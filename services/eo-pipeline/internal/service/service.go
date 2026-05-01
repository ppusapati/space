// Package service holds eo-pipeline business logic.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/ppusapati/space/pkg/errs"
	"github.com/ppusapati/space/services/eo-pipeline/internal/models"
	"github.com/ppusapati/space/services/eo-pipeline/internal/repository"
)

// Pipeline is the service-layer facade.
type Pipeline struct {
	repo  *repository.JobRepository
	IDFn  func() uuid.UUID
	NowFn func() time.Time
}

// New constructs a Pipeline.
func New(repo *repository.JobRepository) *Pipeline {
	return &Pipeline{
		repo:  repo,
		IDFn:  uuid.New,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// SubmitJobInput is the input to [Pipeline.SubmitJob].
type SubmitJobInput struct {
	TenantID       uuid.UUID
	ItemID         uuid.UUID
	Stage          models.JobStage
	ParametersJSON string
	CreatedBy      string
}

// SubmitJob persists a new pending Job.
func (p *Pipeline) SubmitJob(ctx context.Context, in SubmitJobInput) (*models.Job, error) {
	if in.TenantID == uuid.Nil || in.ItemID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id and item_id required")
	}
	if in.Stage == models.StageUnspecified {
		return nil, errs.New(errs.DomainInvalidArgument, "stage required")
	}
	if in.ParametersJSON == "" {
		in.ParametersJSON = "{}"
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	job := &models.Job{
		ID:             p.IDFn(),
		TenantID:       in.TenantID,
		ItemID:         in.ItemID,
		Stage:          in.Stage,
		Status:         models.StatusPending,
		ParametersJSON: in.ParametersJSON,
		CreatedBy:      createdBy,
	}
	return p.repo.Create(ctx, job)
}

// GetJob fetches by id.
func (p *Pipeline) GetJob(ctx context.Context, id uuid.UUID) (*models.Job, error) {
	j, err := p.repo.Get(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errs.New(errs.DomainNotFound, "job %s", id)
	}
	return j, err
}

// ListJobsInput is the input to [Pipeline.ListJobs].
type ListJobsInput struct {
	TenantID      uuid.UUID
	Status        *models.JobStatus
	Stage         *models.JobStage
	CursorCreated *time.Time
	CursorID      uuid.UUID
	Limit         int32
}

// ListJobs returns one page of jobs.
func (p *Pipeline) ListJobs(ctx context.Context, in ListJobsInput) ([]*models.Job, error) {
	if in.TenantID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id required")
	}
	return p.repo.List(ctx, repository.ListParams{
		TenantID:      in.TenantID,
		Status:        in.Status,
		Stage:         in.Stage,
		CursorCreated: in.CursorCreated,
		CursorID:      in.CursorID,
		Limit:         in.Limit,
	})
}

// UpdateStatus transitions a job. Reject illegal transitions
// (terminal → anything) up front so the persistence layer never sees
// inconsistent state.
func (p *Pipeline) UpdateStatus(
	ctx context.Context, id uuid.UUID, target models.JobStatus,
	outputURI, errMsg, updatedBy string,
) (*models.Job, error) {
	cur, err := p.GetJob(ctx, id)
	if err != nil {
		return nil, err
	}
	if cur.IsTerminal() {
		return nil, errs.New(errs.DomainPreconditionFailed,
			"cannot transition terminal job %s (status=%d)", id, cur.Status)
	}
	if target == models.StatusUnspecified {
		return nil, errs.New(errs.DomainInvalidArgument, "target status required")
	}
	if target == models.StatusFailed && errMsg == "" {
		return nil, errs.New(errs.DomainInvalidArgument, "error_message required when failing")
	}
	if target == models.StatusSucceeded && outputURI == "" {
		return nil, errs.New(errs.DomainInvalidArgument, "output_uri required when succeeding")
	}
	updatedByOrSystem := updatedBy
	if updatedByOrSystem == "" {
		updatedByOrSystem = "system"
	}
	return p.repo.UpdateStatus(ctx, id, target, outputURI, errMsg, updatedByOrSystem)
}

// Cancel marks a non-terminal job as cancelled.
func (p *Pipeline) Cancel(ctx context.Context, id uuid.UUID, updatedBy string) (*models.Job, error) {
	return p.UpdateStatus(ctx, id, models.StatusCancelled, "", "cancelled by user", updatedBy)
}
