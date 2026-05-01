// Package httpserver runs an h2c (HTTP/2 cleartext) server suitable for
// ConnectRPC handlers behind a reverse proxy. It also supports plain
// HTTPS when a TLS config is supplied. The Run method handles signals
// and orchestrates a deadline-bounded shutdown.
package httpserver

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Config configures Run.
type Config struct {
	// Addr is the listen address (e.g. ":8080").
	Addr string
	// Handler is the root handler. Use connect.NewHandler to mount
	// services.
	Handler http.Handler
	// TLSConfig switches to HTTPS when non-nil; otherwise the server
	// runs h2c.
	TLSConfig *tls.Config
	// ShutdownTimeout caps the time given to in-flight requests on
	// SIGTERM / SIGINT.
	ShutdownTimeout time.Duration
	// ReadHeaderTimeout caps slow-loris attacks. Defaults to 15 s.
	ReadHeaderTimeout time.Duration
}

// Run blocks until the process receives SIGINT or SIGTERM. It returns
// the first error from ListenAndServe (if any) joined with the
// shutdown error.
func Run(ctx context.Context, cfg Config) error {
	if cfg.Handler == nil {
		return errors.New("httpserver: Handler is required")
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = 15 * time.Second
	}
	if cfg.ReadHeaderTimeout == 0 {
		cfg.ReadHeaderTimeout = 15 * time.Second
	}

	h2s := &http2.Server{}
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           h2c.NewHandler(cfg.Handler, h2s),
		TLSConfig:         cfg.TLSConfig,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}
	if cfg.TLSConfig != nil {
		srv.Handler = cfg.Handler
		_ = http2.ConfigureServer(srv, h2s)
	}

	signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	var (
		serveErr error
		serveWG  sync.WaitGroup
	)
	serveWG.Add(1)
	go func() {
		defer serveWG.Done()
		var err error
		if cfg.TLSConfig != nil {
			err = srv.ListenAndServeTLS("", "")
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr = err
		}
	}()

	<-signalCtx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	shutdownErr := srv.Shutdown(shutdownCtx)
	serveWG.Wait()
	return errors.Join(serveErr, shutdownErr)
}
