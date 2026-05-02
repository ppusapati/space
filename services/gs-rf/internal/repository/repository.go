// Package repository wraps the gs-rf sqlc layer.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/chetana/packages/ulid"

	gsrfdb "github.com/ppusapati/space/services/gs-rf/db/generated"
	"github.com/ppusapati/space/services/gs-rf/internal/mapper"
	"github.com/ppusapati/space/services/gs-rf/internal/models"
)

// ErrNotFound is returned when no row matches.
var ErrNotFound = errors.New("repository: not found")

// Repo persists LinkBudgets and LinkMeasurements.
type Repo struct {
	q    *gsrfdb.Queries
	pool *pgxpool.Pool
}

// New constructs a Repo.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: gsrfdb.New(pool), pool: pool}
}

// ----- LinkBudget ----------------------------------------------------------

type CreateLinkBudgetParams struct {
	ID                 ulid.ID
	TenantID           ulid.ID
	PassID             ulid.ID
	StationID          ulid.ID
	AntennaID          ulid.ID
	SatelliteID        ulid.ID
	CarrierFreqHz      uint64
	TxPowerDBM         float64
	TxGainDBI          float64
	RxGainDBI          float64
	RxNoiseTempK       float64
	BandwidthHz        float64
	SlantRangeKm       float64
	FreeSpaceLossDB    float64
	AtmosphericLossDB  float64
	PolarizationLossDB float64
	PointingLossDB     float64
	PredictedEbN0DB    float64
	PredictedSNRDB     float64
	LinkMarginDB       float64
	Notes              string
	CreatedBy          string
}

func (r *Repo) CreateLinkBudget(ctx context.Context, p CreateLinkBudgetParams) (*models.LinkBudget, error) {
	row, err := r.q.CreateLinkBudget(ctx, gsrfdb.CreateLinkBudgetParams{
		ID:                 mapper.PgUUID(p.ID),
		TenantID:           mapper.PgUUID(p.TenantID),
		PassID:             mapper.PgUUID(p.PassID),
		StationID:          mapper.PgUUID(p.StationID),
		AntennaID:          mapper.PgUUID(p.AntennaID),
		SatelliteID:        mapper.PgUUID(p.SatelliteID),
		CarrierFreqHz:      int64(p.CarrierFreqHz),
		TxPowerDbm:         p.TxPowerDBM,
		TxGainDbi:          p.TxGainDBI,
		RxGainDbi:          p.RxGainDBI,
		RxNoiseTempK:       p.RxNoiseTempK,
		BandwidthHz:        p.BandwidthHz,
		SlantRangeKm:       p.SlantRangeKm,
		FreeSpaceLossDb:    p.FreeSpaceLossDB,
		AtmosphericLossDb:  p.AtmosphericLossDB,
		PolarizationLossDb: p.PolarizationLossDB,
		PointingLossDb:     p.PointingLossDB,
		PredictedEbN0Db:    p.PredictedEbN0DB,
		PredictedSnrDb:     p.PredictedSNRDB,
		LinkMarginDb:       p.LinkMarginDB,
		Notes:              p.Notes,
		CreatedBy:          p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.LinkBudgetFromRow(row), nil
}

func (r *Repo) GetLinkBudget(ctx context.Context, id ulid.ID) (*models.LinkBudget, error) {
	row, err := r.q.GetLinkBudget(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.LinkBudgetFromRow(row), nil
}

type ListLinkBudgetsParams struct {
	TenantID    ulid.ID
	PassID      *ulid.ID
	StationID   *ulid.ID
	SatelliteID *ulid.ID
	PageOffset  int32
	PageSize    int32
}

func (r *Repo) ListLinkBudgetsForTenant(ctx context.Context, p ListLinkBudgetsParams) ([]*models.LinkBudget, int32, error) {
	var passPg, stationPg, satellitePg pgtype.UUID
	if p.PassID != nil {
		passPg = mapper.PgUUID(*p.PassID)
	}
	if p.StationID != nil {
		stationPg = mapper.PgUUID(*p.StationID)
	}
	if p.SatelliteID != nil {
		satellitePg = mapper.PgUUID(*p.SatelliteID)
	}
	total, err := r.q.CountLinkBudgetsForTenant(ctx, gsrfdb.CountLinkBudgetsForTenantParams{
		TenantID:    mapper.PgUUID(p.TenantID),
		PassID:      passPg,
		StationID:   stationPg,
		SatelliteID: satellitePg,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListLinkBudgetsForTenant(ctx, gsrfdb.ListLinkBudgetsForTenantParams{
		TenantID:    mapper.PgUUID(p.TenantID),
		PassID:      passPg,
		StationID:   stationPg,
		SatelliteID: satellitePg,
		PageOffset:  p.PageOffset,
		PageSize:    p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.LinkBudget, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.LinkBudgetFromRow(row))
	}
	return out, int32(total), nil
}

// ----- LinkMeasurement -----------------------------------------------------

type RecordMeasurementParams struct {
	ID             ulid.ID
	TenantID       ulid.ID
	PassID         ulid.ID
	StationID      ulid.ID
	AntennaID      ulid.ID
	SampledAt      time.Time
	RSSIDBM        float64
	SNRDB          float64
	BER            float64
	FER            float64
	FrequencyHz    uint64
	DopplerShiftHz float64
	CreatedBy      string
}

func (r *Repo) RecordMeasurement(ctx context.Context, p RecordMeasurementParams) (*models.LinkMeasurement, error) {
	row, err := r.q.RecordMeasurement(ctx, gsrfdb.RecordMeasurementParams{
		ID:             mapper.PgUUID(p.ID),
		TenantID:       mapper.PgUUID(p.TenantID),
		PassID:         mapper.PgUUID(p.PassID),
		StationID:      mapper.PgUUID(p.StationID),
		AntennaID:      mapper.PgUUID(p.AntennaID),
		SampledAt:      mapper.PgTimestamp(p.SampledAt),
		RssiDbm:        p.RSSIDBM,
		SnrDb:          p.SNRDB,
		Ber:            p.BER,
		Fer:            p.FER,
		FrequencyHz:    int64(p.FrequencyHz),
		DopplerShiftHz: p.DopplerShiftHz,
		CreatedBy:      p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.MeasurementFromRow(row), nil
}

func (r *Repo) GetMeasurement(ctx context.Context, id ulid.ID) (*models.LinkMeasurement, error) {
	row, err := r.q.GetMeasurement(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.MeasurementFromRow(row), nil
}

type ListMeasurementsParams struct {
	TenantID   ulid.ID
	PassID     *ulid.ID
	StationID  *ulid.ID
	TimeStart  time.Time
	TimeEnd    time.Time
	PageOffset int32
	PageSize   int32
}

func (r *Repo) ListMeasurementsForTenant(ctx context.Context, p ListMeasurementsParams) ([]*models.LinkMeasurement, int32, error) {
	var passPg, stationPg pgtype.UUID
	if p.PassID != nil {
		passPg = mapper.PgUUID(*p.PassID)
	}
	if p.StationID != nil {
		stationPg = mapper.PgUUID(*p.StationID)
	}
	total, err := r.q.CountMeasurementsForTenant(ctx, gsrfdb.CountMeasurementsForTenantParams{
		TenantID:  mapper.PgUUID(p.TenantID),
		PassID:    passPg,
		StationID: stationPg,
		TimeStart: mapper.PgTimestampOrNull(p.TimeStart),
		TimeEnd:   mapper.PgTimestampOrNull(p.TimeEnd),
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListMeasurementsForTenant(ctx, gsrfdb.ListMeasurementsForTenantParams{
		TenantID:   mapper.PgUUID(p.TenantID),
		PassID:     passPg,
		StationID:  stationPg,
		TimeStart:  mapper.PgTimestampOrNull(p.TimeStart),
		TimeEnd:    mapper.PgTimestampOrNull(p.TimeEnd),
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.LinkMeasurement, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.MeasurementFromRow(row))
	}
	return out, int32(total), nil
}
