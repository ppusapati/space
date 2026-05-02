package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"p9e.in/chetana/packages/p9log"
	"p9e.in/chetana/packages/registry"
)

// Server implements HTTP API for service registry.
// `logger` is *p9log.Helper (B.1 sweep).
type Server struct {
	registry registry.ServiceRegistry
	logger   *p9log.Helper
	mux      *http.ServeMux
}

// New creates a new API server
func New(reg registry.ServiceRegistry, logger p9log.Logger) *Server {
	s := &Server{
		registry: reg,
		logger:   p9log.NewHelper(logger),
		mux:      http.NewServeMux(),
	}

	// Register handlers
	s.mux.HandleFunc("POST /v1/services/register", s.handleRegister)
	s.mux.HandleFunc("DELETE /v1/services/{instanceId}", s.handleDeregister)
	s.mux.HandleFunc("GET /v1/services/{serviceName}", s.handleGetInstances)
	s.mux.HandleFunc("GET /v1/services/{serviceName}/instances", s.handleGetInstances)
	s.mux.HandleFunc("POST /v1/services/{instanceId}/heartbeat", s.handleHeartbeat)
	s.mux.HandleFunc("GET /v1/health", s.handleHealth)

	return s
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// RegisterRequest is the request body for registering a service
type RegisterRequest struct {
	ServiceName string            `json:"service_name"`
	Host        string            `json:"host"`
	Port        int               `json:"port"`
	Version     string            `json:"version,omitempty"`
	Region      string            `json:"region,omitempty"`
	Zone        string            `json:"zone,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	IsExternal  bool              `json:"is_external,omitempty"`
	ExternalURL string            `json:"external_url,omitempty"`
}

// RegisterResponse is the response for service registration
type RegisterResponse struct {
	InstanceID  string `json:"instance_id"`
	ServiceName string `json:"service_name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Status      string `json:"status"`
}

// GetInstancesResponse is the response for getting service instances
type GetInstancesResponse struct {
	ServiceName string                      `json:"service_name"`
	Instances   []GetInstancesResponseItem `json:"instances"`
}

type GetInstancesResponseItem struct {
	InstanceID    string            `json:"instance_id"`
	Host          string            `json:"host"`
	Port          int               `json:"port"`
	Health        string            `json:"health"`
	Version       string            `json:"version"`
	Region        string            `json:"region"`
	Zone          string            `json:"zone"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	IsExternal    bool              `json:"is_external"`
	ExternalURL   string            `json:"external_url,omitempty"`
	LastHeartbeat string            `json:"last_heartbeat"`
}

// ErrorResponse is a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// HealthResponse is the response for health check
type HealthResponse struct {
	Status string `json:"status"`
}

// handleRegister handles service registration
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid_request", "Failed to decode request body")
		return
	}

	// Validate request
	if req.ServiceName == "" {
		s.writeError(w, http.StatusBadRequest, "missing_field", "service_name is required")
		return
	}
	if req.Host == "" {
		s.writeError(w, http.StatusBadRequest, "missing_field", "host is required")
		return
	}
	if req.Port <= 0 || req.Port > 65535 {
		s.writeError(w, http.StatusBadRequest, "invalid_field", "port must be between 1 and 65535")
		return
	}

	// Generate instance ID if not provided
	instanceID := req.ServiceName + "-" + time.Now().Format("20060102150405")

	// Create service instance
	instance := &registry.ServiceInstance{
		ID:          instanceID,
		ServiceName: req.ServiceName,
		Host:        req.Host,
		Port:        req.Port,
		Version:     req.Version,
		Region:      req.Region,
		Zone:        req.Zone,
		Metadata:    req.Metadata,
		IsExternal:  req.IsExternal,
		ExternalURL: req.ExternalURL,
		Health:      registry.HealthUp,
	}

	// Register with registry
	if err := s.registry.Register(ctx, instance); err != nil {
		s.logger.Error("failed to register instance",
			"service_name", req.ServiceName,
			"error", err,
		)
		s.writeError(w, http.StatusInternalServerError, "registration_failed", err.Error())
		return
	}

	// Return success response
	resp := RegisterResponse{
		InstanceID:  instance.ID,
		ServiceName: instance.ServiceName,
		Host:        instance.Host,
		Port:        instance.Port,
		Status:      "registered",
	}

	s.writeJSON(w, http.StatusCreated, resp)
}

// handleDeregister handles service deregistration
func (s *Server) handleDeregister(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	instanceID := r.PathValue("instanceId")
	if instanceID == "" {
		s.writeError(w, http.StatusBadRequest, "missing_field", "instanceId is required")
		return
	}

	if err := s.registry.Deregister(ctx, instanceID); err != nil {
		s.logger.Error("failed to deregister instance",
			"instance_id", instanceID,
			"error", err,
		)
		s.writeError(w, http.StatusNotFound, "instance_not_found", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleGetInstances handles getting service instances
func (s *Server) handleGetInstances(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	serviceName := r.PathValue("serviceName")
	if serviceName == "" {
		s.writeError(w, http.StatusBadRequest, "missing_field", "serviceName is required")
		return
	}

	instances, err := s.registry.GetInstances(ctx, serviceName)
	if err != nil {
		s.logger.Error("failed to get instances",
			"service_name", serviceName,
			"error", err,
		)
		s.writeError(w, http.StatusInternalServerError, "query_failed", err.Error())
		return
	}

	// Convert to response format
	items := make([]GetInstancesResponseItem, len(instances))
	for i, inst := range instances {
		items[i] = GetInstancesResponseItem{
			InstanceID:    inst.ID,
			Host:          inst.Host,
			Port:          inst.Port,
			Health:        string(inst.Health),
			Version:       inst.Version,
			Region:        inst.Region,
			Zone:          inst.Zone,
			Metadata:      inst.Metadata,
			IsExternal:    inst.IsExternal,
			ExternalURL:   inst.ExternalURL,
			LastHeartbeat: inst.LastHeartbeat.Format(time.RFC3339),
		}
	}

	resp := GetInstancesResponse{
		ServiceName: serviceName,
		Instances:   items,
	}

	s.writeJSON(w, http.StatusOK, resp)
}

// handleHeartbeat handles service heartbeat
func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	instanceID := r.PathValue("instanceId")
	if instanceID == "" {
		s.writeError(w, http.StatusBadRequest, "missing_field", "instanceId is required")
		return
	}

	if err := s.registry.Heartbeat(ctx, instanceID); err != nil {
		s.logger.Error("failed to record heartbeat",
			"instance_id", instanceID,
			"error", err,
		)
		s.writeError(w, http.StatusNotFound, "instance_not_found", err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{
		"status": "heartbeat_recorded",
	})
}

// handleHealth handles registry health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status: "healthy",
	}
	s.writeJSON(w, http.StatusOK, resp)
}

// Helper methods

// writeJSON writes a JSON response
func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func (s *Server) writeError(w http.ResponseWriter, status int, code string, message string) {
	resp := ErrorResponse{
		Error:   code,
		Message: message,
		Code:    code,
	}
	s.writeJSON(w, status, resp)
}
