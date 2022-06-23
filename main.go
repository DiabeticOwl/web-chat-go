// Web Chat in Go is a simple chat built with the websocket protocol (rfc6455)
// on the back end and a free bootstrap template on the front end.
package main

import (
	"bufio"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"time"

	"web-chat-go/hub"
	"web-chat-go/user"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// session type will describe the time of the last activity of a given user.
type session struct {
	un           string
	lastActivity time.Time
}

var tpl *template.Template
var upgrader = websocket.Upgrader{
	// BufferSize determines how many bytes does the CPU handles each message
	// load sent to the websocket, the bigger the size the less the CPU is
	// going to work as less loads would need to be processed.
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// dbUsers is a map that will store the connected users information.
var dbUsers = make(map[string]user.User)

// dbSessions is a map that will store the information of the session
// in which each user is related to. The key will be a cookie.
var dbSessions = make(map[string]session)

// The single instance of "hub" is assigned.
var clientsHub = hub.NewHub()

var tcpConn = make(chan hub.Client)

func init() {
	tpl = template.Must(template.ParseGlob("./assets/html/*.gohtml"))
}

func main() {
	go clientsHub.Run()

	http.HandleFunc("/", index)
	http.HandleFunc("/signup/", signUp)
	http.HandleFunc("/login/", login)
	http.HandleFunc("/logout/", logout)
	http.HandleFunc("/ws/", serveWs)

	http.Handle("/assets/", http.StripPrefix("/assets",
		http.FileServer(http.Dir("./assets"))))

	go tcpServer(":6893")

	err := http.ListenAndServe(":80", nil)
	if err != nil {
		// http.ListenAndServe will panic if the port is in use or the if
		// application doesn't have the rights of using it.
		panic(err)
	}
}

// tcpServer will open a listener in the given address and launch accepted
// connections as goroutines for them to be taken care of.
func tcpServer(address string) {
	// A listener is opened in the port "6893" with the "tcp" network.
	li, err := net.Listen("tcp", address)
	if err != nil {
		// net.Listen will panic if the port is in use or the if
		// application doesn't have the rights of using it.
		panic(err)
	}
	defer li.Close()

	// Infinitely accepts for new connections and sends a goroutine for
	// each one that handles it.
	for {
		conn, err := li.Accept()
		if err != nil {
			panic(err)
		}

		go handleConnection(conn)
	}
}

// handleConnection takes the accepted connection and asks for an
// identification from the user. After a scanner brought by the
// "bufio" package reads each line inputted by the user and prints
// them to the server and any other connection instantiated hub,
// handleConnection will close the connection.
func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Fprintf(conn, "Please identify yourself: ")
	scanner := bufio.NewScanner(conn)
	client := hub.ClientTCP{
		Conn: conn,
	}

	scanner.Scan()

	// Identity of the user.
	un := scanner.Text()
	u, err := user.SearchUser(un)
	if err == nil {
		// TODO: Mask password entry.
		fmt.Fprintf(conn, "Please enter your password: ")

		scanner.Scan()

		err = bcrypt.CompareHashAndPassword(
			u.Password,
			[]byte(u.SaltPass+scanner.Text()),
		)
		if err != nil {
			// TODO: Make a loop for retrying.
			fmt.Fprintf(conn, "\nIncorrect password. Try again.")
			return
		}

		client.User = &u
	} else {
		// else needed as the termination of this function will terminate
		// the WebSocket connection.
		fmt.Fprintf(conn, "\nWelcome in %v.\n", un)

		u.UserName = un
		client.User = &u
	}

	fmt.Printf("User %v has logged in.\n", client.User.UserName)

	// Assigns the client's connection and registers it.
	client.Conn = conn
	clientsHub.RegisterTCP <- &client

	// Sets up every posterior message inputted by the user and broadcasts
	// it through the hub.
	for scanner.Scan() {
		message := hub.ClientMessage{
			Time:    time.Now().Format("2006-01-02 15:04:05"),
			MsgBody: scanner.Bytes(),
			User:    client.User,
		}

		clientsHub.Broadcast <- message
	}

	fmt.Printf("User %v has logged out.\n", client.User.UserName)
}

// serveWs will instantiate a new "hub.Client" pointer and pass it as an
// argument in "hub.ServeClientWs".
func serveWs(w http.ResponseWriter, r *http.Request) {
	u := getUser(w, r)

	client := &hub.Client{
		User: &u,
		Hub:  clientsHub,
		Send: make(chan hub.ClientMessage),
	}

	hub.ServeClientWs(clientsHub, client, w, r)
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

		isSigned := user.IsSigned(user.SearchUser(un))

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
		err = user.AddUser(u)
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

		u, err := user.SearchUser(un)
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
