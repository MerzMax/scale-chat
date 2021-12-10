package main

import (
	"github.com/google/uuid"
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
var clients = make(map[uuid.UUID]Client, 0)

var broadcast = make(chan *MessageWrapper)
var unregisterClient = make(chan *Client)
var registerClient = make(chan *Client)

func main() {
	go broadcastMessages()
	go manageClients()

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

	client, _ := CreateClient(wsConn, make(chan *MessageWrapper), unregisterClient)
	clients[client.Id] = *client

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
				// TODO: Does not work .. Why???
				//select {
				//case client.outgoing <- wrapper:
				//default:
				//	close(client.outgoing)
				//	delete(clients, client.Id)
				//}
			}
		}
	}
}

// Listens for clients on the registerClient and the unregisterClient channels and registers or unregisters the
// transferred client
func manageClients() {
	for {
		select {
		case client := <-registerClient:
			log.Println("Registering new client with id: " + client.Id.String())
			clients[client.Id] = *client
		case client := <-unregisterClient:
			log.Println("Unregister client with id: " + client.Id.String())
			if _, ok := clients[client.Id]; ok {
				delete(clients, client.Id)
				close(client.outgoing)
				log.Println("Client with id: " + client.Id.String() + " was unregistered successfully")
			}
		}
	}
}
