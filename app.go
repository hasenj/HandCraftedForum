package forum

import (
	"forum/cfg"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

var dbInfo vbolt.Info

func OpenDB(dbpath string) *vbolt.DB {
	db := vbolt.Open(dbpath)
	vbolt.InitBuckets(db, &dbInfo)

	// migrations
	const BatchSize = 20

	vbolt.ApplyDBProcess(db, "2024-1217-unify-post-index", func() {
		// delete old indexes
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			tx.DeleteBucket([]byte("user-posts"))
			tx.DeleteBucket([]byte("hashtags"))
			tx.Commit()
		})
		// populate the new index
		vbolt.TxWriteBatches(db, PostsBkt, BatchSize, func(tx *vbolt.Tx, batch []Post) {
			for _, post := range batch {
				UpdatePostIndex(tx, post)
			}
			vbolt.TxCommit(tx)
		})
	})

	return db
}

func MakeApplication() *vbeam.Application {
	vbeam.RunBackServer(cfg.Backport)

	db := OpenDB(cfg.DBPath)

	var app = vbeam.NewApplication("HandCraftedForum", db)
	vbeam.RegisterProc(app, AddUser)
	vbeam.RegisterProc(app, ListUsers)

	vbeam.RegisterProc(app, CreatePost)
	vbeam.RegisterProc(app, QueryPosts)

	return app
}

type Empty struct{}
