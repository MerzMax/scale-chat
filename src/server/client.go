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
	room      string
}

type MessageWrapper struct {
	message         *chat.Message
	processingTimer *prometheus.Timer
	room            string
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

		wrapper := MessageWrapper{message: &message, processingTimer: timer, room: client.room}

		incoming <- &wrapper
	}
}
