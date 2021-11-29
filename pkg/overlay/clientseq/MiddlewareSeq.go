package clientseq

import (
	"context"
	"fmt"
	"github.com/christiansantapaola/SDCC-Multicast/pkg/nameservice/client"
	"github.com/christiansantapaola/SDCC-Multicast/pkg/nameservice/nameservice"
	"github.com/christiansantapaola/SDCC-Multicast/pkg/overlay/clientseq/netseq"
	api "github.com/christiansantapaola/SDCC-Multicast/pkg/overlay/clientseq/pb"
	"github.com/christiansantapaola/SDCC-Multicast/pkg/overlay/clientseq/queue"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"time"
)

/*
	MiddlewareSeq è l'interfaccia con cui l'utente si interfaccia al sistema.
	Si occupa di:
		- si connette alla rete di overlay
		- inizializza le strutture necessarie alla comunicazione
		- identificare il sequencer tra i peer all'interno del gruppo.
		- ricevere messaggi/ inviare nuovi messaggi.
		- se il peer attuale è il sequencer, allora rispedire i messaggi ricevuti a tutti gli altri membri del gruppo.
		- collezionare informazioni sulla rete.
	Il suo utilizzo tuttavia si ferma principalmente per operazioni di Send/Recv.
*/

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
	clientSeq         *netseq.MulticastSender
	logPath           string
	log               *MessageLog
	stop              bool
	verbose           bool
}

/*
	Instanzia un nuovo MiddlewareSeq:
	self: il mio id all 'interno del gruppo
	groupName: nome del gruppo
	logPath: path del file dove verra scritto il log di rete.
	port: porta di ascolte del server grpc.
	nameserver: client grpc del nameserver
	verbose: modalita di operazione debug.
	trySend: indica di non provare a rimandare i messaggi falliti, utile per il debug.
	dopt: opzioni per i client grpc
	sopt: Opzioni per il server grpc
	La scelta del sequencer viene scelta in maniera del tutto arbitraria ed univoca, il peer con ID più piccolo in
	ordine lessicografico diventa il sequencer.
	Il sequencer deve avviare un clientSeq extra rivolto verso se stesso per inviare messaggi.
*/
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
	var clientSeq *netseq.MulticastSender = nil
	sequencerServices, err := group.GetSequencerServices()
	if err != nil {
		return nil, err
	}
	if rank == 0 {
		if verbose {
			log.Printf("netseq.Open(port: %d, sopt: %v, services: %v, dopt: %v)\n", port, sopt, services, dopt)
		}
		socket, err = netseq.Open(context.Background(), port, sopt, services, dopt, verbose)
		if err != nil {
			return nil, err
		}
		clientSeq, err = netseq.NewMulticastSender([]net.Addr{sequencerServices}, dopt)
		if err != nil {
			socket.Close()
			return nil, err
		}
		err = clientSeq.Connect(context.Background())
		if err != nil {
			socket.Close()
			return nil, err
		}
	} else {
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
		clientSeq:         clientSeq,
		log:               NewMessageLog(logf),
		clock:             clock,
		recvQueue:         recvQueue,
		stop:              false,
		verbose:           verbose,
	}
	// wait some arbitrary time to start the server.
	// time.Sleep(1 * time.Second)
	time.Sleep(5 * time.Second)
	go middleware.MiddlewareWork()
	return &middleware, nil
}

func (middleware *MiddlewareSeq) GetGroupID() string {
	return middleware.group.GetMyID()
}

func (middleware *MiddlewareSeq) GetGroupRank() (int, error) {
	return middleware.group.GetRank(middleware.group.GetMyID())
}

/*
	Se il rank è 0 (dato dall'ordine lessicografico) allora sono il sequencer.
*/
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

/*
	Send:
		Invio in multicast:
			Il messaggio viene inviato tramite socket al sequencer.
			Il sequencer invece utilizza il suo client extra per comunicare i suoi messaggi alla sua interfaccia
			server.
*/
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
	if middleware.AmITheSequencer() {
		err = middleware.clientSeq.Send(ctx, &seqMessage)
		if err != nil {
			return err
		}
	} else {
		err = middleware.socket.Send(ctx, &seqMessage)
		if err != nil {
			return err
		}
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
	if middleware.AmITheSequencer() {
		err = middleware.clientSeq.TrySend(ctx, &sysMsg)
		if err != nil {
			return err
		}
	} else {
		err = middleware.socket.TrySend(ctx, &sysMsg)
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
	In MiddlewareWork avviene il lavoro principale del middleware:
		1. Ricezione di un messaggio
		2. se sono il sequencer: faccio lo reinvio a tutti tranne che a me.
		3. aggiungo il messaggio appena ricevuto alla coda di ricezione.
*/

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
		if middleware.AmITheSequencer() {
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
