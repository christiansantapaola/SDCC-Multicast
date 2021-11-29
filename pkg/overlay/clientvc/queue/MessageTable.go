package queue

import (
	"fmt"
	api "github.com/christiansantapaola/SDCC-Multicast/pkg/overlay/clientvc/pb"
	"strconv"
	"strings"
	"sync"
)

/*
	La messageTable nasce con l'intezione di tenere traccia dello stato dei messaggi arrivati, deve:
		1. tenere traccia di tutti i messaggi arrivati che non sono ancora stati ricevuti.
		2. deve saper indicare se un messaggio è pronto per essere rilasciato o no.
	La messageTable si avvale di una struttura di supporto chiamata AckTable che tiene traccia di tutti gli ack
	univoci che sono stati ricevuti per un particolare messaggio.
*/

type AckTable struct {
	ids []string
}

func NewAckTracker() *AckTable {
	return &AckTable{
		ids: make([]string, 0),
	}
}

func IsIn(str string, array []string) bool {
	for _, elem := range array {
		if str == elem {
			return true
		}
	}
	return false
}

func (table *AckTable) Insert(messageId string) {
	if !IsIn(messageId, table.ids) {
		table.ids = append(table.ids, messageId)
	}
}

func (table *AckTable) GetNumAcks() int {
	return len(table.ids)
}

/*
	La messageTable è un oggetto che si occupa di tenere traccia dei messaggi arrivati e del loro stato:
	è implementata come una mappa ID messaggio -> Info sul messaggio.
*/

type TableEntry struct {
	msg   *api.MessageVC
	clock []uint64
	src   string
	acks  *AckTable
}

type MessageTable struct {
	table          map[string]TableEntry
	pendingMessage map[string]*api.MessageVC
	groupSize      int
	mutex          sync.Mutex
}

func NewMessageTable(groupsize int) *MessageTable {
	return &MessageTable{
		table:          make(map[string]TableEntry),
		pendingMessage: make(map[string]*api.MessageVC),
		groupSize:      groupsize,
	}
}

/*
	ParseAck:
		dato un id di un messaggio ne ricava il mittente e il suo clock.
		Un ID di un messaggio è della forma "mittente:clock"
		es: "user:[0 1 2 1 0 0 1]"
*/
func ParseAck(messageId string) (string, []uint64, error) {
	var src string
	var clock []uint64
	toParse := strings.Split(messageId, ":")
	if len(toParse) != 2 {
		return "", nil, fmt.Errorf("messageId '%s' splitted into: %v\n", messageId, toParse)
	}
	src = toParse[0]
	lenght := len(toParse[1])
	res := strings.Split(toParse[1][1:lenght-1], " ")
	for i := 0; i < len(clock); i++ {
		val, err := strconv.ParseUint(res[i], 10, 64)
		if err != nil {
			return "", nil, err
		}
		clock = append(clock, val)
	}
	return src, clock, nil
}

func less(clock1, clock2 []uint64) bool {
	groupSize := len(clock1)
	if len(clock1) != len(clock2) {
		return false
	}
	for k := 0; k < groupSize; k++ {
		if clock1[k] > clock2[k] {
			return false
		}
	}
	return true

}

/*
	Il metodo Insert() rappresenta il cuore della message table.
	Se un messaggio arriva ho 4 possibili casi dati dalle risposte a due domande:
		1. Il messaggio è un ack?
		2. Il messaggio è gia stato registrato?
	Se il messaggio non è un ack e non è stato registrato:
		allora creo una nuova entry nella tabella con le informazioni del messaggio e ritorno
	Se il messaggio è un ack e il messaggio a cui risponde è registrato:
		aggiorno le informazioni dell messaggio, indicando che ha ricevuto un nuovo ack.
	Se il messaggio è un ack e il messaggio a cui risponde non è registrato:
		ho un caso in cui mi è arrivato prima l'ack di un messaggio del messaggio stesso.
		In questo caso devo creare nella tabella una entry in cui dico: in arrivo c'è un messaggio con clock
		dato dall'ack, ma non ho ancora il messaggio stesso.
	Se il messaggio non è un ack ed é stato gia registrato:
		Questo caso risulta dal caso di sopra, quindi aggiorno l'entry gia esistente con il nuovo messaggio.
*/

func (table *MessageTable) Insert(message *api.MessageVC) error {
	table.mutex.Lock()
	defer table.mutex.Unlock()
	if message.GetType() == api.MessageType_ACK {
		entry, exists := table.table[message.GetData()]
		if !exists {
			src, clock, err := ParseAck(message.GetData())
			if err != nil {
				//log.Printf("%s: %v\n", "ParseAck() failed", err)
				return err
			}
			newEntry := TableEntry{msg: nil, acks: NewAckTracker(), src: src, clock: clock}
			newEntry.acks.Insert(message.GetId())
			table.table[message.GetData()] = newEntry
		} else {
			entry.acks.Insert(message.GetId())
		}
	} else {
		entry, exists := table.table[message.GetId()]
		if !exists {
			newEntry := TableEntry{msg: message, acks: NewAckTracker()}
			table.table[message.GetId()] = newEntry
		} else {
			entry.msg = message
			table.table[message.GetId()] = entry

		}
	}
	return nil
}

/*
	IsReady:
		La funzione IsReady si occupa di stabilire se un messaggio è pronto per essere rilasciato all utente oppure no.
		Controlla che il messaggio input abbia ricevuto tutti gli ack e che non ci siano messaggi in arrivo
		con clock minore del suo.
*/
func (table *MessageTable) IsReady(messageID string) (bool, error) {
	table.mutex.Lock()
	defer table.mutex.Unlock()
	entry, exists := table.table[messageID]
	if !exists {
		return false, fmt.Errorf("ID '%s' not found", messageID)
	}
	ackReceived := entry.acks.GetNumAcks() == table.groupSize
	isMin := true
	for _, val := range table.table {
		if val.msg == nil && !less(entry.clock, val.clock) {
			isMin = false
		}
	}
	return ackReceived && isMin, nil
}

func (table *MessageTable) Remove(messageID string) error {
	table.mutex.Lock()
	defer table.mutex.Unlock()
	_, exists := table.table[messageID]
	if !exists {
		return fmt.Errorf("ID '%s' not found", messageID)
	}
	delete(table.table, messageID)
	return nil
}
