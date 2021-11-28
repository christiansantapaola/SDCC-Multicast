package queue

import (
	"container/list"
	api "sdcc/pkg/overlay/clientlc/pb"
	"sync"
)

type MessageLCFIFO struct {
	queue *list.List
	mutex sync.Mutex
}

func NewMessageLCFifo() *MessageLCFIFO {
	return &MessageLCFIFO{queue: list.New()}
}

func (queue *MessageLCFIFO) Len() int { return queue.queue.Len() }

func (queue *MessageLCFIFO) Push(message *api.MessageLC) {
	if message == nil {
		return
	}
	queue.mutex.Lock()
	queue.queue.PushBack(message)
	queue.mutex.Unlock()
}

func (queue *MessageLCFIFO) Pop() *api.MessageLC {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	msg := queue.queue.Front()
	if msg == nil {
		return nil
	}
	if msg.Value == nil {
		return nil
	}
	val := queue.queue.Remove(msg)
	return val.(*api.MessageLC)
}
