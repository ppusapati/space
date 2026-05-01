package algorithms

import (
	"context"
	"testing"
	"time"

	"p9e.in/samavaya/packages/loadbalancer"
	"p9e.in/samavaya/packages/registry"
)

func TestRoundRobinSelection(t *testing.T) {
	lb := NewRoundRobinBalancer()

	// Create test instances
	instances := []*registry.ServiceInstance{
		{ID: "inst-1", ServiceName: "test", Host: "host1", Port: 5001},
		{ID: "inst-2", ServiceName: "test", Host: "host2", Port: 5002},
		{ID: "inst-3", ServiceName: "test", Host: "host3", Port: 5003},
	}

	// Select should cycle through instances
	ctx := context.Background()

	ep1, _ := lb.Select(ctx, instances)
	if ep1.Instance.ID != "inst-1" {
		t.Errorf("Expected inst-1, got %s", ep1.Instance.ID)
	}

	ep2, _ := lb.Select(ctx, instances)
	if ep2.Instance.ID != "inst-2" {
		t.Errorf("Expected inst-2, got %s", ep2.Instance.ID)
	}

	ep3, _ := lb.Select(ctx, instances)
	if ep3.Instance.ID != "inst-3" {
		t.Errorf("Expected inst-3, got %s", ep3.Instance.ID)
	}

	// Should wrap around
	ep4, _ := lb.Select(ctx, instances)
	if ep4.Instance.ID != "inst-1" {
		t.Errorf("Expected inst-1 (wrap), got %s", ep4.Instance.ID)
	}
}

func TestRoundRobinMetrics(t *testing.T) {
	lb := NewRoundRobinBalancer()

	// Record some metrics
	lb.RecordMetrics("inst-1", 100*time.Millisecond, true)
	lb.RecordMetrics("inst-1", 150*time.Millisecond, true)
	lb.RecordMetrics("inst-1", 200*time.Millisecond, false)

	metrics := lb.GetMetrics("inst-1")
	if metrics.SuccessCount != 2 {
		t.Errorf("Expected 2 successes, got %d", metrics.SuccessCount)
	}
	if metrics.FailureCount != 1 {
		t.Errorf("Expected 1 failure, got %d", metrics.FailureCount)
	}
}

func TestRoundRobinConnections(t *testing.T) {
	lb := NewRoundRobinBalancer()

	lb.IncrementConnections("inst-1")
	lb.IncrementConnections("inst-1")

	// Need to get connections through select
	instances := []*registry.ServiceInstance{
		{ID: "inst-1", ServiceName: "test", Host: "host1", Port: 5001},
	}

	ctx := context.Background()
	ep, _ := lb.Select(ctx, instances)
	if ep.ActiveConnections != 2 {
		t.Errorf("Expected 2 active connections, got %d", ep.ActiveConnections)
	}

	lb.DecrementConnections("inst-1")
	ep, _ = lb.Select(ctx, instances)
	if ep.ActiveConnections != 1 {
		t.Errorf("Expected 1 active connection, got %d", ep.ActiveConnections)
	}
}

func TestLeastConnectionsSelection(t *testing.T) {
	lb := NewLeastConnectionsBalancer()

	instances := []*registry.ServiceInstance{
		{ID: "inst-1", ServiceName: "test", Host: "host1", Port: 5001},
		{ID: "inst-2", ServiceName: "test", Host: "host2", Port: 5002},
		{ID: "inst-3", ServiceName: "test", Host: "host3", Port: 5003},
	}

	// Add connections to simulate load
	lb.IncrementConnections("inst-1")
	lb.IncrementConnections("inst-1")
	lb.IncrementConnections("inst-2")
	// inst-3 has 0 connections

	ctx := context.Background()
	ep, _ := lb.Select(ctx, instances)
	if ep.Instance.ID != "inst-3" {
		t.Errorf("Expected inst-3 (least connections), got %s", ep.Instance.ID)
	}
}

func TestLatencyAwareSelection(t *testing.T) {
	lb := NewLatencyAwareBalancer()

	instances := []*registry.ServiceInstance{
		{ID: "inst-1", ServiceName: "test", Host: "host1", Port: 5001},
		{ID: "inst-2", ServiceName: "test", Host: "host2", Port: 5002},
	}

	// Record latencies
	lb.RecordMetrics("inst-1", 200*time.Millisecond, true)
	lb.RecordMetrics("inst-1", 250*time.Millisecond, true)

	lb.RecordMetrics("inst-2", 100*time.Millisecond, true)
	lb.RecordMetrics("inst-2", 150*time.Millisecond, true)

	ctx := context.Background()
	ep, _ := lb.Select(ctx, instances)
	if ep.Instance.ID != "inst-2" {
		t.Errorf("Expected inst-2 (lower latency), got %s", ep.Instance.ID)
	}
}

func TestWeightedRoundRobinSelection(t *testing.T) {
	weights := map[string]int{
		"inst-1": 70, // 70% traffic
		"inst-2": 30, // 30% traffic
	}
	lb := NewWeightedRoundRobinBalancer(weights)

	instances := []*registry.ServiceInstance{
		{ID: "inst-1", ServiceName: "test", Host: "host1", Port: 5001},
		{ID: "inst-2", ServiceName: "test", Host: "host2", Port: 5002},
	}

	// Count selections over many iterations
	ctx := context.Background()
	inst1Count := 0
	inst2Count := 0
	iterations := 1000

	for i := 0; i < iterations; i++ {
		ep, _ := lb.Select(ctx, instances)
		if ep.Instance.ID == "inst-1" {
			inst1Count++
		} else {
			inst2Count++
		}
	}

	// Should be approximately 70/30
	expectedInst1 := iterations * 70 / 100
	expectedInst2 := iterations * 30 / 100

	// Allow 10% tolerance
	tolerance := iterations / 10

	if inst1Count < expectedInst1-tolerance || inst1Count > expectedInst1+tolerance {
		t.Errorf("Expected ~%d selections for inst-1, got %d", expectedInst1, inst1Count)
	}

	if inst2Count < expectedInst2-tolerance || inst2Count > expectedInst2+tolerance {
		t.Errorf("Expected ~%d selections for inst-2, got %d", expectedInst2, inst2Count)
	}
}

func BenchmarkRoundRobinSelect(b *testing.B) {
	lb := NewRoundRobinBalancer()
	instances := []*registry.ServiceInstance{
		{ID: "inst-1", ServiceName: "test", Host: "host1", Port: 5001},
		{ID: "inst-2", ServiceName: "test", Host: "host2", Port: 5002},
		{ID: "inst-3", ServiceName: "test", Host: "host3", Port: 5003},
	}

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lb.Select(ctx, instances)
	}
}

func BenchmarkLatencyAwareSelect(b *testing.B) {
	lb := NewLatencyAwareBalancer()
	instances := []*registry.ServiceInstance{
		{ID: "inst-1", ServiceName: "test", Host: "host1", Port: 5001},
		{ID: "inst-2", ServiceName: "test", Host: "host2", Port: 5002},
		{ID: "inst-3", ServiceName: "test", Host: "host3", Port: 5003},
	}

	// Pre-record some metrics
	for i := 0; i < 100; i++ {
		lb.RecordMetrics("inst-1", 100*time.Millisecond, true)
		lb.RecordMetrics("inst-2", 50*time.Millisecond, true)
		lb.RecordMetrics("inst-3", 150*time.Millisecond, true)
	}

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lb.Select(ctx, instances)
	}
}
