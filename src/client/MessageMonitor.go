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
	ClientId  string
	Type      Type
	Timestamp time.Time
}

func (m MessageEventEntry) ConvertToStringArray() []string {
	return []string{
		strconv.FormatUint(m.MessageId, 10),
		m.ClientId,
		m.Type.String(),
		strconv.FormatUint(uint64(m.Timestamp.UnixMicro()), 10),
	}
}
