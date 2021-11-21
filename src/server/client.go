package main

import (
	"github.com/gorilla/websocket"
	"log"
	"scale-chat/chat"
)

type Client struct {
	wsConn   *websocket.Conn
	outgoing chan *chat.Message
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
		case message := <-client.outgoing:
			data, err := message.ToJSON()
			if err != nil {
				continue
			}

			err = client.wsConn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("Cannot send message via WebSocket", err)
				continue
			}

			MessageCounterVec.WithLabelValues("outgoing").Inc()
		}
	}
}

func (client *Client) HandleIncoming(broadcast chan *chat.Message) {
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

		MessageCounterVec.WithLabelValues("incoming").Inc()

		log.Printf("Received raw message: %s", data)

		message, err := chat.ParseMessage(data)
		if err != nil {
			continue
		}

		broadcast <- &message
	}
}
