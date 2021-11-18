package clientlc

import (
	"sdcc/pkg/overlay/api"
	"sync"
)

type SyncMessageLCQueue struct {
	queue *MessageLCQueue
	mutex sync.Mutex
}

func NewSyncQueueMessageLC() *SyncMessageLCQueue {
	queue := make(MessageLCQueue, 4096)
	return &SyncMessageLCQueue{queue: &queue}
}


func (queue *SyncMessageLCQueue) Push(message *api.MessageLC) {
	if message == nil {
		return
	}
	queue.mutex.Lock()
	queue.queue.Push(message)
	queue.mutex.Unlock()
}

func (queue *SyncMessageLCQueue) Pop() *api.MessageLC {
	if queue.queue.Len() < 1 {
		return nil
	}
	queue.mutex.Lock()
	message := queue.queue.Pop()
	queue.mutex.Unlock()
	return message
}

func (queue *SyncMessageLCQueue) Len() int {
	queue.mutex.Lock()
	length := queue.queue.Len()
	queue.mutex.Unlock()
	return length
}

func (queue *SyncMessageLCQueue) IsEmpty() bool {
	queue.mutex.Lock()
	isEmpty := queue.Len() == 0
	queue.mutex.Unlock()
	return isEmpty
}


