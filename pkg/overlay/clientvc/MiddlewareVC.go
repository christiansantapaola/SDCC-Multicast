package clientvc

import (
	"context"
	"fmt"
	"github.com/christiansantapaola/SDCC-Multicast/pkg/nameservice/client"
	"github.com/christiansantapaola/SDCC-Multicast/pkg/nameservice/nameservice"
	"github.com/christiansantapaola/SDCC-Multicast/pkg/overlay/clientvc/netvc"
	api "github.com/christiansantapaola/SDCC-Multicast/pkg/overlay/clientvc/pb"
	"github.com/christiansantapaola/SDCC-Multicast/pkg/overlay/clientvc/queue"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

/*
	MiddlewareVC è l'interfaccia con cui l'utente si interfaccia al sistema.
	Si occupa di:
		- si connette alla rete di overlay
		- inizializza le strutture necessarie alla comunicazione
		- tenere traccia del clock vettoriale, e aggiornalo correttamente.
		- ricevere messaggi/ mandare ack ai messaggi ricevuti / inviare nuovi messaggi.
		- collezionare informazioni sulla rete.
	Il suo utilizzo tuttavia si ferma principalmente per operazioni di Send/Recv.
*/

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

/*
	Instanza un nuovo Middleware:
	self: il mio id all 'interno del gruppo
	groupName: nome del gruppo
	logPath: path del file dove verra scritto il log di rete.
	port: porta di ascolte del server grpc.
	nameserver: client grpc del nameserver
	verbose: modalita di operazione debug.
	trySend: indica di non provare a rimandare i messaggi falliti, utile per il debug.
	dopt: opzioni per i client grpc
	sopt: Opzioni per il server grpc
*/
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

/*
	Send:
		Invia il messaggio input in broadcast a tutti i membri del gruppo noi inclusi.
*/
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

/*
	Recv:
		Ricevi il prossimo messaggio, la semantica è bloccante.
*/

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

/*
	RecvMsg:
	Ritorna l'intero messaggio è non solo il dato.
*/
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

/*
	SendAck:
	Invia l'ack del messaggio di input in broadcast al gruppo noi inclusi.
*/
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

/*
	SendSys:
		Send per messaggi di sistema.
*/
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

/*
	RecvAndUpdate
	Recv un messaggio dalla MulticastSocket ed aggiorna il clock logico vettoriale correttamente.
	L'intera logica dell'evento Receive è qui dentro.
*/
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

/*
	Lavoro di background del MiddlewareLC:
		1. Riceve un nuovo messaggio ed aggiorna il clock
		2. Se è un messaggio invia l'ack ed inseriscilo nella coda di ricezione.
		3. Se è un ack, non inviare un altro ack ma inseriscilo comunque nella coda di ricezione.
*/
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
		}
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
	}
}

/*
	Stop
		indica la volonta di terminare correttamente la comunicazione multicast.
		NB: i messaggi qui vengono inviati senza resend in caso di fallimento perché se
		i peer iniziano a chiudere aspetteremmo indefinitamente per la risposta di un peer che non è piú on.
*/
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
