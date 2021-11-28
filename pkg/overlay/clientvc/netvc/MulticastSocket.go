package netvc

import (
	"context"
	"google.golang.org/grpc"
	"net"
	api "sdcc/pkg/overlay/clientvc/pb"
	"sync"
	"time"
)

type MulticastSocket struct {
	server     *ServerLC
	sender     *MulticastSender
	serverOpts []grpc.ServerOption
	dialOpts   []grpc.DialOption
	mutex      sync.Mutex
	verbose    bool
}

func Open(ctx context.Context, servicePort int, serverOpts []grpc.ServerOption, senders []net.Addr, dialOpts []grpc.DialOption, verbose bool) (*MulticastSocket, error) {
	server := NewServerLC(verbose)
	sender, err := NewMulticastSender(senders, dialOpts)
	if err != nil {
		return nil, err
	}
	err = sender.Connect(ctx)
	if err != nil {
		return nil, err
	}
	go server.StartServer(servicePort, serverOpts)
	time.Sleep(time.Second)
	return &MulticastSocket{
			server:     server,
			sender:     sender,
			serverOpts: serverOpts,
			dialOpts:   dialOpts,
			verbose:    verbose},
		nil
}

func (socket *MulticastSocket) Send(ctx context.Context, message *api.MessageVC) error {
	socket.mutex.Lock()
	defer socket.mutex.Unlock()
	return socket.sender.Send(ctx, message)
}

func (socket *MulticastSocket) TrySend(ctx context.Context, message *api.MessageVC) error {
	socket.mutex.Lock()
	defer socket.mutex.Unlock()
	return socket.sender.TrySend(ctx, message)
}

func (socket *MulticastSocket) IsEmpty() bool {
	return socket.server.queue.Len() == 0
}

func (socket *MulticastSocket) Recv(ctx context.Context) (*api.MessageVC, error) {
	for {
		lcMessage := socket.server.Pop()
		if lcMessage == nil {
			select {
			case <-time.After(1 * time.Millisecond):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return lcMessage, nil
	}
}

func (socket *MulticastSocket) NonBlockingRecv() *api.MessageVC {
	return socket.server.Pop()
}

func (socket *MulticastSocket) Close() {
	socket.server.server.GracefulStop()
}
