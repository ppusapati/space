package backpressure

import (
	"errors"
	"testing"
)

func TestNewBuffer_DefaultsToDefaultCapacity(t *testing.T) {
	b := NewBuffer(0)
	if b.Capacity() != DefaultCapacity {
		t.Errorf("default cap: %d", b.Capacity())
	}
}

func TestPushPop_FIFO(t *testing.T) {
	b := NewBuffer(3)
	for i := 0; i < 3; i++ {
		if !b.Push(i) {
			t.Errorf("push %d should not drop", i)
		}
	}
	for i := 0; i < 3; i++ {
		v, err := b.Pop()
		if err != nil {
			t.Fatalf("pop: %v", err)
		}
		if v != i {
			t.Errorf("pop order: got %v want %d", v, i)
		}
	}
}

// REQ-FUNC-RT-003: drop-oldest on overflow + counter increments.
func TestPush_OverflowDropsOldest(t *testing.T) {
	b := NewBuffer(2)
	b.Push("a")
	b.Push("b")
	if got := b.Push("c"); got {
		t.Error("expected push to report drop")
	}
	if b.DroppedCount() != 1 {
		t.Errorf("dropped: %d", b.DroppedCount())
	}
	// Buffer should hold ["b", "c"] now.
	v1, _ := b.Pop()
	v2, _ := b.Pop()
	if v1 != "b" || v2 != "c" {
		t.Errorf("got %v %v want b c", v1, v2)
	}
}

func TestPush_DroppedCountAccumulates(t *testing.T) {
	b := NewBuffer(1)
	b.Push("x") // fills
	for i := 0; i < 5; i++ {
		b.Push(i)
	}
	if got := b.DroppedCount(); got != 5 {
		t.Errorf("dropped: got %d want 5", got)
	}
}

func TestPop_EmptyReturnsErr(t *testing.T) {
	b := NewBuffer(2)
	if _, err := b.Pop(); !errors.Is(err, ErrEmpty) {
		t.Errorf("got %v want ErrEmpty", err)
	}
}

func TestLen_TracksOccupancy(t *testing.T) {
	b := NewBuffer(5)
	b.Push(1)
	b.Push(2)
	b.Push(3)
	if b.Len() != 3 {
		t.Errorf("len: %d", b.Len())
	}
	_, _ = b.Pop()
	if b.Len() != 2 {
		t.Errorf("after pop: %d", b.Len())
	}
}

func TestPush_ConcurrentSafe(t *testing.T) {
	b := NewBuffer(1000)
	done := make(chan bool, 4)
	for i := 0; i < 4; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				b.Push(j)
			}
			done <- true
		}()
	}
	for i := 0; i < 4; i++ {
		<-done
	}
	// All 400 pushes fit; no drops.
	if b.Len() != 400 {
		t.Errorf("len: %d want 400", b.Len())
	}
	if b.DroppedCount() != 0 {
		t.Errorf("dropped: %d", b.DroppedCount())
	}
}
