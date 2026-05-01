package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"p9e.in/samavaya/packages/p9log"
	"p9e.in/samavaya/packages/healthcheck/coordinator"
)

// HealthServer implements HTTP API for health status.
// `logger` is *p9log.Helper (B.1 sweep).
type HealthServer struct {
	coordinator *coordinator.Coordinator
	logger      *p9log.Helper
	mux         *http.ServeMux
}

// New creates a new health server
func New(coord *coordinator.Coordinator, logger p9log.Logger) *HealthServer {
	s := &HealthServer{
		coordinator: coord,
		logger:      p9log.NewHelper(logger),
		mux:         http.NewServeMux(),
	}

	// Register handlers
	s.mux.HandleFunc("GET /v1/health/summary", s.handleGetSummary)
	s.mux.HandleFunc("GET /v1/health/service/{service}", s.handleGetService)
	s.mux.HandleFunc("GET /v1/health/service/{service}/instance/{instance}", s.handleGetInstance)
	s.mux.HandleFunc("GET /v1/health/live", s.handleLive)
	s.mux.HandleFunc("GET /v1/health/ready", s.handleReady)

	return s
}

// ServeHTTP implements http.Handler
func (s *HealthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Response types

type SummaryResponse struct {
	Status                string    `json:"status"`
	HealthyServices       int       `json:"healthy_services"`
	UnhealthyServices     int       `json:"unhealthy_services"`
	TotalServices         int       `json:"total_services"`
	HealthyInstances      int       `json:"healthy_instances"`
	UnhealthyInstances    int       `json:"unhealthy_instances"`
	TotalInstances        int       `json:"total_instances"`
	HealthPercent         int       `json:"health_percent"`
	GeneratedAt           string    `json:"generated_at"`
}

type ServiceResponse struct {
	ServiceName         string                      `json:"service_name"`
	Status              string                      `json:"status"`
	HealthyInstances    int                         `json:"healthy_instances"`
	UnhealthyInstances  int                         `json:"unhealthy_instances"`
	TotalInstances      int                         `json:"total_instances"`
	HealthPercent       int                         `json:"health_percent"`
	Instances           map[string]*InstanceResponse `json:"instances"`
	UpdatedAt           string                      `json:"updated_at"`
}

type InstanceResponse struct {
	InstanceID          string                 `json:"instance_id"`
	Status              string                 `json:"status"`
	LastSuccessfulCheck string                 `json:"last_successful_check,omitempty"`
	LastFailedCheck     string                 `json:"last_failed_check,omitempty"`
	FailureCount        int                    `json:"failure_count"`
	SuccessCount        int                    `json:"success_count"`
	LastError           string                 `json:"last_error,omitempty"`
	Details             map[string]interface{} `json:"details,omitempty"`
	UpdatedAt           string                 `json:"updated_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// Handlers

func (s *HealthServer) handleGetSummary(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	summary := s.coordinator.GetSummary(ctx)

	status := "healthy"
	if summary.UnhealthyServices > 0 {
		status = "degraded"
	}
	if summary.HealthyServices == 0 {
		status = "unhealthy"
	}

	resp := SummaryResponse{
		Status:             status,
		HealthyServices:    summary.HealthyServices,
		UnhealthyServices:  summary.UnhealthyServices,
		TotalServices:      summary.TotalServices,
		HealthyInstances:   summary.HealthyInstances,
		UnhealthyInstances: summary.UnhealthyInstances,
		TotalInstances:     summary.TotalInstances,
		HealthPercent:      summary.HealthPercent,
		GeneratedAt:        summary.GeneratedAt.Format(time.RFC3339),
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *HealthServer) handleGetService(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	service := r.PathValue("service")
	if service == "" {
		s.writeError(w, http.StatusBadRequest, "missing_field", "service is required")
		return
	}

	health, err := s.coordinator.GetStatus(ctx, service)
	if err != nil {
		s.logger.Error("failed to get service status",
			"service", service,
			"error", err,
		)
		s.writeError(w, http.StatusNotFound, "not_found", service+" service not found")
		return
	}

	instances := make(map[string]*InstanceResponse)
	if health.Instances != nil {
		for instID, instHealth := range health.Instances {
			instances[instID] = &InstanceResponse{
				InstanceID:          instHealth.InstanceID,
				Status:              string(instHealth.Status),
				LastSuccessfulCheck: instHealth.LastSuccessfulCheck.Format(time.RFC3339),
				LastFailedCheck:     instHealth.LastFailedCheck.Format(time.RFC3339),
				FailureCount:        instHealth.FailureCount,
				SuccessCount:        instHealth.SuccessCount,
				LastError:           instHealth.LastError,
				Details:             instHealth.Details,
				UpdatedAt:           instHealth.UpdatedAt.Format(time.RFC3339),
			}
		}
	}

	resp := ServiceResponse{
		ServiceName:        health.ServiceName,
		Status:             string(health.Status),
		HealthyInstances:   health.HealthyInstances,
		UnhealthyInstances: health.UnhealthyInstances,
		TotalInstances:     health.TotalInstances,
		HealthPercent:      health.HealthPercent,
		Instances:          instances,
		UpdatedAt:          health.UpdatedAt.Format(time.RFC3339),
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *HealthServer) handleGetInstance(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	service := r.PathValue("service")
	instance := r.PathValue("instance")

	if service == "" || instance == "" {
		s.writeError(w, http.StatusBadRequest, "missing_field", "service and instance are required")
		return
	}

	health, err := s.coordinator.GetInstanceStatus(ctx, service, instance)
	if err != nil {
		s.logger.Error("failed to get instance status",
			"service", service,
			"instance", instance,
			"error", err,
		)
		s.writeError(w, http.StatusNotFound, "not_found", "service or instance not found")
		return
	}

	resp := InstanceResponse{
		InstanceID:          health.InstanceID,
		Status:              string(health.Status),
		LastSuccessfulCheck: health.LastSuccessfulCheck.Format(time.RFC3339),
		LastFailedCheck:     health.LastFailedCheck.Format(time.RFC3339),
		FailureCount:        health.FailureCount,
		SuccessCount:        health.SuccessCount,
		LastError:           health.LastError,
		Details:             health.Details,
		UpdatedAt:           health.UpdatedAt.Format(time.RFC3339),
	}

	s.writeJSON(w, http.StatusOK, resp)
}

// handleLive checks if coordinator is running
func (s *HealthServer) handleLive(w http.ResponseWriter, r *http.Request) {
	// Liveness check - just verify we're running
	resp := map[string]interface{}{
		"status": "alive",
		"time":   time.Now().Format(time.RFC3339),
	}
	s.writeJSON(w, http.StatusOK, resp)
}

// handleReady checks if we're ready to serve
func (s *HealthServer) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	summary := s.coordinator.GetSummary(ctx)

	if summary.TotalServices == 0 || summary.HealthyServices == 0 {
		resp := map[string]interface{}{
			"ready": false,
			"reason": "no healthy services",
		}
		s.writeJSON(w, http.StatusServiceUnavailable, resp)
		return
	}

	resp := map[string]interface{}{
		"ready": true,
		"healthy_services": summary.HealthyServices,
		"total_services": summary.TotalServices,
	}
	s.writeJSON(w, http.StatusOK, resp)
}

// Helper methods

func (s *HealthServer) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *HealthServer) writeError(w http.ResponseWriter, status int, code string, message string) {
	resp := ErrorResponse{
		Error:   code,
		Message: message,
		Code:    code,
	}
	s.writeJSON(w, status, resp)
}
