package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"scale-chat/chat"
	"time"
)

type Distributor struct {
	Server         string
	ServerPassword string
	Incoming       chan<- *chat.Message
	Outgoing       <-chan *chat.Message
	Topic          string
	client         redis.Client
	ctx            context.Context
}

type DistributionMessage struct {
	MessageId uint64    `json:"message_id"`
	Text      string    `json:"text"`
	Sender    string    `json:"sender"`
	SentAt    time.Time `json:"sent_at"`
	Room      string    `json:"room"`
	ServerId  string    `json:"server_id"`
}

// UnmarshalBinary a given byte array to a Message
func (distMsg *DistributionMessage) UnmarshalBinary(data []byte) error {
	err := json.Unmarshal(data, distMsg)
	if err != nil {
		log.Printf("Cannot parse message: %v", err)
		return err
	}
	return nil
}

// MarshalBinary a given Message to a byte array
func (distMsg *DistributionMessage) MarshalBinary() ([]byte, error) {
	data, err := json.Marshal(distMsg)
	if err != nil {
		log.Printf("Cannot marshal message: %v", err)
		return data, err
	}
	return data, nil
}

// Connect connects a Distributor to the redis server
func (distr *Distributor) Connect() error {

	distr.client = *redis.NewClient(&redis.Options{
		Addr:     distr.Server,
		Password: distr.ServerPassword,
		DB:       0,
	})

	if err := distr.pingServer(); err != nil {
		return err
	}

	distr.ctx = context.Background()
	return nil
}

func (distr *Distributor) pingServer() error {
	log.Println("Try to ping redis...")
	err := distr.client.Ping(context.Background()).Err()
	if err != nil {
		log.Println("Pinging redis failed. Trying again in 3 seconds.")
		time.Sleep(3 * time.Second)
		err := distr.client.Ping(context.Background()).Err()
		if err != nil {
			return err
		}
	}
	log.Println("Ping succeeded.")
	return nil
}

// Subscribe subscribes the Distributor to a topic
func (distr *Distributor) Subscribe(serverId string) {
	subscription := distr.client.Subscribe(distr.ctx, distr.Topic)

	for msg := range subscription.Channel() {
		timer := prometheus.NewTimer(MessageProcessingTime)

		MessageCounterVec.WithLabelValues("incoming").Inc()

		log.Printf("Received raw distributor distMsg: %s", msg)

		var distMsg DistributionMessage
		err := distMsg.UnmarshalBinary([]byte(msg.Payload))
		if err != nil {
			log.Panicln(err)
		}

		if distMsg.ServerId == serverId {
			continue
		}

		message := chat.Message{
			MessageId: distMsg.MessageId,
			Text:      distMsg.Text,
			Sender:    distMsg.Sender,
			SentAt:    distMsg.SentAt,
			Room:      distMsg.Room,
		}

		wrapper := MessageWrapper{message: &message, processingTimer: timer, sourceDistributor: true}

		incoming <- &wrapper
	}
}

// Publish publishes MessageWrappers written in the outgoing channel
func (distr *Distributor) Publish(serverId string) {
	for message := range distr.Outgoing {

		distMsg := DistributionMessage{
			MessageId: message.MessageId,
			Text:      message.Text,
			Sender:    message.Sender,
			SentAt:    message.SentAt,
			Room:      message.Room,
			ServerId:  serverId,
		}

		err := distr.client.Publish(distr.ctx, distr.Topic, distMsg).Err()
		if err != nil {
			log.Fatal("Failed to publish distMsg via the distributor: ", err)
		}

		fmt.Println("Sent a new distMsg via the distributor: ", distMsg)
	}
}
