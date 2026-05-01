# Health Checking Integration Guide

## Overview

The health check package provides continuous health monitoring for services with automatic failure detection and remediation.

## Quick Start

```go
import (
    "p9e.in/samavaya/packages/packages/healthcheck/checks"
    "p9e.in/samavaya/packages/packages/healthcheck/coordinator"
)

// Create coordinator
coord := coordinator.New(pool, logger)

// Register HTTP check
coord.RegisterCheck("payment-service", "instance-1",
    checks.NewHTTPCheck(healthcheck.HTTPCheckConfig{
        URL:     "http://localhost:8080/health",
        Timeout: 5 * time.Second,
    }))

// Start background checking
go coord.Start(ctx)

// Get status
status, _ := coord.GetStatus(ctx, "payment-service")
fmt.Printf("Service health: %s\n", status.Status)
```

## Health Check Types

### HTTP Check

Sends GET request to health endpoint:

```go
check := checks.NewHTTPCheck(healthcheck.HTTPCheckConfig{
    URL:          "http://localhost:8080/health",
    Method:       "GET",
    Headers:      map[string]string{"Authorization": "Bearer token"},
    SuccessCodes: []int{200, 201},
    Timeout:      5 * time.Second,
})

result, _ := check.Check(ctx)
// Returns: HEALTHY | UNHEALTHY | DEGRADED
```

**Usage:**
- REST APIs
- Web services
- Microservice health endpoints

### TCP Check

Tests TCP port connectivity:

```go
check := checks.NewTCPCheck(healthcheck.TCPCheckConfig{
    Host:    "db.example.com",
    Port:    5432,
    Timeout: 3 * time.Second,
})

result, _ := check.Check(ctx)
```

**Usage:**
- Database connectivity
- Cache systems (Redis, Memcached)
- Message queues
- Generic port availability

### gRPC Check

Uses gRPC Health Check protocol:

```go
check := checks.NewGRPCCheck(healthcheck.GRPCCheckConfig{
    Host:    "localhost",
    Port:    50051,
    Service: "myservice.v1.MyService", // Empty for overall server health
    Timeout: 5 * time.Second,
})

result, _ := check.Check(ctx)
```

**Status values:**
- `SERVING`: Service is healthy
- `NOT_SERVING`: Service has issues
- `UNKNOWN`: Status unknown
- `SERVICE_UNKNOWN`: Service not found

**Usage:**
- gRPC services
- Service mesh integration
- Protocol buffers based systems

### Database Check

Tests database connectivity:

```go
check := checks.NewDatabaseCheck(healthcheck.DatabaseCheckConfig{
    DSN:     "postgres://user:pass@localhost/mydb",
    Query:   "SELECT 1",
    Timeout: 10 * time.Second,
})

result, _ := check.Check(ctx)
```

Or with existing pool:

```go
check := checks.NewDatabaseCheckWithPool(pool, "SELECT 1", 10*time.Second)
result, _ := check.Check(ctx)
```

**Usage:**
- Database health
- Connection pool monitoring
- Query latency checks

## Health States

```
HEALTHY   - Service is functioning normally
UNHEALTHY - Service has critical failures
DEGRADED  - Service is partially functional
UNKNOWN   - Check has not completed yet
```

## Failure Detection

Services transition to UNHEALTHY after:
- N consecutive check failures (default: 3)
- Repeated connection timeouts
- Bad status codes from health endpoint

Recovery requires:
- N consecutive successful checks (default: 2)
- Service to be responding normally

Example transition:

```
Healthy → Failed → Failed → Failed → UNHEALTHY
         (remove from LB)

UNHEALTHY → Success → Success → HEALTHY
                     (add back to LB)
```

## Configuration

### Default Configuration

```go
config := healthcheck.DefaultHealthCheckConfig()
// Interval: 10 seconds
// Timeout: 5 seconds
// UnhealthyThreshold: 3 consecutive failures
// HealthyThreshold: 2 consecutive successes
// IntervalJitter: 1 second
```

### Custom Configuration

```go
coordinator := coordinator.New(pool, logger,
    healthcheck.WithInterval(30 * time.Second),
    healthcheck.WithTimeout(10 * time.Second),
    healthcheck.WithUnhealthyThreshold(5),
)
```

## Coordinator

The coordinator manages all health checks and maintains service state:

```go
coord := coordinator.New(pool, logger)

// Register checks for multiple instances
coord.RegisterCheck("auth-service", "instance-1", authCheck1)
coord.RegisterCheck("auth-service", "instance-2", authCheck2)
coord.RegisterCheck("payment-service", "instance-1", paymentCheck)

// Start background monitoring
go coord.Start(ctx)

// Get service health
svcHealth, _ := coord.GetStatus(ctx, "auth-service")
fmt.Printf("Status: %s, Healthy: %d/%d\n",
    svcHealth.Status,
    svcHealth.HealthyInstances,
    svcHealth.TotalInstances)

// Get instance health
instHealth, _ := coord.GetInstanceStatus(ctx, "auth-service", "instance-1")
fmt.Printf("Instance: %s, Status: %s\n",
    instHealth.InstanceID,
    instHealth.Status)

// Get overall summary
summary := coord.GetSummary(ctx)
fmt.Printf("Health: %d%%\n", summary.HealthPercent)
```

## HTTP REST API

### Get Summary

```bash
GET /v1/health/summary

{
  "status": "healthy",
  "healthy_services": 5,
  "unhealthy_services": 0,
  "total_services": 5,
  "healthy_instances": 15,
  "unhealthy_instances": 0,
  "total_instances": 15,
  "health_percent": 100,
  "generated_at": "2026-03-01T12:00:00Z"
}
```

### Get Service Status

```bash
GET /v1/health/service/{service}

{
  "service_name": "payment-service",
  "status": "healthy",
  "healthy_instances": 3,
  "unhealthy_instances": 0,
  "total_instances": 3,
  "health_percent": 100,
  "instances": {
    "instance-1": {
      "instance_id": "instance-1",
      "status": "HEALTHY",
      "last_successful_check": "2026-03-01T12:00:00Z",
      "failure_count": 0,
      "success_count": 5,
      "updated_at": "2026-03-01T12:00:00Z"
    }
  },
  "updated_at": "2026-03-01T12:00:00Z"
}
```

### Get Instance Status

```bash
GET /v1/health/service/{service}/instance/{instance}

{
  "instance_id": "instance-1",
  "status": "HEALTHY",
  "last_successful_check": "2026-03-01T12:00:00Z",
  "last_failed_check": null,
  "failure_count": 0,
  "success_count": 5,
  "details": {
    "status_code": 200,
    "duration_ms": 45
  },
  "updated_at": "2026-03-01T12:00:00Z"
}
```

### Liveness Check

```bash
GET /v1/health/live

{
  "status": "alive",
  "time": "2026-03-01T12:00:00Z"
}
```

### Readiness Check

```bash
GET /v1/health/ready

{
  "ready": true,
  "healthy_services": 5,
  "total_services": 5
}

# Or if not ready (503)
{
  "ready": false,
  "reason": "no healthy services"
}
```

## Integration Examples

### With Service Registry

```go
// When health check fails, update registry
coordinator := coordinator.New(pool, logger,
    healthcheck.WithOnStatusChange(
        func(ctx context.Context, svc, inst string, old, new healthcheck.HealthStatus) {
            if new == healthcheck.StatusUnhealthy {
                // Mark instance as unhealthy in registry
                registry.UpdateHealth(ctx, svc, inst, false)
            } else if new == healthcheck.StatusHealthy {
                // Mark instance as healthy again
                registry.UpdateHealth(ctx, svc, inst, true)
            }
        }))
```

### With Load Balancer

```go
// Skip unhealthy instances during load balancing
lb := NewLoadBalancer()

for _, instance := range instances {
    health, _ := coordinator.GetInstanceStatus(ctx, service, instance.ID)
    if health.Status == healthcheck.StatusHealthy {
        lb.Add(instance)
    }
}
```

### With Circuit Breaker

```go
// Open circuit if health check fails
health, _ := coordinator.GetStatus(ctx, "downstream-service")
if health.Status == healthcheck.StatusUnhealthy {
    breaker.Open("downstream-service")
}
```

## Monitoring

### Events

```go
coordinator := coordinator.New(pool, logger)

// Subscribe to health events
go func() {
    for event := range coordinator.Events() {
        fmt.Printf("[%s] %s:%s -> %s\n",
            event.CheckType,
            event.ServiceName,
            event.InstanceID,
            event.Result.Status)
    }
}()
```

### Metrics

Track and alert on:
- Service health percentage
- Check latency
- Failure rates
- Time to recover

```go
summary := coordinator.GetSummary(ctx)
metrics.Record("health.services.healthy", float64(summary.HealthyServices))
metrics.Record("health.services.unhealthy", float64(summary.UnhealthyServices))
metrics.Record("health.instances.health_percent", float64(summary.HealthPercent))
```

## Best Practices

### 1. Check Frequency

```go
// Don't check too frequently (overhead)
// Don't check too infrequently (miss failures)

// Recommended defaults:
// - HTTP: 10 seconds
// - TCP: 10 seconds
// - gRPC: 15 seconds
// - Database: 30 seconds
```

### 2. Timeout Configuration

```go
// Timeout should be less than interval
// Typical: interval / 2

interval := 10 * time.Second
timeout := 5 * time.Second  // Half of interval
```

### 3. Failure Thresholds

```go
// Don't mark unhealthy too quickly (avoid flapping)
// But react reasonably fast to real failures

UnhealthyThreshold: 3,  // Mark unhealthy after 3 failures
HealthyThreshold: 2,    // Mark healthy after 2 successes
```

### 4. Cascading Checks

```go
// Check dependencies in order
// If primary fails, try secondary

if httpCheck.Check(ctx).Status == StatusHealthy {
    // Service is up via HTTP
} else if tcpCheck.Check(ctx).Status == StatusHealthy {
    // Service is still responding (degraded)
} else {
    // Service is completely down
}
```

## Troubleshooting

### Flapping Health Status

Problem: Service repeatedly switches between HEALTHY/UNHEALTHY

Solutions:
```go
// Increase thresholds
healthcheck.WithUnhealthyThreshold(5),
healthcheck.WithHealthyThreshold(3),

// Increase check interval
healthcheck.WithInterval(30 * time.Second),

// Increase timeout
healthcheck.WithTimeout(10 * time.Second),
```

### High Check Latency

Problem: Health checks are slow

Solutions:
```go
// Use TCP checks instead of HTTP (faster)
// Use shorter timeouts (fail fast)
// Check fewer services

// Monitor check latency
for event := range coordinator.Events() {
    if event.Result.Duration > 1*time.Second {
        logger.Warn("slow health check", "duration", event.Result.Duration)
    }
}
```

### Too Many Connections

Problem: Health checks open too many connections

Solutions:
```go
// Increase check interval
// Use connection pooling (reuse connections)
// Reduce number of services being checked
```
