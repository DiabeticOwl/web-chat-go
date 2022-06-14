// Hug is a package that describes the Hub and client functionality and
// structure for the "web-chat-go" usage.
package hub

type Hub struct {
	// Map of clients that will be used to broadcast messages to all of them.
	// This collection is used in order to harness the Go's built-in function
	// called "delete".
	clients map[*Client]bool
	// A channel of []byte type that will be used to broadcast messages to
	// the rest of Clients.
	broadcast chan []byte
	// A channel of *Client type that will be used for the registration of
	// a Client.
	register chan *Client
	// A channel of *Client type that will be used for the unregistration of
	// a client.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
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
		// Each Client's unregistration.
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.closeClient(client)
			}
		// Each Client's message to be broadcasted.
		case msg := <-h.broadcast:
			for client := range h.clients {
				select {
				// Sends the message extracted from the broadcast channel to
				// the Client's Send channel.
				case client.Send <- msg:
				default:
					h.closeClient(client)
				}
			}
		}
	}
}
