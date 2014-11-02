package main

import (
	"code.google.com/p/rsc/qr"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"github.com/dgryski/dgoogauth"
	"github.com/lukevers/golem"
	"net/http"
	"strconv"
	"strings"
)

// Handle "/" web
func HandleRoot(w http.ResponseWriter, req *http.Request) {
	if *debugFlag {
		templates = RefreshTemplates(req)
	}

	if IsLoggedIn(w, req) {
		templates.Funcs(AddTemplateFunctions(req)).ExecuteTemplate(w, "index", WhoAmI(req))
	} else {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	}
}

// Handle "/logout" web
func HandleLogout(w http.ResponseWriter, req *http.Request) {
	// Remove cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "user",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Redirect to "/"
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

// Handle "/login" web
func HandleLogin(w http.ResponseWriter, req *http.Request) {
	if *debugFlag {
		templates = RefreshTemplates(req)
	}

	if IsLoggedIn(w, req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
	} else {
		// Check if we have been partly authenticated yet
		session, _ := store.Get(req, "user")
		if session.IsNew {
			templates.Funcs(AddTemplateFunctions(req)).ExecuteTemplate(w, "login", nil)
		} else {
			http.Redirect(w, req, "/login/2fa", http.StatusSeeOther)
		}
	}
}

// Handle "/login/2fa" web
func HandleLogin2FA(w http.ResponseWriter, req *http.Request) {
	if *debugFlag {
		templates = RefreshTemplates(req)
	}

	if IsLoggedIn(w, req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
	} else {
		session, _ := store.Get(req, "user")
		if session.Values["temp"] == "true" {
			templates.Funcs(AddTemplateFunctions(req)).ExecuteTemplate(w, "login_2fa", nil)
		} else {
			http.Redirect(w, req, "/login", http.StatusSeeOther)
		}
	}
}

// Handle POSTS to "/login/2fa" web
func HandleLoginForm2FA(w http.ResponseWriter, req *http.Request) {
	// Check if we have been partly authenticated yet
	session, _ := store.Get(req, "user")
	if session.IsNew {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	} else {
		// Parse our form so we can get values from req.Form
		err = req.ParseForm()
		if err != nil {
			golem.Warnf("Error parsing form: %s", err)
		}

		// Get token from input
		token := req.Form["token"][0]

		// Get user
		user := WhoAmI(req)

		// Configure token
		otpc := &dgoogauth.OTPConfig{
			Secret:      user.TwofaSecret,
			WindowSize:  3,
			HotpCounter: 0,
		}

		// Validate token
		val, err := otpc.Authenticate(token)
		if err != nil {
			golem.Warnf("Error authenticating token: %s", err)
		}

		if val {
			// Validated
			session, _ := store.Get(req, "user")
			session.Values["temp"] = "false"
			session.Save(req, w)

			// Redirect
			http.Redirect(w, req, "/", http.StatusSeeOther)
		} else {
			// Not validated
			http.Redirect(w, req, "/login/2fa", http.StatusSeeOther)
		}
	}
}

// Handle POSTS to "/login" web.
func HandleLoginForm(w http.ResponseWriter, req *http.Request) {
	// Parse our form so we can get values from req.Form
	err = req.ParseForm()
	if err != nil {
		golem.Warnf("Error parsing form: %s", err)
	}

	// Get username/password from input
	username := req.Form["username"][0]
	password := req.Form["password"][0]

	// Query database for user
	var user User
	db.Table("users").Where("username = ?", username).First(&user)

	// Check if usernames match up
	if user.Username == username {
		// Check if passwords match up
		if PasswordMatchesHash(password, user.Password) {
			// Create new session
			session, _ := store.New(req, "user")
			session.Values["username"] = username

			// Save session
			session.Save(req, w)

			// Check if 2fa is enabled
			if user.Twofa {
				session.Values["temp"] = "true"
				session.Save(req, w)

				// Redirect, check 2fa
				http.Redirect(w, req, "/login/2fa", http.StatusSeeOther)
			} else {
				// Redirect, logged in ok
				http.Redirect(w, req, "/", http.StatusSeeOther)
			}
		}
	}

	// If you have gotten this far then you have not been
	// authenticated. Sorry.
	http.Redirect(w, req, "/login", http.StatusSeeOther)
}

// Handle "/settings" web
func HandleSettings(w http.ResponseWriter, req *http.Request) {
	if !IsLoggedIn(w, req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	} else {
		// Refresh the templates
		if *debugFlag {
			templates = RefreshTemplates(req)
		}

		// Execute template
		templates.Funcs(AddTemplateFunctions(req)).ExecuteTemplate(w, "settings", WhoAmI(req))
	}
}

// Handles POST requests to "/settings" which is a page that
// users update their settings at.
func HandleUpdateSettings(w http.ResponseWriter, req *http.Request) {
	if !IsLoggedIn(w, req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	} else {
		// Parse our form so we can get values from req.Form
		err = req.ParseForm()
		if err != nil {
			golem.Warnf("Error parsing form: %s", err)
		}

		username := req.Form["username"][0]
		password := req.Form["password"][0]

		// Figure out who the user is
		u := WhoAmI(req)
		var user User
		db.Table("users").Where("id = ?", u.Id).Find(&user)

		// Update user in database if not empty
		if username != "" {
			user.Username = username
		}

		// Update password in database if not empty
		if password != "" {
			user.Password = HashPassword(password)
		}

		// Update user in memory
		u.Username = user.Username
		u.Password = user.Password

		// Save user
		db.Save(&user)

		// Redirect back to "/settings" when we're done here.
		http.Redirect(w, req, "/settings", http.StatusSeeOther)
	}
}

// Handles GET AJAX requests to "/settings/2fa/generate" which
// generates a QR code for Two Factor Auth.
func HandleGenerate2FA(w http.ResponseWriter, req *http.Request) {
	if !IsLoggedIn(w, req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	} else {
		if req.Header.Get("X-Requested-With") != "XMLHttpRequest" {
			http.Redirect(w, req, "/settings", http.StatusSeeOther)
		} else {
			// Get random secret
			sec := make([]byte, 6)
			_, err = rand.Read(sec)
			if err != nil {
				golem.Warnf("Error creating random secret key: %s", err)
			}

			// Encode secret to base32 string
			secret := base32.StdEncoding.EncodeToString(sec)

			// Create auth string to be encoded as a QR image
			//
			// https://code.google.com/p/google-authenticator/wiki/KeyUriFormat
			// otpauth://totp/Example:alice@google.com?secret=JBSWY3DPEHPK3PXP&issuer=Example
			//
			auth_string := "otpauth://totp/KittensIRC?secret=" + secret + "&issuer=KittensIRC"

			// Encode the QR image
			code, err := qr.Encode(auth_string, qr.L)
			if err != nil {
				golem.Warnf("Error encoding qr code: %s", err)
			}

			// Set temporary session values until we verify 2fa is set
			session, _ := store.Get(req, "user")
			session.Values["secret"] = secret
			session.Save(req, w)

			// Write base64 encoded QR image
			w.Write([]byte(base64.StdEncoding.EncodeToString(code.PNG())))
		}
	}
}

// Handles POST AJAX requests to "/settings/2fa/verify" which
// verifies the first 2FA token.
func HandleVerify2FA(w http.ResponseWriter, req *http.Request) {
	if !IsLoggedIn(w, req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	} else {
		if req.Header.Get("X-Requested-With") != "XMLHttpRequest" {
			http.Redirect(w, req, "/settings", http.StatusSeeOther)
		} else {
			// Get our session
			session, _ := store.Get(req, "user")
			secret := session.Values["secret"]

			// Parse our form so we can get values from req.Form
			err = req.ParseForm()
			if err != nil {
				golem.Warnf("Error parsing form: %s", err)
			}

			// Get token from input
			token := req.Form["token"][0]

			// Get user
			u := WhoAmI(req)

			// Configure token
			otpc := &dgoogauth.OTPConfig{
				Secret:      secret.(string),
				WindowSize:  3,
				HotpCounter: 0,
			}

			// Validate token
			val, err := otpc.Authenticate(token)
			if err != nil {
				golem.Warnf("Error authenticating token: %s", err)
			}

			if val {
				// Update user
				u.Twofa = true
				u.TwofaSecret = secret.(string)

				// Update user in database
				var user User
				db.Table("users").Where("id = ?", u.Id).Find(&user)
				user.Twofa = true
				user.TwofaSecret = secret.(string)
				db.Save(&user)

				// Remove secret from session
				session.Values["secret"] = ""
				session.Save(req, w)

				// Return success
				http.Redirect(w, req, "/settings", http.StatusSeeOther)
			} else {
				// Return error
				http.Error(w, "Wrong token", http.StatusExpectationFailed)
			}
		}
	}
}

// Handles POST AJAX requests to "/settings/2fa/disable" which
// disables 2fa for a users account.
func HandleDisable2FA(w http.ResponseWriter, req *http.Request) {
	if !IsLoggedIn(w, req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	} else {
		if req.Header.Get("X-Requested-With") != "XMLHttpRequest" {
			http.Redirect(w, req, "/settings", http.StatusSeeOther)
		} else {
			// Get user
			u := WhoAmI(req)

			// Update user
			u.Twofa = false
			u.TwofaSecret = ""

			// Update user in database
			var user User
			db.Table("users").Where("id = ?", u.Id).Find(&user)
			user.Twofa = false
			user.TwofaSecret = ""
			db.Save(&user)

			// Return success
			http.Redirect(w, req, "/settings", http.StatusSeeOther)
		}
	}
}

// Handle "/users" web
func HandleUsers(w http.ResponseWriter, req *http.Request) {
	if *debugFlag {
		templates = RefreshTemplates(req)
	}

	// Check if logged in
	if !IsLoggedIn(w, req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	} else {
		// Get the user
		user := WhoAmI(req)

		// Check if user is admin
		if !user.Admin {
			http.Redirect(w, req, "/", http.StatusSeeOther)
		} else {
			templates.Funcs(AddTemplateFunctions(req)).ExecuteTemplate(w, "users", &users)
		}
	}
}

// Handle "/users/new" web POSTs
func HandleNewUser(w http.ResponseWriter, req *http.Request) {
	if *debugFlag {
		templates = RefreshTemplates(req)
	}

	// Check if logged in
	if !IsLoggedIn(w, req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	} else {
		// Get the user
		user := WhoAmI(req)

		// Check if user is admin
		if !user.Admin {
			http.Redirect(w, req, "/", http.StatusSeeOther)
		} else {
			// Parse our form so we can get values from req.Form
			err = req.ParseForm()
			if err != nil {
				golem.Warnf("Error parsing form: %s", err)
			}

			// Parse admin from string to bool
			admin, err := strconv.ParseBool(req.Form["admin"][0])
			if err != nil {
				golem.Warnf("Error parsing admin from string to bool: %s", err)
			}

			// Check if username is set
			username := strings.Trim(req.Form["username"][0], " ")
			password := strings.Trim(req.Form["password"][0], " ")
			if username == "" {
				// Redirect back to "/users"
				http.Redirect(w, req, "/users", http.StatusSeeOther)
			} else {
				// Check if password is set
				if password == "" {
					// Redirect back to "/users"
					http.Redirect(w, req, "/users", http.StatusSeeOther)
				} else {
					// Create user
					newuser := User{
						Username:    username,
						Password:    HashPassword(password),
						Admin:       admin,
						Twofa:       false,
						TwofaSecret: "",
					}

					// Insert new user into database
					db.Create(&newuser)

					// Save new user
					db.Save(&newuser)

					// Update the users array
					users = append(users, &newuser)

					// Redirect back to "/users" when we're done here
					http.Redirect(w, req, "/users", http.StatusSeeOther)
				}
			}
		}
	}
}

// Handle POSTs to "/user/admin" which switches a users administrative
// settings.
func HandleUserAdminSwitch(w http.ResponseWriter, req *http.Request) {
	if *debugFlag {
		templates = RefreshTemplates(req)
	}

	// Check if logged in
	if !IsLoggedIn(w, req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	} else {
		// Check if user is admin
		if !WhoAmI(req).Admin {
			http.Redirect(w, req, "/", http.StatusSeeOther)
		} else {
			// Parse our form so we can get values from req.Form
			err = req.ParseForm()
			if err != nil {
				golem.Warnf("Error parsing form: %s", err)
			}

			// Get the user we're switching admin values for
			id, err := strconv.ParseUint(req.Form["id"][0], 10, 16)
			if err != nil {
				golem.Warnf("Error converting id: %s", err)
			}

			// Update in database
			var user User
			db.Table("users").Where("id = ?", id).First(&user)
			user.Admin = !user.Admin
			db.Save(&user)

			// Update in memory
			for _, v := range users {
				if id == v.Id {
					v.Admin = !v.Admin
				}
			}

			// Return success
			http.Redirect(w, req, "/users", http.StatusSeeOther)
		}
	}
}

// Handle POSTs to "/user/delete" which deletes a user completely
func HandleUserDelete(w http.ResponseWriter, req *http.Request) {
	if *debugFlag {
		templates = RefreshTemplates(req)
	}

	// Check if logged in
	if !IsLoggedIn(w, req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	} else {
		// Check if user is admin
		if !WhoAmI(req).Admin {
			http.Redirect(w, req, "/", http.StatusSeeOther)
		} else {
			// Parse our form so we can get values from req.Form
			err = req.ParseForm()
			if err != nil {
				golem.Warnf("Error parsing form: %s", err)
			}

			// Parse accept from string to bool
			accept, err := strconv.ParseBool(req.Form["accept"][0])
			if err != nil {
				golem.Warnf("Error parsing accept from string to bool: %s", err)
			}

			if !accept {
				http.Redirect(w, req, "/users", http.StatusSeeOther)
			} else {
				// Get the user we're switching admin values for
				username := req.Form["username"][0]
				var id uint64 = 0
				for _, u := range users {
					if u.Username == username {
						id = u.Id
					}
				}

				// Loop through users to find user in memory
				for i, v := range users {
					if id == v.Id {
						// Delete user from memory
						users = append(users[:i], users[i+1:]...)
						// Delete user from database
						db.Unscoped().Table("users").Where("id = ?", id).Delete(&User{})
					}
				}

				// Redirect when we're done here
				http.Redirect(w, req, "/users", http.StatusSeeOther)
			}
		}
	}
}
