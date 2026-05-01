// Package services holds gs-rf business logic.
package services

import (
	"context"
	"errors"
	"time"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	"github.com/ppusapati/space/services/gs-rf/internal/models"
	"github.com/ppusapati/space/services/gs-rf/internal/repository"
)

// RF is the gs-rf service-layer facade.
type RF struct {
	repo  *repository.Repo
	IDFn  func() ulid.ID
	NowFn func() time.Time
}

// New constructs an RF service.
func New(repo *repository.Repo) *RF {
	return &RF{
		repo:  repo,
		IDFn:  ulid.NewMonotonic,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// ----- LinkBudget ----------------------------------------------------------

type CreateLinkBudgetInput struct {
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

func (s *RF) CreateLinkBudget(ctx context.Context, in CreateLinkBudgetInput) (*models.LinkBudget, error) {
	if in.TenantID.IsZero() || in.PassID.IsZero() || in.StationID.IsZero() ||
		in.AntennaID.IsZero() || in.SatelliteID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT",
			"tenant_id, pass_id, station_id, antenna_id, satellite_id required")
	}
	if in.CarrierFreqHz == 0 {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "carrier_freq_hz must be > 0")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return s.repo.CreateLinkBudget(ctx, repository.CreateLinkBudgetParams{
		ID:                 s.IDFn(),
		TenantID:           in.TenantID,
		PassID:             in.PassID,
		StationID:          in.StationID,
		AntennaID:          in.AntennaID,
		SatelliteID:        in.SatelliteID,
		CarrierFreqHz:      in.CarrierFreqHz,
		TxPowerDBM:         in.TxPowerDBM,
		TxGainDBI:          in.TxGainDBI,
		RxGainDBI:          in.RxGainDBI,
		RxNoiseTempK:       in.RxNoiseTempK,
		BandwidthHz:        in.BandwidthHz,
		SlantRangeKm:       in.SlantRangeKm,
		FreeSpaceLossDB:    in.FreeSpaceLossDB,
		AtmosphericLossDB:  in.AtmosphericLossDB,
		PolarizationLossDB: in.PolarizationLossDB,
		PointingLossDB:     in.PointingLossDB,
		PredictedEbN0DB:    in.PredictedEbN0DB,
		PredictedSNRDB:     in.PredictedSNRDB,
		LinkMarginDB:       in.LinkMarginDB,
		Notes:              in.Notes,
		CreatedBy:          createdBy,
	})
}

func (s *RF) GetLinkBudget(ctx context.Context, id ulid.ID) (*models.LinkBudget, error) {
	b, err := s.repo.GetLinkBudget(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("LINK_BUDGET_NOT_FOUND", "link_budget "+id.String())
	}
	return b, err
}

type ListLinkBudgetsInput struct {
	TenantID    ulid.ID
	PassID      *ulid.ID
	StationID   *ulid.ID
	SatelliteID *ulid.ID
	PageOffset  int32
	PageSize    int32
}

func (s *RF) ListLinkBudgetsForTenant(ctx context.Context, in ListLinkBudgetsInput) ([]*models.LinkBudget, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := s.repo.ListLinkBudgetsForTenant(ctx, repository.ListLinkBudgetsParams{
		TenantID:    in.TenantID,
		PassID:      in.PassID,
		StationID:   in.StationID,
		SatelliteID: in.SatelliteID,
		PageOffset:  in.PageOffset,
		PageSize:    in.PageSize,
	})
	if err != nil {
		return nil, models.Page{}, err
	}
	return rows, models.Page{
		TotalCount: total,
		PageOffset: in.PageOffset,
		PageSize:   in.PageSize,
		HasNext:    in.PageOffset+int32(len(rows)) < total,
	}, nil
}

// ----- LinkMeasurement -----------------------------------------------------

type RecordMeasurementInput struct {
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

func (s *RF) RecordMeasurement(ctx context.Context, in RecordMeasurementInput) (*models.LinkMeasurement, error) {
	if in.TenantID.IsZero() || in.PassID.IsZero() || in.StationID.IsZero() || in.AntennaID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT",
			"tenant_id, pass_id, station_id, antenna_id required")
	}
	if in.SampledAt.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "sampled_at required")
	}
	if in.BER < 0 || in.BER > 1 {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "ber must be in [0, 1]")
	}
	if in.FER < 0 || in.FER > 1 {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "fer must be in [0, 1]")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return s.repo.RecordMeasurement(ctx, repository.RecordMeasurementParams{
		ID:             s.IDFn(),
		TenantID:       in.TenantID,
		PassID:         in.PassID,
		StationID:      in.StationID,
		AntennaID:      in.AntennaID,
		SampledAt:      in.SampledAt,
		RSSIDBM:        in.RSSIDBM,
		SNRDB:          in.SNRDB,
		BER:            in.BER,
		FER:            in.FER,
		FrequencyHz:    in.FrequencyHz,
		DopplerShiftHz: in.DopplerShiftHz,
		CreatedBy:      createdBy,
	})
}

func (s *RF) GetMeasurement(ctx context.Context, id ulid.ID) (*models.LinkMeasurement, error) {
	m, err := s.repo.GetMeasurement(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("MEASUREMENT_NOT_FOUND", "measurement "+id.String())
	}
	return m, err
}

type ListMeasurementsInput struct {
	TenantID   ulid.ID
	PassID     *ulid.ID
	StationID  *ulid.ID
	TimeStart  time.Time
	TimeEnd    time.Time
	PageOffset int32
	PageSize   int32
}

func (s *RF) ListMeasurementsForTenant(ctx context.Context, in ListMeasurementsInput) ([]*models.LinkMeasurement, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	if !in.TimeStart.IsZero() && !in.TimeEnd.IsZero() && in.TimeEnd.Before(in.TimeStart) {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "time_end must be >= time_start")
	}
	rows, total, err := s.repo.ListMeasurementsForTenant(ctx, repository.ListMeasurementsParams{
		TenantID:   in.TenantID,
		PassID:     in.PassID,
		StationID:  in.StationID,
		TimeStart:  in.TimeStart,
		TimeEnd:    in.TimeEnd,
		PageOffset: in.PageOffset,
		PageSize:   in.PageSize,
	})
	if err != nil {
		return nil, models.Page{}, err
	}
	return rows, models.Page{
		TotalCount: total,
		PageOffset: in.PageOffset,
		PageSize:   in.PageSize,
		HasNext:    in.PageOffset+int32(len(rows)) < total,
	}, nil
}
