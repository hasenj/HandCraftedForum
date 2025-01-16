package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"forum"
	"forum/cfg"

	"go.hasen.dev/generic"
	"go.hasen.dev/vbolt"
)

func getData(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func getResource(name string, target any) error {
	localPath := "hn/" + name
	url := "https://hacker-news.firebaseio.com/v0/" + name

	data, err := os.ReadFile(localPath)
	if err != nil {
		// fetch content from url and save to localPath
		data, err = getData(url)
		if err != nil {
			return err
		}
		err = os.WriteFile(localPath, data, 0644)
		if err != nil {
			return err
		}
	}

	json.Unmarshal(data, target)
	// fmt.Println(string(data))

	return err
}

func getItem(id int, item *PostItem) error {
	res := fmt.Sprintf("item/%d.json", id)
	return getResource(res, item)
}

type PostItem struct {
	Id       int    `json:"id"`
	Username string `json:"by"`

	Parent int   `json:"parent"`
	Kids   []int `json:"kids"`

	Time  int64  `json:"time"`
	Title string `json:"title"`
	Text  string `json:"text"`

	Type string `json:"type"`
	URL  string `json:"url"`
}

func main() {
	os.MkdirAll("hn/item", 0755)

	if len(os.Args) < 2 {
		fmt.Println("Pass the post Id as the first parameter")
		return
	}

	postId, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("invalid post id:", os.Args[1])
		fmt.Println(err)
		return
	}

	downloadPostById(postId)
}

func downloadPostById(postId int) {
	userIdMap := make(map[string]int)
	userIds := make([]string, 0)
	addedUsers := make([]string, 0)

	posts := make([]PostItem, 0)

	postIds := []int{
		postId,
	}
	postIt := 0

	for postIt < len(postIds) {
		post := generic.AllocAppend(&posts)
		nextPostId := postIds[postIt]
		postIt++
		fmt.Printf("Downloading post: %d   \r", nextPostId)
		err := getItem(nextPostId, post)
		if err != nil {
			fmt.Println("Error:", err)
		}
		generic.Append(&postIds, post.Kids...)
		generic.Append(&userIds, post.Username)
	}
	fmt.Printf("Downloaded %d posts         \n", len(postIds))

	db := forum.OpenDB(cfg.DBPath)
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		for i := range posts {
			postItem := &posts[i]
			if postItem.Username != "" {
				userId, ok := userIdMap[postItem.Username]
				if !ok {
					if !vbolt.Read(tx, forum.UsernameBkt, postItem.Username, &userId) {
						user := forum.AddUserTx(tx, forum.AddUserRequest{
							Username: postItem.Username,
							Email:    postItem.Username + ".hn@example.com",
						}, nil)
						userId = user.Id
						generic.Append(&addedUsers, postItem.Username)
						fmt.Printf("Creating user: %s     \r", postItem.Username)
					}
					userIdMap[postItem.Username] = userId
				}
			}
		}
		fmt.Printf("Created %d users      \n", len(addedUsers))

		for i := range posts {
			postItem := &posts[i]
			var post forum.Post
			post.Id = postItem.Id
			var parts []string
			if postItem.URL != "" {
				title := fmt.Sprintf(`<a href="%s">%s</a>`, postItem.URL, postItem.Title)
				parts = append(parts, title)
			}
			parts = append(parts, postItem.Text)
			post.Content = strings.Join(parts, "<br />")
			post.CreatedAt = time.Unix(postItem.Time, 0)
			post.UserId = userIdMap[postItem.Username]
			post.ParentId = postItem.Parent
			forum.SavePost(tx, &post)
			fmt.Printf("Saving post: %d    \r", post.Id)
		}
		fmt.Printf("Done         \n")

		vbolt.TxCommit(tx)
	})

	getItem(postId, &posts[0])
}
