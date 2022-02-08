package main

import (
	"context"
	"encoding/json"
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
	Message  chat.Message `json:"message"`
	ServerId string       `json:"server_id"`
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
func (distMsg DistributionMessage) MarshalBinary() ([]byte, error) {
	data, err := json.Marshal(distMsg)
	if err != nil {
		log.Printf("Cannot marshal message: %v", err)
		return data, err
	}
	return data, nil
}

// Ping connects a Distributor to the redis server
func (distr *Distributor) Ping() error {

	distr.client = *redis.NewClient(&redis.Options{
		Addr:     distr.Server,
		Password: distr.ServerPassword,
		DB:       0,
	})

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

	distr.ctx = context.Background()
	return nil
}

// Subscribe subscribes the Distributor to a topic
func (distr *Distributor) Subscribe(serverId string) {
	subscription := distr.client.Subscribe(distr.ctx, distr.Topic)

	for msg := range subscription.Channel() {
		timer := prometheus.NewTimer(MessageProcessingTime)

		MessageCounterVec.WithLabelValues("incoming").Inc()

		var distMsg DistributionMessage
		err := distMsg.UnmarshalBinary([]byte(msg.Payload))
		if err != nil {
			log.Panicln(err)
		}

		if distMsg.ServerId == serverId {
			continue
		}

		wrapper := MessageWrapper{message: &distMsg.Message, processingTimer: timer, source: DISTRIBUTOR}

		incoming <- &wrapper
	}
}

// Publish publishes MessageWrappers written in the outgoing channel
func (distr *Distributor) Publish(serverId string) {
	for message := range distr.Outgoing {

		distMsg := DistributionMessage{
			Message:  *message,
			ServerId: serverId,
		}

		err := distr.client.Publish(distr.ctx, distr.Topic, distMsg).Err()
		if err != nil {
			log.Fatal("Failed to publish distMsg via the distributor: ", err)
		}

		log.Println("Sent a new distMsg via the distributor: ", distMsg)
	}
}
