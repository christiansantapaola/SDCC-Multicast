package netvc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	pb "sdcc/pkg/overlay/clientvc/pb"
	"sdcc/pkg/overlay/clientvc/queue"
)

/*
	Implementazione del server grpc del servizio MessageQueueVC:
		La struttura offre un metodo remote:
		- `Enqueue()`
		il quale riceve un messaggio protobuf è lo immette in una coda FIFO
		A quest ultimo è contrapposto un metodo locale:
		- `Pop()`
		Che rimuove il primo elemento dalla coda FIFO.
*/

type ServerLC struct {
	pb.UnimplementedMessageQueueVCServer
	queue    *queue.MessageVCFIFO
	listener *net.Listener
	server   *grpc.Server
	verbose  bool
}

func NewServerLC(verbose bool) *ServerLC {
	rec := ServerLC{queue: queue.NewMessageVCFifo(), verbose: verbose}
	return &rec
}

func (receiver *ServerLC) StartServer(port int, opts []grpc.ServerOption) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	receiver.listener = &lis
	receiver.server = grpc.NewServer(opts...)
	pb.RegisterMessageQueueVCServer(receiver.server, receiver)
	err = receiver.server.Serve(lis)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
}

func (receiver *ServerLC) Stop() {
	receiver.server.GracefulStop()
}

// grpc function, can be called by remote host
func (receiver *ServerLC) Enqueue(ctx context.Context, message *pb.MessageVC) (*pb.EnqueueReply, error) {
	if receiver.verbose {
		log.Printf("[ServerLC] Ready to enqueue Message: %s, %s, %s\n", message.GetSrc(), message.GetId(), message.GetData())
	}
	receiver.queue.Push(message)
	return &pb.EnqueueReply{}, nil
}

// local function
func (receiver *ServerLC) Pop() *pb.MessageVC {
	return receiver.queue.Pop()
}
