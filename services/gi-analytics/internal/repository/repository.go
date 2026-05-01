// Package repository wraps the gi-analytics sqlc layer.
package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/samavaya/packages/ulid"

	gianadb "github.com/ppusapati/space/services/gi-analytics/db/generated"
	"github.com/ppusapati/space/services/gi-analytics/internal/mapper"
	"github.com/ppusapati/space/services/gi-analytics/internal/models"
)

var ErrNotFound = errors.New("repository: not found")

type Repo struct {
	q    *gianadb.Queries
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: gianadb.New(pool), pool: pool}
}

type SubmitAnalysisJobParams struct {
	ID             ulid.ID
	TenantID       ulid.ID
	Type           models.AnalysisType
	Status         models.AnalysisStatus
	InputURIs      []string
	ParametersJSON string
	CreatedBy      string
}

func (r *Repo) SubmitAnalysisJob(ctx context.Context, p SubmitAnalysisJobParams) (*models.AnalysisJob, error) {
	params := strings.TrimSpace(p.ParametersJSON)
	if params == "" {
		params = "{}"
	}
	row, err := r.q.SubmitAnalysisJob(ctx, gianadb.SubmitAnalysisJobParams{
		ID:             mapper.PgUUID(p.ID),
		TenantID:       mapper.PgUUID(p.TenantID),
		Type:           int32(p.Type),
		Status:         int32(p.Status),
		InputUris:      p.InputURIs,
		ParametersJson: []byte(params),
		CreatedBy:      p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.AnalysisJobFromRow(row), nil
}

func (r *Repo) GetAnalysisJob(ctx context.Context, id ulid.ID) (*models.AnalysisJob, error) {
	row, err := r.q.GetAnalysisJob(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.AnalysisJobFromRow(row), nil
}

type ListAnalysisJobsParams struct {
	TenantID   ulid.ID
	Status     *models.AnalysisStatus
	Type       *models.AnalysisType
	PageOffset int32
	PageSize   int32
}

func (r *Repo) ListAnalysisJobsForTenant(ctx context.Context, p ListAnalysisJobsParams) ([]*models.AnalysisJob, int32, error) {
	var statusPtr, typePtr *int32
	if p.Status != nil {
		v := int32(*p.Status)
		statusPtr = &v
	}
	if p.Type != nil {
		v := int32(*p.Type)
		typePtr = &v
	}
	total, err := r.q.CountAnalysisJobsForTenant(ctx, gianadb.CountAnalysisJobsForTenantParams{
		TenantID: mapper.PgUUID(p.TenantID),
		Status:   statusPtr,
		Type:     typePtr,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListAnalysisJobsForTenant(ctx, gianadb.ListAnalysisJobsForTenantParams{
		TenantID:   mapper.PgUUID(p.TenantID),
		Status:     statusPtr,
		Type:       typePtr,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.AnalysisJob, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.AnalysisJobFromRow(row))
	}
	return out, int32(total), nil
}

type UpdateAnalysisJobStatusParams struct {
	ID                 ulid.ID
	Status             models.AnalysisStatus
	OutputURI          string
	ResultsSummaryJSON string
	ErrorMessage       string
	UpdatedBy          string
}

func (r *Repo) UpdateAnalysisJobStatus(ctx context.Context, p UpdateAnalysisJobStatusParams) (*models.AnalysisJob, error) {
	row, err := r.q.UpdateAnalysisJobStatus(ctx, gianadb.UpdateAnalysisJobStatusParams{
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
	return mapper.AnalysisJobFromRow(row), nil
}
