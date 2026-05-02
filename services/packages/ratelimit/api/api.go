package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"p9e.in/chetana/packages/p9log"
	"p9e.in/chetana/packages/ratelimit/backend"
)

// PolicyServer implements HTTP API for rate limit policy management.
// `logger` is *p9log.Helper (B.1 sweep).
type PolicyServer struct {
	limiter *backend.PostgresRateLimiter
	logger  *p9log.Helper
	mux     *http.ServeMux
}

// New creates a new policy server
func New(limiter *backend.PostgresRateLimiter, logger p9log.Logger) *PolicyServer {
	s := &PolicyServer{
		limiter: limiter,
		logger:  p9log.NewHelper(logger),
		mux:     http.NewServeMux(),
	}

	// Register handlers
	s.mux.HandleFunc("GET /v1/ratelimit/stats/{key}", s.handleGetStats)
	s.mux.HandleFunc("POST /v1/ratelimit/reset/{key}", s.handleReset)
	s.mux.HandleFunc("PUT /v1/ratelimit/{key}", s.handleSetLimit)
	s.mux.HandleFunc("POST /v1/ratelimit/check/{key}", s.handleCheck)

	return s
}

// ServeHTTP implements http.Handler
func (s *PolicyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Request/Response types

type SetLimitRequest struct {
	LimitPerSecond int64 `json:"limit_per_second"`
}

type CheckRequest struct {
	// Number of tokens to check (default 1)
	Tokens int `json:"tokens"`
}

type CheckResponse struct {
	Allowed bool   `json:"allowed"`
	Key     string `json:"key"`
	Stats   *StatsResponse `json:"stats"`
}

type StatsResponse struct {
	Key           string                 `json:"key"`
	AllowedCount  int64                  `json:"allowed_count"`
	RejectedCount int64                  `json:"rejected_count"`
	CurrentLimit  int64                  `json:"current_limit"`
	WindowStart   string                 `json:"window_start"`
	WindowEnd     string                 `json:"window_end"`
	Metrics       map[string]interface{} `json:"metrics,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// Handlers

func (s *PolicyServer) handleGetStats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	key := r.PathValue("key")
	if key == "" {
		s.writeError(w, http.StatusBadRequest, "missing_field", "key is required")
		return
	}

	stats, err := s.limiter.GetStats(ctx, key)
	if err != nil {
		s.logger.Error("failed to get stats",
			"key", key,
			"error", err,
		)
		s.writeError(w, http.StatusInternalServerError, "query_failed", err.Error())
		return
	}

	resp := StatsResponse{
		Key:           stats.Key,
		AllowedCount:  stats.AllowedCount,
		RejectedCount: stats.RejectedCount,
		CurrentLimit:  stats.CurrentLimit,
		WindowStart:   stats.WindowStart.Format(time.RFC3339),
		WindowEnd:     stats.WindowEnd.Format(time.RFC3339),
		Metrics:       stats.Metrics,
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *PolicyServer) handleSetLimit(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	key := r.PathValue("key")
	if key == "" {
		s.writeError(w, http.StatusBadRequest, "missing_field", "key is required")
		return
	}

	var req SetLimitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid_request", "Failed to decode request body")
		return
	}

	if req.LimitPerSecond <= 0 {
		s.writeError(w, http.StatusBadRequest, "invalid_field", "limit_per_second must be positive")
		return
	}

	if err := s.limiter.SetLimit(ctx, key, req.LimitPerSecond); err != nil {
		s.logger.Error("failed to set limit",
			"key", key,
			"limit", req.LimitPerSecond,
			"error", err,
		)
		s.writeError(w, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	resp := map[string]interface{}{
		"key":                key,
		"limit_per_second":   req.LimitPerSecond,
		"status":             "updated",
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *PolicyServer) handleReset(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	key := r.PathValue("key")
	if key == "" {
		s.writeError(w, http.StatusBadRequest, "missing_field", "key is required")
		return
	}

	if err := s.limiter.Reset(ctx, key); err != nil {
		s.logger.Error("failed to reset rate limit",
			"key", key,
			"error", err,
		)
		s.writeError(w, http.StatusInternalServerError, "reset_failed", err.Error())
		return
	}

	resp := map[string]interface{}{
		"key":    key,
		"status": "reset",
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *PolicyServer) handleCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	key := r.PathValue("key")
	if key == "" {
		s.writeError(w, http.StatusBadRequest, "missing_field", "key is required")
		return
	}

	var req CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Tokens = 1 // Default to 1 token
	}

	if req.Tokens <= 0 {
		req.Tokens = 1
	}

	allowed, err := s.limiter.AllowN(ctx, key, req.Tokens)
	if err != nil {
		s.logger.Error("failed to check rate limit",
			"key", key,
			"error", err,
		)
		s.writeError(w, http.StatusInternalServerError, "check_failed", err.Error())
		return
	}

	stats, _ := s.limiter.GetStats(ctx, key)

	status := http.StatusOK
	if !allowed {
		status = http.StatusTooManyRequests
	}

	resp := CheckResponse{
		Allowed: allowed,
		Key:     key,
	}

	if stats != nil {
		resp.Stats = &StatsResponse{
			Key:           stats.Key,
			AllowedCount:  stats.AllowedCount,
			RejectedCount: stats.RejectedCount,
			CurrentLimit:  stats.CurrentLimit,
			WindowStart:   stats.WindowStart.Format(time.RFC3339),
			WindowEnd:     stats.WindowEnd.Format(time.RFC3339),
			Metrics:       stats.Metrics,
		}
	}

	s.writeJSON(w, status, resp)
}

// Helper methods

func (s *PolicyServer) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *PolicyServer) writeError(w http.ResponseWriter, status int, code string, message string) {
	resp := ErrorResponse{
		Error:   code,
		Message: message,
		Code:    code,
	}
	s.writeJSON(w, status, resp)
}
