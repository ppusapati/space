// Package mapper converts proto / domain / sqlc types for eo-pipeline.
package mapper

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"p9e.in/chetana/packages/api/v1/fields"
	"p9e.in/chetana/packages/api/v1/pagination"
	"p9e.in/chetana/packages/ulid"

	pbpipe "github.com/ppusapati/space/services/eo-pipeline/api"
	eopipelinedb "github.com/ppusapati/space/services/eo-pipeline/db/generated"
	"github.com/ppusapati/space/services/eo-pipeline/internal/models"
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

// PageRequest extracts (offset, size) from a PaginationRequest.
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

// JobFromRow converts a sqlc Job row to the domain model.
func JobFromRow(row eopipelinedb.Job) *models.Job {
	j := &models.Job{
		ID:             FromPgUUID(row.ID),
		TenantID:       FromPgUUID(row.TenantID),
		ItemID:         FromPgUUID(row.ItemID),
		Stage:          models.JobStage(row.Stage),
		Status:         models.JobStatus(row.Status),
		ParametersJSON: string(row.ParametersJson),
		OutputURI:      row.OutputUri,
		ErrorMessage:   row.ErrorMessage,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
		CreatedBy:      row.CreatedBy,
		UpdatedBy:      row.UpdatedBy,
	}
	if row.StartedAt.Valid {
		j.StartedAt = row.StartedAt.Time
	}
	if row.FinishedAt.Valid {
		j.FinishedAt = row.FinishedAt.Time
	}
	return j
}

// JobToProto converts a domain Job to its proto form.
func JobToProto(j *models.Job) *pbpipe.Job {
	if j == nil {
		return nil
	}
	out := &pbpipe.Job{
		Id:             j.ID.String(),
		TenantId:       j.TenantID.String(),
		ItemId:         j.ItemID.String(),
		Stage:          pbpipe.JobStage(j.Stage),
		Status:         pbpipe.JobStatus(j.Status),
		ParametersJson: j.ParametersJSON,
		OutputUri:      j.OutputURI,
		ErrorMessage:   j.ErrorMessage,
		Fields:         FieldsToProto(j.ID, j.CreatedBy, j.CreatedAt, j.UpdatedBy, j.UpdatedAt),
	}
	if !j.StartedAt.IsZero() {
		out.StartedAt = timestamppb.New(j.StartedAt)
	}
	if !j.FinishedAt.IsZero() {
		out.FinishedAt = timestamppb.New(j.FinishedAt)
	}
	return out
}
