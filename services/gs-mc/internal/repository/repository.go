// Package repository wraps the gs-mc sqlc layer.
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/samavaya/packages/ulid"

	gsmcdb "github.com/ppusapati/space/services/gs-mc/db/generated"
	"github.com/ppusapati/space/services/gs-mc/internal/mapper"
	"github.com/ppusapati/space/services/gs-mc/internal/models"
)

// ErrNotFound is returned when no row matches.
var ErrNotFound = errors.New("repository: not found")

// Repo persists GroundStations and Antennas.
type Repo struct {
	q    *gsmcdb.Queries
	pool *pgxpool.Pool
}

// New constructs a Repo.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: gsmcdb.New(pool), pool: pool}
}

// ----- GroundStations ------------------------------------------------------

type CreateGroundStationParams struct {
	ID           ulid.ID
	TenantID     ulid.ID
	Slug         string
	Name         string
	CountryCode  string
	LatitudeDeg  float64
	LongitudeDeg float64
	AltitudeM    float64
	CreatedBy    string
}

func (r *Repo) CreateGroundStation(ctx context.Context, p CreateGroundStationParams) (*models.GroundStation, error) {
	row, err := r.q.CreateGroundStation(ctx, gsmcdb.CreateGroundStationParams{
		ID:           mapper.PgUUID(p.ID),
		TenantID:     mapper.PgUUID(p.TenantID),
		Slug:         p.Slug,
		Name:         p.Name,
		CountryCode:  p.CountryCode,
		LatitudeDeg:  p.LatitudeDeg,
		LongitudeDeg: p.LongitudeDeg,
		AltitudeM:    p.AltitudeM,
		CreatedBy:    p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.GroundStationFromRow(row), nil
}

func (r *Repo) GetGroundStation(ctx context.Context, id ulid.ID) (*models.GroundStation, error) {
	row, err := r.q.GetGroundStation(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.GroundStationFromRow(row), nil
}

func (r *Repo) ListGroundStationsForTenant(
	ctx context.Context, tenantID ulid.ID, offset, size int32,
) ([]*models.GroundStation, int32, error) {
	total, err := r.q.CountGroundStationsForTenant(ctx, mapper.PgUUID(tenantID))
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListGroundStationsForTenant(ctx, gsmcdb.ListGroundStationsForTenantParams{
		TenantID:   mapper.PgUUID(tenantID),
		PageOffset: offset,
		PageSize:   size,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.GroundStation, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.GroundStationFromRow(row))
	}
	return out, int32(total), nil
}

func (r *Repo) DeprecateGroundStation(ctx context.Context, id ulid.ID, updatedBy string) (*models.GroundStation, error) {
	row, err := r.q.DeprecateGroundStation(ctx, gsmcdb.DeprecateGroundStationParams{
		ID:        mapper.PgUUID(id),
		UpdatedBy: updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.GroundStationFromRow(row), nil
}

// ----- Antennas ------------------------------------------------------------

type CreateAntennaParams struct {
	ID              ulid.ID
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

func (r *Repo) CreateAntenna(ctx context.Context, p CreateAntennaParams) (*models.Antenna, error) {
	row, err := r.q.CreateAntenna(ctx, gsmcdb.CreateAntennaParams{
		ID:              mapper.PgUUID(p.ID),
		TenantID:        mapper.PgUUID(p.TenantID),
		StationID:       mapper.PgUUID(p.StationID),
		Slug:            p.Slug,
		Name:            p.Name,
		Band:            int32(p.Band),
		MinFreqHz:       int64(p.MinFreqHz),
		MaxFreqHz:       int64(p.MaxFreqHz),
		Polarization:    int32(p.Polarization),
		GainDbi:         p.GainDBI,
		SlewRateDegPerS: p.SlewRateDegPerS,
		CreatedBy:       p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.AntennaFromRow(row), nil
}

func (r *Repo) GetAntenna(ctx context.Context, id ulid.ID) (*models.Antenna, error) {
	row, err := r.q.GetAntenna(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.AntennaFromRow(row), nil
}

type ListAntennasParams struct {
	TenantID   ulid.ID
	StationID  *ulid.ID
	Band       *models.FrequencyBand
	PageOffset int32
	PageSize   int32
}

func (r *Repo) ListAntennasForTenant(ctx context.Context, p ListAntennasParams) ([]*models.Antenna, int32, error) {
	var bandPtr *int32
	if p.Band != nil {
		v := int32(*p.Band)
		bandPtr = &v
	}
	var stationPg pgtype.UUID
	if p.StationID != nil {
		stationPg = mapper.PgUUID(*p.StationID)
	}
	total, err := r.q.CountAntennasForTenant(ctx, gsmcdb.CountAntennasForTenantParams{
		TenantID:  mapper.PgUUID(p.TenantID),
		StationID: stationPg,
		Band:      bandPtr,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListAntennasForTenant(ctx, gsmcdb.ListAntennasForTenantParams{
		TenantID:   mapper.PgUUID(p.TenantID),
		StationID:  stationPg,
		Band:       bandPtr,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.Antenna, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.AntennaFromRow(row))
	}
	return out, int32(total), nil
}

func (r *Repo) DeprecateAntenna(ctx context.Context, id ulid.ID, updatedBy string) (*models.Antenna, error) {
	row, err := r.q.DeprecateAntenna(ctx, gsmcdb.DeprecateAntennaParams{
		ID:        mapper.PgUUID(id),
		UpdatedBy: updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.AntennaFromRow(row), nil
}
