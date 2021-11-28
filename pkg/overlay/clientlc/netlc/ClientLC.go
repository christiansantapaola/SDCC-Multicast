package netlc

import (
	"context"
	"google.golang.org/grpc"
	"net"
	pb "sdcc/pkg/overlay/clientlc/pb"
)

type ClientLC struct {
	Dial     *grpc.ClientConn
	ClientLC pb.MessageQueueLCClient
}

func NewClientLC(service net.Addr, opt []grpc.DialOption) (*ClientLC, error) {
	dial, err := grpc.Dial(service.String(), opt...)
	if err != nil {
		return nil, err
	}
	client := pb.NewMessageQueueLCClient(dial)
	return &ClientLC{Dial: dial, ClientLC: client}, nil
}

func (sender *ClientLC) Enqueue(ctx context.Context, message *pb.MessageLC) (*pb.EnqueueReply, error) {
	return sender.ClientLC.Enqueue(ctx, message)
}
