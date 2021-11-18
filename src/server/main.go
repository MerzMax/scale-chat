package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"scale-chat/chat"
)

var upgrader = websocket.Upgrader{} // use default options
var chatHistory = make([]*chat.Message, 0)
var clients = make([]*Client, 0)

var broadcast = make(chan *chat.Message)

func main() {
	go broadcastMessages()

	http.HandleFunc("/hello", hello)
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/", demoHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// Event handler for the /ws endpoint
func wsHandler(writer http.ResponseWriter, req *http.Request) {

	// Allow requests from every origin
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// Upgrade the http connection to ws
	wsConn, err := upgrader.Upgrade(writer, req, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}

	outgoing := make(chan *chat.Message) // TODO: Add buffer?
	client := Client{wsConn: wsConn, outgoing: outgoing}
	clients = append(clients, &client)

	go client.HandleOutgoing()
	go client.HandleIncoming(broadcast)
}

// Event handler for the /hello endpoint
func hello(writer http.ResponseWriter, req *http.Request) {
	log.Println("/hello endpoint requested")
	writer.Write([]byte("Hello World!"))
}

func demoHandler(writer http.ResponseWriter, req *http.Request) {
	log.Println("/ endpoint requested")
	http.ServeFile(writer, req, "./demo.html")
}

// Method sends a message to the connected loadtest-client
func replyMessage(wsConn *websocket.Conn) {
	message := "Hello you! :)"
	err := wsConn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Println("Error during sending message:", err)
	}
	log.Printf("Me: %s", message)
}

func broadcastMessages() {
	for {
		select {
		case message := <-broadcast:
			chatHistory = append(chatHistory, message)
			for _, client := range clients {
				client.outgoing <- message
			}
		}
	}
}