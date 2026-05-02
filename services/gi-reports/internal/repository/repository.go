// Package repository wraps the gi-reports sqlc layer.
package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/chetana/packages/ulid"

	giredb "github.com/ppusapati/space/services/gi-reports/db/generated"
	"github.com/ppusapati/space/services/gi-reports/internal/mapper"
	"github.com/ppusapati/space/services/gi-reports/internal/models"
)

var ErrNotFound = errors.New("repository: not found")

type Repo struct {
	q    *giredb.Queries
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: giredb.New(pool), pool: pool}
}

// ----- Template -----------------------------------------------------------

type CreateTemplateParams struct {
	ID               ulid.ID
	TenantID         ulid.ID
	Slug             string
	Name             string
	Description      string
	TemplateURI      string
	Format           models.ReportFormat
	ParametersSchema string
	CreatedBy        string
}

func (r *Repo) CreateTemplate(ctx context.Context, p CreateTemplateParams) (*models.ReportTemplate, error) {
	row, err := r.q.CreateTemplate(ctx, giredb.CreateTemplateParams{
		ID:               mapper.PgUUID(p.ID),
		TenantID:         mapper.PgUUID(p.TenantID),
		Slug:             p.Slug,
		Name:             p.Name,
		Description:      p.Description,
		TemplateUri:      p.TemplateURI,
		Format:           int32(p.Format),
		ParametersSchema: p.ParametersSchema,
		CreatedBy:        p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.TemplateFromRow(row), nil
}

func (r *Repo) GetTemplate(ctx context.Context, id ulid.ID) (*models.ReportTemplate, error) {
	row, err := r.q.GetTemplate(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.TemplateFromRow(row), nil
}

type ListTemplatesParams struct {
	TenantID   ulid.ID
	Format     *models.ReportFormat
	PageOffset int32
	PageSize   int32
}

func (r *Repo) ListTemplatesForTenant(ctx context.Context, p ListTemplatesParams) ([]*models.ReportTemplate, int32, error) {
	var formatPtr *int32
	if p.Format != nil {
		v := int32(*p.Format)
		formatPtr = &v
	}
	total, err := r.q.CountTemplatesForTenant(ctx, giredb.CountTemplatesForTenantParams{
		TenantID: mapper.PgUUID(p.TenantID),
		Format:   formatPtr,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListTemplatesForTenant(ctx, giredb.ListTemplatesForTenantParams{
		TenantID:   mapper.PgUUID(p.TenantID),
		Format:     formatPtr,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.ReportTemplate, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.TemplateFromRow(row))
	}
	return out, int32(total), nil
}

func (r *Repo) DeprecateTemplate(ctx context.Context, id ulid.ID, updatedBy string) (*models.ReportTemplate, error) {
	row, err := r.q.DeprecateTemplate(ctx, giredb.DeprecateTemplateParams{
		ID:        mapper.PgUUID(id),
		UpdatedBy: updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.TemplateFromRow(row), nil
}

// ----- Report -------------------------------------------------------------

type GenerateReportParams struct {
	ID             ulid.ID
	TenantID       ulid.ID
	TemplateID     ulid.ID
	Status         models.ReportStatus
	ParametersJSON string
	CreatedBy      string
}

func (r *Repo) GenerateReport(ctx context.Context, p GenerateReportParams) (*models.Report, error) {
	params := strings.TrimSpace(p.ParametersJSON)
	if params == "" {
		params = "{}"
	}
	row, err := r.q.GenerateReport(ctx, giredb.GenerateReportParams{
		ID:             mapper.PgUUID(p.ID),
		TenantID:       mapper.PgUUID(p.TenantID),
		TemplateID:     mapper.PgUUID(p.TemplateID),
		Status:         int32(p.Status),
		ParametersJson: []byte(params),
		CreatedBy:      p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.ReportFromRow(row), nil
}

func (r *Repo) GetReport(ctx context.Context, id ulid.ID) (*models.Report, error) {
	row, err := r.q.GetReport(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.ReportFromRow(row), nil
}

type ListReportsParams struct {
	TenantID   ulid.ID
	TemplateID *ulid.ID
	Status     *models.ReportStatus
	PageOffset int32
	PageSize   int32
}

func (r *Repo) ListReportsForTenant(ctx context.Context, p ListReportsParams) ([]*models.Report, int32, error) {
	var statusPtr *int32
	if p.Status != nil {
		v := int32(*p.Status)
		statusPtr = &v
	}
	var templatePg pgtype.UUID
	if p.TemplateID != nil {
		templatePg = mapper.PgUUID(*p.TemplateID)
	}
	total, err := r.q.CountReportsForTenant(ctx, giredb.CountReportsForTenantParams{
		TenantID:   mapper.PgUUID(p.TenantID),
		TemplateID: templatePg,
		Status:     statusPtr,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListReportsForTenant(ctx, giredb.ListReportsForTenantParams{
		TenantID:   mapper.PgUUID(p.TenantID),
		TemplateID: templatePg,
		Status:     statusPtr,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.Report, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.ReportFromRow(row))
	}
	return out, int32(total), nil
}

func (r *Repo) UpdateReportStatus(
	ctx context.Context, id ulid.ID, status models.ReportStatus, outputURI, errorMessage, updatedBy string,
) (*models.Report, error) {
	row, err := r.q.UpdateReportStatus(ctx, giredb.UpdateReportStatusParams{
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
	return mapper.ReportFromRow(row), nil
}
