package queue

import (
	"container/list"
	api "sdcc/pkg/overlay/clientseq/pb"
	"sync"
)

type MessageSeqFIFO struct {
	queue *list.List
	mutex sync.Mutex
}

func NewMessageSeqFifo() *MessageSeqFIFO {
	return &MessageSeqFIFO{queue: list.New()}
}

func (queue *MessageSeqFIFO) Len() int { return queue.queue.Len() }

func (queue *MessageSeqFIFO) Push(message *api.MessageSeq) {
	if message == nil {
		return
	}
	queue.mutex.Lock()
	queue.queue.PushBack(message)
	queue.mutex.Unlock()
}

func (queue *MessageSeqFIFO) Pop() *api.MessageSeq {
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
	return val.(*api.MessageSeq)
}
