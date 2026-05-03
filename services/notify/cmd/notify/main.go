// Command notify is the chetana notify service.
//
// REQ-FUNC-PLT-NOTIFY-001..004 + design.md §4.7.
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

	"github.com/ppusapati/space/services/notify/internal/email"
	"github.com/ppusapati/space/services/notify/internal/sms"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("notify failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := loadConfig()
	logger.Info("notify starting",
		slog.String("http_addr", cfg.HTTPAddr),
		slog.String("metrics_addr", cfg.MetricsAddr),
		slog.String("ses_endpoint", cfg.SESEndpoint),
		slog.String("sns_endpoint", cfg.SNSEndpoint),
		slog.String("version", cfg.Version),
		slog.String("git_sha", cfg.GitSHA),
	)

	// REQ-FUNC-PLT-NOTIFY-004: assert + log FIPS endpoints at boot.
	if err := email.FIPSAsserts(cfg.SESEndpoint); err != nil {
		return err
	}
	logger.Info("ses fips endpoint verified", slog.String("endpoint", cfg.SESEndpoint))

	if err := sms.FIPSAsserts(cfg.SNSEndpoint); err != nil {
		return err
	}
	logger.Info("sns fips endpoint verified", slog.String("endpoint", cfg.SNSEndpoint))

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
	logger.Info("notify stopped cleanly")
	return nil
}

type notifyConfig struct {
	HTTPAddr    string
	MetricsAddr string
	DatabaseDSN string
	SESEndpoint string
	SNSEndpoint string
	Version     string
	GitSHA      string
}

func loadConfig() notifyConfig {
	r := region.Active().String()
	return notifyConfig{
		HTTPAddr:    getenvOr("HTTP_ADDR", ":8083"),
		MetricsAddr: getenvOr("METRICS_ADDR", ":9093"),
		DatabaseDSN: getenvOr("DATABASE_URL", region.PostgresDSN("notify")),
		SESEndpoint: getenvOr("CHETANA_SES_ENDPOINT", "https://email-fips."+r+".amazonaws.com"),
		SNSEndpoint: getenvOr("CHETANA_SNS_ENDPOINT", "https://sns-fips."+r+".amazonaws.com"),
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
