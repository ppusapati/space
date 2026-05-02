// Package services holds gs-mc business logic.
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	"github.com/ppusapati/space/services/gs-mc/internal/models"
	"github.com/ppusapati/space/services/gs-mc/internal/repository"
)

// MissionControl is the gs-mc service-layer facade.
type MissionControl struct {
	repo  *repository.Repo
	IDFn  func() ulid.ID
	NowFn func() time.Time
}

// New constructs a MissionControl service.
func New(repo *repository.Repo) *MissionControl {
	return &MissionControl{
		repo:  repo,
		IDFn:  ulid.NewMonotonic,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// ----- GroundStation -------------------------------------------------------

type CreateGroundStationInput struct {
	TenantID     ulid.ID
	Slug         string
	Name         string
	CountryCode  string
	LatitudeDeg  float64
	LongitudeDeg float64
	AltitudeM    float64
	CreatedBy    string
}

func (m *MissionControl) CreateGroundStation(ctx context.Context, in CreateGroundStationInput) (*models.GroundStation, error) {
	if in.TenantID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	in.Slug = strings.TrimSpace(in.Slug)
	in.Name = strings.TrimSpace(in.Name)
	in.CountryCode = strings.ToUpper(strings.TrimSpace(in.CountryCode))
	if in.Slug == "" || in.Name == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "slug and name required")
	}
	if len(in.CountryCode) != 2 {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "country_code must be ISO-3166 alpha-2")
	}
	if in.LatitudeDeg < -90 || in.LatitudeDeg > 90 {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "latitude must be in [-90, 90]")
	}
	if in.LongitudeDeg < -180 || in.LongitudeDeg > 180 {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "longitude must be in [-180, 180]")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return m.repo.CreateGroundStation(ctx, repository.CreateGroundStationParams{
		ID:           m.IDFn(),
		TenantID:     in.TenantID,
		Slug:         in.Slug,
		Name:         in.Name,
		CountryCode:  in.CountryCode,
		LatitudeDeg:  in.LatitudeDeg,
		LongitudeDeg: in.LongitudeDeg,
		AltitudeM:    in.AltitudeM,
		CreatedBy:    createdBy,
	})
}

func (m *MissionControl) GetGroundStation(ctx context.Context, id ulid.ID) (*models.GroundStation, error) {
	s, err := m.repo.GetGroundStation(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("STATION_NOT_FOUND", "ground_station "+id.String())
	}
	return s, err
}

func (m *MissionControl) ListGroundStationsForTenant(
	ctx context.Context, tenantID ulid.ID, offset, size int32,
) ([]*models.GroundStation, models.Page, error) {
	if tenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := m.repo.ListGroundStationsForTenant(ctx, tenantID, offset, size)
	if err != nil {
		return nil, models.Page{}, err
	}
	return rows, models.Page{
		TotalCount: total,
		PageOffset: offset,
		PageSize:   size,
		HasNext:    offset+int32(len(rows)) < total,
	}, nil
}

func (m *MissionControl) DeprecateGroundStation(ctx context.Context, id ulid.ID, updatedBy string) (*models.GroundStation, error) {
	if updatedBy == "" {
		updatedBy = "system"
	}
	s, err := m.repo.DeprecateGroundStation(ctx, id, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("STATION_NOT_FOUND", "ground_station "+id.String())
	}
	return s, err
}

// ----- Antenna -------------------------------------------------------------

type CreateAntennaInput struct {
	TenantID        ulid.ID
	StationID       ulid.ID
	Slug            string
	Name            string
	Band            models.FrequencyBand
	MinFreqHz       uint64
	MaxFreqHz       uint64
	Polarization    models.Polarization
	GainDBI         float64
	SlewRateDegPerS float64
	CreatedBy       string
}

func (m *MissionControl) CreateAntenna(ctx context.Context, in CreateAntennaInput) (*models.Antenna, error) {
	if in.TenantID.IsZero() || in.StationID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id and station_id required")
	}
	in.Slug = strings.TrimSpace(in.Slug)
	in.Name = strings.TrimSpace(in.Name)
	if in.Slug == "" || in.Name == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "slug and name required")
	}
	if in.Band == models.BandUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "band required")
	}
	if in.Polarization == models.PolUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "polarization required")
	}
	if in.MaxFreqHz < in.MinFreqHz {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "max_freq_hz must be >= min_freq_hz")
	}
	// Verify the station exists and belongs to the same tenant.
	st, err := m.repo.GetGroundStation(ctx, in.StationID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, pkgerrors.BadRequest("STATION_NOT_FOUND",
				"ground_station "+in.StationID.String())
		}
		return nil, err
	}
	if st.TenantID != in.TenantID {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "station tenant mismatch")
	}
	if !st.Active {
		return nil, pkgerrors.New(412, "STATION_DEPRECATED", "ground_station is deprecated")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return m.repo.CreateAntenna(ctx, repository.CreateAntennaParams{
		ID:              m.IDFn(),
		TenantID:        in.TenantID,
		StationID:       in.StationID,
		Slug:            in.Slug,
		Name:            in.Name,
		Band:            in.Band,
		MinFreqHz:       in.MinFreqHz,
		MaxFreqHz:       in.MaxFreqHz,
		Polarization:    in.Polarization,
		GainDBI:         in.GainDBI,
		SlewRateDegPerS: in.SlewRateDegPerS,
		CreatedBy:       createdBy,
	})
}

func (m *MissionControl) GetAntenna(ctx context.Context, id ulid.ID) (*models.Antenna, error) {
	a, err := m.repo.GetAntenna(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("ANTENNA_NOT_FOUND", "antenna "+id.String())
	}
	return a, err
}

type ListAntennasInput struct {
	TenantID   ulid.ID
	StationID  *ulid.ID
	Band       *models.FrequencyBand
	PageOffset int32
	PageSize   int32
}

func (m *MissionControl) ListAntennasForTenant(ctx context.Context, in ListAntennasInput) ([]*models.Antenna, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := m.repo.ListAntennasForTenant(ctx, repository.ListAntennasParams{
		TenantID:   in.TenantID,
		StationID:  in.StationID,
		Band:       in.Band,
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

func (m *MissionControl) DeprecateAntenna(ctx context.Context, id ulid.ID, updatedBy string) (*models.Antenna, error) {
	if updatedBy == "" {
		updatedBy = "system"
	}
	a, err := m.repo.DeprecateAntenna(ctx, id, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("ANTENNA_NOT_FOUND", "antenna "+id.String())
	}
	return a, err
}
