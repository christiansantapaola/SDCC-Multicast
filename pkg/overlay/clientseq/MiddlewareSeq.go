package clientseq

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"sdcc/pkg/nameservice/client"
	"sdcc/pkg/nameservice/nameservice"
	"sdcc/pkg/overlay/clientseq/netseq"
	api "sdcc/pkg/overlay/clientseq/pb"
	"sdcc/pkg/overlay/clientseq/queue"
	"time"
)

type SystemEvent int

const (
	EXIT  SystemEvent = 0
	FATAL             = 1
)

func (event SystemEvent) String() string {
	switch event {
	case EXIT:
		return "EXIT"
	case FATAL:
		return "FATAL"
	default:
		return "NOT IMPLEMENTED"
	}
}

func EventFromString(eventStr string) (SystemEvent, error) {
	switch eventStr {
	case "EXIT":
		return EXIT, nil
	case "FATAL":
		return FATAL, nil
	default:
		return 0, fmt.Errorf("'%s' NOT IMPLEMENTED", eventStr)
	}
}

type MiddlewareSeq struct {
	nameServiceClient *nameservice.NameServiceClient
	group             *MulticastGroup
	clock             *Clock
	recvQueue         *queue.MessageSeqFIFO
	socket            *netseq.MulticastSocket
	logPath           string
	log               *MessageLog
	stop              bool
	verbose           bool
}

func NewMiddlewareSeq(self, groupName, logPath string, port int, nameserver *nameservice.NameServiceClient, verbose bool, dopt []grpc.DialOption, sopt []grpc.ServerOption) (*MiddlewareSeq, error) {
	if verbose {
		log.Printf("client.GetAddressGroup(context.Background(), *nameserver: %v, groupName:%s)\n", nameserver, groupName)
	}
	logf, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	users, err := client.GetAddressGroup(context.Background(), *nameserver, groupName)
	if err != nil {
		return nil, err
	}
	if verbose {
		log.Printf("group := NewGroup(self: %s, users: %v)\n\n", self, users)
	}
	group := NewGroup(self, users)
	rank, err := group.GetMyRank()
	if err != nil {
		return nil, err
	}
	services, _, _, err := group.GetUsersInfo()
	if err != nil {
		return nil, err
	}
	var socket *netseq.MulticastSocket
	if rank == 0 {
		if verbose {
			log.Printf("netseq.Open(port: %d, sopt: %v, services: %v, dopt: %v)\n", port, sopt, services, dopt)
		}
		socket, err = netseq.Open(context.Background(), port, sopt, services, dopt, verbose)
		if err != nil {
			return nil, err
		}
	} else {
		sequencerServices, err := group.GetSequencerServices()
		if err != nil {
			return nil, err
		}
		sequencer := []net.Addr{sequencerServices}
		if verbose {
			log.Printf("netseq.Open(port: %d, sopt: %v, services: %v, dopt: %v)\n", port, sopt, sequencer, dopt)
		}
		socket, err = netseq.Open(context.Background(), port, sopt, sequencer, dopt, verbose)
		if err != nil {
			return nil, err
		}
	}

	if verbose {
		log.Printf("recvQueue := queue.NewMessageLCRecvQueue(ids: %v)\n", services)
	}
	recvQueue := queue.NewMessageSeqFifo()
	clock := NewClock()
	middleware := MiddlewareSeq{
		nameServiceClient: nameserver,
		group:             group,
		socket:            socket,
		log:               NewMessageLog(logf),
		clock:             clock,
		recvQueue:         recvQueue,
		stop:              false,
		verbose:           verbose,
	}
	// wait some arbitrary time to start the server.
	// time.Sleep(1 * time.Second)
	go middleware.MiddlewareWork()
	return &middleware, nil
}

func (middleware *MiddlewareSeq) GetGroupID() string {
	return middleware.group.GetMyID()
}

func (middleware *MiddlewareSeq) GetGroupRank() (int, error) {
	return middleware.group.GetRank(middleware.group.GetMyID())
}

func (middleware *MiddlewareSeq) AmITheSequencer() bool {
	rank, err := middleware.GetGroupRank()
	if err != nil {
		return false
	}
	return rank == 0
}

func getID(self string, clock uint64) string {
	return fmt.Sprintf("%s:%d", ShortID(self), clock)
}

func ShortID(self string) string {
	return self[:6]
}

func (middleware *MiddlewareSeq) Send(ctx context.Context, message string) error {
	clock := middleware.clock.Increase()
	src := middleware.group.GetMyID()
	id := getID(src, clock)
	seqMessage := api.MessageSeq{Type: api.MessageType_APPLICATION, Clock: clock, Src: src, Id: id, Data: message}
	if middleware.verbose {
		log.Printf("[SEND] Sending message with id '%s' with clock '%d' and data '%s'\n", seqMessage.GetId(), clock, seqMessage.GetData())
	}
	err := middleware.log.Log(TO_SEND, &seqMessage)
	if err != nil {
		return err
	}
	err = middleware.socket.Send(ctx, &seqMessage)
	if err != nil {
		return err
	}
	err = middleware.log.Log(SENT, &seqMessage)
	if err != nil {
		return err
	}
	return nil
}

func (middleware *MiddlewareSeq) Recv(ctx context.Context) (string, error) {
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			msg := middleware.recvQueue.Pop()
			if msg == nil {
				time.Sleep(1 * time.Millisecond)
				continue
			}
			if msg.GetType() == api.MessageType_SYSTEM {
				middleware.ExecSystemMessage(msg)
			}
			return msg.GetData(), nil
		}
	}
}

func (middleware *MiddlewareSeq) RecvMsg(ctx context.Context) (*api.MessageSeq, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			msg := middleware.recvQueue.Pop()
			if msg == nil {
				time.Sleep(1 * time.Millisecond)
				continue
			}
			return msg, nil
		}
	}
}

func (middleware *MiddlewareSeq) SendSys(ctx context.Context, event SystemEvent) error {
	clock := middleware.clock.Increase()
	src := middleware.group.GetMyID()
	id := getID(middleware.group.GetMyID(), clock)
	sysMsg := api.MessageSeq{Type: api.MessageType_SYSTEM, Clock: clock, Src: src, Id: id, Data: event.String()}
	if middleware.verbose {
		log.Printf("[SEND: ACK] ID: '%s' clock: '%d' data: %s\n", sysMsg.GetId(), clock, sysMsg.GetData())
	}
	err := middleware.log.Log(TO_SEND, &sysMsg)
	if err != nil {
		return err
	}
	err = middleware.socket.TrySend(ctx, &sysMsg)
	if err != nil {
		return err
	}
	err = middleware.log.Log(SENT, &sysMsg)
	if err != nil {
		return err
	}
	return nil
}

func (middleware *MiddlewareSeq) MiddlewareWork() {
	for {
		msg, err := middleware.socket.Recv(context.Background())
		if err != nil {
			if middleware.stop {
				return
			}
			log.Printf("[MiddlewareWork] %v\n", err)
		}
		if middleware.verbose {
			log.Printf("[RECV] Received message from '%s' of type '%s' with clock '%d', id '%s' and data '%s'\n",
				ShortID(msg.GetSrc()), msg.GetType().String(), msg.GetClock(), msg.GetId(), msg.GetData())
			//log.Printf("[RECV] clock update after receiving message '%s' to '%d'\n", msg.GetId(), clock)
		}
		if middleware.AmITheSequencer() && msg.GetSrc() != middleware.group.GetMyID() {
			if middleware.verbose {
				log.Printf("[RELAY] message from '%s' of type '%s' with clock '%d', id '%s' and data '%s'\n",
					ShortID(msg.GetSrc()), msg.GetType().String(), msg.GetClock(), msg.GetId(), msg.GetData())
			}
			if msg.GetType() == api.MessageType_SYSTEM {
				err := middleware.socket.TryRelay(context.Background(), msg, 0)
				if err != nil {
					log.Fatalln(err)
				}
			} else {
				err := middleware.socket.Relay(context.Background(), msg, 0)
				if err != nil {
					log.Fatalln(err)
				}

			}
		}
		err = middleware.log.Log(RECEIVED, msg)
		if err != nil {
			log.Fatalln(err)
		}
		middleware.recvQueue.Push(msg)
	}
}

func (middleware *MiddlewareSeq) Stop() {
	err := middleware.SendSys(context.Background(), EXIT)
	if err != nil {
		log.Printf("[Middleware.Stop()] %v\n", err)
	}
	for {
		msg, err := middleware.RecvMsg(context.Background())
		if err != nil {
			return
		}
		if msg.GetType() != api.MessageType_SYSTEM {
			continue
		} else {
			middleware.ExecSystemMessage(msg)
		}
	}
	//middleware.stop = true
	//middleware.socket.Close()
}

func (middleware *MiddlewareSeq) GetGroupSize() int {
	return len(middleware.group.users)
}

func (middleware *MiddlewareSeq) ExecSystemMessage(message *api.MessageSeq) {
	switch message.GetData() {
	case "EXIT":
		middleware.socket.Close()
		os.Exit(0)
	case "FATAL":
		middleware.socket.Close()
		os.Exit(1)
	default:
		// do nothing
		return
	}
}