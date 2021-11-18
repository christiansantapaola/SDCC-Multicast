package clientlc

import "sdcc/pkg/overlay/api"

type Message struct {
	src string
	id string
	clock uint64
	message string
}

func NewMessage(apiMsg *api.MessageLC) *Message {
	return &Message{src: apiMsg.GetSrc(), id: apiMsg.GetId(), clock: apiMsg.GetClock(), message: apiMsg.GetData()}
}

func (m *Message) GetSrc() string {
	return m.src
}

func (m *Message) GetId() string {
	return m.id
}

func (m *Message) GetData() string {
	return m.message
}

func (m *Message) GetClock() uint64 {
	return m.clock
}