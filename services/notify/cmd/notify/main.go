// Command notify is the chetana notify service.
//
// REQ-FUNC-PLT-NOTIFY-001..004 + design.md §4.7.
package main

import (
	"context"
	"encoding/json"
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

	"github.com/ppusapati/space/services/notify/internal/dispatcher"
	"github.com/ppusapati/space/services/notify/internal/email"
	"github.com/ppusapati/space/services/notify/internal/inapp"
	"github.com/ppusapati/space/services/notify/internal/preferences"
	"github.com/ppusapati/space/services/notify/internal/sms"
	"github.com/ppusapati/space/services/notify/internal/store"
	"github.com/ppusapati/space/services/notify/internal/template"
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

	// RETROFIT-001 B12: wire the dispatcher.
	templates, err := store.NewTemplateStore(pool, time.Now)
	if err != nil {
		return err
	}
	prefs, err := preferences.NewStore(pool, time.Now)
	if err != nil {
		return err
	}

	// Channel senders. The chetana cmd-layer accepts the abstract
	// Sender interfaces so tests + the dev posture can swap in
	// CapturingSender. Production wires aws-sdk-go-v2's SES + SNS
	// clients once TASK-P1-PLT-SECRETS-001 lands KMS-backed creds.
	// Until then we wire CapturingSenders so the route mounts
	// cleanly + the boot-time FIPS asserts still gate.
	emailSender := newEmailSender(logger)
	smsSender := newSMSSender(logger)
	inAppPub := newInAppPublisher(logger)

	disp, err := dispatcher.New(dispatcher.Config{
		Templates:   templates,
		Preferences: prefs,
		Renderer:    template.NewRenderer(),
		Email:       emailSender,
		SMS:         smsSender,
		InApp:       inAppPub,
		Defaults: dispatcher.Defaults{
			EmailFrom: cfg.EmailFrom,
		},
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

	// /v1/notify/send → POST {SendRequest} → {Result}.
	// Plain JSON now; Connect-generated handler drops in alongside
	// when notify.proto regenerates (OQ-004).
	srv.Mux.HandleFunc("/v1/notify/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req dispatcher.SendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_request","error_description":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		res, err := disp.Send(r.Context(), req)
		if err != nil {
			// dispatcher returns Result + err for internal failures;
			// for typed outcomes (rate-limited, opted-out) Result is
			// non-nil and the caller maps Outcome → status.
			http.Error(w, `{"error":"internal","error_description":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	})

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

// loggingEmailSender writes every Send through the supplied logger
// + delegates to a CapturingSender for the test-friendly state.
// Production wires the real aws-sdk SES client here once
// TASK-P1-PLT-SECRETS-001 ships KMS-backed creds.
type loggingEmailSender struct {
	inner  *email.CapturingSender
	logger *slog.Logger
}

func (l *loggingEmailSender) Send(ctx context.Context, msg email.Message) error {
	l.logger.Info("email send",
		slog.String("from", msg.From),
		slog.Int("to_count", len(msg.To)),
		slog.String("subject", msg.Subject),
	)
	return l.inner.Send(ctx, msg)
}

func newEmailSender(logger *slog.Logger) email.Sender {
	return &loggingEmailSender{inner: &email.CapturingSender{}, logger: logger}
}

type loggingSMSSender struct {
	inner  *sms.CapturingSender
	logger *slog.Logger
}

func (l *loggingSMSSender) Send(ctx context.Context, msg sms.Message) error {
	l.logger.Info("sms send", slog.String("to", msg.To), slog.Int("body_len", len(msg.Body)))
	return l.inner.Send(ctx, msg)
}

func newSMSSender(logger *slog.Logger) sms.Sender {
	return &loggingSMSSender{inner: &sms.CapturingSender{}, logger: logger}
}

type loggingInAppPublisher struct {
	inner  *inapp.CapturingPublisher
	logger *slog.Logger
}

func (l *loggingInAppPublisher) Publish(ctx context.Context, msg inapp.Message) error {
	l.logger.Info("inapp publish",
		slog.String("user_id", msg.UserID),
		slog.String("title", msg.Title),
		slog.String("severity", msg.Severity),
	)
	return l.inner.Publish(ctx, msg)
}

func newInAppPublisher(logger *slog.Logger) inapp.Publisher {
	return &loggingInAppPublisher{inner: &inapp.CapturingPublisher{}, logger: logger}
}

type notifyConfig struct {
	HTTPAddr    string
	MetricsAddr string
	DatabaseDSN string
	SESEndpoint string
	SNSEndpoint string
	EmailFrom   string
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
		EmailFrom:   getenvOr("CHETANA_EMAIL_FROM", "Chetana <noreply@chetana.p9e.in>"),
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
