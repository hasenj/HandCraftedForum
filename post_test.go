package forum

import (
	"os"
	"testing"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

func TestPosting(t *testing.T) {
	testDBPath := "test.db"

	db := OpenDB("test.db")
	defer os.Remove(testDBPath)

	withContext := func(fn func(ctx *vbeam.Context)) {
		var ctx vbeam.Context
		ctx.Tx = vbolt.ReadTx(db)
		defer vbolt.TxClose(ctx.Tx)
		fn(&ctx)
	}

	// data for creating posts
	reqs := []CreatePostReq{
		{UserId: 1, Content: "Hello #World #T1"},
		{UserId: 1, Content: "Hello #World #T2"},
		{UserId: 1, Content: "#Hello World #T3"},

		{UserId: 2, Content: "Hello #World #T1"},
		{UserId: 2, Content: "#Hello World #T2"},

		{UserId: 3, Content: "#Hello #World #T1"},
	}

	// expected counts for each query
	queryCounts := map[string]int{
		"t:T1": 3,
		"t:T2": 2,
		"t:T3": 1,

		"t:World": 4,
		"t:Hello": 3,

		"u:1": 3,
		"u:2": 2,
		"u:3": 1,
	}

	// create posts
	for _, req := range reqs {
		withContext(func(ctx *vbeam.Context) {
			_, err := CreatePost(ctx, req)
			if err != nil {
				t.Fatalf("Post Creation Failed: %v", err)
			}
		})
	}

	// query posts
	for tag, count := range queryCounts {
		withContext(func(ctx *vbeam.Context) {
			res, err := QueryPosts(ctx, PostsQuery{Query: tag})
			if err != nil {
				t.Fatal(err)
			}
			if len(res.Posts) != count {
				t.Fatalf("Expected: %d, actual: %d", count, len(res.Posts))
			}
		})
	}
}
