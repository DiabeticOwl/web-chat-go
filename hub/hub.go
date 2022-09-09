// Hub is a package that describes the Hubs and clients functionalities and
// structure for the "web-chat-go" usage.
package hub

import (
	"bufio"
	"database/sql"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
	"web-chat-go/user"

	"golang.org/x/crypto/bcrypt"
)

type Hub struct {
	// Map of WebSocket clients that will be used to broadcast messages to the
	// rest of its own type.
	// This collection is used in order to harness the Go's built-in function
	// called "delete".
	clientsWS  map[*ClientWS]bool
	clientsTCP map[*ClientTCP]bool
	// A channel of []byte type that will be used to broadcast messages to
	// the rest of Clients.
	Broadcast chan ClientMessage
	// A channel of *Client type will be used for the registration of
	// a Client.
	RegisterWS  chan *ClientWS
	RegisterTCP chan *ClientTCP
	// A channel of *Client type will be used for the unregistration of
	// a client.
	UnregisterWS  chan *ClientWS
	UnregisterTCP chan *ClientTCP
}

func NewHub() *Hub {
	return &Hub{
		clientsWS:     make(map[*ClientWS]bool),
		clientsTCP:    make(map[*ClientTCP]bool),
		Broadcast:     make(chan ClientMessage),
		RegisterWS:    make(chan *ClientWS),
		RegisterTCP:   make(chan *ClientTCP),
		UnregisterWS:  make(chan *ClientWS),
		UnregisterTCP: make(chan *ClientTCP),
	}
}

// closeClient cleans the hub's collection of Clients and closes the given
// Client's Send channel.
func (h *Hub) closeClient(client *ClientWS) {
	delete(h.clientsWS, client)

	client.Conn.Close()
	close(client.Send)
}

func (h *Hub) RunChat() {
	for {
		select {
		// Each Client's registration.
		case client := <-h.RegisterWS:
			h.clientsWS[client] = true
		// Each ClientTCP's registration.
		case client := <-h.RegisterTCP:
			h.clientsTCP[client] = true
		// Each Client's unregistration.
		case client := <-h.UnregisterWS:
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

			wg.Wait()
		}
	}
}

// handleTCPConnection takes the accepted connection and asks for an
// identification from the user. After a scanner brought by the
// "bufio" package reads each line inputted by the user and prints
// them to the server and any other connection instantiated hub,
// handleTCPConnection will close the connection.
func (h *Hub) handleTCPConnection(conn net.Conn, dbConn *sql.DB) {
	defer conn.Close()

	fmt.Fprintf(conn, "Please identify yourself: ")
	scanner := bufio.NewScanner(conn)
	clientTCP := ClientTCP{
		Conn:      conn,
		UserColor: randomHexColor(),
	}

	scanner.Scan()

	// Identity of the user.
	un := scanner.Text()
	u, err := user.SearchUser(dbConn, un)
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

		clientTCP.User = &u
	} else {
		// else needed as the termination of this function will terminate
		// the WebSocket connection.
		fmt.Fprintf(conn, "\nWelcome in %v.\n", un)

		u.UserName = un
		clientTCP.User = &u
	}

	fmt.Printf("User %v has logged in.\n", clientTCP.User.UserName)

	// Assigns the client's connection and registers it.
	clientTCP.Conn = conn
	h.RegisterTCP <- &clientTCP

	// Sets up every posterior message inputted by the user and broadcasts
	// it through the hub.
	for scanner.Scan() {
		message := ClientMessage{
			Time:     time.Now().Format("2006-01-02 15:04:05"),
			MsgBody:  scanner.Bytes(),
			MsgColor: clientTCP.UserColor,
			User:     clientTCP.User,
		}

		h.Broadcast <- message
	}

	h.UnregisterTCP <- &clientTCP
	fmt.Printf("User %v has logged out.\n", clientTCP.User.UserName)
}

// RunTCPServer will open a listener in the given address and launch accepted
// connections as goroutines for them to be taken care of.
func (h *Hub) RunTCPServer(address string, dbConn *sql.DB) {
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

		go h.handleTCPConnection(conn, dbConn)
	}
}

func randomHexColor() string {
	digits := strings.Split("0 1 2 3 4 5 6 7 8 9 a b c d e f", " ")
	hexCode := "#"

	for len(hexCode) < 7 {
		rInt := rand.Intn(len(digits))

		hexCode += digits[rInt]
	}

	return hexCode
}
