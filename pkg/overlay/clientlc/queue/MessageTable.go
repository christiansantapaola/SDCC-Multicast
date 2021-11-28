package queue

import (
	"fmt"
	"log"
	api "sdcc/pkg/overlay/clientlc/pb"
	"strconv"
	"strings"
	"sync"
)

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

type TableEntry struct {
	msg   *api.MessageLC
	clock uint64
	src   string
	acks  *AckTable
}

type MessageTable struct {
	table          map[string]TableEntry
	pendingMessage map[string]*api.MessageLC
	groupSize      int
	mutex          sync.Mutex
}

func NewMessageTable(groupsize int) *MessageTable {
	return &MessageTable{
		table:          make(map[string]TableEntry),
		pendingMessage: make(map[string]*api.MessageLC),
		groupSize:      groupsize,
	}
}

func ParseAck(messageId string) (string, uint64, error) {
	var src string
	var clock uint64
	toParse := strings.Split(messageId, ":")
	if len(toParse) != 2 {
		return "", 0, fmt.Errorf("messageId '%s' splitted into: %v\n", messageId, toParse)
	}
	src = toParse[0]
	clock, err := strconv.ParseUint(toParse[1], 10, 64)
	if err != nil {
		return "", 0, nil
	}
	return src, clock, nil
}

func less(src1, src2 string, clock1, clock2 uint64) bool {
	if clock1 == clock2 {
		return src1 < src2
	} else {
		return clock1 < clock2
	}
}

func (table *MessageTable) Insert(message *api.MessageLC) error {
	table.mutex.Lock()
	defer table.mutex.Unlock()
	if message.GetType() == api.MessageType_ACK {
		entry, exists := table.table[message.GetData()]
		if !exists {
			log.Printf("ACK ARRIVED BEFORE MESSAGE WITH ID '%s' AND DATA '%s'\n", message.GetId(), message.GetData())
			src, clock, err := ParseAck(message.GetData())
			if err != nil {
				//log.Printf("%s: %v\n", "ParseAck() failed", err)
				return err
			}
			//log.Printf("ID PARSED INTO '%s' AND '%d'\n", src, clock)
			newEntry := TableEntry{msg: nil, acks: NewAckTracker(), src: src, clock: clock}
			//log.Printf("NEW ENTRY: %v\n", entry)
			newEntry.acks.Insert(message.GetId())
			//log.Printf("INSERT ACK: %d", newEntry.acks.GetNumAcks())
			table.table[message.GetData()] = newEntry
			//log.Printf("INSERT NEW ENTRY: %v\n", table.table[message.GetData()])
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

func (table *MessageTable) IsReady(messageID string) (bool, error) {
	table.mutex.Lock()
	defer table.mutex.Unlock()
	//log.Printf("IsReady(%s)\n", messageID)
	entry, exists := table.table[messageID]
	if !exists {
		//log.Printf("entry does not exists\n")
		return false, fmt.Errorf("ID '%s' not found", messageID)
	}
	//log.Printf("table.table[messageID] = %v\n", table.table[messageID])
	ackReceived := entry.acks.GetNumAcks() == table.groupSize
	//log.Printf("%d == %d : %t", entry.acks.GetNumAcks(), table.groupSize, ackReceived)
	isMin := true
	for _, val := range table.table {
		//log.Printf("val: %v\n", val)
		if val.msg == nil && !less(entry.src, val.src, entry.clock, val.clock) {
			//log.Printf("IS PENDING AND LESS THAN ENTRY", val)
			isMin = false
		}
	}
	//fmt.Printf("%t && %t == %t\n", ackReceived, isMin, ackReceived && isMin)
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
