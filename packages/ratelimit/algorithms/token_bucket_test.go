package algorithms

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucketAllow(t *testing.T) {
	// 100 capacity, 50 req/sec
	limiter := NewTokenBucketLimiter(100, 50.0)

	ctx := context.Background()

	// Should allow initial requests up to capacity
	for i := 0; i < 100; i++ {
		ok, _ := limiter.Allow(ctx, "test-key")
		if !ok {
			t.Fatalf("Request %d should be allowed", i)
		}
	}

	// Next request should be rejected
	ok, _ := limiter.Allow(ctx, "test-key")
	if ok {
		t.Fatal("Request should be rejected after capacity exhausted")
	}
}

func TestTokenBucketRefill(t *testing.T) {
	// 10 capacity, 5 req/sec
	limiter := NewTokenBucketLimiter(10, 5.0)

	ctx := context.Background()

	// Exhaust bucket
	for i := 0; i < 10; i++ {
		limiter.Allow(ctx, "test-key")
	}

	// Should be rejected
	ok, _ := limiter.Allow(ctx, "test-key")
	if ok {
		t.Fatal("Request should be rejected")
	}

	// Wait for refill (1 second = 5 tokens at 5 req/sec)
	time.Sleep(1100 * time.Millisecond)

	// Should now have ~5 tokens
	allowCount := 0
	for i := 0; i < 10; i++ {
		ok, _ := limiter.Allow(ctx, "test-key")
		if ok {
			allowCount++
		} else {
			break
		}
	}

	if allowCount < 4 || allowCount > 6 {
		t.Errorf("Expected ~5 requests allowed, got %d", allowCount)
	}
}

func TestTokenBucketAllowN(t *testing.T) {
	limiter := NewTokenBucketLimiter(100, 50.0)

	ctx := context.Background()

	// Request 50 tokens at once
	ok, _ := limiter.AllowN(ctx, "test-key", 50)
	if !ok {
		t.Fatal("Should allow 50 tokens")
	}

	// Request another 50
	ok, _ = limiter.AllowN(ctx, "test-key", 50)
	if !ok {
		t.Fatal("Should allow another 50 tokens")
	}

	// Next 50 should fail
	ok, _ = limiter.AllowN(ctx, "test-key", 50)
	if ok {
		t.Fatal("Should not allow 50 tokens")
	}
}

func TestTokenBucketReserve(t *testing.T) {
	limiter := NewTokenBucketLimiter(10, 5.0)

	ctx := context.Background()

	// Exhaust bucket
	for i := 0; i < 10; i++ {
		limiter.Allow(ctx, "test-key")
	}

	// Reserve should indicate delay
	res, _ := limiter.Reserve(ctx, "test-key")
	if res.OK {
		t.Fatal("Reserve should indicate not OK after capacity exhausted")
	}

	if res.Delay == 0 {
		t.Fatal("Reserve should indicate delay needed")
	}
}

func TestTokenBucketMultipleKeys(t *testing.T) {
	limiter := NewTokenBucketLimiter(10, 5.0)

	ctx := context.Background()

	// Exhaust key1
	for i := 0; i < 10; i++ {
		limiter.Allow(ctx, "key1")
	}

	// key2 should still have capacity
	ok, _ := limiter.Allow(ctx, "key2")
	if !ok {
		t.Fatal("key2 should have capacity")
	}
}

func BenchmarkTokenBucketAllow(b *testing.B) {
	limiter := NewTokenBucketLimiter(1000, 1000.0)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(ctx, "bench-key")
	}
}

func BenchmarkTokenBucketAllowN(b *testing.B) {
	limiter := NewTokenBucketLimiter(10000, 10000.0)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.AllowN(ctx, "bench-key", 100)
	}
}

func TestBBRAllow(t *testing.T) {
	bbr := NewBBRLimiter()

	ctx := context.Background()

	// Should allow initial requests
	ok, _ := bbr.Allow(ctx, "test-key")
	if !ok {
		t.Fatal("Should allow initial request")
	}

	// Should track inflight
	stats, _ := bbr.GetStats(ctx, "test-key")
	if stats.Metrics["inflight"] != int64(1) {
		t.Errorf("Expected inflight=1, got %v", stats.Metrics["inflight"])
	}
}

func TestBBRRecordRTT(t *testing.T) {
	bbr := NewBBRLimiter()

	// Record RTT
	bbr.RecordRTT(10 * time.Millisecond)

	stats, _ := bbr.GetStats(context.Background(), "test")
	rttMs := stats.Metrics["rtt_ms"].(int64)

	if rttMs < 9 || rttMs > 11 {
		t.Errorf("Expected RTT ~10ms, got %dms", rttMs)
	}
}

func TestBBRStateTransition(t *testing.T) {
	bbr := NewBBRLimiter()

	ctx := context.Background()

	// Start in STARTUP
	stats, _ := bbr.GetStats(ctx, "test")
	if stats.Metrics["state"] != "STARTUP" {
		t.Errorf("Expected STARTUP state, got %v", stats.Metrics["state"])
	}

	// Record deliveries to trigger state transitions
	for i := 0; i < 50; i++ {
		bbr.RecordDelivery(10, 10*time.Millisecond)
	}

	stats, _ = bbr.GetStats(ctx, "test")
	state := stats.Metrics["state"]
	t.Logf("State after deliveries: %v", state)
	// Should transition through states
}

func BenchmarkBBRAllow(b *testing.B) {
	bbr := NewBBRLimiter()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bbr.Allow(ctx, "bench-key")
	}
}
