package forum

import (
	"strings"
	"time"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
	"go.hasen.dev/vpack"
)

type Post struct {
	Id        int
	UserId    int
	CreatedAt time.Time

	Content string
}

func PackPost(self *Post, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.Id, buf)
	vpack.Int(&self.UserId, buf)
	vpack.UnixTime(&self.CreatedAt, buf)
	vpack.String(&self.Content, buf)
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
		content = content[end:]
		const MaxTagLen = 20
		if len(tag) > MaxTagLen {
			tag = tag[:MaxTagLen]
		}
		tags = append(tags, tag)
	}
	return
}

var PostsBkt = vbolt.Bucket(&dbInfo, "posts", vpack.FInt, PackPost)

// UserPostsIdx term: user id. priority: timestamp. target: post id
var UserPostsIdx = vbolt.IndexExt(&dbInfo, "user-posts", vpack.FInt, vpack.UnixTimeKey, vpack.FInt)

// HashTagsIdx term: hashtag, priority: timestamp, term: post id
var HashTagsIdx = vbolt.IndexExt(&dbInfo, "hashtags", vpack.StringZ, vpack.UnixTimeKey, vpack.FInt)

type CreatePostReq struct {
	UserId  int
	Content string
}

func CreatePost(ctx *vbeam.Context, req CreatePostReq) (post Post, err error) {
	// don't bother validating anything for now
	// TODO: use sessions, and validate content before saving to db

	const MaxPostSize = 1024 * 2
	if len(req.Content) > MaxPostSize {
		req.Content = req.Content[:MaxPostSize]
	}

	vbeam.UseWriteTx(ctx)

	post.Id = vbolt.NextIntId(ctx.Tx, PostsBkt)
	post.UserId = req.UserId
	post.Content = req.Content
	post.CreatedAt = time.Now()

	vbolt.Write(ctx.Tx, PostsBkt, post.Id, &post)

	vbolt.SetTargetSingleTermExt(
		ctx.Tx,         // transaction
		UserPostsIdx,   // index reference
		post.Id,        // target
		post.CreatedAt, // priority
		post.UserId,    // term (single)
	)

	tags := ExtractHashTags(post.Content)
	vbolt.SetTargetTermsUniform(
		ctx.Tx,         // transaction
		HashTagsIdx,    // index reference
		post.Id,        // target
		tags,           // terms (slice)
		post.CreatedAt, // priority (same for all terms)
	)

	vbolt.TxCommit(ctx.Tx)

	return
}

type Posts struct {
	Posts []Post
}

type ByUserReq struct {
	UserId int
}

func PostsByUser(ctx *vbeam.Context, req ByUserReq) (resp Posts, err error) {
	const Limit = 100
	var window = vbolt.Window{Limit: Limit}
	var postIds []int
	vbolt.ReadTermTargets(
		ctx.Tx,       // the transaction
		UserPostsIdx, // the index
		req.UserId,   // the query term
		&postIds,     // slice to store matching targets
		window,       // query windowing
	)
	vbolt.ReadSlice(ctx.Tx, PostsBkt, postIds, &resp.Posts)
	return
}

type ByHashtagReq struct {
	Hashtag string
}

func PostsByHashtag(ctx *vbeam.Context, req ByHashtagReq) (resp Posts, err error) {
	const Limit = 100
	var window = vbolt.Window{Limit: Limit}
	var postIds []int
	vbolt.ReadTermTargets(
		ctx.Tx,      // the transaction
		HashTagsIdx, // the index
		req.Hashtag, // the query term
		&postIds,    // slice to store matching targets
		window,      // query windowing
	)
	vbolt.ReadSlice(ctx.Tx, PostsBkt, postIds, &resp.Posts)
	return
}
