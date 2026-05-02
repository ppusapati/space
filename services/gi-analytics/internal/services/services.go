// Package services holds gi-analytics business logic.
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	"github.com/ppusapati/space/services/gi-analytics/internal/models"
	"github.com/ppusapati/space/services/gi-analytics/internal/repository"
)

type Analytics struct {
	repo  *repository.Repo
	IDFn  func() ulid.ID
	NowFn func() time.Time
}

func New(repo *repository.Repo) *Analytics {
	return &Analytics{
		repo:  repo,
		IDFn:  ulid.NewMonotonic,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

type SubmitAnalysisJobInput struct {
	TenantID       ulid.ID
	Type           models.AnalysisType
	InputURIs      []string
	ParametersJSON string
	CreatedBy      string
}

func (a *Analytics) SubmitAnalysisJob(ctx context.Context, in SubmitAnalysisJobInput) (*models.AnalysisJob, error) {
	if in.TenantID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	if in.Type == models.TypeUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "type required")
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
	return a.repo.SubmitAnalysisJob(ctx, repository.SubmitAnalysisJobParams{
		ID:             a.IDFn(),
		TenantID:       in.TenantID,
		Type:           in.Type,
		Status:         models.StatusQueued,
		InputURIs:      cleaned,
		ParametersJSON: in.ParametersJSON,
		CreatedBy:      createdBy,
	})
}

func (a *Analytics) GetAnalysisJob(ctx context.Context, id ulid.ID) (*models.AnalysisJob, error) {
	j, err := a.repo.GetAnalysisJob(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("ANALYSIS_JOB_NOT_FOUND", "analysis_job "+id.String())
	}
	return j, err
}

type ListAnalysisJobsInput struct {
	TenantID   ulid.ID
	Status     *models.AnalysisStatus
	Type       *models.AnalysisType
	PageOffset int32
	PageSize   int32
}

func (a *Analytics) ListAnalysisJobsForTenant(ctx context.Context, in ListAnalysisJobsInput) ([]*models.AnalysisJob, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := a.repo.ListAnalysisJobsForTenant(ctx, repository.ListAnalysisJobsParams{
		TenantID:   in.TenantID,
		Status:     in.Status,
		Type:       in.Type,
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

type UpdateAnalysisJobStatusInput struct {
	ID                 ulid.ID
	Status             models.AnalysisStatus
	OutputURI          string
	ResultsSummaryJSON string
	ErrorMessage       string
	UpdatedBy          string
}

func (a *Analytics) UpdateAnalysisJobStatus(ctx context.Context, in UpdateAnalysisJobStatusInput) (*models.AnalysisJob, error) {
	if in.Status == models.StatusUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "status required")
	}
	current, err := a.GetAnalysisJob(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if !validTransition(current.Status, in.Status) {
		return nil, pkgerrors.New(412, "ILLEGAL_TRANSITION",
			"illegal analysis_job status transition")
	}
	if in.UpdatedBy == "" {
		in.UpdatedBy = "system"
	}
	updated, err := a.repo.UpdateAnalysisJobStatus(ctx, repository.UpdateAnalysisJobStatusParams{
		ID:                 in.ID,
		Status:             in.Status,
		OutputURI:          in.OutputURI,
		ResultsSummaryJSON: in.ResultsSummaryJSON,
		ErrorMessage:       in.ErrorMessage,
		UpdatedBy:          in.UpdatedBy,
	})
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("ANALYSIS_JOB_NOT_FOUND", "analysis_job "+in.ID.String())
	}
	return updated, err
}

func (a *Analytics) CancelAnalysisJob(ctx context.Context, id ulid.ID, updatedBy string) (*models.AnalysisJob, error) {
	return a.UpdateAnalysisJobStatus(ctx, UpdateAnalysisJobStatusInput{
		ID:           id,
		Status:       models.StatusCanceled,
		ErrorMessage: "canceled by user",
		UpdatedBy:    updatedBy,
	})
}

func validTransition(from, to models.AnalysisStatus) bool {
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
