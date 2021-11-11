package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{} // use default options

func main() {
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/ws", wsHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// Event handler for the /ws endpoint
func wsHandler(writer http.ResponseWriter, req *http.Request) {
	// Upgrade the http connection to ws
	wsConnection, err := upgrader.Upgrade(writer, req, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer wsConnection.Close()

	// Listening for incoming messages
	for {
		_, message, err := wsConnection.ReadMessage()
		if err != nil {
			log.Println("Error during reading message:", err)
			break
		}

		log.Printf("Received Message: %s", message)


	}
}

// Event handler for the /hello endpoint
func hello(writer http.ResponseWriter, req *http.Request) {
	log.Println("/hello endpoint requested")
	writer.Write([]byte("Hello World!"))
}

// Method sends a message to the connected client
func replyMessage(wsConnection *websocket.Conn) {
	message := "Hello you! :)"
	err := wsConnection.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Println("Error during sending message:", err)
	}
	log.Printf("Me: %s",  message)
}
