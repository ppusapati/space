// Package mapper converts proto / domain / sqlc types for sat-command.
package mapper

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"p9e.in/chetana/packages/api/v1/fields"
	"p9e.in/chetana/packages/api/v1/pagination"
	"p9e.in/chetana/packages/ulid"

	pbcmd "github.com/ppusapati/space/services/sat-command/api"
	satcmddb "github.com/ppusapati/space/services/sat-command/db/generated"
	"github.com/ppusapati/space/services/sat-command/internal/models"
)

// PgUUID converts a ulid.ID into a pgtype.UUID payload (16 bytes).
func PgUUID(id ulid.ID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

// PgUUIDOrNull treats ulid.Zero as NULL.
func PgUUIDOrNull(id ulid.ID) pgtype.UUID {
	if id.IsZero() {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

// FromPgUUID extracts ulid.ID from pgtype.UUID. Returns ulid.Zero on NULL.
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

// PgTimestampOrNull returns NULL when t is the zero value.
func PgTimestampOrNull(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{Valid: false}
	}
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

// ----- CommandDef ----------------------------------------------------------

// CommandDefFromRow converts a sqlc row to the domain model.
func CommandDefFromRow(row satcmddb.CommandDef) *models.CommandDef {
	return &models.CommandDef{
		ID:               FromPgUUID(row.ID),
		TenantID:         FromPgUUID(row.TenantID),
		SatelliteID:      FromPgUUID(row.SatelliteID),
		Subsystem:        row.Subsystem,
		Name:             row.Name,
		Opcode:           uint32(row.Opcode),
		ParametersSchema: row.ParametersSchema,
		Description:      row.Description,
		Active:           row.Active,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
		CreatedBy:        row.CreatedBy,
		UpdatedBy:        row.UpdatedBy,
	}
}

// CommandDefToProto converts a domain CommandDef to its proto form.
func CommandDefToProto(c *models.CommandDef) *pbcmd.CommandDef {
	if c == nil {
		return nil
	}
	out := &pbcmd.CommandDef{
		Id:               c.ID.String(),
		TenantId:         c.TenantID.String(),
		Subsystem:        c.Subsystem,
		Name:             c.Name,
		Opcode:           c.Opcode,
		ParametersSchema: c.ParametersSchema,
		Description:      c.Description,
		Active:           c.Active,
		Fields:           FieldsToProto(c.ID, c.CreatedBy, c.CreatedAt, c.UpdatedBy, c.UpdatedAt),
	}
	if !c.SatelliteID.IsZero() {
		out.SatelliteId = c.SatelliteID.String()
	}
	return out
}

// ----- UplinkRequest -------------------------------------------------------

// UplinkFromRow converts a sqlc row to the domain model.
func UplinkFromRow(row satcmddb.UplinkRequest) *models.UplinkRequest {
	u := &models.UplinkRequest{
		ID:               FromPgUUID(row.ID),
		TenantID:         FromPgUUID(row.TenantID),
		SatelliteID:      FromPgUUID(row.SatelliteID),
		CommandDefID:     FromPgUUID(row.CommandDefID),
		ParametersJSON:   row.ParametersJson,
		ScheduledRelease: row.ScheduledRelease.Time,
		Status:           models.UplinkStatus(row.Status),
		SequenceNumber:   uint64(row.SequenceNumber),
		GatewayID:        row.GatewayID,
		SubmittedAt:      row.SubmittedAt.Time,
		ErrorMessage:     row.ErrorMessage,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
		CreatedBy:        row.CreatedBy,
		UpdatedBy:        row.UpdatedBy,
	}
	if row.ReleasedAt.Valid {
		u.ReleasedAt = row.ReleasedAt.Time
	}
	if row.AckedAt.Valid {
		u.AckedAt = row.AckedAt.Time
	}
	if row.CompletedAt.Valid {
		u.CompletedAt = row.CompletedAt.Time
	}
	return u
}

// UplinkToProto converts a domain UplinkRequest to its proto form.
func UplinkToProto(u *models.UplinkRequest) *pbcmd.UplinkRequest {
	if u == nil {
		return nil
	}
	out := &pbcmd.UplinkRequest{
		Id:               u.ID.String(),
		TenantId:         u.TenantID.String(),
		SatelliteId:      u.SatelliteID.String(),
		CommandDefId:     u.CommandDefID.String(),
		ParametersJson:   u.ParametersJSON,
		ScheduledRelease: timestamppb.New(u.ScheduledRelease),
		Status:           pbcmd.UplinkStatus(u.Status),
		SequenceNumber:   u.SequenceNumber,
		GatewayId:        u.GatewayID,
		SubmittedAt:      timestamppb.New(u.SubmittedAt),
		ErrorMessage:     u.ErrorMessage,
		Fields:           FieldsToProto(u.ID, u.CreatedBy, u.CreatedAt, u.UpdatedBy, u.UpdatedAt),
	}
	if !u.ReleasedAt.IsZero() {
		out.ReleasedAt = timestamppb.New(u.ReleasedAt)
	}
	if !u.AckedAt.IsZero() {
		out.AckedAt = timestamppb.New(u.AckedAt)
	}
	if !u.CompletedAt.IsZero() {
		out.CompletedAt = timestamppb.New(u.CompletedAt)
	}
	return out
}
