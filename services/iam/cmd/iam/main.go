// Command iam is the chetana Identity & Access Management service.
//
// Phase 1: TASK-P1-IAM-001 ships login + Argon2id + rate-limit +
// lockout. JWT issuance (TASK-P1-IAM-002), MFA (003), WebAuthn (004),
// OIDC (005), SAML (006), session manager (007), password reset
// (008), and GDPR SAR/erasure (009) layer on top in subsequent
// tasks.
//
// REQ-FUNC-PLT-IAM-001 / REQ-FUNC-PLT-IAM-003.
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

	"p9e.in/chetana/packages/crypto"
	dbmigrate "p9e.in/chetana/packages/db/migrate"
	"p9e.in/chetana/packages/observability/serverobs"

	"github.com/ppusapati/space/services/iam/internal/config"
	"github.com/ppusapati/space/services/iam/internal/login"
	"github.com/ppusapati/space/services/iam/internal/store"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("iam failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	// 1. FIPS self-check (REQ-NFR-SEC-001). Refuses to serve if
	// CHETANA_REQUIRE_FIPS=1 and provider != boringcrypto.
	if err := crypto.AssertFIPS(logger); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger.Info("iam starting",
		slog.String("http_addr", cfg.HTTPAddr),
		slog.String("metrics_addr", cfg.MetricsAddr),
		slog.String("tenant_id", cfg.TenantID),
		slog.String("version", cfg.Version),
		slog.String("git_sha", cfg.GitSHA),
	)

	// 2. Apply platform-wide migrations (timescaledb / postgis /
	// pg_trgm extensions + retention policies if applicable). Per-
	// service schema migrations are run separately; this just
	// guarantees the cluster baseline is in place.
	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), 60*time.Second)
	if err := dbmigrate.EnsureUp(migrateCtx, dbmigrate.Config{DSN: cfg.DatabaseDSN, Logger: logger}); err != nil {
		migrateCancel()
		return err
	}
	migrateCancel()

	// 3. Open the Postgres pool.
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 15*time.Second)
	pool, err := pgxpool.New(dbCtx, cfg.DatabaseDSN)
	dbCancel()
	if err != nil {
		return err
	}
	defer pool.Close()

	// 4. Open the Redis client.
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer func() { _ = rdb.Close() }()

	// 5. Wire the login handler.
	users := store.NewStore(pool)
	limiter := login.NewIPLimiter(rdb, login.IPLimiterConfig{})
	audit := login.NopAudit{} // TASK-P1-AUDIT-001 supplies the real Kafka emitter
	handler, err := login.NewHandler(limiter, users, audit, login.HandlerConfig{
		TenantID: cfg.TenantID,
	})
	if err != nil {
		return err
	}
	_ = handler // wired into the Connect mux once iam.proto codegen lands

	// 6. Build the observability HTTP surface (/health, /ready, /metrics).
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

	// Connect RPC handler registration lands once iam.proto is
	// generated (BSR auth required — OQ-004). The /v1/iam routes
	// will be registered against srv.Mux there.

	// Optional readiness probe: dial the dependencies at boot so
	// unrecoverable failures surface immediately rather than after
	// the first /ready scrape.
	if err := pool.Ping(context.Background()); err != nil {
		logger.Warn("postgres ping failed at boot", slog.Any("err", err))
	}
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Warn("redis ping failed at boot", slog.Any("err", err))
	}

	// 7. Run until SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := srv.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("iam stopped cleanly")
	return nil
}
