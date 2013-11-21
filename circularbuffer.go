// Circular buffer implementation.
// Features:
//  - no memory allocations during push/pop
//  - nonblocking push (ie: evict old data)
//  - pop item from the top
//  - concurrent (through locking)
//
// Semantics of eviction:
//  - empty Evict callback - return evicted item when
//    calling NBPush.
//  - Evict callback present - return nil to NBPush
//    and call Evict() inline.
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
	avail  chan bool // poor man's semaphore. len(avail) is always equal to (size + pos - start) % size
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

// Nonblocking push. If the Evict callback is not set returns the
// evicted item (if any), otherwise nil.
func (b *CircularBuffer) NBPush(v interface{}) interface{} {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.buffer[b.pos] = v
	b.pos = (b.pos + 1) % b.size
	if b.pos == b.start {
		// Remove old item from the bottom of the stack to
		// free the space for the new one. This doesn't change
		// the length of the stack, so no need to touch avail.
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
	b.buffer[b.pos] = nil
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
