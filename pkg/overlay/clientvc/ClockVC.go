package clientvc

import (
	"sync"
)

type Clock struct {
	clock []uint64
	rank  int
	mutex sync.Mutex
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	} else {
		return b
	}
}

func NewClock(groupSize int, rank int) *Clock {
	clock := make([]uint64, groupSize)
	for i := 0; i < len(clock); i++ {
		clock[i] = 0
	}
	return &Clock{clock: clock, rank: rank}
}

func (cl *Clock) Lock() {
	cl.mutex.Lock()
}

func (cl *Clock) Unlock() {
	cl.mutex.Unlock()
}

func (cl *Clock) Increase(rank int) []uint64 {
	cl.clock[rank] = cl.clock[rank] + 1
	return cl.clock

}

func (cl *Clock) Update(clock []uint64) []uint64 {
	cl.clock[cl.rank] = cl.clock[cl.rank] + 1
	for i := 0; i < len(cl.clock); i++ {
		cl.clock[i] = max(cl.clock[i], clock[i])
	}
	return cl.clock
}

func (cl *Clock) GetClock() []uint64 {
	clock := cl.clock
	return clock
}
