// Package orchestrator contains unit tests for saga orchestrator
package orchestrator

import (
	"context"
	"errors"
	"testing"
	"time"

	"p9e.in/chetana/packages/saga"
	"p9e.in/chetana/packages/saga/models"
)

// Mock implementations for testing

type mockStepExecutor struct {
	executionFunc func(ctx context.Context, sagaID string, stepNum int, stepDef *saga.StepDefinition) (*models.StepResult, error)
	statusFunc    func(ctx context.Context, sagaID string, stepNum int) (*models.StepExecution, error)
}

func (m *mockStepExecutor) ExecuteStep(ctx context.Context, sagaID string, stepNum int, stepDef *saga.StepDefinition) (*models.StepResult, error) {
	if m.executionFunc != nil {
		return m.executionFunc(ctx, sagaID, stepNum, stepDef)
	}
	return &models.StepResult{
		StepNumber:     int32(stepNum),
		Status:         models.StepStatusSuccess,
		ExecutionTimeMs: 100,
	}, nil
}

func (m *mockStepExecutor) GetStepStatus(ctx context.Context, sagaID string, stepNum int) (*models.StepExecution, error) {
	if m.statusFunc != nil {
		return m.statusFunc(ctx, sagaID, stepNum)
	}
	return &models.StepExecution{
		SagaID:     sagaID,
		StepNumber: int32(stepNum),
		Status:     models.StepStatusSuccess,
	}, nil
}

type mockTimeoutHandler struct {
	setupFunc  func(ctx context.Context, sagaID string, stepNum int, timeoutSeconds int32) error
	cancelFunc func(sagaID string, stepNum int) error
	checkFunc  func(sagaID string, stepNum int) (bool, error)
	retryFunc  func(sagaType string, stepNum int) (*saga.RetryConfiguration, error)
}

func (m *mockTimeoutHandler) SetupStepTimeout(ctx context.Context, sagaID string, stepNum int, timeoutSeconds int32) error {
	if m.setupFunc != nil {
		return m.setupFunc(ctx, sagaID, stepNum, timeoutSeconds)
	}
	return nil
}

func (m *mockTimeoutHandler) CancelStepTimeout(sagaID string, stepNum int) error {
	if m.cancelFunc != nil {
		return m.cancelFunc(sagaID, stepNum)
	}
	return nil
}

func (m *mockTimeoutHandler) CheckExpired(sagaID string, stepNum int) (bool, error) {
	if m.checkFunc != nil {
		return m.checkFunc(sagaID, stepNum)
	}
	return false, nil
}

func (m *mockTimeoutHandler) GetRetryConfig(sagaType string, stepNum int) (*saga.RetryConfiguration, error) {
	if m.retryFunc != nil {
		return m.retryFunc(sagaType, stepNum)
	}
	return &saga.RetryConfiguration{
		MaxRetries:        3,
		InitialBackoffMs:  100,
		MaxBackoffMs:      1000,
		BackoffMultiplier: 2.0,
		JitterFraction:    0.1,
	}, nil
}

type mockEventPublisher struct {
	publishFunc func(eventType string) error
}

func (m *mockEventPublisher) PublishStepStarted(ctx context.Context, execution *models.SagaExecution, stepNum int32) error {
	if m.publishFunc != nil {
		return m.publishFunc("step_started")
	}
	return nil
}

func (m *mockEventPublisher) PublishStepCompleted(ctx context.Context, execution *models.SagaExecution, stepNum int32, result *models.StepResult) error {
	if m.publishFunc != nil {
		return m.publishFunc("step_completed")
	}
	return nil
}

func (m *mockEventPublisher) PublishStepFailed(ctx context.Context, execution *models.SagaExecution, stepNum int32, err error) error {
	if m.publishFunc != nil {
		return m.publishFunc("step_failed")
	}
	return nil
}

func (m *mockEventPublisher) PublishStepRetrying(ctx context.Context, execution *models.SagaExecution, stepNum int32, err error) error {
	if m.publishFunc != nil {
		return m.publishFunc("step_retrying")
	}
	return nil
}

func (m *mockEventPublisher) PublishSagaCompleted(ctx context.Context, execution *models.SagaExecution) error {
	if m.publishFunc != nil {
		return m.publishFunc("saga_completed")
	}
	return nil
}

func (m *mockEventPublisher) PublishSagaFailed(ctx context.Context, execution *models.SagaExecution) error {
	if m.publishFunc != nil {
		return m.publishFunc("saga_failed")
	}
	return nil
}

func (m *mockEventPublisher) PublishCompensationStarted(ctx context.Context, execution *models.SagaExecution) error {
	if m.publishFunc != nil {
		return m.publishFunc("compensation_started")
	}
	return nil
}

func (m *mockEventPublisher) PublishCompensationCompleted(ctx context.Context, execution *models.SagaExecution) error {
	if m.publishFunc != nil {
		return m.publishFunc("compensation_completed")
	}
	return nil
}

type mockRepository struct {
	createFunc func(ctx context.Context, execution *models.SagaExecution) error
	getFunc    func(ctx context.Context, sagaID string) (*models.SagaExecution, error)
	updateFunc func(ctx context.Context, execution *models.SagaExecution) error
}

func (m *mockRepository) CreateExecution(ctx context.Context, execution *models.SagaExecution) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, execution)
	}
	return nil
}

func (m *mockRepository) GetExecution(ctx context.Context, sagaID string) (*models.SagaExecution, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, sagaID)
	}
	return &models.SagaExecution{ID: sagaID}, nil
}

func (m *mockRepository) UpdateExecution(ctx context.Context, execution *models.SagaExecution) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, execution)
	}
	return nil
}

type mockExecutionLogRepository struct {
	createFunc func(ctx context.Context, log *models.StepExecution) error
	getFunc    func(ctx context.Context, sagaID string) ([]*models.StepExecution, error)
}

func (m *mockExecutionLogRepository) CreateExecutionLog(ctx context.Context, log *models.StepExecution) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, log)
	}
	return nil
}

func (m *mockExecutionLogRepository) GetExecutionLog(ctx context.Context, sagaID string) ([]*models.StepExecution, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, sagaID)
	}
	return make([]*models.StepExecution, 0), nil
}

type mockSagaHandler struct {
	sagaTypeFunc        func() string
	stepDefsFunc        func() []*saga.StepDefinition
	stepDefFunc         func(stepNum int32) (*saga.StepDefinition, error)
	validateInputFunc   func(input *saga.SagaExecutionInput) error
}

func (m *mockSagaHandler) SagaType() string {
	if m.sagaTypeFunc != nil {
		return m.sagaTypeFunc()
	}
	return "TEST-SAGA"
}

func (m *mockSagaHandler) GetStepDefinitions() []*saga.StepDefinition {
	if m.stepDefsFunc != nil {
		return m.stepDefsFunc()
	}
	return []*saga.StepDefinition{
		{
			StepNumber:    1,
			ServiceName:   "test-service",
			HandlerMethod: "TestHandler",
			IsCritical:    true,
		},
	}
}

func (m *mockSagaHandler) GetStepDefinition(stepNum int32) (*saga.StepDefinition, error) {
	if m.stepDefFunc != nil {
		return m.stepDefFunc(stepNum)
	}
	return &saga.StepDefinition{
		StepNumber:    stepNum,
		ServiceName:   "test-service",
		HandlerMethod: "TestHandler",
	}, nil
}

func (m *mockSagaHandler) ValidateInput(input *saga.SagaExecutionInput) error {
	if m.validateInputFunc != nil {
		return m.validateInputFunc(input)
	}
	return nil
}

// Tests

func TestSagaOrchestratorExecuteSaga_Success(t *testing.T) {
	registry := NewSagaRegistry()
	config := &saga.DefaultConfig{
		DefaultTimeoutSeconds:  60,
		DefaultMaxRetries:      3,
		DefaultInitialBackoff:  time.Second,
		DefaultMaxBackoff:      30 * time.Second,
		BackoffMultiplier:      2.0,
		JitterFraction:         0.1,
		CircuitBreakerThreshold: 5,
		CircuitBreakerResetMs:  60000,
	}

	orchestrator := NewSagaOrchestratorImpl(
		registry,
		&mockStepExecutor{},
		&mockTimeoutHandler{},
		&mockEventPublisher{},
		&mockRepository{},
		&mockExecutionLogRepository{},
		config,
	)

	handler := &mockSagaHandler{
		stepDefsFunc: func() []*saga.StepDefinition {
			return []*saga.StepDefinition{
				{
					StepNumber:    1,
					ServiceName:   "test-service",
					HandlerMethod: "TestHandler",
					IsCritical:    true,
				},
			}
		},
	}

	registry.RegisterHandler("TEST-SAGA", handler)

	ctx := context.Background()
	input := &saga.SagaExecutionInput{
		TenantID:  "tenant-1",
		CompanyID: "company-1",
		BranchID:  "branch-1",
	}

	execution, err := orchestrator.ExecuteSaga(ctx, "TEST-SAGA", input)

	if err != nil {
		t.Errorf("ExecuteSaga failed: %v", err)
	}

	if execution == nil {
		t.Fatal("execution is nil")
	}

	if execution.Status != models.SagaStatusCompleted {
		t.Errorf("Expected status %s, got %s", models.SagaStatusCompleted, execution.Status)
	}

	if execution.TenantID != "tenant-1" {
		t.Errorf("Expected TenantID tenant-1, got %s", execution.TenantID)
	}
}

func TestSagaOrchestratorExecuteSaga_InvalidSagaType(t *testing.T) {
	registry := NewSagaRegistry()
	config := &saga.DefaultConfig{}

	orchestrator := NewSagaOrchestratorImpl(
		registry,
		&mockStepExecutor{},
		&mockTimeoutHandler{},
		&mockEventPublisher{},
		&mockRepository{},
		&mockExecutionLogRepository{},
		config,
	)

	ctx := context.Background()
	input := &saga.SagaExecutionInput{}

	_, err := orchestrator.ExecuteSaga(ctx, "INVALID-SAGA", input)

	if err == nil {
		t.Fatal("Expected error for invalid saga type")
	}
}

func TestSagaOrchestratorRegisterHandler_Success(t *testing.T) {
	registry := NewSagaRegistry()
	config := &saga.DefaultConfig{}

	orchestrator := NewSagaOrchestratorImpl(
		registry,
		&mockStepExecutor{},
		&mockTimeoutHandler{},
		&mockEventPublisher{},
		&mockRepository{},
		&mockExecutionLogRepository{},
		config,
	)

	handler := &mockSagaHandler{}
	err := orchestrator.RegisterSagaHandler("TEST-SAGA", handler)

	if err != nil {
		t.Errorf("RegisterSagaHandler failed: %v", err)
	}

	// Verify handler was registered
	if !registry.HasHandler("TEST-SAGA") {
		t.Fatal("Handler not registered")
	}
}

func TestSagaOrchestratorGetExecution_Success(t *testing.T) {
	registry := NewSagaRegistry()
	config := &saga.DefaultConfig{}

	expectedExecution := &models.SagaExecution{
		ID:     "saga-123",
		Status: models.SagaStatusCompleted,
	}

	mockRepo := &mockRepository{
		getFunc: func(ctx context.Context, sagaID string) (*models.SagaExecution, error) {
			return expectedExecution, nil
		},
	}

	orchestrator := NewSagaOrchestratorImpl(
		registry,
		&mockStepExecutor{},
		&mockTimeoutHandler{},
		&mockEventPublisher{},
		mockRepo,
		&mockExecutionLogRepository{},
		config,
	)

	ctx := context.Background()
	execution, err := orchestrator.GetExecution(ctx, "saga-123")

	if err != nil {
		t.Errorf("GetExecution failed: %v", err)
	}

	if execution.ID != "saga-123" {
		t.Errorf("Expected saga ID saga-123, got %s", execution.ID)
	}
}

func TestSagaOrchestratorGetExecutionTimeline_Success(t *testing.T) {
	registry := NewSagaRegistry()
	config := &saga.DefaultConfig{}

	expectedTimeline := []*models.StepExecution{
		{
			SagaID:     "saga-123",
			StepNumber: 1,
			Status:     models.StepStatusSuccess,
		},
	}

	mockLogRepo := &mockExecutionLogRepository{
		getFunc: func(ctx context.Context, sagaID string) ([]*models.StepExecution, error) {
			return expectedTimeline, nil
		},
	}

	orchestrator := NewSagaOrchestratorImpl(
		registry,
		&mockStepExecutor{},
		&mockTimeoutHandler{},
		&mockEventPublisher{},
		&mockRepository{},
		mockLogRepo,
		config,
	)

	ctx := context.Background()
	timeline, err := orchestrator.GetExecutionTimeline(ctx, "saga-123")

	if err != nil {
		t.Errorf("GetExecutionTimeline failed: %v", err)
	}

	if len(timeline) != 1 {
		t.Errorf("Expected 1 step in timeline, got %d", len(timeline))
	}

	if timeline[0].StepNumber != 1 {
		t.Errorf("Expected step 1, got %d", timeline[0].StepNumber)
	}
}

func TestSagaRegistryRegisterHandler_Duplicate(t *testing.T) {
	registry := NewSagaRegistry()
	handler := &mockSagaHandler{}

	err1 := registry.RegisterHandler("TEST-SAGA", handler)
	if err1 != nil {
		t.Fatalf("First registration failed: %v", err1)
	}

	err2 := registry.RegisterHandler("TEST-SAGA", handler)
	if err2 == nil {
		t.Fatal("Expected error when registering duplicate handler")
	}
}

func TestSagaRegistryGetHandler_NotFound(t *testing.T) {
	registry := NewSagaRegistry()

	_, err := registry.GetHandler("NONEXISTENT-SAGA")
	if err == nil {
		t.Fatal("Expected error for nonexistent handler")
	}
}

func TestSagaRegistryGetAllHandlers(t *testing.T) {
	registry := NewSagaRegistry()
	handler1 := &mockSagaHandler{}
	handler2 := &mockSagaHandler{}

	registry.RegisterHandler("SAGA-1", handler1)
	registry.RegisterHandler("SAGA-2", handler2)

	handlers := registry.GetAllHandlers()

	if len(handlers) != 2 {
		t.Errorf("Expected 2 handlers, got %d", len(handlers))
	}
}

func TestExecutionPlannerPlanExecution_Success(t *testing.T) {
	stepDefs := []*saga.StepDefinition{
		{
			StepNumber:    1,
			ServiceName:   "service-1",
			HandlerMethod: "Method1",
			IsCritical:    true,
		},
		{
			StepNumber:    2,
			ServiceName:   "service-2",
			HandlerMethod: "Method2",
			IsCritical:    true,
		},
	}

	planner := NewExecutionPlanner(stepDefs)
	plan, err := planner.PlanExecution()

	if err != nil {
		t.Errorf("PlanExecution failed: %v", err)
	}

	if len(plan) != 2 {
		t.Errorf("Expected 2 steps in plan, got %d", len(plan))
	}
}

func TestExecutionPlannerCanExecuteStep_Success(t *testing.T) {
	stepDefs := []*saga.StepDefinition{
		{
			StepNumber:    1,
			ServiceName:   "service-1",
			HandlerMethod: "Method1",
		},
		{
			StepNumber:    2,
			ServiceName:   "service-2",
			HandlerMethod: "Method2",
		},
	}

	planner := NewExecutionPlanner(stepDefs)
	executionState := map[string]interface{}{
		"step_1_result": "success",
	}

	canExecute, err := planner.CanExecuteStep(2, executionState)

	if err != nil {
		t.Errorf("CanExecuteStep failed: %v", err)
	}

	if !canExecute {
		t.Error("Expected step 2 to be executable")
	}
}

func TestExecutionPlannerGetCriticalPath(t *testing.T) {
	stepDefs := []*saga.StepDefinition{
		{
			StepNumber:    1,
			ServiceName:   "service-1",
			HandlerMethod: "Method1",
			IsCritical:    true,
		},
		{
			StepNumber:    2,
			ServiceName:   "service-2",
			HandlerMethod: "Method2",
			IsCritical:    false,
		},
		{
			StepNumber:    3,
			ServiceName:   "service-3",
			HandlerMethod: "Method3",
			IsCritical:    true,
		},
	}

	planner := NewExecutionPlanner(stepDefs)
	critical := planner.GetCriticalPath()

	if len(critical) != 2 {
		t.Errorf("Expected 2 critical steps, got %d", len(critical))
	}
}

func TestExecutionPlannerEstimateExecutionTime(t *testing.T) {
	stepDefs := []*saga.StepDefinition{
		{
			StepNumber:    1,
			ServiceName:   "service-1",
			HandlerMethod: "Method1",
			TimeoutSeconds: 10,
		},
		{
			StepNumber:    2,
			ServiceName:   "service-2",
			HandlerMethod: "Method2",
			TimeoutSeconds: 20,
		},
	}

	planner := NewExecutionPlanner(stepDefs)
	estimatedTime := planner.EstimateExecutionTime()

	if estimatedTime != 30 {
		t.Errorf("Expected estimated time 30s, got %ds", estimatedTime)
	}
}
