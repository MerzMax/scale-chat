package client

import (
	"bufio"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"scale-chat/chat"
	"strings"
	"time"
)

var consoleReader = bufio.NewReader(os.Stdin)

type Client struct {
	wsConnection     websocket.Conn
	id               string
	CloseConnection  chan os.Signal
	ServerUrl        string
	IsLoadTestClient bool
	MsgSize          int
	MsgFrequency     int
	MsgEvents        chan *MessageEventEntry
}

func (client *Client) Start() error {
	if client.IsLoadTestClient {
		client.id = uuid.New().String()
	} else {
		log.Printf("Client started in loadtest mode. Please input your id: ")
		input, err := consoleReader.ReadString('\n')
		// convert CRLF to LF
		client.id = strings.Replace(input, "\n", "", -1)
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
				log.Println("Error while closing the ws connection gracefully: ", err)
				return err
			}

			// Timeout for connection close
			select {
			case <-time.After(time.Second):
				log.Println("Terminating. Timeout for connection close.")
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

		receivedAt := time.Now()

		message, err := chat.ParseMessage(data)
		if err != nil {
			continue
		}

		// If the loadtest mode is activated, there will be added a new message event with the metadata of this message.
		if client.IsLoadtestClient {
			var msgEventEntry = MessageEventEntry{
				ClientId:  client.id,
				SenderId: message.Sender,
				MessageId: message.MessageID,
				TimeStamp: receivedAt,
				Type: Received,
			}
			client.MsgEvents <- &msgEventEntry
			return
		}

		log.Printf("%v", message)
	}
}

// Handles outgoing ws messages
func sendHandler(client *Client) {
	defer client.wsConnection.Close()

	// Each message has an id to be able to follow the message in the message flow.
	var messageId uint64 = 1

	for {
		var text string
		if client.IsLoadTestClient {
			time.Sleep(time.Duration(client.MsgFrequency) * time.Millisecond)
			// The string "a" is exactly one byte. When we want to send a message with a specific byte size we can
			// repeat the string to reach the message size we want to have-
			text = strings.Repeat("a", client.MsgSize)
		} else {
			log.Printf("Please input the message you want to send:")
			input, err := consoleReader.ReadString('\n')
			// convert CRLF to LF
			text = strings.Replace(input, "\n", "", -1)
			if err != nil {
				log.Printf("Failed to read the input. Try again...")
				continue
			}
		}

		message := chat.Message{
			MessageId: messageId,
			Text:      text,
			Sender:    client.id,
			SentAt:    time.Now(),
		}

		data, err := message.ToJSON()
		if err != nil {
			continue
		}

		ts := time.Now()

		err = client.wsConnection.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("Error while sending message:", err)
			return
		}

		// If the loadtest mode is activated, there will be added a new message event with the metadata of this message.
		if client.IsLoadTestClient {
			var msgEventEntry = MessageEventEntry{
				ClientId:  client.id,
				SenderId: client.id,
				MessageId: message.MessageID,
				TimeStamp: ts,
				Type: Sent,
			}
			client.MsgEvents <- &msgEventEntry
		}
		messageId++
	}
}
