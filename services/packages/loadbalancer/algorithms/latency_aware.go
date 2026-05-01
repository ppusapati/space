package algorithms

import (
	"context"
	"fmt"
	"sync"
	"time"

	"p9e.in/samavaya/packages/loadbalancer"
	"p9e.in/samavaya/packages/registry"
)

// LatencyAwareBalancer selects endpoints based on lowest latency
// Uses exponential weighted moving average for smooth latency tracking
type LatencyAwareBalancer struct {
	metrics map[string]*loadbalancer.EndpointMetrics
	conns   map[string]int64
	mu      sync.RWMutex

	// Alpha for EWMA calculation (0.0-1.0, default 0.3)
	// Higher value gives more weight to recent measurements
	alpha float64
}

// NewLatencyAwareBalancer creates a new latency-aware load balancer
func NewLatencyAwareBalancer() *LatencyAwareBalancer {
	return &LatencyAwareBalancer{
		metrics: make(map[string]*loadbalancer.EndpointMetrics),
		conns:   make(map[string]int64),
		alpha:   0.3, // Default EWMA alpha
	}
}

// Select returns the endpoint with the lowest average latency
func (lab *LatencyAwareBalancer) Select(ctx context.Context, endpoints []*registry.ServiceInstance) (*loadbalancer.Endpoint, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints available")
	}

	lab.mu.RLock()
	defer lab.mu.RUnlock()

	var selected *loadbalancer.Endpoint
	minLatency := time.Duration(^uint64(0) >> 1) // Max duration

	for _, ep := range endpoints {
		metrics := lab.getOrCreateMetricsLocked(ep.ID)

		// Use average latency, fallback to max latency if no data
		latency := metrics.AvgLatency
		if latency == 0 {
			latency = metrics.MaxLatency
		}

		// If still no data, treat as fast
		if latency == 0 {
			latency = 1 * time.Millisecond
		}

		if latency < minLatency {
			minLatency = latency
			selected = &loadbalancer.Endpoint{
				Instance:          ep,
				Weight:            100,
				ActiveConnections: lab.conns[ep.ID],
				Metrics:           metrics,
			}
		}
	}

	if selected == nil && len(endpoints) > 0 {
		selected = &loadbalancer.Endpoint{
			Instance:          endpoints[0],
			Weight:            100,
			ActiveConnections: lab.conns[endpoints[0].ID],
			Metrics:           lab.getOrCreateMetricsLocked(endpoints[0].ID),
		}
	}

	return selected, nil
}

// RecordMetrics records metrics for an endpoint using EWMA
func (lab *LatencyAwareBalancer) RecordMetrics(instanceID string, latency time.Duration, success bool) {
	lab.mu.Lock()
	defer lab.mu.Unlock()

	metrics := lab.getOrCreateMetricsLocked(instanceID)

	metrics.TotalRequests++
	if success {
		metrics.SuccessCount++
	} else {
		metrics.FailureCount++
	}

	// Update min/max latency
	if latency < metrics.MinLatency || metrics.MinLatency == 0 {
		metrics.MinLatency = latency
	}
	if latency > metrics.MaxLatency {
		metrics.MaxLatency = latency
	}

	// Update exponential weighted moving average
	if metrics.AvgLatency == 0 {
		metrics.AvgLatency = latency
	} else {
		// EWMA = alpha * current + (1 - alpha) * previous
		ewma := float64(metrics.AvgLatency) * (1 - lab.alpha)
		ewma += float64(latency) * lab.alpha
		metrics.AvgLatency = time.Duration(ewma)
	}

	// Update percentiles (simplified - would use histogram in production)
	metrics.P50Latency = metrics.AvgLatency
	metrics.P95Latency = time.Duration(float64(metrics.MaxLatency) * 0.95)
	metrics.P99Latency = time.Duration(float64(metrics.MaxLatency) * 0.99)

	if metrics.TotalRequests > 0 {
		metrics.ErrorRate = float64(metrics.FailureCount) / float64(metrics.TotalRequests)
	}

	metrics.LastUpdate = time.Now()
}

// IncrementConnections increments the active connection count
func (lab *LatencyAwareBalancer) IncrementConnections(instanceID string) {
	lab.mu.Lock()
	defer lab.mu.Unlock()
	lab.conns[instanceID]++
}

// DecrementConnections decrements the active connection count
func (lab *LatencyAwareBalancer) DecrementConnections(instanceID string) {
	lab.mu.Lock()
	defer lab.mu.Unlock()
	if count, ok := lab.conns[instanceID]; ok && count > 0 {
		lab.conns[instanceID]--
	}
}

// Reset clears internal state
func (lab *LatencyAwareBalancer) Reset() {
	lab.mu.Lock()
	defer lab.mu.Unlock()

	lab.metrics = make(map[string]*loadbalancer.EndpointMetrics)
	lab.conns = make(map[string]int64)
}

// GetMetrics returns metrics for an endpoint
func (lab *LatencyAwareBalancer) GetMetrics(instanceID string) *loadbalancer.EndpointMetrics {
	lab.mu.RLock()
	defer lab.mu.RUnlock()
	return lab.metrics[instanceID]
}

// SetAlpha sets the EWMA alpha value (0.0-1.0)
func (lab *LatencyAwareBalancer) SetAlpha(alpha float64) {
	if alpha < 0.0 {
		alpha = 0.0
	}
	if alpha > 1.0 {
		alpha = 1.0
	}

	lab.mu.Lock()
	defer lab.mu.Unlock()
	lab.alpha = alpha
}

// Helper methods

func (lab *LatencyAwareBalancer) getOrCreateMetricsLocked(instanceID string) *loadbalancer.EndpointMetrics {
	if metrics, ok := lab.metrics[instanceID]; ok {
		return metrics
	}

	metrics := &loadbalancer.EndpointMetrics{
		LastUpdate: time.Now(),
	}
	lab.metrics[instanceID] = metrics
	return metrics
}
