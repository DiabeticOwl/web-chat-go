package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"web-chat-go/user"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type session struct {
	un           string
	lastActivity time.Time
}

var tpl *template.Template
var upgrader = websocket.Upgrader{
	// BufferSize determines how many bytes does the CPU handles each load,
	// the bigger the size the less the CPU is going to work as less loads
	// would need to be processed.
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// UserName, User
var dbUsers = make(map[string]user.User)

// Session ID, UserName
var dbSessions = make(map[string]session)

func init() {
	tpl = template.Must(template.ParseGlob("./templates/html/*.gohtml"))
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/signup/", signUp)
	http.HandleFunc("/login/", login)
	http.HandleFunc("/logout/", logout)
	http.HandleFunc("/receive/", receiveM)

	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./templates"))))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "index.gohtml", getUser(w, r))
}

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

		isSigned := user.IsSigned(user.SearchUser(un))

		if isSigned {
			http.Error(w,
				"The submitted username is already in use.",
				http.StatusForbidden)

			return
		}

		c := setCookie(w)
		dbSessions[c.Value] = session{
			un:           un,
			lastActivity: time.Now(),
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
		err = user.AddUser(u)
		if err != nil {
			http.Error(w,
				"Internal Server Error",
				http.StatusInternalServerError)

			panic(err)
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	tpl.ExecuteTemplate(w, "signup.gohtml", nil)
}

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

		u, err := user.SearchUser(un)
		if err != nil {
			http.Error(w,
				"Internal Server Error",
				http.StatusInternalServerError)

			panic(err)
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

func receiveM(w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
	u := getUser(w, r)

	for {
		// Read message from browser
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		// Print the message to the console
		fmt.Printf("%s sent: %s\n", u.UserName, string(msg))

		tNow := []byte(fmt.Sprint(time.Now().Format("2006-01-02 15:04:05"), "|"))
		msg = append(tNow, msg...)
		// TODO: Change "Sender" for "Receiver" according to who the
		// connection represents.
		msg = append(msg, "|Sender"...)

		// Write message back to browser
		if err = conn.WriteMessage(msgType, msg); err != nil {
			return
		}
	}
}
