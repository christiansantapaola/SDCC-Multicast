package clientseq

import (
	"fmt"
	"os"
	api "sdcc/pkg/overlay/clientseq/pb"
	"sync"
	"time"
)

type MessageLog struct {
	out   *os.File
	log   map[string]logEntry
	mutex sync.Mutex
}

type Status int

const (
	TO_SEND  Status = 0
	SENT            = 1
	RECEIVED        = 2
)

func (status Status) String() string {
	switch status {
	case TO_SEND:
		return "TO_SEND"
	case SENT:
		return "SENT"
	case RECEIVED:
		return "RECEIVED"
	default:
		return "NOT IMPLEMENTED"
	}
}

type logEntry struct {
	Status  Status
	Message *api.MessageSeq
}

func NewMessageLog(out *os.File) *MessageLog {
	return &MessageLog{
		out: out,
		log: make(map[string]logEntry),
	}
}

func (log *MessageLog) WriteLog(status Status, message *api.MessageSeq) error {
	str := fmt.Sprintf("[%s] STATUS: %s MESSAGE: %s\n", time.Now().Format("2006-01 02-15:04:05"), status.String(), message.GetId())
	_, err := log.out.WriteString(str)
	if err != nil {
		return err
	}
	return nil

}

func (log *MessageLog) Log(status Status, message *api.MessageSeq) error {
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

func (log *MessageLog) GetMessageToSend() []*api.MessageSeq {
	log.mutex.Lock()
	defer log.mutex.Unlock()
	toSend := make([]*api.MessageSeq, 0)
	for _, val := range log.log {
		if val.Status == TO_SEND {
			toSend = append(toSend, val.Message)
		}
	}
	return toSend
}
