package main

import (
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/lukevers/golem"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Session store for users
var store = sessions.NewFilesystemStore(
	// Path
	"app/sessions",
	// Secret key with strength set to 64
	[]byte(securecookie.GenerateRandomKey(64)),
)

type User struct {
	// ID is an int64 that is a users identification
	// number.
	Id uint64

	// Username is a string with max-size set to 255
	// and is the username that a user will use when
	// logging in to the web interface.
	Username string `sql:"size:255;unique"`

	// Password is a string with max-size set to 255
	// and is the password that a user will use when
	// logging in to the web interface.
	Password string `sql:"size:255"`

	// Admin is a bool that specifies if the current
	// user is an administrator or not.
	Admin bool

	// Twofa (2fa) is a bool that specifies if the
	// current user is using 2fa or not.
	Twofa bool

	// TwofaSecret is a base32 encoded string of the
	// 2fa secret.
	TwofaSecret string

	// CreatedAt is a timestamp of when the specific
	// user was created at.
	CreatedAt time.Time

	// UpdatedAt is a timestamp of when the specific
	// user was last updated at.
	UpdatedAt time.Time
}

// Hash Password takes a string and hashes that password
// and returns it as a string. It handles errors that are
// returned from bcrypt.GenerateFromPassword, and is a
// wrapper around having to use []byte everywhere.
func HashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		golem.Warnf("Error hashing password: %s", err)
	}

	return string(hash)
}

// Password Matches Hash takes a plaintext password and uses
// bcrypt.CompareHashAndPassword to check against the hashed
// password we're checking against from the database. The
// func from bcrypt returns nil if the passwords match, and
// an error otherwise, so we're checking if bcrypt's func
// returns nil or not and that's how we're determining if the
// hashes match or not.
func PasswordMatchesHash(password string, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// Clean Sessions is a func that is ran when the app is started
// and removes all the old session data from "app/sessions" because
// the old sessions are useless since a secret key is randomly
// generated on start.
func CleanSessions() {
	// Find all sessions
	sessions, err := filepath.Glob("app/sessions/session_*")
	if err != nil {
		golem.Warnf("Error finding sessions: %s", err)
	}

	// Loop through sessions and delete
	for _, s := range sessions {
		err = os.Remove(s)
		if err != nil {
			golem.Warnf("Error deleting session: %s", err)
		}
	}
}

// Is Logged In checks if the user has a session or not.
// If the user does not have a session that matches with
// what we have, then the user is not logged in.
func IsLoggedIn(w http.ResponseWriter, req *http.Request) bool {
	// Check for session
	session, _ := store.Get(req, "user")
	if session.IsNew {
		return false
	}

	// Check if user has Twofa
	user := WhoAmI(req)
	if user == nil {
		// If user is nil then the user probably deleted their own user.
		// If we get here, we should remove the session immediately
		http.SetCookie(w, &http.Cookie{
			Name:   "user",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		return false
	} else {
		if user.Twofa {
			// Check if temporary session
			if session.Values["temp"] == "true" {
				return false
			} else {
				return true
			}
		} else {
			return true
		}
	}
}

// WhoAmI figures out who exactly is using the current
// session (what user is), and it returns the *User from
// the slice of Users that we have.
func WhoAmI(req *http.Request) *User {
	session, _ := store.Get(req, "user")

	for _, user := range users {
		if session.Values["username"] == user.Username {
			return user
		}
	}

	return nil
}
