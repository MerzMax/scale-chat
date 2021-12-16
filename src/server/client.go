package main

import (
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"scale-chat/chat"
	"sync"
)

type Client struct {
	wsConn    *websocket.Conn
	outgoing  chan *MessageWrapper
	waitGroup *sync.WaitGroup
}

type MessageWrapper struct {
	message         *chat.Message
	processingTimer *prometheus.Timer
}

// chatHistory is the history of all chat messages
var chatHistory = make([]*chat.Message, 0)

// clients that are connected to the server
var clients = make([]*Client, 0)

// incoming messages are sent through this channel
var incoming = make(chan *MessageWrapper)

// StartClient starts a client's incoming and outgoing message handlers
// and waits until the connection breaks to remove the client
func StartClient(wsConn *websocket.Conn) {
	outgoing := make(chan *MessageWrapper, 100)

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)

	client := Client{
		wsConn:    wsConn,
		outgoing:  outgoing,
		waitGroup: &waitGroup,
	}
	clients = append(clients, &client)

	go client.HandleOutgoing()
	go client.HandleIncoming(incoming)

	// Wait for both handlers
	log.Println("Started a client")
	waitGroup.Wait()

	// Remove client from the list of active clients
	removeClient(&client)

	// Try to close websocket connection
	_ = closeWsConn(wsConn)

	log.Println("Client is gone")
}

//HandleOutgoing sends outgoing messages to the client's websocket connection
func (client *Client) HandleOutgoing() {
	defer func() {
		log.Println("Client's outgoing handler finished")
		client.waitGroup.Done()
	}()

	for wrapper := range client.outgoing {
		data, err := wrapper.message.ToJSON()
		if err != nil {
			continue
		}

		err = client.wsConn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("Cannot send message via WebSocket", err)
			return
		}

		wrapper.processingTimer.ObserveDuration()
		MessageCounterVec.WithLabelValues("outgoing").Inc()
	}
}

// HandleIncoming reads new messages from the websocket connection
// and broadcasts them to the other clients
func (client *Client) HandleIncoming(incoming chan<- *MessageWrapper) {
	defer func() {
		log.Println("Client's incoming handler finished")
		client.waitGroup.Done()
	}()

	for {
		_, data, err := client.wsConn.ReadMessage()
		if err != nil {
			log.Println("Cannot read message on websocket connection:", err)
			return
		}

		timer := prometheus.NewTimer(MessageProcessingTime)

		MessageCounterVec.WithLabelValues("incoming").Inc()

		log.Printf("Received raw message: %s", data)

		message, err := chat.ParseMessage(data)
		if err != nil {
			continue
		}

		wrapper := MessageWrapper{message: &message, processingTimer: timer}

		incoming <- &wrapper
	}
}

// BroadcastMessages listens for messages on the incoming channel and sends them to all connected clients
func BroadcastMessages() {
	for wrapper := range incoming {
		chatHistory = append(chatHistory, wrapper.message)

		for _, client := range clients {
			// By providing a default case, we avoid blocking the main broadcasting loop
			// in case the buffer of the outgoing channel is full.
			select {
			case client.outgoing <- wrapper:
			default:
				log.Println("Client's outgoing channel is full, skipping the message")
			}
		}
	}
}

// removeClient filters through the slice of active clients and removes the supplied reference.
func removeClient(client *Client) {
	log.Println("Removing client from list of active clients")
	filteredClients := make([]*Client, 0)
	for _, c := range clients {
		if c != client {
			filteredClients = append(filteredClients, c)
		}
	}
	clients = filteredClients
}

// closeWsConn tries to close the websocket connection
func closeWsConn(wsConn *websocket.Conn) error {
	log.Println("Trying to close websocket connection")
	err := wsConn.Close()
	if err != nil {
		log.Println("Failed to close websocket connection gracefully")
		return err
	}
	return nil
}
