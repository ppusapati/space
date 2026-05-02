// Reference service demonstrating the serverobs wiring required by
// every chetana Go service.
//
// TASK-P0-OBS-001 — REQ-FUNC-CMN-001/002/003, REQ-NFR-OBS-001,
// REQ-NFR-SEC-001.
//
// Build & run:
//
//	go run ./observability/serverobs/example
//	curl http://localhost:8080/health     # {"status":"ok",...}
//	curl http://localhost:8080/ready      # {"ok":true,"deps":[...]}
//	curl http://localhost:9090/metrics    # Prometheus text format
//
// Set CHETANA_REQUIRE_FIPS=1 on a non-boringcrypto build to see the
// FIPS self-check refuse to start.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"p9e.in/chetana/packages/crypto"
	"p9e.in/chetana/packages/observability/serverobs"
)

// build identifiers — overridden at link time via:
//
//	go build -ldflags "-X main.version=v1.2.3 -X main.gitSHA=abc123"
var (
	version = "v0.0.0-example"
	gitSHA  = "0000000"
)

// stubDep returns a DepCheck that flips between healthy and degraded
// based on the file flag /tmp/chetana-example-degrade. Useful for
// poking at /ready by `touch`-ing the flag during a curl loop.
type stubDep struct{ name, flagPath string }

func (s stubDep) Name() string { return s.name }
func (s stubDep) Check(ctx context.Context) error {
	if _, err := os.Stat(s.flagPath); err == nil {
		return errors.New("simulated upstream failure (flag file present)")
	}
	return nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// FIPS self-check. In production builds with GOEXPERIMENT=boringcrypto
	// and CHETANA_REQUIRE_FIPS=1 this enforces the contract; otherwise it
	// just logs the active provider for operator visibility.
	if err := crypto.AssertFIPS(logger); err != nil {
		logger.Error("FIPS self-check failed; refusing to start", slog.Any("err", err))
		os.Exit(1)
	}

	srv := serverobs.NewServer(
		serverobs.ServerConfig{Addr: ":8080"},
		serverobs.ObservabilityConfig{
			Build: serverobs.BuildInfo{Version: version, GitSHA: gitSHA},
			DepChecks: []serverobs.DepCheck{
				stubDep{name: "stub-upstream", flagPath: "/tmp/chetana-example-degrade"},
			},
			MetricsAddr:   ":9090",
			ReadyCacheTTL: 5 * time.Second,
		},
	)

	// Register a placeholder /hello so operators can confirm the main
	// mux is alive alongside /health and /ready.
	srv.Mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello from chetana example service\n"))
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("example service starting",
		slog.String("main_addr", ":8080"),
		slog.String("metrics_addr", ":9090"),
		slog.String("version", version),
		slog.String("git_sha", gitSHA),
	)

	if err := srv.Run(ctx); err != nil {
		logger.Error("server exited with error", slog.Any("err", err))
		os.Exit(1)
	}
	logger.Info("server stopped cleanly")
}
