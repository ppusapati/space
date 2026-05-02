// Package mapper converts proto / domain / sqlc types for sat-telemetry.
package mapper

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"p9e.in/chetana/packages/api/v1/fields"
	"p9e.in/chetana/packages/api/v1/pagination"
	"p9e.in/chetana/packages/ulid"

	pbtlm "github.com/ppusapati/space/services/sat-telemetry/api"
	sattlmdb "github.com/ppusapati/space/services/sat-telemetry/db/generated"
	"github.com/ppusapati/space/services/sat-telemetry/internal/models"
)

// PgUUID converts a ulid.ID to a pgtype.UUID payload (16 bytes).
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

// ----- Channel -------------------------------------------------------------

// ChannelFromRow converts a sqlc row to the domain model.
func ChannelFromRow(row sattlmdb.Channel) *models.Channel {
	return &models.Channel{
		ID:           FromPgUUID(row.ID),
		TenantID:     FromPgUUID(row.TenantID),
		SatelliteID:  FromPgUUID(row.SatelliteID),
		Subsystem:    row.Subsystem,
		Name:         row.Name,
		Units:        row.Units,
		ValueType:    models.ChannelValueType(row.ValueType),
		MinValue:     row.MinValue,
		MaxValue:     row.MaxValue,
		SampleRateHz: row.SampleRateHz,
		Active:       row.Active,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
		CreatedBy:    row.CreatedBy,
		UpdatedBy:    row.UpdatedBy,
	}
}

// ChannelToProto converts a domain Channel to its proto form.
func ChannelToProto(c *models.Channel) *pbtlm.Channel {
	if c == nil {
		return nil
	}
	return &pbtlm.Channel{
		Id:           c.ID.String(),
		TenantId:     c.TenantID.String(),
		SatelliteId:  c.SatelliteID.String(),
		Subsystem:    c.Subsystem,
		Name:         c.Name,
		Units:        c.Units,
		ValueType:    pbtlm.ChannelValueType(c.ValueType),
		MinValue:     c.MinValue,
		MaxValue:     c.MaxValue,
		SampleRateHz: c.SampleRateHz,
		Active:       c.Active,
		Fields:       FieldsToProto(c.ID, c.CreatedBy, c.CreatedAt, c.UpdatedBy, c.UpdatedAt),
	}
}

// ----- Frame ---------------------------------------------------------------

// FrameFromRow converts a sqlc TelemetryFrame row to the domain model.
func FrameFromRow(row sattlmdb.TelemetryFrame) *models.Frame {
	return &models.Frame{
		ID:               FromPgUUID(row.ID),
		TenantID:         FromPgUUID(row.TenantID),
		SatelliteID:      FromPgUUID(row.SatelliteID),
		APID:             uint32(row.Apid),
		VirtualChannel:   uint32(row.VirtualChannel),
		SequenceCount:    uint64(row.SequenceCount),
		SatTime:          row.SatTime.Time,
		GroundTime:       row.GroundTime.Time,
		PayloadSizeBytes: uint64(row.PayloadSizeBytes),
		PayloadSHA256:    row.PayloadSha256,
		FrameType:        row.FrameType,
		CreatedBy:        row.CreatedBy,
	}
}

// FrameToProto converts a domain Frame to its proto form.
func FrameToProto(f *models.Frame) *pbtlm.TelemetryFrame {
	if f == nil {
		return nil
	}
	return &pbtlm.TelemetryFrame{
		Id:               f.ID.String(),
		TenantId:         f.TenantID.String(),
		SatelliteId:      f.SatelliteID.String(),
		Apid:             f.APID,
		VirtualChannel:   f.VirtualChannel,
		SequenceCount:    f.SequenceCount,
		SatTime:          timestamppb.New(f.SatTime),
		GroundTime:       timestamppb.New(f.GroundTime),
		PayloadSizeBytes: f.PayloadSizeBytes,
		PayloadSha256:    f.PayloadSHA256,
		FrameType:        f.FrameType,
		Fields:           FieldsToProto(f.ID, f.CreatedBy, f.GroundTime, f.CreatedBy, f.GroundTime),
	}
}

// ----- Sample --------------------------------------------------------------

// SampleFromRow converts a TelemetrySample row to the domain model.
func SampleFromRow(row sattlmdb.TelemetrySample) *models.Sample {
	return &models.Sample{
		ID:          FromPgUUID(row.ID),
		TenantID:    FromPgUUID(row.TenantID),
		SatelliteID: FromPgUUID(row.SatelliteID),
		FrameID:     FromPgUUID(row.FrameID),
		ChannelID:   FromPgUUID(row.ChannelID),
		SampleTime:  row.SampleTime.Time,
		ValueDouble: row.ValueDouble,
		ValueInt:    row.ValueInt,
		ValueBool:   row.ValueBool,
		ValueText:   row.ValueText,
		IngestedAt:  row.IngestedAt.Time,
	}
}

// SampleToProto converts a domain Sample to its proto form.
func SampleToProto(s *models.Sample) *pbtlm.TelemetrySample {
	if s == nil {
		return nil
	}
	frameID := ""
	if !s.FrameID.IsZero() {
		frameID = s.FrameID.String()
	}
	return &pbtlm.TelemetrySample{
		Id:          s.ID.String(),
		TenantId:    s.TenantID.String(),
		SatelliteId: s.SatelliteID.String(),
		FrameId:     frameID,
		ChannelId:   s.ChannelID.String(),
		SampleTime:  timestamppb.New(s.SampleTime),
		ValueDouble: s.ValueDouble,
		ValueInt:    s.ValueInt,
		ValueBool:   s.ValueBool,
		ValueText:   s.ValueText,
	}
}
