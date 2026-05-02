// Package compensation contains unit tests for compensation engine
package compensation

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
	executeFunc func(ctx context.Context, sagaID string, stepNum int, stepDef *saga.StepDefinition) (*models.StepResult, error)
}

func (m *mockStepExecutor) ExecuteStep(
	ctx context.Context,
	sagaID string,
	stepNum int,
	stepDef *saga.StepDefinition,
) (*models.StepResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, sagaID, stepNum, stepDef)
	}
	return &models.StepResult{
		StepNumber:      int32(stepNum),
		Status:          models.StepStatusSuccess,
		ExecutionTimeMs: 100,
	}, nil
}

func (m *mockStepExecutor) GetStepStatus(
	ctx context.Context,
	sagaID string,
	stepNum int,
) (*models.StepExecution, error) {
	return &models.StepExecution{
		SagaID:     sagaID,
		StepNumber: int32(stepNum),
		Status:     models.StepStatusSuccess,
	}, nil
}

type mockEventPublisher struct {
	publishFunc func(eventType string) error
}

func (m *mockEventPublisher) PublishStepStarted(ctx context.Context, execution *models.SagaExecution, stepNum int32) error {
	return nil
}

func (m *mockEventPublisher) PublishStepCompleted(ctx context.Context, execution *models.SagaExecution, stepNum int32, result *models.StepResult) error {
	return nil
}

func (m *mockEventPublisher) PublishStepFailed(ctx context.Context, execution *models.SagaExecution, stepNum int32, err error) error {
	return nil
}

func (m *mockEventPublisher) PublishStepRetrying(ctx context.Context, execution *models.SagaExecution, stepNum int32, err error) error {
	return nil
}

func (m *mockEventPublisher) PublishSagaCompleted(ctx context.Context, execution *models.SagaExecution) error {
	return nil
}

func (m *mockEventPublisher) PublishSagaFailed(ctx context.Context, execution *models.SagaExecution) error {
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
	getFunc    func(ctx context.Context, sagaID string) (*models.SagaExecution, error)
	updateFunc func(ctx context.Context, execution *models.SagaExecution) error
}

func (m *mockRepository) CreateExecution(ctx context.Context, execution *models.SagaExecution) error {
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

type mockLogRepository struct {
	getFunc    func(ctx context.Context, sagaID string) ([]*models.StepExecution, error)
	createFunc func(ctx context.Context, log *models.StepExecution) error
}

func (m *mockLogRepository) CreateExecutionLog(ctx context.Context, log *models.StepExecution) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, log)
	}
	return nil
}

func (m *mockLogRepository) GetExecutionLog(ctx context.Context, sagaID string) ([]*models.StepExecution, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, sagaID)
	}
	return make([]*models.StepExecution, 0), nil
}

// Tests

func TestCompensationEngineStartCompensation_Success(t *testing.T) {
	mockExecutor := &mockStepExecutor{}
	mockPublisher := &mockEventPublisher{}
	mockRepo := &mockRepository{}
	mockLogRepo := &mockLogRepository{
		getFunc: func(ctx context.Context, sagaID string) ([]*models.StepExecution, error) {
			return []*models.StepExecution{
				{
					SagaID:     sagaID,
					StepNumber: 1,
					Status:     models.StepStatusSuccess,
				},
			}, nil
		},
	}

	engine := NewCompensationEngineImpl(mockExecutor, mockPublisher, mockRepo, mockLogRepo)

	ctx := context.Background()
	now := time.Now()
	execution := &models.SagaExecution{
		ID:        "saga-123",
		SagaType:  "SAGA-S01",
		Status:    models.SagaStatusFailed,
		StartedAt: &now,
	}

	stepDefs := []*saga.StepDefinition{
		{
			StepNumber:    1,
			ServiceName:   "service-1",
			HandlerMethod: "Method1",
			CompensationSteps: []*saga.StepDefinition{
				{
					StepNumber:    1,
					HandlerMethod: "CancelMethod1",
				},
			},
		},
	}

	err := engine.StartCompensation(ctx, execution, stepDefs)

	if err != nil {
		t.Errorf("StartCompensation failed: %v", err)
	}

	if execution.Status != models.SagaStatusCompensated {
		t.Errorf("Expected compensated status, got %s", execution.Status)
	}

	if execution.CompensationStatus != models.CompensationCompleted {
		t.Errorf("Expected completed compensation status, got %s", execution.CompensationStatus)
	}
}

func TestCompensationEngineExecuteCompensation_Success(t *testing.T) {
	mockExecutor := &mockStepExecutor{}
	mockPublisher := &mockEventPublisher{}
	mockRepo := &mockRepository{}
	mockLogRepo := &mockLogRepository{}

	engine := NewCompensationEngineImpl(mockExecutor, mockPublisher, mockRepo, mockLogRepo)

	ctx := context.Background()
	execution := &models.SagaExecution{
		ID: "saga-123",
	}

	compSteps := []*saga.StepDefinition{
		{
			StepNumber:    1,
			HandlerMethod: "CancelMethod1",
			IsCritical:    true,
		},
	}

	err := engine.ExecuteCompensation(ctx, execution, 1, compSteps)

	if err != nil {
		t.Errorf("ExecuteCompensation failed: %v", err)
	}
}

func TestCompensationEngineGetCompensationStatus_Success(t *testing.T) {
	mockExecutor := &mockStepExecutor{}
	mockPublisher := &mockEventPublisher{}
	now := time.Now()
	mockRepo := &mockRepository{
		getFunc: func(ctx context.Context, sagaID string) (*models.SagaExecution, error) {
			return &models.SagaExecution{
				ID:                  sagaID,
				CompensationStatus:  models.CompensationCompleted,
				Status:              models.SagaStatusCompensated,
				StartedAt:           &now,
			}, nil
		},
	}
	mockLogRepo := &mockLogRepository{}

	engine := NewCompensationEngineImpl(mockExecutor, mockPublisher, mockRepo, mockLogRepo)

	ctx := context.Background()
	status, err := engine.GetCompensationStatus(ctx, "saga-123")

	if err != nil {
		t.Errorf("GetCompensationStatus failed: %v", err)
	}

	if status != models.CompensationCompleted {
		t.Errorf("Expected completed status, got %s", status)
	}
}

func TestCompensationEngineCanCompensate_Success(t *testing.T) {
	mockExecutor := &mockStepExecutor{}
	mockPublisher := &mockEventPublisher{}
	now := time.Now()
	mockRepo := &mockRepository{}
	mockLogRepo := &mockLogRepository{
		getFunc: func(ctx context.Context, sagaID string) ([]*models.StepExecution, error) {
			return []*models.StepExecution{
				{
					SagaID:     sagaID,
					StepNumber: 1,
					Status:     models.StepStatusSuccess,
				},
			}, nil
		},
	}

	engine := NewCompensationEngineImpl(mockExecutor, mockPublisher, mockRepo, mockLogRepo)

	ctx := context.Background()
	execution := &models.SagaExecution{
		ID:     "saga-123",
		Status: models.SagaStatusFailed,
		StartedAt: &now,
	}

	stepDefs := []*saga.StepDefinition{
		{
			StepNumber:    1,
			HandlerMethod: "Method1",
			IsCritical:    true,
			CompensationSteps: []*saga.StepDefinition{
				{
					StepNumber:    1,
					HandlerMethod: "CancelMethod1",
				},
			},
		},
	}

	canCompensate, err := engine.CanCompensate(ctx, execution, stepDefs)

	if err != nil {
		t.Errorf("CanCompensate failed: %v", err)
	}

	if !canCompensate {
		t.Fatal("Expected to be able to compensate")
	}
}

func TestCompensationEngineCanCompensate_InvalidStatus(t *testing.T) {
	mockExecutor := &mockStepExecutor{}
	mockPublisher := &mockEventPublisher{}
	now := time.Now()
	mockRepo := &mockRepository{}
	mockLogRepo := &mockLogRepository{}

	engine := NewCompensationEngineImpl(mockExecutor, mockPublisher, mockRepo, mockLogRepo)

	ctx := context.Background()
	execution := &models.SagaExecution{
		ID:        "saga-123",
		Status:    models.SagaStatusCompleted, // Wrong status
		StartedAt: &now,
	}

	stepDefs := []*saga.StepDefinition{}

	canCompensate, err := engine.CanCompensate(ctx, execution, stepDefs)

	if err == nil {
		t.Fatal("Expected error for invalid status")
	}

	if canCompensate {
		t.Fatal("Expected not to be able to compensate")
	}
}

func TestCompensationEngineRetryCompensation_Success(t *testing.T) {
	mockExecutor := &mockStepExecutor{}
	mockPublisher := &mockEventPublisher{}
	mockRepo := &mockRepository{}
	mockLogRepo := &mockLogRepository{
		getFunc: func(ctx context.Context, sagaID string) ([]*models.StepExecution, error) {
			return make([]*models.StepExecution, 0), nil
		},
	}

	engine := NewCompensationEngineImpl(mockExecutor, mockPublisher, mockRepo, mockLogRepo)

	ctx := context.Background()
	now := time.Now()
	execution := &models.SagaExecution{
		ID:                  "saga-123",
		Status:              models.SagaStatusCompensating,
		CompensationStatus:  models.CompensationFailed,
		StartedAt:           &now,
	}

	stepDefs := []*saga.StepDefinition{}

	err := engine.RetryCompensation(ctx, execution, stepDefs)

	if err != nil {
		t.Errorf("RetryCompensation failed: %v", err)
	}

	if execution.CompensationStatus != models.CompensationCompleted {
		t.Errorf("Expected completed status, got %s", execution.CompensationStatus)
	}
}

// Compensation Log Repository Tests

func TestCompensationLogRepositoryCreateLog_Success(t *testing.T) {
	repo := NewCompensationLogRepositoryImpl()

	ctx := context.Background()
	log := &CompensationLog{
		StepNumber:          1,
		CompensationStepNum: 1,
		Status:              models.StepStatusSuccess,
	}

	err := repo.CreateCompensationLog(ctx, "saga-123", log)

	if err != nil {
		t.Errorf("CreateCompensationLog failed: %v", err)
	}

	if log.SagaID != "saga-123" {
		t.Errorf("Expected saga-123, got %s", log.SagaID)
	}
}

func TestCompensationLogRepositoryGetLogs_Success(t *testing.T) {
	repo := NewCompensationLogRepositoryImpl()

	ctx := context.Background()
	log1 := &CompensationLog{StepNumber: 1, CompensationStepNum: 1, Status: models.StepStatusSuccess}
	log2 := &CompensationLog{StepNumber: 2, CompensationStepNum: 1, Status: models.StepStatusSuccess}

	repo.CreateCompensationLog(ctx, "saga-123", log1)
	repo.CreateCompensationLog(ctx, "saga-123", log2)

	logs, err := repo.GetCompensationLogs(ctx, "saga-123")

	if err != nil {
		t.Errorf("GetCompensationLogs failed: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(logs))
	}
}

func TestCompensationLogRepositoryGetLogsByStep_Success(t *testing.T) {
	repo := NewCompensationLogRepositoryImpl()

	ctx := context.Background()
	log1 := &CompensationLog{StepNumber: 1, CompensationStepNum: 1, Status: models.StepStatusSuccess}
	log2 := &CompensationLog{StepNumber: 2, CompensationStepNum: 1, Status: models.StepStatusSuccess}

	repo.CreateCompensationLog(ctx, "saga-123", log1)
	repo.CreateCompensationLog(ctx, "saga-123", log2)

	logs, err := repo.GetCompensationLogsByStep(ctx, "saga-123", 1)

	if err != nil {
		t.Errorf("GetCompensationLogsByStep failed: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}

	if logs[0].StepNumber != 1 {
		t.Errorf("Expected step 1, got %d", logs[0].StepNumber)
	}
}

func TestCompensationLogRepositoryGetStatistics_Success(t *testing.T) {
	repo := NewCompensationLogRepositoryImpl()

	ctx := context.Background()
	log1 := &CompensationLog{
		StepNumber:      1,
		Status:          models.StepStatusSuccess,
		ExecutionTimeMs: 100,
	}
	log2 := &CompensationLog{
		StepNumber:      2,
		Status:          models.StepStatusSuccess,
		ExecutionTimeMs: 200,
	}

	repo.CreateCompensationLog(ctx, "saga-123", log1)
	repo.CreateCompensationLog(ctx, "saga-123", log2)

	stats, err := repo.GetCompensationStatistics(ctx, "saga-123")

	if err != nil {
		t.Errorf("GetCompensationStatistics failed: %v", err)
	}

	if stats["totalLogs"].(int) != 2 {
		t.Errorf("Expected 2 total logs, got %d", stats["totalLogs"])
	}

	if stats["successfulLogs"].(int) != 2 {
		t.Errorf("Expected 2 successful logs, got %d", stats["successfulLogs"])
	}

	averageTime := stats["averageTimeMs"].(float64)
	if averageTime != 150.0 {
		t.Errorf("Expected average 150.0ms, got %f", averageTime)
	}
}

func TestCompensationLogRepositoryUpdateLog_Success(t *testing.T) {
	repo := NewCompensationLogRepositoryImpl()

	ctx := context.Background()
	log := &CompensationLog{
		StepNumber: 1,
		Status:     models.StepStatusSuccess,
	}

	repo.CreateCompensationLog(ctx, "saga-123", log)

	// Update log
	log.Status = models.StepStatusFailed
	err := repo.UpdateCompensationLog(ctx, "saga-123", log)

	if err != nil {
		t.Errorf("UpdateCompensationLog failed: %v", err)
	}

	logs, _ := repo.GetCompensationLogs(ctx, "saga-123")
	if logs[0].Status != models.StepStatusFailed {
		t.Errorf("Expected failed status, got %s", logs[0].Status)
	}
}

func TestCompensationLogRepositoryDeleteLogs_Success(t *testing.T) {
	repo := NewCompensationLogRepositoryImpl()

	ctx := context.Background()
	log := &CompensationLog{StepNumber: 1, Status: models.StepStatusSuccess}

	repo.CreateCompensationLog(ctx, "saga-123", log)

	err := repo.DeleteCompensationLogs(ctx, "saga-123")

	if err != nil {
		t.Errorf("DeleteCompensationLogs failed: %v", err)
	}

	logs, _ := repo.GetCompensationLogs(ctx, "saga-123")
	if len(logs) != 0 {
		t.Errorf("Expected 0 logs after delete, got %d", len(logs))
	}
}

func TestCompensationLogRepositoryClearLogs(t *testing.T) {
	repo := NewCompensationLogRepositoryImpl()

	ctx := context.Background()
	log1 := &CompensationLog{StepNumber: 1, Status: models.StepStatusSuccess}
	log2 := &CompensationLog{StepNumber: 2, Status: models.StepStatusSuccess}

	repo.CreateCompensationLog(ctx, "saga-123", log1)
	repo.CreateCompensationLog(ctx, "saga-456", log2)

	repo.ClearLogs()

	count := repo.GetLogCount()
	if count != 0 {
		t.Errorf("Expected 0 logs after clear, got %d", count)
	}
}
