// Package timeout contains unit tests for timeout and circuit breaker
package timeout

import (
	"context"
	"errors"
	"testing"
	"time"

	"p9e.in/chetana/packages/saga"
	"p9e.in/chetana/packages/saga/models"
)

// Timeout Handler Tests

func TestTimeoutHandlerSetupStepTimeout_Success(t *testing.T) {
	defaultConfig := &saga.RetryConfiguration{
		MaxRetries:        3,
		InitialBackoffMs:  1000,
		MaxBackoffMs:      30000,
		BackoffMultiplier: 2.0,
	}

	handler := NewTimeoutHandlerImpl(defaultConfig, nil)

	ctx := context.Background()
	err := handler.SetupStepTimeout(ctx, "saga-123", 1, 60)

	if err != nil {
		t.Errorf("SetupStepTimeout failed: %v", err)
	}

	// Verify timeout was set
	count := handler.GetActiveTimeoutCount()
	if count != 1 {
		t.Errorf("Expected 1 active timeout, got %d", count)
	}
}

func TestTimeoutHandlerSetupStepTimeout_InvalidSagaID(t *testing.T) {
	handler := NewTimeoutHandlerImpl(nil, nil)

	ctx := context.Background()
	err := handler.SetupStepTimeout(ctx, "", 1, 60)

	if err == nil {
		t.Fatal("Expected error for empty saga ID")
	}
}

func TestTimeoutHandlerSetupStepTimeout_InvalidStepNum(t *testing.T) {
	handler := NewTimeoutHandlerImpl(nil, nil)

	ctx := context.Background()
	err := handler.SetupStepTimeout(ctx, "saga-123", 0, 60)

	if err == nil {
		t.Fatal("Expected error for invalid step number")
	}
}

func TestTimeoutHandlerCancelStepTimeout_Success(t *testing.T) {
	handler := NewTimeoutHandlerImpl(nil, nil)

	ctx := context.Background()
	handler.SetupStepTimeout(ctx, "saga-123", 1, 60)

	err := handler.CancelStepTimeout("saga-123", 1)

	if err != nil {
		t.Errorf("CancelStepTimeout failed: %v", err)
	}

	// Verify timeout was cancelled
	expired, err := handler.CheckExpired("saga-123", 1)
	if err != nil {
		t.Errorf("CheckExpired failed: %v", err)
	}

	if expired {
		t.Error("Expected cancelled timeout to return false for expired")
	}
}

func TestTimeoutHandlerCheckExpired_NotExpired(t *testing.T) {
	handler := NewTimeoutHandlerImpl(nil, nil)

	ctx := context.Background()
	handler.SetupStepTimeout(ctx, "saga-123", 1, 60)

	expired, err := handler.CheckExpired("saga-123", 1)

	if err != nil {
		t.Errorf("CheckExpired failed: %v", err)
	}

	if expired {
		t.Error("Expected timeout to not be expired")
	}
}

func TestTimeoutHandlerGetRetryConfig_Default(t *testing.T) {
	defaultConfig := &saga.RetryConfiguration{
		MaxRetries:        3,
		InitialBackoffMs:  1000,
		MaxBackoffMs:      30000,
		BackoffMultiplier: 2.0,
	}

	handler := NewTimeoutHandlerImpl(defaultConfig, nil)

	config, err := handler.GetRetryConfig("SAGA-TEST", 1)

	if err != nil {
		t.Errorf("GetRetryConfig failed: %v", err)
	}

	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", config.MaxRetries)
	}
}

func TestTimeoutHandlerGetRetryConfig_Custom(t *testing.T) {
	defaultConfig := &saga.RetryConfiguration{
		MaxRetries:        3,
		InitialBackoffMs:  1000,
		MaxBackoffMs:      30000,
		BackoffMultiplier: 2.0,
	}

	customConfig := &saga.RetryConfiguration{
		MaxRetries:        5,
		InitialBackoffMs:  500,
		MaxBackoffMs:      20000,
		BackoffMultiplier: 1.5,
	}

	strategies := map[string]*saga.RetryConfiguration{
		"SAGA-S01": customConfig,
	}

	handler := NewTimeoutHandlerImpl(defaultConfig, strategies)

	config, err := handler.GetRetryConfig("SAGA-S01", 1)

	if err != nil {
		t.Errorf("GetRetryConfig failed: %v", err)
	}

	if config.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries 5, got %d", config.MaxRetries)
	}
}

func TestTimeoutHandlerRegisterRetryStrategy_Success(t *testing.T) {
	handler := NewTimeoutHandlerImpl(nil, nil)

	customConfig := &saga.RetryConfiguration{
		MaxRetries:        5,
		InitialBackoffMs:  500,
		MaxBackoffMs:      20000,
		BackoffMultiplier: 1.5,
	}

	err := handler.RegisterRetryStrategy("SAGA-S01", customConfig)

	if err != nil {
		t.Errorf("RegisterRetryStrategy failed: %v", err)
	}

	// Verify registration
	config, _ := handler.GetRetryConfig("SAGA-S01", 1)
	if config.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries 5, got %d", config.MaxRetries)
	}
}

func TestTimeoutHandlerGetActiveTimeoutCount(t *testing.T) {
	handler := NewTimeoutHandlerImpl(nil, nil)

	ctx := context.Background()
	handler.SetupStepTimeout(ctx, "saga-123", 1, 60)
	handler.SetupStepTimeout(ctx, "saga-123", 2, 60)
	handler.SetupStepTimeout(ctx, "saga-456", 1, 60)

	count := handler.GetActiveTimeoutCount()

	if count != 3 {
		t.Errorf("Expected 3 active timeouts, got %d", count)
	}
}

func TestTimeoutHandlerCleanupExpiredTimeouts(t *testing.T) {
	handler := NewTimeoutHandlerImpl(nil, nil)

	ctx := context.Background()
	// Setup timeouts with very short duration
	handler.SetupStepTimeout(ctx, "saga-123", 1, 1) // 1 second

	// Wait for expiration
	time.Sleep(1500 * time.Millisecond)

	cleaned := handler.CleanupExpiredTimeouts()

	if cleaned != 1 {
		t.Errorf("Expected 1 cleaned timeout, got %d", cleaned)
	}
}

// Circuit Breaker Tests

func TestCircuitBreakerCall_ClosedSuccess(t *testing.T) {
	cb := NewCircuitBreakerImpl(5, 60*time.Second)

	callCount := 0
	err := cb.Call(func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("Call failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}

	status := cb.GetStatus()
	if status != models.CircuitBreakerClosed {
		t.Errorf("Expected closed status, got %s", status)
	}
}

func TestCircuitBreakerCall_ClosedFailure(t *testing.T) {
	cb := NewCircuitBreakerImpl(3, 60*time.Second) // Threshold of 3 failures

	// Cause 3 failures
	for i := 0; i < 3; i++ {
		cb.Call(func() error {
			return errors.New("service error")
		})
	}

	// Circuit should now be open
	status := cb.GetStatus()
	if status != models.CircuitBreakerOpen {
		t.Errorf("Expected open status after threshold, got %s", status)
	}

	// Next call should be rejected
	err := cb.Call(func() error {
		return nil // This won't be called
	})

	if err == nil {
		t.Fatal("Expected error when circuit is open")
	}
}

func TestCircuitBreakerCall_OpenRejects(t *testing.T) {
	cb := NewCircuitBreakerImpl(1, 60*time.Second)

	// Cause 1 failure to open circuit
	cb.Call(func() error {
		return errors.New("error")
	})

	// Circuit is now open, should reject calls
	callCount := 0
	err := cb.Call(func() error {
		callCount++
		return nil
	})

	if err == nil {
		t.Fatal("Expected error from open circuit")
	}

	if callCount != 0 {
		t.Errorf("Expected no calls when open, got %d", callCount)
	}
}

func TestCircuitBreakerCall_HalfOpenSuccess(t *testing.T) {
	cb := NewCircuitBreakerImpl(1, 1*time.Millisecond) // Very short reset timeout

	// Open the circuit
	cb.Call(func() error {
		return errors.New("error")
	})

	// Wait for reset timeout
	time.Sleep(10 * time.Millisecond)

	// Transition to half-open
	cb.TransitionToHalfOpen()

	// Call should succeed and close circuit
	err := cb.Call(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Half-open call failed: %v", err)
	}

	status := cb.GetStatus()
	if status != models.CircuitBreakerClosed {
		t.Errorf("Expected closed status after recovery, got %s", status)
	}
}

func TestCircuitBreakerReset_Success(t *testing.T) {
	cb := NewCircuitBreakerImpl(1, 60*time.Second)

	// Open the circuit
	cb.Call(func() error {
		return errors.New("error")
	})

	// Reset — the interface dropped the error return on 2026-04-19 (B.8).
	cb.Reset()

	status := cb.GetStatus()
	if status != models.CircuitBreakerClosed {
		t.Errorf("Expected closed status after reset, got %s", status)
	}
}

func TestCircuitBreakerGetStats(t *testing.T) {
	cb := NewCircuitBreakerImpl(5, 60*time.Second)

	cb.Call(func() error {
		return nil
	})

	stats := cb.GetStats()

	if stats["failureCount"] != int32(0) {
		t.Errorf("Expected 0 failures, got %d", stats["failureCount"])
	}

	if stats["successCount"] != int32(1) {
		t.Errorf("Expected 1 success, got %d", stats["successCount"])
	}
}

func TestCircuitBreakerFailureThreshold(t *testing.T) {
	cb := NewCircuitBreakerImpl(5, 60*time.Second)

	// Cause exactly 5 failures (threshold)
	for i := 0; i < 5; i++ {
		cb.Call(func() error {
			return errors.New("error")
		})
	}

	// Circuit should be open
	status := cb.GetStatus()
	if status != models.CircuitBreakerOpen {
		t.Errorf("Expected open after threshold, got %s", status)
	}
}
