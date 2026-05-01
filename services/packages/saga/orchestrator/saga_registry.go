// Package orchestrator provides saga registry for handler management
package orchestrator

import (
	"fmt"
	"sync"

	"p9e.in/samavaya/packages/saga"
)

// SagaRegistry manages registration and retrieval of saga handlers
type SagaRegistry struct {
	mu       sync.RWMutex
	handlers map[string]saga.SagaHandler
}

// NewSagaRegistry creates a new saga registry
func NewSagaRegistry() *SagaRegistry {
	return &SagaRegistry{
		handlers: make(map[string]saga.SagaHandler),
	}
}

// RegisterHandler registers a handler for a saga type
func (r *SagaRegistry) RegisterHandler(sagaType string, handler saga.SagaHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if sagaType == "" {
		return fmt.Errorf("saga type cannot be empty")
	}

	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	// Check if already registered
	if _, exists := r.handlers[sagaType]; exists {
		return fmt.Errorf("handler for saga type %s is already registered", sagaType)
	}

	r.handlers[sagaType] = handler
	return nil
}

// GetHandler retrieves a handler for a saga type
func (r *SagaRegistry) GetHandler(sagaType string) (saga.SagaHandler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, exists := r.handlers[sagaType]
	if !exists {
		return nil, fmt.Errorf("no handler registered for saga type %s", sagaType)
	}

	return handler, nil
}

// HasHandler checks if a handler exists for a saga type
func (r *SagaRegistry) HasHandler(sagaType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.handlers[sagaType]
	return exists
}

// GetAllHandlers returns all registered handlers
func (r *SagaRegistry) GetAllHandlers() map[string]saga.SagaHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]saga.SagaHandler)
	for k, v := range r.handlers {
		result[k] = v
	}

	return result
}

// RemoveHandler removes a handler for a saga type (for testing)
func (r *SagaRegistry) RemoveHandler(sagaType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.handlers[sagaType]; !exists {
		return fmt.Errorf("handler for saga type %s not found", sagaType)
	}

	delete(r.handlers, sagaType)
	return nil
}

// ClearHandlers removes all handlers (for testing)
func (r *SagaRegistry) ClearHandlers() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlers = make(map[string]saga.SagaHandler)
}

// GetHandlerCount returns number of registered handlers
func (r *SagaRegistry) GetHandlerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.handlers)
}
