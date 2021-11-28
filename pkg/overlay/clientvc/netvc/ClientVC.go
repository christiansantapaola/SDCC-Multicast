package netvc

import (
	"context"
	"google.golang.org/grpc"
	"net"
	pb "sdcc/pkg/overlay/clientvc/pb"
)

type ClientVC struct {
	Dial     *grpc.ClientConn
	ClientLC pb.MessageQueueVCClient
}

func NewClientVC(service net.Addr, opt []grpc.DialOption) (*ClientVC, error) {
	dial, err := grpc.Dial(service.String(), opt...)
	if err != nil {
		return nil, err
	}
	client := pb.NewMessageQueueVCClient(dial)
	return &ClientVC{Dial: dial, ClientLC: client}, nil
}

func (sender *ClientVC) Enqueue(ctx context.Context, message *pb.MessageVC) (*pb.EnqueueReply, error) {
	return sender.ClientLC.Enqueue(ctx, message)
}
