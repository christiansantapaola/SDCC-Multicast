package netvc

import (
	"context"
	pb "github.com/christiansantapaola/SDCC-Multicast/pkg/overlay/clientvc/pb"
	"google.golang.org/grpc"
	"net"
)

/*
	ClientVC è un oggetto che gestisce una connesione client con il servizio grpc MessageQueueLC
	Si occupa di aprire una connessione è di chiamare i metodi rpc sul servizio.
*/

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
