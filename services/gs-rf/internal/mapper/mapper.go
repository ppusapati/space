// Package mapper converts proto / domain / sqlc types for gs-rf.
package mapper

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"p9e.in/samavaya/packages/api/v1/fields"
	"p9e.in/samavaya/packages/api/v1/pagination"
	"p9e.in/samavaya/packages/ulid"

	pbrf "github.com/ppusapati/space/services/gs-rf/api"
	gsrfdb "github.com/ppusapati/space/services/gs-rf/db/generated"
	"github.com/ppusapati/space/services/gs-rf/internal/models"
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

// ----- LinkBudget ----------------------------------------------------------

func LinkBudgetFromRow(row gsrfdb.LinkBudget) *models.LinkBudget {
	return &models.LinkBudget{
		ID:                 FromPgUUID(row.ID),
		TenantID:           FromPgUUID(row.TenantID),
		PassID:             FromPgUUID(row.PassID),
		StationID:          FromPgUUID(row.StationID),
		AntennaID:          FromPgUUID(row.AntennaID),
		SatelliteID:        FromPgUUID(row.SatelliteID),
		CarrierFreqHz:      uint64(row.CarrierFreqHz),
		TxPowerDBM:         row.TxPowerDbm,
		TxGainDBI:          row.TxGainDbi,
		RxGainDBI:          row.RxGainDbi,
		RxNoiseTempK:       row.RxNoiseTempK,
		BandwidthHz:        row.BandwidthHz,
		SlantRangeKm:       row.SlantRangeKm,
		FreeSpaceLossDB:    row.FreeSpaceLossDb,
		AtmosphericLossDB:  row.AtmosphericLossDb,
		PolarizationLossDB: row.PolarizationLossDb,
		PointingLossDB:     row.PointingLossDb,
		PredictedEbN0DB:    row.PredictedEbN0Db,
		PredictedSNRDB:     row.PredictedSnrDb,
		LinkMarginDB:       row.LinkMarginDb,
		Notes:              row.Notes,
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
		CreatedBy:          row.CreatedBy,
		UpdatedBy:          row.UpdatedBy,
	}
}

func LinkBudgetToProto(b *models.LinkBudget) *pbrf.LinkBudget {
	if b == nil {
		return nil
	}
	return &pbrf.LinkBudget{
		Id:                 b.ID.String(),
		TenantId:           b.TenantID.String(),
		PassId:             b.PassID.String(),
		StationId:          b.StationID.String(),
		AntennaId:          b.AntennaID.String(),
		SatelliteId:        b.SatelliteID.String(),
		CarrierFreqHz:      b.CarrierFreqHz,
		TxPowerDbm:         b.TxPowerDBM,
		TxGainDbi:          b.TxGainDBI,
		RxGainDbi:          b.RxGainDBI,
		RxNoiseTempK:       b.RxNoiseTempK,
		BandwidthHz:        b.BandwidthHz,
		SlantRangeKm:       b.SlantRangeKm,
		FreeSpaceLossDb:    b.FreeSpaceLossDB,
		AtmosphericLossDb:  b.AtmosphericLossDB,
		PolarizationLossDb: b.PolarizationLossDB,
		PointingLossDb:     b.PointingLossDB,
		PredictedEbN0Db:    b.PredictedEbN0DB,
		PredictedSnrDb:     b.PredictedSNRDB,
		LinkMarginDb:       b.LinkMarginDB,
		Notes:              b.Notes,
		Fields:             FieldsToProto(b.ID, b.CreatedBy, b.CreatedAt, b.UpdatedBy, b.UpdatedAt),
	}
}

// ----- LinkMeasurement -----------------------------------------------------

func MeasurementFromRow(row gsrfdb.LinkMeasurement) *models.LinkMeasurement {
	return &models.LinkMeasurement{
		ID:             FromPgUUID(row.ID),
		TenantID:       FromPgUUID(row.TenantID),
		PassID:         FromPgUUID(row.PassID),
		StationID:      FromPgUUID(row.StationID),
		AntennaID:      FromPgUUID(row.AntennaID),
		SampledAt:      row.SampledAt.Time,
		RSSIDBM:        row.RssiDbm,
		SNRDB:          row.SnrDb,
		BER:            row.Ber,
		FER:            row.Fer,
		FrequencyHz:    uint64(row.FrequencyHz),
		DopplerShiftHz: row.DopplerShiftHz,
		CreatedAt:      row.CreatedAt.Time,
		CreatedBy:      row.CreatedBy,
	}
}

func MeasurementToProto(m *models.LinkMeasurement) *pbrf.LinkMeasurement {
	if m == nil {
		return nil
	}
	return &pbrf.LinkMeasurement{
		Id:             m.ID.String(),
		TenantId:       m.TenantID.String(),
		PassId:         m.PassID.String(),
		StationId:      m.StationID.String(),
		AntennaId:      m.AntennaID.String(),
		SampledAt:      timestamppb.New(m.SampledAt),
		RssiDbm:        m.RSSIDBM,
		SnrDb:          m.SNRDB,
		Ber:            m.BER,
		Fer:            m.FER,
		FrequencyHz:    m.FrequencyHz,
		DopplerShiftHz: m.DopplerShiftHz,
		Fields:         FieldsToProto(m.ID, m.CreatedBy, m.CreatedAt, m.CreatedBy, m.CreatedAt),
	}
}
