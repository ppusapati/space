// Package services holds gi-reports business logic.
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	"github.com/ppusapati/space/services/gi-reports/internal/models"
	"github.com/ppusapati/space/services/gi-reports/internal/repository"
)

type Reports struct {
	repo  *repository.Repo
	IDFn  func() ulid.ID
	NowFn func() time.Time
}

func New(repo *repository.Repo) *Reports {
	return &Reports{
		repo:  repo,
		IDFn:  ulid.NewMonotonic,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// ----- Template -----------------------------------------------------------

type CreateTemplateInput struct {
	TenantID         ulid.ID
	Slug             string
	Name             string
	Description      string
	TemplateURI      string
	Format           models.ReportFormat
	ParametersSchema string
	CreatedBy        string
}

func (r *Reports) CreateTemplate(ctx context.Context, in CreateTemplateInput) (*models.ReportTemplate, error) {
	if in.TenantID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	in.Slug = strings.TrimSpace(in.Slug)
	in.Name = strings.TrimSpace(in.Name)
	if in.Slug == "" || in.Name == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "slug and name required")
	}
	if in.Format == models.FormatUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "format required")
	}
	if strings.TrimSpace(in.TemplateURI) == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "template_uri required")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return r.repo.CreateTemplate(ctx, repository.CreateTemplateParams{
		ID:               r.IDFn(),
		TenantID:         in.TenantID,
		Slug:             in.Slug,
		Name:             in.Name,
		Description:      in.Description,
		TemplateURI:      in.TemplateURI,
		Format:           in.Format,
		ParametersSchema: in.ParametersSchema,
		CreatedBy:        createdBy,
	})
}

func (r *Reports) GetTemplate(ctx context.Context, id ulid.ID) (*models.ReportTemplate, error) {
	t, err := r.repo.GetTemplate(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("TEMPLATE_NOT_FOUND", "report_template "+id.String())
	}
	return t, err
}

type ListTemplatesInput struct {
	TenantID   ulid.ID
	Format     *models.ReportFormat
	PageOffset int32
	PageSize   int32
}

func (r *Reports) ListTemplatesForTenant(ctx context.Context, in ListTemplatesInput) ([]*models.ReportTemplate, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := r.repo.ListTemplatesForTenant(ctx, repository.ListTemplatesParams{
		TenantID:   in.TenantID,
		Format:     in.Format,
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

func (r *Reports) DeprecateTemplate(ctx context.Context, id ulid.ID, updatedBy string) (*models.ReportTemplate, error) {
	if updatedBy == "" {
		updatedBy = "system"
	}
	t, err := r.repo.DeprecateTemplate(ctx, id, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("TEMPLATE_NOT_FOUND", "report_template "+id.String())
	}
	return t, err
}

// ----- Report -------------------------------------------------------------

type GenerateReportInput struct {
	TenantID       ulid.ID
	TemplateID     ulid.ID
	ParametersJSON string
	CreatedBy      string
}

func (r *Reports) GenerateReport(ctx context.Context, in GenerateReportInput) (*models.Report, error) {
	if in.TenantID.IsZero() || in.TemplateID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id and template_id required")
	}
	tpl, err := r.repo.GetTemplate(ctx, in.TemplateID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, pkgerrors.BadRequest("TEMPLATE_NOT_FOUND",
				"report_template "+in.TemplateID.String())
		}
		return nil, err
	}
	if tpl.TenantID != in.TenantID {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "template tenant mismatch")
	}
	if !tpl.Active {
		return nil, pkgerrors.New(412, "TEMPLATE_DEPRECATED", "report_template is deprecated")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return r.repo.GenerateReport(ctx, repository.GenerateReportParams{
		ID:             r.IDFn(),
		TenantID:       in.TenantID,
		TemplateID:     in.TemplateID,
		Status:         models.StatusQueued,
		ParametersJSON: in.ParametersJSON,
		CreatedBy:      createdBy,
	})
}

func (r *Reports) GetReport(ctx context.Context, id ulid.ID) (*models.Report, error) {
	rep, err := r.repo.GetReport(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("REPORT_NOT_FOUND", "report "+id.String())
	}
	return rep, err
}

type ListReportsInput struct {
	TenantID   ulid.ID
	TemplateID *ulid.ID
	Status     *models.ReportStatus
	PageOffset int32
	PageSize   int32
}

func (r *Reports) ListReportsForTenant(ctx context.Context, in ListReportsInput) ([]*models.Report, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := r.repo.ListReportsForTenant(ctx, repository.ListReportsParams{
		TenantID:   in.TenantID,
		TemplateID: in.TemplateID,
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

func (r *Reports) UpdateReportStatus(
	ctx context.Context, id ulid.ID, status models.ReportStatus, outputURI, errorMessage, updatedBy string,
) (*models.Report, error) {
	if status == models.StatusUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "status required")
	}
	current, err := r.GetReport(ctx, id)
	if err != nil {
		return nil, err
	}
	if !validReportTransition(current.Status, status) {
		return nil, pkgerrors.New(412, "ILLEGAL_TRANSITION",
			"illegal report status transition")
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	updated, err := r.repo.UpdateReportStatus(ctx, id, status, outputURI, errorMessage, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("REPORT_NOT_FOUND", "report "+id.String())
	}
	return updated, err
}

func validReportTransition(from, to models.ReportStatus) bool {
	switch from {
	case models.StatusQueued:
		return to == models.StatusGenerating || to == models.StatusCanceled
	case models.StatusGenerating:
		return to == models.StatusCompleted || to == models.StatusFailed || to == models.StatusCanceled
	case models.StatusCompleted, models.StatusFailed, models.StatusCanceled:
		return false
	default:
		return false
	}
}
