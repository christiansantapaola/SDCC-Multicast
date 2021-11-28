package netseq

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
	pb "sdcc/pkg/overlay/clientseq/pb"
	"time"
)

type MulticastSender struct {
	services []net.Addr
	opts     []grpc.DialOption
	clients  []*ClientSeq
}

func NewMulticastSender(services []net.Addr, opts []grpc.DialOption) (*MulticastSender, error) {
	clients := make([]*ClientSeq, len(services))
	for i := 0; i < len(clients); i++ {
		clients[i] = nil
	}
	multicastSender := MulticastSender{clients: clients, services: services, opts: opts}
	return &multicastSender, nil
}

func (multicastSender *MulticastSender) Connect(ctx context.Context) error {
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
				log.Printf("[MulticastSender.TryConnect()] %v\n", err)
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

func (multicastSender *MulticastSender) Send(ctx context.Context, message *pb.MessageSeq) error {
	numClient := len(multicastSender.clients)
	for i := 0; i < numClient; i++ {
		_, err := multicastSender.clients[i].Enqueue(ctx, message)
		for err != nil {
			time.Sleep(1 * time.Second)
			log.Printf("[MulticastSender.Send()] %v\n", err)
			_, err = multicastSender.clients[i].Enqueue(ctx, message)
		}
	}
	return nil
}

func (multicastSender *MulticastSender) TrySend(ctx context.Context, message *pb.MessageSeq) error {
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

func (multicastSender *MulticastSender) Relay(ctx context.Context, message *pb.MessageSeq, myRank int) error {
	numClient := len(multicastSender.clients)
	for i := 0; i < numClient; i++ {
		if i != myRank {
			_, err := multicastSender.clients[i].Enqueue(ctx, message)
			for err != nil {
				log.Printf("[MulticastSender.Relay()] %v\n", err)
				_, err = multicastSender.clients[i].Enqueue(ctx, message)
			}
		}
	}
	return nil

}

func (multicastSender *MulticastSender) TryRelay(ctx context.Context, message *pb.MessageSeq, myRank int) error {
	numClient := len(multicastSender.clients)
	for i := 0; i < numClient; i++ {
		if i != myRank {
			_, err := multicastSender.clients[i].Enqueue(ctx, message)
			if err != nil {
				log.Printf("[MulticastSender.TryRelay()] %v\n", err)
				continue
			}
		}
	}
	return nil

}
