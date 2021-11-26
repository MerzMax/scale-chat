package client

import "time"

type MessageEventEntry struct {
	MessageId uint64
	ClientId string
	Type Type
	TimeStamp time.Time
}

type Type int

const (
	Sent = iota
	Received
)

func (t Type) String() string {
	return []string{"Sent", "Received"}[t]
}




