package clientseq

import (
	"sync"
)

/*
	Implementazione di un clock scalare thread-safe.
	In questa implementazione il suo scopo è di dare un numero univoco ad ogni messaggio da inviare.
*/

type Clock struct {
	clock uint64
	mutex sync.Mutex
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	} else {
		return b
	}
}

func NewClock() *Clock {
	return &Clock{clock: 0}
}

func (cl *Clock) Increase() uint64 {
	cl.mutex.Lock()
	cl.clock = cl.clock + 1
	clock := cl.clock
	cl.mutex.Unlock()
	return clock

}

func (cl *Clock) Update(clock uint64) uint64 {
	cl.mutex.Lock()
	cl.clock = max(cl.clock+1, clock)
	clockCpy := cl.clock
	cl.mutex.Unlock()
	return clockCpy
}

func (cl *Clock) GetClock() uint64 {
	cl.mutex.Lock()
	clock := cl.clock
	cl.mutex.Unlock()
	return clock
}
