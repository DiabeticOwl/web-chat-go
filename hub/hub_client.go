package hub

import (
	"fmt"
	"net"
	"net/http"
	"time"
	"web-chat-go/user"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client type is a struct that will describe a websocket client.
// A client is a registered user in the application's hub that
// is able to send messages to others of its own type.
type Client struct {
	User *user.User
	Hub  *Hub
	Conn *websocket.Conn
	Send chan ClientMessage
}

// ClientTCP type is a struct that will describe a TCP client.
type ClientTCP struct {
	User *user.User
	Conn net.Conn
}

// ClientMessage type is a struct that will describe a message to be
// sent to the entire collection of Clients in the application's hub.
type ClientMessage struct {
	Time    string
	MsgBody []byte
	User    *user.User
}

// readMessages defers the closure of the Client and enables an
// implementation of a Reader logic that will read each message sent from the
// web application through the WebSocket. The read message will be broadcasted
// to the entire Hub's collection of clients.
func (c *Client) readMessages() {
	// This will unregister the given Client and close it's connection
	// to the web application.
	defer c.Hub.closeClient(c)

	for {
		// Read message from browser
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			// This clause means: if err is not any of the following, panic.
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
				websocket.CloseNormalClosure,
			) {
				panic(err)
			}

			fmt.Printf("User %v has logged out.\n", c.User.UserName)
			break
		}

		message := ClientMessage{
			Time:    time.Now().Format("2006-01-02 15:04:05"),
			MsgBody: msg,
			User:    c.User,
		}

		c.Hub.Broadcast <- message
	}
}

// writeMessages defers the closure of the Client's connection to the web
// application and enables an implementation of a Writer logic that will
// extract all messages that the "Send" channel in the Client instance has.
func (c *Client) writeMessages() {
	defer c.Conn.Close()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				panic(err)
				// return
			}

			// Formatting the message with the following pattern:
			// Time of the message | Message's body.
			msg := fmt.Sprintf("%s|%s", message.Time, message.MsgBody)
			w.Write([]byte(msg))

			if err := w.Close(); err != nil {
				panic(err)
				// return
			}
		}
	}
}

// ServeClientWs will upgrade the passed connection to a WebSocket protocol
// and register the passed Client to the hub instance. Later the client's
// "writeMessages" and "readMessages" methods will be launched to different
// goroutines.
func ServeClientWs(
	hub *Hub, client *Client,
	w http.ResponseWriter, r *http.Request,
) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
		// return
	}

	client.Conn = conn

	client.Hub.registerWS <- client

	fmt.Printf("User %v has logged in.\n", client.User.UserName)

	go client.writeMessages()
	go client.readMessages()
}
