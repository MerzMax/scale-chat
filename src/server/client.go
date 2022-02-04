package main

import (
	"encoding/json"
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
	message           *chat.Message
	processingTimer   *prometheus.Timer
	sourceDistributor bool
}

//HandleOutgoing sends outgoing messages to the client's websocket connection
func (client *Client) HandleOutgoing() {
	defer func() {
		log.Println("Client's outgoing handler finished")
		client.waitGroup.Done()
	}()

	for wrapper := range client.outgoing {
		data, err := wrapper.message.MarshalBinary()
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

		var message chat.Message
		err = message.UnmarshalBinary(data)
		if err != nil {
			continue
		}

		wrapper := MessageWrapper{message: &message, processingTimer: timer, sourceDistributor: false}

		incoming <- &wrapper
	}
}

// MarshalBinary marshal a MessageWrapper to a byte array
func (mw *MessageWrapper) MarshalBinary() ([]byte, error) {
	data, err := json.Marshal(mw)
	if err != nil {
		log.Printf("Cannot marshal message: %v", err)
		return data, err
	}
	return data, nil
}

// UnmarshalBinary unmarshal a byte array to a MessageWrapper
func (mw *MessageWrapper) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, mw); err != nil {
		return err
	}
	return nil
}
