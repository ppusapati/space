package checks

import (
	"context"
	"fmt"
	"net"
	"time"

	"p9e.in/samavaya/packages/healthcheck"
)

// TCPCheck implements TCP port health checking
type TCPCheck struct {
	Host    string
	Port    int
	Timeout time.Duration
	name    string
}

// NewTCPCheck creates a new TCP health check
func NewTCPCheck(cfg healthcheck.TCPCheckConfig) *TCPCheck {
	check := &TCPCheck{
		Host:    cfg.Host,
		Port:    cfg.Port,
		Timeout: cfg.Timeout,
		name:    fmt.Sprintf("tcp:%s:%d", cfg.Host, cfg.Port),
	}

	if check.Timeout == 0 {
		check.Timeout = 3 * time.Second
	}

	return check
}

// Check performs the TCP health check
func (tc *TCPCheck) Check(ctx context.Context) (*healthcheck.CheckResult, error) {
	start := time.Now()
	result := &healthcheck.CheckResult{
		CheckedAt: start,
		Details:   make(map[string]interface{}),
	}

	result.Details["host"] = tc.Host
	result.Details["port"] = tc.Port

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, tc.Timeout)
	defer cancel()

	// Attempt TCP connection
	address := fmt.Sprintf("%s:%d", tc.Host, tc.Port)
	dialer := &net.Dialer{
		Timeout: tc.Timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	result.Duration = time.Since(start)
	result.Details["duration_ms"] = result.Duration.Milliseconds()

	if err != nil {
		result.Status = healthcheck.StatusUnhealthy
		result.Error = err.Error()
		result.Message = fmt.Sprintf("Failed to connect to %s: %v", address, err)
		return result, err
	}

	// Close connection
	if err := conn.Close(); err != nil {
		result.Status = healthcheck.StatusDegraded
		result.Message = "Connected but failed to close"
		result.Details["close_error"] = err.Error()
		return result, nil
	}

	result.Status = healthcheck.StatusHealthy
	result.Message = fmt.Sprintf("Connected to %s", address)

	return result, nil
}

// Type returns the check type
func (tc *TCPCheck) Type() healthcheck.CheckType {
	return healthcheck.CheckTypeTCP
}

// Name returns the check name
func (tc *TCPCheck) Name() string {
	return tc.name
}

// SetName sets the check name
func (tc *TCPCheck) SetName(name string) {
	tc.name = name
}
