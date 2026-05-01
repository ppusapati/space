// Package services holds gi-predict business logic.
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	"github.com/ppusapati/space/services/gi-predict/internal/models"
	"github.com/ppusapati/space/services/gi-predict/internal/repository"
)

type Predict struct {
	repo  *repository.Repo
	IDFn  func() ulid.ID
	NowFn func() time.Time
}

func New(repo *repository.Repo) *Predict {
	return &Predict{
		repo:  repo,
		IDFn:  ulid.NewMonotonic,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

type SubmitForecastJobInput struct {
	TenantID       ulid.ID
	Type           models.ForecastType
	ModelID        ulid.ID
	InputURIs      []string
	HorizonDays    int32
	ParametersJSON string
	CreatedBy      string
}

func (p *Predict) SubmitForecastJob(ctx context.Context, in SubmitForecastJobInput) (*models.ForecastJob, error) {
	if in.TenantID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	if in.Type == models.TypeUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "type required")
	}
	if in.HorizonDays <= 0 || in.HorizonDays > 3650 {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "horizon_days must be in (0, 3650]")
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
	return p.repo.SubmitForecastJob(ctx, repository.SubmitForecastJobParams{
		ID:             p.IDFn(),
		TenantID:       in.TenantID,
		Type:           in.Type,
		Status:         models.StatusQueued,
		ModelID:        in.ModelID,
		InputURIs:      cleaned,
		HorizonDays:    in.HorizonDays,
		ParametersJSON: in.ParametersJSON,
		CreatedBy:      createdBy,
	})
}

func (p *Predict) GetForecastJob(ctx context.Context, id ulid.ID) (*models.ForecastJob, error) {
	j, err := p.repo.GetForecastJob(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("FORECAST_JOB_NOT_FOUND", "forecast_job "+id.String())
	}
	return j, err
}

type ListForecastJobsInput struct {
	TenantID   ulid.ID
	Status     *models.ForecastStatus
	Type       *models.ForecastType
	PageOffset int32
	PageSize   int32
}

func (p *Predict) ListForecastJobsForTenant(ctx context.Context, in ListForecastJobsInput) ([]*models.ForecastJob, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := p.repo.ListForecastJobsForTenant(ctx, repository.ListForecastJobsParams{
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

type UpdateForecastJobStatusInput struct {
	ID                 ulid.ID
	Status             models.ForecastStatus
	OutputURI          string
	ResultsSummaryJSON string
	ErrorMessage       string
	UpdatedBy          string
}

func (p *Predict) UpdateForecastJobStatus(ctx context.Context, in UpdateForecastJobStatusInput) (*models.ForecastJob, error) {
	if in.Status == models.StatusUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "status required")
	}
	current, err := p.GetForecastJob(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if !validTransition(current.Status, in.Status) {
		return nil, pkgerrors.New(412, "ILLEGAL_TRANSITION",
			"illegal forecast_job status transition")
	}
	if in.UpdatedBy == "" {
		in.UpdatedBy = "system"
	}
	updated, err := p.repo.UpdateForecastJobStatus(ctx, repository.UpdateForecastJobStatusParams{
		ID:                 in.ID,
		Status:             in.Status,
		OutputURI:          in.OutputURI,
		ResultsSummaryJSON: in.ResultsSummaryJSON,
		ErrorMessage:       in.ErrorMessage,
		UpdatedBy:          in.UpdatedBy,
	})
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("FORECAST_JOB_NOT_FOUND", "forecast_job "+in.ID.String())
	}
	return updated, err
}

func (p *Predict) CancelForecastJob(ctx context.Context, id ulid.ID, updatedBy string) (*models.ForecastJob, error) {
	return p.UpdateForecastJobStatus(ctx, UpdateForecastJobStatusInput{
		ID:           id,
		Status:       models.StatusCanceled,
		ErrorMessage: "canceled by user",
		UpdatedBy:    updatedBy,
	})
}

func validTransition(from, to models.ForecastStatus) bool {
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
