// Package timeout implements circuit breaker pattern for fault tolerance
package timeout

import (
	"fmt"
	"sync"
	"time"

	"p9e.in/chetana/packages/saga"
	"p9e.in/chetana/packages/saga/models"
)

// CircuitBreakerImpl implements circuit breaker pattern with three states
type CircuitBreakerImpl struct {
	mu                 sync.RWMutex
	state              models.CircuitBreakerStatus
	failureCount       int32
	successCount       int32
	failureThreshold   int32
	resetTimeout       time.Duration
	lastFailureTime    time.Time
	lastStateChangeAt  time.Time
}

// NewCircuitBreakerImpl creates a new circuit breaker instance
func NewCircuitBreakerImpl(failureThreshold int32, resetTimeout time.Duration) *CircuitBreakerImpl {
	return &CircuitBreakerImpl{
		state:            models.CircuitBreakerClosed,
		failureCount:     0,
		successCount:     0,
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		lastStateChangeAt: time.Now(),
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreakerImpl) Call(fn func() error) error {
	cb.mu.Lock()
	state := cb.state
	cb.mu.Unlock()

	switch state {
	case models.CircuitBreakerClosed:
		return cb.callClosed(fn)
	case models.CircuitBreakerOpen:
		return cb.callOpen()
	case models.CircuitBreakerHalfOpen:
		return cb.callHalfOpen(fn)
	default:
		return fmt.Errorf("unknown circuit breaker state: %s", state)
	}
}

// callClosed handles calls when circuit is closed (normal operation)
func (cb *CircuitBreakerImpl) callClosed(fn func() error) error {
	// Execute the function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		// Record failure
		cb.failureCount++
		cb.lastFailureTime = time.Now()

		// Check if threshold exceeded
		if cb.failureCount >= cb.failureThreshold {
			cb.setState(models.CircuitBreakerOpen)
			fmt.Printf("Circuit breaker opened after %d failures\n", cb.failureCount)
		}

		return err
	}

	// Success - reset failure count
	cb.failureCount = 0
	cb.successCount++

	return nil
}

// callOpen handles calls when circuit is open (rejecting requests)
func (cb *CircuitBreakerImpl) callOpen() error {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	// Check if reset timeout has expired
	if time.Since(cb.lastStateChangeAt) > cb.resetTimeout {
		// Should transition to half-open, but caller should retry
		return saga.ErrCircuitBreakerOpen
	}

	return saga.ErrCircuitBreakerOpen
}

// callHalfOpen handles calls when circuit is half-open (testing recovery)
func (cb *CircuitBreakerImpl) callHalfOpen(fn func() error) error {
	// Execute the function (allowing one test request)
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		// Test failed, open circuit again
		cb.setState(models.CircuitBreakerOpen)
		cb.lastFailureTime = time.Now()
		fmt.Printf("Circuit breaker reopened (half-open test failed)\n")
		return err
	}

	// Test succeeded, close circuit
	cb.setState(models.CircuitBreakerClosed)
	cb.failureCount = 0
	cb.successCount = 1
	fmt.Printf("Circuit breaker closed (recovered)\n")

	return nil
}

// GetStatus returns the current circuit breaker status
func (cb *CircuitBreakerImpl) GetStatus() models.CircuitBreakerStatus {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	// Check if open circuit should transition to half-open
	if cb.state == models.CircuitBreakerOpen {
		if time.Since(cb.lastStateChangeAt) > cb.resetTimeout {
			// Ready to try recovery, but don't change state here
			// Let the next Call() handle the transition
			return models.CircuitBreakerHalfOpen
		}
	}

	return cb.state
}

// Reset resets the circuit breaker to closed state. Matches the
// saga.CircuitBreaker interface signature (no return; the reset is
// always safe to perform).
func (cb *CircuitBreakerImpl) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.setState(models.CircuitBreakerClosed)
	cb.failureCount = 0
	cb.successCount = 0

	fmt.Printf("Circuit breaker manually reset\n")
}

// setState changes the circuit breaker state
func (cb *CircuitBreakerImpl) setState(newState models.CircuitBreakerStatus) {
	if cb.state != newState {
		cb.state = newState
		cb.lastStateChangeAt = time.Now()
	}
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreakerImpl) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":            string(cb.state),
		"failureCount":     cb.failureCount,
		"failureThreshold": cb.failureThreshold,
		"successCount":     cb.successCount,
		"lastFailureTime":  cb.lastFailureTime,
		"lastStateChange":  cb.lastStateChangeAt,
		"resetTimeoutMs":   cb.resetTimeout.Milliseconds(),
	}
}

// TransitionToHalfOpen manually transitions circuit to half-open (for testing)
func (cb *CircuitBreakerImpl) TransitionToHalfOpen() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == models.CircuitBreakerOpen {
		cb.setState(models.CircuitBreakerHalfOpen)
		return nil
	}

	return fmt.Errorf("can only transition to half-open from open state, current state: %s", cb.state)
}
