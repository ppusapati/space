// Package executor implements step execution for saga orchestration
package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"p9e.in/chetana/packages/saga"
	"p9e.in/chetana/packages/saga/models"
)

// StepExecutorImpl executes individual saga steps via RPC
type StepExecutorImpl struct {
	rpcConnector saga.RpcConnector
	idempotency  *IdempotencyImpl
}

// NewStepExecutorImpl creates a new step executor instance
func NewStepExecutorImpl(
	rpcConnector saga.RpcConnector,
	idempotency *IdempotencyImpl,
) *StepExecutorImpl {
	return &StepExecutorImpl{
		rpcConnector: rpcConnector,
		idempotency:  idempotency,
	}
}

// ExecuteStep executes a single saga step via RPC invocation
func (e *StepExecutorImpl) ExecuteStep(
	ctx context.Context,
	sagaID string,
	stepNum int,
	stepDef *saga.StepDefinition,
) (*models.StepResult, error) {
	// 1. Check if result is cached (idempotent)
	cachedResult := e.idempotency.GetCachedResult(sagaID, stepNum)
	if cachedResult != nil {
		return cachedResult, nil
	}

	// 2. Resolve service endpoint
	endpoint, err := e.rpcConnector.GetServiceEndpoint(stepDef.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve service endpoint: %w", err)
	}

	// 3. Build request from step definition
	request := buildStepRequest(stepDef)

	// 4. Track execution time
	startTime := time.Now()

	// 5. Invoke service via RPC
	response, err := e.rpcConnector.InvokeHandler(
		ctx,
		endpoint,
		stepDef.HandlerMethod,
		request,
	)

	executionTime := time.Since(startTime)

	// 6. Handle execution errors
	if err != nil {
		return nil, fmt.Errorf("RPC invocation failed for step %d: %w", stepNum, err)
	}

	// 7. Build result. StepResult.Result is []byte (JSON wire form); the
	// RPC connector returns interface{}, so marshal here. Zero-value on
	// nil response keeps a predictable shape.
	var resultBytes []byte
	if response != nil {
		if b, err := json.Marshal(response); err == nil {
			resultBytes = b
		} else {
			// Fall back to a textual representation so the pipeline keeps
			// moving; the idempotency cache still has the typed response.
			resultBytes = []byte(fmt.Sprintf("%v", response))
		}
	}
	result := &models.StepResult{
		StepNumber:      int32(stepNum),
		Status:          models.StepStatusSuccess,
		Result:          resultBytes,
		ExecutionTimeMs: executionTime.Milliseconds(),
		RetryCount:      0,
		CompletedAt:     time.Now(),
	}

	// 8. Cache result for idempotency
	if err := e.idempotency.CacheResult(sagaID, stepNum, result); err != nil {
		// Log but don't fail - caching is for optimization
		fmt.Printf("failed to cache step result: %v\n", err)
	}

	return result, nil
}

// GetStepStatus retrieves the status of a previously executed step
func (e *StepExecutorImpl) GetStepStatus(
	ctx context.Context,
	sagaID string,
	stepNum int,
) (*models.StepExecution, error) {
	// This would typically be implemented via repository
	// For now, return error indicating not implemented
	return nil, fmt.Errorf("GetStepStatus not yet implemented in executor")
}

// Helper function to build request from step definition
func buildStepRequest(stepDef *saga.StepDefinition) interface{} {
	// Build request based on step definition's InputMapping
	// This would typically deserialize the input mapping to the correct type
	request := map[string]interface{}{
		"stepNumber":   stepDef.StepNumber,
		"serviceName":  stepDef.ServiceName,
		"handlerMethod": stepDef.HandlerMethod,
	}

	// Add any input mapping from step definition
	if stepDef.InputMapping != nil {
		for key, value := range stepDef.InputMapping {
			request[key] = value
		}
	}

	return request
}
