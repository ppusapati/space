// Command scheduler is the chetana distributed scheduler.
//
// REQ-FUNC-CMN-006 + design.md §3.1.
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
	"github.com/redis/go-redis/v9"

	"p9e.in/chetana/packages/observability/serverobs"
	"p9e.in/chetana/packages/region"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("scheduler failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := loadConfig()
	logger.Info("scheduler starting",
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

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer func() { _ = rdb.Close() }()

	srv := serverobs.NewServer(
		serverobs.ServerConfig{Addr: cfg.HTTPAddr},
		serverobs.ObservabilityConfig{
			Build:       serverobs.BuildInfo{Version: cfg.Version, GitSHA: cfg.GitSHA},
			MetricsAddr: cfg.MetricsAddr,
			DepChecks: []serverobs.DepCheck{
				serverobs.PostgresCheck{Pool: pool},
				serverobs.RedisCheck{Client: rdb},
			},
		},
	)
	if err := pool.Ping(context.Background()); err != nil {
		logger.Warn("postgres ping failed at boot", slog.Any("err", err))
	}
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Warn("redis ping failed at boot", slog.Any("err", err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := srv.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("scheduler stopped cleanly")
	return nil
}

type schedulerConfig struct {
	HTTPAddr    string
	MetricsAddr string
	DatabaseDSN string
	RedisAddr   string
	Version     string
	GitSHA      string
}

func loadConfig() schedulerConfig {
	return schedulerConfig{
		HTTPAddr:    getenvOr("HTTP_ADDR", ":8085"),
		MetricsAddr: getenvOr("METRICS_ADDR", ":9095"),
		DatabaseDSN: getenvOr("DATABASE_URL", region.PostgresDSN("scheduler")),
		RedisAddr:   getenvOr("CHETANA_REDIS_ADDR", region.Active().String()+".redis.chetana.internal:6379"),
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
