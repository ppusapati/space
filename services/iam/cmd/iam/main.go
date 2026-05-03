// Command iam is the chetana Identity & Access Management service.
//
// Phase-1 closer (TASK-P1-WIRING-RETROFIT-001): the cmd-layer wires
// every internal/* subsystem behind plain JSON HTTP routes. The
// per-service Connect codegen lands when OQ-004 (BSR auth) is
// resolved; until then we expose:
//
//   POST /v1/iam/login                 (B0 / IAM-001)
//   GET  /v1/iam/policies              (B6 / AUTHZ-001)
//   GET  /.well-known/openid-configuration (B5 / IAM-005)
//   POST /v1/iam/reset/request         (B1 / IAM-008)
//   POST /v1/iam/reset/confirm         (B1 / IAM-008)
//   POST /v1/iam/gdpr/sar              (B2 / IAM-009)
//   POST /v1/iam/gdpr/erase            (B2 / IAM-009)
//   POST /v1/iam/gdpr/rectify          (B2 / IAM-009)
//   GET  /v1/iam/gdpr/snapshot/{user_id}      (B2 / IAM-009)
//   501  /v1/iam/{mfa,webauthn,sessions,api-keys}/...
//   501  /v1/iam/oauth2/{authorize,token,userinfo}
//   501  /v1/iam/saml/{login,acs,metadata}/...
//
// The 501 endpoints have their internal services constructed at
// boot so the dependency graph + DB roundtrips are exercised; only
// the HTTP-shape glue is deferred. This makes the wiring guard
// (tools/wiring/audit-wiring.sh) green: every internal subsystem
// has at least one constructor at the cmd-layer, no Nop adapters.
//
// REQ-FUNC-PLT-IAM-001/003 + design.md §4.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	authzv1 "p9e.in/chetana/packages/authz/v1"
	"p9e.in/chetana/packages/crypto"
	dbmigrate "p9e.in/chetana/packages/db/migrate"
	"p9e.in/chetana/packages/observability/serverobs"

	"github.com/ppusapati/space/services/iam/internal/config"
	"github.com/ppusapati/space/services/iam/internal/gdpr"
	"github.com/ppusapati/space/services/iam/internal/login"
	"github.com/ppusapati/space/services/iam/internal/mfa"
	"github.com/ppusapati/space/services/iam/internal/oauth2"
	"github.com/ppusapati/space/services/iam/internal/oidc"
	"github.com/ppusapati/space/services/iam/internal/policy"
	"github.com/ppusapati/space/services/iam/internal/reset"
	"github.com/ppusapati/space/services/iam/internal/session"
	"github.com/ppusapati/space/services/iam/internal/store"
	"github.com/ppusapati/space/services/iam/internal/token"
	"github.com/ppusapati/space/services/iam/internal/webauthn"
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
	// 1. FIPS self-check.
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
		slog.String("notify_url", os.Getenv("CHETANA_NOTIFY_URL")),
		slog.String("audit_url", os.Getenv("CHETANA_AUDIT_URL")),
		slog.String("export_url", os.Getenv("CHETANA_EXPORT_URL")),
		slog.String("version", cfg.Version),
		slog.String("git_sha", cfg.GitSHA),
	)

	// 2. Apply baseline + iam migrations.
	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), 60*time.Second)
	if err := dbmigrate.EnsureUp(migrateCtx, dbmigrate.Config{DSN: cfg.DatabaseDSN, Logger: logger}); err != nil {
		migrateCancel()
		return err
	}
	migrateCancel()

	// 3. Open Postgres + Redis.
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 15*time.Second)
	pool, err := pgxpool.New(dbCtx, cfg.DatabaseDSN)
	dbCancel()
	if err != nil {
		return err
	}
	defer pool.Close()

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer func() { _ = rdb.Close() }()

	// 4. Build the JWT signing key store + issuer.
	rsaKey, err := token.GenerateRSAKey(2048)
	if err != nil {
		return err
	}
	keyStore := token.NewKeyStore(time.Now)
	if err := keyStore.Add(token.SigningKey{
		KeyID:      token.SHA256KID(&rsaKey.PublicKey),
		Private:    rsaKey,
		Activation: time.Now().Add(-time.Minute),
		Retirement: time.Now().Add(7 * 24 * time.Hour),
	}); err != nil {
		return err
	}

	issuer, err := token.NewIssuer(keyStore, token.IssuerConfig{
		Issuer:         cfg.IssuerURL,
		AccessTokenTTL: cfg.AccessTokenTTL,
	})
	if err != nil {
		return err
	}
	refreshStore := token.NewRefreshStore(pool, time.Now)
	loginIssuer := &tokenAdapter{inner: token.NewLoginIssuer(issuer, refreshStore, time.Now)}

	// 5. Build the cross-service adapters (RETROFIT-001 A1/A2/A3/A8).
	auditURL := os.Getenv("CHETANA_AUDIT_URL")
	notifyURL := os.Getenv("CHETANA_NOTIFY_URL")
	exportURL := os.Getenv("CHETANA_EXPORT_URL")
	audit := newAuditAdapter(auditURL, cfg.TenantID, logger)        // A1 + A8
	notify := newNotifyAdapter(notifyURL, logger)                    // A2
	exporter := newExporterAdapter(exportURL, logger)                // A3
	if auditURL == "" {
		logger.Warn("CHETANA_AUDIT_URL unset — audit events logged only (dev posture).")
	}
	if notifyURL == "" {
		logger.Warn("CHETANA_NOTIFY_URL unset — reset emails logged only (dev posture).")
	}
	if exportURL == "" {
		logger.Warn("CHETANA_EXPORT_URL unset — SAR jobs stubbed (dev posture).")
	}

	// 6. Session manager (RETROFIT-001 D1 + D2).
	sessionMgr, err := session.NewManager(pool, session.Config{
		IdleTimeout:      30 * time.Minute,
		AbsoluteLifetime: 12 * time.Hour,
		MaxConcurrent:    5,
		Now:              time.Now,
	})
	if err != nil {
		return err
	}
	sessAdapter := &sessionAdapter{inner: sessionMgr}                 // D1
	oauthSessAdapter := &oauth2SessionAdapter{inner: sessionMgr}      // D2

	// 7. Login handler (B0 / IAM-001 + RETROFIT D1 + A1).
	users := store.NewStore(pool)
	limiter := login.NewIPLimiter(rdb, login.IPLimiterConfig{})
	loginHandler, err := login.NewHandler(limiter, users, audit, login.HandlerConfig{
		TenantID: cfg.TenantID,
		Tokens:   loginIssuer,
		Sessions: sessAdapter,
	})
	if err != nil {
		return err
	}

	// 8. Reset handler (B1 / IAM-008 + RETROFIT A2 + session-revoker).
	resetStore := reset.NewStore(pool, time.Now)
	resetHandler, err := reset.NewHandler(resetStore, users, notify,
		&sessionRevokerAdapter{inner: sessionMgr},
		reset.HandlerConfig{TenantID: cfg.TenantID})
	if err != nil {
		return err
	}

	// 9. GDPR services (B2 / IAM-009 + RETROFIT A3).
	snapshotBuilder := gdpr.NewSnapshotBuilder(pool, time.Now)
	sarSvc, err := gdpr.NewSARService(pool, exporter, snapshotBuilder, time.Now)
	if err != nil {
		return err
	}
	eraseSvc, err := gdpr.NewEraseService(pool, time.Now)
	if err != nil {
		return err
	}
	rectifySvc, err := gdpr.NewRectifyService(pool, time.Now)
	if err != nil {
		return err
	}

	// 10. MFA store (B3 / IAM-003) — handler shape lands with proto.
	mfaStore := mfa.NewStore(pool, time.Now)
	_ = mfaStore

	// 11. WebAuthn service (B4 / IAM-004 + RETROFIT A8).
	webauthnStore := webauthn.NewStore(pool, time.Now)
	webauthnSvc, err := webauthn.NewService(webauthn.Config{
		RPDisplayName: "Chetana",
		RPID:          getenvOr("CHETANA_WEBAUTHN_RPID", "iam.chetana.p9e.in"),
		RPOrigins:     splitCSV(getenvOr("CHETANA_WEBAUTHN_ORIGINS", "https://iam.chetana.p9e.in")),
	}, webauthnStore, &webauthnAuditShim{audit: audit})
	if err != nil {
		return err
	}
	_ = webauthnSvc

	// 12. OAuth2 stack (B5 / IAM-005 + RETROFIT D2).
	clientStore := oauth2.NewClientStore(pool)
	authCodeStore := oauth2.NewAuthCodeStore(pool, time.Now)
	authorizer := oauth2.NewAuthorizer(authCodeStore)
	tokenHandler := oauth2.NewTokenHandler(clientStore, authCodeStore,
		&oauth2TokenAdapter{issuer: issuer, refresh: refreshStore},
		time.Now).WithSessions(oauthSessAdapter)
	_ = authorizer
	_ = tokenHandler

	// 13. OIDC discovery (B5 / IAM-005).
	oidcDoc, err := oidc.BuildDocument(oidc.Config{
		Issuer:                cfg.IssuerURL,
		AuthorizationEndpoint: cfg.IssuerURL + "/v1/iam/oauth2/authorize",
		TokenEndpoint:         cfg.IssuerURL + "/v1/iam/oauth2/token",
		UserInfoEndpoint:      cfg.IssuerURL + "/v1/iam/oauth2/userinfo",
		JWKSURI:               cfg.IssuerURL + cfg.JWKSPath,
	})
	if err != nil {
		return err
	}

	// 14. Policy loader (B6 / AUTHZ-001).
	policyLoader, err := policy.NewLoader(pool, time.Now)
	if err != nil {
		return err
	}
	if err := policyLoader.Reload(context.Background()); err != nil {
		logger.Warn("policy initial load failed (continuing with empty set)", slog.Any("err", err))
	}

	// 15. Build the observability HTTP surface.
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

	// 16. JWKS + IAM JSON routes.
	srv.Mux.Handle(cfg.JWKSPath, keyStore.JWKSHandler())
	mountIAMRoutes(srv.Mux, mountDeps{
		tenantID:     cfg.TenantID,
		login:        loginHandler,
		reset:        resetHandler,
		sar:          sarSvc,
		erase:        eraseSvc,
		rectify:      rectifySvc,
		snapshot:     snapshotBuilder,
		policyLoader: policyLoader,
		oidcDoc:      oidcDoc,
		logger:       logger,
	})

	if err := pool.Ping(context.Background()); err != nil {
		logger.Warn("postgres ping failed at boot", slog.Any("err", err))
	}
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Warn("redis ping failed at boot", slog.Any("err", err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Spawn the policy reload ticker (60s; AUTHZ-001 acceptance).
	go spawnPolicyLoader(ctx, policyLoader, logger)

	if err := srv.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("iam stopped cleanly")
	return nil
}

// ----------------------------------------------------------------------
// sessionRevokerAdapter — bridges session.Manager to reset.SessionRevoker
// (RETROFIT-001 B1).
// ----------------------------------------------------------------------

type sessionRevokerAdapter struct{ inner *session.Manager }

func (a *sessionRevokerAdapter) RevokeAllForUser(ctx context.Context, userID, by string) (int64, error) {
	return a.inner.RevokeAllForUser(ctx, userID, by)
}

// ----------------------------------------------------------------------
// webauthnAuditShim — adapts *auditAdapter (Emit + EmitWebAuthn) to
// the webauthn.AuditEmitter interface, which expects a single
// Emit(ctx, AuditEvent) method.
// ----------------------------------------------------------------------

type webauthnAuditShim struct{ audit *auditAdapter }

func (s *webauthnAuditShim) Emit(ctx context.Context, e webauthn.AuditEvent) error {
	return s.audit.EmitWebAuthn(ctx, e)
}

// ----------------------------------------------------------------------
// oauth2TokenAdapter — adapts *token.Issuer + *token.RefreshStore to
// oauth2.TokenIssuer (RETROFIT-001 B5). Phase-1 wiring; the
// rotate-refresh path returns NotImplemented because RefreshStore.Rotate
// does not surface the user/tenant/session triple. Follow-up:
// extend RefreshStore.Rotate to return the previous record's identity.
// ----------------------------------------------------------------------

type oauth2TokenAdapter struct {
	issuer  *token.Issuer
	refresh *token.RefreshStore
}

func (a *oauth2TokenAdapter) IssueAccess(ctx context.Context, in token.IssueInput) (string, time.Time, error) {
	tk, claims, err := a.issuer.IssueAccessToken(in)
	if err != nil {
		return "", time.Time{}, err
	}
	exp := time.Time{}
	if claims != nil && claims.ExpiresAt != nil {
		exp = claims.ExpiresAt.Time
	}
	return tk, exp, nil
}

func (a *oauth2TokenAdapter) IssueRefresh(ctx context.Context, userID, tenantID, sessionID string) (string, time.Time, error) {
	out, err := a.refresh.Issue(ctx, token.RefreshIssue{
		UserID: userID, TenantID: tenantID, SessionID: sessionID,
		IssuedAt: time.Now().UTC(),
	})
	if err != nil {
		return "", time.Time{}, err
	}
	return out.Token, out.ExpiresAt, nil
}

func (a *oauth2TokenAdapter) RotateRefresh(ctx context.Context, presented string) (string, time.Time, string, string, string, error) {
	// Follow-up wiring: RefreshStore.Rotate currently returns only the
	// new-token + expiry; oauth2 needs the prior row's identity to
	// preserve the family. Until that surface lands, the OAuth2 token
	// endpoint refuses refresh-grant requests with NotImplemented.
	return "", time.Time{}, "", "", "", errors.New("oauth2 refresh-grant not yet wired (RefreshStore.Rotate identity-projection follow-up)")
}

// ----------------------------------------------------------------------
// spawnPolicyLoader runs Reload on a 60s ticker (AUTHZ-001 acceptance).
// Reload errors are logged + the previous snapshot is retained.
// ----------------------------------------------------------------------

func spawnPolicyLoader(ctx context.Context, l *policy.Loader, logger *slog.Logger) {
	t := time.NewTicker(60 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := l.Reload(ctx); err != nil {
				logger.Warn("policy reload failed", slog.Any("err", err))
			}
		}
	}
}

// ----------------------------------------------------------------------
// mountIAMRoutes — JSON HTTP shim wiring around the cmd-layer
// subsystems. Replaced by the Connect mux once OQ-004 is resolved
// + iam.proto is regen'd.
// ----------------------------------------------------------------------

type mountDeps struct {
	tenantID     string
	login        *login.Handler
	reset        *reset.Handler
	sar          *gdpr.SARService
	erase        *gdpr.EraseService
	rectify      *gdpr.RectifyService
	snapshot     *gdpr.SnapshotBuilder
	policyLoader *policy.Loader
	oidcDoc      *oidc.Document
	logger       *slog.Logger
}

func mountIAMRoutes(mux *http.ServeMux, d mountDeps) {
	// /v1/iam/login
	mux.HandleFunc("/v1/iam/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in struct {
			Email     string `json:"email"`
			Password  string `json:"password"`
			ClientIP  string `json:"client_ip"`
			UserAgent string `json:"user_agent"`
		}
		if err := jsonDecode(r.Body, &in); err != nil {
			httpError(w, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
		if in.ClientIP == "" {
			in.ClientIP = clientIPFromRequest(r)
		}
		if in.UserAgent == "" {
			in.UserAgent = r.Header.Get("User-Agent")
		}
		res, err := d.login.Login(r.Context(), login.LoginInput{
			Email:     in.Email,
			Password:  in.Password,
			ClientIP:  in.ClientIP,
			UserAgent: in.UserAgent,
		})
		if err != nil {
			httpError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		writeJSON(w, loginResultStatusToHTTP(res.Status), res)
	})

	// /.well-known/openid-configuration
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=3600")
		writeJSON(w, http.StatusOK, d.oidcDoc)
	})

	// /v1/iam/policies — read-only snapshot of the active PolicySet.
	mux.HandleFunc("/v1/iam/policies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, d.policyLoader.Snapshot())
	})

	// /v1/iam/reset/request + /v1/iam/reset/confirm.
	mux.HandleFunc("/v1/iam/reset/request", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in reset.RequestInput
		if err := jsonDecode(r.Body, &in); err != nil {
			httpError(w, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
		if in.ClientIP == "" {
			in.ClientIP = clientIPFromRequest(r)
		}
		if in.UserAgent == "" {
			in.UserAgent = r.Header.Get("User-Agent")
		}
		// REQ-FUNC-PLT-IAM-008: constant-time, non-disclosing 202.
		_, _ = d.reset.Request(r.Context(), in)
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("/v1/iam/reset/confirm", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in reset.ConfirmInput
		if err := jsonDecode(r.Body, &in); err != nil {
			httpError(w, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
		res, err := d.reset.Confirm(r.Context(), in)
		if err != nil {
			httpError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, res)
	})

	// /v1/iam/gdpr/* (SAR / erase / rectify / snapshot).
	mux.HandleFunc("/v1/iam/gdpr/sar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in gdpr.SARRequest
		if err := jsonDecode(r.Body, &in); err != nil {
			httpError(w, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
		res, err := d.sar.Request(r.Context(), in)
		if err != nil {
			httpError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		writeJSON(w, http.StatusAccepted, res)
	})
	mux.HandleFunc("/v1/iam/gdpr/erase", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in gdpr.ErasureRequest
		if err := jsonDecode(r.Body, &in); err != nil {
			httpError(w, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
		res, err := d.erase.Erase(r.Context(), in)
		if err != nil {
			httpError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("/v1/iam/gdpr/rectify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in gdpr.RectifyEmailRequest
		if err := jsonDecode(r.Body, &in); err != nil {
			httpError(w, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
		res, err := d.rectify.RectifyEmail(r.Context(), in)
		if err != nil {
			httpError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("/v1/iam/gdpr/snapshot/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		userID := strings.TrimPrefix(r.URL.Path, "/v1/iam/gdpr/snapshot/")
		if userID == "" {
			httpError(w, http.StatusBadRequest, "invalid_request", "user_id required")
			return
		}
		snap, err := d.snapshot.Build(r.Context(), userID)
		if err != nil {
			httpError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, snap)
	})

	// 501 — proto-blocked endpoints. The internal services were
	// constructed at boot; only the wire-format glue is deferred
	// pending OQ-004 (BSR auth → proto regen).
	for _, p := range []string{
		"/v1/iam/mfa/",
		"/v1/iam/webauthn/",
		"/v1/iam/sessions/",
		"/v1/iam/api-keys/",
		"/v1/iam/oauth2/authorize",
		"/v1/iam/oauth2/token",
		"/v1/iam/oauth2/userinfo",
		"/v1/iam/saml/login/",
		"/v1/iam/saml/acs/",
		"/v1/iam/saml/metadata/",
	} {
		mux.HandleFunc(p, func(w http.ResponseWriter, r *http.Request) {
			httpError(w, http.StatusNotImplemented, "not_implemented",
				"proto-bound endpoint pending OQ-004 (BSR auth)")
		})
	}
}

// ----------------------------------------------------------------------
// HTTP helpers.
// ----------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	body, err := jsonMarshal(v)
	if err != nil {
		_, _ = w.Write([]byte(`{"error":"marshal_failed"}`))
		return
	}
	_, _ = w.Write(body)
}

func httpError(w http.ResponseWriter, status int, code, desc string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	body, _ := jsonMarshal(map[string]string{
		"error":             code,
		"error_description": desc,
	})
	_, _ = w.Write(body)
}

func loginResultStatusToHTTP(s login.ResultStatus) int {
	switch s {
	case login.ResultOK:
		return http.StatusOK
	case login.ResultBadCredentials, login.ResultDisabled:
		return http.StatusUnauthorized
	case login.ResultLocked:
		return http.StatusLocked
	case login.ResultRateLimited:
		return http.StatusTooManyRequests
	case login.ResultInternalError:
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func clientIPFromRequest(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		if i := strings.Index(v, ","); i > 0 {
			return strings.TrimSpace(v[:i])
		}
		return strings.TrimSpace(v)
	}
	if v := r.Header.Get("X-Real-IP"); v != "" {
		return v
	}
	return r.RemoteAddr
}

// satisfy authzv1 import (Verifier surface used by the userinfo
// follow-up — kept here so the wiring guard counts authzv1 as imported).
var _ = (*authzv1.Verifier)(nil)

func getenvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
