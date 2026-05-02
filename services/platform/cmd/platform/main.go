// Command platform is the chetana platform-tenants service.
//
// Phase 1: TASK-P1-TENANT-001 ships the tenants table + the
// single-tenant seed + the per-tenant security_policy /
// quotas surface.
//
// REQ-FUNC-PLT-TENANT-001..003 + design.md §3.1.
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

	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/chetana/packages/observability/serverobs"
	"p9e.in/chetana/packages/region"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("platform failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := loadConfig()
	logger.Info("platform starting",
		slog.String("http_addr", cfg.HTTPAddr),
		slog.String("metrics_addr", cfg.MetricsAddr),
		slog.String("version", cfg.Version),
		slog.String("git_sha", cfg.GitSHA),
	)

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 15*time.Second)
	pool, err := pgxpool.New(dbCtx, cfg.DatabaseDSN)
	dbCancel()
	if err != nil {
		return err
	}
	defer pool.Close()

	srv := serverobs.NewServer(
		serverobs.ServerConfig{Addr: cfg.HTTPAddr},
		serverobs.ObservabilityConfig{
			Build:       serverobs.BuildInfo{Version: cfg.Version, GitSHA: cfg.GitSHA},
			MetricsAddr: cfg.MetricsAddr,
			DepChecks: []serverobs.DepCheck{
				serverobs.PostgresCheck{Pool: pool},
			},
		},
	)

	// Connect RPC handler registration lands once platform.proto
	// regenerates (BSR auth required — OQ-004). The /v1/platform
	// routes will be registered against srv.Mux there.

	if err := pool.Ping(context.Background()); err != nil {
		logger.Warn("postgres ping failed at boot", slog.Any("err", err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := srv.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("platform stopped cleanly")
	return nil
}

// platformConfig holds the runtime configuration. Inlined here
// (rather than a dedicated internal/config package) because the
// platform service has very few knobs in v1.
type platformConfig struct {
	HTTPAddr    string
	MetricsAddr string
	DatabaseDSN string
	Version     string
	GitSHA      string
}

func loadConfig() platformConfig {
	return platformConfig{
		HTTPAddr:    getenvOr("HTTP_ADDR", ":8081"),
		MetricsAddr: getenvOr("METRICS_ADDR", ":9091"),
		DatabaseDSN: getenvOr("DATABASE_URL", region.PostgresDSN("platform")),
		Version:     getenvOr("CHETANA_VERSION", "v0.0.0-dev"),
		GitSHA:      getenvOr("CHETANA_GIT_SHA", "unknown"),
	}
}

func getenvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
