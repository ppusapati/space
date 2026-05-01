// Package mappers converts proto / domain / sqlc types for sat-telemetry.
package mappers

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/ppusapati/space/api/p9e/space/common/v1"
	satv1 "github.com/ppusapati/space/api/p9e/space/satsubsys/v1"
	sattlmdb "github.com/ppusapati/space/services/sat-telemetry/db/generated"
	"github.com/ppusapati/space/services/sat-telemetry/internal/models"
)

// PgUUID wraps uuid.UUID for sqlc.
func PgUUID(id uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: id, Valid: true} }

// FromPgUUID returns the uuid.UUID portion of a pgtype.UUID.
func FromPgUUID(p pgtype.UUID) uuid.UUID {
	if !p.Valid {
		return uuid.Nil
	}
	return p.Bytes
}

// PgUUIDPtr converts *uuid.UUID to pgtype.UUID.
func PgUUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

// PgUUIDOrNull converts a uuid.UUID to pgtype.UUID, treating uuid.Nil as NULL.
func PgUUIDOrNull(id uuid.UUID) pgtype.UUID {
	if id == uuid.Nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

// PgTimestamp converts time.Time to pgtype.Timestamptz (always Valid=true).
func PgTimestamp(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// PgTimestampPtr converts *time.Time to pgtype.Timestamptz.
func PgTimestampPtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// ChannelFromRow converts a sqlc Channel row to a domain Channel.
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
func ChannelToProto(c *models.Channel) *satv1.Channel {
	if c == nil {
		return nil
	}
	return &satv1.Channel{
		Id:           c.ID.String(),
		TenantId:     c.TenantID.String(),
		SatelliteId:  c.SatelliteID.String(),
		Subsystem:    c.Subsystem,
		Name:         c.Name,
		Units:        c.Units,
		ValueType:    satv1.ChannelValueType(c.ValueType),
		MinValue:     c.MinValue,
		MaxValue:     c.MaxValue,
		SampleRateHz: c.SampleRateHz,
		Active:       c.Active,
		Audit: &commonv1.AuditFields{
			CreatedAt: timestamppb.New(c.CreatedAt),
			UpdatedAt: timestamppb.New(c.UpdatedAt),
			CreatedBy: c.CreatedBy,
			UpdatedBy: c.UpdatedBy,
		},
	}
}

// FrameFromRow converts a sqlc TelemetryFrame row to a domain Frame.
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
func FrameToProto(f *models.Frame) *satv1.TelemetryFrame {
	if f == nil {
		return nil
	}
	return &satv1.TelemetryFrame{
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
		Audit: &commonv1.AuditFields{
			CreatedAt: timestamppb.New(f.GroundTime),
			UpdatedAt: timestamppb.New(f.GroundTime),
			CreatedBy: f.CreatedBy,
			UpdatedBy: f.CreatedBy,
		},
	}
}

// SampleFromQueryRow converts a TelemetrySample row to a domain Sample.
func SampleFromQueryRow(row sattlmdb.TelemetrySample) *models.Sample {
	frameID := uuid.Nil
	if row.FrameID.Valid {
		frameID = row.FrameID.Bytes
	}
	return &models.Sample{
		ID:          FromPgUUID(row.ID),
		TenantID:    FromPgUUID(row.TenantID),
		SatelliteID: FromPgUUID(row.SatelliteID),
		FrameID:     frameID,
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
func SampleToProto(s *models.Sample) *satv1.TelemetrySample {
	if s == nil {
		return nil
	}
	frameID := ""
	if s.FrameID != uuid.Nil {
		frameID = s.FrameID.String()
	}
	return &satv1.TelemetrySample{
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
