// Package timeout implements timeout and retry management for saga steps
package timeout

import (
	"context"
	"fmt"
	"sync"
	"time"

	"p9e.in/samavaya/packages/saga"
)

// TimeoutHandlerImpl manages step timeouts and retry configurations
type TimeoutHandlerImpl struct {
	mu                       sync.RWMutex
	activeTimeouts           map[string]*TimeoutRecord // key: "sagaID:stepNum"
	retryStrategies          map[string]*saga.RetryConfiguration
	defaultRetryConfig       *saga.RetryConfiguration
	timeoutCheckInterval     time.Duration
	circuitBreakerInstances  map[string]*CircuitBreakerImpl
}

// TimeoutRecord tracks an active timeout
type TimeoutRecord struct {
	SagaID      string
	StepNum     int
	TimeoutAt   time.Time
	CancelledAt *time.Time
	CreatedAt   time.Time
}

// NewTimeoutHandlerImpl creates a new timeout handler instance
func NewTimeoutHandlerImpl(
	defaultRetryConfig *saga.RetryConfiguration,
	retryStrategies map[string]*saga.RetryConfiguration,
) *TimeoutHandlerImpl {
	if defaultRetryConfig == nil {
		defaultRetryConfig = &saga.RetryConfiguration{
			MaxRetries:         3,
			InitialBackoffMs:   1000,
			MaxBackoffMs:       30000,
			BackoffMultiplier:  2.0,
			JitterFraction:     0.1,
		}
	}

	if retryStrategies == nil {
		retryStrategies = make(map[string]*saga.RetryConfiguration)
	}

	return &TimeoutHandlerImpl{
		activeTimeouts:          make(map[string]*TimeoutRecord),
		retryStrategies:         retryStrategies,
		defaultRetryConfig:      defaultRetryConfig,
		timeoutCheckInterval:    100 * time.Millisecond,
		circuitBreakerInstances: make(map[string]*CircuitBreakerImpl),
	}
}

// SetupStepTimeout sets up a timeout for a saga step
func (h *TimeoutHandlerImpl) SetupStepTimeout(
	ctx context.Context,
	sagaID string,
	stepNum int,
	timeoutSeconds int32,
) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 1. Validate inputs
	if sagaID == "" {
		return fmt.Errorf("saga ID cannot be empty")
	}
	if stepNum < 1 {
		return fmt.Errorf("step number must be positive")
	}
	if timeoutSeconds <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	// 2. Create timeout record
	key := generateTimeoutKey(sagaID, stepNum)
	now := time.Now()
	record := &TimeoutRecord{
		SagaID:    sagaID,
		StepNum:   stepNum,
		TimeoutAt: now.Add(time.Duration(timeoutSeconds) * time.Second),
		CreatedAt: now,
	}

	// 3. Store active timeout
	h.activeTimeouts[key] = record

	// 4. Schedule timeout check (could be moved to background worker)
	go h.checkTimeout(ctx, sagaID, stepNum, record.TimeoutAt)

	return nil
}

// CancelStepTimeout cancels a timeout for a completed step
func (h *TimeoutHandlerImpl) CancelStepTimeout(sagaID string, stepNum int) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 1. Find timeout record
	key := generateTimeoutKey(sagaID, stepNum)
	record, exists := h.activeTimeouts[key]
	if !exists {
		// Timeout not found, which is okay - might have already expired
		return nil
	}

	// 2. Mark as cancelled
	now := time.Now()
	record.CancelledAt = &now

	return nil
}

// CheckExpired checks if a saga/step has expired
func (h *TimeoutHandlerImpl) CheckExpired(sagaID string, stepNum int) (bool, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 1. Find timeout record
	key := generateTimeoutKey(sagaID, stepNum)
	record, exists := h.activeTimeouts[key]
	if !exists {
		return false, nil // No timeout set
	}

	// 2. Check if cancelled
	if record.CancelledAt != nil {
		return false, nil // Cancelled, not expired
	}

	// 3. Check if expired
	return time.Now().After(record.TimeoutAt), nil
}

// GetRetryConfig returns retry configuration for a saga type and step
func (h *TimeoutHandlerImpl) GetRetryConfig(sagaType string, stepNum int) (*saga.RetryConfiguration, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 1. Check for saga-type specific strategy
	if strategy, exists := h.retryStrategies[sagaType]; exists {
		return strategy, nil
	}

	// 2. Return default configuration
	return h.defaultRetryConfig, nil
}

// RegisterRetryStrategy registers a custom retry strategy for a saga type
func (h *TimeoutHandlerImpl) RegisterRetryStrategy(
	sagaType string,
	config *saga.RetryConfiguration,
) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if sagaType == "" {
		return fmt.Errorf("saga type cannot be empty")
	}
	if config == nil {
		return fmt.Errorf("retry config cannot be nil")
	}

	h.retryStrategies[sagaType] = config
	return nil
}

// GetCircuitBreaker gets or creates a circuit breaker for a service
func (h *TimeoutHandlerImpl) GetCircuitBreaker(serviceName string) saga.CircuitBreaker {
	h.mu.Lock()
	defer h.mu.Unlock()

	cb, exists := h.circuitBreakerInstances[serviceName]
	if !exists {
		cb = NewCircuitBreakerImpl(5, 60*time.Second)
		h.circuitBreakerInstances[serviceName] = cb
	}

	return cb
}

// GetActiveTimeoutCount returns number of active timeouts (for debugging)
func (h *TimeoutHandlerImpl) GetActiveTimeoutCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	now := time.Now()
	for _, record := range h.activeTimeouts {
		if record.CancelledAt == nil && now.Before(record.TimeoutAt) {
			count++
		}
	}

	return count
}

// GetExpiredCount returns number of expired timeouts
func (h *TimeoutHandlerImpl) GetExpiredCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	now := time.Now()
	for _, record := range h.activeTimeouts {
		if record.CancelledAt == nil && now.After(record.TimeoutAt) {
			count++
		}
	}

	return count
}

// CleanupExpiredTimeouts removes expired timeout records
func (h *TimeoutHandlerImpl) CleanupExpiredTimeouts() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	count := 0
	now := time.Now()

	for key, record := range h.activeTimeouts {
		if now.After(record.TimeoutAt) {
			delete(h.activeTimeouts, key)
			count++
		}
	}

	return count
}

// checkTimeout runs as a goroutine to check if timeout has expired
func (h *TimeoutHandlerImpl) checkTimeout(ctx context.Context, sagaID string, stepNum int, timeoutAt time.Time) {
	timer := time.NewTimer(time.Until(timeoutAt))
	defer timer.Stop()

	select {
	case <-timer.C:
		// Timeout expired, mark it
		h.mu.Lock()
		key := generateTimeoutKey(sagaID, stepNum)
		record, exists := h.activeTimeouts[key]
		if exists && record.CancelledAt == nil {
			// Log timeout event - in real implementation, would publish event
			fmt.Printf("Step timeout expired: saga=%s, step=%d\n", sagaID, stepNum)
		}
		h.mu.Unlock()
	case <-ctx.Done():
		// Context cancelled
		return
	}
}

// Helper function to generate timeout key
func generateTimeoutKey(sagaID string, stepNum int) string {
	return fmt.Sprintf("timeout:%s:%d", sagaID, stepNum)
}
