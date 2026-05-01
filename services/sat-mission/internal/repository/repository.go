// Package repository wraps the sat-mission sqlc layer.
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/samavaya/packages/ulid"

	satmissiondb "github.com/ppusapati/space/services/sat-mission/db/generated"
	"github.com/ppusapati/space/services/sat-mission/internal/mapper"
	"github.com/ppusapati/space/services/sat-mission/internal/models"
)

// ErrNotFound is returned when no row matches.
var ErrNotFound = errors.New("repository: not found")

// Repo persists Satellites.
type Repo struct {
	q    *satmissiondb.Queries
	pool *pgxpool.Pool
}

// New constructs a Repo.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: satmissiondb.New(pool), pool: pool}
}

// RegisterSatelliteParams holds the input for [Repo.RegisterSatellite].
type RegisterSatelliteParams struct {
	ID                      ulid.ID
	TenantID                ulid.ID
	Name                    string
	NoradID                 string
	InternationalDesignator string
	ConfigJSON              string
	CreatedBy               string
}

// RegisterSatellite inserts a new satellite row.
func (r *Repo) RegisterSatellite(ctx context.Context, p RegisterSatelliteParams) (*models.Satellite, error) {
	row, err := r.q.RegisterSatellite(ctx, satmissiondb.RegisterSatelliteParams{
		ID:                      mapper.PgUUID(p.ID),
		TenantID:                mapper.PgUUID(p.TenantID),
		Name:                    p.Name,
		NoradID:                 p.NoradID,
		InternationalDesignator: p.InternationalDesignator,
		ConfigJson:              []byte(mapper.NormalizeJSON(p.ConfigJSON)),
		CreatedBy:               p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.SatelliteFromRow(row), nil
}

// GetSatellite returns a satellite by id.
func (r *Repo) GetSatellite(ctx context.Context, id ulid.ID) (*models.Satellite, error) {
	row, err := r.q.GetSatellite(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.SatelliteFromRow(row), nil
}

// ListSatellitesForTenant returns one page of satellites.
func (r *Repo) ListSatellitesForTenant(
	ctx context.Context, tenantID ulid.ID, offset, size int32,
) ([]*models.Satellite, int32, error) {
	total, err := r.q.CountSatellitesForTenant(ctx, mapper.PgUUID(tenantID))
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListSatellitesForTenant(ctx, satmissiondb.ListSatellitesForTenantParams{
		TenantID:   mapper.PgUUID(tenantID),
		PageOffset: offset,
		PageSize:   size,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.Satellite, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.SatelliteFromRow(row))
	}
	return out, int32(total), nil
}

// UpdateTLE updates the TLE lines for a satellite.
func (r *Repo) UpdateTLE(ctx context.Context, id ulid.ID, line1, line2, updatedBy string) (*models.Satellite, error) {
	row, err := r.q.UpdateTLE(ctx, satmissiondb.UpdateTLEParams{
		ID:        mapper.PgUUID(id),
		TleLine1:  line1,
		TleLine2:  line2,
		UpdatedBy: updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.SatelliteFromRow(row), nil
}

// UpdateOrbitalState updates the last_state_* columns.
func (r *Repo) UpdateOrbitalState(ctx context.Context, id ulid.ID, s models.OrbitalState, updatedBy string) (*models.Satellite, error) {
	row, err := r.q.UpdateOrbitalState(ctx, satmissiondb.UpdateOrbitalStateParams{
		ID:        mapper.PgUUID(id),
		RxKm:      s.RxKm,
		RyKm:      s.RyKm,
		RzKm:      s.RzKm,
		VxKmS:     s.VxKmS,
		VyKmS:     s.VyKmS,
		VzKmS:     s.VzKmS,
		Epoch:     mapper.PgTimestamp(s.Epoch),
		UpdatedBy: updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.SatelliteFromRow(row), nil
}

// SetMode updates the satellite's current mode.
func (r *Repo) SetMode(ctx context.Context, id ulid.ID, mode models.SatelliteMode, updatedBy string) (*models.Satellite, error) {
	row, err := r.q.SetMode(ctx, satmissiondb.SetModeParams{
		ID:        mapper.PgUUID(id),
		Mode:      int32(mode),
		UpdatedBy: updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.SatelliteFromRow(row), nil
}
