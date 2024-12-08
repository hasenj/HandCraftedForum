package forum

import (
	"forum/cfg"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

func MakeApplication() *vbeam.Application {
	vbeam.RunBackServer(cfg.Backport)
	db := vbolt.Open(cfg.DBPath)
	var app = vbeam.NewApplication("HandCraftedForum", db)
	vbeam.RegisterProc(app, AddUser)
	vbeam.RegisterProc(app, ListUsers)
	return app
}

// global (but volatile) list of usernames
var usernames = make([]string, 0)

type AddUserRequest struct {
	Username string
}

type UserListResponse struct {
	AllUsernames []string
}

func AddUser(ctx *vbeam.Context, req AddUserRequest) (resp UserListResponse, err error) {
	usernames = append(usernames, req.Username)
	resp.AllUsernames = usernames
	return
}

type Empty struct{}

func ListUsers(ctx *vbeam.Context, req Empty) (resp UserListResponse, err error) {
	resp.AllUsernames = usernames
	return
}
