package chat

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"time"
)

type Message struct {
	Text            string    `json:"text"`
	Sender          string    `json:"sender"`
	SentAt          time.Time `json:"sent_at"`
	ProcessingTimer *prometheus.Timer
}

func ParseMessage(data []byte) (Message, error) {
	var message Message
	err := json.Unmarshal(data, &message)
	if err != nil {
		log.Printf("Cannot parse message: %v", err)
		return message, err
	}
	return message, nil
}

func (message *Message) ToJSON() ([]byte, error) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Cannot marshal message: %v", err)
		return data, err
	}
	return data, nil
}
