// File that contains global variables for the application.
package main

import (
	"html/template"
	"time"
	"web-chat-go/hub"
	"web-chat-go/user"

	"github.com/gorilla/websocket"
)

// session type will describe the time of the last activity of a given user.
type session struct {
	un           string
	lastActivity time.Time
}

// debug describes whether the application is run on a debugging or a
// production environment.
var debug = true

var tpl = template.Must(template.ParseGlob("./assets/html/*.gohtml"))
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

var tcpConn = make(chan hub.ClientWS)

// The connection to the database is opened and assigned to dbConn.
var dbConn = connect()
