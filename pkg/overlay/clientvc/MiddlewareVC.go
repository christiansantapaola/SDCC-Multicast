package clientvc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"os"
	"sdcc/pkg/nameservice/client"
	"sdcc/pkg/nameservice/nameservice"
	"sdcc/pkg/overlay/clientvc/netvc"
	api "sdcc/pkg/overlay/clientvc/pb"
	"sdcc/pkg/overlay/clientvc/queue"
	"sync"
	"time"
)

type SystemEvent int

const (
	EXIT  SystemEvent = 0
	FATAL             = 1
	START             = 2
)

func (event SystemEvent) String() string {
	switch event {
	case EXIT:
		return "EXIT"
	case FATAL:
		return "FATAL"
	case START:
		return "START"
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
	case "START":
		return START, nil
	default:
		return 0, fmt.Errorf("'%s' NOT IMPLEMENTED", eventStr)
	}
}

type MiddlewareVC struct {
	nameServiceClient *nameservice.NameServiceClient
	group             *MulticastGroup
	clock             *Clock
	sysRecvQueue      *queue.MessageVCRecvQueue
	recvQueue         *queue.MessageVCRecvQueue
	socket            *netvc.MulticastSocket
	logPath           string
	log               *MessageLog
	sendMutex         sync.Mutex
	stop              bool
	verbose           bool
	trySend           bool
}

func NewMiddlewareLC(self, groupName, logPath string, port int, nameserver *nameservice.NameServiceClient, verbose bool, trySend bool, dopt []grpc.DialOption, sopt []grpc.ServerOption) (*MiddlewareVC, error) {
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
	if verbose {
		log.Printf("group := NewGroup(self: %s, users: %v)\n\n", self, users)
	}
	group := NewGroup(self, users)
	services, ids, _, err := group.GetUsersInfo()
	if err != nil {
		return nil, err
	}
	if verbose {
		log.Printf("netseq.Open(port: %d, sopt: %v, services: %v, dopt: %v)\n", port, sopt, services, dopt)
	}
	socket, err := netvc.Open(context.Background(), port, sopt, services, dopt, verbose)
	if err != nil {
		return nil, err
	}
	if verbose {
		log.Printf("recvQueue := queue.NewMessageVCRecvQueue(ids: %v)\n", ids)
	}
	recvQueue := queue.NewMessageVCRecvQueue(ids, verbose)
	rank, err := group.GetMyRank()
	if err != nil {
		return nil, err
	}
	clock := NewClock(group.GetGroupSize(), rank)
	middleware := MiddlewareVC{
		nameServiceClient: nameserver,
		group:             group,
		socket:            socket,
		log:               NewMessageLog(logf),
		clock:             clock,
		recvQueue:         recvQueue,
		stop:              false,
		verbose:           verbose,
		trySend:           trySend,
	}
	// wait some arbitrary time to start the server.
	time.Sleep(time.Duration(1+rand.Intn(10)) * time.Second)
	go middleware.RecvWork()
	return &middleware, nil
}

func getID(self string, clock []uint64) string {
	return fmt.Sprintf("%s:%v", self, clock)
}

func ShortID(self string) string {
	return self[:6]
}

func (middleware *MiddlewareVC) Send(ctx context.Context, message string) error {
	middleware.clock.Lock()
	defer middleware.clock.Unlock()
	src := middleware.group.GetMyID()
	myrank, err := middleware.group.GetMyRank()
	if err != nil {
		return err
	}
	clock := middleware.clock.Increase(myrank)
	id := getID(src, clock)
	lcMessage := api.MessageVC{Type: api.MessageType_APPLICATION, Clock: clock, Src: src, Id: id, Data: message}
	if middleware.verbose {
		log.Printf("[SEND] Sending message with id '%s' with clock '%d' and data '%s'\n", lcMessage.GetId(), clock, lcMessage.GetData())
	}
	err = middleware.log.Log(TO_SEND, &lcMessage)
	if err != nil {
		return err
	}
	if middleware.trySend {
		err = middleware.socket.TrySend(ctx, &lcMessage)
		if err != nil {
			middleware.log.Log(FAILED_TO_SEND, &lcMessage)
			return err
		}
	} else {
		err = middleware.socket.Send(ctx, &lcMessage)
		if err != nil {
			middleware.log.Log(FAILED_TO_SEND, &lcMessage)
			return err
		}
	}
	err = middleware.log.Log(SENT, &lcMessage)
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (middleware *MiddlewareVC) Recv(ctx context.Context) (string, error) {
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

func (middleware *MiddlewareVC) RecvMsg(ctx context.Context) (*api.MessageVC, error) {
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

func (middleware *MiddlewareVC) SendAck(ctx context.Context, message *api.MessageVC) error {
	middleware.clock.Lock()
	defer middleware.clock.Unlock()
	myrank, err := middleware.group.GetMyRank()
	if err != nil {
		return err
	}
	clock := middleware.clock.Increase(myrank)
	src := middleware.group.GetMyID()
	id := getID(middleware.group.GetMyID(), clock)
	ack := api.MessageVC{Type: api.MessageType_ACK, Clock: clock, Src: src, Id: id, Data: message.GetId()}
	if middleware.verbose {
		log.Printf("[SEND: ACK] ID: '%s' clock: '%d' data: %s\n", ack.GetId(), clock, ack.GetData())
	}
	err = middleware.log.Log(TO_SEND_ACK, &ack)
	if err != nil {
		return err
	}
	if middleware.trySend {
		err = middleware.socket.TrySend(ctx, &ack)
		if err != nil {
			errlog := middleware.log.Log(FAILED_TO_SEND, &ack)
			if errlog != nil {
				return err
			}
			return err
		}
	} else {
		err = middleware.socket.Send(ctx, &ack)
		if err != nil {
			errlog := middleware.log.Log(FAILED_TO_SEND, &ack)
			if errlog != nil {
				return err
			}
			return err
		}
	}
	err = middleware.log.Log(SENT_ACK, &ack)
	if err != nil {
		return err
	}
	return nil
}

func (middleware *MiddlewareVC) SendSys(ctx context.Context, event SystemEvent, try bool) error {
	middleware.clock.Lock()
	defer middleware.clock.Unlock()
	myrank, err := middleware.group.GetMyRank()
	if err != nil {
		return err
	}
	clock := middleware.clock.Increase(myrank)
	src := middleware.group.GetMyID()
	id := getID(middleware.group.GetMyID(), clock)
	sysMsg := api.MessageVC{Type: api.MessageType_SYSTEM, Clock: clock, Src: src, Id: id, Data: event.String()}
	if middleware.verbose {
		log.Printf("[SEND: ACK] ID: '%s' clock: '%d' data: %s\n", sysMsg.GetId(), clock, sysMsg.GetData())
	}
	err = middleware.log.Log(TO_SEND, &sysMsg)
	if err != nil {
		return err
	}
	if try {
		err = middleware.socket.TrySend(ctx, &sysMsg)
		if err != nil {
			return err
		}
	} else {
		err = middleware.socket.Send(ctx, &sysMsg)
		if err != nil {
			return err
		}
	}
	err = middleware.log.Log(SENT, &sysMsg)
	if err != nil {
		return err
	}
	return nil
}

func (middleware *MiddlewareVC) RecvAndUpdate() *api.MessageVC {
	middleware.clock.Lock()
	defer middleware.clock.Unlock()
	msg := middleware.socket.NonBlockingRecv()
	if msg == nil {
		return nil
	} else {
		middleware.clock.Update(msg.GetClock())
		return msg
	}
}

func (middleware *MiddlewareVC) RecvWork() {
	for {
		var err error
		msg := middleware.RecvAndUpdate()
		if msg == nil {
			time.Sleep(1 * time.Millisecond)
			continue
		}
		if middleware.verbose {
			log.Printf("[RECV] Received message from '%s' of type '%s' with clock '%d', id '%s' and data '%s' -> new clock :%d\n",
				ShortID(msg.GetSrc()), msg.GetType().String(), msg.GetClock(), msg.GetId(), msg.GetData(), middleware.clock.GetClock())
			//log.Printf("[RECV] clock update after receiving message '%s' to '%d'\n", msg.GetId(), clock)
		}
		//if msg.GetType() == api.MessageType_SYSTEM {
		//	middleware.ExecSystemMessage(msg)
		//	continue
		//}
		if msg.GetType() == api.MessageType_ACK {
			err = middleware.log.Log(RECEIVED_ACK, msg)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			err = middleware.log.Log(RECEIVED, msg)
			if err != nil {
				log.Fatalln(err)
			}
		}
		err = middleware.recvQueue.Push(msg)
		if err != nil {
			log.Fatalln(err)
		}
		if msg.GetType() != api.MessageType_ACK {
			err = middleware.SendAck(context.Background(), msg)
			if err != nil {
				log.Fatalln(err)
			}
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func (middleware *MiddlewareVC) Stop() {
	err := middleware.SendSys(context.Background(), EXIT, true)
	if err != nil {
		log.Println(err)
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

func (middleware *MiddlewareVC) GetGroupSize() int {
	return len(middleware.group.users)
}

func (middleware *MiddlewareVC) ExecSystemMessage(message *api.MessageVC) {
	switch message.GetData() {
	case "EXIT":
		middleware.socket.Close()
		os.Exit(0)
	case "FATAL":
		middleware.socket.Close()
		os.Exit(1)
	case "START":
		// Do Nothing
		return
	default:
		// do nothing
		return
	}
}

func (middleware *MiddlewareVC) GetGroupID() string {
	return middleware.group.GetMyID()
}

func (middleware *MiddlewareVC) GetShortID() string {
	return middleware.group.GetShortID()
}

func (middleware *MiddlewareVC) GetRank() (int, error) {
	return middleware.group.GetMyRank()
}
