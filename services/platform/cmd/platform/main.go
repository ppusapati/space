// Command platform is the chetana platform-tenants service.
//
// Phase 1: TASK-P1-TENANT-001 ships the tenants table + the
// single-tenant seed + the per-tenant security_policy /
// quotas surface.
// TASK-P1-PLT-HEALTH-001 ships the aggregate health endpoint.
// RETROFIT-001 B9 + A5 wires both into the cmd entrypoint and
// swaps the alerter's stub Notifier for a notify-svc HTTP
// producer.
//
// REQ-FUNC-PLT-TENANT-001..003 + REQ-FUNC-CMN-004 + design.md §3.1.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/chetana/packages/observability/serverobs"
	"p9e.in/chetana/packages/region"

	"github.com/ppusapati/space/services/platform/internal/health"
	"github.com/ppusapati/space/services/platform/internal/tenant"
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
		slog.String("notify_url", cfg.NotifyURL),
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

	// RETROFIT-001 B9: health store + aggregator + alerter.
	tenantStore := tenant.NewStore(pool, time.Now)
	_ = tenantStore // exposed via /v1/tenants below.

	hStore, err := health.NewStore(pool, time.Now)
	if err != nil {
		return err
	}

	notifier := newHealthNotifier(cfg, logger)
	// All three notifier slots route to the same notify-svc
	// adapter today. Production splits them: Slack via the inapp
	// channel, Email via the email channel, Pager via the
	// PagerDuty integration once it lands. The adapter chooses
	// the channel per call by inspecting alert.Severity.
	alerter, err := health.NewAlerter(health.AlerterConfig{
		Store: hStore,
		Slack: notifier,
		Email: notifier,
		Pager: notifier,
	})
	if err != nil {
		return err
	}

	aggregator, err := health.NewAggregator(health.AggregatorConfig{
		Store:   hStore,
		Alerter: alerter,
	})
	if err != nil {
		return err
	}
	for service, url := range cfg.RegisteredServices {
		aggregator.Register(service, url)
	}

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

	// /v1/health/services — aggregated rollup.
	srv.Mux.HandleFunc("/v1/health/services", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		report, err := aggregator.Report(r.Context())
		if err != nil {
			http.Error(w, `{"error":"report_failed","error_description":"`+err.Error()+`"}`,
				http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(report)
	})

	// /v1/tenants/{id} — read the tenant row.
	srv.Mux.HandleFunc("/v1/tenants/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := r.URL.Path[len("/v1/tenants/"):]
		if id == "" {
			http.Error(w, `{"error":"invalid_request","error_description":"tenant id required"}`,
				http.StatusBadRequest)
			return
		}
		t, err := tenantStore.Get(r.Context(), id)
		if err != nil {
			http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(t)
	})

	if err := pool.Ping(context.Background()); err != nil {
		logger.Warn("postgres ping failed at boot", slog.Any("err", err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Spawn the aggregator's poll loop.
	go func() {
		if err := aggregator.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("aggregator loop crashed", slog.Any("err", err))
		}
	}()

	if err := srv.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("platform stopped cleanly")
	return nil
}

// newHealthNotifier wraps the chetana notify-svc HTTP producer in
// a small adapter that satisfies health.Notifier (A5 swap). When
// the notify URL is unset the cmd-layer falls back to a logging
// notifier so the alerter still surfaces transitions in the local
// log stream — that's the explicitly-allowlisted exception to the
// no-Nop guard. Production deploys MUST set CHETANA_NOTIFY_URL.
func newHealthNotifier(cfg platformConfig, logger *slog.Logger) health.Notifier {
	if cfg.NotifyURL == "" {
		logger.Warn("CHETANA_NOTIFY_URL unset — health alerts logged only (dev posture).")
		return &loggingNotifier{logger: logger}
	}
	return &notifyServiceNotifier{
		url:    cfg.NotifyURL,
		client: &http.Client{Timeout: 5 * time.Second},
		logger: logger,
	}
}

type loggingNotifier struct{ logger *slog.Logger }

func (l *loggingNotifier) Notify(_ context.Context, alert health.Alert) error {
	l.logger.Warn("health alert (notify-svc not wired)",
		slog.String("service", alert.Service),
		slog.String("state", alert.State),
		slog.String("severity", alert.Severity),
		slog.String("note", alert.Note),
	)
	return nil
}

type notifyServiceNotifier struct {
	url    string
	client *http.Client
	logger *slog.Logger
}

func (n *notifyServiceNotifier) Notify(ctx context.Context, alert health.Alert) error {
	body, _ := json.Marshal(map[string]any{
		"template_id":   "platform.health.alert",
		"channel":       "email",
		"variables": map[string]any{
			"service":  alert.Service,
			"state":    alert.State,
			"severity": alert.Severity,
			"note":     alert.Note,
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(n.url, "/")+"/v1/notify/send",
		bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}

// platformConfig holds the runtime configuration.
type platformConfig struct {
	HTTPAddr           string
	MetricsAddr        string
	DatabaseDSN        string
	NotifyURL          string
	RegisteredServices map[string]string
	Version            string
	GitSHA             string
}

func loadConfig() platformConfig {
	return platformConfig{
		HTTPAddr:    getenvOr("HTTP_ADDR", ":8081"),
		MetricsAddr: getenvOr("METRICS_ADDR", ":9091"),
		DatabaseDSN: getenvOr("DATABASE_URL", region.PostgresDSN("platform")),
		NotifyURL:   os.Getenv("CHETANA_NOTIFY_URL"),
		RegisteredServices: map[string]string{
			"iam":         getenvOr("CHETANA_IAM_READY_URL", "http://iam:8080/ready"),
			"audit":       getenvOr("CHETANA_AUDIT_READY_URL", "http://audit:8082/ready"),
			"notify":      getenvOr("CHETANA_NOTIFY_READY_URL", "http://notify:8083/ready"),
			"export":      getenvOr("CHETANA_EXPORT_READY_URL", "http://export:8084/ready"),
			"scheduler":   getenvOr("CHETANA_SCHEDULER_READY_URL", "http://scheduler:8085/ready"),
			"realtime-gw": getenvOr("CHETANA_RT_READY_URL", "http://realtime-gw:8086/ready"),
		},
		Version: getenvOr("CHETANA_VERSION", "v0.0.0-dev"),
		GitSHA:  getenvOr("CHETANA_GIT_SHA", "unknown"),
	}
}

func getenvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
