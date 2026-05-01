-- Service Registry Schema
-- This schema defines the tables for the dynamic service discovery system

-- Main service registry table
CREATE TABLE IF NOT EXISTS service_registry (
    -- Unique identifier for this instance
    instance_id UUID PRIMARY KEY,

    -- Service name (e.g., "user-service", "payment-service")
    service_name VARCHAR(255) NOT NULL,

    -- Host/IP address of the instance
    hostname VARCHAR(255) NOT NULL,

    -- Port number where service is listening
    port INT NOT NULL CHECK (port > 0 AND port <= 65535),

    -- Current health status: UP, DOWN, WARNING, UNKNOWN
    health_status VARCHAR(50) DEFAULT 'UP',

    -- Custom metadata stored as JSONB (region, zone, version, etc.)
    metadata JSONB DEFAULT '{}',

    -- Whether this is an external service (partner API, 3rd-party)
    is_external BOOLEAN DEFAULT FALSE,

    -- External URL (only populated if is_external is true)
    external_url TEXT,

    -- Service version
    version VARCHAR(100),

    -- Region (e.g., "us-east-1", "eu-west-1")
    region VARCHAR(100),

    -- Zone/Availability zone (e.g., "us-east-1a")
    zone VARCHAR(100),

    -- Last heartbeat timestamp - used for detecting stale instances
    last_heartbeat TIMESTAMP DEFAULT NOW(),

    -- When the instance was registered
    registered_at TIMESTAMP DEFAULT NOW(),

    -- When the instance was created in the registry
    created_at TIMESTAMP DEFAULT NOW(),

    -- When the instance was last updated
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Index for efficient queries by service name and health status
CREATE INDEX IF NOT EXISTS idx_service_registry_service_health
    ON service_registry(service_name, health_status);

-- Index for heartbeat-based cleanup queries
CREATE INDEX IF NOT EXISTS idx_service_registry_heartbeat
    ON service_registry(last_heartbeat);

-- Index for service name queries
CREATE INDEX IF NOT EXISTS idx_service_registry_service_name
    ON service_registry(service_name);

-- Index for health status queries
CREATE INDEX IF NOT EXISTS idx_service_registry_health_status
    ON service_registry(health_status);

-- Index for region/zone filtering (canary deployments)
CREATE INDEX IF NOT EXISTS idx_service_registry_region_zone
    ON service_registry(region, zone);

-- Index for external services
CREATE INDEX IF NOT EXISTS idx_service_registry_is_external
    ON service_registry(is_external);

-- Health check configuration table
-- This table stores how to health-check each service
CREATE TABLE IF NOT EXISTS health_check_configs (
    service_name VARCHAR(255) PRIMARY KEY,

    -- Health check type: http, tcp, grpc, database, custom
    check_type VARCHAR(50) NOT NULL,

    -- Endpoint/host to check
    endpoint VARCHAR(255),

    -- Port number for health check
    port INT,

    -- Interval between health checks (in seconds)
    check_interval_seconds INT DEFAULT 15,

    -- Timeout for health check (in seconds)
    timeout_seconds INT DEFAULT 5,

    -- Number of consecutive failures before marking unhealthy
    failure_threshold INT DEFAULT 3,

    -- Number of consecutive successes before marking healthy
    success_threshold INT DEFAULT 2,

    -- HTTP-specific: path to check
    path TEXT,

    -- HTTP-specific: HTTP method (GET, POST, etc.)
    method VARCHAR(10),

    -- gRPC-specific: service name for health check
    grpc_service VARCHAR(255),

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Routing policies table
-- This table stores routing policies for each service (for service mesh)
CREATE TABLE IF NOT EXISTS routing_policies (
    service_name VARCHAR(255) PRIMARY KEY,

    -- Version constraint (optional, e.g., "v1.2.x")
    version_constraint VARCHAR(100),

    -- Region preference (optional)
    region VARCHAR(100),

    -- Load balancing algorithm: round_robin, least_conn, weighted, latency, random
    algorithm VARCHAR(50) DEFAULT 'round_robin',

    -- Circuit breaker configuration (JSON)
    circuit_breaker_config JSONB,

    -- Retry policy configuration (JSON)
    retry_config JSONB,

    -- Timeout policy configuration (JSON)
    timeout_config JSONB,

    -- Canary deployment configuration (JSON)
    canary_config JSONB,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Rate limiting tracking table
-- This table tracks rate limit state for distributed rate limiting
CREATE TABLE IF NOT EXISTS rate_limits (
    -- Composite key: service_name + client_id
    key VARCHAR(255) PRIMARY KEY,

    -- Service name
    service_name VARCHAR(255) NOT NULL,

    -- Optional client ID (for per-client limits)
    client_id VARCHAR(255),

    -- Number of allowed requests in current window
    allowed_requests BIGINT DEFAULT 0,

    -- Number of rejected requests in current window
    rejected_requests BIGINT DEFAULT 0,

    -- Current limit (requests per second)
    limit_per_second INT DEFAULT 100,

    -- Start of current rate limit window
    window_start TIMESTAMP DEFAULT NOW(),

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Index for efficient lookups by service and client
CREATE INDEX IF NOT EXISTS idx_rate_limits_service_client
    ON rate_limits(service_name, client_id);

-- Metrics table for observability
-- This table stores per-service metrics
CREATE TABLE IF NOT EXISTS service_metrics (
    service_name VARCHAR(255) PRIMARY KEY,

    -- Total number of requests
    request_count BIGINT DEFAULT 0,

    -- Total number of errors
    error_count BIGINT DEFAULT 0,

    -- Average latency in milliseconds
    avg_latency_ms BIGINT,

    -- 50th percentile latency
    p50_latency_ms BIGINT,

    -- 95th percentile latency
    p95_latency_ms BIGINT,

    -- 99th percentile latency
    p99_latency_ms BIGINT,

    -- Number of rate-limited requests
    rate_limited_count BIGINT DEFAULT 0,

    -- Number of circuit breaker trips
    circuit_breaker_trips BIGINT DEFAULT 0,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Service dependencies table
-- This table tracks the dependency graph between services
CREATE TABLE IF NOT EXISTS service_dependencies (
    -- Source service
    from_service VARCHAR(255) NOT NULL,

    -- Target service
    to_service VARCHAR(255) NOT NULL,

    -- Number of calls from source to target
    call_count BIGINT DEFAULT 0,

    -- Number of errors in calls
    error_count BIGINT DEFAULT 0,

    -- Average latency for calls
    avg_latency_ms BIGINT,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (from_service, to_service)
);

-- Request tracing table
-- This table stores trace data for request tracing across services
CREATE TABLE IF NOT EXISTS request_traces (
    -- Unique trace ID
    trace_id VARCHAR(255) PRIMARY KEY,

    -- Parent trace ID (for distributed tracing)
    parent_trace_id VARCHAR(255),

    -- Service that generated this span
    service_name VARCHAR(255) NOT NULL,

    -- Method/operation name
    method VARCHAR(255),

    -- HTTP status code or gRPC status code
    status_code INT,

    -- Duration in milliseconds
    duration_ms BIGINT,

    -- Detailed spans data (JSON)
    spans JSONB,

    created_at TIMESTAMP DEFAULT NOW()
);

-- Index for efficient trace lookups
CREATE INDEX IF NOT EXISTS idx_request_traces_service_time
    ON request_traces(service_name, created_at DESC);

-- Index for distributed tracing (by parent trace ID)
CREATE INDEX IF NOT EXISTS idx_request_traces_parent_id
    ON request_traces(parent_trace_id);

-- Create a function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for automatic timestamp updates
CREATE TRIGGER trigger_service_registry_updated_at
    BEFORE UPDATE ON service_registry
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_health_check_configs_updated_at
    BEFORE UPDATE ON health_check_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_routing_policies_updated_at
    BEFORE UPDATE ON routing_policies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_service_metrics_updated_at
    BEFORE UPDATE ON service_metrics
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_service_dependencies_updated_at
    BEFORE UPDATE ON service_dependencies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
