// Package mapper converts proto / domain / sqlc types for gs-mc.
package mapper

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"p9e.in/samavaya/packages/api/v1/fields"
	"p9e.in/samavaya/packages/api/v1/pagination"
	"p9e.in/samavaya/packages/ulid"

	pbmc "github.com/ppusapati/space/services/gs-mc/api"
	gsmcdb "github.com/ppusapati/space/services/gs-mc/db/generated"
	"github.com/ppusapati/space/services/gs-mc/internal/models"
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

// ----- GroundStation -------------------------------------------------------

func GroundStationFromRow(row gsmcdb.GroundStation) *models.GroundStation {
	return &models.GroundStation{
		ID:           FromPgUUID(row.ID),
		TenantID:     FromPgUUID(row.TenantID),
		Slug:         row.Slug,
		Name:         row.Name,
		CountryCode:  row.CountryCode,
		LatitudeDeg:  row.LatitudeDeg,
		LongitudeDeg: row.LongitudeDeg,
		AltitudeM:    row.AltitudeM,
		Active:       row.Active,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
		CreatedBy:    row.CreatedBy,
		UpdatedBy:    row.UpdatedBy,
	}
}

func GroundStationToProto(s *models.GroundStation) *pbmc.GroundStation {
	if s == nil {
		return nil
	}
	return &pbmc.GroundStation{
		Id:           s.ID.String(),
		TenantId:     s.TenantID.String(),
		Slug:         s.Slug,
		Name:         s.Name,
		CountryCode:  s.CountryCode,
		LatitudeDeg:  s.LatitudeDeg,
		LongitudeDeg: s.LongitudeDeg,
		AltitudeM:    s.AltitudeM,
		Active:       s.Active,
		Fields:       FieldsToProto(s.ID, s.CreatedBy, s.CreatedAt, s.UpdatedBy, s.UpdatedAt),
	}
}

// ----- Antenna -------------------------------------------------------------

func AntennaFromRow(row gsmcdb.Antenna) *models.Antenna {
	return &models.Antenna{
		ID:              FromPgUUID(row.ID),
		TenantID:        FromPgUUID(row.TenantID),
		StationID:       FromPgUUID(row.StationID),
		Slug:            row.Slug,
		Name:            row.Name,
		Band:            models.FrequencyBand(row.Band),
		MinFreqHz:       uint64(row.MinFreqHz),
		MaxFreqHz:       uint64(row.MaxFreqHz),
		Polarization:    models.Polarization(row.Polarization),
		GainDBI:         row.GainDbi,
		SlewRateDegPerS: row.SlewRateDegPerS,
		Active:          row.Active,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		CreatedBy:       row.CreatedBy,
		UpdatedBy:       row.UpdatedBy,
	}
}

func AntennaToProto(a *models.Antenna) *pbmc.Antenna {
	if a == nil {
		return nil
	}
	return &pbmc.Antenna{
		Id:               a.ID.String(),
		TenantId:         a.TenantID.String(),
		StationId:        a.StationID.String(),
		Slug:             a.Slug,
		Name:             a.Name,
		Band:             pbmc.FrequencyBand(a.Band),
		MinFreqHz:        a.MinFreqHz,
		MaxFreqHz:        a.MaxFreqHz,
		Polarization:     pbmc.Polarization(a.Polarization),
		GainDbi:          a.GainDBI,
		SlewRateDegPerS:  a.SlewRateDegPerS,
		Active:           a.Active,
		Fields:           FieldsToProto(a.ID, a.CreatedBy, a.CreatedAt, a.UpdatedBy, a.UpdatedAt),
	}
}
