package main

import (
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"scale-chat/chat"
)

// WebSocket connection configuration
var upgrader = websocket.Upgrader{
	ReadBufferSize:  128,
	WriteBufferSize: 128,
}

var chatHistory = make([]*chat.Message, 0)
var clients = make([]*Client, 0)

var broadcast = make(chan *MessageWrapper)
var unregister = make(chan *Client)

func main() {
	go broadcastMessages()

	// Register separate ServeMux instances for public endpoints and internal metrics
	publicMux := http.NewServeMux()
	internalMux := http.NewServeMux()

	// Register public endpoints
	publicMux.HandleFunc("/", demoHandler)
	publicMux.HandleFunc("/ws", wsHandler)

	// Register Prometheus endpoint
	internalMux.Handle("/metrics", promhttp.Handler())

	// Initiate Prometheus monitoring
	InitMonitoring()

	// Listen on internal metrics port
	go func() {
		err := http.ListenAndServe(":8081", internalMux)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Listen on public endpoint port
	err := http.ListenAndServe(":8080", publicMux)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Server Started and running!")
}

// Event handler for the /ws endpoint
func wsHandler(writer http.ResponseWriter, req *http.Request) {

	// Upgrade the http connection to ws
	wsConn, err := upgrader.Upgrade(writer, req, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}

	outgoing := make(chan *MessageWrapper) // TODO: Add buffer?

	client := {wsConn: wsConn, outgoing: outgoing, unregister: unregister}
	clients = append(clients, &client)

	go client.HandleOutgoing()
	go client.HandleIncoming(broadcast)
}

// Handles the / endpoint and serves the demo html chat client
func demoHandler(writer http.ResponseWriter, req *http.Request) {
	log.Println("serving demo HTML")
	http.ServeFile(writer, req, "./demo.html")
}

// Listens for messages on the broadcast channel and sends them to all connected clients
func broadcastMessages() {
	for {
		select {
		case wrapper := <-broadcast:
			chatHistory = append(chatHistory, wrapper.message)
			for _, client := range clients {
				client.outgoing <- wrapper
			}
		}
	}
}

func unregisterClients() {
	for {
		select {
		case client := <- unregister:
			clients
		}
	}
}