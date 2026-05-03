// Command export is the chetana export service.
//
// REQ-FUNC-CMN-005 + design.md §3.1, §5.2.
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

	"github.com/ppusapati/space/services/export/internal/s3"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("export failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := loadConfig()
	logger.Info("export starting",
		slog.String("http_addr", cfg.HTTPAddr),
		slog.String("metrics_addr", cfg.MetricsAddr),
		slog.String("s3_endpoint", cfg.S3Endpoint),
		slog.String("s3_bucket", cfg.S3Bucket),
		slog.String("version", cfg.Version),
		slog.String("git_sha", cfg.GitSHA),
	)

	if err := s3.FIPSAsserts(cfg.S3Endpoint); err != nil {
		return err
	}
	logger.Info("s3 fips endpoint verified", slog.String("endpoint", cfg.S3Endpoint))

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
	if err := pool.Ping(context.Background()); err != nil {
		logger.Warn("postgres ping failed at boot", slog.Any("err", err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := srv.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("export stopped cleanly")
	return nil
}

type exportConfig struct {
	HTTPAddr    string
	MetricsAddr string
	DatabaseDSN string
	S3Endpoint  string
	S3Bucket    string
	Version     string
	GitSHA      string
}

func loadConfig() exportConfig {
	r := region.Active().String()
	return exportConfig{
		HTTPAddr:    getenvOr("HTTP_ADDR", ":8084"),
		MetricsAddr: getenvOr("METRICS_ADDR", ":9094"),
		DatabaseDSN: getenvOr("DATABASE_URL", region.PostgresDSN("export")),
		S3Endpoint:  getenvOr("CHETANA_S3_ENDPOINT", "https://s3-fips."+r+".amazonaws.com"),
		S3Bucket:    getenvOr("CHETANA_EXPORT_BUCKET", "chetana-exports"),
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
