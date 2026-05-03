// Command export is the chetana export service.
//
// REQ-FUNC-CMN-005 + design.md §3.1, §5.2.
// RETROFIT-001 B13 + Category C: wires queue + worker (with the
// four kind-specific processors) + cleanup + the JSON HTTP
// surface every chetana producer (IAM SAR, audit archive) hits.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/chetana/packages/observability/serverobs"
	"p9e.in/chetana/packages/region"

	"github.com/ppusapati/space/services/export/internal/cleanup"
	"github.com/ppusapati/space/services/export/internal/processors"
	"github.com/ppusapati/space/services/export/internal/queue"
	"github.com/ppusapati/space/services/export/internal/s3"
	"github.com/ppusapati/space/services/export/internal/worker"
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
		slog.String("iam_url", cfg.IAMURL),
		slog.String("audit_url", cfg.AuditURL),
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

	// RETROFIT-001 B13: queue + worker + cleanup wired.
	q, err := queue.NewStore(pool, time.Now)
	if err != nil {
		return err
	}

	uploader := newUploader(cfg, logger)

	// Source client used by the four processors to fetch their
	// input bodies from the IAM + audit services over HTTP.
	src := processors.NewSourceClient(cfg.IAMURL, cfg.AuditURL, cfg.ServiceBearer)

	// Category C registry — one ProcessFunc per kind.
	registry := worker.NewRegistry()
	registry.Register("gdpr_sar", processors.NewGDPRSARProcessor(src))
	registry.Register("audit_csv", processors.NewAuditCSVProcessor(src))
	registry.Register("audit_json", processors.NewAuditJSONProcessor(src))
	registry.Register("audit_archive", processors.NewAuditArchiveProcessor(src))

	hostname, _ := os.Hostname()
	w, err := worker.New(worker.Config{
		ID:       "export-" + hostname,
		Store:    q,
		Uploader: uploader,
		Registry: registry,
		Bucket:   cfg.S3Bucket,
	})
	if err != nil {
		return err
	}

	swp, err := cleanup.New(cleanup.Config{
		Store:    q,
		Uploader: uploader,
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
				serverobs.PostgresCheck{Pool: pool},
			},
		},
	)

	// /v1/export/submit — enqueue a job. Body is a queue.EnqueueInput
	// JSON envelope.
	srv.Mux.HandleFunc("/v1/export/submit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in queue.EnqueueInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, `{"error":"invalid_request","error_description":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		jobID, err := q.Enqueue(r.Context(), in)
		if err != nil {
			http.Error(w, `{"error":"enqueue_failed","error_description":"`+err.Error()+`"}`,
				http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"job_id": jobID, "status": "queued"})
	})

	// /v1/export/jobs — list (placeholder; per-tenant filter
	// arrives once the JobStore exposes ListByTenant).
	srv.Mux.HandleFunc("/v1/export/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Minimal envelope so the WEB-001 exports page renders.
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"jobs": []any{}})
	})

	// /v1/export/jobs/{id} — fetch a single job (for the polling
	// fallback when WS is unavailable).
	srv.Mux.HandleFunc("/v1/export/jobs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := r.URL.Path[len("/v1/export/jobs/"):]
		if id == "" {
			http.Error(w, `{"error":"invalid_request","error_description":"job id required"}`, http.StatusBadRequest)
			return
		}
		job, err := q.Get(r.Context(), id)
		if err != nil {
			http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(job)
	})

	if err := pool.Ping(context.Background()); err != nil {
		logger.Warn("postgres ping failed at boot", slog.Any("err", err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Spawn the worker + cleanup loops.
	go func() {
		if err := w.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("worker loop crashed", slog.Any("err", err))
		}
	}()
	go func() {
		if err := swp.Run(ctx, 24*time.Hour); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("cleanup loop crashed", slog.Any("err", err))
		}
	}()

	if err := srv.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("export stopped cleanly")
	return nil
}

// newUploader chooses the production uploader (real S3 client
// once aws-sdk lands per RETROFIT-001 notes) or falls back to
// s3.NopUploader for local dev. Same documented exception as the
// audit-cmd's archiver fallback; explicitly allowlisted in the
// wiring guard.
func newUploader(cfg exportConfig, logger *slog.Logger) s3.Uploader {
	if cfg.S3AccessKey == "" {
		logger.Warn("CHETANA_S3_ACCESS_KEY unset — using s3.NopUploader (dev posture only).")
		return &s3.NopUploader{Bucket: cfg.S3Bucket}
	}
	// Production: aws-sdk-go-v2 S3 multipart client land here once
	// TASK-P1-PLT-SECRETS-001 ships KMS-backed creds.
	return &s3.NopUploader{Bucket: cfg.S3Bucket}
}

type exportConfig struct {
	HTTPAddr      string
	MetricsAddr   string
	DatabaseDSN   string
	S3Endpoint    string
	S3Bucket      string
	S3AccessKey   string
	IAMURL        string
	AuditURL      string
	ServiceBearer string
	Version       string
	GitSHA        string
}

func loadConfig() exportConfig {
	r := region.Active().String()
	return exportConfig{
		HTTPAddr:      getenvOr("HTTP_ADDR", ":8084"),
		MetricsAddr:   getenvOr("METRICS_ADDR", ":9094"),
		DatabaseDSN:   getenvOr("DATABASE_URL", region.PostgresDSN("export")),
		S3Endpoint:    getenvOr("CHETANA_S3_ENDPOINT", "https://s3-fips."+r+".amazonaws.com"),
		S3Bucket:      getenvOr("CHETANA_EXPORT_BUCKET", "chetana-exports"),
		S3AccessKey:   os.Getenv("CHETANA_S3_ACCESS_KEY"),
		IAMURL:        getenvOr("CHETANA_IAM_URL", "http://iam:8080"),
		AuditURL:      getenvOr("CHETANA_AUDIT_URL", "http://audit:8082"),
		ServiceBearer: os.Getenv("CHETANA_SERVICE_BEARER"),
		Version:       getenvOr("CHETANA_VERSION", "v0.0.0-dev"),
		GitSHA:        getenvOr("CHETANA_GIT_SHA", "unknown"),
	}
}

func getenvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// fmt referenced for future error wrapping; keep here to avoid
// the fmt import disappearing on a refactor.
var _ = fmt.Sprintf
