package circularbuffer

import (
	"sync"
)

type CircularBuffer struct {
	start  uint // idx of first used cell
	pos    uint // idx of first unused cell
	buffer []interface{}
	size   uint
	avail  chan bool // poor man's semaphore
	lock   sync.Mutex
	Evict  func(interface{})
}

func NewCircularBuffer(size uint) *CircularBuffer {
	b := new(CircularBuffer)
	b.buffer = make([]interface{}, size)
	b.size = size
	b.avail = make(chan bool, size)
	return b
}

func (b *CircularBuffer) NBPush(v interface{}) {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.buffer[b.pos] = v
	b.pos = (b.pos + 1) % b.size
	if b.pos == b.start {
		evictv := b.buffer[b.start]
		b.buffer[b.start] = nil
		b.start = (b.start + 1) % b.size
		if b.Evict != nil {
			b.Evict(evictv)
		}
	} else {
		select {
		case b.avail <- true:
		default:
			panic("Sending to avail channel must never block")
		}
	}
}

func (b *CircularBuffer) Get() interface{} {
	_ = <-b.avail

	b.lock.Lock()
	defer b.lock.Unlock()

	if b.start == b.pos {
		panic("Trying to get from empty buffer")
	}

	v := b.buffer[b.start]
	b.start = (b.start + 1) % b.size

	return v
}

func (b *CircularBuffer) Pop() interface{} {
	_ = <-b.avail

	b.lock.Lock()
	defer b.lock.Unlock()

	if b.start == b.pos {
		panic("Can't pop from empty buffer")
	}

	b.pos = (b.size + b.pos - 1) % b.size
	v := b.buffer[b.pos]
	b.buffer[b.pos] = nil

	return v
}

func (b *CircularBuffer) Empty() bool {
	return len(b.avail) == 0
}
