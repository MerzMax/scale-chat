package main

import (
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
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
	wsConn, err := upgrader.Upgrade(writer, req, nil)
	if err != nil {
		log.Print("Cannot upgrade to websocket connection:", err)
		return
	}

	StartClient(wsConn)
}

// Handles the / endpoint and serves the demo html chat client
func demoHandler(writer http.ResponseWriter, req *http.Request) {
	log.Println("serving demo HTML")
	http.ServeFile(writer, req, "./demo.html")
}
