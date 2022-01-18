package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net"
	"net/http"
)

// WebSocket connection configuration
var upgrader = websocket.Upgrader{
	ReadBufferSize:  128,
	WriteBufferSize: 128,
}

func main() {
	go BroadcastMessages()

	// Register separate ServeMux instances for public endpoints and internal metrics
	publicMux := mux.NewRouter()
	internalMux := http.NewServeMux()

	// Register public endpoints
	publicMux.HandleFunc("/", demoHandler)
	publicMux.HandleFunc("/ws", wsHandler)
	publicMux.HandleFunc("/ws/{chatId}", wsHandler)

	// Register Prometheus endpoint
	internalMux.Handle("/metrics", promhttp.Handler())

	// Initiate Prometheus monitoring
	InitMonitoring()

	// Listen on internal metrics port
	go func() {
		l, err := net.Listen("tcp", ":8081")
		if err != nil {
			log.Fatal("Could not listen on metrics port: ", err)
		}

		log.Println("Metrics server will be listening for incoming requests on port: 8081")

		if err := http.Serve(l, internalMux); err != nil {
			log.Fatal("Serving the metrics server failed:", err)
		}
	}()

	// Listen on public endpoint port
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Could not listen on chat server port: ", err)
	}

	log.Println("Chat server will be listening for incoming requests on port: 8080")

	if err := http.Serve(l, publicMux); err != nil {
		log.Fatal("Serving the chat server failed:", err)
	}
}

// Event handler for the /ws endpoint
func wsHandler(writer http.ResponseWriter, req *http.Request) {

	log.Println("Got new connection")

	vars := mux.Vars(req)
	chatId := vars["chatId"]

	log.Println(chatId)

	wsConn, err := upgrader.Upgrade(writer, req, nil)
	if err != nil {
		log.Print("Cannot upgrade to websocket connection:", err)
		return
	}

	StartClient(wsConn, chatId)
}

// Handles the / endpoint and serves the demo html chat client
func demoHandler(writer http.ResponseWriter, req *http.Request) {
	log.Println("serving demo HTML")
	http.ServeFile(writer, req, "./demo.html")
}
