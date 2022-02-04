package client

import (
	"bufio"
	"context"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"scale-chat/chat"
	"strings"
	"sync"
	"time"
)

var consoleReader = bufio.NewReader(os.Stdin)

type Client struct {
	Context          context.Context
	WaitGroup        *sync.WaitGroup
	wsConnection     websocket.Conn
	id               string
	CloseConnection  chan os.Signal
	ServerUrl        string
	IsLoadTestClient bool
	MsgSize          int
	MsgFrequency     int
	MsgEvents        chan<- *MessageEventEntry
	Room             string
}

func (client *Client) Start() error {
	defer client.WaitGroup.Done()

	if client.IsLoadTestClient {
		client.id = uuid.New().String()
	} else {
		log.Printf("Client started in loadtest mode.")
		log.Printf("Please input your id:")
		input, err := consoleReader.ReadString('\n')
		// convert CRLF to LF
		client.id = strings.Replace(input, "\n", "", -1)
		if err != nil || len(client.id) < 1 {
			log.Printf("Failed to read the name input. Using default name: MuM")
			client.id = "MuM"
		}

		log.Printf("Please input your chat room:")
		input, err = consoleReader.ReadString('\n')
		// convert CRLF to LF
		client.Room = strings.Replace(input, "\n", "", -1)
		if err != nil || len(client.Room) < 1 {
			log.Printf("Failed to read the chat room input. Using default id: test")
			client.Room = "test"
		}
	}

	// Connection Establishment
	wsConnection, _, err := websocket.DefaultDialer.Dial(client.ServerUrl+"/"+client.Room, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}

	client.wsConnection = *wsConnection

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(2)

	// Start Goroutine that listens on incoming messages
	receiveCtx, receiveCancelFunc := context.WithCancel(context.Background())
	go client.receiveHandler(receiveCtx, waitGroup)

	// Start Goroutine that sends a message every second
	sendCtx, sendCancelFunc := context.WithCancel(context.Background())
	go client.sendHandler(sendCtx, waitGroup)

	// Waiting for shutdown...
	<-client.Context.Done()

	log.Println("Client got interrupted, closing connection...")

	sendCancelFunc()
	receiveCancelFunc()

	waitGroup.Wait()

	// Closing the connection gracefully
	err = wsConnection.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("Error while closing the ws connection gracefully: ", err)
		return err
	}

	err = wsConnection.Close()
	if err != nil {
		log.Println("Cannot close websocket connection", err)
		return err
	}

	return nil
}

// Handles incoming ws messages
func (client *Client) receiveHandler(ctx context.Context, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	incomingMessages := make(chan *[]byte)

	// Convert blocking ReadMessage call into channel
	go func() {
		for {
			_, data, err := client.wsConnection.ReadMessage()
			if err != nil {
				close(incomingMessages)
				return
			}

			incomingMessages <- &data
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-incomingMessages:
			if !ok {
				log.Println("incomingMessages channel was closed")
				return
			}

			receivedAt := time.Now()

			var message chat.Message
			err := message.Unmarshal(*data)
			if err != nil {
				continue
			}

			// If the load test mode is activated, there will be added a new message event with the metadata of this message.
			if client.IsLoadTestClient {
				var msgEventEntry = MessageEventEntry{
					ClientId:  client.id,
					SenderId:  message.Sender,
					MessageId: message.MessageId,
					TimeStamp: receivedAt,
					Type:      Received,
				}
				client.MsgEvents <- &msgEventEntry
			}

			log.Printf("%v", message)
		}
	}
}

// Handles outgoing ws messages
func (client *Client) sendHandler(ctx context.Context, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	// Each message has an id to be able to follow the message in the message flow.
	var messageId uint64 = 1

	for {
		select {
		case <-ctx.Done():
			return
		default:
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

			data, err := message.Marshal()
			if err != nil {
				continue
			}

			ts := time.Now()

			err = client.wsConnection.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("Error while sending message:", err)
				return
			}

			// If the load test mode is activated, there will be added a new message event with the metadata of this message.
			if client.IsLoadTestClient {
				var msgEventEntry = MessageEventEntry{
					ClientId:  client.id,
					SenderId:  client.id,
					MessageId: message.MessageId,
					TimeStamp: ts,
					Type:      Sent,
				}
				client.MsgEvents <- &msgEventEntry
			}
			messageId++
		}
	}
}
