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

type Client struct {
	User *user.User
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
}

type ClientTCP struct {
	User *user.User
	Conn net.Conn
}

// clientClose will unregister the given Client and close it's connection
// to the web application.
func (c *Client) clientClose() {
	c.Hub.unregister <- c
	c.Conn.Close()
}

// readMessages defers the closure of the Client and enables an
// implementation of a Reader logic that will read each message sent from the
// web application through the WebSocket. The read message will be broadcasted
// to the entire Hub's collection of clients.
func (c *Client) readMessages() {
	defer c.clientClose()

	for {
		// Read message from browser
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err, websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				fmt.Println("The websocket closed unexpectedly.")
				break
			} else if websocket.IsCloseError(
				err, websocket.CloseGoingAway,
			) {
				fmt.Printf("User %v has logged out.\n", c.User.UserName)
				break
			}
			panic(err)
		}

		tNow := []byte(fmt.Sprint(time.Now().Format("2006-01-02 15:04:05"), "|"))
		msg = append(tNow, msg...)

		msg = append(msg, []byte(fmt.Sprintf("|%v", c.User.UserName))...)

		c.Hub.Broadcast <- msg
	}
}

// writeMessages defers the closure of the Client's connection to the web
// application and enables an implementation of a Writer logic that will
// extract all messages that the "Send" channel in the Client instance has.
func (c *Client) writeMessages() {
	defer c.Conn.Close()

	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				panic(err)
				// return
			}

			w.Write(msg)

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

	client.Hub.register <- client

	fmt.Printf("User %v has logged in.\n", client.User.UserName)

	go client.writeMessages()
	go client.readMessages()
}
