// Package repository wraps the eo-pipeline sqlc layer.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	eopipelinedb "github.com/ppusapati/space/services/eo-pipeline/db/generated"
	"github.com/ppusapati/space/services/eo-pipeline/internal/mappers"
	"github.com/ppusapati/space/services/eo-pipeline/internal/models"
)

// ErrNotFound is returned when no row matches.
var ErrNotFound = errors.New("repository: not found")

// JobRepository persists Jobs.
type JobRepository struct {
	q    *eopipelinedb.Queries
	pool *pgxpool.Pool
}

// NewJobRepository constructs a JobRepository.
func NewJobRepository(pool *pgxpool.Pool) *JobRepository {
	return &JobRepository{q: eopipelinedb.New(pool), pool: pool}
}

// Create inserts a new Job.
func (r *JobRepository) Create(ctx context.Context, j *models.Job) (*models.Job, error) {
	row, err := r.q.CreateJob(ctx, eopipelinedb.CreateJobParams{
		ID:             mappers.PgUUID(j.ID),
		TenantID:       mappers.PgUUID(j.TenantID),
		ItemID:         mappers.PgUUID(j.ItemID),
		Stage:          int32(j.Stage),
		Status:         int32(j.Status),
		ParametersJson: []byte(j.ParametersJSON),
		CreatedBy:      j.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mappers.JobFromRow(row), nil
}

// Get fetches a Job by id.
func (r *JobRepository) Get(ctx context.Context, id uuid.UUID) (*models.Job, error) {
	row, err := r.q.GetJob(ctx, mappers.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mappers.JobFromRow(row), nil
}

// ListParams groups the List inputs.
type ListParams struct {
	TenantID       uuid.UUID
	Status         *models.JobStatus
	Stage          *models.JobStage
	CursorCreated  *time.Time
	CursorID       uuid.UUID
	Limit          int32
}

// List returns one page of jobs newest-first.
func (r *JobRepository) List(ctx context.Context, p ListParams) ([]*models.Job, error) {
	var statusArg, stageArg *int32
	if p.Status != nil {
		v := int32(*p.Status)
		statusArg = &v
	}
	if p.Stage != nil {
		v := int32(*p.Stage)
		stageArg = &v
	}
	rows, err := r.q.ListJobs(ctx, eopipelinedb.ListJobsParams{
		TenantID:        mappers.PgUUID(p.TenantID),
		Status:          statusArg,
		Stage:           stageArg,
		CursorCreatedAt: pgtypeTSPtr(p.CursorCreated),
		CursorID:        mappers.PgUUID(p.CursorID),
		Lim:             p.Limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*models.Job, 0, len(rows))
	for _, row := range rows {
		out = append(out, mappers.JobFromRow(row))
	}
	return out, nil
}

// UpdateStatus transitions a job and stamps timestamps.
func (r *JobRepository) UpdateStatus(
	ctx context.Context, id uuid.UUID, status models.JobStatus,
	outputURI, errMsg, updatedBy string,
) (*models.Job, error) {
	row, err := r.q.UpdateJobStatus(ctx, eopipelinedb.UpdateJobStatusParams{
		ID:           mappers.PgUUID(id),
		Status:       int32(status),
		Column3:      outputURI,
		Column4:      errMsg,
		UpdatedBy:    updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mappers.JobFromRow(row), nil
}

func pgtypeTSPtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}
