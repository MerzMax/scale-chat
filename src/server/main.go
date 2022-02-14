package main

import (
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net"
	"net/http"
	"os"
	"scale-chat/chat"
	"strconv"
)

// WebSocket connection configuration
var upgrader = websocket.Upgrader{
	ReadBufferSize:  128,
	WriteBufferSize: 128,
}

func main() {
	// Load env variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Couldn't load .env file.")
	} else {
		log.Println("Loaded a configuration via .env.")
	}

	enableDist := false
	envEnableDist := os.Getenv("ENABLE_DIST")
	if envEnableDist != "" {
		enableDist, err = strconv.ParseBool(envEnableDist)
		if err != nil {
			log.Println("Distributor will be disabled")
		}
		log.Println("Distributor will be enabled")
	} else {
		log.Println("Could not read ENABLE_DIST env variable")
		log.Println("Distributor will be disabled.")
	}

	var distributeIncoming chan *chat.Message
	var distributeOutgoing chan *chat.Message
	if enableDist {
		serverId := uuid.New().String()
		log.Println("ServerId for distribution: ", serverId)

		distributeIncoming = make(chan *chat.Message)
		distributeOutgoing = make(chan *chat.Message)
		distr := Distributor{
			Server:         os.Getenv("DIST_SERVER"),
			ServerPassword: os.Getenv("DIST_SERVER_PASSWORD"),
			Topic:          os.Getenv("DIST_TOPIC"),
			Incoming:       distributeIncoming,
			Outgoing:       distributeOutgoing,
		}

		err = distr.Ping()
		if err != nil {
			log.Panicln("Couldn't connect to the distributor. Pinging failed", err)
		}

		go distr.Subscribe(serverId)
		go distr.Publish(serverId)
	}

	go BroadcastMessages(enableDist, distributeOutgoing)

	// Register separate ServeMux instances for public endpoints and internal metrics
	publicMux := mux.NewRouter()
	internalMux := http.NewServeMux()

	// Register public endpoints
	publicMux.HandleFunc("/", demoHandler)
	publicMux.HandleFunc("/ws", wsHandler)
	publicMux.HandleFunc("/ws/{room}", wsHandler)

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
	room := vars["room"]

	wsConn, err := upgrader.Upgrade(writer, req, nil)
	if err != nil {
		log.Print("Cannot upgrade to websocket connection:", err)
		return
	}

	StartClient(wsConn, room)
}

// Handles the / endpoint and serves the demo html chat client
func demoHandler(writer http.ResponseWriter, req *http.Request) {
	log.Println("serving demo HTML")
	http.ServeFile(writer, req, "./demo.html")
}
