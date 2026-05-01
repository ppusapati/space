// Package compensation provides compensation log management
package compensation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"p9e.in/samavaya/packages/saga/models"
)

// CompensationLogRepositoryImpl manages compensation logs in memory
// In production, this would be backed by PostgreSQL via SQLC
type CompensationLogRepositoryImpl struct {
	mu   sync.RWMutex
	logs map[string][]*CompensationLog // key: sagaID
}

// CompensationLog represents a compensation action
type CompensationLog struct {
	ID                  string
	SagaID              string
	StepNumber          int32
	CompensationStepNum int32
	Status              models.StepExecutionStatus
	ErrorMessage        string
	ExecutedAt          time.Time
	ExecutionTimeMs     int64
}

// NewCompensationLogRepositoryImpl creates a new compensation log repository
func NewCompensationLogRepositoryImpl() *CompensationLogRepositoryImpl {
	return &CompensationLogRepositoryImpl{
		logs: make(map[string][]*CompensationLog),
	}
}

// CreateCompensationLog creates a new compensation log entry
func (r *CompensationLogRepositoryImpl) CreateCompensationLog(
	ctx context.Context,
	sagaID string,
	log *CompensationLog,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 1. Validate inputs
	if sagaID == "" {
		return fmt.Errorf("saga ID cannot be empty")
	}
	if log == nil {
		return fmt.Errorf("log cannot be nil")
	}

	// 2. Generate ID if not set
	if log.ID == "" {
		log.ID = fmt.Sprintf("comp-log-%d", time.Now().UnixNano())
	}

	// 3. Set saga ID
	log.SagaID = sagaID

	// 4. Add to logs
	r.logs[sagaID] = append(r.logs[sagaID], log)

	return nil
}

// GetCompensationLogs retrieves all compensation logs for a saga
func (r *CompensationLogRepositoryImpl) GetCompensationLogs(
	ctx context.Context,
	sagaID string,
) ([]*CompensationLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. Validate input
	if sagaID == "" {
		return nil, fmt.Errorf("saga ID cannot be empty")
	}

	// 2. Get logs
	logs, exists := r.logs[sagaID]
	if !exists {
		return make([]*CompensationLog, 0), nil
	}

	// 3. Return copy of logs
	result := make([]*CompensationLog, len(logs))
	copy(result, logs)

	return result, nil
}

// GetCompensationLogsByStep retrieves compensation logs for a specific step
func (r *CompensationLogRepositoryImpl) GetCompensationLogsByStep(
	ctx context.Context,
	sagaID string,
	stepNumber int32,
) ([]*CompensationLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. Get logs for saga
	logs, exists := r.logs[sagaID]
	if !exists {
		return make([]*CompensationLog, 0), nil
	}

	// 2. Filter by step number
	result := make([]*CompensationLog, 0)
	for _, log := range logs {
		if log.StepNumber == stepNumber {
			result = append(result, log)
		}
	}

	return result, nil
}

// UpdateCompensationLog updates an existing compensation log
func (r *CompensationLogRepositoryImpl) UpdateCompensationLog(
	ctx context.Context,
	sagaID string,
	log *CompensationLog,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 1. Validate inputs
	if sagaID == "" {
		return fmt.Errorf("saga ID cannot be empty")
	}
	if log == nil {
		return fmt.Errorf("log cannot be nil")
	}

	// 2. Find and update log
	logs, exists := r.logs[sagaID]
	if !exists {
		return fmt.Errorf("no logs found for saga %s", sagaID)
	}

	for i, existingLog := range logs {
		if existingLog.ID == log.ID {
			logs[i] = log
			return nil
		}
	}

	return fmt.Errorf("log %s not found", log.ID)
}

// DeleteCompensationLogs deletes all compensation logs for a saga
func (r *CompensationLogRepositoryImpl) DeleteCompensationLogs(
	ctx context.Context,
	sagaID string,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 1. Validate input
	if sagaID == "" {
		return fmt.Errorf("saga ID cannot be empty")
	}

	// 2. Delete logs
	delete(r.logs, sagaID)

	return nil
}

// GetCompensationStatistics retrieves compensation statistics for a saga
func (r *CompensationLogRepositoryImpl) GetCompensationStatistics(
	ctx context.Context,
	sagaID string,
) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. Get logs
	logs, exists := r.logs[sagaID]
	if !exists {
		return map[string]interface{}{
			"totalLogs":          0,
			"successfulLogs":     0,
			"failedLogs":         0,
			"averageTimeMs":      0.0,
			"totalCompensations": 0,
		}, nil
	}

	// 2. Calculate statistics
	successCount := 0
	failCount := 0
	totalTimeMs := int64(0)

	for _, log := range logs {
		if log.Status == models.StepStatusSuccess {
			successCount++
		} else if log.Status == models.StepStatusFailed {
			failCount++
		}
		totalTimeMs += log.ExecutionTimeMs
	}

	averageTimeMs := 0.0
	if len(logs) > 0 {
		averageTimeMs = float64(totalTimeMs) / float64(len(logs))
	}

	return map[string]interface{}{
		"totalLogs":          len(logs),
		"successfulLogs":     successCount,
		"failedLogs":         failCount,
		"averageTimeMs":      averageTimeMs,
		"totalCompensations": len(logs),
	}, nil
}

// ClearLogs clears all compensation logs (for testing)
func (r *CompensationLogRepositoryImpl) ClearLogs() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logs = make(map[string][]*CompensationLog)
}

// GetLogCount returns the total number of logs
func (r *CompensationLogRepositoryImpl) GetLogCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, logs := range r.logs {
		count += len(logs)
	}

	return count
}
