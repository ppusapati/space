// Command realtime-gw is the chetana realtime WebSocket gateway.
//
// REQ-FUNC-RT-001..006 + design.md §4.3.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"

	"p9e.in/chetana/packages/observability/serverobs"
	"p9e.in/chetana/packages/region"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("realtime-gw failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := loadConfig()
	logger.Info("realtime-gw starting",
		slog.String("http_addr", cfg.HTTPAddr),
		slog.String("metrics_addr", cfg.MetricsAddr),
		slog.String("redis_addr", cfg.RedisAddr),
		slog.String("version", cfg.Version),
		slog.String("git_sha", cfg.GitSHA),
	)

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer func() { _ = rdb.Close() }()

	srv := serverobs.NewServer(
		serverobs.ServerConfig{Addr: cfg.HTTPAddr},
		serverobs.ObservabilityConfig{
			Build:       serverobs.BuildInfo{Version: cfg.Version, GitSHA: cfg.GitSHA},
			MetricsAddr: cfg.MetricsAddr,
			DepChecks: []serverobs.DepCheck{
				serverobs.RedisCheck{Client: rdb},
			},
		},
	)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Warn("redis ping failed at boot", slog.Any("err", err))
	}

	// /v1/rt WS upgrade handler registers post-OQ-004 once the
	// authzv1.Verifier + topic.Authorizer + ws.Server can be
	// constructed against a real IAM JWKS endpoint.
	_ = srv.Mux

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := srv.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("realtime-gw stopped cleanly")
	return nil
}

type rtConfig struct {
	HTTPAddr    string
	MetricsAddr string
	RedisAddr   string
	Version     string
	GitSHA      string
}

func loadConfig() rtConfig {
	return rtConfig{
		HTTPAddr:    getenvOr("HTTP_ADDR", ":8086"),
		MetricsAddr: getenvOr("METRICS_ADDR", ":9096"),
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
