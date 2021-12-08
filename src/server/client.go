package main

import (
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"scale-chat/chat"
)

type Client struct {
	wsConn   *websocket.Conn
	outgoing chan *MessageWrapper
}

type MessageWrapper struct {
	message         *chat.Message
	processingTimer *prometheus.Timer
}

func (client *Client) HandleOutgoing() {
	defer func() {
		err := client.wsConn.Close()
		if err != nil {
			log.Println("Cannot close WebSocket", err)
			return
		}
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
				continue
			}

			wrapper.processingTimer.ObserveDuration()
			MessageCounterVec.WithLabelValues("outgoing").Inc()
		}
	}
}

func (client *Client) HandleIncoming(broadcast chan *MessageWrapper) {
	defer func() {
		err := client.wsConn.Close()
		if err != nil {
			log.Println("Cannot close WebSocket", err)
			return
		}
	}()

	for {
		_, data, err := client.wsConn.ReadMessage()
		if err != nil {
			log.Println("Error during reading message:", err)
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

		broadcast <- &wrapper
	}
}
