package clientlc

import (
	"context"
	"google.golang.org/grpc"
	"net"
	"sdcc/pkg/overlay/api"
)

type SenderLC struct {
	Dial     *grpc.ClientConn
	ClientLC api.ReceiverLCClient
}

func NewSenderLC(service net.Addr, opt []grpc.DialOption) (*SenderLC, error) {
	dial, err := grpc.Dial(service.String(), opt...)
	if err != nil {
		return nil, err
	}
	client := api.NewReceiverLCClient(dial)
	return &SenderLC{Dial: dial, ClientLC: client}, nil
}

func (sender *SenderLC) Send(ctx context.Context, messageType api.MessageType, clock uint64, source, id, data string) (*api.SendReply, error) {
	msg := &api.MessageLC{Type: messageType, Src: source, Clock: clock, Id: id, Data: data}
	reply, err := sender.ClientLC.Send(ctx, msg)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
