// Package services holds eo-pipeline business logic.
package services

import (
	"context"
	"errors"
	"time"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	"github.com/ppusapati/space/services/eo-pipeline/internal/models"
	"github.com/ppusapati/space/services/eo-pipeline/internal/repository"
)

// Pipeline is the eo-pipeline service-layer facade.
type Pipeline struct {
	repo  *repository.Repo
	IDFn  func() ulid.ID
	NowFn func() time.Time
}

// New constructs a Pipeline service.
func New(repo *repository.Repo) *Pipeline {
	return &Pipeline{
		repo:  repo,
		IDFn:  ulid.NewMonotonic,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// SubmitJobInput is the input for [Pipeline.SubmitJob].
type SubmitJobInput struct {
	TenantID       ulid.ID
	ItemID         ulid.ID
	Stage          models.JobStage
	ParametersJSON string
	CreatedBy      string
}

// SubmitJob persists a new pending job.
func (p *Pipeline) SubmitJob(ctx context.Context, in SubmitJobInput) (*models.Job, error) {
	if in.TenantID.IsZero() || in.ItemID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id and item_id required")
	}
	if in.Stage == models.StageUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "stage required")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return p.repo.CreateJob(ctx, repository.CreateJobParams{
		ID:             p.IDFn(),
		TenantID:       in.TenantID,
		ItemID:         in.ItemID,
		Stage:          in.Stage,
		Status:         models.StatusPending,
		ParametersJSON: in.ParametersJSON,
		CreatedBy:      createdBy,
	})
}

// GetJob fetches a job by id.
func (p *Pipeline) GetJob(ctx context.Context, id ulid.ID) (*models.Job, error) {
	j, err := p.repo.GetJob(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("JOB_NOT_FOUND", "job "+id.String())
	}
	return j, err
}

// ListJobsInput is the input for [Pipeline.ListJobsForTenant].
type ListJobsInput struct {
	TenantID   ulid.ID
	Status     *models.JobStatus
	Stage      *models.JobStage
	PageOffset int32
	PageSize   int32
}

// ListJobsForTenant returns one page of jobs.
func (p *Pipeline) ListJobsForTenant(ctx context.Context, in ListJobsInput) ([]*models.Job, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := p.repo.ListJobsForTenant(ctx, repository.ListJobsParams{
		TenantID:   in.TenantID,
		Status:     in.Status,
		Stage:      in.Stage,
		PageOffset: in.PageOffset,
		PageSize:   in.PageSize,
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

// UpdateJobStatus updates a job's status (with started_at/finished_at side
// effects driven by SQL). Validates legal transitions.
func (p *Pipeline) UpdateJobStatus(
	ctx context.Context, id ulid.ID, status models.JobStatus, outputURI, errorMessage, updatedBy string,
) (*models.Job, error) {
	if status == models.StatusUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "status required")
	}
	current, err := p.GetJob(ctx, id)
	if err != nil {
		return nil, err
	}
	if !validTransition(current.Status, status) {
		return nil, pkgerrors.New(412, "ILLEGAL_TRANSITION",
			"illegal job status transition")
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	updated, err := p.repo.UpdateJobStatus(ctx, id, status, outputURI, errorMessage, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("JOB_NOT_FOUND", "job "+id.String())
	}
	return updated, err
}

// CancelJob marks a job CANCELLED (only if not already terminal).
func (p *Pipeline) CancelJob(ctx context.Context, id ulid.ID, updatedBy string) (*models.Job, error) {
	return p.UpdateJobStatus(ctx, id, models.StatusCancelled, "", "cancelled by user", updatedBy)
}

// validTransition enforces the job-status transition graph:
//   PENDING   -> RUNNING | CANCELLED
//   RUNNING   -> SUCCEEDED | FAILED | CANCELLED
//   SUCCEEDED, FAILED, CANCELLED — terminal.
func validTransition(from, to models.JobStatus) bool {
	switch from {
	case models.StatusPending:
		return to == models.StatusRunning || to == models.StatusCancelled
	case models.StatusRunning:
		return to == models.StatusSucceeded || to == models.StatusFailed || to == models.StatusCancelled
	case models.StatusSucceeded, models.StatusFailed, models.StatusCancelled:
		return false
	default:
		return false
	}
}
