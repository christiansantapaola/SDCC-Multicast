package queue

import (
	"container/list"
	api "sdcc/pkg/overlay/clientvc/pb"
	"sync"
)

/*
	Implementazione di una coda FIFO, basata su linked list (list.list).
	La coda FIFO è thread-safe grazie all utilizzo di un mutex
*/

type MessageVCFIFO struct {
	queue *list.List
	mutex sync.Mutex
}

func NewMessageVCFifo() *MessageVCFIFO {
	return &MessageVCFIFO{queue: list.New()}
}

func (queue *MessageVCFIFO) Len() int { return queue.queue.Len() }

func (queue *MessageVCFIFO) Push(message *api.MessageVC) {
	if message == nil {
		return
	}
	queue.mutex.Lock()
	queue.queue.PushBack(message)
	queue.mutex.Unlock()
}

func (queue *MessageVCFIFO) Pop() *api.MessageVC {
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
	return val.(*api.MessageVC)
}
