// Web Chat in Go is a simple chat built with the websocket protocol (rfc6455)
// on the back end and a free bootstrap template on the front end.
package main

import "net/http"

func main() {
	// Closes the connection to the database when the application ends.
	defer dbConn.Close()

	go clientsHub.RunChat()
	go clientsHub.RunTCPServer(":6893", dbConn)

	http.HandleFunc("/", index)
	http.HandleFunc("/signup/", signUp)
	http.HandleFunc("/login/", login)
	http.HandleFunc("/logout/", logout)
	http.HandleFunc("/ws/", serveWs)

	http.Handle("/assets/", http.StripPrefix("/assets",
		http.FileServer(http.Dir("./assets"))))

	// Port 8080 is used for debugging purposes.
	// In deployment port 80 is recommended.
	port := ":8080"
	if !debug {
		port = ":80"
	}
	err := http.ListenAndServe(port, nil)
	if err != nil {
		// http.ListenAndServe will panic if the port is in use or the if
		// application doesn't have the rights of using it.
		panic(err)
	}
}
