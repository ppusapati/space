// Package events contains unit tests for event publishing
package events

import (
	"context"
	"testing"
	"time"

	"p9e.in/samavaya/packages/saga/models"
)

// Event Publisher Tests

func TestEventPublisherPublishStepStarted_Success(t *testing.T) {
	mockProducer := &MockKafkaProducer{}
	publisher := NewEventPublisherImpl("saga-events", mockProducer)

	ctx := context.Background()
	now := time.Now()
	execution := &models.SagaExecution{
		ID:        "saga-123",
		SagaType:  "SAGA-S01",
		StartedAt: &now,
	}

	err := publisher.PublishStepStarted(ctx, execution, 1)

	if err != nil {
		t.Errorf("PublishStepStarted failed: %v", err)
	}

	messages := mockProducer.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].Topic != "saga-events" {
		t.Errorf("Expected topic saga-events, got %s", messages[0].Topic)
	}

	if messages[0].Key != "saga-123" {
		t.Errorf("Expected key saga-123, got %s", messages[0].Key)
	}
}

func TestEventPublisherPublishStepCompleted_Success(t *testing.T) {
	mockProducer := &MockKafkaProducer{}
	publisher := NewEventPublisherImpl("saga-events", mockProducer)

	ctx := context.Background()
	now := time.Now()
	execution := &models.SagaExecution{
		ID:        "saga-123",
		SagaType:  "SAGA-S01",
		StartedAt: &now,
	}

	result := &models.StepResult{
		StepNumber:      1,
		Status:          models.StepStatusSuccess,
		ExecutionTimeMs: 100,
		RetryCount:      0,
	}

	err := publisher.PublishStepCompleted(ctx, execution, 1, result)

	if err != nil {
		t.Errorf("PublishStepCompleted failed: %v", err)
	}

	messages := mockProducer.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestEventPublisherPublishStepFailed_Success(t *testing.T) {
	mockProducer := &MockKafkaProducer{}
	publisher := NewEventPublisherImpl("saga-events", mockProducer)

	ctx := context.Background()
	now := time.Now()
	execution := &models.SagaExecution{
		ID:        "saga-123",
		SagaType:  "SAGA-S01",
		StartedAt: &now,
	}

	err := publisher.PublishStepFailed(ctx, execution, 1, errors.New("step failed"))

	if err != nil {
		t.Errorf("PublishStepFailed failed: %v", err)
	}

	messages := mockProducer.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestEventPublisherPublishStepRetrying_Success(t *testing.T) {
	mockProducer := &MockKafkaProducer{}
	publisher := NewEventPublisherImpl("saga-events", mockProducer)

	ctx := context.Background()
	now := time.Now()
	execution := &models.SagaExecution{
		ID:        "saga-123",
		SagaType:  "SAGA-S01",
		StartedAt: &now,
	}

	err := publisher.PublishStepRetrying(ctx, execution, 1, errors.New("retry"))

	if err != nil {
		t.Errorf("PublishStepRetrying failed: %v", err)
	}

	messages := mockProducer.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestEventPublisherPublishSagaCompleted_Success(t *testing.T) {
	mockProducer := &MockKafkaProducer{}
	publisher := NewEventPublisherImpl("saga-events", mockProducer)

	ctx := context.Background()
	now := time.Now()
	execution := &models.SagaExecution{
		ID:         "saga-123",
		SagaType:   "SAGA-S01",
		StartedAt:  &now,
		CompletedAt: &now,
		CurrentStep: 3,
		TotalSteps:  3,
	}

	err := publisher.PublishSagaCompleted(ctx, execution)

	if err != nil {
		t.Errorf("PublishSagaCompleted failed: %v", err)
	}

	messages := mockProducer.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestEventPublisherPublishSagaFailed_Success(t *testing.T) {
	mockProducer := &MockKafkaProducer{}
	publisher := NewEventPublisherImpl("saga-events", mockProducer)

	ctx := context.Background()
	now := time.Now()
	execution := &models.SagaExecution{
		ID:           "saga-123",
		SagaType:     "SAGA-S01",
		StartedAt:    &now,
		CompletedAt:  &now,
		CurrentStep:  2,
		TotalSteps:   3,
		ErrorCode:    "STEP_FAILED",
		ErrorMessage: "Step 2 failed",
	}

	err := publisher.PublishSagaFailed(ctx, execution)

	if err != nil {
		t.Errorf("PublishSagaFailed failed: %v", err)
	}

	messages := mockProducer.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestEventPublisherPublishCompensationStarted_Success(t *testing.T) {
	mockProducer := &MockKafkaProducer{}
	publisher := NewEventPublisherImpl("saga-events", mockProducer)

	ctx := context.Background()
	now := time.Now()
	execution := &models.SagaExecution{
		ID:               "saga-123",
		SagaType:         "SAGA-S01",
		StartedAt:        &now,
		CurrentStep:      2,
		CompensationStatus: models.CompensationRunning,
	}

	err := publisher.PublishCompensationStarted(ctx, execution)

	if err != nil {
		t.Errorf("PublishCompensationStarted failed: %v", err)
	}

	messages := mockProducer.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestEventPublisherPublishCompensationCompleted_Success(t *testing.T) {
	mockProducer := &MockKafkaProducer{}
	publisher := NewEventPublisherImpl("saga-events", mockProducer)

	ctx := context.Background()
	now := time.Now()
	execution := &models.SagaExecution{
		ID:                  "saga-123",
		SagaType:           "SAGA-S01",
		StartedAt:          &now,
		CompensationStatus: models.CompensationCompleted,
	}

	err := publisher.PublishCompensationCompleted(ctx, execution)

	if err != nil {
		t.Errorf("PublishCompensationCompleted failed: %v", err)
	}

	messages := mockProducer.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestEventPublisherGetPublishedEvents(t *testing.T) {
	mockProducer := &MockKafkaProducer{}
	publisher := NewEventPublisherImpl("saga-events", mockProducer)

	ctx := context.Background()
	now := time.Now()
	execution := &models.SagaExecution{
		ID:        "saga-123",
		SagaType:  "SAGA-S01",
		StartedAt: &now,
	}

	publisher.PublishStepStarted(ctx, execution, 1)
	publisher.PublishStepCompleted(ctx, execution, 1, &models.StepResult{})

	events := publisher.GetPublishedEvents()

	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	if events[0].EventType != models.SagaEventTypeStepStarted {
		t.Errorf("Expected first event to be step started")
	}

	if events[1].EventType != models.SagaEventTypeStepCompleted {
		t.Errorf("Expected second event to be step completed")
	}
}

func TestEventPublisherGetPublishedEventsBySaga(t *testing.T) {
	mockProducer := &MockKafkaProducer{}
	publisher := NewEventPublisherImpl("saga-events", mockProducer)

	ctx := context.Background()
	now := time.Now()

	execution1 := &models.SagaExecution{
		ID:        "saga-123",
		SagaType:  "SAGA-S01",
		StartedAt: &now,
	}

	execution2 := &models.SagaExecution{
		ID:        "saga-456",
		SagaType:  "SAGA-S01",
		StartedAt: &now,
	}

	publisher.PublishStepStarted(ctx, execution1, 1)
	publisher.PublishStepStarted(ctx, execution2, 1)
	publisher.PublishStepStarted(ctx, execution1, 2)

	events := publisher.GetPublishedEventsBySaga("saga-123")

	if len(events) != 2 {
		t.Errorf("Expected 2 events for saga-123, got %d", len(events))
	}

	for _, event := range events {
		if event.SagaID != "saga-123" {
			t.Errorf("Expected saga-123, got %s", event.SagaID)
		}
	}
}

func TestEventPublisherClearBuffer(t *testing.T) {
	mockProducer := &MockKafkaProducer{}
	publisher := NewEventPublisherImpl("saga-events", mockProducer)

	ctx := context.Background()
	now := time.Now()
	execution := &models.SagaExecution{
		ID:        "saga-123",
		SagaType:  "SAGA-S01",
		StartedAt: &now,
	}

	publisher.PublishStepStarted(ctx, execution, 1)

	publisher.ClearBuffer()

	events := publisher.GetPublishedEvents()
	if len(events) != 0 {
		t.Errorf("Expected 0 events after clear, got %d", len(events))
	}
}

// Import errors package if not already present
import "errors"
