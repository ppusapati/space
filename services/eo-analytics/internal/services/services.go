// Package services holds eo-analytics business logic.
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	"github.com/ppusapati/space/services/eo-analytics/internal/models"
	"github.com/ppusapati/space/services/eo-analytics/internal/repository"
)

// Analytics is the eo-analytics service-layer facade.
type Analytics struct {
	repo  *repository.Repo
	IDFn  func() ulid.ID
	NowFn func() time.Time
}

// New constructs an Analytics service.
func New(repo *repository.Repo) *Analytics {
	return &Analytics{
		repo:  repo,
		IDFn:  ulid.NewMonotonic,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// ----- Models --------------------------------------------------------------

// RegisterModelInput is the input for [Analytics.RegisterModel].
type RegisterModelInput struct {
	TenantID     ulid.ID
	Name         string
	Version      string
	Task         models.InferenceTask
	Framework    string
	ArtefactURI  string
	MetadataJSON string
	CreatedBy    string
}

// RegisterModel persists a new model.
func (a *Analytics) RegisterModel(ctx context.Context, in RegisterModelInput) (*models.Model, error) {
	if in.TenantID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Version = strings.TrimSpace(in.Version)
	in.Framework = strings.TrimSpace(in.Framework)
	in.ArtefactURI = strings.TrimSpace(in.ArtefactURI)
	if in.Name == "" || in.Version == "" || in.Framework == "" || in.ArtefactURI == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "name, version, framework, artefact_uri required")
	}
	if in.Task == models.TaskUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "task required")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return a.repo.RegisterModel(ctx, repository.RegisterModelParams{
		ID:           a.IDFn(),
		TenantID:     in.TenantID,
		Name:         in.Name,
		Version:      in.Version,
		Task:         in.Task,
		Framework:    in.Framework,
		ArtefactURI:  in.ArtefactURI,
		MetadataJSON: in.MetadataJSON,
		CreatedBy:    createdBy,
	})
}

// GetModel fetches a model by id.
func (a *Analytics) GetModel(ctx context.Context, id ulid.ID) (*models.Model, error) {
	m, err := a.repo.GetModel(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("MODEL_NOT_FOUND", "model "+id.String())
	}
	return m, err
}

// ListModelsInput is the input for [Analytics.ListModelsForTenant].
type ListModelsInput struct {
	TenantID   ulid.ID
	Task       *models.InferenceTask
	PageOffset int32
	PageSize   int32
}

// ListModelsForTenant returns one page of models.
func (a *Analytics) ListModelsForTenant(ctx context.Context, in ListModelsInput) ([]*models.Model, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := a.repo.ListModelsForTenant(ctx, repository.ListModelsParams{
		TenantID:   in.TenantID,
		Task:       in.Task,
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

// DeactivateModel deactivates a model.
func (a *Analytics) DeactivateModel(ctx context.Context, id ulid.ID, updatedBy string) (*models.Model, error) {
	if id.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "id required")
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	m, err := a.repo.DeactivateModel(ctx, id, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("MODEL_NOT_FOUND", "model "+id.String())
	}
	return m, err
}

// ----- InferenceJobs -------------------------------------------------------

// SubmitInferenceJobInput is the input for [Analytics.SubmitInferenceJob].
type SubmitInferenceJobInput struct {
	TenantID  ulid.ID
	ModelID   ulid.ID
	ItemID    ulid.ID
	CreatedBy string
}

// SubmitInferenceJob persists a new queued inference job. Verifies the model
// exists, belongs to the tenant, and is active.
func (a *Analytics) SubmitInferenceJob(ctx context.Context, in SubmitInferenceJobInput) (*models.InferenceJob, error) {
	if in.TenantID.IsZero() || in.ModelID.IsZero() || in.ItemID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id, model_id, item_id required")
	}
	m, err := a.repo.GetModel(ctx, in.ModelID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, pkgerrors.BadRequest("MODEL_NOT_FOUND", "model "+in.ModelID.String())
		}
		return nil, err
	}
	if m.TenantID != in.TenantID {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "model tenant mismatch")
	}
	if !m.Active {
		return nil, pkgerrors.New(412, "MODEL_INACTIVE", "model is not active")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return a.repo.CreateInferenceJob(ctx, repository.CreateInferenceJobParams{
		ID:        a.IDFn(),
		TenantID:  in.TenantID,
		ModelID:   in.ModelID,
		ItemID:    in.ItemID,
		Status:    models.StatusQueued,
		CreatedBy: createdBy,
	})
}

// GetInferenceJob fetches an inference job by id.
func (a *Analytics) GetInferenceJob(ctx context.Context, id ulid.ID) (*models.InferenceJob, error) {
	j, err := a.repo.GetInferenceJob(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("INFERENCE_JOB_NOT_FOUND", "inference_job "+id.String())
	}
	return j, err
}

// ListInferenceJobsInput is the input for [Analytics.ListInferenceJobsForTenant].
type ListInferenceJobsInput struct {
	TenantID   ulid.ID
	Status     *models.InferenceJobStatus
	PageOffset int32
	PageSize   int32
}

// ListInferenceJobsForTenant returns one page of jobs.
func (a *Analytics) ListInferenceJobsForTenant(ctx context.Context, in ListInferenceJobsInput) ([]*models.InferenceJob, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := a.repo.ListInferenceJobsForTenant(ctx, repository.ListInferenceJobsParams{
		TenantID:   in.TenantID,
		Status:     in.Status,
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

// UpdateInferenceJobStatus updates the status of an inference job. Validates
// the transition graph.
func (a *Analytics) UpdateInferenceJobStatus(
	ctx context.Context, id ulid.ID, status models.InferenceJobStatus, outputURI, errorMessage, updatedBy string,
) (*models.InferenceJob, error) {
	if status == models.StatusUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "status required")
	}
	current, err := a.GetInferenceJob(ctx, id)
	if err != nil {
		return nil, err
	}
	if !validTransition(current.Status, status) {
		return nil, pkgerrors.New(412, "ILLEGAL_TRANSITION",
			"illegal inference_job status transition")
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	updated, err := a.repo.UpdateInferenceJobStatus(ctx, id, status, outputURI, errorMessage, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("INFERENCE_JOB_NOT_FOUND", "inference_job "+id.String())
	}
	return updated, err
}

// validTransition enforces the inference job status transition graph:
//
//	QUEUED    -> RUNNING | FAILED
//	RUNNING   -> SUCCEEDED | FAILED
//	SUCCEEDED, FAILED — terminal.
func validTransition(from, to models.InferenceJobStatus) bool {
	switch from {
	case models.StatusQueued:
		return to == models.StatusRunning || to == models.StatusFailed
	case models.StatusRunning:
		return to == models.StatusSucceeded || to == models.StatusFailed
	case models.StatusSucceeded, models.StatusFailed:
		return false
	default:
		return false
	}
}
