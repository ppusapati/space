// Package serverobs provides the per-process observability surface
// (/health, /ready, /metrics) plus a small HTTP server bundle that
// exposes them. It complements packages/connect/server, which owns the
// ConnectRPC interceptor wiring; serverobs is intentionally
// dependency-light so it can be tested in isolation without pulling in
// the heavy proto/Postgres/Kafka chain that the interceptor stack
// requires.
//
// → REQ-FUNC-CMN-001 (/health), REQ-FUNC-CMN-002 (/ready),
//   REQ-FUNC-CMN-003 (/metrics), REQ-NFR-OBS-001 (OTel),
//   REQ-NFR-SEC-001 (FIPS).
//
// → design.md §4.1.3 (FIPS), §4.7.
//
// Typical wiring from a service entrypoint:
//
//	srv := serverobs.NewServer(serverobs.ServerConfig{Addr: ":8080"},
//	    serverobs.ObservabilityConfig{
//	        Build: serverobs.BuildInfo{Version: version, GitSHA: commit},
//	        DepChecks: []serverobs.DepCheck{
//	            serverobs.PostgresCheck{Pool: pool},
//	            serverobs.RedisCheck{Client: redisClient},
//	        },
//	    },
//	)
//	connect.RegisterHandlers(srv.Mux, ...)   // wire RPC handlers
//	if err := crypto.AssertFIPS(slog.Default()); err != nil { return err }
//	return srv.Run(ctx)
package serverobs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ServerConfig holds the listen-address + timeouts for the main HTTP
// server. The metrics server uses MetricsAddr from ObservabilityConfig.
type ServerConfig struct {
	// Addr is the main listen address (e.g. ":8080"). Required.
	Addr string

	// ReadHeaderTimeout, ReadTimeout, WriteTimeout, IdleTimeout follow
	// http.Server semantics; defaults are applied when zero.
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

// withDefaults fills zero-value timeouts with conservative values. The
// totals are well below typical k8s probe deadlines.
func (c ServerConfig) withDefaults() ServerConfig {
	if c.ReadHeaderTimeout == 0 {
		c.ReadHeaderTimeout = 30 * time.Second
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = 60 * time.Second
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 60 * time.Second
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 120 * time.Second
	}
	return c
}

// ObservabilityConfig configures the per-server observability surface
// (health, readiness, metrics, FIPS check).
//
// → REQ-FUNC-CMN-001, REQ-FUNC-CMN-002, REQ-FUNC-CMN-003,
//   REQ-NFR-OBS-001, REQ-NFR-SEC-001
type ObservabilityConfig struct {
	// Build identifies the binary in /health and chetana_build_info.
	// REQUIRED — services must set Version and GitSHA at build time
	// (typically via -ldflags "-X main.version=...").
	Build BuildInfo

	// MetricsAddr is the listen address for the dedicated metrics
	// server (default ":9090"). The main RPC mux does NOT serve
	// /metrics — operators scrape the dedicated port.
	MetricsAddr string

	// DepChecks lists the readiness probes aggregated by /ready.
	// Empty slice = /ready always reports OK.
	DepChecks []DepCheck

	// ReadyCacheTTL overrides the default 5s readiness cache window.
	ReadyCacheTTL time.Duration
}

// DefaultObservabilityConfig returns sensible defaults. Callers MUST
// populate Build before passing the result to NewServer.
func DefaultObservabilityConfig() ObservabilityConfig {
	return ObservabilityConfig{
		MetricsAddr:   ":9090",
		ReadyCacheTTL: readyCacheTTL,
	}
}

// Server bundles the main RPC server and the metrics server. Returned
// by NewServer; callers wire their ConnectRPC handlers via Server.Mux
// and start with Server.Run.
type Server struct {
	// Mux is the HTTP mux for the main port. Register ConnectRPC
	// handlers here. /health and /ready are pre-registered.
	Mux *http.ServeMux

	// HTTP is the main HTTP server. Caller may wrap with h2c / CORS
	// before calling Run; the wrapping points are documented in
	// design.md §4.7.
	HTTP *http.Server

	cfg             ObservabilityConfig
	metricsServer   *http.Server
	metricsRegistry *metricsRegistry
	ready           *readyChecker
	startedAt       time.Time
}

// NewServer constructs a Server with /health, /ready, and /metrics
// pre-wired. The main RPC handlers are added by the caller via the
// returned Mux. The metrics port is started by Server.Run.
//
// Acceptance criterion #1: a service constructed via NewServer exposes
// /health, /ready, and /metrics on the documented ports without further
// configuration.
func NewServer(srvCfg ServerConfig, obsCfg ObservabilityConfig) *Server {
	srvCfg = srvCfg.withDefaults()
	if obsCfg.MetricsAddr == "" {
		obsCfg.MetricsAddr = ":9090"
	}
	if obsCfg.ReadyCacheTTL <= 0 {
		obsCfg.ReadyCacheTTL = readyCacheTTL
	}

	rc := newReadyChecker(obsCfg.DepChecks, obsCfg.ReadyCacheTTL)
	mr := newMetricsRegistry(obsCfg.Build)
	startedAt := time.Now()

	mainMux := http.NewServeMux()
	wrap := httpMetricsMiddleware(mr)
	mainMux.Handle("/health", wrap(healthHandler(obsCfg.Build, startedAt)))
	mainMux.Handle("/ready", wrap(readyHandler(rc)))

	httpServer := &http.Server{
		Addr:              srvCfg.Addr,
		Handler:           mainMux,
		ReadHeaderTimeout: srvCfg.ReadHeaderTimeout,
		ReadTimeout:       srvCfg.ReadTimeout,
		WriteTimeout:      srvCfg.WriteTimeout,
		IdleTimeout:       srvCfg.IdleTimeout,
	}

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", wrap(metricsHandler(mr, rc)))
	metricsServer := &http.Server{
		Addr:              obsCfg.MetricsAddr,
		Handler:           metricsMux,
		ReadHeaderTimeout: srvCfg.ReadHeaderTimeout,
		ReadTimeout:       srvCfg.ReadTimeout,
		WriteTimeout:      srvCfg.WriteTimeout,
		IdleTimeout:       srvCfg.IdleTimeout,
	}

	return &Server{
		Mux:             mainMux,
		HTTP:            httpServer,
		cfg:             obsCfg,
		metricsServer:   metricsServer,
		metricsRegistry: mr,
		ready:           rc,
		startedAt:       startedAt,
	}
}

// Run starts the metrics server and the main HTTP server. It returns
// when ctx is cancelled or either server fails. Both servers are then
// gracefully shut down within a 30-second budget.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() {
		if err := s.metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("metrics server: %w", err)
		}
	}()

	go func() {
		if err := s.HTTP.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("main server: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		_ = s.shutdown()
		return err
	}
	return s.shutdown()
}

// shutdown gracefully closes both servers within a 30-second budget.
func (s *Server) shutdown() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	mErr := s.metricsServer.Shutdown(shutdownCtx)
	hErr := s.HTTP.Shutdown(shutdownCtx)
	if hErr != nil {
		return hErr
	}
	return mErr
}
