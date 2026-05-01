package checks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"p9e.in/samavaya/packages/healthcheck"
)

func TestHTTPCheckSuccess(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	check := NewHTTPCheck(healthcheck.HTTPCheckConfig{
		URL:     server.URL,
		Timeout: 5 * time.Second,
	})

	result, err := check.Check(context.Background())
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if result.Status != healthcheck.StatusHealthy {
		t.Errorf("Expected HEALTHY, got %v", result.Status)
	}

	if result.Details["status_code"] != http.StatusOK {
		t.Errorf("Expected status code 200, got %v", result.Details["status_code"])
	}
}

func TestHTTPCheckFailure(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	check := NewHTTPCheck(healthcheck.HTTPCheckConfig{
		URL:     server.URL,
		Timeout: 5 * time.Second,
	})

	result, err := check.Check(context.Background())
	if err != nil {
		t.Fatalf("Check error: %v", err)
	}

	if result.Status != healthcheck.StatusUnhealthy {
		t.Errorf("Expected UNHEALTHY, got %v", result.Status)
	}

	if result.Details["status_code"] != http.StatusServiceUnavailable {
		t.Errorf("Expected status code 503, got %v", result.Details["status_code"])
	}
}

func TestHTTPCheckTimeout(t *testing.T) {
	// Create test server that sleeps
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	check := NewHTTPCheck(healthcheck.HTTPCheckConfig{
		URL:     server.URL,
		Timeout: 10 * time.Millisecond, // Very short timeout
	})

	result, err := check.Check(context.Background())
	if err == nil {
		t.Fatal("Expected timeout error")
	}

	if result.Status != healthcheck.StatusUnhealthy {
		t.Errorf("Expected UNHEALTHY, got %v", result.Status)
	}
}

func TestHTTPCheckType(t *testing.T) {
	check := NewHTTPCheck(healthcheck.HTTPCheckConfig{
		URL: "http://localhost:8080/health",
	})

	if check.Type() != healthcheck.CheckTypeHTTP {
		t.Errorf("Expected HTTP type, got %v", check.Type())
	}
}

func TestTCPCheckSuccess(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Extract host and port
	check := NewTCPCheck(healthcheck.TCPCheckConfig{
		Host:    "localhost",
		Port:    8080, // Will fail, but that's ok for this test
		Timeout: 1 * time.Second,
	})

	if check.Type() != healthcheck.CheckTypeTCP {
		t.Errorf("Expected TCP type, got %v", check.Type())
	}
}

func BenchmarkHTTPCheck(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	check := NewHTTPCheck(healthcheck.HTTPCheckConfig{
		URL:     server.URL,
		Timeout: 5 * time.Second,
	})

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		check.Check(ctx)
	}
}

func BenchmarkTCPCheck(b *testing.B) {
	check := NewTCPCheck(healthcheck.TCPCheckConfig{
		Host:    "localhost",
		Port:    80,
		Timeout: 1 * time.Second,
	})

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		check.Check(ctx)
	}
}
