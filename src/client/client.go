package client

import (
	"github.com/gorilla/websocket"
	"log"
	"os"
	"scale-chat/chat"
	"time"
)

type Client struct {
	ServerUrl    string
	SysInterrupt chan os.Signal
}

func (client *Client) Start() error{
	// Connection Establishment
	serverUrl := "ws://localhost:8080" + "/ws"
	wsConnection, _, err := websocket.DefaultDialer.Dial(serverUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer wsConnection.Close()

	// Start Goroutine that listens on incoming messages
	go receiveHandler(wsConnection)

	// Start Goroutine that sends a message every second
	go sendHandler(wsConnection)

	for {
		select {
		// Listening for system interrupt
		case <-client.SysInterrupt:
			log.Println("System interrupt")

			// Closing the connection gracefully
			err := wsConnection.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error while closing the ws connection: ", err)
				return err
			}

			// Timeout for connection close
			select {
			case <-time.After(time.Second):
			}
			return nil
		}
	}
}

// The method handles incoming ws messages
func receiveHandler(wsConnection *websocket.Conn) {
	for {
		_, data, err := wsConnection.ReadMessage()
		if err != nil {
			log.Println("Error while receiving message:", err)
			return
		}

		message, err := chat.ParseMessage(data)
		if err != nil {
			continue
		}

		log.Printf("%v", message)
	}
}

func sendHandler(wsConnection *websocket.Conn) {
	for {
		message := chat.Message{
			Text:   "Hello",
			Sender: "Max",
			SentAt: time.Now(),
		}

		data, err := message.ToJSON()
		if err != nil {
			continue
		}

		err = wsConnection.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("Error while sending message:", err)
			return
		}
		//log.Printf("Me: %s", message)
		time.Sleep(time.Second)
	}
}