// Package repository wraps the eo-pipeline sqlc layer.
package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/samavaya/packages/ulid"

	eopipelinedb "github.com/ppusapati/space/services/eo-pipeline/db/generated"
	"github.com/ppusapati/space/services/eo-pipeline/internal/mapper"
	"github.com/ppusapati/space/services/eo-pipeline/internal/models"
)

// ErrNotFound is returned when no row matches.
var ErrNotFound = errors.New("repository: not found")

// Repo persists Jobs.
type Repo struct {
	q    *eopipelinedb.Queries
	pool *pgxpool.Pool
}

// New constructs a Repo.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: eopipelinedb.New(pool), pool: pool}
}

// CreateJobParams holds the input for [Repo.CreateJob].
type CreateJobParams struct {
	ID             ulid.ID
	TenantID       ulid.ID
	ItemID         ulid.ID
	Stage          models.JobStage
	Status         models.JobStatus
	ParametersJSON string
	CreatedBy      string
}

// CreateJob inserts a new job row.
func (r *Repo) CreateJob(ctx context.Context, p CreateJobParams) (*models.Job, error) {
	params := strings.TrimSpace(p.ParametersJSON)
	if params == "" {
		params = "{}"
	}
	row, err := r.q.CreateJob(ctx, eopipelinedb.CreateJobParams{
		ID:             mapper.PgUUID(p.ID),
		TenantID:       mapper.PgUUID(p.TenantID),
		ItemID:         mapper.PgUUID(p.ItemID),
		Stage:          int32(p.Stage),
		Status:         int32(p.Status),
		ParametersJson: []byte(params),
		CreatedBy:      p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.JobFromRow(row), nil
}

// GetJob returns a job by id.
func (r *Repo) GetJob(ctx context.Context, id ulid.ID) (*models.Job, error) {
	row, err := r.q.GetJob(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.JobFromRow(row), nil
}

// ListJobsParams holds the input for [Repo.ListJobsForTenant].
type ListJobsParams struct {
	TenantID   ulid.ID
	Status     *models.JobStatus
	Stage      *models.JobStage
	PageOffset int32
	PageSize   int32
}

// ListJobsForTenant returns one page of jobs.
func (r *Repo) ListJobsForTenant(ctx context.Context, p ListJobsParams) ([]*models.Job, int32, error) {
	var statusPtr *int32
	if p.Status != nil {
		v := int32(*p.Status)
		statusPtr = &v
	}
	var stagePtr *int32
	if p.Stage != nil {
		v := int32(*p.Stage)
		stagePtr = &v
	}
	total, err := r.q.CountJobsForTenant(ctx, eopipelinedb.CountJobsForTenantParams{
		TenantID: mapper.PgUUID(p.TenantID),
		Status:   statusPtr,
		Stage:    stagePtr,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListJobsForTenant(ctx, eopipelinedb.ListJobsForTenantParams{
		TenantID:   mapper.PgUUID(p.TenantID),
		Status:     statusPtr,
		Stage:      stagePtr,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.Job, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.JobFromRow(row))
	}
	return out, int32(total), nil
}

// UpdateJobStatus updates the status of a job (and timestamp side-effects).
func (r *Repo) UpdateJobStatus(
	ctx context.Context, id ulid.ID, status models.JobStatus, outputURI, errorMessage, updatedBy string,
) (*models.Job, error) {
	row, err := r.q.UpdateJobStatus(ctx, eopipelinedb.UpdateJobStatusParams{
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
	return mapper.JobFromRow(row), nil
}
