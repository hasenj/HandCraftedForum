package forum

import (
	"forum/cfg"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

var dbInfo vbolt.Info

func MakeApplication() *vbeam.Application {
	vbeam.RunBackServer(cfg.Backport)

	db := vbolt.Open(cfg.DBPath)
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		vbolt.TxRawBucket(tx, "proc") // special
		vbolt.EnsureBuckets(tx, &dbInfo)
		tx.Commit()
	})

	var app = vbeam.NewApplication("HandCraftedForum", db)
	vbeam.RegisterProc(app, AddUser)
	vbeam.RegisterProc(app, ListUsers)
	return app
}

type Empty struct{}
