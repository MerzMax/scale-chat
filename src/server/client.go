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

// History of all chat messages
var chatHistory = make([]*chat.Message, 0)

// Clients that are connected to the server
var clients = make([]*Client, 0)

// Channel that incoming messages are sent through
var incoming = make(chan *MessageWrapper)

func StartClient(wsConn *websocket.Conn) {
	outgoing := make(chan *MessageWrapper) // TODO: Add buffer?

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)

	client := Client{
		wsConn:    wsConn,
		outgoing:  outgoing,
		waitGroup: &waitGroup,
	}
	clients = append(clients, &client)

	log.Println("Starting outgoing and incoming handlers for client")
	go client.HandleOutgoing()
	go client.HandleIncoming(incoming)

	// Wait for both handlers
	log.Println("Waiting for client to close the connection")
	waitGroup.Wait()

	// Remove client from the list of active clients
	log.Println("Removing client from list of active clients")
	filteredClients := make([]*Client, 0)
	for _, c := range clients {
		if c != &client {
			filteredClients = append(filteredClients, c)
		}
	}
	clients = filteredClients

	// Try to close websocket connection
	log.Println("Trying to close websocket connection")
	err := wsConn.Close()
	if err != nil {
		log.Println("Failed to close websocket connection gracefully")
	}

	// Close outgoing client channel
	log.Println("Closing outgoing client channel")
	close(client.outgoing)
}

func (client *Client) HandleOutgoing() {
	defer func() {
		log.Println("Defer HandleOutgoing")
		client.waitGroup.Done()
	}()

	for {
		select {
		case wrapper := <-client.outgoing:
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
}

func (client *Client) HandleIncoming(incoming chan *MessageWrapper) {
	defer func() {
		log.Println("Defer HandleIncoming")
		client.waitGroup.Done()
	}()

	for {
		_, data, err := client.wsConn.ReadMessage()
		if err != nil {
			log.Println("Cannot read message on websocket connection:", err)
			break
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

// Listen for messages on the incoming channel and sends them to all connected clients
func broadcastMessages() {
	for {
		select {
		case wrapper := <-incoming:
			chatHistory = append(chatHistory, wrapper.message)
			for _, client := range clients {
				client.outgoing <- wrapper
			}
		}
	}
}
