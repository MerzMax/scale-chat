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

func wsHandler(writer http.ResponseWriter, req *http.Request) {
	// UPGRADE THE CONNECTION TO WS
	wsConnection, err := upgrader.Upgrade(writer, req, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer wsConnection.Close()

	// EVENT LOOP
	for {
		messageType, message, err := wsConnection.ReadMessage()
		if err != nil {
			log.Println("Error during reading message: ", err)
			break
		}

		log.Println()
		log.Printf("message: %s", message)
		log.Printf("messageType: %d", messageType)
		log.Println()

		err = wsConnection.WriteMessage(messageType, []byte("Hello you"))
		if err != nil {
			log.Println("Error during sending message: ", err)
			break
		}
	}
}

func hello(writer http.ResponseWriter, req *http.Request) {
	log.Println("/hello endpoint requested")
	writer.Write([]byte("Hello World!"))
	return
}
