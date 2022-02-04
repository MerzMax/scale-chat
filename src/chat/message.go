package chat

import (
	"encoding/json"
	"log"
	"time"
)

type Message struct {
	MessageId uint64    `json:"message_id"`
	Text      string    `json:"text"`
	Sender    string    `json:"sender"`
	SentAt    time.Time `json:"sent_at"`
}

// Unmarshal a given byte array to a Message
func (msg *Message) Unmarshal(data []byte) error {
	err := json.Unmarshal(data, msg)
	if err != nil {
		log.Printf("Cannot parse message: %v", err)
		return err
	}
	return nil
}

// Marshal a given Message to a byte array
func (msg *Message) Marshal() ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Cannot marshal message: %v", err)
		return data, err
	}
	return data, nil
}
