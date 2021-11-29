package netseq

import (
	"context"
	"google.golang.org/grpc"
	"net"
	pb "sdcc/pkg/overlay/clientseq/pb"
)

/*
	ClientSeq è un oggetto che gestisce una connesione client con il servizio grpc MessageQueueSeq
	Si occupa di aprire una connessione è di chiamare i metodi rpc sul servizio.
*/

type ClientSeq struct {
	Dial      *grpc.ClientConn
	clientSeq pb.MessageQueueSeqClient
}

func NewClientSeq(service net.Addr, opt []grpc.DialOption) (*ClientSeq, error) {
	dial, err := grpc.Dial(service.String(), opt...)
	if err != nil {
		return nil, err
	}
	client := pb.NewMessageQueueSeqClient(dial)
	return &ClientSeq{Dial: dial, clientSeq: client}, nil
}

func (sender *ClientSeq) Enqueue(ctx context.Context, message *pb.MessageSeq) (*pb.EnqueueReply, error) {
	return sender.clientSeq.Enqueue(ctx, message)
}
