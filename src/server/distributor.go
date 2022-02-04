package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

type Distributor struct {
	Server         string
	ServerPassword string
	Incoming       chan<- *MessageWrapper
	client         redis.Client
	ctx            context.Context
}

// Connect connects a Distributor to the redis server
func (distributor *Distributor) Connect() error {

	distributor.client = *redis.NewClient(&redis.Options{
		Addr:     distributor.Server,
		Password: distributor.ServerPassword,
		DB:       0,
	})

	if err := distributor.pingServer(); err != nil {
		return err
	}

	distributor.ctx = context.Background()
	return nil
}

func (distributor *Distributor) pingServer() error {
	err := distributor.client.Ping(context.Background()).Err()
	if err != nil {
		// Sleep for 3 seconds and wait for Redis to initialize
		time.Sleep(3 * time.Second)
		err := distributor.client.Ping(context.Background()).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

// SubscribeToTopics subscribes the Distributor to a list of provided topics
func (distributor *Distributor) SubscribeToTopics(topics ...string) error {
	for _, topic := range topics {
		t := distributor.client.Subscribe(distributor.ctx, topic)
		// Get the Channel to use
		channel := t.Channel()

		for msg := range channel {
			var msgWrapper *MessageWrapper
			err := msgWrapper.UnmarshalBinary([]byte(msg.Payload))
			if err != nil {
				return err
			}

			incoming <- msgWrapper

			fmt.Println("Received a new message via redis: ", msg)
		}
	}
	return nil
}

// PublishMessage publishes a MessageWrapper to the provided topic
func (distributor *Distributor) PublishMessage(topic string, msgWrapper MessageWrapper) error {
	data, err := msgWrapper.MarshalBinary()
	if err != nil {
		return err
	}

	err = distributor.client.Publish(distributor.ctx, topic, data).Err()
	if err != nil {
		return err
	}
	return nil
}
