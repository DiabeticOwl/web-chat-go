package main

import (
	"net/http"
	"time"
	"web-chat-go/user"

	"github.com/google/uuid"
)

func setCookie(w http.ResponseWriter) *http.Cookie {
	c := &http.Cookie{
		Name:     "session",
		Value:    uuid.NewString(),
		Path:     "/",
		HttpOnly: true,
	}

	http.SetCookie(w, c)

	return c
}

// getUser returns the user found in the session.
// Returns a zero value of type user if none found.
func getUser(w http.ResponseWriter, r *http.Request) user.User {
	var u user.User

	c, err := r.Cookie("session")
	if err != nil {
		return u
	}

	if sess, ok := dbSessions[c.Value]; ok {
		u, err = user.SearchUser(sess.un)
		if err != nil {
			panic(err)
		}

		restartLength(w, c)
	}

	return u
}

// alreadyLoggedIn returns a bool indicating whether the user is already
// in session or not.
func alreadyLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	c, err := r.Cookie("session")
	if err != nil {
		return false
	}
	_, ok := dbSessions[c.Value]

	if ok {
		restartLength(w, c)
	}

	return ok
}

func restartLength(w http.ResponseWriter, c *http.Cookie) {
	sess := dbSessions[c.Value]
	sess.lastActivity = time.Now()

	dbSessions[c.Value] = sess

	// c.MaxAge = sessionLength
	c.Path = "/"

	http.SetCookie(w, c)
}
