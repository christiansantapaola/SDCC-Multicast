package netlc

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
	pb "sdcc/pkg/overlay/clientlc/pb"
	"sync"
	"time"
)

type MulticastSender struct {
	services []net.Addr
	opts     []grpc.DialOption
	clients  []*ClientLC
	mutex    sync.Mutex
}

func NewMulticastSender(services []net.Addr, opts []grpc.DialOption) (*MulticastSender, error) {
	clients := make([]*ClientLC, len(services))
	for i := 0; i < len(clients); i++ {
		clients[i] = nil
	}
	multicastSender := MulticastSender{clients: clients, services: services, opts: opts}
	return &multicastSender, nil
}

func (multicastSender *MulticastSender) Connect(ctx context.Context) error {
	multicastSender.mutex.Lock()
	defer multicastSender.mutex.Unlock()
	for !multicastSender.AreAllConnected() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := multicastSender.TryConnect()
			if err != nil {
				time.Sleep(1 * time.Second)
			}
		}
	}
	return nil
}

func (multicastSender *MulticastSender) TryConnect() error {
	var errRes error = nil
	for i := 0; i < len(multicastSender.clients); i++ {
		if multicastSender.clients[i] == nil {
			clients, err := NewClientLC(multicastSender.services[i], multicastSender.opts)
			if err != nil {
				log.Println(err)
				errRes = err
			}
			multicastSender.clients[i] = clients
		}
	}
	return errRes
}

func (multicastSender *MulticastSender) AreAllConnected() bool {
	for i := 0; i < len(multicastSender.clients); i++ {
		if multicastSender.clients[i] == nil {
			return false
		}
	}
	return true
}

func isAllTrue(vec []bool) bool {
	for _, boolean := range vec {
		if !boolean {
			return false
		}
	}
	return true
}

func (multicastSender *MulticastSender) Send(ctx context.Context, message *pb.MessageLC) error {
	numClient := len(multicastSender.clients)
	for i := 0; i < numClient; i++ {
		_, err := multicastSender.clients[i].Enqueue(ctx, message)
		for err != nil {
			log.Printf("[MulticastSender.Send()] %v\n", err)
			time.Sleep(1 * time.Second)
			_, err = multicastSender.clients[i].Enqueue(ctx, message)
		}
	}
	return nil
}

func (multicastSender *MulticastSender) TrySend(ctx context.Context, message *pb.MessageLC) error {
	numClient := len(multicastSender.clients)
	for i := 0; i < numClient; i++ {
		_, err := multicastSender.clients[i].Enqueue(ctx, message)
		if err != nil {
			log.Printf("[MulticastSender.TrySend()] %v\n", err)
			continue
		}
	}
	return nil
}
