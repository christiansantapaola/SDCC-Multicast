package queue

import (
	"container/list"
	api "github.com/christiansantapaola/SDCC-Multicast/pkg/overlay/clientlc/pb"
	"sync"
)

/*
	Implementazione di una coda FIFO, basata su linked list (list.list).
	La coda FIFO è thread-safe grazie all utilizzo di un mutex
*/

type MessageLCFIFO struct {
	queue *list.List
	mutex sync.Mutex
}

func NewMessageLCFifo() *MessageLCFIFO {
	return &MessageLCFIFO{queue: list.New()}
}

func (queue *MessageLCFIFO) Len() int {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	return queue.queue.Len()
}

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
