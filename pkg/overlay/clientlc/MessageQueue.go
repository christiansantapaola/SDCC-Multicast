package clientlc

type MessageQueue []*Message

func (queue MessageQueue) Len() int      { return len(queue) }
func (queue MessageQueue) IsEmpty() bool { return len(queue) == 0 }
func (queue MessageQueue) Less(i, j int) bool {
	if queue[i].GetClock() == queue[j].GetClock() {
		return queue[i].GetSrc() == queue[j].GetSrc()
	} else {
		return queue[i].GetClock() < queue[j].GetClock()
	}
}
func (queue MessageQueue) Swap(i, j int) { queue[i], queue[j] = queue[j], queue[i] }

func (queue *MessageQueue) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*queue = append(*queue, x.(*Message))
}

func (queue *MessageQueue) Pop() interface{} {
	old := *queue
	n := len(old)
	if n == 0 {
		return nil
	}
	x := old[n-1]
	*queue = old[0 : n-1]
	return x
}

func (queue MessageQueue) Peek() interface{} {
	n := len(queue)
	return queue[n - 1]
}
