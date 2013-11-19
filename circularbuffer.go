// Circular buffer implementation.
// Features:
//  - no memory allocations during push/pop
//  - nonblocking push (ie: evict old data)
//  - pop item from the top
//  - concurrent (through locking)
//
package circularbuffer

import (
	"sync"
)

type StackPusher interface {
	NBPush(interface{}) interface{}
}

type StackGetter interface {
	Get() interface{}
	Pop() interface{}
}

type CircularBuffer struct {
	start  uint // idx of first used cell
	pos    uint // idx of first unused cell
	buffer []interface{}
	size   uint
	avail  chan bool // poor man's semaphore
	lock   sync.Mutex
}

// Create CircularBuffer object with a prealocated buffer of a given size.
func NewCircularBuffer(size uint) *CircularBuffer {
	return &CircularBuffer{
		buffer: make([]interface{}, size),
		size: size,
		avail: make(chan bool, size),
	}
}

// Nonblocking push. Returns the evicted item or nil.
func (b *CircularBuffer) NBPush(v interface{}) interface{} {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.buffer[b.pos] = v
	b.pos = (b.pos + 1) % b.size
	if b.pos == b.start {
		evictv := b.buffer[b.start]
		b.buffer[b.start] = nil
		b.start = (b.start + 1) % b.size
		return evictv
	} else {
		select {
		case b.avail <- true:
		default:
			panic("Sending to avail channel must never block")
		}
		return nil
	}
}

// Get an item from the beginning of the queue (oldest), blocking.
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

// Blocking pop an item from the end of the queue (newest), blocking.
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

// Is the buffer empty?
func (b *CircularBuffer) Empty() bool {
	// b.avail is a channel, no need for a lock
	return len(b.avail) == 0
}

// Length of the buffer
func (b *CircularBuffer) Length() int {
	// b.avail is a channel, no need for a lock
	return len(b.avail)
}
