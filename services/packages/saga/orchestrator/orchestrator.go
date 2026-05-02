// Package orchestrator implements the SagaOrchestrator interface
package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"p9e.in/chetana/packages/saga"
	"p9e.in/chetana/packages/saga/models"
)

// SagaOrchestratorImpl orchestrates saga execution with step coordination
type SagaOrchestratorImpl struct {
	mu                sync.RWMutex
	registry          *SagaRegistry
	executor          saga.SagaStepExecutor
	timeoutHandler    saga.SagaTimeoutHandler
	eventPublisher    saga.SagaEventPublisher
	repository        saga.SagaRepository
	execLogRepository saga.SagaExecutionLogRepository
	config            *saga.DefaultConfig
}

// NewSagaOrchestratorImpl creates new saga orchestrator instance
func NewSagaOrchestratorImpl(
	registry *SagaRegistry,
	executor saga.SagaStepExecutor,
	timeoutHandler saga.SagaTimeoutHandler,
	eventPublisher saga.SagaEventPublisher,
	repository saga.SagaRepository,
	execLogRepository saga.SagaExecutionLogRepository,
	config *saga.DefaultConfig,
) *SagaOrchestratorImpl {
	return &SagaOrchestratorImpl{
		registry:          registry,
		executor:          executor,
		timeoutHandler:    timeoutHandler,
		eventPublisher:    eventPublisher,
		repository:        repository,
		execLogRepository: execLogRepository,
		config:            config,
	}
}

// ExecuteSaga starts a new saga execution from step 1
func (o *SagaOrchestratorImpl) ExecuteSaga(
	ctx context.Context,
	sagaType string,
	input *saga.SagaExecutionInput,
) (*models.SagaExecution, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Validate saga type is registered
	handler, err := o.registry.GetHandler(sagaType)
	if err != nil {
		return nil, fmt.Errorf("failed to get handler for saga type %s: %w", sagaType, err)
	}

	// Validate input
	if err := handler.ValidateInput(input); err != nil {
		return nil, fmt.Errorf("invalid saga input: %w", err)
	}

	// Get step definitions
	stepDefs := handler.GetStepDefinitions()
	if len(stepDefs) == 0 {
		return nil, fmt.Errorf("saga type %s has no step definitions", sagaType)
	}

	// Create saga execution record
	now := time.Now()
	execution := &models.SagaExecution{
		ID:             generateSagaID(),
		TenantID:       input.TenantID,
		CompanyID:      input.CompanyID,
		BranchID:       input.BranchID,
		SagaType:       sagaType,
		Status:         models.SagaStatusRunning,
		CurrentStep:    1,
		TotalSteps:     int32(len(stepDefs)),
		StartedAt:      &now,
		TimeoutSeconds: input.TimeoutSeconds,
		ExpiresAt:      calculateExpiryTime(now, input.TimeoutSeconds),
		SagaDefinition: marshalStepDefinitions(stepDefs),
		ExecutionState: make(map[string]interface{}),
		Metadata:       input.Metadata,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	// Persist saga execution
	if err := o.repository.CreateExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to create saga execution: %w", err)
	}

	// Publish saga started event
	if err := o.eventPublisher.PublishSagaStarted(ctx, execution); err != nil {
		// Log warning but don't fail - saga is already created
		fmt.Printf("failed to publish saga started event: %v\n", err)
	}

	// Execute steps sequentially
	if err := o.executeSteps(ctx, execution, stepDefs, handler); err != nil {
		// Update execution status to FAILED
		execution.Status = models.SagaStatusFailed
		execution.ErrorMessage = err.Error()
		execution.ErrorCode = "STEP_EXECUTION_ERROR"

		// Publish saga failed event
		if publishErr := o.eventPublisher.PublishSagaFailed(ctx, execution); publishErr != nil {
			fmt.Printf("failed to publish saga failed event: %v\n", publishErr)
		}

		// Persist updated execution
		if persistErr := o.repository.UpdateExecution(ctx, execution); persistErr != nil {
			fmt.Printf("failed to persist failed execution: %v\n", persistErr)
		}

		// Start compensation
		if compErr := o.startCompensation(ctx, execution, stepDefs); compErr != nil {
			return execution, fmt.Errorf("saga failed with step error and compensation failed: %w", compErr)
		}

		return execution, fmt.Errorf("saga execution failed: %w", err)
	}

	// Mark saga as completed
	now = time.Now()
	execution.Status = models.SagaStatusCompleted
	execution.CompletedAt = &now
	execution.UpdatedAt = &now

	if err := o.repository.UpdateExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to update saga execution to completed: %w", err)
	}

	// Publish saga completed event
	if err := o.eventPublisher.PublishSagaCompleted(ctx, execution); err != nil {
		fmt.Printf("failed to publish saga completed event: %v\n", err)
	}

	return execution, nil
}

// ResumeSaga resumes a failed saga from the last successful step
func (o *SagaOrchestratorImpl) ResumeSaga(
	ctx context.Context,
	sagaID string,
) (*models.SagaExecution, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Retrieve existing execution
	execution, err := o.repository.GetExecution(ctx, sagaID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve saga execution: %w", err)
	}

	// Only resume if failed or partially compensated
	if execution.Status != models.SagaStatusFailed && execution.Status != models.SagaStatusCompensated {
		return nil, fmt.Errorf("cannot resume saga in status %s", execution.Status)
	}

	// Get handler to retrieve step definitions
	handler, err := o.registry.GetHandler(execution.SagaType)
	if err != nil {
		return nil, fmt.Errorf("failed to get handler for saga type %s: %w", execution.SagaType, err)
	}

	stepDefs := handler.GetStepDefinitions()

	// Resume from current step
	execution.Status = models.SagaStatusRunning
	now := time.Now()
	execution.UpdatedAt = &now

	if err := o.repository.UpdateExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to update saga execution status: %w", err)
	}

	// Execute remaining steps
	if err := o.executeSteps(ctx, execution, stepDefs, handler); err != nil {
		execution.Status = models.SagaStatusFailed
		execution.ErrorMessage = err.Error()

		if persistErr := o.repository.UpdateExecution(ctx, execution); persistErr != nil {
			fmt.Printf("failed to persist updated execution: %v\n", persistErr)
		}

		return execution, fmt.Errorf("saga resume failed: %w", err)
	}

	// Mark as completed
	now = time.Now()
	execution.Status = models.SagaStatusCompleted
	execution.CompletedAt = &now
	execution.UpdatedAt = &now

	if err := o.repository.UpdateExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to update saga to completed: %w", err)
	}

	return execution, nil
}

// GetExecution retrieves current saga execution state
func (o *SagaOrchestratorImpl) GetExecution(
	ctx context.Context,
	sagaID string,
) (*models.SagaExecution, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return o.repository.GetExecution(ctx, sagaID)
}

// GetExecutionTimeline retrieves all steps executed for a saga
func (o *SagaOrchestratorImpl) GetExecutionTimeline(
	ctx context.Context,
	sagaID string,
) ([]*models.StepExecution, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return o.execLogRepository.GetExecutionLog(ctx, sagaID)
}

// RegisterSagaHandler registers a handler for a saga type
func (o *SagaOrchestratorImpl) RegisterSagaHandler(
	sagaType string,
	handler saga.SagaHandler,
) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.registry.RegisterHandler(sagaType, handler)
}

// executeSteps executes all remaining steps in sequence
func (o *SagaOrchestratorImpl) executeSteps(
	ctx context.Context,
	execution *models.SagaExecution,
	stepDefs []*saga.StepDefinition,
	handler saga.SagaHandler,
) error {
	// Execute each step starting from current
	for i := int(execution.CurrentStep) - 1; i < len(stepDefs); i++ {
		stepDef := stepDefs[i]
		stepNum := int32(i + 1)

		// Check if saga has expired
		if execution.ExpiresAt != nil && time.Now().After(*execution.ExpiresAt) {
			return saga.NewSagaError(
				execution.ID,
				stepNum,
				"TIMEOUT",
				"saga execution timeout",
				nil,
				true,
			)
		}

		// Publish step started event
		if err := o.eventPublisher.PublishStepStarted(ctx, execution, stepNum); err != nil {
			fmt.Printf("failed to publish step started event: %v\n", err)
		}

		// Setup timeout for this step
		if stepDef.TimeoutSeconds > 0 {
			if err := o.timeoutHandler.SetupStepTimeout(ctx, execution.ID, int(stepNum), stepDef.TimeoutSeconds); err != nil {
				fmt.Printf("failed to setup step timeout: %v\n", err)
			}
		}

		// Execute step with retry logic
		var stepResult *models.StepResult
		var stepErr error

		retryConfig, err := o.timeoutHandler.GetRetryConfig(execution.SagaType, int(stepNum))
		if err != nil {
			retryConfig = &saga.RetryConfiguration{
				MaxRetries:         o.config.DefaultMaxRetries,
				InitialBackoffMs:   int32(o.config.DefaultInitialBackoff.Milliseconds()),
				MaxBackoffMs:       int32(o.config.DefaultMaxBackoff.Milliseconds()),
				BackoffMultiplier:  o.config.BackoffMultiplier,
				JitterFraction:     o.config.JitterFraction,
			}
		}

		stepResult, stepErr = o.executeStepWithRetry(ctx, execution, stepDef, stepNum, retryConfig)

		// Cancel timeout
		if err := o.timeoutHandler.CancelStepTimeout(execution.ID, int(stepNum)); err != nil {
			fmt.Printf("failed to cancel step timeout: %v\n", err)
		}

		if stepErr != nil {
			// Log step error
			if err := o.execLogRepository.CreateExecutionLog(ctx, &models.StepExecution{
				SagaID:        execution.ID,
				StepNumber:    stepNum,
				Status:        models.StepStatusFailed,
				ErrorMessage:  stepErr.Error(),
				ExecutedAt:    time.Now(),
				ExecutionTime: 0,
			}); err != nil {
				fmt.Printf("failed to log step execution: %v\n", err)
			}

			// Publish step failed event
			if pubErr := o.eventPublisher.PublishStepFailed(ctx, execution, stepNum, stepErr); pubErr != nil {
				fmt.Printf("failed to publish step failed event: %v\n", pubErr)
			}

			// Check if step is critical
			if stepDef.IsCritical {
				return stepErr
			}

			// For non-critical steps, continue
			continue
		}

		// Log successful step execution
		if err := o.execLogRepository.CreateExecutionLog(ctx, &models.StepExecution{
			SagaID:        execution.ID,
			StepNumber:    stepNum,
			Status:        models.StepStatusSuccess,
			Result:        stepResult.Result,
			ExecutedAt:    time.Now(),
			ExecutionTime: stepResult.ExecutionTimeMs,
			RetryCount:    stepResult.RetryCount,
		}); err != nil {
			fmt.Printf("failed to log step execution: %v\n", err)
		}

		// Publish step completed event
		if err := o.eventPublisher.PublishStepCompleted(ctx, execution, stepNum, stepResult); err != nil {
			fmt.Printf("failed to publish step completed event: %v\n", err)
		}

		// Store step result in execution state
		execution.ExecutionState[fmt.Sprintf("step_%d_result", stepNum)] = stepResult.Result
		execution.CurrentStep = stepNum + 1

		// Update execution after each step
		now := time.Now()
		execution.UpdatedAt = &now
		if err := o.repository.UpdateExecution(ctx, execution); err != nil {
			fmt.Printf("failed to update execution after step: %v\n", err)
		}
	}

	return nil
}

// executeStepWithRetry executes a step with exponential backoff retry
func (o *SagaOrchestratorImpl) executeStepWithRetry(
	ctx context.Context,
	execution *models.SagaExecution,
	stepDef *saga.StepDefinition,
	stepNum int32,
	retryConfig *saga.RetryConfiguration,
) (*models.StepResult, error) {
	var lastErr error
	var retryCount int32 = 0

	for retryCount <= retryConfig.MaxRetries {
		stepResult, err := o.executor.ExecuteStep(ctx, execution.ID, int(stepNum), stepDef)

		if err == nil {
			stepResult.RetryCount = retryCount
			return stepResult, nil
		}

		lastErr = err
		retryCount++

		if retryCount > retryConfig.MaxRetries {
			break
		}

		// Publish retry event
		if pubErr := o.eventPublisher.PublishStepRetrying(ctx, execution, stepNum, err); pubErr != nil {
			fmt.Printf("failed to publish step retrying event: %v\n", pubErr)
		}

		// Calculate backoff with exponential increase and jitter
		backoffMs := calculateBackoff(
			retryCount-1,
			int64(retryConfig.InitialBackoffMs),
			int64(retryConfig.MaxBackoffMs),
			retryConfig.BackoffMultiplier,
			retryConfig.JitterFraction,
		)

		// Wait before retry
		select {
		case <-time.After(time.Duration(backoffMs) * time.Millisecond):
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled during retry backoff: %w", ctx.Err())
		}
	}

	return nil, fmt.Errorf("step execution failed after %d retries: %w", retryCount, lastErr)
}

// startCompensation initiates compensation for failed saga
func (o *SagaOrchestratorImpl) startCompensation(
	ctx context.Context,
	execution *models.SagaExecution,
	stepDefs []*saga.StepDefinition,
) error {
	execution.Status = models.SagaStatusCompensating
	execution.CompensationStatus = models.CompensationRunning
	now := time.Now()
	execution.UpdatedAt = &now

	if err := o.repository.UpdateExecution(ctx, execution); err != nil {
		return fmt.Errorf("failed to update execution status to compensating: %w", err)
	}

	// Publish compensation started event
	if err := o.eventPublisher.PublishCompensationStarted(ctx, execution); err != nil {
		fmt.Printf("failed to publish compensation started event: %v\n", err)
	}

	// Get timeline of executed steps
	timeline, err := o.execLogRepository.GetExecutionLog(ctx, execution.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve execution timeline: %w", err)
	}

	// Compensate in reverse order
	for i := len(timeline) - 1; i >= 0; i-- {
		stepExec := timeline[i]

		if stepExec.Status != models.StepStatusSuccess {
			continue
		}

		stepNum := stepExec.StepNumber
		stepDef := stepDefs[stepNum-1]

		// Execute compensation steps if defined
		if len(stepDef.CompensationSteps) == 0 {
			continue
		}

		// Compensation steps are executed in reverse order for saga rollback.
		// This will be implemented in Phase 1
	}

	execution.Status = models.SagaStatusCompensated
	execution.CompensationStatus = models.CompensationCompleted
	now = time.Now()
	execution.UpdatedAt = &now

	if err := o.repository.UpdateExecution(ctx, execution); err != nil {
		return fmt.Errorf("failed to update execution status to compensated: %w", err)
	}

	// Publish compensation completed event
	if err := o.eventPublisher.PublishCompensationCompleted(ctx, execution); err != nil {
		fmt.Printf("failed to publish compensation completed event: %v\n", err)
	}

	return nil
}

// Helper functions

func generateSagaID() string {
	// In production, use ULID generation
	return fmt.Sprintf("SAGA-%d", time.Now().UnixNano())
}

func calculateExpiryTime(startTime time.Time, timeoutSeconds int32) *time.Time {
	if timeoutSeconds <= 0 {
		return nil
	}
	expiry := startTime.Add(time.Duration(timeoutSeconds) * time.Second)
	return &expiry
}

func marshalStepDefinitions(stepDefs []*saga.StepDefinition) []byte {
	data, _ := json.Marshal(stepDefs)
	return data
}

func calculateBackoff(
	retryNum int32,
	initialMs int64,
	maxMs int64,
	multiplier float64,
	jitterFraction float64,
) int64 {
	// Exponential backoff: min(initial * multiplier^retry, max)
	backoff := int64(float64(initialMs) * exponentialPower(multiplier, float64(retryNum)))
	if backoff > maxMs {
		backoff = maxMs
	}

	// Add jitter: ±(jitterFraction * backoff)
	jitterMax := int64(float64(backoff) * jitterFraction)
	jitter := (time.Now().UnixNano() % (2*jitterMax + 1)) - jitterMax

	return backoff + jitter
}

func exponentialPower(base float64, exponent float64) float64 {
	result := 1.0
	for i := 0.0; i < exponent; i++ {
		result *= base
	}
	return result
}
