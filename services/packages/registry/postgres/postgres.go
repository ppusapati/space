package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"p9e.in/samavaya/packages/p9log"
	"p9e.in/samavaya/packages/registry"
)

// PostgresRegistry implements registry.RegistryBackend using PostgreSQL.
// `logger` is *p9log.Helper (B.1 sweep).
type PostgresRegistry struct {
	pool   *pgxpool.Pool
	logger *p9log.Helper
}

// New creates a new PostgreSQL-backed registry
func New(pool *pgxpool.Pool, logger p9log.Logger) *PostgresRegistry {
	return &PostgresRegistry{
		pool:   pool,
		logger: p9log.NewHelper(logger),
	}
}

// Register stores a service instance
func (pr *PostgresRegistry) Register(ctx context.Context, instance *registry.ServiceInstance) error {
	const query = `
		INSERT INTO service_registry (
			instance_id, service_name, hostname, port, health_status, metadata,
			is_external, external_url, version, region, zone, last_heartbeat,
			registered_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
		ON CONFLICT (instance_id) DO UPDATE SET
			health_status = $5,
			metadata = $6,
			external_url = $8,
			version = $9,
			region = $10,
			zone = $11,
			last_heartbeat = $12,
			updated_at = $15
	`

	metadataJSON, err := json.Marshal(instance.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = pr.pool.Exec(ctx, query,
		instance.ID,
		instance.ServiceName,
		instance.Host,
		instance.Port,
		instance.Health,
		metadataJSON,
		instance.IsExternal,
		instance.ExternalURL,
		instance.Version,
		instance.Region,
		instance.Zone,
		instance.LastHeartbeat,
		instance.RegisteredAt,
		instance.CreatedAt,
		instance.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to register instance: %w", err)
	}

	return nil
}

// Deregister removes a service instance
func (pr *PostgresRegistry) Deregister(ctx context.Context, instanceID string) error {
	const query = `DELETE FROM service_registry WHERE instance_id = $1`

	result, err := pr.pool.Exec(ctx, query, instanceID)
	if err != nil {
		return fmt.Errorf("failed to deregister instance: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("instance not found: %s", instanceID)
	}

	return nil
}

// GetInstance retrieves a specific instance by ID
func (pr *PostgresRegistry) GetInstance(ctx context.Context, instanceID string) (*registry.ServiceInstance, error) {
	const query = `
		SELECT instance_id, service_name, hostname, port, health_status, metadata,
		       is_external, external_url, version, region, zone, last_heartbeat,
		       registered_at, created_at, updated_at
		FROM service_registry
		WHERE instance_id = $1
	`

	return pr.scanInstance(ctx, query, instanceID)
}

// GetInstances retrieves all healthy instances of a service
func (pr *PostgresRegistry) GetInstances(ctx context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	const query = `
		SELECT instance_id, service_name, hostname, port, health_status, metadata,
		       is_external, external_url, version, region, zone, last_heartbeat,
		       registered_at, created_at, updated_at
		FROM service_registry
		WHERE service_name = $1 AND health_status = $2
		ORDER BY instance_id
	`

	return pr.scanInstances(ctx, query, serviceName, registry.HealthUp)
}

// GetInstancesByHealth retrieves instances filtered by health status
func (pr *PostgresRegistry) GetInstancesByHealth(ctx context.Context, serviceName string, health registry.HealthStatus) ([]*registry.ServiceInstance, error) {
	const query = `
		SELECT instance_id, service_name, hostname, port, health_status, metadata,
		       is_external, external_url, version, region, zone, last_heartbeat,
		       registered_at, created_at, updated_at
		FROM service_registry
		WHERE service_name = $1 AND health_status = $2
		ORDER BY instance_id
	`

	return pr.scanInstances(ctx, query, serviceName, health)
}

// GetAllInstances retrieves all registered instances
func (pr *PostgresRegistry) GetAllInstances(ctx context.Context) ([]*registry.ServiceInstance, error) {
	const query = `
		SELECT instance_id, service_name, hostname, port, health_status, metadata,
		       is_external, external_url, version, region, zone, last_heartbeat,
		       registered_at, created_at, updated_at
		FROM service_registry
		ORDER BY service_name, instance_id
	`

	rows, err := pr.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query instances: %w", err)
	}
	defer rows.Close()

	var instances []*registry.ServiceInstance
	for rows.Next() {
		instance, err := pr.parseRow(rows)
		if err != nil {
			pr.logger.Error("failed to parse instance row", "error", err)
			continue
		}
		instances = append(instances, instance)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading rows: %w", err)
	}

	return instances, nil
}

// UpdateHealth updates the health status of an instance
func (pr *PostgresRegistry) UpdateHealth(ctx context.Context, instanceID string, health registry.HealthStatus) error {
	const query = `
		UPDATE service_registry
		SET health_status = $1, updated_at = $2
		WHERE instance_id = $3
	`

	result, err := pr.pool.Exec(ctx, query, health, time.Now(), instanceID)
	if err != nil {
		return fmt.Errorf("failed to update health: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("instance not found: %s", instanceID)
	}

	return nil
}

// UpdateMetadata updates the metadata of an instance
func (pr *PostgresRegistry) UpdateMetadata(ctx context.Context, instanceID string, metadata map[string]string) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	const query = `
		UPDATE service_registry
		SET metadata = $1, updated_at = $2
		WHERE instance_id = $3
	`

	result, err := pr.pool.Exec(ctx, query, metadataJSON, time.Now(), instanceID)
	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("instance not found: %s", instanceID)
	}

	return nil
}

// Heartbeat records a heartbeat for an instance
func (pr *PostgresRegistry) Heartbeat(ctx context.Context, instanceID string) error {
	const query = `
		UPDATE service_registry
		SET last_heartbeat = $1, updated_at = $2
		WHERE instance_id = $3
	`

	result, err := pr.pool.Exec(ctx, query, time.Now(), time.Now(), instanceID)
	if err != nil {
		return fmt.Errorf("failed to record heartbeat: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("instance not found: %s", instanceID)
	}

	return nil
}

// CleanupStale removes instances that haven't heartbeated within TTL
func (pr *PostgresRegistry) CleanupStale(ctx context.Context, ttl time.Duration) (int, error) {
	const query = `
		DELETE FROM service_registry
		WHERE last_heartbeat < now() - interval '1 second' * $1
	`

	result, err := pr.pool.Exec(ctx, query, int64(ttl.Seconds()))
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup stale instances: %w", err)
	}

	return int(result.RowsAffected()), nil
}

// Close closes the connection (noop for pool-based backend)
func (pr *PostgresRegistry) Close() error {
	// Pool is managed externally, don't close it
	return nil
}

// Helper functions

// scanInstance executes a query and scans a single instance
func (pr *PostgresRegistry) scanInstance(ctx context.Context, query string, args ...interface{}) (*registry.ServiceInstance, error) {
	row := pr.pool.QueryRow(ctx, query, args...)
	instance, err := pr.parseRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("instance not found")
		}
		return nil, fmt.Errorf("failed to scan instance: %w", err)
	}
	return instance, nil
}

// scanInstances executes a query and scans multiple instances
func (pr *PostgresRegistry) scanInstances(ctx context.Context, query string, args ...interface{}) ([]*registry.ServiceInstance, error) {
	rows, err := pr.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query instances: %w", err)
	}
	defer rows.Close()

	var instances []*registry.ServiceInstance
	for rows.Next() {
		instance, err := pr.parseRow(rows)
		if err != nil {
			pr.logger.Error("failed to parse instance row", "error", err)
			continue
		}
		instances = append(instances, instance)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading rows: %w", err)
	}

	return instances, nil
}

// parseRow parses a database row into a ServiceInstance
func (pr *PostgresRegistry) parseRow(row interface {
	Scan(...interface{}) error
}) (*registry.ServiceInstance, error) {
	var instance registry.ServiceInstance
	var metadataJSON []byte

	err := row.Scan(
		&instance.ID,
		&instance.ServiceName,
		&instance.Host,
		&instance.Port,
		&instance.Health,
		&metadataJSON,
		&instance.IsExternal,
		&instance.ExternalURL,
		&instance.Version,
		&instance.Region,
		&instance.Zone,
		&instance.LastHeartbeat,
		&instance.RegisteredAt,
		&instance.CreatedAt,
		&instance.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse metadata JSON
	if metadataJSON != nil {
		err = json.Unmarshal(metadataJSON, &instance.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}
	if instance.Metadata == nil {
		instance.Metadata = make(map[string]string)
	}

	return &instance, nil
}
