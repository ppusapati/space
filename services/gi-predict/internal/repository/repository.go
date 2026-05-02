// Package repository wraps the gi-predict sqlc layer.
package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/chetana/packages/ulid"

	gipdb "github.com/ppusapati/space/services/gi-predict/db/generated"
	"github.com/ppusapati/space/services/gi-predict/internal/mapper"
	"github.com/ppusapati/space/services/gi-predict/internal/models"
)

var ErrNotFound = errors.New("repository: not found")

type Repo struct {
	q    *gipdb.Queries
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: gipdb.New(pool), pool: pool}
}

type SubmitForecastJobParams struct {
	ID             ulid.ID
	TenantID       ulid.ID
	Type           models.ForecastType
	Status         models.ForecastStatus
	ModelID        ulid.ID
	InputURIs      []string
	HorizonDays    int32
	ParametersJSON string
	CreatedBy      string
}

func (r *Repo) SubmitForecastJob(ctx context.Context, p SubmitForecastJobParams) (*models.ForecastJob, error) {
	params := strings.TrimSpace(p.ParametersJSON)
	if params == "" {
		params = "{}"
	}
	row, err := r.q.SubmitForecastJob(ctx, gipdb.SubmitForecastJobParams{
		ID:             mapper.PgUUID(p.ID),
		TenantID:       mapper.PgUUID(p.TenantID),
		Type:           int32(p.Type),
		Status:         int32(p.Status),
		ModelID:        mapper.PgUUIDOrNull(p.ModelID),
		InputUris:      p.InputURIs,
		HorizonDays:    p.HorizonDays,
		ParametersJson: []byte(params),
		CreatedBy:      p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.ForecastJobFromRow(row), nil
}

func (r *Repo) GetForecastJob(ctx context.Context, id ulid.ID) (*models.ForecastJob, error) {
	row, err := r.q.GetForecastJob(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.ForecastJobFromRow(row), nil
}

type ListForecastJobsParams struct {
	TenantID   ulid.ID
	Status     *models.ForecastStatus
	Type       *models.ForecastType
	PageOffset int32
	PageSize   int32
}

func (r *Repo) ListForecastJobsForTenant(ctx context.Context, p ListForecastJobsParams) ([]*models.ForecastJob, int32, error) {
	var statusPtr, typePtr *int32
	if p.Status != nil {
		v := int32(*p.Status)
		statusPtr = &v
	}
	if p.Type != nil {
		v := int32(*p.Type)
		typePtr = &v
	}
	total, err := r.q.CountForecastJobsForTenant(ctx, gipdb.CountForecastJobsForTenantParams{
		TenantID: mapper.PgUUID(p.TenantID),
		Status:   statusPtr,
		Type:     typePtr,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListForecastJobsForTenant(ctx, gipdb.ListForecastJobsForTenantParams{
		TenantID:   mapper.PgUUID(p.TenantID),
		Status:     statusPtr,
		Type:       typePtr,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.ForecastJob, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.ForecastJobFromRow(row))
	}
	return out, int32(total), nil
}

type UpdateForecastJobStatusParams struct {
	ID                 ulid.ID
	Status             models.ForecastStatus
	OutputURI          string
	ResultsSummaryJSON string
	ErrorMessage       string
	UpdatedBy          string
}

func (r *Repo) UpdateForecastJobStatus(ctx context.Context, p UpdateForecastJobStatusParams) (*models.ForecastJob, error) {
	row, err := r.q.UpdateForecastJobStatus(ctx, gipdb.UpdateForecastJobStatusParams{
		ID:                 mapper.PgUUID(p.ID),
		Status:             int32(p.Status),
		OutputUri:          p.OutputURI,
		ResultsSummaryJson: p.ResultsSummaryJSON,
		ErrorMessage:       p.ErrorMessage,
		UpdatedBy:          p.UpdatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.ForecastJobFromRow(row), nil
}
