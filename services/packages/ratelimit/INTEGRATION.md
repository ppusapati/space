# Rate Limiter Integration Guide

## Overview

The rate limiter package provides multiple algorithms for controlling request flow across 125+ microservices.

## Integration Points

### 1. Service Mesh Integration

```go
import "p9e.in/samavaya/packages/packages/mesh"
import "p9e.in/samavaya/packages/packages/ratelimit/backend"

// In mesh initialization
limiter := backend.NewPostgresRateLimiter(pool, logger)

// Before routing request
key := fmt.Sprintf("%s:%s", serviceName, clientID)
allowed, err := limiter.Allow(ctx, key)
if !allowed {
    return fmt.Errorf("rate limit exceeded for %s", serviceName)
}

// After response
if err != nil {
    limiter.RecordFailure(ctx, key)
}
```

### 2. Load Balancer Integration

```go
import "p9e.in/samavaya/packages/packages/loadbalancer"
import "p9e.in/samavaya/packages/packages/ratelimit/algorithms"

// Create load balancer
lb := algorithms.NewRoundRobinBalancer()

// When recording metrics
endpoint, _ := lb.Select(ctx, instances)
lb.RecordMetrics(endpoint.Instance.ID, latency, success)
```

### 3. Circuit Breaker Integration

```go
import "p9e.in/samavaya/packages/packages/circuitbreaker"
import "p9e.in/samavaya/packages/packages/ratelimit"

// When circuit breaker is OPEN
result, _ := breaker.Check(ctx, "service:op")
if !result.Allowed {
    // Also apply rate limiting backpressure
    rateLimiter.SetLimit(ctx, key, newLimit/2)
}
```

### 4. HTTP Middleware Integration

```go
import "p9e.in/samavaya/packages/packages/ratelimit/api"

// Create API handler
limiterAPI := api.New(limiter, logger)

// Add to HTTP router
http.Handle("/v1/ratelimit/", limiterAPI)

// Use in middleware
func RateLimitMiddleware(limiter ratelimit.Limiter) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        clientID := getClientID(r) // From auth header
        serviceName := "my-service"
        key := fmt.Sprintf("%s:%s", serviceName, clientID)

        allowed, _ := limiter.Allow(r.Context(), key)
        if !allowed {
            w.Header().Set("Retry-After", "1")
            w.WriteHeader(http.StatusTooManyRequests)
            return
        }

        // Continue processing
    }
}
```

## Database Schema

The rate limiter uses the `rate_limits` table created in Phase 1:

```sql
CREATE TABLE rate_limits (
    key VARCHAR(255) PRIMARY KEY,
    service_name VARCHAR(255) NOT NULL,
    client_id VARCHAR(255),
    allowed_requests BIGINT DEFAULT 0,
    rejected_requests BIGINT DEFAULT 0,
    limit_per_second INT DEFAULT 100,
    window_start TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## Algorithms

### Token Bucket

Simple and predictable:
```go
limiter := algorithms.NewTokenBucketLimiter(
    100,  // Capacity
    50.0, // Refill rate (tokens per second)
)
```

**Use when:**
- Predictable traffic patterns
- Need simple burst handling
- Per-API key limits

### BBR (Bottleneck Bandwidth and RTT)

Adaptive and efficient:
```go
bbr := algorithms.NewBBRLimiter()

// Record actual latencies
bbr.RecordRTT(latency)

// Track deliveries
bbr.RecordDelivery(numPackets, latency)
```

**Use when:**
- High-throughput services
- Variable latency
- Want automatic congestion control

### Distributed (PostgreSQL)

Multi-instance coordination:
```go
limiter := backend.NewPostgresRateLimiter(pool, logger)

// Automatically syncs every 10 seconds
// Uses advisory locks for coordination
```

**Use when:**
- Multiple instances of same service
- Need consistent rate limiting
- Can tolerate eventual consistency

## Per-Service vs Per-Client

### Per-Service Limit

```go
key := fmt.Sprintf("payment-service")
allowed, _ := limiter.Allow(ctx, key)
```

Limits total requests to a service: 100 req/sec

### Per-Client Limit

```go
clientID := extractClientID(r.Header.Get("Authorization"))
key := fmt.Sprintf("payment-service:%s", clientID)
allowed, _ := limiter.Allow(ctx, key)
```

Limits per API key/user: 10 req/sec per client

### Per-Endpoint Limit

```go
endpoint := r.URL.Path
key := fmt.Sprintf("payment-service:%s:%s", clientID, endpoint)
allowed, _ := limiter.Allow(ctx, key)
```

Different limits for different operations

## Monitoring

### Via HTTP API

```bash
# Check stats
curl -X GET http://localhost:8080/v1/ratelimit/stats/payment-service

# Set new limit
curl -X PUT http://localhost:8080/v1/ratelimit/payment-service \
  -d '{"limit_per_second": 200}'

# Check if request allowed
curl -X POST http://localhost:8080/v1/ratelimit/check/payment-service:client-1 \
  -d '{"tokens": 1}'

# Reset limiter
curl -X POST http://localhost:8080/v1/ratelimit/reset/payment-service
```

### Via Code

```go
stats, _ := limiter.GetStats(ctx, "payment-service")
fmt.Printf("Allowed: %d, Rejected: %d, Limit: %d\n",
    stats.AllowedCount, stats.RejectedCount, stats.CurrentLimit)
```

## Error Handling

```go
allowed, err := limiter.Allow(ctx, key)
if err != nil {
    // Database/infrastructure error
    logger.Error("rate limiter error", "error", err)
    // Fail open or closed based on policy
}

if !allowed {
    // Rate limit hit - normal backpressure
    return ErrRateLimitExceeded
}
```

## Configuration

### Per-Service Limits

In routing policy:

```json
{
  "service_name": "payment-service",
  "rate_limit": {
    "algorithm": "token_bucket",
    "limit_per_second": 1000,
    "burst_capacity": 2000
  }
}
```

### Adaptive Limits

```go
config := ratelimit.RateLimitConfig{
    Algorithm:       ratelimit.AlgorithmAdaptive,
    DefaultLimit:    100,
    EnableAdaptive:  true,
    AdaptiveHighLoad: 0.8,  // Reduce at 80% load
    AdaptiveLowLoad: 0.2,   // Increase at 20% load
}
```

## Testing

```bash
# Run tests
go test ./packages/ratelimit/...

# With coverage
go test -cover ./packages/ratelimit/...

# Benchmarks
go test -bench=. ./packages/ratelimit/algorithms/
```

## Performance

- **Token Bucket**: <1μs per Allow() call
- **BBR**: <10μs per Allow() call (with RTT tracking)
- **Distributed**: ~10ms per Allow() call (with DB round trip)

Use local caching with distributed limiter to achieve <100μs latency:
- Cache validity: 10 seconds (configurable)
- Sync interval: 10 seconds (configurable)
- Fallback: Local buffer when cache misses
