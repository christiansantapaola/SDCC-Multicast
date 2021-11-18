package clientlc

import (
	"context"
)
import pb "sdcc/pkg/overlay/api"


type ReceiverLC struct {
	pb.UnimplementedReceiverLCServer
	self string
	queue *SyncMessageLCQueue
	clock *Clock
}

func NewReceiverLC(self string, queue *SyncMessageLCQueue, clock *Clock) *ReceiverLC {
	return &ReceiverLC{self: self, queue: queue, clock: clock}
}

// Send send for the clients, is recv for the server.
func (receiver *ReceiverLC) Send(ctx context.Context, message *pb.MessageLC) (*pb.SendReply, error) {
	receiver.queue.Push(message)
	return &pb.SendReply{}, nil
}

func (receiver *ReceiverLC) Pop() *pb.MessageLC {
	return receiver.queue.Pop()
}