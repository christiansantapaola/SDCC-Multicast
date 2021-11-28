package queue

import (
	"fmt"
	api "sdcc/pkg/overlay/clientseq/pb"
)

type AckTable struct {
	size int
}

func NewAckTracker() *AckTable {
	return &AckTable{
		size: 0,
	}
}

func (table *AckTable) Insert(message *api.MessageSeq) error {
	if message.GetType() != api.MessageType_ACK {
		return fmt.Errorf("message is not ACK")
	}
	table.size++
	return nil
}

func (table *AckTable) GetNumAcks() int {
	return table.size
}

type TableEntry struct {
	message *api.MessageSeq
	acks    *AckTable
}

type MessageTable struct {
	table map[string]TableEntry
}

func NewMessageTable() *MessageTable {
	return &MessageTable{
		table: make(map[string]TableEntry),
	}
}

func (table *MessageTable) Insert(message *api.MessageSeq) {
	// update ack
	if message.GetType() == api.MessageType_ACK {
		entry, exists := table.table[message.GetData()]
		if !exists {
			newEntry := TableEntry{message: nil, acks: NewAckTracker()}
			newEntry.acks.Insert(message)
			table.table[message.GetData()] = newEntry
		} else {
			entry.acks.Insert(message)
		}
	} else {
		_, exists := table.table[message.GetId()]
		if !exists {
			newEntry := TableEntry{message: message, acks: NewAckTracker()}
			table.table[message.GetId()] = newEntry
		} else {
			return
		}
	}
}

func (table *MessageTable) GetMessageInfo(messageID string) (*api.MessageSeq, *AckTable, error) {
	entry, exists := table.table[messageID]
	if !exists {
		return nil, nil, fmt.Errorf("ID '%s' not found", messageID)
	}
	return entry.message, entry.acks, nil
}

func (table *MessageTable) Remove(messageID string) (*api.MessageSeq, *AckTable, error) {
	value, exists := table.table[messageID]
	if !exists {
		return nil, nil, fmt.Errorf("ID '%s' not found", messageID)
	}
	msg := value.message
	acks := value.acks
	delete(table.table, messageID)
	return msg, acks, nil
}
