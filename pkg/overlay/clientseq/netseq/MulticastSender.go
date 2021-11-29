package netseq

import (
	"context"
	pb "github.com/christiansantapaola/SDCC-Multicast/pkg/overlay/clientseq/pb"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

/*
	MulticastSender si occcupa di gestire multipli ClientSeq.
	Apre le connessioni con il metodo `Connect()` è ritorna solo nel caso in cui tutti le connessioni
	che deve gestire sono attive.
	Il metodo `Send()` si occupa di inviare il messaggio dato a tutti i servizi gestiti.
	Il metodo `Relay()` prende in input il rank di un processo ed invia a tutti i processi tranne a quello
	di cui abbiamo inviato il rank.
	I metodi Try*() si comportano come la loro controparte ma non riprovano l'operazione in caso di errore.
*/

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

/*
	Connect:
		ctx context.Context: contesto per dare un timeout.
	Connect si assicura che tutti i client grpc siano instanziati è connessi.
	ritorna errore in caso di timeout scaduto.
*/

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
			clients, err := NewClientSeq(multicastSender.services[i], multicastSender.opts)
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

/*
	Send:
	Chiama la funziona remota Enqueue su tutti i clientSeq gestiti da multicastSender.
*/

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

/*
	Relay:
	Chiama la funzione remota Enqueue su tutti i clientSeq tranne sul clientSeq con indice `rank`.
*/

func (multicastSender *MulticastSender) Relay(ctx context.Context, message *pb.MessageSeq, rank int) error {
	numClient := len(multicastSender.clients)
	for i := 0; i < numClient; i++ {
		if i != rank {
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
