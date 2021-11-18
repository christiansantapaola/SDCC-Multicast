package clientlc

import (
	"sdcc/pkg/overlay/api"
)

type MessageLCQueue []*api.MessageLC

func (mq MessageLCQueue) Len() int { return len(mq) }
func (mq *MessageLCQueue) Push(msg *api.MessageLC) {
	if msg == nil {
		return
	}
	*mq = append(*mq, msg)
}

func (mq *MessageLCQueue) Pop() *api.MessageLC {
	n := len(*mq)
	if n == 0 {
		return nil
	}
	val := (*mq)[0]
	*mq = (*mq)[1:n-1]
	return val
}

