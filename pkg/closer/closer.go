// Package closer collects shutdown hooks and runs them in LIFO order
// with a hard deadline. It is a tiny replacement for the
// uber-fx-style lifecycle bookkeeping every service needs, written so
// it has no third-party dependencies.
package closer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Func is a shutdown function. It receives a context bounded by the
// caller's deadline.
type Func func(context.Context) error

// Closer aggregates Funcs and runs them in LIFO order on Run.
type Closer struct {
	mu    sync.Mutex
	funcs []labelled
}

type labelled struct {
	label string
	fn    Func
}

// Add registers `fn` for shutdown. `label` is included in the joined
// error and is useful to diagnose hangs.
func (c *Closer) Add(label string, fn Func) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.funcs = append(c.funcs, labelled{label: label, fn: fn})
}

// Run invokes every registered Func in LIFO order with a per-call
// budget of `timeout`. Errors are joined with errors.Join.
func (c *Closer) Run(ctx context.Context, timeout time.Duration) error {
	c.mu.Lock()
	fns := append([]labelled(nil), c.funcs...)
	c.funcs = nil
	c.mu.Unlock()

	var errs []error
	for i := len(fns) - 1; i >= 0; i-- {
		l := fns[i]
		callCtx, cancel := context.WithTimeout(ctx, timeout)
		err := l.fn(callCtx)
		cancel()
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", l.label, err))
		}
	}
	return errors.Join(errs...)
}
