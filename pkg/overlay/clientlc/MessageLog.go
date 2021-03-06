package clientlc

import (
	"fmt"
	api "github.com/christiansantapaola/SDCC-Multicast/pkg/overlay/clientlc/pb"
	"os"
	"sync"
	"time"
)

/*
	MessageLog è una struttura che scrive il log di rete delle operazioni da compiere/compiute.
	Pensato per essere utilizzato come un log look-ahead.
	Prima dico quale operazione voglio eseguire la eseguo, poi termino l'operazione con l'evento terminate:
	- TO_SEND -> (FAILED_TO_SEND/SENT)
	- TO_SEND_ACK -> (FAILED_TO_SEND/SENT_ACK)
	- RECEIVED / RECEIVED_ACK
	Il log viene scritto su un file indicato alla creazione del log.
*/

type MessageLog struct {
	out   *os.File
	log   map[string]logEntry
	mutex sync.Mutex
}

type Status int

const (
	TO_SEND        Status = 0
	SENT                  = 1
	RECEIVED              = 2
	FAILED_TO_SEND        = 3
	TO_SEND_ACK           = 4
	SENT_ACK              = 5
	RECEIVED_ACK          = 6
)

func (status Status) String() string {
	switch status {
	case TO_SEND:
		return "TO_SEND"
	case SENT:
		return "SENT"
	case RECEIVED:
		return "RECEIVED"
	case FAILED_TO_SEND:
		return "FAILED_TO_SEND"
	case TO_SEND_ACK:
		return "TO_SEND_ACK"
	case SENT_ACK:
		return "SENT_ACK"
	case RECEIVED_ACK:
		return "RECEIVED_ACK"
	default:
		return "NOT IMPLEMENTED"
	}
}

type logEntry struct {
	Status  Status
	Message *api.MessageLC
}

func NewMessageLog(out *os.File) *MessageLog {
	return &MessageLog{
		out: out,
		log: make(map[string]logEntry),
	}
}

func (log *MessageLog) WriteLog(status Status, message *api.MessageLC) error {
	var str string
	if status == TO_SEND_ACK || status == SENT_ACK || status == RECEIVED_ACK {
		str = fmt.Sprintf("[%s] STATUS: %s ID: %s TO MESSAGE: %s\n", time.Now().Format("2006-01 02-15:04:05"), status.String(), message.GetId(), message.GetData())
	} else {
		str = fmt.Sprintf("[%s] STATUS: %s MESSAGE ID: %s DATA: %s\n", time.Now().Format("2006-01 02-15:04:05"), status.String(), message.GetId(), message.GetData())
	}
	_, err := log.out.WriteString(str)
	if err != nil {
		return err
	}
	return nil

}

func (log *MessageLog) Log(status Status, message *api.MessageLC) error {
	log.mutex.Lock()
	if val, exists := log.log[message.GetId()]; !exists {
		newEntry := logEntry{Message: message, Status: status}
		log.log[message.GetId()] = newEntry
	} else {
		val.Status = status
	}
	log.mutex.Unlock()
	err := log.WriteLog(status, message)
	if err != nil {
		return err
	}
	return nil
}
