// Package mapper converts proto / domain / sqlc types for gs-scheduler.
package mapper

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"p9e.in/chetana/packages/api/v1/fields"
	"p9e.in/chetana/packages/api/v1/pagination"
	"p9e.in/chetana/packages/ulid"

	pbsch "github.com/ppusapati/space/services/gs-scheduler/api"
	gsschdb "github.com/ppusapati/space/services/gs-scheduler/db/generated"
	"github.com/ppusapati/space/services/gs-scheduler/internal/models"
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

// ----- ContactPass --------------------------------------------------------

func ContactPassFromRow(row gsschdb.ContactPass) *models.ContactPass {
	return &models.ContactPass{
		ID:              FromPgUUID(row.ID),
		TenantID:        FromPgUUID(row.TenantID),
		StationID:       FromPgUUID(row.StationID),
		SatelliteID:     FromPgUUID(row.SatelliteID),
		AOSTime:         row.AosTime.Time,
		TCATime:         row.TcaTime.Time,
		LOSTime:         row.LosTime.Time,
		MaxElevationDeg: row.MaxElevationDeg,
		AOSAzimuthDeg:   row.AosAzimuthDeg,
		LOSAzimuthDeg:   row.LosAzimuthDeg,
		Source:          row.Source,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		CreatedBy:       row.CreatedBy,
		UpdatedBy:       row.UpdatedBy,
	}
}

func ContactPassToProto(p *models.ContactPass) *pbsch.ContactPass {
	if p == nil {
		return nil
	}
	return &pbsch.ContactPass{
		Id:              p.ID.String(),
		TenantId:        p.TenantID.String(),
		StationId:       p.StationID.String(),
		SatelliteId:     p.SatelliteID.String(),
		AosTime:         timestamppb.New(p.AOSTime),
		TcaTime:         timestamppb.New(p.TCATime),
		LosTime:         timestamppb.New(p.LOSTime),
		MaxElevationDeg: p.MaxElevationDeg,
		AosAzimuthDeg:   p.AOSAzimuthDeg,
		LosAzimuthDeg:   p.LOSAzimuthDeg,
		Source:          p.Source,
		Fields:          FieldsToProto(p.ID, p.CreatedBy, p.CreatedAt, p.UpdatedBy, p.UpdatedAt),
	}
}

// ----- Booking ------------------------------------------------------------

func BookingFromRow(row gsschdb.Booking) *models.Booking {
	b := &models.Booking{
		ID:           FromPgUUID(row.ID),
		TenantID:     FromPgUUID(row.TenantID),
		PassID:       FromPgUUID(row.PassID),
		Priority:     row.Priority,
		Status:       models.BookingStatus(row.Status),
		Purpose:      row.Purpose,
		Notes:        row.Notes,
		ErrorMessage: row.ErrorMessage,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
		CreatedBy:    row.CreatedBy,
		UpdatedBy:    row.UpdatedBy,
	}
	if row.ScheduledAt.Valid {
		b.ScheduledAt = row.ScheduledAt.Time
	}
	if row.CompletedAt.Valid {
		b.CompletedAt = row.CompletedAt.Time
	}
	return b
}

func BookingToProto(b *models.Booking) *pbsch.Booking {
	if b == nil {
		return nil
	}
	out := &pbsch.Booking{
		Id:           b.ID.String(),
		TenantId:     b.TenantID.String(),
		PassId:       b.PassID.String(),
		Priority:     b.Priority,
		Status:       pbsch.BookingStatus(b.Status),
		Purpose:      b.Purpose,
		Notes:        b.Notes,
		ErrorMessage: b.ErrorMessage,
		Fields:       FieldsToProto(b.ID, b.CreatedBy, b.CreatedAt, b.UpdatedBy, b.UpdatedAt),
	}
	if !b.ScheduledAt.IsZero() {
		out.ScheduledAt = timestamppb.New(b.ScheduledAt)
	}
	if !b.CompletedAt.IsZero() {
		out.CompletedAt = timestamppb.New(b.CompletedAt)
	}
	return out
}
