// Package compensation implements saga compensation logic
package compensation

import (
	"context"
	"fmt"
	"time"

	"p9e.in/samavaya/packages/saga"
	"p9e.in/samavaya/packages/saga/models"
)

// CompensationEngineImpl implements the SagaCompensationEngine interface
type CompensationEngineImpl struct {
	stepExecutor   saga.SagaStepExecutor
	eventPublisher saga.SagaEventPublisher
	repository     saga.SagaRepository
	logRepository  saga.SagaExecutionLogRepository
}

// NewCompensationEngineImpl creates a new compensation engine instance
func NewCompensationEngineImpl(
	stepExecutor saga.SagaStepExecutor,
	eventPublisher saga.SagaEventPublisher,
	repository saga.SagaRepository,
	logRepository saga.SagaExecutionLogRepository,
) *CompensationEngineImpl {
	return &CompensationEngineImpl{
		stepExecutor:   stepExecutor,
		eventPublisher: eventPublisher,
		repository:     repository,
		logRepository:  logRepository,
	}
}

// StartCompensation initiates compensation for a failed saga
func (ce *CompensationEngineImpl) StartCompensation(
	ctx context.Context,
	execution *models.SagaExecution,
	stepDefs []*saga.StepDefinition,
) error {
	// 1. Update saga status to COMPENSATING
	execution.Status = models.SagaStatusCompensating
	execution.CompensationStatus = models.CompensationRunning
	now := time.Now()
	execution.UpdatedAt = &now

	if err := ce.repository.UpdateExecution(ctx, execution); err != nil {
		return fmt.Errorf("failed to update execution status to compensating: %w", err)
	}

	// 2. Publish compensation started event
	if err := ce.eventPublisher.PublishCompensationStarted(ctx, execution); err != nil {
		fmt.Printf("failed to publish compensation started event: %v\n", err)
	}

	// 3. Execute compensation steps
	if err := ce.executeCompensation(ctx, execution, stepDefs); err != nil {
		// Mark as compensation failed
		execution.CompensationStatus = models.CompensationFailed
		now := time.Now()
		execution.UpdatedAt = &now

		if persistErr := ce.repository.UpdateExecution(ctx, execution); persistErr != nil {
			fmt.Printf("failed to persist compensation failure: %v\n", persistErr)
		}

		return fmt.Errorf("compensation execution failed: %w", err)
	}

	// 4. Mark compensation as completed
	execution.CompensationStatus = models.CompensationCompleted
	execution.Status = models.SagaStatusCompensated
	now = time.Now()
	execution.UpdatedAt = &now

	if err := ce.repository.UpdateExecution(ctx, execution); err != nil {
		return fmt.Errorf("failed to update execution to compensated: %w", err)
	}

	// 5. Publish compensation completed event
	if err := ce.eventPublisher.PublishCompensationCompleted(ctx, execution); err != nil {
		fmt.Printf("failed to publish compensation completed event: %v\n", err)
	}

	return nil
}

// ExecuteCompensation executes compensation for a specific step
func (ce *CompensationEngineImpl) ExecuteCompensation(
	ctx context.Context,
	execution *models.SagaExecution,
	stepNum int32,
	compensationSteps []*saga.StepDefinition,
) error {
	// 1. Validate step number
	if stepNum < 1 {
		return fmt.Errorf("invalid step number: %d", stepNum)
	}

	if len(compensationSteps) == 0 {
		// No compensation steps defined, which is fine
		return nil
	}

	// 2. Execute compensation steps in order
	for _, compStep := range compensationSteps {
		// 3. Execute compensation step
		result, err := ce.stepExecutor.ExecuteStep(ctx, execution.ID, int(compStep.StepNumber), compStep)
		if err != nil {
			// Log compensation error
			if logErr := ce.logRepository.CreateExecutionLog(ctx, &models.StepExecution{
				SagaID:       execution.ID,
				StepNumber:   compStep.StepNumber,
				Status:       models.StepStatusFailed,
				ErrorMessage: err.Error(),
				ExecutedAt:   time.Now(),
			}); logErr != nil {
				fmt.Printf("failed to log compensation error: %v\n", logErr)
			}

			// Check if step is critical
			if compStep.IsCritical {
				return fmt.Errorf("critical compensation step %d failed: %w", compStep.StepNumber, err)
			}

			// For non-critical steps, continue
			continue
		}

		// 4. Log successful compensation
		if err := ce.logRepository.CreateExecutionLog(ctx, &models.StepExecution{
			SagaID:        execution.ID,
			StepNumber:    compStep.StepNumber,
			Status:        models.StepStatusSuccess,
			Result:        result.Result,
			ExecutedAt:    time.Now(),
			ExecutionTime: result.ExecutionTimeMs,
		}); err != nil {
			fmt.Printf("failed to log compensation success: %v\n", err)
		}
	}

	return nil
}

// GetCompensationStatus retrieves the compensation status of a saga
func (ce *CompensationEngineImpl) GetCompensationStatus(
	ctx context.Context,
	sagaID string,
) (models.CompensationStatus, error) {
	// 1. Get execution
	execution, err := ce.repository.GetExecution(ctx, sagaID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve execution: %w", err)
	}

	// 2. Return compensation status
	return execution.CompensationStatus, nil
}

// executeCompensation executes compensation in reverse order
func (ce *CompensationEngineImpl) executeCompensation(
	ctx context.Context,
	execution *models.SagaExecution,
	stepDefs []*saga.StepDefinition,
) error {
	// 1. Get execution timeline (steps that were executed)
	timeline, err := ce.logRepository.GetExecutionLog(ctx, execution.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve execution timeline: %w", err)
	}

	// 2. Reverse the timeline (compensate in reverse order)
	for i := len(timeline) - 1; i >= 0; i-- {
		stepExec := timeline[i]

		// 3. Only compensate successful steps
		if stepExec.Status != models.StepStatusSuccess {
			continue
		}

		// 4. Get the step definition for this executed step
		if int(stepExec.StepNumber) > len(stepDefs) {
			continue
		}

		stepDef := stepDefs[stepExec.StepNumber-1]

		// 5. Execute compensation steps if defined. CompensationSteps is
		// []int32 (step-number references); resolve those to the actual
		// *StepDefinition entries before calling ExecuteCompensation.
		if len(stepDef.CompensationSteps) > 0 {
			resolved := make([]*saga.StepDefinition, 0, len(stepDef.CompensationSteps))
			for _, compStepNum := range stepDef.CompensationSteps {
				if int(compStepNum) < 1 || int(compStepNum) > len(stepDefs) {
					continue
				}
				resolved = append(resolved, stepDefs[compStepNum-1])
			}
			if err := ce.ExecuteCompensation(ctx, execution, stepExec.StepNumber, resolved); err != nil {
				// Continue with other steps, but track that compensation failed
				fmt.Printf("compensation for step %d failed: %v\n", stepExec.StepNumber, err)
			}
		}
	}

	return nil
}

// CanCompensate checks if a saga can be compensated
func (ce *CompensationEngineImpl) CanCompensate(
	ctx context.Context,
	execution *models.SagaExecution,
	stepDefs []*saga.StepDefinition,
) (bool, error) {
	// 1. Check execution status
	switch execution.Status {
	case models.SagaStatusFailed, models.SagaStatusCompensating, models.SagaStatusCompensated:
		// Can be compensated
	default:
		return false, fmt.Errorf("cannot compensate saga in status %s", execution.Status)
	}

	// 2. Check if any steps have uncompensatable operations
	timeline, err := ce.logRepository.GetExecutionLog(ctx, execution.ID)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve timeline: %w", err)
	}

	for _, stepExec := range timeline {
		if stepExec.Status != models.StepStatusSuccess {
			continue
		}

		// Check if step has compensation
		if int(stepExec.StepNumber) > len(stepDefs) {
			return false, fmt.Errorf("invalid step number in timeline: %d", stepExec.StepNumber)
		}

		stepDef := stepDefs[stepExec.StepNumber-1]
		if len(stepDef.CompensationSteps) == 0 && stepDef.IsCritical {
			// Critical step with no compensation = uncompensatable
			return false, fmt.Errorf("no compensation defined for critical step %d", stepExec.StepNumber)
		}
	}

	return true, nil
}

// RetryCompensation retries a failed compensation
func (ce *CompensationEngineImpl) RetryCompensation(
	ctx context.Context,
	execution *models.SagaExecution,
	stepDefs []*saga.StepDefinition,
) error {
	// 1. Verify compensation was previously attempted
	if execution.CompensationStatus == models.CompensationNotStarted {
		return fmt.Errorf("compensation has not been started")
	}

	// 2. Reset compensation status
	execution.CompensationStatus = models.CompensationRunning
	now := time.Now()
	execution.UpdatedAt = &now

	if err := ce.repository.UpdateExecution(ctx, execution); err != nil {
		return fmt.Errorf("failed to update execution: %w", err)
	}

	// 3. Execute compensation again
	if err := ce.executeCompensation(ctx, execution, stepDefs); err != nil {
		execution.CompensationStatus = models.CompensationFailed
		now := time.Now()
		execution.UpdatedAt = &now

		if persistErr := ce.repository.UpdateExecution(ctx, execution); persistErr != nil {
			fmt.Printf("failed to persist retry failure: %v\n", persistErr)
		}

		return fmt.Errorf("compensation retry failed: %w", err)
	}

	// 4. Mark as completed
	execution.CompensationStatus = models.CompensationCompleted
	now = time.Now()
	execution.UpdatedAt = &now

	if err := ce.repository.UpdateExecution(ctx, execution); err != nil {
		return fmt.Errorf("failed to update execution to completed: %w", err)
	}

	// 5. Publish event
	if err := ce.eventPublisher.PublishCompensationCompleted(ctx, execution); err != nil {
		fmt.Printf("failed to publish compensation completed event: %v\n", err)
	}

	return nil
}

// GetCompensationLog retrieves the compensation log for a saga
func (ce *CompensationEngineImpl) GetCompensationLog(
	ctx context.Context,
	sagaID string,
) ([]*models.StepExecution, error) {
	// Get execution log and filter for compensation steps
	timeline, err := ce.logRepository.GetExecutionLog(ctx, sagaID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve execution log: %w", err)
	}

	// Return full timeline (includes both forward and compensation steps)
	return timeline, nil
}
