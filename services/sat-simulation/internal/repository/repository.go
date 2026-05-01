// Package repository wraps the sat-simulation sqlc layer.
package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/samavaya/packages/ulid"

	satsimdb "github.com/ppusapati/space/services/sat-simulation/db/generated"
	"github.com/ppusapati/space/services/sat-simulation/internal/mapper"
	"github.com/ppusapati/space/services/sat-simulation/internal/models"
)

// ErrNotFound is returned when no row matches.
var ErrNotFound = errors.New("repository: not found")

// Repo persists Scenarios and SimulationRuns.
type Repo struct {
	q    *satsimdb.Queries
	pool *pgxpool.Pool
}

// New constructs a Repo.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: satsimdb.New(pool), pool: pool}
}

// ----- Scenario ------------------------------------------------------------

// CreateScenarioParams holds the input for [Repo.CreateScenario].
type CreateScenarioParams struct {
	ID          ulid.ID
	TenantID    ulid.ID
	Slug        string
	Title       string
	Description string
	SpecJSON    string
	CreatedBy   string
}

// CreateScenario inserts a new scenarios row.
func (r *Repo) CreateScenario(ctx context.Context, p CreateScenarioParams) (*models.Scenario, error) {
	spec := strings.TrimSpace(p.SpecJSON)
	if spec == "" {
		spec = "{}"
	}
	row, err := r.q.CreateScenario(ctx, satsimdb.CreateScenarioParams{
		ID:          mapper.PgUUID(p.ID),
		TenantID:    mapper.PgUUID(p.TenantID),
		Slug:        p.Slug,
		Title:       p.Title,
		Description: p.Description,
		SpecJson:    []byte(spec),
		CreatedBy:   p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.ScenarioFromRow(row), nil
}

// GetScenario returns a scenario by id.
func (r *Repo) GetScenario(ctx context.Context, id ulid.ID) (*models.Scenario, error) {
	row, err := r.q.GetScenario(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.ScenarioFromRow(row), nil
}

// ListScenariosForTenant returns one page of scenarios.
func (r *Repo) ListScenariosForTenant(
	ctx context.Context, tenantID ulid.ID, offset, size int32,
) ([]*models.Scenario, int32, error) {
	total, err := r.q.CountScenariosForTenant(ctx, mapper.PgUUID(tenantID))
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListScenariosForTenant(ctx, satsimdb.ListScenariosForTenantParams{
		TenantID:   mapper.PgUUID(tenantID),
		PageOffset: offset,
		PageSize:   size,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.Scenario, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.ScenarioFromRow(row))
	}
	return out, int32(total), nil
}

// DeprecateScenario marks a scenario inactive.
func (r *Repo) DeprecateScenario(ctx context.Context, id ulid.ID, updatedBy string) (*models.Scenario, error) {
	row, err := r.q.DeprecateScenario(ctx, satsimdb.DeprecateScenarioParams{
		ID:        mapper.PgUUID(id),
		UpdatedBy: updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.ScenarioFromRow(row), nil
}

// ----- SimulationRun -------------------------------------------------------

// StartRunParams holds the input for [Repo.StartRun].
type StartRunParams struct {
	ID             ulid.ID
	TenantID       ulid.ID
	SatelliteID    ulid.ID
	ScenarioID     ulid.ID
	Mode           models.SimulationMode
	Status         models.RunStatus
	ParametersJSON string
	CreatedBy      string
}

// StartRun inserts a new simulation_runs row.
func (r *Repo) StartRun(ctx context.Context, p StartRunParams) (*models.SimulationRun, error) {
	params := strings.TrimSpace(p.ParametersJSON)
	if params == "" {
		params = "{}"
	}
	row, err := r.q.StartRun(ctx, satsimdb.StartRunParams{
		ID:             mapper.PgUUID(p.ID),
		TenantID:       mapper.PgUUID(p.TenantID),
		SatelliteID:    mapper.PgUUID(p.SatelliteID),
		ScenarioID:     mapper.PgUUID(p.ScenarioID),
		Mode:           int32(p.Mode),
		Status:         int32(p.Status),
		ParametersJson: []byte(params),
		CreatedBy:      p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.RunFromRow(row), nil
}

// GetRun returns a simulation run by id.
func (r *Repo) GetRun(ctx context.Context, id ulid.ID) (*models.SimulationRun, error) {
	row, err := r.q.GetRun(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.RunFromRow(row), nil
}

// ListRunsParams holds the input for [Repo.ListRunsForTenant].
type ListRunsParams struct {
	TenantID    ulid.ID
	SatelliteID *ulid.ID
	ScenarioID  *ulid.ID
	Status      *models.RunStatus
	Mode        *models.SimulationMode
	PageOffset  int32
	PageSize    int32
}

// ListRunsForTenant returns one page of simulation runs.
func (r *Repo) ListRunsForTenant(ctx context.Context, p ListRunsParams) ([]*models.SimulationRun, int32, error) {
	var statusPtr *int32
	if p.Status != nil {
		v := int32(*p.Status)
		statusPtr = &v
	}
	var modePtr *int32
	if p.Mode != nil {
		v := int32(*p.Mode)
		modePtr = &v
	}
	var satellitePg pgtype.UUID
	if p.SatelliteID != nil {
		satellitePg = mapper.PgUUID(*p.SatelliteID)
	}
	var scenarioPg pgtype.UUID
	if p.ScenarioID != nil {
		scenarioPg = mapper.PgUUID(*p.ScenarioID)
	}
	total, err := r.q.CountRunsForTenant(ctx, satsimdb.CountRunsForTenantParams{
		TenantID:    mapper.PgUUID(p.TenantID),
		SatelliteID: satellitePg,
		ScenarioID:  scenarioPg,
		Status:      statusPtr,
		Mode:        modePtr,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListRunsForTenant(ctx, satsimdb.ListRunsForTenantParams{
		TenantID:    mapper.PgUUID(p.TenantID),
		SatelliteID: satellitePg,
		ScenarioID:  scenarioPg,
		Status:      statusPtr,
		Mode:        modePtr,
		PageOffset:  p.PageOffset,
		PageSize:    p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.SimulationRun, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.RunFromRow(row))
	}
	return out, int32(total), nil
}

// UpdateRunStatusParams holds the input for [Repo.UpdateRunStatus].
type UpdateRunStatusParams struct {
	ID           ulid.ID
	Status       models.RunStatus
	LogURI       string
	TelemetryURI string
	ResultsJSON  string
	Score        float64
	ErrorMessage string
	UpdatedBy    string
}

// UpdateRunStatus updates a run with optional artefact fields and side
// effects (started_at, finished_at) driven by status.
func (r *Repo) UpdateRunStatus(ctx context.Context, p UpdateRunStatusParams) (*models.SimulationRun, error) {
	row, err := r.q.UpdateRunStatus(ctx, satsimdb.UpdateRunStatusParams{
		ID:           mapper.PgUUID(p.ID),
		Status:       int32(p.Status),
		LogUri:       p.LogURI,
		TelemetryUri: p.TelemetryURI,
		ResultsJson:  p.ResultsJSON,
		Score:        p.Score,
		ErrorMessage: p.ErrorMessage,
		UpdatedBy:    p.UpdatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.RunFromRow(row), nil
}
