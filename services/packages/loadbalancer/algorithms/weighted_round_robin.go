package algorithms

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"p9e.in/samavaya/packages/loadbalancer"
	"p9e.in/samavaya/packages/registry"
)

// WeightedRoundRobinBalancer implements weighted round-robin load balancing
// Used for canary deployments, A/B testing, and traffic gradual rollouts
type WeightedRoundRobinBalancer struct {
	index   int64
	weights map[string]int // instance_id -> weight (0-100)
	metrics map[string]*loadbalancer.EndpointMetrics
	conns   map[string]int64
	mu      sync.RWMutex
}

// NewWeightedRoundRobinBalancer creates a new weighted round-robin load balancer
func NewWeightedRoundRobinBalancer(weights map[string]int) *WeightedRoundRobinBalancer {
	if weights == nil {
		weights = make(map[string]int)
	}

	return &WeightedRoundRobinBalancer{
		index:   0,
		weights: weights,
		metrics: make(map[string]*loadbalancer.EndpointMetrics),
		conns:   make(map[string]int64),
	}
}

// Select returns the next endpoint based on weights
func (wrb *WeightedRoundRobinBalancer) Select(ctx context.Context, endpoints []*registry.ServiceInstance) (*loadbalancer.Endpoint, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints available")
	}

	wrb.mu.RLock()
	defer wrb.mu.RUnlock()

	// Build weighted list
	var totalWeight int
	var weightedList []*registry.ServiceInstance

	for _, ep := range endpoints {
		weight := wrb.weights[ep.ID]
		if weight == 0 {
			weight = 100 // Default weight
		}

		// Add endpoint to weighted list (weight times)
		for i := 0; i < weight; i++ {
			weightedList = append(weightedList, ep)
		}

		totalWeight += weight
	}

	if len(weightedList) == 0 {
		return nil, fmt.Errorf("no endpoints with weight available")
	}

	// Select based on index
	currentIndex := atomic.AddInt64(&wrb.index, 1) - 1
	selectedIdx := int(currentIndex) % len(weightedList)
	selectedEp := weightedList[selectedIdx]

	endpoint := &loadbalancer.Endpoint{
		Instance:          selectedEp,
		Weight:            wrb.weights[selectedEp.ID],
		ActiveConnections: wrb.conns[selectedEp.ID],
		Metrics:           wrb.getOrCreateMetricsLocked(selectedEp.ID),
	}

	return endpoint, nil
}

// UpdateWeights updates the weights for endpoints
func (wrb *WeightedRoundRobinBalancer) UpdateWeights(weights map[string]int) {
	wrb.mu.Lock()
	defer wrb.mu.Unlock()

	wrb.weights = weights
}

// SetWeight sets the weight for a single endpoint
func (wrb *WeightedRoundRobinBalancer) SetWeight(instanceID string, weight int) {
	if weight < 0 {
		weight = 0
	}
	if weight > 100 {
		weight = 100
	}

	wrb.mu.Lock()
	defer wrb.mu.Unlock()

	wrb.weights[instanceID] = weight
}

// GetWeight gets the current weight for an endpoint
func (wrb *WeightedRoundRobinBalancer) GetWeight(instanceID string) int {
	wrb.mu.RLock()
	defer wrb.mu.RUnlock()

	weight := wrb.weights[instanceID]
	if weight == 0 {
		return 100 // Default weight
	}
	return weight
}

// RecordMetrics records metrics for an endpoint
func (wrb *WeightedRoundRobinBalancer) RecordMetrics(instanceID string, latency time.Duration, success bool) {
	wrb.mu.Lock()
	defer wrb.mu.Unlock()

	metrics := wrb.getOrCreateMetricsLocked(instanceID)

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
func (wrb *WeightedRoundRobinBalancer) IncrementConnections(instanceID string) {
	wrb.mu.Lock()
	defer wrb.mu.Unlock()
	wrb.conns[instanceID]++
}

// DecrementConnections decrements the active connection count
func (wrb *WeightedRoundRobinBalancer) DecrementConnections(instanceID string) {
	wrb.mu.Lock()
	defer wrb.mu.Unlock()
	if count, ok := wrb.conns[instanceID]; ok && count > 0 {
		wrb.conns[instanceID]--
	}
}

// Reset clears internal state
func (wrb *WeightedRoundRobinBalancer) Reset() {
	wrb.mu.Lock()
	defer wrb.mu.Unlock()

	wrb.index = 0
	wrb.metrics = make(map[string]*loadbalancer.EndpointMetrics)
	wrb.conns = make(map[string]int64)
}

// GetMetrics returns metrics for an endpoint
func (wrb *WeightedRoundRobinBalancer) GetMetrics(instanceID string) *loadbalancer.EndpointMetrics {
	wrb.mu.RLock()
	defer wrb.mu.RUnlock()
	return wrb.metrics[instanceID]
}

// Helper methods

func (wrb *WeightedRoundRobinBalancer) getOrCreateMetricsLocked(instanceID string) *loadbalancer.EndpointMetrics {
	if metrics, ok := wrb.metrics[instanceID]; ok {
		return metrics
	}

	metrics := &loadbalancer.EndpointMetrics{
		LastUpdate: time.Now(),
	}
	wrb.metrics[instanceID] = metrics
	return metrics
}
