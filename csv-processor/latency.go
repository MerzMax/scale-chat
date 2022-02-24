package main

import (
	"errors"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"
)

type MessageLatency struct {
	id         string
	sender     string
	receiver   string
	sentAt     time.Time
	receivedAt time.Time
	latency    int64
}

func (m *MessageLatency) StringArray() []string {
	return []string{
		m.id,
		m.sender,
		m.receiver,
		m.sentAt.Format(time.RFC3339Nano),
		m.receivedAt.Format(time.RFC3339Nano),
		strconv.FormatInt(m.latency, 10),
	}
}

func calcLatency(sent Message, received Message) MessageLatency {
	return MessageLatency{
		id:         received.id,
		sender:     received.sender,
		receiver:   received.receiver,
		sentAt:     sent.timestamp,
		receivedAt: received.timestamp,
		latency:    received.timestamp.UnixMicro() - sent.timestamp.UnixMicro(),
	}
}

func convertToMessageLatencies(messages []Message) ([]MessageLatency, error) {
	sent := filterSentMessages(messages)

	if len(sent) < 1 {
		return nil, errors.New("did not find any sent messages")
	}

	var results []MessageLatency
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	wg.Add(len(sent))

	for _, sentMessage := range sent {
		sentMessage := sentMessage
		go func() {
			defer wg.Done()
			received := findReceivedMessages(sentMessage, messages)

			var partialResults []MessageLatency
			for _, receivedMessage := range received {
				partialResults = append(partialResults, calcLatency(sentMessage, receivedMessage))
			}

			mu.Lock()
			results = append(results, partialResults...)
			mu.Unlock()
		}()
	}

	wg.Wait()

	return results, nil
}

func CalculateLatency(loadTest LoadTest, messages []Message) [][]string {
	received, err := convertToMessageLatencies(messages)
	if err != nil {
		log.Fatal("Did not find any sent messages in ", loadTest.filename)
	}

	sort.Slice(received, func(i, j int) bool {
		return received[i].sentAt.Before(received[j].sentAt)
	})

	results := make([][]string, len(received)+1)
	results[0] = []string{"id", "sender", "receiver", "sent_at", "received_at", "latency"}

	for i, c := range received {
		results[i+1] = c.StringArray()
	}

	return results
}
