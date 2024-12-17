package forum

import (
	"fmt"
	"strings"
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

	Content string
	Tags    []string
}

func PackPost(self *Post, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.Id, buf)
	vpack.Int(&self.UserId, buf)
	vpack.UnixTime(&self.CreatedAt, buf)
	vpack.String(&self.Content, buf)
	vpack.Slice(&self.Tags, vpack.String, buf)
}

// ExtractHashTags returns a list of hashtags in content, without the hash itself
func ExtractHashTags(content string) (tags []string) {
	for len(content) > 0 {
		start := strings.IndexByte(content, '#')
		if start == -1 {
			break
		}
		start++ // skip the hash itself
		end := strings.IndexAny(content[start:], " \n\t")
		if end == -1 {
			end = len(content[start:])
		}
		tag := content[start : start+end]
		content = content[start+end:]
		const MaxTagLen = 20
		if len(tag) > MaxTagLen {
			tag = tag[:MaxTagLen]
		}
		tags = append(tags, tag)
	}
	return
}

var PostsBkt = vbolt.Bucket(&dbInfo, "posts", vpack.FInt, PackPost)

// PostsIdx term: string, priority: timestamp, target: post id
var PostsIdx = vbolt.IndexExt(&dbInfo, "posts_by", vpack.StringZ, vpack.UnixTimeKey, vpack.FInt)

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
	post.Tags = ExtractHashTags(content)

	vbolt.Write(ctx.Tx, PostsBkt, post.Id, &post)
	UpdatePostIndex(ctx.Tx, post)

	vbolt.TxCommit(ctx.Tx)
	return
}

func UpdatePostIndex(tx *vbolt.Tx, post Post) {
	terms := make([]string, 0, len(post.Tags)+3)
	generic.Append(&terms, fmt.Sprintf("u:%d", post.UserId))
	generic.Append(&terms, fmt.Sprintf("y:%d", post.CreatedAt.Year()))
	generic.Append(&terms, fmt.Sprintf("m:%s", post.CreatedAt.Format("2006.01")))
	for _, tag := range post.Tags {
		generic.Append(&terms, "t:"+tag)
	}
	priority := post.CreatedAt
	vbolt.SetTargetTermsUniform(
		tx,       // transaction
		PostsIdx, // index reference
		post.Id,  // target
		terms,    // terms (slice)
		priority, // priority (same for all terms)
	)
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
