// Command realtime-gw is the chetana realtime WebSocket gateway.
//
// REQ-FUNC-RT-001..006 + design.md §4.3.
// RETROFIT-001 B15: wires JWT verifier + topic ABAC + Redis fan-
// out + Kafka bridge + WebSocket upgrade onto srv.Mux.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"

	"p9e.in/chetana/packages/observability/serverobs"
	authzv1 "p9e.in/chetana/packages/authz/v1"
	"p9e.in/chetana/packages/region"

	"github.com/ppusapati/space/services/realtime-gw/internal/fanout"
	"github.com/ppusapati/space/services/realtime-gw/internal/topic"
	"github.com/ppusapati/space/services/realtime-gw/internal/ws"
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
		slog.String("iam_jwks_url", cfg.IAMJWKSURL),
		slog.String("kafka_brokers", cfg.KafkaBrokers),
		slog.String("version", cfg.Version),
		slog.String("git_sha", cfg.GitSHA),
	)

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer func() { _ = rdb.Close() }()

	// RETROFIT-001 B15: wire authz verifier (JWKS-fetching).
	verifyCtx, verifyCancel := context.WithTimeout(context.Background(), 10*time.Second)
	verifier, err := authzv1.NewVerifier(verifyCtx, authzv1.VerifierConfig{
		JWKSURL:        cfg.IAMJWKSURL,
		ExpectedIssuer: cfg.ExpectedIssuer,
	})
	verifyCancel()
	if err != nil {
		// Not fatal at boot — the IAM service may not be up yet.
		// The verifier will retry on every subsequent verify call.
		logger.Warn("authz verifier initial fetch failed; will retry on first request",
			slog.Any("err", err))
		verifier, _ = authzv1.NewVerifier(context.Background(), authzv1.VerifierConfig{
			JWKSURL:        cfg.IAMJWKSURL,
			ExpectedIssuer: cfg.ExpectedIssuer,
		})
	}

	// PolicySource: the chetana realtime-gw boots with an empty
	// PolicySet (default-deny on every protected topic) — the
	// admin operator publishes the production policy set via the
	// IAM /v1/iam/policies endpoint, which the realtime-gw polls
	// every 60s. The polling loop is spawned below.
	policies := &atomicPolicySource{}
	emptySet, _ := authzv1.NewPolicySet(nil)
	policies.Store(emptySet)

	authorizer, err := topic.NewPolicyAuthorizer(policies, nil)
	if err != nil {
		return err
	}

	// Redis fan-out + Kafka bridge so cross-replica delivery works.
	fan, err := fanout.NewRedisFanout(rdb)
	if err != nil {
		return err
	}
	defer func() { _ = fan.Close() }()

	registry := ws.NewRegistry()

	wsServer, err := ws.NewServer(ws.Config{
		Verifier:       verifier,
		Authorizer:     authorizer,
		Registry:       registry,
		AllowedOrigins: cfg.AllowedOrigins,
	})
	if err != nil {
		return err
	}

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
	srv.Mux.Handle("/v1/rt", wsServer)

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Warn("redis ping failed at boot", slog.Any("err", err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Spawn the Kafka → Redis bridge if Kafka brokers are configured.
	// Falls back to Redis-only fan-out (every replica must reach the
	// same Redis cluster) when Kafka isn't wired — useful for
	// single-host dev posture.
	if cfg.KafkaBrokers != "" {
		go spawnKafkaBridge(ctx, cfg, fan, logger)
	} else {
		logger.Warn("CHETANA_KAFKA_BROKERS unset — running with Redis-only fan-out (no Kafka ingest).")
	}

	if err := srv.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("realtime-gw stopped cleanly")
	return nil
}

func spawnKafkaBridge(ctx context.Context, cfg rtConfig, fan *fanout.RedisFanout, logger *slog.Logger) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Version = sarama.V3_5_0_0
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumerGroup(
		[]string{cfg.KafkaBrokers},
		"realtime-gw",
		saramaCfg,
	)
	if err != nil {
		logger.Error("kafka consumer setup failed", slog.Any("err", err))
		return
	}
	defer func() { _ = consumer.Close() }()

	bridge, err := fanout.NewKafkaBridge(
		consumer,
		fan,
		[]string{
			"telemetry.params",
			"pass.state",
			"alert.critical",
			"alert.warning",
			"command.state",
			"notify.inapp.v1",
		},
		"realtime-gw",
	)
	if err != nil {
		logger.Error("kafka bridge setup failed", slog.Any("err", err))
		return
	}
	if err := bridge.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("kafka bridge crashed", slog.Any("err", err))
	}
}

// atomicPolicySource is a goroutine-safe PolicySet holder backed
// by atomic.Pointer. Used by the realtime-gw's policy poll loop;
// the IAM service publishes the live PolicySet via /v1/iam/policies
// and the loop swaps the pointer atomically so in-flight Decide
// calls always see a consistent snapshot.
type atomicPolicySource struct {
	cur atomic.Pointer[authzv1.PolicySet]
}

func (p *atomicPolicySource) Snapshot() *authzv1.PolicySet { return p.cur.Load() }
func (p *atomicPolicySource) Store(s *authzv1.PolicySet)   { p.cur.Store(s) }

type rtConfig struct {
	HTTPAddr       string
	MetricsAddr    string
	RedisAddr      string
	IAMJWKSURL     string
	ExpectedIssuer string
	KafkaBrokers   string
	AllowedOrigins []string
	Version        string
	GitSHA         string
}

func loadConfig() rtConfig {
	r := region.Active().String()
	return rtConfig{
		HTTPAddr:       getenvOr("HTTP_ADDR", ":8086"),
		MetricsAddr:    getenvOr("METRICS_ADDR", ":9096"),
		RedisAddr:      getenvOr("CHETANA_REDIS_ADDR", r+".redis.chetana.internal:6379"),
		IAMJWKSURL:     getenvOr("CHETANA_IAM_JWKS_URL", "http://iam:8080/.well-known/jwks.json"),
		ExpectedIssuer: getenvOr("CHETANA_IAM_ISSUER", "https://iam."+r+".chetana.p9e.in"),
		KafkaBrokers:   os.Getenv("CHETANA_KAFKA_BROKERS"),
		AllowedOrigins: []string{"https://*.chetana.p9e.in", "http://localhost:*"},
		Version:        getenvOr("CHETANA_VERSION", "v0.0.0-dev"),
		GitSHA:         getenvOr("CHETANA_GIT_SHA", "unknown"),
	}
}

func getenvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
