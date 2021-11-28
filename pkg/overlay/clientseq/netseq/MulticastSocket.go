package netseq

import (
	"context"
	"google.golang.org/grpc"
	"net"
	api "sdcc/pkg/overlay/clientseq/pb"
	"sync"
	"time"
)

type MulticastSocket struct {
	server     *ServerSeq
	sender     *MulticastSender
	serverOpts []grpc.ServerOption
	dialOpts   []grpc.DialOption
	mutex      sync.Mutex
	verbose    bool
}

func Open(ctx context.Context, servicePort int, serverOpts []grpc.ServerOption, senders []net.Addr, dialOpts []grpc.DialOption, verbose bool) (*MulticastSocket, error) {
	server := NewServerSeq(verbose)
	sender, err := NewMulticastSender(senders, dialOpts)
	if err != nil {
		return nil, err
	}
	err = sender.Connect(ctx)
	if err != nil {
		return nil, err
	}
	go server.StartServer(servicePort, serverOpts)
	return &MulticastSocket{
			server:     server,
			sender:     sender,
			serverOpts: serverOpts,
			dialOpts:   dialOpts},
		nil
}

func (socket *MulticastSocket) Send(ctx context.Context, message *api.MessageSeq) error {
	socket.mutex.Lock()
	defer socket.mutex.Unlock()
	return socket.sender.Send(ctx, message)
}

func (socket *MulticastSocket) Relay(ctx context.Context, message *api.MessageSeq, myRank int) error {
	socket.mutex.Lock()
	defer socket.mutex.Unlock()
	return socket.sender.Relay(ctx, message, myRank)
}

func (socket *MulticastSocket) TrySend(ctx context.Context, message *api.MessageSeq) error {
	socket.mutex.Lock()
	defer socket.mutex.Unlock()
	return socket.sender.TrySend(ctx, message)
}

func (socket *MulticastSocket) TryRelay(ctx context.Context, message *api.MessageSeq, myRank int) error {
	socket.mutex.Lock()
	defer socket.mutex.Unlock()
	return socket.sender.TryRelay(ctx, message, myRank)
}

func (socket *MulticastSocket) Recv(ctx context.Context) (*api.MessageSeq, error) {
	for {
		seqMessage := socket.server.Pop()
		if seqMessage == nil {
			select {
			case <-time.After(1 * time.Millisecond):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return seqMessage, nil
	}
}

func (socket *MulticastSocket) Close() {
	socket.server.server.GracefulStop()
}