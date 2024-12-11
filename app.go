package forum

import (
	"forum/cfg"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

var dbInfo vbolt.Info

func OpenDB(dbpath string) *vbolt.DB {
	db := vbolt.Open(dbpath)
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		vbolt.TxRawBucket(tx, "proc") // special
		vbolt.EnsureBuckets(tx, &dbInfo)
		tx.Commit()
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
	vbeam.RegisterProc(app, PostsByUser)
	vbeam.RegisterProc(app, PostsByHashtag)

	return app
}

type Empty struct{}
