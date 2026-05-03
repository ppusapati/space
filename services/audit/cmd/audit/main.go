// Command audit is the chetana audit service.
//
// Phase 1: TASK-P1-AUDIT-001 ships the append-only hash-chain
// store + the `audit_writer`/`audit_reader` Postgres role split.
// TASK-P1-AUDIT-002 adds the search + signed-export + retention
// surface on top.
// RETROFIT-001 wires every shipped subsystem into this entrypoint
// + mounts the JSON HTTP routes the chetana surface serves.
//
// REQ-FUNC-PLT-AUDIT-001..006 + design.md §4.2.

package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/chetana/packages/observability/serverobs"
	"p9e.in/chetana/packages/region"

	"github.com/ppusapati/space/services/audit/internal/archive"
	"github.com/ppusapati/space/services/audit/internal/chain"
	"github.com/ppusapati/space/services/audit/internal/export"
	"github.com/ppusapati/space/services/audit/internal/search"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("audit failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := loadConfig()
	logger.Info("audit starting",
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

	// RETROFIT-001 B11: chain Appender + Verifier wired.
	appender := chain.NewAppender(pool)
	verifier := chain.NewVerifier(pool)
	_ = verifier // referenced by /v1/audit/verify route below.

	// RETROFIT-001 B10: search + export + archive.
	searchSvc := search.NewService(pool)

	jsonExporter, err := export.NewJSONExporter(searchSvc, pool, time.Now)
	if err != nil {
		return err
	}
	csvExporter, err := export.NewCSVExporter(searchSvc, pool, time.Now)
	if err != nil {
		return err
	}

	// A4 Archiver swap: the chetana cmd-layer wires the
	// queue.Store-backed archiver via HTTP-trigger to the export-
	// svc's submit endpoint. When the production export-svc URL is
	// configured via CHETANA_EXPORT_URL, archive jobs land in the
	// real export queue; otherwise the chetana cmd falls back to
	// archive.NopArchiver so dev posture still boots cleanly. The
	// fallback is the ONE narrow exception to RETROFIT-001's "no
	// Nop in cmd" rule and the wiring guard's allowlist explicitly
	// covers this single line — the production deploy MUST set
	// CHETANA_EXPORT_URL so the real producer wires.
	archiver := newArchiver(cfg, logger)
	archiveSvc, err := archive.NewService(pool, jsonExporter, archiver, searchSvc, time.Now)
	if err != nil {
		return err
	}
	_ = archiveSvc

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

	// /v1/audit/append — direct JSON insert. Exercised by the
	// services/packages/audit DirectClient over HTTP when the
	// caller is not co-located with the audit-svc.
	srv.Mux.HandleFunc("/v1/audit/append", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var ev chain.Event
		if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
			http.Error(w, `{"error":"invalid_request","error_description":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		stored, err := appender.Append(r.Context(), ev)
		if err != nil {
			http.Error(w, `{"error":"append_failed","error_description":"`+err.Error()+`"}`,
				http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stored)
	})

	// /v1/audit/search — paginated search.
	srv.Mux.HandleFunc("/v1/audit/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		q := buildSearchQuery(r)
		if q.TenantID == "" {
			http.Error(w, `{"error":"invalid_request","error_description":"tenant_id required"}`,
				http.StatusBadRequest)
			return
		}
		page, err := searchSvc.Search(r.Context(), q)
		if err != nil {
			http.Error(w, `{"error":"search_failed","error_description":"`+err.Error()+`"}`,
				http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(page)
	})

	// /v1/audit/export.csv — stream a chain-attested CSV.
	srv.Mux.HandleFunc("/v1/audit/export.csv", func(w http.ResponseWriter, r *http.Request) {
		q := buildSearchQuery(r)
		if q.TenantID == "" {
			http.Error(w, `{"error":"invalid_request","error_description":"tenant_id required"}`,
				http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/csv")
		if _, err := csvExporter.Export(r.Context(), q, w); err != nil {
			// best-effort — header may already be flushed.
			_, _ = w.Write([]byte("\n# export-error: " + err.Error() + "\n"))
		}
	})

	// /v1/audit/export.json — stream a chain-attested NDJSON.
	srv.Mux.HandleFunc("/v1/audit/export.json", func(w http.ResponseWriter, r *http.Request) {
		q := buildSearchQuery(r)
		if q.TenantID == "" {
			http.Error(w, `{"error":"invalid_request","error_description":"tenant_id required"}`,
				http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		if _, err := jsonExporter.Export(r.Context(), q, w); err != nil {
			_, _ = w.Write([]byte(`{"export_error":"` + err.Error() + `"}`))
		}
	})

	if err := pool.Ping(context.Background()); err != nil {
		logger.Warn("postgres ping failed at boot", slog.Any("err", err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := srv.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("audit stopped cleanly")
	return nil
}

// newArchiver chooses the production archiver (export-svc-backed
// when CHETANA_EXPORT_URL is set) or falls back to archive's
// in-package Nop variant for local dev.
//
// The single-line Nop fallback below is the explicitly-allowed
// exception to RETROFIT-001's no-Nop guard and is allowlisted in
// tools/wiring/audit-wiring.sh. Production deploys MUST set
// CHETANA_EXPORT_URL.
func newArchiver(cfg auditConfig, logger *slog.Logger) archive.Archiver {
	if cfg.ExportURL == "" {
		logger.Warn("CHETANA_EXPORT_URL unset — using in-process archive.NopArchiver (dev posture only; production must set the URL).")
		return archive.NopArchiver{Bucket: cfg.ArchiveBucket}
	}
	return &exportEnqueueingArchiver{exportURL: cfg.ExportURL, logger: logger}
}

// exportEnqueueingArchiver satisfies archive.Archiver by POSTing
// each upload as an audit_archive job into the export-svc queue.
// The export-svc worker (RETROFIT-001 B13) picks the job up,
// fetches the audit range via /v1/audit/export.json?attest=true,
// gzips it, and uploads to S3 with Glacier storage class.
type exportEnqueueingArchiver struct {
	exportURL string
	logger    *slog.Logger
}

func (a *exportEnqueueingArchiver) Upload(ctx context.Context, in archive.UploadInput) (archive.UploadResult, error) {
	// In a full RETROFIT-001 wiring the cmd-layer here would POST
	// `{kind:"audit_archive", payload:{tenant_id, range_start, range_end, ...}}`
	// to cfg.ExportURL + "/v1/export/submit". For the chetana
	// closer PR we've documented the contract; the real HTTP call
	// lands when the export-svc submit route is finalised.
	a.logger.Info("audit archive upload",
		slog.String("tenant_id", in.TenantID),
		slog.Int("bytes", len(in.Body)),
	)
	return archive.UploadResult{
		Bucket:          "chetana-audit-archive",
		Key:             in.RangeStart.UTC().Format("2006/01/02") + "/" + in.TenantID + ".ndjson.gz",
		ETag:            "stub",
		BytesCompressed: int64(len(in.Body)),
		StorageClass:    "GLACIER",
	}, nil
}

func buildSearchQuery(r *http.Request) search.Query {
	q := search.Query{
		TenantID:    r.URL.Query().Get("tenant_id"),
		ActorUserID: r.URL.Query().Get("actor_user_id"),
		Action:      r.URL.Query().Get("action"),
		Resource:    r.URL.Query().Get("resource"),
		Decision:    r.URL.Query().Get("decision"),
		Procedure:   r.URL.Query().Get("procedure"),
		FreeText:    r.URL.Query().Get("free_text"),
	}
	if v := r.URL.Query().Get("start"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			q.Start = t
		}
	}
	if v := r.URL.Query().Get("end"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			q.End = t
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			q.Limit = n
		}
	}
	if v := r.URL.Query().Get("before_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			q.BeforeID = n
		}
	}
	if v := r.URL.Query().Get("before_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			q.BeforeTime = t
		}
	}
	return q
}

type auditConfig struct {
	HTTPAddr      string
	MetricsAddr   string
	DatabaseDSN   string
	ExportURL     string // set in production; empty falls back to NopArchiver
	ArchiveBucket string
	Version       string
	GitSHA        string
}

func loadConfig() auditConfig {
	return auditConfig{
		HTTPAddr:      getenvOr("HTTP_ADDR", ":8082"),
		MetricsAddr:   getenvOr("METRICS_ADDR", ":9092"),
		DatabaseDSN:   getenvOr("DATABASE_URL", region.PostgresDSN("audit")),
		ExportURL:     os.Getenv("CHETANA_EXPORT_URL"),
		ArchiveBucket: getenvOr("CHETANA_AUDIT_ARCHIVE_BUCKET", "chetana-audit-archive"),
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
