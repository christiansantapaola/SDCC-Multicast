package queue

import (
	"container/heap"
	"fmt"
	"log"
	pb "sdcc/pkg/overlay/clientvc/pb"
	"sync"
)

type MessageHeap []*pb.MessageVC

func (queue MessageHeap) Len() int      { return len(queue) }
func (queue MessageHeap) IsEmpty() bool { return len(queue) == 0 }
func (queue MessageHeap) Less(i, j int) bool {
	groupSize := len(queue[i].GetClock())
	for k := 0; k < groupSize; k++ {
		if queue[i].GetClock()[k] > queue[j].GetClock()[k] {
			return false
		}
	}
	return true
}
func (queue MessageHeap) Swap(i, j int) { queue[i], queue[j] = queue[j], queue[i] }

func (queue *MessageHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*queue = append(*queue, x.(*pb.MessageVC))
}

func (queue *MessageHeap) Pop() interface{} {
	old := *queue
	n := len(old)
	if n == 0 {
		return nil
	}
	x := old[n-1]
	*queue = old[0 : n-1]
	return x
}

func (queue MessageHeap) Peek() *pb.MessageVC {
	n := len(queue)
	if n > 0 {
		return queue[0]
	} else {
		return nil
	}
}

type MessageVCRecvQueue struct {
	queue        *MessageHeap
	groups       []string
	messageTable *MessageTable
	mutex        sync.Mutex
	verbose      bool
	groupSize    int
}

func NewMessageVCRecvQueue(groups []string, verbose bool) *MessageVCRecvQueue {
	queue := make(MessageHeap, 0)
	return &MessageVCRecvQueue{
		queue:        &queue,
		groups:       groups,
		messageTable: NewMessageTable(len(groups)),
		verbose:      verbose,
		groupSize:    len(groups),
	}
}

func (queue *MessageVCRecvQueue) Push(message *pb.MessageVC) error {
	if message == nil {
		return fmt.Errorf("messageTable to Push() is nil")
	}
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if queue.verbose {
		log.Printf("[MessageVCRecVQueue:INFO] Push(messageTable: {src: %s, type: %s, clock: %d, id: %s})\n",
			message.GetSrc(), message.GetType().String(), message.GetClock(), message.GetId())
	}
	queue.messageTable.Insert(message)
	if message.GetType() != pb.MessageType_ACK {
		heap.Push(queue.queue, message)
	}
	return nil
}

func (queue *MessageVCRecvQueue) Pop() *pb.MessageVC {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if queue.queue.Len() < 1 {
		return nil
	}
	msg := heap.Pop(queue.queue).(*pb.MessageVC)
	if msg == nil {
		// queue is empty
		return nil
	}
	isReady, err := queue.messageTable.IsReady(msg.GetId())
	if err != nil {
		heap.Push(queue.queue, msg)
		return nil
	}
	if isReady {
		return msg
	} else {
		heap.Push(queue.queue, msg)
		return nil
	}
}

func (queue *MessageVCRecvQueue) Len() int {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	length := queue.queue.Len()
	return length
}

func (queue *MessageVCRecvQueue) IsEmpty() bool {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	isEmpty := queue.Len() == 0
	return isEmpty
}
