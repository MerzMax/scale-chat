package client

import (
	"strconv"
	"time"
)

type Type int

const (
	Sent = iota
	Received
)

func (t Type) String() string {
	return []string{"Sent", "Received"}[t]
}

type MessageEventEntry struct {
	MessageId uint64
	SenderId  string
	ClientId  string
	Type      Type
	TimeStamp time.Time
}

func (m MessageEventEntry) StringArray() []string {
	return []string{
		strconv.FormatUint(m.MessageId, 10),
		m.SenderId,
		m.ClientId,
		m.Type.String(),
		strconv.FormatUint(
			uint64(
				m.TimeStamp.UnixMicro()),
			10)}
}
