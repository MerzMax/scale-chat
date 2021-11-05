package main

import (
	"github.com/gorilla/websocket"
	"log"
	"time"
)

func main() {
	// Connection Establishment
	serverUrl := "ws://localhost:8080" + "/ws"
	wsConnection, _, err := websocket.DefaultDialer.Dial(serverUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer wsConnection.Close()

	// Start Goroutine that listens on incoming messages
	go receiveHandler(wsConnection)

	// Sends a message every second
	for{
		message := "Hello?"
		err = wsConnection.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("Error while sending message:", err)
			return
		}
		log.Println("Me: " +  message)
		time.Sleep(time.Second)
	}
}

// The method handles incoming ws messages
func receiveHandler(wsConnection *websocket.Conn){
	for {
		_, msg, err := wsConnection.ReadMessage()
		if err != nil {
			log.Println("Error while receiving message:", err)
			return
		}
		log.Printf("Received Message: %s\n", msg)
	}
}
