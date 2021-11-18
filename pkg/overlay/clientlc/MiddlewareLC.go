package clientlc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"sdcc/pkg/overlay/api"
	"time"
)

type MiddlewareLC struct {
	self string
	groupSize uint64
	table map[string]uint64
	messageQueue *SyncMessageQUeue
	clock *Clock
	router *Router
	stop bool
}

func NewMiddlewareLC(self string, services []net.Addr, opt []grpc.DialOption) (*MiddlewareLC, error){
	router, err := NewRouter(self, services, opt)
	if err != nil {
		return nil, err
	}
	return &MiddlewareLC{
		self:         self,
		groupSize:    uint64(len(services)),
		table:        make(map[string]uint64),
		messageQueue: NewSyncMessageQueue(),
		clock:        NewClock(),
		router:       router,
		stop:         false,
	}, nil
}

func (middleware *MiddlewareLC) push(message *api.MessageLC) {
	if _, ok := middleware.table[message.Id]; ok {
		middleware.table[message.Id]++
	} else {
		middleware.table[message.Id] = 1
	}
	if message.GetType() == api.MessageType_APPLICATION {
		umsg := NewMessage(message)
		middleware.messageQueue.Push(umsg)
	}
}

func (middleware *MiddlewareLC) pop() *Message {
	msg := middleware.messageQueue.Pop()
	if middleware.table[msg.GetId()] < middleware.groupSize {
		middleware.messageQueue.Push(msg)
		return nil
	} else {
		return msg
	}
}

func (middleware *MiddlewareLC) GetID(clock uint64) string {
	return fmt.Sprintf("%s:%d", middleware.self, clock)
}

func (middleware *MiddlewareLC) Send(ctx context.Context, message string) error {
	clock := middleware.clock.Increase()
	id := middleware.GetID(clock)
	err := middleware.router.Send(ctx, api.MessageType_APPLICATION, clock, middleware.self, id, message)
	if err != nil {
		return err
	}
	return nil
}

func (middleware *MiddlewareLC) RecvWork() {
	for !middleware.stop {
		msg := middleware.router.Recv()
		if msg == nil {
			time.Sleep(1 * time.Millisecond)
			continue
		}
		middleware.push(msg)
	}
}

func (middleware *MiddlewareLC) Recv(ctx context.Context) *Message {
	for {
		msg := middleware.pop()
		if msg == nil {
			select {
				case <-time.After(1 * time.Millisecond):
					continue
			case <-ctx.Done():
				msg = middleware.pop()
				if msg == nil {
					return nil
				} else {
					middleware.clock.Update(msg.GetClock())
					return msg
				}
			}
 		} else {
			middleware.clock.Update(msg.GetClock())
			return msg
		}
	}
}
