package algorithms

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"p9e.in/samavaya/packages/loadbalancer"
	"p9e.in/samavaya/packages/registry"
)

// RandomBalancer selects endpoints randomly
// Useful as a fallback or for simple workloads without affinity requirements
type RandomBalancer struct {
	rng     *rand.Rand
	metrics map[string]*loadbalancer.EndpointMetrics
	conns   map[string]int64
	mu      sync.RWMutex
}

// NewRandomBalancer creates a new random load balancer
func NewRandomBalancer() *RandomBalancer {
	return &RandomBalancer{
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
		metrics: make(map[string]*loadbalancer.EndpointMetrics),
		conns:   make(map[string]int64),
	}
}

// Select returns a random endpoint
func (rb *RandomBalancer) Select(ctx context.Context, endpoints []*registry.ServiceInstance) (*loadbalancer.Endpoint, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints available")
	}

	rb.mu.RLock()
	defer rb.mu.RUnlock()

	// Select random index
	selectedIdx := rb.rng.Intn(len(endpoints))
	selectedEp := endpoints[selectedIdx]

	endpoint := &loadbalancer.Endpoint{
		Instance:          selectedEp,
		Weight:            100,
		ActiveConnections: rb.conns[selectedEp.ID],
		Metrics:           rb.getOrCreateMetricsLocked(selectedEp.ID),
	}

	return endpoint, nil
}

// RecordMetrics records metrics for an endpoint
func (rb *RandomBalancer) RecordMetrics(instanceID string, latency time.Duration, success bool) {
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

	if metrics.AvgLatency == 0 {
		metrics.AvgLatency = latency
	} else {
		metrics.AvgLatency = (metrics.AvgLatency + latency) / 2
	}

	metrics.ErrorRate = float64(metrics.FailureCount) / float64(metrics.TotalRequests)
	metrics.LastUpdate = time.Now()
}

// IncrementConnections increments the active connection count
func (rb *RandomBalancer) IncrementConnections(instanceID string) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.conns[instanceID]++
}

// DecrementConnections decrements the active connection count
func (rb *RandomBalancer) DecrementConnections(instanceID string) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if count, ok := rb.conns[instanceID]; ok && count > 0 {
		rb.conns[instanceID]--
	}
}

// Reset clears internal state
func (rb *RandomBalancer) Reset() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.metrics = make(map[string]*loadbalancer.EndpointMetrics)
	rb.conns = make(map[string]int64)
}

// GetMetrics returns metrics for an endpoint
func (rb *RandomBalancer) GetMetrics(instanceID string) *loadbalancer.EndpointMetrics {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.metrics[instanceID]
}

// Helper methods

func (rb *RandomBalancer) getOrCreateMetricsLocked(instanceID string) *loadbalancer.EndpointMetrics {
	if metrics, ok := rb.metrics[instanceID]; ok {
		return metrics
	}

	metrics := &loadbalancer.EndpointMetrics{
		LastUpdate: time.Now(),
	}
	rb.metrics[instanceID] = metrics
	return metrics
}
