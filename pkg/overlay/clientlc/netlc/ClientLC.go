package netlc

import (
	"context"
	pb "github.com/christiansantapaola/SDCC-Multicast/pkg/overlay/clientlc/pb"
	"google.golang.org/grpc"
	"net"
)

/*
	ClientLC è un oggetto che gestisce una connesione client con il servizio grpc MessageQueueLC
	Si occupa di aprire una connessione è di chiamare i metodi rpc sul servizio.
*/

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
