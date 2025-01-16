package forum

import (
	"errors"
	"fmt"
	"time"

	"go.hasen.dev/generic"
	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
	"go.hasen.dev/vpack"
)

type Post struct {
	Id        int
	UserId    int
	CreatedAt time.Time

	ParentId int
	Content  string
}

func PackPost(self *Post, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.Id, buf)
	vpack.Int(&self.UserId, buf)
	vpack.Int(&self.ParentId, buf)
	vpack.UnixTime(&self.CreatedAt, buf)
	vpack.String(&self.Content, buf)
}

var PostsBkt = vbolt.Bucket(&dbInfo, "posts", vpack.FInt, PackPost)

// PostsIdx term: string, priority: timestamp, target: post id
var PostsIdx = vbolt.IndexExt(&dbInfo, "posts_by",
	vpack.StringZ, vpack.UnixTimeKey, vpack.FInt)

// PostRepliesIdx term: post_id, priority: timestamp target: ancestor post ids
var PostRepliesIdx = vbolt.IndexExt(&dbInfo, "posts_replies",
	vpack.FInt, vpack.UnixTimeKey, vpack.FInt)

func SavePost(tx *vbolt.Tx, post *Post) {
	vbolt.Write(tx, PostsBkt, post.Id, post)
	UpdatePostIndex(tx, post)
}

func UpdatePostIndex(tx *vbolt.Tx, post *Post) {
	terms := make([]string, 0, 3)
	generic.Append(&terms, fmt.Sprintf("u:%d", post.UserId))
	generic.Append(&terms, fmt.Sprintf("y:%d", post.CreatedAt.Year()))
	generic.Append(&terms, fmt.Sprintf("m:%s", post.CreatedAt.Format("2006.01")))
	priority := post.CreatedAt
	vbolt.SetTargetTermsUniform(
		tx,       // transaction
		PostsIdx, // index reference
		post.Id,  // target
		terms,    // terms (slice)
		priority, // priority (same for all terms)
	)

	if post.ParentId != 0 {
		ancestors := make([]int, 0)
		generic.Append(&ancestors, post.ParentId)
		vbolt.IterateTarget(tx, PostRepliesIdx, post.ParentId,
			func(ancestorId int, _priority time.Time) bool {
				generic.Append(&ancestors, ancestorId)
				return true
			})
		vbolt.SetTargetTermsUniform(
			tx,
			PostRepliesIdx,
			post.Id,
			ancestors,
			post.CreatedAt,
		)
	}
}

type CreatePostReq struct {
	UserId  int
	Content string
}

func CreatePost(ctx *vbeam.Context, req CreatePostReq) (post Post, err error) {
	// don't bother validating anything for now
	// TODO: use sessions, and validate content before saving to db

	const MaxPostSize = 1024 * 2

	content := req.Content

	if len(content) > MaxPostSize {
		content = content[:MaxPostSize]
	}

	vbeam.UseWriteTx(ctx)

	post.Id = vbolt.NextIntId(ctx.Tx, PostsBkt)
	post.UserId = req.UserId
	post.Content = content
	post.CreatedAt = time.Now()

	SavePost(ctx.Tx, &post)

	vbolt.TxCommit(ctx.Tx)
	return
}

type PostsQuery struct {
	Query  string
	Cursor []byte
}

type PostsResponse struct {
	Posts      []Post
	NextParams PostsQuery
}

const Limit = 2

func QueryPosts(ctx *vbeam.Context, req PostsQuery) (resp PostsResponse, err error) {
	var window = vbolt.Window{
		Limit:     Limit,
		Direction: vbolt.IterateReverse,
		Cursor:    req.Cursor,
	}
	var postIds []int
	resp.NextParams = req
	resp.NextParams.Cursor = vbolt.ReadTermTargets(
		ctx.Tx,    // the transaction
		PostsIdx,  // the index
		req.Query, // the query term
		&postIds,  // slice to store matching targets
		window,    // query windowing
	)
	vbolt.ReadSlice(ctx.Tx, PostsBkt, postIds, &resp.Posts)
	generic.EnsureSliceNotNil(&resp.Posts)
	generic.EnsureSliceNotNil(&resp.NextParams.Cursor)
	return
}

type PostQuery struct {
	PostId int
}

type PostResponse struct {
	PostIds []int
	Posts   map[int]Post
	Replies map[int]int
	Users   map[int]User
}

var PostNotFound = errors.New("PostNotFound")

func GetPost(ctx *vbeam.Context, req PostQuery) (resp PostResponse, err error) {
	generic.InitMap(&resp.Posts)
	generic.InitMap(&resp.Replies)
	generic.InitMap(&resp.Users)

	// parents
	vbolt.IterateTarget(ctx.Tx, PostRepliesIdx, req.PostId, func(parentId int, _p time.Time) bool {
		generic.Append(&resp.PostIds, parentId)
		return true
	})

	// self
	generic.Append(&resp.PostIds, req.PostId)

	// children
	vbolt.ReadTermTargets(ctx.Tx, PostRepliesIdx, req.PostId, &resp.PostIds, vbolt.Window{})

	vbolt.ReadSliceToMap(ctx.Tx, PostsBkt, resp.PostIds, resp.Posts)

	// reply counts
	for _, postId := range resp.PostIds {
		var replies int
		vbolt.ReadTermCount(ctx.Tx, PostRepliesIdx, &postId, &replies)
		resp.Replies[postId] = replies
	}

	// users
	for _, post := range resp.Posts {
		_, found := resp.Users[post.UserId]
		if !found {
			var user User
			if vbolt.Read(ctx.Tx, UsersBkt, post.UserId, &user) {
				resp.Users[post.UserId] = user
			}
		}
	}
	return
}
