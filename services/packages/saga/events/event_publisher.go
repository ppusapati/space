// Package events implements event publishing for saga lifecycle events
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"p9e.in/chetana/packages/saga/models"
)

// EventPublisherImpl publishes saga events to event bus (Kafka)
type EventPublisherImpl struct {
	mu              sync.RWMutex
	topic           string
	eventBuffer     []*models.SagaEvent
	bufferSize      int
	kafkaProducer   KafkaProducer
	eventIDGen      func() string
}

// KafkaProducer interface for event publishing
type KafkaProducer interface {
	ProduceMessage(ctx context.Context, topic, key string, value []byte) error
	Close() error
}

// MockKafkaProducer for testing
type MockKafkaProducer struct {
	messages []ProducedMessage
	mu       sync.Mutex
}

type ProducedMessage struct {
	Topic string
	Key   string
	Value []byte
}

func (m *MockKafkaProducer) ProduceMessage(ctx context.Context, topic, key string, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = append(m.messages, ProducedMessage{
		Topic: topic,
		Key:   key,
		Value: value,
	})

	return nil
}

func (m *MockKafkaProducer) Close() error {
	return nil
}

func (m *MockKafkaProducer) GetMessages() []ProducedMessage {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]ProducedMessage, len(m.messages))
	copy(result, m.messages)
	return result
}

// NewEventPublisherImpl creates a new event publisher instance
func NewEventPublisherImpl(
	topic string,
	kafkaProducer KafkaProducer,
) *EventPublisherImpl {
	return &EventPublisherImpl{
		topic:         topic,
		eventBuffer:   make([]*models.SagaEvent, 0, 100),
		bufferSize:    100,
		kafkaProducer: kafkaProducer,
		eventIDGen: func() string {
			return fmt.Sprintf("event-%d", time.Now().UnixNano())
		},
	}
}

// PublishSagaStarted publishes a saga-lifecycle "started" event. Added
// 2026-04-19 (B.8) to match the orchestrator's expectation that a saga
// emits a start event distinct from step-level starts.
func (p *EventPublisherImpl) PublishSagaStarted(
	ctx context.Context,
	execution *models.SagaExecution,
) error {
	event := &models.SagaEvent{
		EventID:   p.eventIDGen(),
		SagaID:    execution.ID,
		SagaType:  execution.SagaType,
		EventType: models.SagaEventType("SAGA.STARTED"),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"status":     "started",
			"totalSteps": execution.TotalSteps,
		},
	}

	return p.publishEvent(ctx, event)
}

// PublishStepStarted publishes a step started event
func (p *EventPublisherImpl) PublishStepStarted(
	ctx context.Context,
	execution *models.SagaExecution,
	stepNum int32,
) error {
	event := &models.SagaEvent{
		EventID:    p.eventIDGen(),
		SagaID:     execution.ID,
		SagaType:   execution.SagaType,
		EventType:  models.SagaEventStepStarted,
		StepNumber: stepNum,
		Timestamp:  time.Now(),
		Data: map[string]interface{}{
			"status": "started",
		},
	}

	return p.publishEvent(ctx, event)
}

// PublishStepCompleted publishes a step completed event
func (p *EventPublisherImpl) PublishStepCompleted(
	ctx context.Context,
	execution *models.SagaExecution,
	stepNum int32,
	result *models.StepResult,
) error {
	event := &models.SagaEvent{
		EventID:    p.eventIDGen(),
		SagaID:     execution.ID,
		SagaType:   execution.SagaType,
		EventType:  models.SagaEventStepCompleted,
		StepNumber: stepNum,
		Timestamp:  time.Now(),
		Data: map[string]interface{}{
			"status":          "completed",
			"executionTimeMs": result.ExecutionTimeMs,
			"retryCount":      result.RetryCount,
		},
	}

	return p.publishEvent(ctx, event)
}

// PublishStepFailed publishes a step failed event
func (p *EventPublisherImpl) PublishStepFailed(
	ctx context.Context,
	execution *models.SagaExecution,
	stepNum int32,
	err error,
) error {
	event := &models.SagaEvent{
		EventID:    p.eventIDGen(),
		SagaID:     execution.ID,
		SagaType:   execution.SagaType,
		EventType:  models.SagaEventStepFailed,
		StepNumber: stepNum,
		Timestamp:  time.Now(),
		Data: map[string]interface{}{
			"status": "failed",
			"error":  err.Error(),
		},
	}

	return p.publishEvent(ctx, event)
}

// PublishStepRetrying publishes a step retrying event
func (p *EventPublisherImpl) PublishStepRetrying(
	ctx context.Context,
	execution *models.SagaExecution,
	stepNum int32,
	err error,
) error {
	event := &models.SagaEvent{
		EventID:    p.eventIDGen(),
		SagaID:     execution.ID,
		SagaType:   execution.SagaType,
		EventType:  models.SagaEventStepRetrying,
		StepNumber: stepNum,
		Timestamp:  time.Now(),
		Data: map[string]interface{}{
			"status": "retrying",
			"error":  err.Error(),
		},
	}

	return p.publishEvent(ctx, event)
}

// PublishSagaCompleted publishes a saga completed event
func (p *EventPublisherImpl) PublishSagaCompleted(
	ctx context.Context,
	execution *models.SagaExecution,
) error {
	event := &models.SagaEvent{
		EventID:   p.eventIDGen(),
		SagaID:    execution.ID,
		SagaType:  execution.SagaType,
		EventType: models.SagaEventSagaCompleted,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"status":           "completed",
			"stepsExecuted":    execution.CurrentStep,
			"totalSteps":       execution.TotalSteps,
			"executionTimeMs":  execution.CompletedAt.Sub(*execution.StartedAt).Milliseconds(),
		},
	}

	return p.publishEvent(ctx, event)
}

// PublishSagaFailed publishes a saga failed event
func (p *EventPublisherImpl) PublishSagaFailed(
	ctx context.Context,
	execution *models.SagaExecution,
) error {
	event := &models.SagaEvent{
		EventID:   p.eventIDGen(),
		SagaID:    execution.ID,
		SagaType:  execution.SagaType,
		EventType: models.SagaEventSagaFailed,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"status":          "failed",
			"errorCode":       execution.ErrorCode,
			"errorMessage":    execution.ErrorMessage,
			"failedAtStep":    execution.CurrentStep,
			"totalSteps":      execution.TotalSteps,
		},
	}

	return p.publishEvent(ctx, event)
}

// PublishCompensationStarted publishes a compensation started event
func (p *EventPublisherImpl) PublishCompensationStarted(
	ctx context.Context,
	execution *models.SagaExecution,
) error {
	event := &models.SagaEvent{
		EventID:   p.eventIDGen(),
		SagaID:    execution.ID,
		SagaType:  execution.SagaType,
		EventType: models.SagaEventCompensationStarted,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"status":     "compensation_started",
			"failedStep": execution.CurrentStep,
		},
	}

	return p.publishEvent(ctx, event)
}

// PublishCompensationCompleted publishes a compensation completed event
func (p *EventPublisherImpl) PublishCompensationCompleted(
	ctx context.Context,
	execution *models.SagaExecution,
) error {
	event := &models.SagaEvent{
		EventID:   p.eventIDGen(),
		SagaID:    execution.ID,
		SagaType:  execution.SagaType,
		EventType: models.SagaEventCompensationCompleted,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"status":                  string(execution.CompensationStatus),
			"compensationCompleted":   true,
		},
	}

	return p.publishEvent(ctx, event)
}

// publishEvent publishes an event to Kafka
func (p *EventPublisherImpl) publishEvent(ctx context.Context, event *models.SagaEvent) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 1. Serialize event to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// 2. Use saga ID as partition key for ordering
	partitionKey := event.SagaID

	// 3. Publish to Kafka
	if err := p.kafkaProducer.ProduceMessage(ctx, p.topic, partitionKey, payload); err != nil {
		return fmt.Errorf("failed to produce message to Kafka: %w", err)
	}

	// 4. Buffer event for local retrieval
	p.eventBuffer = append(p.eventBuffer, event)
	if len(p.eventBuffer) > p.bufferSize {
		// Keep only most recent events
		p.eventBuffer = p.eventBuffer[len(p.eventBuffer)-p.bufferSize:]
	}

	return nil
}

// GetPublishedEvents returns published events (for testing)
func (p *EventPublisherImpl) GetPublishedEvents() []*models.SagaEvent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*models.SagaEvent, len(p.eventBuffer))
	copy(result, p.eventBuffer)
	return result
}

// GetPublishedEventsBySaga returns events for a specific saga
func (p *EventPublisherImpl) GetPublishedEventsBySaga(sagaID string) []*models.SagaEvent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*models.SagaEvent, 0)
	for _, event := range p.eventBuffer {
		if event.SagaID == sagaID {
			result = append(result, event)
		}
	}

	return result
}

// ClearBuffer clears the event buffer (for testing)
func (p *EventPublisherImpl) ClearBuffer() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.eventBuffer = make([]*models.SagaEvent, 0, p.bufferSize)
}
