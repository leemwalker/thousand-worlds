package gamemap

import (
	"errors"
	"sync"
)

var (
	ErrConcurrencyLimitExceeded = errors.New("map render concurrency limit exceeded")
)

// RendererPool manages resources for map rendering, including memory buffers
// and concurrency limits.
type RendererPool struct {
	bufferPool *sync.Pool
	semaphore  chan struct{}
}

// NewRendererPool creates a pool with a strict concurrency limit.
func NewRendererPool(limit int) *RendererPool {
	return &RendererPool{
		bufferPool: &sync.Pool{
			New: func() interface{} {
				// Pre-allocate a reasonable buffer size (e.g. 4MB) to avoid initial resize
				// 2048x2048 RGBA is 16MB, but WebP is much smaller.
				// We'll let it grow as needed.
				return make([]byte, 0, 4*1024*1024)
			},
		},
		semaphore: make(chan struct{}, limit),
	}
}

// AcquireConcurrencySlot attempts to acquire a slot in the semaphore.
// It returns true if acquired, false if the limit is reached.
// It does NOT block. It implements the "Reject Immediately" strategy.
func (p *RendererPool) AcquireConcurrencySlot() bool {
	select {
	case p.semaphore <- struct{}{}:
		return true
	default:
		return false
	}
}

// ReleaseConcurrencySlot releases a slot in the semaphore.
func (p *RendererPool) ReleaseConcurrencySlot() {
	<-p.semaphore
}

// GetBuffer gets a buffer from the pool with at least the requested capacity.
func (p *RendererPool) GetBuffer(capacity int) []byte {
	buf := p.bufferPool.Get().([]byte)
	if cap(buf) < capacity {
		return make([]byte, 0, capacity)
	}
	return buf[:0]
}

// ReturnBuffer returns a buffer to the pool.
func (p *RendererPool) ReturnBuffer(buf []byte) {
	p.bufferPool.Put(buf)
}
