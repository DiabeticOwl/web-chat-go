package hub

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
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

func (h *Hub) closeClient(client *Client) {
	delete(h.clients, client)
	close(client.Send)
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.closeClient(client)
			}
		case msg := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.Send <- msg:
				default:
					h.closeClient(client)
				}
			}
		}
	}
}
