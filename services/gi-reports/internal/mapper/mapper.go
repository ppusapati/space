// Package mapper converts proto / domain / sqlc types for gi-reports.
package mapper

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"p9e.in/chetana/packages/api/v1/fields"
	"p9e.in/chetana/packages/api/v1/pagination"
	"p9e.in/chetana/packages/ulid"

	pbre "github.com/ppusapati/space/services/gi-reports/api"
	giredb "github.com/ppusapati/space/services/gi-reports/db/generated"
	"github.com/ppusapati/space/services/gi-reports/internal/models"
)

func PgUUID(id ulid.ID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}
func PgUUIDOrNull(id ulid.ID) pgtype.UUID {
	if id.IsZero() {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}
func FromPgUUID(p pgtype.UUID) ulid.ID {
	if !p.Valid {
		return ulid.Zero
	}
	return ulid.ID(p.Bytes)
}

func PageRequest(p *pagination.PaginationRequest) (offset, size int32) {
	if p == nil {
		return 0, 50
	}
	offset = p.GetPageOffset()
	if offset < 0 {
		offset = 0
	}
	size = p.GetPageSize()
	switch {
	case size <= 0:
		size = 50
	case size > 500:
		size = 500
	}
	return offset, size
}

func PageResponse(p models.Page) *pagination.PaginationResponse {
	return &pagination.PaginationResponse{
		TotalCount: p.TotalCount,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
		HasNext:    p.HasNext,
	}
}

func FieldsToProto(id ulid.ID, createdBy string, createdAt time.Time, updatedBy string, updatedAt time.Time) *fields.Fields {
	out := &fields.Fields{Uuid: id.String(), IsActive: true}
	if createdBy != "" {
		out.CreatedBy = wrapperspb.String(createdBy)
	}
	if !createdAt.IsZero() {
		out.CreatedAt = timestamppb.New(createdAt)
	}
	if updatedBy != "" {
		out.UpdatedBy = wrapperspb.String(updatedBy)
	}
	if !updatedAt.IsZero() {
		out.UpdatedAt = timestamppb.New(updatedAt)
	}
	return out
}

// ----- ReportTemplate -----------------------------------------------------

func TemplateFromRow(row giredb.ReportTemplate) *models.ReportTemplate {
	return &models.ReportTemplate{
		ID:               FromPgUUID(row.ID),
		TenantID:         FromPgUUID(row.TenantID),
		Slug:             row.Slug,
		Name:             row.Name,
		Description:      row.Description,
		TemplateURI:      row.TemplateUri,
		Format:           models.ReportFormat(row.Format),
		ParametersSchema: row.ParametersSchema,
		Active:           row.Active,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
		CreatedBy:        row.CreatedBy,
		UpdatedBy:        row.UpdatedBy,
	}
}

func TemplateToProto(t *models.ReportTemplate) *pbre.ReportTemplate {
	if t == nil {
		return nil
	}
	return &pbre.ReportTemplate{
		Id:               t.ID.String(),
		TenantId:         t.TenantID.String(),
		Slug:             t.Slug,
		Name:             t.Name,
		Description:      t.Description,
		TemplateUri:      t.TemplateURI,
		Format:           pbre.ReportFormat(t.Format),
		ParametersSchema: t.ParametersSchema,
		Active:           t.Active,
		Fields:           FieldsToProto(t.ID, t.CreatedBy, t.CreatedAt, t.UpdatedBy, t.UpdatedAt),
	}
}

// ----- Report -------------------------------------------------------------

func ReportFromRow(row giredb.Report) *models.Report {
	r := &models.Report{
		ID:             FromPgUUID(row.ID),
		TenantID:       FromPgUUID(row.TenantID),
		TemplateID:     FromPgUUID(row.TemplateID),
		Status:         models.ReportStatus(row.Status),
		ParametersJSON: string(row.ParametersJson),
		OutputURI:      row.OutputUri,
		ErrorMessage:   row.ErrorMessage,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
		CreatedBy:      row.CreatedBy,
		UpdatedBy:      row.UpdatedBy,
	}
	if row.StartedAt.Valid {
		r.StartedAt = row.StartedAt.Time
	}
	if row.FinishedAt.Valid {
		r.FinishedAt = row.FinishedAt.Time
	}
	return r
}

func ReportToProto(r *models.Report) *pbre.Report {
	if r == nil {
		return nil
	}
	out := &pbre.Report{
		Id:             r.ID.String(),
		TenantId:       r.TenantID.String(),
		TemplateId:     r.TemplateID.String(),
		Status:         pbre.ReportStatus(r.Status),
		ParametersJson: r.ParametersJSON,
		OutputUri:      r.OutputURI,
		ErrorMessage:   r.ErrorMessage,
		Fields:         FieldsToProto(r.ID, r.CreatedBy, r.CreatedAt, r.UpdatedBy, r.UpdatedAt),
	}
	if !r.StartedAt.IsZero() {
		out.StartedAt = timestamppb.New(r.StartedAt)
	}
	if !r.FinishedAt.IsZero() {
		out.FinishedAt = timestamppb.New(r.FinishedAt)
	}
	return out
}
