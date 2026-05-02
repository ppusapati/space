// Package executor contains unit tests for saga step execution
package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	"p9e.in/chetana/packages/saga"
	"p9e.in/chetana/packages/saga/models"
)

// Mock RPC Connector for testing

type mockRpcConnector struct {
	endpoints           map[string]string
	invokeFunc          func(ctx context.Context, endpoint, method string, req interface{}) (interface{}, error)
	getEndpointFunc     func(serviceName string) (string, error)
	registerServiceFunc func(serviceName, endpoint string) error
}

func (m *mockRpcConnector) InvokeHandler(
	ctx context.Context,
	endpoint string,
	handlerMethod string,
	request interface{},
) (interface{}, error) {
	if m.invokeFunc != nil {
		return m.invokeFunc(ctx, endpoint, handlerMethod, request)
	}
	return map[string]interface{}{"status": "success"}, nil
}

func (m *mockRpcConnector) GetServiceEndpoint(serviceName string) (string, error) {
	if m.getEndpointFunc != nil {
		return m.getEndpointFunc(serviceName)
	}
	if endpoint, exists := m.endpoints[serviceName]; exists {
		return endpoint, nil
	}
	return "", errors.New("service not found")
}

func (m *mockRpcConnector) RegisterService(serviceName string, endpoint string) error {
	if m.registerServiceFunc != nil {
		return m.registerServiceFunc(serviceName, endpoint)
	}
	if m.endpoints == nil {
		m.endpoints = make(map[string]string)
	}
	m.endpoints[serviceName] = endpoint
	return nil
}

// Tests

func TestStepExecutorExecuteStep_Success(t *testing.T) {
	idempotency := NewIdempotencyImpl(1*time.Minute, 1000)
	mockConnector := &mockRpcConnector{
		endpoints: map[string]string{
			"order-service": "http://order-service:8080",
		},
	}

	executor := NewStepExecutorImpl(mockConnector, idempotency)

	ctx := context.Background()
	stepDef := &saga.StepDefinition{
		StepNumber:    1,
		ServiceName:   "order-service",
		HandlerMethod: "CreateOrder",
		IsCritical:    true,
	}

	result, err := executor.ExecuteStep(ctx, "saga-123", 1, stepDef)

	if err != nil {
		t.Errorf("ExecuteStep failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if result.Status != models.StepStatusSuccess {
		t.Errorf("Expected status %s, got %s", models.StepStatusSuccess, result.Status)
	}

	if result.StepNumber != 1 {
		t.Errorf("Expected step number 1, got %d", result.StepNumber)
	}
}

func TestStepExecutorExecuteStep_ServiceNotFound(t *testing.T) {
	idempotency := NewIdempotencyImpl(1*time.Minute, 1000)
	mockConnector := &mockRpcConnector{
		getEndpointFunc: func(serviceName string) (string, error) {
			return "", errors.New("service not found")
		},
	}

	executor := NewStepExecutorImpl(mockConnector, idempotency)

	ctx := context.Background()
	stepDef := &saga.StepDefinition{
		StepNumber:    1,
		ServiceName:   "unknown-service",
		HandlerMethod: "TestMethod",
	}

	_, err := executor.ExecuteStep(ctx, "saga-123", 1, stepDef)

	if err == nil {
		t.Fatal("Expected error for unknown service")
	}
}

func TestStepExecutorExecuteStep_RpcInvocationError(t *testing.T) {
	idempotency := NewIdempotencyImpl(1*time.Minute, 1000)
	mockConnector := &mockRpcConnector{
		endpoints: map[string]string{
			"order-service": "http://order-service:8080",
		},
		invokeFunc: func(ctx context.Context, endpoint, method string, req interface{}) (interface{}, error) {
			return nil, errors.New("RPC call failed")
		},
	}

	executor := NewStepExecutorImpl(mockConnector, idempotency)

	ctx := context.Background()
	stepDef := &saga.StepDefinition{
		StepNumber:    1,
		ServiceName:   "order-service",
		HandlerMethod: "CreateOrder",
	}

	_, err := executor.ExecuteStep(ctx, "saga-123", 1, stepDef)

	if err == nil {
		t.Fatal("Expected error for RPC invocation failure")
	}
}

func TestStepExecutorExecuteStep_Idempotent(t *testing.T) {
	idempotency := NewIdempotencyImpl(1*time.Minute, 1000)
	mockConnector := &mockRpcConnector{
		endpoints: map[string]string{
			"order-service": "http://order-service:8080",
		},
	}

	executor := NewStepExecutorImpl(mockConnector, idempotency)

	ctx := context.Background()
	stepDef := &saga.StepDefinition{
		StepNumber:    1,
		ServiceName:   "order-service",
		HandlerMethod: "CreateOrder",
	}

	// First execution
	result1, err1 := executor.ExecuteStep(ctx, "saga-123", 1, stepDef)
	if err1 != nil {
		t.Errorf("First execution failed: %v", err1)
	}

	// Second execution (should be cached)
	result2, err2 := executor.ExecuteStep(ctx, "saga-123", 1, stepDef)
	if err2 != nil {
		t.Errorf("Second execution failed: %v", err2)
	}

	// Results should be identical (same cached result)
	if result1.StepNumber != result2.StepNumber {
		t.Error("Step numbers don't match")
	}

	if result1.Result != result2.Result {
		t.Error("Results don't match for idempotent execution")
	}
}

func TestRpcConnectorRegisterService_Success(t *testing.T) {
	connector := NewRpcConnectorImpl()

	err := connector.RegisterService("order-service", "http://order-service:8080")

	if err != nil {
		t.Errorf("RegisterService failed: %v", err)
	}

	// Verify registration
	endpoint, err := connector.GetServiceEndpoint("order-service")
	if err != nil {
		t.Errorf("GetServiceEndpoint failed: %v", err)
	}

	if endpoint != "http://order-service:8080" {
		t.Errorf("Expected endpoint http://order-service:8080, got %s", endpoint)
	}
}

func TestRpcConnectorGetServiceEndpoint_NotFound(t *testing.T) {
	connector := NewRpcConnectorImpl()

	_, err := connector.GetServiceEndpoint("unknown-service")

	if err == nil {
		t.Fatal("Expected error for unknown service")
	}
}

func TestRpcConnectorInvokeHandler_Success(t *testing.T) {
	connector := NewRpcConnectorImpl()

	err := connector.RegisterService("order-service", "http://order-service:8080")
	if err != nil {
		t.Fatalf("RegisterService failed: %v", err)
	}

	ctx := context.Background()
	response, err := connector.InvokeHandler(
		ctx,
		"http://order-service:8080",
		"CreateOrder",
		map[string]interface{}{"orderId": "123"},
	)

	if err != nil {
		t.Errorf("InvokeHandler failed: %v", err)
	}

	if response == nil {
		t.Fatal("response is nil")
	}
}

func TestRpcConnectorRegisterService_EmptyServiceName(t *testing.T) {
	connector := NewRpcConnectorImpl()

	err := connector.RegisterService("", "http://endpoint:8080")

	if err == nil {
		t.Fatal("Expected error for empty service name")
	}
}

func TestRpcConnectorRegisterService_EmptyEndpoint(t *testing.T) {
	connector := NewRpcConnectorImpl()

	err := connector.RegisterService("order-service", "")

	if err == nil {
		t.Fatal("Expected error for empty endpoint")
	}
}

func TestIdempotencyGetCachedResult_Hit(t *testing.T) {
	idempotency := NewIdempotencyImpl(1*time.Minute, 1000)

	// Cache a result
	result := &models.StepResult{
		StepNumber: 1,
		Status:     models.StepStatusSuccess,
		Result:     "success",
	}

	err := idempotency.CacheResult("saga-123", 1, result)
	if err != nil {
		t.Fatalf("CacheResult failed: %v", err)
	}

	// Retrieve cached result
	cachedResult := idempotency.GetCachedResult("saga-123", 1)

	if cachedResult == nil {
		t.Fatal("Expected cached result, got nil")
	}

	if cachedResult.StepNumber != 1 {
		t.Errorf("Expected step number 1, got %d", cachedResult.StepNumber)
	}
}

func TestIdempotencyGetCachedResult_Miss(t *testing.T) {
	idempotency := NewIdempotencyImpl(1*time.Minute, 1000)

	cachedResult := idempotency.GetCachedResult("saga-123", 1)

	if cachedResult != nil {
		t.Fatal("Expected nil for uncached result")
	}
}

func TestIdempotencyGetCachedResult_Expired(t *testing.T) {
	idempotency := NewIdempotencyImpl(10*time.Millisecond, 1000) // Very short TTL

	result := &models.StepResult{
		StepNumber: 1,
		Status:     models.StepStatusSuccess,
		Result:     "success",
	}

	err := idempotency.CacheResult("saga-123", 1, result)
	if err != nil {
		t.Fatalf("CacheResult failed: %v", err)
	}

	// Wait for expiration
	time.Sleep(50 * time.Millisecond)

	cachedResult := idempotency.GetCachedResult("saga-123", 1)

	if cachedResult != nil {
		t.Fatal("Expected nil for expired result")
	}
}

func TestIdempotencyIsDuplicate_True(t *testing.T) {
	idempotency := NewIdempotencyImpl(1*time.Minute, 1000)

	result := &models.StepResult{
		StepNumber: 1,
		Status:     models.StepStatusSuccess,
		Result:     "success",
	}

	idempotency.CacheResult("saga-123", 1, result)

	isDuplicate := idempotency.IsDuplicate("saga-123", 1)

	if !isDuplicate {
		t.Fatal("Expected true for duplicate")
	}
}

func TestIdempotencyIsDuplicate_False(t *testing.T) {
	idempotency := NewIdempotencyImpl(1*time.Minute, 1000)

	isDuplicate := idempotency.IsDuplicate("saga-123", 1)

	if isDuplicate {
		t.Fatal("Expected false for non-duplicate")
	}
}

func TestIdempotencyGetCacheStats(t *testing.T) {
	idempotency := NewIdempotencyImpl(1*time.Minute, 1000)

	result := &models.StepResult{
		StepNumber: 1,
		Status:     models.StepStatusSuccess,
		Result:     "success",
	}

	idempotency.CacheResult("saga-123", 1, result)
	idempotency.CacheResult("saga-123", 2, result)

	stats := idempotency.GetCacheStats()

	if stats["totalCached"].(int) != 2 {
		t.Errorf("Expected 2 cached items, got %d", stats["totalCached"])
	}
}

func TestIdempotencyClearCache(t *testing.T) {
	idempotency := NewIdempotencyImpl(1*time.Minute, 1000)

	result := &models.StepResult{
		StepNumber: 1,
		Status:     models.StepStatusSuccess,
		Result:     "success",
	}

	idempotency.CacheResult("saga-123", 1, result)

	idempotency.ClearCache()

	cachedResult := idempotency.GetCachedResult("saga-123", 1)

	if cachedResult != nil {
		t.Fatal("Expected nil after clearing cache")
	}
}
