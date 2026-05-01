package checks

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	"p9e.in/samavaya/packages/healthcheck"
)

// GRPCCheck implements gRPC health checking using Health proto
type GRPCCheck struct {
	Host    string
	Port    int
	Service string
	Timeout time.Duration
	name    string
	conn    *grpc.ClientConn
}

// NewGRPCCheck creates a new gRPC health check
func NewGRPCCheck(cfg healthcheck.GRPCCheckConfig) *GRPCCheck {
	check := &GRPCCheck{
		Host:    cfg.Host,
		Port:    cfg.Port,
		Service: cfg.Service,
		Timeout: cfg.Timeout,
		name:    fmt.Sprintf("grpc:%s:%d/%s", cfg.Host, cfg.Port, cfg.Service),
	}

	if check.Timeout == 0 {
		check.Timeout = 5 * time.Second
	}

	if check.Service == "" {
		check.Service = "" // Empty service checks overall server health
	}

	return check
}

// Check performs the gRPC health check
func (gc *GRPCCheck) Check(ctx context.Context) (*healthcheck.CheckResult, error) {
	start := time.Now()
	result := &healthcheck.CheckResult{
		CheckedAt: start,
		Details:   make(map[string]interface{}),
	}

	result.Details["host"] = gc.Host
	result.Details["port"] = gc.Port
	result.Details["service"] = gc.Service

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, gc.Timeout)
	defer cancel()

	// Connect to gRPC server
	address := fmt.Sprintf("%s:%d", gc.Host, gc.Port)
	conn, err := grpc.DialContext(ctx, address,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	result.Duration = time.Since(start)
	result.Details["duration_ms"] = result.Duration.Milliseconds()

	if err != nil {
		result.Status = healthcheck.StatusUnhealthy
		result.Error = err.Error()
		result.Message = fmt.Sprintf("Failed to dial gRPC server at %s: %v", address, err)
		return result, err
	}
	defer conn.Close()

	// Create health client
	client := grpc_health_v1.NewHealthClient(conn)

	// Check service health
	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: gc.Service,
	})

	if err != nil {
		// Handle specific gRPC errors
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Unimplemented {
			result.Status = healthcheck.StatusDegraded
			result.Message = "Service does not implement Health proto"
			result.Details["error"] = "Unimplemented"
			return result, nil
		}

		result.Status = healthcheck.StatusUnhealthy
		result.Error = err.Error()
		result.Message = fmt.Sprintf("Health check failed: %v", err)
		return result, err
	}

	// Check response status
	switch resp.Status {
	case grpc_health_v1.HealthCheckResponse_SERVING:
		result.Status = healthcheck.StatusHealthy
		result.Message = "Service is serving"

	case grpc_health_v1.HealthCheckResponse_NOT_SERVING:
		result.Status = healthcheck.StatusUnhealthy
		result.Message = "Service is not serving"
		result.Error = "Service not serving"

	case grpc_health_v1.HealthCheckResponse_UNKNOWN:
		result.Status = healthcheck.StatusUnknown
		result.Message = "Service health status unknown"

	case grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN:
		result.Status = healthcheck.StatusUnhealthy
		result.Message = fmt.Sprintf("Service %q is unknown", gc.Service)
		result.Error = "Service unknown"

	default:
		result.Status = healthcheck.StatusUnknown
		result.Message = fmt.Sprintf("Unknown status: %v", resp.Status)
	}

	result.Details["status"] = resp.Status.String()

	return result, nil
}

// Type returns the check type
func (gc *GRPCCheck) Type() healthcheck.CheckType {
	return healthcheck.CheckTypeGRPC
}

// Name returns the check name
func (gc *GRPCCheck) Name() string {
	return gc.name
}

// SetName sets the check name
func (gc *GRPCCheck) SetName(name string) {
	gc.name = name
}

// Close closes the gRPC connection
func (gc *GRPCCheck) Close() error {
	if gc.conn != nil {
		return gc.conn.Close()
	}
	return nil
}
