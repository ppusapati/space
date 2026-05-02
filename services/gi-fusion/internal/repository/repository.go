// Package repository wraps the gi-fusion sqlc layer.
package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/chetana/packages/ulid"

	gifudb "github.com/ppusapati/space/services/gi-fusion/db/generated"
	"github.com/ppusapati/space/services/gi-fusion/internal/mapper"
	"github.com/ppusapati/space/services/gi-fusion/internal/models"
)

var ErrNotFound = errors.New("repository: not found")

type Repo struct {
	q    *gifudb.Queries
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: gifudb.New(pool), pool: pool}
}

type SubmitFusionJobParams struct {
	ID             ulid.ID
	TenantID       ulid.ID
	Method         models.FusionMethod
	Status         models.FusionStatus
	InputURIs      []string
	ParametersJSON string
	CreatedBy      string
}

func (r *Repo) SubmitFusionJob(ctx context.Context, p SubmitFusionJobParams) (*models.FusionJob, error) {
	params := strings.TrimSpace(p.ParametersJSON)
	if params == "" {
		params = "{}"
	}
	row, err := r.q.SubmitFusionJob(ctx, gifudb.SubmitFusionJobParams{
		ID:             mapper.PgUUID(p.ID),
		TenantID:       mapper.PgUUID(p.TenantID),
		Method:         int32(p.Method),
		Status:         int32(p.Status),
		InputUris:      p.InputURIs,
		ParametersJson: []byte(params),
		CreatedBy:      p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.FusionJobFromRow(row), nil
}

func (r *Repo) GetFusionJob(ctx context.Context, id ulid.ID) (*models.FusionJob, error) {
	row, err := r.q.GetFusionJob(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.FusionJobFromRow(row), nil
}

type ListFusionJobsParams struct {
	TenantID   ulid.ID
	Status     *models.FusionStatus
	Method     *models.FusionMethod
	PageOffset int32
	PageSize   int32
}

func (r *Repo) ListFusionJobsForTenant(ctx context.Context, p ListFusionJobsParams) ([]*models.FusionJob, int32, error) {
	var statusPtr, methodPtr *int32
	if p.Status != nil {
		v := int32(*p.Status)
		statusPtr = &v
	}
	if p.Method != nil {
		v := int32(*p.Method)
		methodPtr = &v
	}
	total, err := r.q.CountFusionJobsForTenant(ctx, gifudb.CountFusionJobsForTenantParams{
		TenantID: mapper.PgUUID(p.TenantID),
		Status:   statusPtr,
		Method:   methodPtr,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListFusionJobsForTenant(ctx, gifudb.ListFusionJobsForTenantParams{
		TenantID:   mapper.PgUUID(p.TenantID),
		Status:     statusPtr,
		Method:     methodPtr,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.FusionJob, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.FusionJobFromRow(row))
	}
	return out, int32(total), nil
}

func (r *Repo) UpdateFusionJobStatus(
	ctx context.Context, id ulid.ID, status models.FusionStatus, outputURI, errorMessage, updatedBy string,
) (*models.FusionJob, error) {
	row, err := r.q.UpdateFusionJobStatus(ctx, gifudb.UpdateFusionJobStatusParams{
		ID:           mapper.PgUUID(id),
		Status:       int32(status),
		OutputUri:    outputURI,
		ErrorMessage: errorMessage,
		UpdatedBy:    updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.FusionJobFromRow(row), nil
}
