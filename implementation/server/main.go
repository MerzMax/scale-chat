package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{} // use default options

func main() {
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/ws", upgradeToWs)
	err := http.ListenAndServe(":8080", nil)
	checkError(err)
}

func upgradeToWs (writer http.ResponseWriter, req *http.Request) {
	wsConnection, err := upgrader.Upgrade(writer, req, nil)
	checkError(err)
	defer wsConnection.Close()

	for {
		messageType, message, err := wsConnection.ReadMessage()
		checkError(err)

		log.Printf("message: %s", message)
		log.Printf("messageType: %d", messageType)

		err = wsConnection.WriteMessage(messageType, []byte("Hello you"))
		checkError(err)
	}


}

func hello(writer http.ResponseWriter, req *http.Request) {
	log.Println("/hello endpoint requested")
	writer.Write([]byte("Hello World!"))
	return
}

func checkError(err error){
	if err != nil {
		log.Fatal(err)
	}
}
