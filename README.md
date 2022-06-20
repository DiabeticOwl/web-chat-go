# Web Chat in Go

Web Chat in Go is a simple chat built with the [websocket protocol](https://datatracker.ietf.org/doc/html/rfc6455) on the back end and a [free bootstrap template](https://www.bootdey.com/snippets/view/chat-app) on the front end.

The logic used is an implementation of the [websocket package](https://github.com/gorilla/websocket) in the Go programming language with the application of hubs and clients principles to connect the server with the user in the web page.

User data is stored on a local [PostgreSQL](https://www.postgresql.org/) database and, alongside the entire app, will be served in the cloud in the future.

Web Chat in Go also enables another port (6893) in which users can connect with the command "telnet" and join the group chat by identifying themselves as an already logged in user or as a guest user.

## Client

The client logic consists consists on a struct that contains information of the user connected to the application, the connection itself, the instance of the Hub it belongs and a channel that will be used to send the written message to the rest of clients in the program.

## Hub

The hub logic consists on a struct that contains a collection of clients, a channel that will be used to broadcast all the messages sent to it to the rest of clients and two channels, one for the registration of clients and the other for the unregistration of them.
