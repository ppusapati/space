// Package heartbeat implements the 30-second ping/pong + idle-
// close machinery for every WS connection.
//
// → REQ-FUNC-RT-004 (30s ping/pong; idle close).
//
// Tracker is the small per-connection state machine the WS
// reader + writer share. The reader calls TouchPong() on every
// pong frame; the writer's ticker calls ShouldClose() before
// every send to decide if the peer has stopped responding.

package heartbeat

import (
	"sync"
	"time"
)

// Defaults per REQ-FUNC-RT-004.
const (
	DefaultInterval = 30 * time.Second
	// IdleClose triggers when no pong has arrived within
	// 2 × Interval. 60s gives a misbehaving network one full
	// missed-ping window before we hard-close.
	DefaultIdleClose = 60 * time.Second
)

// Tracker tracks pong arrival times.
type Tracker struct {
	mu        sync.Mutex
	lastPong  time.Time
	interval  time.Duration
	idleClose time.Duration
	now       func() time.Time
}

// Config configures Tracker.
type Config struct {
	Interval  time.Duration
	IdleClose time.Duration
	Now       func() time.Time
}

// New builds a Tracker. The clock starts at Now().
func New(cfg Config) *Tracker {
	if cfg.Interval <= 0 {
		cfg.Interval = DefaultInterval
	}
	if cfg.IdleClose <= 0 {
		cfg.IdleClose = DefaultIdleClose
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Tracker{
		lastPong:  cfg.Now(),
		interval:  cfg.Interval,
		idleClose: cfg.IdleClose,
		now:       cfg.Now,
	}
}

// TouchPong records a freshly-arrived pong.
func (t *Tracker) TouchPong() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastPong = t.now()
}

// ShouldClose reports whether the connection has been silent
// past the IdleClose horizon.
func (t *Tracker) ShouldClose() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.now().Sub(t.lastPong) > t.idleClose
}

// Interval returns the configured ping cadence.
func (t *Tracker) Interval() time.Duration { return t.interval }

// IdleHorizon returns the configured idle-close horizon.
func (t *Tracker) IdleHorizon() time.Duration { return t.idleClose }
