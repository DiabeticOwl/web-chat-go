// Hub is a package that describes the Hubs and clients functionalities and
// structure for the "web-chat-go" usage.
package hub

import (
	"fmt"
	"strings"
)

type Hub struct {
	// Map of clients that will be used to broadcast messages to all of them.
	// This collection is used in order to harness the Go's built-in function
	// called "delete".
	clients    map[*Client]bool
	clientsTCP map[*ClientTCP]bool
	// A channel of []byte type that will be used to broadcast messages to
	// the rest of Clients.
	Broadcast chan []byte
	// A channel of *Client type will be used for the registration of
	// a Client.
	register    chan *Client
	RegisterTCP chan *ClientTCP
	// A channel of *Client type will be used for the unregistration of
	// a client.
	unregister    chan *Client
	UnregisterTCP chan *ClientTCP
}

func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		clientsTCP:    make(map[*ClientTCP]bool),
		Broadcast:     make(chan []byte),
		register:      make(chan *Client),
		RegisterTCP:   make(chan *ClientTCP),
		unregister:    make(chan *Client),
		UnregisterTCP: make(chan *ClientTCP),
	}
}

// closeClient cleans the hub's collection of Clients and closes the given
// Client's Send channel.
func (h *Hub) closeClient(client *Client) {
	delete(h.clients, client)
	close(client.Send)
}

func (h *Hub) Run() {
	for {
		select {
		// Each Client's registration.
		case client := <-h.register:
			h.clients[client] = true
			// Each ClientTCP's registration.
		case client := <-h.RegisterTCP:
			h.clientsTCP[client] = true
		// Each Client's unregistration.
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.closeClient(client)
			}
			// Each Client's unregistration.
		case client := <-h.UnregisterTCP:
			if _, ok := h.clientsTCP[client]; ok {
				delete(h.clientsTCP, client)
			}
		// Each Client's message to be broadcasted.
		// Each message is passed to two goroutines that will range hub's
		// maps of clients.
		case msg := <-h.Broadcast:
			// If the message comes from the
			go func() {
				for client := range h.clients {
					select {
					// Sends the message extracted from the broadcast channel
					// to the Client's Send channel.
					case client.Send <- msg:
					default:
						h.closeClient(client)
					}
				}
			}()

			for client := range h.clientsTCP {
				msgDet := strings.Split(string(msg), "|")
				msg := fmt.Sprintf(
					"%v - User %v says: %v",
					msgDet[0],
					msgDet[2],
					msgDet[1],
				)

				fmt.Fprintln(client.Conn, msg)
			}
		}
	}
}
