package client

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"scale-chat/chat"
	"strings"
	"time"
)

type Client struct {
	wsConnection 	 	websocket.Conn
	id           	 	string
	CloseConnection  	chan string
	ServerUrl		 	string
	IsLoadtestClient	bool
	MsgSize			int
	MsgFrequency	int
}

func (client *Client) Start() error{
	if client.IsLoadtestClient{
		client.id = uuid.New().String()
	} else {
		log.Printf("Client started in loadtest mode. Please input your id: ")
		_, err := fmt.Scanln(&client.id)
		if err != nil || len(client.id) < 1 {
			log.Printf("Failed to read the name input. Using default id: MuM")
			client.id = "MuM"
		}
	}

	// Connection Establishment
	wsConnection, _, err := websocket.DefaultDialer.Dial(client.ServerUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer wsConnection.Close()

	client.wsConnection = *wsConnection

	// Start Goroutine that listens on incoming messages
	go receiveHandler(client)

	// Start Goroutine that sends a message every second
	go sendHandler(client)

	for {
		select {
		case s := <-client.CloseConnection:
			log.Printf("Closing connection... Reason: %s", s)

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

// Handles incoming ws messages
func receiveHandler(client *Client) {
	for {
		_, data, err := client.wsConnection.ReadMessage()
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

// Handles outgoing ws messages
func sendHandler(client *Client) {
	for {

		var text string
		if client.IsLoadtestClient {
			time.Sleep(time.Duration(client.MsgFrequency) * time.Millisecond)
			// The string "a" is exactly one byte. When we want to send a message with a specific byte size we can
			// repeat the string to reach the message size we want to have-
			text = strings.Repeat("a", client.MsgSize)
		} else {
			_, err := fmt.Scanln(&text)
			if err != nil {
				log.Printf("Failed to read the input. Try again...")
				continue
			}
		}

		message := chat.Message{
			Text:   text,
			Sender: client.id,
			SentAt: time.Now(),
		}

		data, err := message.ToJSON()
		if err != nil {
			continue
		}

		err = client.wsConnection.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("Error while sending message:", err)
			return
		}
	}
}