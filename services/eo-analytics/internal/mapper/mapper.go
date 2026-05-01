// Package mapper converts proto / domain / sqlc types for eo-analytics.
package mapper

import (
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"p9e.in/samavaya/packages/api/v1/fields"
	"p9e.in/samavaya/packages/api/v1/pagination"
	"p9e.in/samavaya/packages/ulid"

	pban "github.com/ppusapati/space/services/eo-analytics/api"
	eoanalyticsdb "github.com/ppusapati/space/services/eo-analytics/db/generated"
	"github.com/ppusapati/space/services/eo-analytics/internal/models"
)

// PgUUID converts a ulid.ID into a pgtype.UUID payload (16 bytes).
func PgUUID(id ulid.ID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

// FromPgUUID extracts ulid.ID from a pgtype.UUID. Returns ulid.Zero on NULL.
func FromPgUUID(p pgtype.UUID) ulid.ID {
	if !p.Valid {
		return ulid.Zero
	}
	return ulid.ID(p.Bytes)
}

// PgTimestamp converts time.Time to pgtype.Timestamptz.
func PgTimestamp(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// PageRequest extracts (offset, size) from a PaginationRequest, applying
// defaults: page_size defaults to 50, capped at 500; offset defaults to 0.
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

// PageResponse builds a PaginationResponse.
func PageResponse(p models.Page) *pagination.PaginationResponse {
	return &pagination.PaginationResponse{
		TotalCount: p.TotalCount,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
		HasNext:    p.HasNext,
	}
}

// FieldsToProto builds an audit Fields message.
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

// ----- Model ---------------------------------------------------------------

// ModelFromRow converts a sqlc Model row to the domain model.
func ModelFromRow(row eoanalyticsdb.Model) *models.Model {
	return &models.Model{
		ID:           FromPgUUID(row.ID),
		TenantID:     FromPgUUID(row.TenantID),
		Name:         row.Name,
		Version:      row.Version,
		Task:         models.InferenceTask(row.Task),
		Framework:    row.Framework,
		ArtefactURI:  row.ArtefactUri,
		MetadataJSON: string(row.MetadataJson),
		Active:       row.Active,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
		CreatedBy:    row.CreatedBy,
		UpdatedBy:    row.UpdatedBy,
	}
}

// ModelToProto converts a domain Model to its proto form.
func ModelToProto(m *models.Model) *pban.Model {
	if m == nil {
		return nil
	}
	return &pban.Model{
		Id:           m.ID.String(),
		TenantId:     m.TenantID.String(),
		Name:         m.Name,
		Version:      m.Version,
		Task:         pban.InferenceTask(m.Task),
		Framework:    m.Framework,
		ArtefactUri:  m.ArtefactURI,
		MetadataJson: m.MetadataJSON,
		Active:       m.Active,
		Fields:       FieldsToProto(m.ID, m.CreatedBy, m.CreatedAt, m.UpdatedBy, m.UpdatedAt),
	}
}

// NormalizeMetadataJSON returns "{}" for empty / whitespace input, otherwise
// trims and returns as-is.
func NormalizeMetadataJSON(s string) string {
	t := strings.TrimSpace(s)
	if t == "" {
		return "{}"
	}
	return t
}

// ----- InferenceJob --------------------------------------------------------

// InferenceJobFromRow converts a sqlc InferenceJob row to the domain model.
func InferenceJobFromRow(row eoanalyticsdb.InferenceJob) *models.InferenceJob {
	j := &models.InferenceJob{
		ID:           FromPgUUID(row.ID),
		TenantID:     FromPgUUID(row.TenantID),
		ModelID:      FromPgUUID(row.ModelID),
		ItemID:       FromPgUUID(row.ItemID),
		Status:       models.InferenceJobStatus(row.Status),
		OutputURI:    row.OutputUri,
		ErrorMessage: row.ErrorMessage,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
		CreatedBy:    row.CreatedBy,
		UpdatedBy:    row.UpdatedBy,
	}
	if row.StartedAt.Valid {
		j.StartedAt = row.StartedAt.Time
	}
	if row.FinishedAt.Valid {
		j.FinishedAt = row.FinishedAt.Time
	}
	return j
}

// InferenceJobToProto converts a domain InferenceJob to its proto form.
func InferenceJobToProto(j *models.InferenceJob) *pban.InferenceJob {
	if j == nil {
		return nil
	}
	out := &pban.InferenceJob{
		Id:           j.ID.String(),
		TenantId:     j.TenantID.String(),
		ModelId:      j.ModelID.String(),
		ItemId:       j.ItemID.String(),
		Status:       pban.InferenceJobStatus(j.Status),
		OutputUri:    j.OutputURI,
		ErrorMessage: j.ErrorMessage,
		Fields:       FieldsToProto(j.ID, j.CreatedBy, j.CreatedAt, j.UpdatedBy, j.UpdatedAt),
	}
	if !j.StartedAt.IsZero() {
		out.StartedAt = timestamppb.New(j.StartedAt)
	}
	if !j.FinishedAt.IsZero() {
		out.FinishedAt = timestamppb.New(j.FinishedAt)
	}
	return out
}
