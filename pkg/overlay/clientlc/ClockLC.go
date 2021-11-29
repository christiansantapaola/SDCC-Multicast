package clientlc

import (
	"sync"
)

/*
	Semplice struttura dati che tiene traccia in maniera thread-safe del clock scalare locale.
	IL suo funzionamento consiste nel:
		1. prendere il lock con Lock()
		2. eseguire l'aggiornamento con Increase() / Update()
		3. rilasciare il lock con Unlock()
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

func (cl *Clock) Lock() {
	cl.mutex.Lock()
}

func (cl *Clock) Unlock() {
	cl.mutex.Unlock()
}

func (cl *Clock) Increase() uint64 {
	cl.clock = cl.clock + 1
	clock := cl.clock
	return clock

}

func (cl *Clock) Update(clock uint64) uint64 {
	cl.clock = max(cl.clock+1, clock)
	clockCpy := cl.clock
	return clockCpy
}

func (cl *Clock) GetClock() uint64 {
	clock := cl.clock
	return clock
}
