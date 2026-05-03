package heartbeat

import (
	"sync"
	"testing"
	"time"
)

func TestNew_Defaults(t *testing.T) {
	tr := New(Config{})
	if tr.Interval() != DefaultInterval {
		t.Errorf("interval default: %v", tr.Interval())
	}
	if tr.IdleHorizon() != DefaultIdleClose {
		t.Errorf("idle default: %v", tr.IdleHorizon())
	}
}

func TestShouldClose_FreshPongIsAlive(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	clk := &testClock{t: now}
	tr := New(Config{Now: clk.now})
	if tr.ShouldClose() {
		t.Error("fresh tracker should not close")
	}
}

func TestShouldClose_FiresPastHorizon(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	clk := &testClock{t: now}
	tr := New(Config{Now: clk.now, IdleClose: 60 * time.Second})

	// Advance past the horizon.
	clk.advance(61 * time.Second)
	if !tr.ShouldClose() {
		t.Error("past horizon should close")
	}
}

func TestTouchPong_ResetsHorizon(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	clk := &testClock{t: now}
	tr := New(Config{Now: clk.now, IdleClose: 60 * time.Second})

	clk.advance(50 * time.Second)
	tr.TouchPong()
	clk.advance(50 * time.Second) // 100s since the original start, but only 50s since the pong
	if tr.ShouldClose() {
		t.Error("recent pong should keep alive")
	}
}

func TestTouchPong_ConcurrentSafe(t *testing.T) {
	tr := New(Config{})
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				tr.TouchPong()
				_ = tr.ShouldClose()
			}
		}()
	}
	wg.Wait()
}

type testClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *testClock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *testClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}
