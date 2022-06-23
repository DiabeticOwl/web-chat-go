// Hub is a package that describes the Hubs and clients functionalities and
// structure for the "web-chat-go" usage.
package hub

import (
	"fmt"
	"sync"
)

type Hub struct {
	// Map of WebSocket clients that will be used to broadcast messages to the
	// rest of its own type.
	// This collection is used in order to harness the Go's built-in function
	// called "delete".
	clientsWS  map[*Client]bool
	clientsTCP map[*ClientTCP]bool
	// A channel of []byte type that will be used to broadcast messages to
	// the rest of Clients.
	Broadcast chan ClientMessage
	// A channel of *Client type will be used for the registration of
	// a Client.
	registerWS  chan *Client
	RegisterTCP chan *ClientTCP
	// A channel of *Client type will be used for the unregistration of
	// a client.
	unregisterWS  chan *Client
	UnregisterTCP chan *ClientTCP
}

func NewHub() *Hub {
	return &Hub{
		clientsWS:     make(map[*Client]bool),
		clientsTCP:    make(map[*ClientTCP]bool),
		Broadcast:     make(chan ClientMessage),
		registerWS:    make(chan *Client),
		RegisterTCP:   make(chan *ClientTCP),
		unregisterWS:  make(chan *Client),
		UnregisterTCP: make(chan *ClientTCP),
	}
}

// closeClient cleans the hub's collection of Clients and closes the given
// Client's Send channel.
func (h *Hub) closeClient(client *Client) {
	delete(h.clientsWS, client)
	close(client.Send)
}

func (h *Hub) Run() {
	for {
		select {
		// Each Client's registration.
		case client := <-h.registerWS:
			h.clientsWS[client] = true
		// Each ClientTCP's registration.
		case client := <-h.RegisterTCP:
			h.clientsTCP[client] = true
		// Each Client's unregistration.
		case client := <-h.unregisterWS:
			if _, ok := h.clientsWS[client]; ok {
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
		case message := <-h.Broadcast:
			var wg = &sync.WaitGroup{}

			wg.Add(2)

			go func() {
				defer wg.Done()

				for client := range h.clientsWS {
					select {
					// Sends the message extracted from the broadcast channel
					// to the Client's Send channel.
					case client.Send <- message:
					default:
						h.closeClient(client)
					}
				}
			}()

			go func() {
				defer wg.Done()

				for client := range h.clientsTCP {
					if client.User.UserName == message.User.UserName {
						continue
					}

					msg := fmt.Sprintf(
						"%s %s > %s",
						message.Time,
						message.User.UserName,
						message.MsgBody,
					)

					fmt.Fprintln(client.Conn, msg)
				}
			}()
		}
	}
}
