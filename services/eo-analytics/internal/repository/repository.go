// Package repository wraps the eo-analytics sqlc layer.
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/samavaya/packages/ulid"

	eoanalyticsdb "github.com/ppusapati/space/services/eo-analytics/db/generated"
	"github.com/ppusapati/space/services/eo-analytics/internal/mapper"
	"github.com/ppusapati/space/services/eo-analytics/internal/models"
)

// ErrNotFound is returned when no row matches.
var ErrNotFound = errors.New("repository: not found")

// Repo persists Models and InferenceJobs.
type Repo struct {
	q    *eoanalyticsdb.Queries
	pool *pgxpool.Pool
}

// New constructs a Repo.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: eoanalyticsdb.New(pool), pool: pool}
}

// ----- Models --------------------------------------------------------------

// RegisterModelParams holds the input for [Repo.RegisterModel].
type RegisterModelParams struct {
	ID           ulid.ID
	TenantID     ulid.ID
	Name         string
	Version      string
	Task         models.InferenceTask
	Framework    string
	ArtefactURI  string
	MetadataJSON string
	CreatedBy    string
}

// RegisterModel inserts a new model row.
func (r *Repo) RegisterModel(ctx context.Context, p RegisterModelParams) (*models.Model, error) {
	row, err := r.q.RegisterModel(ctx, eoanalyticsdb.RegisterModelParams{
		ID:           mapper.PgUUID(p.ID),
		TenantID:     mapper.PgUUID(p.TenantID),
		Name:         p.Name,
		Version:      p.Version,
		Task:         int32(p.Task),
		Framework:    p.Framework,
		ArtefactUri:  p.ArtefactURI,
		MetadataJson: []byte(mapper.NormalizeMetadataJSON(p.MetadataJSON)),
		CreatedBy:    p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.ModelFromRow(row), nil
}

// GetModel returns a model by id.
func (r *Repo) GetModel(ctx context.Context, id ulid.ID) (*models.Model, error) {
	row, err := r.q.GetModel(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.ModelFromRow(row), nil
}

// ListModelsParams holds the input for [Repo.ListModelsForTenant].
type ListModelsParams struct {
	TenantID   ulid.ID
	Task       *models.InferenceTask
	PageOffset int32
	PageSize   int32
}

// ListModelsForTenant returns one page of models.
func (r *Repo) ListModelsForTenant(ctx context.Context, p ListModelsParams) ([]*models.Model, int32, error) {
	var taskPtr *int32
	if p.Task != nil {
		v := int32(*p.Task)
		taskPtr = &v
	}
	total, err := r.q.CountModelsForTenant(ctx, eoanalyticsdb.CountModelsForTenantParams{
		TenantID: mapper.PgUUID(p.TenantID),
		Task:     taskPtr,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListModelsForTenant(ctx, eoanalyticsdb.ListModelsForTenantParams{
		TenantID:   mapper.PgUUID(p.TenantID),
		Task:       taskPtr,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.Model, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.ModelFromRow(row))
	}
	return out, int32(total), nil
}

// DeactivateModel deactivates a model.
func (r *Repo) DeactivateModel(ctx context.Context, id ulid.ID, updatedBy string) (*models.Model, error) {
	row, err := r.q.DeactivateModel(ctx, eoanalyticsdb.DeactivateModelParams{
		ID:        mapper.PgUUID(id),
		UpdatedBy: updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.ModelFromRow(row), nil
}

// ----- InferenceJobs -------------------------------------------------------

// CreateInferenceJobParams holds the input for [Repo.CreateInferenceJob].
type CreateInferenceJobParams struct {
	ID        ulid.ID
	TenantID  ulid.ID
	ModelID   ulid.ID
	ItemID    ulid.ID
	Status    models.InferenceJobStatus
	CreatedBy string
}

// CreateInferenceJob inserts a new inference job row.
func (r *Repo) CreateInferenceJob(ctx context.Context, p CreateInferenceJobParams) (*models.InferenceJob, error) {
	row, err := r.q.CreateInferenceJob(ctx, eoanalyticsdb.CreateInferenceJobParams{
		ID:        mapper.PgUUID(p.ID),
		TenantID:  mapper.PgUUID(p.TenantID),
		ModelID:   mapper.PgUUID(p.ModelID),
		ItemID:    mapper.PgUUID(p.ItemID),
		Status:    int32(p.Status),
		CreatedBy: p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.InferenceJobFromRow(row), nil
}

// GetInferenceJob returns an inference job by id.
func (r *Repo) GetInferenceJob(ctx context.Context, id ulid.ID) (*models.InferenceJob, error) {
	row, err := r.q.GetInferenceJob(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.InferenceJobFromRow(row), nil
}

// ListInferenceJobsParams holds the input for [Repo.ListInferenceJobsForTenant].
type ListInferenceJobsParams struct {
	TenantID   ulid.ID
	Status     *models.InferenceJobStatus
	PageOffset int32
	PageSize   int32
}

// ListInferenceJobsForTenant returns one page of inference jobs.
func (r *Repo) ListInferenceJobsForTenant(ctx context.Context, p ListInferenceJobsParams) ([]*models.InferenceJob, int32, error) {
	var statusPtr *int32
	if p.Status != nil {
		v := int32(*p.Status)
		statusPtr = &v
	}
	total, err := r.q.CountInferenceJobsForTenant(ctx, eoanalyticsdb.CountInferenceJobsForTenantParams{
		TenantID: mapper.PgUUID(p.TenantID),
		Status:   statusPtr,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListInferenceJobsForTenant(ctx, eoanalyticsdb.ListInferenceJobsForTenantParams{
		TenantID:   mapper.PgUUID(p.TenantID),
		Status:     statusPtr,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.InferenceJob, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.InferenceJobFromRow(row))
	}
	return out, int32(total), nil
}

// UpdateInferenceJobStatus updates the status of a job.
func (r *Repo) UpdateInferenceJobStatus(
	ctx context.Context, id ulid.ID, status models.InferenceJobStatus, outputURI, errorMessage, updatedBy string,
) (*models.InferenceJob, error) {
	row, err := r.q.UpdateInferenceJobStatus(ctx, eoanalyticsdb.UpdateInferenceJobStatusParams{
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
	return mapper.InferenceJobFromRow(row), nil
}
