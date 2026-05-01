package algorithms

import (
	"context"
	"fmt"
	"sync"
	"time"

	"p9e.in/samavaya/packages/loadbalancer"
	"p9e.in/samavaya/packages/registry"
)

// LeastConnectionsBalancer selects endpoints with the fewest active connections
// Ideal for long-lived connections or streaming scenarios
type LeastConnectionsBalancer struct {
	metrics map[string]*loadbalancer.EndpointMetrics
	conns   map[string]int64
	mu      sync.RWMutex
}

// NewLeastConnectionsBalancer creates a new least-connections load balancer
func NewLeastConnectionsBalancer() *LeastConnectionsBalancer {
	return &LeastConnectionsBalancer{
		metrics: make(map[string]*loadbalancer.EndpointMetrics),
		conns:   make(map[string]int64),
	}
}

// Select returns the endpoint with the fewest active connections
func (lcb *LeastConnectionsBalancer) Select(ctx context.Context, endpoints []*registry.ServiceInstance) (*loadbalancer.Endpoint, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints available")
	}

	lcb.mu.RLock()
	defer lcb.mu.RUnlock()

	var selected *loadbalancer.Endpoint
	minConnections := int64(^uint64(0) >> 1) // Max int64

	for _, ep := range endpoints {
		connCount := lcb.conns[ep.ID]

		if connCount < minConnections {
			minConnections = connCount
			selected = &loadbalancer.Endpoint{
				Instance:          ep,
				Weight:            100,
				ActiveConnections: connCount,
				Metrics:           lcb.getOrCreateMetricsLocked(ep.ID),
			}
		}
	}

	if selected == nil && len(endpoints) > 0 {
		selected = &loadbalancer.Endpoint{
			Instance:          endpoints[0],
			Weight:            100,
			ActiveConnections: lcb.conns[endpoints[0].ID],
			Metrics:           lcb.getOrCreateMetricsLocked(endpoints[0].ID),
		}
	}

	return selected, nil
}

// RecordMetrics records metrics for an endpoint
func (lcb *LeastConnectionsBalancer) RecordMetrics(instanceID string, latency time.Duration, success bool) {
	lcb.mu.Lock()
	defer lcb.mu.Unlock()

	metrics := lcb.getOrCreateMetricsLocked(instanceID)

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

	if metrics.AvgLatency == 0 {
		metrics.AvgLatency = latency
	} else {
		metrics.AvgLatency = (metrics.AvgLatency + latency) / 2
	}

	metrics.ErrorRate = float64(metrics.FailureCount) / float64(metrics.TotalRequests)
	metrics.LastUpdate = time.Now()
}

// IncrementConnections increments the active connection count
func (lcb *LeastConnectionsBalancer) IncrementConnections(instanceID string) {
	lcb.mu.Lock()
	defer lcb.mu.Unlock()
	lcb.conns[instanceID]++
}

// DecrementConnections decrements the active connection count
func (lcb *LeastConnectionsBalancer) DecrementConnections(instanceID string) {
	lcb.mu.Lock()
	defer lcb.mu.Unlock()
	if count, ok := lcb.conns[instanceID]; ok && count > 0 {
		lcb.conns[instanceID]--
	}
}

// Reset clears internal state
func (lcb *LeastConnectionsBalancer) Reset() {
	lcb.mu.Lock()
	defer lcb.mu.Unlock()

	lcb.metrics = make(map[string]*loadbalancer.EndpointMetrics)
	lcb.conns = make(map[string]int64)
}

// GetMetrics returns metrics for an endpoint
func (lcb *LeastConnectionsBalancer) GetMetrics(instanceID string) *loadbalancer.EndpointMetrics {
	lcb.mu.RLock()
	defer lcb.mu.RUnlock()
	return lcb.metrics[instanceID]
}

// Helper methods

func (lcb *LeastConnectionsBalancer) getOrCreateMetricsLocked(instanceID string) *loadbalancer.EndpointMetrics {
	if metrics, ok := lcb.metrics[instanceID]; ok {
		return metrics
	}

	metrics := &loadbalancer.EndpointMetrics{
		LastUpdate: time.Now(),
	}
	lcb.metrics[instanceID] = metrics
	return metrics
}
