// Package backpressure caps per-connection-per-topic message
// rate and drops the OLDEST queued message on overflow.
//
// → REQ-FUNC-RT-003 (per-connection rate cap; drop-oldest on
//                     overflow with metric).
// → REQ-FUNC-RT-006 (chetana_rt_dropped_total{reason="overflow"}
//                     increments under load).
//
// Algorithm: a per-(connection, topic) ring buffer (FIFO with
// fixed capacity). Producers Push; consumers Pop. When the buffer
// is full, Push evicts the OLDEST entry, increments the dropped
// counter, and inserts the new message at the tail.

package backpressure

import (
	"errors"
	"sync"
)

// DefaultCapacity is the per-(connection, topic) message buffer
// size. 1000 mirrors the chetana RT design's per-topic budget;
// sized so a 30-second slow client doesn't immediately drop on
// every message.
const DefaultCapacity = 1000

// Buffer is a fixed-capacity FIFO with drop-oldest semantics.
type Buffer struct {
	mu      sync.Mutex
	cap     int
	items   []any
	dropped int64
}

// NewBuffer returns a Buffer with `capacity` slots. capacity <= 0
// → DefaultCapacity.
func NewBuffer(capacity int) *Buffer {
	if capacity <= 0 {
		capacity = DefaultCapacity
	}
	return &Buffer{
		cap:   capacity,
		items: make([]any, 0, capacity),
	}
}

// Push enqueues `msg`. Returns true when the message was added
// without dropping anything; false when the oldest message was
// evicted to make room.
func (b *Buffer) Push(msg any) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.items) >= b.cap {
		// drop-oldest
		copy(b.items, b.items[1:])
		b.items = b.items[:len(b.items)-1]
		b.dropped++
		b.items = append(b.items, msg)
		return false
	}
	b.items = append(b.items, msg)
	return true
}

// Pop dequeues the oldest message. Returns ErrEmpty when empty.
func (b *Buffer) Pop() (any, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.items) == 0 {
		return nil, ErrEmpty
	}
	msg := b.items[0]
	copy(b.items, b.items[1:])
	b.items = b.items[:len(b.items)-1]
	return msg, nil
}

// Len returns the current buffer occupancy.
func (b *Buffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.items)
}

// DroppedCount returns the cumulative dropped-oldest count.
// The cmd-layer wires this to the
// `chetana_rt_dropped_total{reason="overflow"}` metric.
func (b *Buffer) DroppedCount() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.dropped
}

// Capacity returns the configured capacity.
func (b *Buffer) Capacity() int { return b.cap }

// ErrEmpty is returned by Pop on an empty buffer.
var ErrEmpty = errors.New("backpressure: buffer empty")
