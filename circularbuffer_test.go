package circularbuffer

import (
	"testing"
)

func (b *CircularBuffer) verifyIsEmpty() bool {
	b.lock.Lock()
	defer b.lock.Unlock()

	e := len(b.avail) == 0
	if e {
		if b.pos != b.start {
			panic("desychronized state")
		}
	}
	return e
}

func TestSyncGet(t *testing.T) {
	c := NewCircularBuffer(10)

	for i := 0; i < 4; i++ {
		c.NBPush(i)
	}

	for i := 0; i < 4; i++ {
		v := c.Get().(int)
		if i != v {
			t.Error(v)
		}
	}

	if c.verifyIsEmpty() != true {
		t.Error("not empty")
	}
}

func TestSyncOverflow(t *testing.T) {
	c := NewCircularBuffer(10) // max 9 items in the buffer

	evicted := 0
	c.Evict = func(v interface{}) {
		evicted += 1
		if v.(int) != 0 {
			t.Error(v)
		}
	}

	for i := 0; i < 10; i++ {
		c.NBPush(i)
	}

	for i := 1; i < 10; i++ {
		v := c.Get().(int)
		if i != v {
			t.Error(v)
		}
	}

	if c.verifyIsEmpty() != true {
		t.Error("not empty")
	}

	if evicted != 1 {
		t.Error("wrong evict count", evicted)
	}
}

func TestAsyncGet(t *testing.T) {
	c := NewCircularBuffer(10)

	go func() {
		for i := 0; i < 4; i++ {
			v := c.Get().(int)
			if i != v {
				t.Error(i)
			}
		}

		if c.verifyIsEmpty() != true {
			t.Error("not empty")
		}
	}()

	c.NBPush(0)
	c.NBPush(1)
	c.NBPush(2)
	c.NBPush(3)
}

func TestSyncPop(t *testing.T) {
	c := NewCircularBuffer(10)

	c.NBPush(3)
	c.NBPush(2)
	c.NBPush(1)
	c.NBPush(0)

	for i := 0; i < 4; i++ {
		v := c.Pop().(int)
		if i != v {
			t.Error(v)
		}
	}

	if c.verifyIsEmpty() != true {
		t.Error("not empty")
	}
}

func TestASyncPop(t *testing.T) {
	c := NewCircularBuffer(10)

	go func() {
		for i := 0; i < 4; i++ {
			v := c.Pop().(int)
			if i != v {
				t.Error(v)
			}
		}

		if c.verifyIsEmpty() != true {
			t.Error("not empty")
		}
	}()

	c.NBPush(3)
	c.NBPush(2)
	c.NBPush(1)
	c.NBPush(0)
}
