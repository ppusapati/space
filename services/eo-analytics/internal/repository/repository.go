// Package repository wraps eo-analytics sqlc.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	eoanalyticsdb "github.com/ppusapati/space/services/eo-analytics/db/generated"
	"github.com/ppusapati/space/services/eo-analytics/internal/mappers"
	"github.com/ppusapati/space/services/eo-analytics/internal/models"
)

// ErrNotFound is returned when no row matches.
var ErrNotFound = errors.New("repository: not found")

// Repository aggregates the entity repositories.
type Repository struct {
	Models *ModelRepository
	Jobs   *InferenceJobRepository
}

// New constructs a Repository.
func New(pool *pgxpool.Pool) *Repository {
	q := eoanalyticsdb.New(pool)
	return &Repository{
		Models: &ModelRepository{q: q, pool: pool},
		Jobs:   &InferenceJobRepository{q: q, pool: pool},
	}
}

// ModelRepository persists Models.
type ModelRepository struct {
	q    *eoanalyticsdb.Queries
	pool *pgxpool.Pool
}

// Register inserts a new Model.
func (r *ModelRepository) Register(ctx context.Context, m *models.Model) (*models.Model, error) {
	row, err := r.q.RegisterModel(ctx, eoanalyticsdb.RegisterModelParams{
		ID:           mappers.PgUUID(m.ID),
		TenantID:     mappers.PgUUID(m.TenantID),
		Name:         m.Name,
		Version:      m.Version,
		Task:         int32(m.Task),
		Framework:    m.Framework,
		ArtefactUri:  m.ArtefactURI,
		MetadataJson: []byte(m.MetadataJSON),
		CreatedBy:    m.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mappers.ModelFromRow(row), nil
}

// Get fetches a Model by id.
func (r *ModelRepository) Get(ctx context.Context, id uuid.UUID) (*models.Model, error) {
	row, err := r.q.GetModel(ctx, mappers.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mappers.ModelFromRow(row), nil
}

// List returns one page of Models for the tenant.
func (r *ModelRepository) List(
	ctx context.Context, tenantID uuid.UUID, task *models.InferenceTask,
	cursorTS *time.Time, cursorID uuid.UUID, limit int32,
) ([]*models.Model, error) {
	var taskArg *int32
	if task != nil {
		v := int32(*task)
		taskArg = &v
	}
	rows, err := r.q.ListModels(ctx, eoanalyticsdb.ListModelsParams{
		TenantID:        mappers.PgUUID(tenantID),
		Task:            taskArg,
		CursorCreatedAt: mappers.PgTimestampPtr(cursorTS),
		CursorID:        mappers.PgUUID(cursorID),
		Lim:             limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*models.Model, 0, len(rows))
	for _, row := range rows {
		out = append(out, mappers.ModelFromRow(row))
	}
	return out, nil
}

// Deactivate marks a Model inactive.
func (r *ModelRepository) Deactivate(ctx context.Context, id uuid.UUID, updatedBy string) (*models.Model, error) {
	row, err := r.q.DeactivateModel(ctx, eoanalyticsdb.DeactivateModelParams{
		ID:        mappers.PgUUID(id),
		UpdatedBy: updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mappers.ModelFromRow(row), nil
}

// InferenceJobRepository persists InferenceJobs.
type InferenceJobRepository struct {
	q    *eoanalyticsdb.Queries
	pool *pgxpool.Pool
}

// Create inserts a new pending InferenceJob.
func (r *InferenceJobRepository) Create(ctx context.Context, j *models.InferenceJob) (*models.InferenceJob, error) {
	row, err := r.q.CreateInferenceJob(ctx, eoanalyticsdb.CreateInferenceJobParams{
		ID:        mappers.PgUUID(j.ID),
		TenantID:  mappers.PgUUID(j.TenantID),
		ModelID:   mappers.PgUUID(j.ModelID),
		ItemID:    mappers.PgUUID(j.ItemID),
		Status:    int32(j.Status),
		CreatedBy: j.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mappers.InferenceJobFromRow(row), nil
}

// Get returns by id.
func (r *InferenceJobRepository) Get(ctx context.Context, id uuid.UUID) (*models.InferenceJob, error) {
	row, err := r.q.GetInferenceJob(ctx, mappers.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mappers.InferenceJobFromRow(row), nil
}

// List returns one page of jobs filtered optionally by status.
func (r *InferenceJobRepository) List(
	ctx context.Context, tenantID uuid.UUID, status *models.InferenceJobStatus,
	cursorTS *time.Time, cursorID uuid.UUID, limit int32,
) ([]*models.InferenceJob, error) {
	var statusArg *int32
	if status != nil {
		v := int32(*status)
		statusArg = &v
	}
	rows, err := r.q.ListInferenceJobs(ctx, eoanalyticsdb.ListInferenceJobsParams{
		TenantID:        mappers.PgUUID(tenantID),
		Status:          statusArg,
		CursorCreatedAt: mappers.PgTimestampPtr(cursorTS),
		CursorID:        mappers.PgUUID(cursorID),
		Lim:             limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*models.InferenceJob, 0, len(rows))
	for _, row := range rows {
		out = append(out, mappers.InferenceJobFromRow(row))
	}
	return out, nil
}

// UpdateStatus transitions a job.
func (r *InferenceJobRepository) UpdateStatus(
	ctx context.Context, id uuid.UUID, status models.InferenceJobStatus,
	outputURI, errMsg, updatedBy string,
) (*models.InferenceJob, error) {
	row, err := r.q.UpdateInferenceJobStatus(ctx, eoanalyticsdb.UpdateInferenceJobStatusParams{
		ID:        mappers.PgUUID(id),
		Status:    int32(status),
		Column3:   outputURI,
		Column4:   errMsg,
		UpdatedBy: updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mappers.InferenceJobFromRow(row), nil
}
