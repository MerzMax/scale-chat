package main

import (
	"errors"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"
)

type CountComparison struct {
	sentAt   time.Time
	expected uint
	actual   uint
}

func (c *CountComparison) Percent() float64 {
	return (float64(c.actual) / float64(c.expected)) * 100
}

func (c *CountComparison) StringArray() []string {
	return []string{
		c.sentAt.Format(time.RFC3339Nano),
		strconv.Itoa(int(c.actual)),
		strconv.Itoa(int(c.expected)),
	}
}

func compareMessageCounts(sent Message, clientCount uint, received []Message) CountComparison {
	return CountComparison{
		sentAt:   sent.timestamp,
		expected: clientCount,
		actual:   uint(len(received)),
	}
}

func compareExpectedAndActual(clientCount uint, messages []Message) ([]CountComparison, error) {
	sent := filterSentMessages(messages)

	if len(sent) < 1 {
		return nil, errors.New("did not find any sent messages")
	}

	var comparisons []CountComparison
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	wg.Add(len(sent))

	for _, message := range sent {
		message := message
		go func() {
			defer wg.Done()
			received := findReceivedMessages(message, messages)
			comparison := compareMessageCounts(message, clientCount, received)

			mu.Lock()
			comparisons = append(comparisons, comparison)
			mu.Unlock()
		}()
	}

	wg.Wait()

	return comparisons, nil
}

func CompareCount(loadTest LoadTest, messages []Message) [][]string {
	comparisons, err := compareExpectedAndActual(loadTest.roomSize, messages)
	if err != nil {
		log.Fatal("Did not find any sent messages in ", loadTest.filename)
	}

	sort.Slice(comparisons, func(i, j int) bool {
		return comparisons[i].sentAt.Before(comparisons[j].sentAt)
	})

	results := make([][]string, len(comparisons)+1)
	results[0] = []string{"sent_at", "received", "sent"}

	for i, c := range comparisons {
		results[i+1] = c.StringArray()
	}

	return results
}
