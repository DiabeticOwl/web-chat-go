package main

import (
	"fmt"
	"net/http"
	"time"
	"web-chat-go/hub"
	"web-chat-go/user"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// serveWs will instantiate a new "hub.Client" pointer and pass it as an
// argument in "hub.ServeClientWs".
func serveWs(w http.ResponseWriter, r *http.Request) {
	u := getUser(w, r)

	clientWS := &hub.ClientWS{
		User: &u,
		Hub:  clientsHub,
		Send: make(chan hub.ClientMessage),
	}

	hub.ServeClientWs(clientsHub, clientWS, w, r)
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
		return
	}

	tpl.ExecuteTemplate(w, "index.gohtml", getUser(w, r))
}

// signUp will execute the given template and save the data that the user
// posted through the form displayed, encrypting the password with an added
// salt and the "bcrypt" package. signUp will redirect to "/" if
// "alreadyLoggedIn" returns true.
func signUp(w http.ResponseWriter, r *http.Request) {
	if alreadyLoggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		un := r.FormValue("username")
		pw := r.FormValue("password")
		fn := r.FormValue("firstname")
		ln := r.FormValue("lastname")

		// More validations would be appropriate.
		if un == "" || pw == "" || fn == "" || ln == "" {
			http.Error(w,
				"Please fill all fields before proceeding through the SignUp.",
				http.StatusBadRequest)

			return
		}

		isSigned := user.IsSigned(user.SearchUser(dbConn, un))

		if isSigned {
			http.Error(w,
				"The submitted username is already in use.",
				http.StatusForbidden)

			return
		}

		saltPass := uuid.NewString()
		// Encrypting password with bcrypt.
		sb, err := bcrypt.GenerateFromPassword(
			[]byte(saltPass+pw),
			bcrypt.DefaultCost,
		)
		if err != nil {
			http.Error(w,
				"Internal Server Error",
				http.StatusInternalServerError)

			// Put panic instead of return since this error might not be very
			// clear so panic will help more in debugging.
			panic(err)
		}

		u := user.User{
			UserName:  un,
			SaltPass:  saltPass,
			Password:  sb,
			FirstName: fn,
			LastName:  ln,
		}
		err = user.AddUser(dbConn, u)
		if err != nil {
			http.Error(w,
				"Internal Server Error",
				http.StatusInternalServerError)

			panic(err)
		}

		c := setCookie(w)
		dbSessions[c.Value] = session{
			un:           un,
			lastActivity: time.Now(),
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	tpl.ExecuteTemplate(w, "signup.gohtml", nil)
}

// login will execute the given template and receive the information passed
// from the form displayed and check if it is in our database of users.
// If it is, the function will create a session and assign it to that user.
// If it is not, then the user will be redirected to "/".
func login(w http.ResponseWriter, r *http.Request) {
	if alreadyLoggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		un := r.FormValue("username")
		pw := r.FormValue("password")

		// More validations would be appropriate.
		if un == "" || pw == "" {
			http.Error(w,
				"Please fill all fields before proceeding through the LogIn.",
				http.StatusBadRequest)

			return
		}

		u, err := user.SearchUser(dbConn, un)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w,
				"Incorrect Username or Password.",
				http.StatusForbidden)
			return
		}

		err = bcrypt.CompareHashAndPassword(
			u.Password,
			[]byte(u.SaltPass+pw),
		)
		if err != nil {
			http.Error(w,
				"Incorrect Username or Password.",
				http.StatusForbidden)
			return
		}

		c := setCookie(w)
		dbSessions[c.Value] = session{
			un:           un,
			lastActivity: time.Now(),
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tpl.ExecuteTemplate(w, "login.gohtml", nil)
}

// logout will check that the user is logged in and remove their session and
// record from "dbSessions". If the user isn't logged in or the cookie isn't
// found it will redirect to "/".
func logout(w http.ResponseWriter, r *http.Request) {
	if !alreadyLoggedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// err is thrown because it is already checked in "alreadyLoggedIn".
	c, _ := r.Cookie("session")

	c.MaxAge = -1
	c.Path = "/"
	http.SetCookie(w, c)

	delete(dbSessions, c.Value)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
