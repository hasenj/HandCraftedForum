package forum

import (
	"errors"

	"go.hasen.dev/generic"
	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
	"go.hasen.dev/vpack"
	"golang.org/x/crypto/bcrypt"
)

// Models
// =============================================================================

type User struct {
	Id       int
	Username string
	Email    string
	IsAdmin  bool
}

func PackUser(self *User, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.Id, buf)
	vpack.String(&self.Username, buf)
	vpack.String(&self.Email, buf)
	vpack.Bool(&self.IsAdmin, buf)
}

// Buckets
// =============================================================================

var UsersBkt = vbolt.Bucket(&dbInfo, "users", vpack.FInt, PackUser)

// user id => hashed passwd
var PasswdBkt = vbolt.Bucket(&dbInfo, "passwd", vpack.FInt, vpack.ByteSlice)

// this is to ensure username uniqueness
// username => userid
var UsernameBkt = vbolt.Bucket(&dbInfo, "username", vpack.StringZ, vpack.Int)

// Procedures
// =============================================================================

type AddUserRequest struct {
	Username string
	Email    string
	Password string
}

type UserListResponse struct {
	Users []User
}

func fetchUsers(tx *vbolt.Tx) (users []User) {
	vbolt.IterateAll(tx, UsersBkt, func(key int, value User) bool {
		generic.Append(&users, value)
		return true
	})
	return
}

func isUsernameValid(name string) bool {
	if len(name) < 3 {
		return false
	}
	for _, c := range name {
		if c > 0xff {
			return false
		}
		// alpha, ok
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			continue
		}
		// numbers, ok
		if c >= '0' && c <= '9' {
			continue
		}
		// special separators, ok
		if c == '.' || c == '-' || c == '_' {
			continue
		}
		// unrecognized character
		return false
	}
	// if we reach here, no problems were found
	return true
}

func isPasswordValid(pwd string) bool {
	// cannot hash a password with length over 72
	return len(pwd) >= 8 && len(pwd) <= 72
}

var UsernameTaken = errors.New("UsernameTaken")
var UsernameInvalid = errors.New("UsernameInvalid")
var PasswordInvalid = errors.New("PasswordInvalid")

// AddUserTx adds the user without any validation; if username exists it will be
// over written!
func AddUserTx(tx *vbolt.Tx, req AddUserRequest, hash []byte) User {
	var user User
	user.Id = vbolt.NextIntId(tx, UsersBkt)
	user.Username = req.Username
	user.Email = req.Email
	user.IsAdmin = user.Id < 2

	vbolt.Write(tx, UsersBkt, user.Id, &user)
	vbolt.Write(tx, PasswdBkt, user.Id, &hash)
	vbolt.Write(tx, UsernameBkt, user.Username, &user.Id)
	return user
}

func ValidateUserTx(tx *vbolt.Tx, req AddUserRequest) error {
	if !isUsernameValid(req.Username) {
		return UsernameInvalid
	}

	// check username is not already taken
	if vbolt.HasKey(tx, UsernameBkt, req.Username) {
		return UsernameTaken
	}

	if !isPasswordValid(req.Password) {
		return PasswordInvalid
	}

	return nil
}

func AddUser(ctx *vbeam.Context, req AddUserRequest) (resp UserListResponse, err error) {
	err = ValidateUserTx(ctx.Tx, req)
	if err != nil {
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	// start a write transaction!!
	vbeam.UseWriteTx(ctx)

	AddUserTx(ctx.Tx, req, hash)

	resp.Users = fetchUsers(ctx.Tx)
	generic.EnsureSliceNotNil(&resp.Users)

	vbolt.TxCommit(ctx.Tx)
	return
}

func ListUsers(ctx *vbeam.Context, req Empty) (resp UserListResponse, err error) {
	resp.Users = fetchUsers(ctx.Tx)
	generic.EnsureSliceNotNil(&resp.Users)
	return
}
