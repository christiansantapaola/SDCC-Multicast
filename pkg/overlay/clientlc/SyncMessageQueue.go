package clientlc

import (
	"container/heap"
	"sync"
)

type SyncMessageQUeue struct {
	queue *MessageQueue
	mutex sync.Mutex
}

func NewSyncMessageQueue() *SyncMessageQUeue {
	queue := make(MessageQueue, 4096)
	return &SyncMessageQUeue{queue: &queue}
}


func (queue *SyncMessageQUeue) Push(message *Message) {
	if message == nil {
		return
	}
	queue.mutex.Lock()
	heap.Push(queue.queue, message)
	queue.mutex.Unlock()
}

func (queue *SyncMessageQUeue) Pop() *Message {
	if queue.queue.Len() < 1 {
		return nil
	}
	queue.mutex.Lock()
	message := heap.Pop(queue.queue).(*Message)
	queue.mutex.Unlock()
	return message
}

func (queue *SyncMessageQUeue) Len() int {
	queue.mutex.Lock()
	length := queue.queue.Len()
	queue.mutex.Unlock()
	return length
}

func (queue *SyncMessageQUeue) IsEmpty() bool {
	queue.mutex.Lock()
	isEmpty := queue.queue.IsEmpty()
	queue.mutex.Unlock()
	return isEmpty
}
