package main

import (
	"flag"
	"github.com/gorilla/mux"
	"github.com/lukevers/golem"
	"html/template"
	"net/http"
	"strconv"
)

var (
	err   error
	users []*User
)

var templates = template.Must(template.New("").Funcs(AddTemplateFunctions(nil)).ParseGlob("app/views/*"))

func main() {
	// Parse flags
	flag.Parse()

	// Initalize database
	initalizeDB()

	// Remove old sessions
	CleanSessions()

	// Create web server
	r := mux.NewRouter()

	// Handles GET requests for "/" which is our root page.
	r.HandleFunc("/", HandleRoot)

	// Handles GET requests to "/login" which displays a form that a user
	// can use to try and login.
	r.HandleFunc("/login", HandleLogin).Methods("GET")

	// Handles POST requests for "/login" which tests if a user is
	// logging in with correct details or not.
	r.HandleFunc("/login", HandleLoginForm).Methods("POST")

	// Handles GET requests to "/login/2fa" which displays a form that a
	// user can use to try their 2fa token on.
	r.HandleFunc("/login/2fa", HandleLogin2FA).Methods("GET")

	// Handles POST requests to "/login/2fa" which tests if a users 2fa
	// token is correct or not.
	r.HandleFunc("/login/2fa", HandleLoginForm2FA).Methods("POST")

	// Handle logout requests which removes the session and logs the user out
	r.HandleFunc("/logout", HandleLogout)

	// Handles GET requests for "/settings" which is a page where users
	// can update their settings.
	r.HandleFunc("/settings", HandleSettings).Methods("GET")

	// Handles POST requests for "/settings" which is a page where users
	// can update their settings. POSTing here will update settings.
	r.HandleFunc("/settings", HandleUpdateSettings).Methods("POST")

	// Handles GET requests for "/settings/2fa/generate" which is a page
	// that generates a QR code for Two Factor Auth.
	r.HandleFunc("/settings/2fa/generate", HandleGenerate2FA).Methods("GET")

	// Handles POST requests for "/settings/2fa/verify" which checks to
	// see if a 2FA token is correct. If it is, then we remove the temp
	// secret key from the session and add it to the database.
	r.HandleFunc("/settings/2fa/verify", HandleVerify2FA).Methods("POST")

	// Handles POST requests for "/settings/2fa/disable" which disables
	// 2FA for the users account.
	r.HandleFunc("/settings/2fa/disable", HandleDisable2FA).Methods("POST")

	// Handles GET requests for "/users" which is an admin-only page
	r.HandleFunc("/users", HandleUsers).Methods("GET")

	// Handles POST requests for "/users/new" which is a form where
	// new users can be added.
	r.HandleFunc("/users/new", HandleNewUser).Methods("POST")

	// Handles POST requests for "/users/delete" which is how users
	// can be deleted.
	r.HandleFunc("/users/delete", HandleUserDelete).Methods("POST")

	// Handles POST requests for "/users/admin" which is a form where
	// administrators can promote/demote users.
	r.HandleFunc("/users/admin", HandleUserAdminSwitch).Methods("POST")

	// Handle all other static files and folders (eg. CSS/JS).
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))

	// Get all users
	db.Find(&users, &User{})

	// Start web server
	golem.Infof("Starting webserver on port %s", strconv.Itoa(*portFlag))
	http.Handle("/", r)
	http.ListenAndServe(*interfaceFlag+":"+strconv.Itoa(*portFlag), nil)
}
