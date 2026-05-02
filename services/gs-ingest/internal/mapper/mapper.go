// Package mapper converts proto / domain / sqlc types for gs-ingest.
package mapper

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"p9e.in/chetana/packages/api/v1/fields"
	"p9e.in/chetana/packages/api/v1/pagination"
	"p9e.in/chetana/packages/ulid"

	pbing "github.com/ppusapati/space/services/gs-ingest/api"
	gsingdb "github.com/ppusapati/space/services/gs-ingest/db/generated"
	"github.com/ppusapati/space/services/gs-ingest/internal/models"
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

func PgTimestamp(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func PgTimestampOrNull(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
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

// ----- IngestSession ------------------------------------------------------

func IngestSessionFromRow(row gsingdb.IngestSession) *models.IngestSession {
	s := &models.IngestSession{
		ID:             FromPgUUID(row.ID),
		TenantID:       FromPgUUID(row.TenantID),
		BookingID:      FromPgUUID(row.BookingID),
		PassID:         FromPgUUID(row.PassID),
		StationID:      FromPgUUID(row.StationID),
		SatelliteID:    FromPgUUID(row.SatelliteID),
		Status:         models.IngestStatus(row.Status),
		FramesReceived: uint64(row.FramesReceived),
		BytesReceived:  uint64(row.BytesReceived),
		ErrorMessage:   row.ErrorMessage,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
		CreatedBy:      row.CreatedBy,
		UpdatedBy:      row.UpdatedBy,
	}
	if row.StartedAt.Valid {
		s.StartedAt = row.StartedAt.Time
	}
	if row.CompletedAt.Valid {
		s.CompletedAt = row.CompletedAt.Time
	}
	return s
}

func IngestSessionToProto(s *models.IngestSession) *pbing.IngestSession {
	if s == nil {
		return nil
	}
	out := &pbing.IngestSession{
		Id:             s.ID.String(),
		TenantId:       s.TenantID.String(),
		BookingId:      s.BookingID.String(),
		PassId:         s.PassID.String(),
		StationId:      s.StationID.String(),
		SatelliteId:    s.SatelliteID.String(),
		Status:         pbing.IngestStatus(s.Status),
		FramesReceived: s.FramesReceived,
		BytesReceived:  s.BytesReceived,
		ErrorMessage:   s.ErrorMessage,
		Fields:         FieldsToProto(s.ID, s.CreatedBy, s.CreatedAt, s.UpdatedBy, s.UpdatedAt),
	}
	if !s.StartedAt.IsZero() {
		out.StartedAt = timestamppb.New(s.StartedAt)
	}
	if !s.CompletedAt.IsZero() {
		out.CompletedAt = timestamppb.New(s.CompletedAt)
	}
	return out
}

// ----- DownlinkFrame ------------------------------------------------------

func DownlinkFrameFromRow(row gsingdb.DownlinkFrame) *models.DownlinkFrame {
	return &models.DownlinkFrame{
		ID:               FromPgUUID(row.ID),
		TenantID:         FromPgUUID(row.TenantID),
		SessionID:        FromPgUUID(row.SessionID),
		APID:             uint32(row.Apid),
		VirtualChannel:   uint32(row.VirtualChannel),
		SequenceCount:    uint64(row.SequenceCount),
		GroundTime:       row.GroundTime.Time,
		PayloadSizeBytes: uint64(row.PayloadSizeBytes),
		PayloadSHA256:    row.PayloadSha256,
		PayloadURI:       row.PayloadUri,
		FrameType:        row.FrameType,
		CreatedAt:        row.CreatedAt.Time,
		CreatedBy:        row.CreatedBy,
	}
}

func DownlinkFrameToProto(f *models.DownlinkFrame) *pbing.DownlinkFrame {
	if f == nil {
		return nil
	}
	return &pbing.DownlinkFrame{
		Id:               f.ID.String(),
		TenantId:         f.TenantID.String(),
		SessionId:        f.SessionID.String(),
		Apid:             f.APID,
		VirtualChannel:   f.VirtualChannel,
		SequenceCount:    f.SequenceCount,
		GroundTime:       timestamppb.New(f.GroundTime),
		PayloadSizeBytes: f.PayloadSizeBytes,
		PayloadSha256:    f.PayloadSHA256,
		PayloadUri:       f.PayloadURI,
		FrameType:        f.FrameType,
		Fields:           FieldsToProto(f.ID, f.CreatedBy, f.CreatedAt, f.CreatedBy, f.CreatedAt),
	}
}
