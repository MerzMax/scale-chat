package main

import (
	"github.com/gorilla/websocket"
	"log"
	"scale-chat/chat"
	"sync"
)

// messageBufferSize is the buffer size of the incoming and outgoing message channels
const messageBufferSize = 100

// chatHistory is the history of all chat messages
var chatHistory = make([]*chat.Message, 0)

// clients that are connected to the server
var clients = make([]*Client, 0)

// incoming messages are sent through this channel
var incoming = make(chan *MessageWrapper, messageBufferSize)

// StartClient starts a client's incoming and outgoing message handlers
// and waits until the connection breaks to remove the client
func StartClient(wsConn *websocket.Conn, room string) {
	outgoing := make(chan *MessageWrapper, messageBufferSize)

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)

	client := Client{
		wsConn:    wsConn,
		outgoing:  outgoing,
		waitGroup: &waitGroup,
		room:      room,
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

// BroadcastMessages listens for messages on the incoming channel and sends them to all connected clients
func BroadcastMessages(enableDistribution bool, outgoing chan<- *chat.Message) {
	for wrapper := range incoming {
		chatHistory = append(chatHistory, wrapper.message)

		if enableDistribution && wrapper.source != DISTRIBUTOR {
			outgoing <- wrapper.message
		}

		for _, client := range clients {
			if wrapper.message.Room != client.room {
				continue
			}

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
