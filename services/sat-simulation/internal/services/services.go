// Package services holds sat-simulation business logic.
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	"github.com/ppusapati/space/services/sat-simulation/internal/models"
	"github.com/ppusapati/space/services/sat-simulation/internal/repository"
)

// Simulation is the sat-simulation service-layer facade.
type Simulation struct {
	repo  *repository.Repo
	IDFn  func() ulid.ID
	NowFn func() time.Time
}

// New constructs a Simulation service.
func New(repo *repository.Repo) *Simulation {
	return &Simulation{
		repo:  repo,
		IDFn:  ulid.NewMonotonic,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// ----- Scenario ------------------------------------------------------------

// CreateScenarioInput is the input for [Simulation.CreateScenario].
type CreateScenarioInput struct {
	TenantID    ulid.ID
	Slug        string
	Title       string
	Description string
	SpecJSON    string
	CreatedBy   string
}

// CreateScenario persists a new scenario.
func (s *Simulation) CreateScenario(ctx context.Context, in CreateScenarioInput) (*models.Scenario, error) {
	if in.TenantID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	in.Slug = strings.TrimSpace(in.Slug)
	in.Title = strings.TrimSpace(in.Title)
	if in.Slug == "" || in.Title == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "slug and title required")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return s.repo.CreateScenario(ctx, repository.CreateScenarioParams{
		ID:          s.IDFn(),
		TenantID:    in.TenantID,
		Slug:        in.Slug,
		Title:       in.Title,
		Description: in.Description,
		SpecJSON:    in.SpecJSON,
		CreatedBy:   createdBy,
	})
}

// GetScenario fetches a scenario by id.
func (s *Simulation) GetScenario(ctx context.Context, id ulid.ID) (*models.Scenario, error) {
	sc, err := s.repo.GetScenario(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("SCENARIO_NOT_FOUND", "scenario "+id.String())
	}
	return sc, err
}

// ListScenariosForTenant returns one page of scenarios.
func (s *Simulation) ListScenariosForTenant(
	ctx context.Context, tenantID ulid.ID, offset, size int32,
) ([]*models.Scenario, models.Page, error) {
	if tenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := s.repo.ListScenariosForTenant(ctx, tenantID, offset, size)
	if err != nil {
		return nil, models.Page{}, err
	}
	return rows, models.Page{
		TotalCount: total,
		PageOffset: offset,
		PageSize:   size,
		HasNext:    offset+int32(len(rows)) < total,
	}, nil
}

// DeprecateScenario marks a scenario inactive.
func (s *Simulation) DeprecateScenario(ctx context.Context, id ulid.ID, updatedBy string) (*models.Scenario, error) {
	if updatedBy == "" {
		updatedBy = "system"
	}
	sc, err := s.repo.DeprecateScenario(ctx, id, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("SCENARIO_NOT_FOUND", "scenario "+id.String())
	}
	return sc, err
}

// ----- SimulationRun -------------------------------------------------------

// StartRunInput is the input for [Simulation.StartRun].
type StartRunInput struct {
	TenantID       ulid.ID
	SatelliteID    ulid.ID
	ScenarioID     ulid.ID
	Mode           models.SimulationMode
	ParametersJSON string
	CreatedBy      string
}

// StartRun persists a new run in QUEUED status. Validates that the scenario
// exists, belongs to the same tenant, and is active.
func (s *Simulation) StartRun(ctx context.Context, in StartRunInput) (*models.SimulationRun, error) {
	if in.TenantID.IsZero() || in.SatelliteID.IsZero() || in.ScenarioID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT",
			"tenant_id, satellite_id, scenario_id required")
	}
	if in.Mode == models.ModeUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "mode required")
	}
	sc, err := s.repo.GetScenario(ctx, in.ScenarioID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, pkgerrors.BadRequest("SCENARIO_NOT_FOUND",
				"scenario "+in.ScenarioID.String())
		}
		return nil, err
	}
	if sc.TenantID != in.TenantID {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "scenario tenant mismatch")
	}
	if !sc.Active {
		return nil, pkgerrors.New(412, "SCENARIO_DEPRECATED", "scenario is deprecated")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return s.repo.StartRun(ctx, repository.StartRunParams{
		ID:             s.IDFn(),
		TenantID:       in.TenantID,
		SatelliteID:    in.SatelliteID,
		ScenarioID:     in.ScenarioID,
		Mode:           in.Mode,
		Status:         models.StatusQueued,
		ParametersJSON: in.ParametersJSON,
		CreatedBy:      createdBy,
	})
}

// GetRun fetches a run by id.
func (s *Simulation) GetRun(ctx context.Context, id ulid.ID) (*models.SimulationRun, error) {
	r, err := s.repo.GetRun(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("RUN_NOT_FOUND", "run "+id.String())
	}
	return r, err
}

// ListRunsInput is the input for [Simulation.ListRunsForTenant].
type ListRunsInput struct {
	TenantID    ulid.ID
	SatelliteID *ulid.ID
	ScenarioID  *ulid.ID
	Status      *models.RunStatus
	Mode        *models.SimulationMode
	PageOffset  int32
	PageSize    int32
}

// ListRunsForTenant returns one page of runs.
func (s *Simulation) ListRunsForTenant(ctx context.Context, in ListRunsInput) ([]*models.SimulationRun, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := s.repo.ListRunsForTenant(ctx, repository.ListRunsParams{
		TenantID:    in.TenantID,
		SatelliteID: in.SatelliteID,
		ScenarioID:  in.ScenarioID,
		Status:      in.Status,
		Mode:        in.Mode,
		PageOffset:  in.PageOffset,
		PageSize:    in.PageSize,
	})
	if err != nil {
		return nil, models.Page{}, err
	}
	return rows, models.Page{
		TotalCount: total,
		PageOffset: in.PageOffset,
		PageSize:   in.PageSize,
		HasNext:    in.PageOffset+int32(len(rows)) < total,
	}, nil
}

// CancelRun marks a run CANCELED.
func (s *Simulation) CancelRun(ctx context.Context, id ulid.ID, reason, updatedBy string) (*models.SimulationRun, error) {
	msg := strings.TrimSpace(reason)
	if msg == "" {
		msg = "canceled by user"
	}
	return s.UpdateRunStatus(ctx, UpdateRunStatusInput{
		ID:           id,
		Status:       models.StatusCanceled,
		ErrorMessage: msg,
		UpdatedBy:    updatedBy,
	})
}

// UpdateRunStatusInput is the input for [Simulation.UpdateRunStatus].
type UpdateRunStatusInput struct {
	ID           ulid.ID
	Status       models.RunStatus
	LogURI       string
	TelemetryURI string
	ResultsJSON  string
	Score        float64
	ErrorMessage string
	UpdatedBy    string
}

// UpdateRunStatus transitions a run to a new status. Valid transitions:
//
//	QUEUED  -> RUNNING | CANCELED
//	RUNNING -> COMPLETED | FAILED | CANCELED
//	COMPLETED, FAILED, CANCELED — terminal.
func (s *Simulation) UpdateRunStatus(ctx context.Context, in UpdateRunStatusInput) (*models.SimulationRun, error) {
	if in.Status == models.StatusUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "status required")
	}
	current, err := s.GetRun(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if !validRunTransition(current.Status, in.Status) {
		return nil, pkgerrors.New(412, "ILLEGAL_TRANSITION",
			"illegal run status transition")
	}
	if in.UpdatedBy == "" {
		in.UpdatedBy = "system"
	}
	updated, err := s.repo.UpdateRunStatus(ctx, repository.UpdateRunStatusParams{
		ID:           in.ID,
		Status:       in.Status,
		LogURI:       in.LogURI,
		TelemetryURI: in.TelemetryURI,
		ResultsJSON:  in.ResultsJSON,
		Score:        in.Score,
		ErrorMessage: in.ErrorMessage,
		UpdatedBy:    in.UpdatedBy,
	})
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("RUN_NOT_FOUND", "run "+in.ID.String())
	}
	return updated, err
}

func validRunTransition(from, to models.RunStatus) bool {
	switch from {
	case models.StatusQueued:
		return to == models.StatusRunning || to == models.StatusCanceled
	case models.StatusRunning:
		return to == models.StatusCompleted || to == models.StatusFailed || to == models.StatusCanceled
	case models.StatusCompleted, models.StatusFailed, models.StatusCanceled:
		return false
	default:
		return false
	}
}
