package main

import "time"

type Message struct {
	id        string
	sender    string
	receiver  string
	msgType   string
	timestamp time.Time
}

func removeDuplicates(messages []Message) []Message {
	keys := make(map[string]bool)
	var filtered []Message

	for _, message := range messages {
		key := message.id + message.sender
		if _, val := keys[key]; !val {
			keys[key] = true
			filtered = append(filtered, message)
		}
	}

	return filtered
}

func filterSentMessages(messages []Message) []Message {
	var sent []Message

	for _, message := range messages {
		if message.msgType == "Sent" {
			sent = append(sent, message)
		}
	}

	return removeDuplicates(sent)
}

func findReceivedMessages(sent Message, messages []Message) []Message {
	var received []Message

	for _, message := range messages {
		if message.msgType == "Received" && message.id == sent.id && message.sender == sent.sender {
			received = append(received, message)
		}
	}

	return received
}
