package algorithms

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"p9e.in/chetana/packages/loadbalancer"
	"p9e.in/chetana/packages/registry"
)

// RoundRobinBalancer implements round-robin load balancing
// Cycles through endpoints sequentially, wrapping around when reaching the end
type RoundRobinBalancer struct {
	index   int64
	metrics map[string]*loadbalancer.EndpointMetrics
	conns   map[string]int64
	mu      sync.RWMutex
}

// NewRoundRobinBalancer creates a new round-robin load balancer
func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{
		index:   0,
		metrics: make(map[string]*loadbalancer.EndpointMetrics),
		conns:   make(map[string]int64),
	}
}

// Select returns the next endpoint in round-robin order
func (rb *RoundRobinBalancer) Select(ctx context.Context, endpoints []*registry.ServiceInstance) (*loadbalancer.Endpoint, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints available")
	}

	// Get next index
	currentIndex := atomic.AddInt64(&rb.index, 1) - 1
	selectedIdx := int(currentIndex) % len(endpoints)

	endpoint := &loadbalancer.Endpoint{
		Instance:          endpoints[selectedIdx],
		Weight:            100,
		ActiveConnections: rb.getConnections(endpoints[selectedIdx].ID),
		Metrics:           rb.getOrCreateMetrics(endpoints[selectedIdx].ID),
	}

	return endpoint, nil
}

// RecordMetrics records metrics for an endpoint
func (rb *RoundRobinBalancer) RecordMetrics(instanceID string, latency time.Duration, success bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	metrics := rb.getOrCreateMetricsLocked(instanceID)

	metrics.TotalRequests++
	if success {
		metrics.SuccessCount++
	} else {
		metrics.FailureCount++
	}

	// Update latency stats
	if latency < metrics.MinLatency || metrics.MinLatency == 0 {
		metrics.MinLatency = latency
	}
	if latency > metrics.MaxLatency {
		metrics.MaxLatency = latency
	}

	// Simple average calculation (can be improved with rolling average)
	if metrics.AvgLatency == 0 {
		metrics.AvgLatency = latency
	} else {
		metrics.AvgLatency = (metrics.AvgLatency + latency) / 2
	}

	metrics.ErrorRate = float64(metrics.FailureCount) / float64(metrics.TotalRequests)
	metrics.LastUpdate = time.Now()
}

// IncrementConnections increments the active connection count
func (rb *RoundRobinBalancer) IncrementConnections(instanceID string) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.conns[instanceID]++
}

// DecrementConnections decrements the active connection count
func (rb *RoundRobinBalancer) DecrementConnections(instanceID string) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if count, ok := rb.conns[instanceID]; ok && count > 0 {
		rb.conns[instanceID]--
	}
}

// Reset clears internal state
func (rb *RoundRobinBalancer) Reset() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.index = 0
	rb.metrics = make(map[string]*loadbalancer.EndpointMetrics)
	rb.conns = make(map[string]int64)
}

// GetMetrics returns metrics for an endpoint
func (rb *RoundRobinBalancer) GetMetrics(instanceID string) *loadbalancer.EndpointMetrics {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.metrics[instanceID]
}

// Helper methods

func (rb *RoundRobinBalancer) getOrCreateMetrics(instanceID string) *loadbalancer.EndpointMetrics {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.getOrCreateMetricsLocked(instanceID)
}

func (rb *RoundRobinBalancer) getOrCreateMetricsLocked(instanceID string) *loadbalancer.EndpointMetrics {
	if metrics, ok := rb.metrics[instanceID]; ok {
		return metrics
	}

	metrics := &loadbalancer.EndpointMetrics{
		LastUpdate: time.Now(),
	}
	rb.metrics[instanceID] = metrics
	return metrics
}

func (rb *RoundRobinBalancer) getConnections(instanceID string) int64 {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.conns[instanceID]
}
