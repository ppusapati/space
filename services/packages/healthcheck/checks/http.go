package checks

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"p9e.in/samavaya/packages/healthcheck"
)

// HTTPCheck implements HTTP health checking
type HTTPCheck struct {
	URL          string
	Method       string
	Headers      map[string]string
	SuccessCodes []int
	Timeout      time.Duration
	client       *http.Client
	name         string
}

// NewHTTPCheck creates a new HTTP health check
func NewHTTPCheck(cfg healthcheck.HTTPCheckConfig) *HTTPCheck {
	check := &HTTPCheck{
		URL:          cfg.URL,
		Method:       cfg.Method,
		Headers:      cfg.Headers,
		SuccessCodes: cfg.SuccessCodes,
		Timeout:      cfg.Timeout,
		name:         fmt.Sprintf("http:%s", cfg.URL),
	}

	if check.Method == "" {
		check.Method = "GET"
	}

	if len(check.SuccessCodes) == 0 {
		check.SuccessCodes = []int{http.StatusOK}
	}

	if check.Timeout == 0 {
		check.Timeout = 5 * time.Second
	}

	check.client = &http.Client{
		Timeout: check.Timeout,
	}

	return check
}

// Check performs the HTTP health check
func (hc *HTTPCheck) Check(ctx context.Context) (*healthcheck.CheckResult, error) {
	start := time.Now()
	result := &healthcheck.CheckResult{
		CheckedAt: start,
		Details:   make(map[string]interface{}),
	}

	// Create request with context
	ctx, cancel := context.WithTimeout(ctx, hc.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, hc.Method, hc.URL, nil)
	if err != nil {
		result.Status = healthcheck.StatusUnhealthy
		result.Error = err.Error()
		result.Message = fmt.Sprintf("Failed to create request: %v", err)
		result.Duration = time.Since(start)
		return result, err
	}

	// Add headers
	for key, value := range hc.Headers {
		req.Header.Set(key, value)
	}

	// Perform request
	resp, err := hc.client.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		result.Status = healthcheck.StatusUnhealthy
		result.Error = err.Error()
		result.Message = fmt.Sprintf("Request failed: %v", err)
		result.Details["url"] = hc.URL
		return result, err
	}

	defer resp.Body.Close()

	// Check status code
	isSuccess := false
	for _, code := range hc.SuccessCodes {
		if resp.StatusCode == code {
			isSuccess = true
			break
		}
	}

	result.Details["status_code"] = resp.StatusCode
	result.Details["url"] = hc.URL
	result.Details["method"] = hc.Method
	result.Details["duration_ms"] = result.Duration.Milliseconds()

	if !isSuccess {
		result.Status = healthcheck.StatusUnhealthy
		result.Message = fmt.Sprintf("Unexpected status code: %d", resp.StatusCode)
		result.Error = fmt.Sprintf("Expected one of %v, got %d", hc.SuccessCodes, resp.StatusCode)
		return result, nil
	}

	// Read response body (limited)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*10)) // 10KB limit
	if err != nil {
		result.Status = healthcheck.StatusDegraded
		result.Message = "Failed to read response body"
		result.Details["body_error"] = err.Error()
	} else {
		result.Details["body_size"] = len(body)
	}

	result.Status = healthcheck.StatusHealthy
	result.Message = "Health check passed"

	return result, nil
}

// Type returns the check type
func (hc *HTTPCheck) Type() healthcheck.CheckType {
	return healthcheck.CheckTypeHTTP
}

// Name returns the check name
func (hc *HTTPCheck) Name() string {
	return hc.name
}

// SetName sets the check name
func (hc *HTTPCheck) SetName(name string) {
	hc.name = name
}
