package bufpool

import (
	"sync"
)

var thePool bufPool

// Get retrieves a buffer of the appropriate length from the buffer pool or
// allocates a new one. Get may choose to ignore the pool and treat it as empty.
// Callers should not assume any relation between values passed to Put and the
// values returned by Get.
//
// If no suitable buffer exists in the pool, Get creates one.
func Get(length int) *Buffer {
	return thePool.Get(length)
}

// Put returns a buffer to the buffer pool.
func Put(buf *Buffer) {
	thePool.Put(buf)
}

type bufPool struct {
	pools [steps]sync.Pool
}

func (p *bufPool) Get(length int) *Buffer {
	if length == 0 {
		return nil
	}
	if length > maxPoolSize {
		return NewBuffer(make([]byte, length))
	}

	idx := index(length)
	if bufIface := p.pools[idx].Get(); bufIface != nil {
		buf := bufIface.(*Buffer)
		buf.reset(length)
		return buf
	}

	b := make([]byte, length, indexSize(idx))
	return NewBuffer(b)
}

func (p *bufPool) Put(buf *Buffer) {
	length := buf.Cap()
	if length > maxPoolSize || length < minSize {
		return // drop it
	}

	idx := prevIndex(length)
	p.pools[idx].Put(buf)
}
