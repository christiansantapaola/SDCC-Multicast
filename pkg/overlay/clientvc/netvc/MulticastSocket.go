package netvc

import (
	"context"
	api "github.com/christiansantapaola/SDCC-Multicast/pkg/overlay/clientvc/pb"
	"google.golang.org/grpc"
	"net"
	"sync"
	"time"
)

/*
	Una MulticastSocket è un oggetto composto da:
		1. un server grpc che ha funzione di canale di ricezione.
		2. un MulticastSender che ha funzione di un canale di trasmissione.
	Questa struttura astrae la comunicazione p2p basata su un server e vari client grpc dando una singola
	interfaccia per entrambe le situazioni.
*/

type MulticastSocket struct {
	server     *ServerLC
	sender     *MulticastSender
	serverOpts []grpc.ServerOption
	dialOpts   []grpc.DialOption
	mutex      sync.Mutex
	verbose    bool
}

/*
	Instanzia una Nuova MulticastSocket:
		- Instanzia su thread separato il server grpc
		- Instanzia i vari client grpc.
*/

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
