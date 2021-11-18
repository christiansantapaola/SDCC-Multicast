package clientlc

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
	"sdcc/pkg/overlay/api"
)

type Router struct {
	server  *ReceiverLC
	clients []*SenderLC
}

func NewRouter(self string, services []net.Addr, opt []grpc.DialOption) (*Router, error) {
	queue := NewSyncQueueMessageLC()
	clock := NewClock()
	server := NewReceiverLC(self, queue, clock)
	clients := make([]*SenderLC, len(services))
	for _, addr := range services {
		client, err := NewSenderLC(addr, opt)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	router := Router{server: server, clients: clients,}
	return &router, nil
}

func (router *Router) StartServer(serverService net.Addr, opt []grpc.ServerOption) {
	lis, err := net.Listen(serverService.Network(), serverService.String())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(opt...)
	api.RegisterReceiverLCServer(grpcServer, router.server)
	err = grpcServer.Serve(lis)
	if err != nil {
		return 
	}
}

func (router *Router) Send(ctx context.Context,mType api.MessageType, clock uint64, src, id, data string, ) error {
	for _, client := range router.clients {
		_, err := client.Send(ctx, mType, clock, src, id, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (router *Router) Recv() *api.MessageLC {
	return router.server.Pop()
}


