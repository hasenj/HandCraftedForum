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
	vbolt.ApplyDBProcess(db, "2025-0114-reset-posts", func() {
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			tx.DeleteBucket([]byte(PostsBkt.Name))
			tx.CreateBucket([]byte(PostsBkt.Name))
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

	vbeam.RegisterProc(app, GetPost)

	return app
}

type Empty struct{}
