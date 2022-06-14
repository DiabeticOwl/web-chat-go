package hub

import (
	"fmt"
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

func (c *Client) clientClose() {
	c.Hub.unregister <- c
	c.Conn.Close()
}

func (c *Client) readMessages() {
	defer c.clientClose()

	for {
		// Read message from browser
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Println("The websocket closed unexpectedly.")
			}

			break
		}

		tNow := []byte(fmt.Sprint(time.Now().Format("2006-01-02 15:04:05"), "|"))
		msg = append(tNow, msg...)
		// TODO: Change "Sender" for "Receiver" according to who the
		// connection represents.
		msg = append(msg, "|Sender"...)

		c.Hub.broadcast <- msg
	}
}

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

	go client.writeMessages()
	go client.readMessages()
}
