package main

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"scale-chat/chat"
)

type Client struct {
	Id         uuid.UUID
	wsConn     *websocket.Conn
	outgoing   chan *MessageWrapper
	unregister chan *Client
}

type MessageWrapper struct {
	message         *chat.Message
	processingTimer *prometheus.Timer
}

func CreateClient(wsConn *websocket.Conn, outgoing chan *MessageWrapper, unregister chan *Client) (*Client, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		log.Println("Cannot generate UUID")
		return nil, err
	}
	return &Client{
		Id:         id,
		wsConn:     wsConn,
		outgoing:   outgoing,
		unregister: unregister,
	}, nil
}

func (client *Client) HandleOutgoing() {
	defer func() {
		client.wsConn.Close()
	}()

	for {
		select {
		case wrapper, ok := <-client.outgoing:
			if !ok {
				log.Println("The outgoing channel was closed by the server.")
				client.wsConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := client.wsConn.WriteJSON(wrapper.message)
			if err != nil {
				log.Println("Cannot send message via WebSocket", err)
				return
			}

			wrapper.processingTimer.ObserveDuration()
			MessageCounterVec.WithLabelValues("outgoing").Inc()
		}
	}
}

func (client *Client) HandleIncoming(broadcast chan *MessageWrapper) {
	defer func() {
		unregisterClient <- client
		client.wsConn.Close()
	}()

	for {
		_, data, err := client.wsConn.ReadMessage()
		if err != nil {
			log.Println("Error during reading message:", err)
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

		broadcast <- &wrapper
	}
}
