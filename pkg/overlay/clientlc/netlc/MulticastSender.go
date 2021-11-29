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

/*
	MulticastSender si occcupa di gestire multipli ClientLC.
	Apre le connessioni con il metodo `Connect()` è ritorna solo nel caso in cui tutti le connessioni
	che deve gestire sono attive.
	Il metodo `Send()` si occupa di inviare il messaggio dato a tutti i servizi gestiti.
	I metodi Try*() si comportano come la loro controparte ma non riprovano l'operazione in caso di errore.
*/

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

/*
	Connect:
		ctx context.Context: contesto per dare un timeout.
	Connect si assicura che tutti i client grpc siano instanziati è connessi.
	ritorna errore in caso di timeout scaduto.
*/

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

/*
	Send:
		Invia il messaggio passato da parametro.
	Send invia il messaggio a tutti i client rpc gestiti.
	Se non riesce a inviare un messaggio (fallimento della chiamata rpc) continuera a provare fino a che non riesce.
	I tentativi di comunicazione sono spaziati da un secondo.
*/
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

/*
	TrySend:
		Come Send, solo che in caso di fallimento non riprova a rimandare il messaggio.
*/

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
