package netseq

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	pb "sdcc/pkg/overlay/clientseq/pb"
	"sdcc/pkg/overlay/clientseq/queue"
)

/*
	Implementazione del server grpc del servizio MessageQueueSeq
	Il server viene fatto partire su un altro thread da quello chiamante.
	`ServerSeq` offre una coda FIFO con un metodo remoto Enqueue e un metodo locale Pop per prendere i messaggi invocati.
*/

type ServerSeq struct {
	pb.UnimplementedMessageQueueSeqServer
	queue    *queue.MessageSeqFIFO
	listener *net.Listener
	server   *grpc.Server
	verbose  bool
}

func NewServerSeq(verbose bool) *ServerSeq {
	rec := ServerSeq{queue: queue.NewMessageSeqFifo(), verbose: verbose}
	return &rec
}

func (receiver *ServerSeq) StartServer(port int, opts []grpc.ServerOption) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	receiver.listener = &lis
	receiver.server = grpc.NewServer(opts...)
	pb.RegisterMessageQueueSeqServer(receiver.server, receiver)
	err = receiver.server.Serve(lis)
	if err != nil {
		log.Fatalf("[ServerSeq] %v\n", err)
	}
}

func (receiver *ServerSeq) Stop() {
	receiver.server.GracefulStop()
}

// grpc function, can be called by remote host
func (receiver *ServerSeq) Enqueue(ctx context.Context, message *pb.MessageSeq) (*pb.EnqueueReply, error) {
	if receiver.verbose {
		log.Printf("[ServerSeq] Ready to enqueue Message: %s, %s, %s\n", message.GetSrc(), message.GetId(), message.GetData())
	}
	receiver.queue.Push(message)
	return &pb.EnqueueReply{}, nil
}

// local function
func (receiver *ServerSeq) Pop() *pb.MessageSeq {
	return receiver.queue.Pop()
}
