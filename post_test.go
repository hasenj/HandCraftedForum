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

	reqs := []CreatePostReq{
		{UserId: 1, Content: "Hello #World #T1"},
		{UserId: 1, Content: "Hello #World #T2"},
		{UserId: 1, Content: "#Hello World #T3"},

		{UserId: 2, Content: "Hello #World #T1"},
		{UserId: 2, Content: "#Hello World #T2"},

		{UserId: 3, Content: "#Hello #World #T1"},
	}

	tagsCounts := map[string]int{
		"T1": 3,
		"T2": 2,
		"T3": 1,

		"World": 4,
		"Hello": 3,
	}

	for _, req := range reqs {
		withContext(func(ctx *vbeam.Context) {
			_, err := CreatePost(ctx, req)
			if err != nil {
				t.Fatalf("Post Creation Failed: %v", err)
			}
		})
	}

	for tag, count := range tagsCounts {
		withContext(func(ctx *vbeam.Context) {
			res, err := PostsByHashtag(ctx, ByHashtagReq{tag})
			if err != nil {
				t.Fatal(err)
			}
			if len(res.Posts) != count {
				t.Fatalf("Expected: %d, actual: %d", count, len(res.Posts))
			}
		})
	}
}
