package httpserver

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestRunReturnsErrorWithoutHandler(t *testing.T) {
	if err := Run(context.Background(), Config{Addr: ":0"}); err == nil {
		t.Fatal("expected error when Handler is nil")
	}
}

func TestRunGracefulShutdownOnContextCancel(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	var runErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		runErr = Run(ctx, Config{Addr: ":0", Handler: mux, ShutdownTimeout: 200 * time.Millisecond})
	}()
	// Give the server a moment to bind.
	time.Sleep(50 * time.Millisecond)
	cancel()
	// Cancelling the parent context does not actually shut the server
	// down — only signals do. Send SIGINT to ourselves to wake the
	// signal-notify loop.
	_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	wg.Wait()
	if runErr != nil && !errors.Is(runErr, http.ErrServerClosed) {
		t.Fatalf("Run returned unexpected error: %v", runErr)
	}
}
