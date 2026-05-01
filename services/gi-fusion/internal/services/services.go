// Package services holds gi-fusion business logic.
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	"github.com/ppusapati/space/services/gi-fusion/internal/models"
	"github.com/ppusapati/space/services/gi-fusion/internal/repository"
)

type Fusion struct {
	repo  *repository.Repo
	IDFn  func() ulid.ID
	NowFn func() time.Time
}

func New(repo *repository.Repo) *Fusion {
	return &Fusion{
		repo:  repo,
		IDFn:  ulid.NewMonotonic,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

type SubmitFusionJobInput struct {
	TenantID       ulid.ID
	Method         models.FusionMethod
	InputURIs      []string
	ParametersJSON string
	CreatedBy      string
}

func (f *Fusion) SubmitFusionJob(ctx context.Context, in SubmitFusionJobInput) (*models.FusionJob, error) {
	if in.TenantID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	if in.Method == models.MethodUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "method required")
	}
	cleaned := make([]string, 0, len(in.InputURIs))
	for _, u := range in.InputURIs {
		if t := strings.TrimSpace(u); t != "" {
			cleaned = append(cleaned, t)
		}
	}
	if len(cleaned) == 0 {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "at least one input_uri required")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return f.repo.SubmitFusionJob(ctx, repository.SubmitFusionJobParams{
		ID:             f.IDFn(),
		TenantID:       in.TenantID,
		Method:         in.Method,
		Status:         models.StatusQueued,
		InputURIs:      cleaned,
		ParametersJSON: in.ParametersJSON,
		CreatedBy:      createdBy,
	})
}

func (f *Fusion) GetFusionJob(ctx context.Context, id ulid.ID) (*models.FusionJob, error) {
	j, err := f.repo.GetFusionJob(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("FUSION_JOB_NOT_FOUND", "fusion_job "+id.String())
	}
	return j, err
}

type ListFusionJobsInput struct {
	TenantID   ulid.ID
	Status     *models.FusionStatus
	Method     *models.FusionMethod
	PageOffset int32
	PageSize   int32
}

func (f *Fusion) ListFusionJobsForTenant(ctx context.Context, in ListFusionJobsInput) ([]*models.FusionJob, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := f.repo.ListFusionJobsForTenant(ctx, repository.ListFusionJobsParams{
		TenantID:   in.TenantID,
		Status:     in.Status,
		Method:     in.Method,
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

func (f *Fusion) UpdateFusionJobStatus(
	ctx context.Context, id ulid.ID, status models.FusionStatus, outputURI, errorMessage, updatedBy string,
) (*models.FusionJob, error) {
	if status == models.StatusUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "status required")
	}
	current, err := f.GetFusionJob(ctx, id)
	if err != nil {
		return nil, err
	}
	if !validTransition(current.Status, status) {
		return nil, pkgerrors.New(412, "ILLEGAL_TRANSITION",
			"illegal fusion_job status transition")
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	updated, err := f.repo.UpdateFusionJobStatus(ctx, id, status, outputURI, errorMessage, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("FUSION_JOB_NOT_FOUND", "fusion_job "+id.String())
	}
	return updated, err
}

func (f *Fusion) CancelFusionJob(ctx context.Context, id ulid.ID, updatedBy string) (*models.FusionJob, error) {
	return f.UpdateFusionJobStatus(ctx, id, models.StatusCanceled, "", "canceled by user", updatedBy)
}

func validTransition(from, to models.FusionStatus) bool {
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
