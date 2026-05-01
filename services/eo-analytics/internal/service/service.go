// Package service holds eo-analytics business logic.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/ppusapati/space/pkg/errs"
	"github.com/ppusapati/space/services/eo-analytics/internal/models"
	"github.com/ppusapati/space/services/eo-analytics/internal/repository"
)

// Analytics is the eo-analytics service-layer facade.
type Analytics struct {
	repo  *repository.Repository
	IDFn  func() uuid.UUID
	NowFn func() time.Time
}

// New constructs an Analytics service.
func New(repo *repository.Repository) *Analytics {
	return &Analytics{
		repo:  repo,
		IDFn:  uuid.New,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// RegisterModelInput is the input to [Analytics.RegisterModel].
type RegisterModelInput struct {
	TenantID     uuid.UUID
	Name         string
	Version      string
	Task         models.InferenceTask
	Framework    string
	ArtefactURI  string
	MetadataJSON string
	CreatedBy    string
}

// RegisterModel persists a new Model.
func (a *Analytics) RegisterModel(ctx context.Context, in RegisterModelInput) (*models.Model, error) {
	if in.TenantID == uuid.Nil || in.Name == "" || in.Version == "" {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id, name, version required")
	}
	if in.Task == models.TaskUnspecified {
		return nil, errs.New(errs.DomainInvalidArgument, "task required")
	}
	if in.Framework == "" || in.ArtefactURI == "" {
		return nil, errs.New(errs.DomainInvalidArgument, "framework and artefact_uri required")
	}
	if in.MetadataJSON == "" {
		in.MetadataJSON = "{}"
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	m := &models.Model{
		ID:           a.IDFn(),
		TenantID:     in.TenantID,
		Name:         in.Name,
		Version:      in.Version,
		Task:         in.Task,
		Framework:    in.Framework,
		ArtefactURI:  in.ArtefactURI,
		MetadataJSON: in.MetadataJSON,
		Active:       true,
		CreatedBy:    createdBy,
	}
	return a.repo.Models.Register(ctx, m)
}

// GetModel fetches a Model.
func (a *Analytics) GetModel(ctx context.Context, id uuid.UUID) (*models.Model, error) {
	m, err := a.repo.Models.Get(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errs.New(errs.DomainNotFound, "model %s", id)
	}
	return m, err
}

// ListModelsInput is the input to [Analytics.ListModels].
type ListModelsInput struct {
	TenantID      uuid.UUID
	Task          *models.InferenceTask
	CursorCreated *time.Time
	CursorID      uuid.UUID
	Limit         int32
}

// ListModels returns one page of Models.
func (a *Analytics) ListModels(ctx context.Context, in ListModelsInput) ([]*models.Model, error) {
	if in.TenantID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id required")
	}
	return a.repo.Models.List(ctx, in.TenantID, in.Task, in.CursorCreated, in.CursorID, in.Limit)
}

// DeactivateModel marks a Model inactive.
func (a *Analytics) DeactivateModel(ctx context.Context, id uuid.UUID, updatedBy string) (*models.Model, error) {
	if updatedBy == "" {
		updatedBy = "system"
	}
	m, err := a.repo.Models.Deactivate(ctx, id, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errs.New(errs.DomainNotFound, "model %s", id)
	}
	return m, err
}

// SubmitInferenceJobInput is the input to [Analytics.SubmitInferenceJob].
type SubmitInferenceJobInput struct {
	TenantID  uuid.UUID
	ModelID   uuid.UUID
	ItemID    uuid.UUID
	CreatedBy string
}

// SubmitInferenceJob persists a new pending InferenceJob. The model
// must exist, belong to the same tenant, and be active.
func (a *Analytics) SubmitInferenceJob(ctx context.Context, in SubmitInferenceJobInput) (*models.InferenceJob, error) {
	if in.TenantID == uuid.Nil || in.ModelID == uuid.Nil || in.ItemID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id, model_id, item_id required")
	}
	m, err := a.GetModel(ctx, in.ModelID)
	if err != nil {
		return nil, err
	}
	if m.TenantID != in.TenantID {
		return nil, errs.New(errs.DomainPermissionDenied, "model belongs to a different tenant")
	}
	if !m.Active {
		return nil, errs.New(errs.DomainPreconditionFailed, "model is deactivated")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	j := &models.InferenceJob{
		ID:        a.IDFn(),
		TenantID:  in.TenantID,
		ModelID:   in.ModelID,
		ItemID:    in.ItemID,
		Status:    models.StatusPending,
		CreatedBy: createdBy,
	}
	return a.repo.Jobs.Create(ctx, j)
}

// GetInferenceJob fetches by id.
func (a *Analytics) GetInferenceJob(ctx context.Context, id uuid.UUID) (*models.InferenceJob, error) {
	j, err := a.repo.Jobs.Get(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errs.New(errs.DomainNotFound, "inference_job %s", id)
	}
	return j, err
}

// ListInferenceJobsInput is the input to [Analytics.ListInferenceJobs].
type ListInferenceJobsInput struct {
	TenantID      uuid.UUID
	Status        *models.InferenceJobStatus
	CursorCreated *time.Time
	CursorID      uuid.UUID
	Limit         int32
}

// ListInferenceJobs returns one page of InferenceJobs.
func (a *Analytics) ListInferenceJobs(ctx context.Context, in ListInferenceJobsInput) ([]*models.InferenceJob, error) {
	if in.TenantID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id required")
	}
	return a.repo.Jobs.List(ctx, in.TenantID, in.Status, in.CursorCreated, in.CursorID, in.Limit)
}

// UpdateInferenceJobStatus transitions an InferenceJob.
func (a *Analytics) UpdateInferenceJobStatus(
	ctx context.Context, id uuid.UUID, target models.InferenceJobStatus,
	outputURI, errMsg, updatedBy string,
) (*models.InferenceJob, error) {
	cur, err := a.GetInferenceJob(ctx, id)
	if err != nil {
		return nil, err
	}
	if cur.IsTerminal() {
		return nil, errs.New(errs.DomainPreconditionFailed,
			"cannot transition terminal inference job %s (status=%d)", id, cur.Status)
	}
	if target == models.StatusUnspecified {
		return nil, errs.New(errs.DomainInvalidArgument, "target status required")
	}
	if target == models.StatusFailed && errMsg == "" {
		return nil, errs.New(errs.DomainInvalidArgument, "error_message required when failing")
	}
	if target == models.StatusSucceeded && outputURI == "" {
		return nil, errs.New(errs.DomainInvalidArgument, "output_uri required when succeeding")
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	return a.repo.Jobs.UpdateStatus(ctx, id, target, outputURI, errMsg, updatedBy)
}
