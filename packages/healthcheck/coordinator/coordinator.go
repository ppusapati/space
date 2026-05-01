package coordinator

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/samavaya/packages/p9log"
	"p9e.in/samavaya/packages/healthcheck"
)

// Coordinator manages health checking for all services and instances.
//
// Logger field stores *p9log.Helper (not the raw p9log.Logger interface) so
// the level-methods Debug/Info/Warn/Error are available. The constructor
// still accepts p9log.Logger for caller flexibility. See roadmap task B.1.
type Coordinator struct {
	pool           *pgxpool.Pool
	logger         *p9log.Helper
	options        healthcheck.Options
	checks         map[string]map[string]healthcheck.Checker // service -> instance -> checker
	health         map[string]*healthcheck.ServiceHealth     // service -> health
	mu             sync.RWMutex
	stopChan       chan struct{}
	wg             sync.WaitGroup
	eventChan      chan *healthcheck.Event
	sequence       atomic.Int64
	serviceUpdated map[string]time.Time // Track last update per service
}

// New creates a new health check coordinator
func New(pool *pgxpool.Pool, logger p9log.Logger, opts ...healthcheck.Option) *Coordinator {
	options := healthcheck.DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &Coordinator{
		pool:           pool,
		logger:         p9log.NewHelper(logger),
		options:        options,
		checks:         make(map[string]map[string]healthcheck.Checker),
		health:         make(map[string]*healthcheck.ServiceHealth),
		stopChan:       make(chan struct{}),
		eventChan:      make(chan *healthcheck.Event, 100),
		serviceUpdated: make(map[string]time.Time),
	}
}

// RegisterCheck registers a health check for a service instance
func (c *Coordinator) RegisterCheck(service, instance string, checker healthcheck.Checker) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.checks[service] == nil {
		c.checks[service] = make(map[string]healthcheck.Checker)
	}

	c.checks[service][instance] = checker

	// Initialize health
	if c.health[service] == nil {
		c.health[service] = &healthHealth{
			ServiceName: service,
			Instances:   make(map[string]*healthcheck.InstanceHealth),
			UpdatedAt:   time.Now(),
		}
	}

	if c.health[service].Instances == nil {
		c.health[service].Instances = make(map[string]*healthcheck.InstanceHealth)
	}

	c.health[service].Instances[instance] = &healthcheck.InstanceHealth{
		InstanceID:  instance,
		ServiceName: service,
		Status:      healthcheck.StatusUnknown,
		UpdatedAt:   time.Now(),
	}

	c.logger.Info("registered health check",
		"service", service,
		"instance", instance,
		"check_type", checker.Type(),
	)
}

// Start begins the background health checking loop
func (c *Coordinator) Start(ctx context.Context) error {
	if !c.options.Enabled {
		c.logger.Info("health checking is disabled")
		return nil
	}

	c.logger.Info("starting health check coordinator")

	// Start check workers for each service
	c.mu.RLock()
	services := make([]string, 0, len(c.checks))
	for service := range c.checks {
		services = append(services, service)
	}
	c.mu.RUnlock()

	for _, service := range services {
		c.startServiceChecker(ctx, service)
	}

	return nil
}

// Stop stops the health checking loop
func (c *Coordinator) Stop(ctx context.Context) error {
	c.logger.Info("stopping health check coordinator")
	close(c.stopChan)
	c.wg.Wait()
	close(c.eventChan)
	return nil
}

// GetStatus returns current health status of a service
func (c *Coordinator) GetStatus(ctx context.Context, service string) (*healthcheck.ServiceHealth, error) {
	c.mu.RLock()
	health, ok := c.health[service]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("service not found: %s", service)
	}

	return health, nil
}

// GetInstanceStatus returns health status of a specific instance
func (c *Coordinator) GetInstanceStatus(ctx context.Context, service, instance string) (*healthcheck.InstanceHealth, error) {
	c.mu.RLock()
	health, ok := c.health[service]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("service not found: %s", service)
	}

	if health.Instances == nil {
		return nil, fmt.Errorf("no instances found for service: %s", service)
	}

	instance_health, ok := health.Instances[instance]
	if !ok {
		return nil, fmt.Errorf("instance not found: %s", instance)
	}

	return instance_health, nil
}

// GetSummary returns overall health check summary
func (c *Coordinator) GetSummary(ctx context.Context) *healthcheck.Summary {
	c.mu.RLock()
	defer c.mu.RUnlock()

	summary := &healthcheck.Summary{
		GeneratedAt: time.Now(),
	}

	for _, svc := range c.health {
		summary.TotalServices++
		if svc.Status == healthcheck.StatusHealthy {
			summary.HealthyServices++
		} else {
			summary.UnhealthyServices++
		}

		summary.TotalInstances += svc.TotalInstances
		summary.HealthyInstances += svc.HealthyInstances
		summary.UnhealthyInstances += svc.UnhealthyInstances
	}

	if summary.TotalInstances > 0 {
		summary.HealthPercent = (summary.HealthyInstances * 100) / summary.TotalInstances
	}

	return summary
}

// Events returns a channel for health check events
func (c *Coordinator) Events() <-chan *healthcheck.Event {
	return c.eventChan
}

// Helper methods

// startServiceChecker starts background checker for a service
func (c *Coordinator) startServiceChecker(ctx context.Context, service string) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		ticker := time.NewTicker(c.options.Config.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-c.stopChan:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.runServiceChecks(ctx, service)

				// Add jitter to prevent thundering herd
				jitter := time.Duration(rand.Int63n(int64(c.options.Config.IntervalJitter)))
				time.Sleep(jitter)
			}
		}
	}()
}

// runServiceChecks runs all checks for a service
func (c *Coordinator) runServiceChecks(ctx context.Context, service string) {
	c.mu.RLock()
	instances := make(map[string]healthcheck.Checker)
	for inst, checker := range c.checks[service] {
		instances[inst] = checker
	}
	c.mu.RUnlock()

	// Run checks in parallel
	var wg sync.WaitGroup
	results := make(map[string]*healthcheck.CheckResult)
	resultsMu := sync.Mutex{}

	for instance, checker := range instances {
		wg.Add(1)
		go func(inst string, chk healthcheck.Checker) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(ctx, c.options.Config.Timeout)
			defer cancel()

			result, err := chk.Check(ctx)
			if err != nil {
				c.logger.Error("health check failed",
					"service", service,
					"instance", inst,
					"check", chk.Name(),
					"error", err,
				)

				if c.options.OnCheckFailure != nil {
					c.options.OnCheckFailure(ctx, service, inst, err)
				}
			}

			result.Sequence = c.sequence.Add(1)

			resultsMu.Lock()
			results[inst] = result
			resultsMu.Unlock()

			// Emit event
			c.emitEvent(&healthcheck.Event{
				Type:        "check_completed",
				ServiceName: service,
				InstanceID:  inst,
				CheckType:   chk.Type(),
				Result:      result,
				Timestamp:   time.Now(),
			})
		}(instance, checker)
	}

	wg.Wait()

	// Update health status
	c.updateServiceHealth(service, results)
}

// updateServiceHealth updates service health based on check results
func (c *Coordinator) updateServiceHealth(service string, results map[string]*healthcheck.CheckResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.health[service] == nil {
		return
	}

	svcHealth := c.health[service]
	healthyCount := 0
	unhealthyCount := 0

	for instance, result := range results {
		if svcHealth.Instances == nil {
			svcHealth.Instances = make(map[string]*healthcheck.InstanceHealth)
		}

		instHealth := svcHealth.Instances[instance]
		if instHealth == nil {
			instHealth = &healthcheck.InstanceHealth{
				InstanceID:  instance,
				ServiceName: service,
			}
			svcHealth.Instances[instance] = instHealth
		}

		oldStatus := instHealth.Status

		// Update based on result
		switch result.Status {
		case healthcheck.StatusHealthy:
			instHealth.SuccessCount++
			instHealth.FailureCount = 0
			if instHealth.SuccessCount >= c.options.Config.HealthyThreshold {
				instHealth.Status = healthcheck.StatusHealthy
			}
			instHealth.LastSuccessfulCheck = time.Now()

		case healthcheck.StatusUnhealthy:
			instHealth.FailureCount++
			instHealth.SuccessCount = 0
			if instHealth.FailureCount >= c.options.Config.UnhealthyThreshold {
				instHealth.Status = healthcheck.StatusUnhealthy
			}
			instHealth.LastFailedCheck = time.Now()
			instHealth.LastError = result.Error

		default:
			instHealth.Status = result.Status
		}

		instHealth.Details = result.Details
		instHealth.UpdatedAt = time.Now()

		// Notify on status change
		if oldStatus != instHealth.Status && c.options.OnStatusChange != nil {
			c.options.OnStatusChange(context.Background(), service, instance, oldStatus, instHealth.Status)
		}

		// Track counts
		if instHealth.Status == healthcheck.StatusHealthy {
			healthyCount++
		} else {
			unhealthyCount++
		}
	}

	// Update service health
	svcHealth.HealthyInstances = healthyCount
	svcHealth.UnhealthyInstances = unhealthyCount
	svcHealth.TotalInstances = len(results)

	if svcHealth.TotalInstances > 0 {
		svcHealth.HealthPercent = (healthyCount * 100) / svcHealth.TotalInstances
	}

	// Determine overall service status
	switch {
	case healthyCount == 0:
		svcHealth.Status = healthcheck.StatusUnhealthy
	case unhealthyCount > 0:
		svcHealth.Status = healthcheck.StatusDegraded
	default:
		svcHealth.Status = healthcheck.StatusHealthy
	}

	svcHealth.UpdatedAt = time.Now()
	c.serviceUpdated[service] = time.Now()
}

// emitEvent sends an event to the event channel
func (c *Coordinator) emitEvent(event *healthcheck.Event) {
	select {
	case c.eventChan <- event:
	case <-c.stopChan:
	}
}

// Stub for proper type alias
type healthHealth = healthcheck.ServiceHealth
