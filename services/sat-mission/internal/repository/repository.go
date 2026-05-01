// Package repository wraps the sat-mission sqlc layer.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	satmissiondb "github.com/ppusapati/space/services/sat-mission/db/generated"
	"github.com/ppusapati/space/services/sat-mission/internal/mappers"
	"github.com/ppusapati/space/services/sat-mission/internal/models"
)

// ErrNotFound is returned when no row matches.
var ErrNotFound = errors.New("repository: not found")

// SatelliteRepository persists Satellites.
type SatelliteRepository struct {
	q    *satmissiondb.Queries
	pool *pgxpool.Pool
}

// NewSatelliteRepository constructs a SatelliteRepository.
func NewSatelliteRepository(pool *pgxpool.Pool) *SatelliteRepository {
	return &SatelliteRepository{q: satmissiondb.New(pool), pool: pool}
}

// Register inserts a new Satellite.
func (r *SatelliteRepository) Register(ctx context.Context, s *models.Satellite) (*models.Satellite, error) {
	row, err := r.q.RegisterSatellite(ctx, satmissiondb.RegisterSatelliteParams{
		ID:                      mappers.PgUUID(s.ID),
		TenantID:                mappers.PgUUID(s.TenantID),
		Name:                    s.Name,
		NoradID:                 s.NORADID,
		InternationalDesignator: s.InternationalDesignator,
		ConfigJson:              []byte(s.ConfigJSON),
		CreatedBy:               s.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mappers.SatelliteFromRow(row), nil
}

// Get fetches by id.
func (r *SatelliteRepository) Get(ctx context.Context, id uuid.UUID) (*models.Satellite, error) {
	row, err := r.q.GetSatellite(ctx, mappers.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mappers.SatelliteFromRow(row), nil
}

// List returns one page of Satellites for a tenant.
func (r *SatelliteRepository) List(
	ctx context.Context, tenantID uuid.UUID, cursorTS *time.Time, cursorID uuid.UUID, limit int32,
) ([]*models.Satellite, error) {
	rows, err := r.q.ListSatellites(ctx, satmissiondb.ListSatellitesParams{
		TenantID:        mappers.PgUUID(tenantID),
		CursorCreatedAt: mappers.PgTimestampPtr(cursorTS),
		CursorID:        mappers.PgUUID(cursorID),
		Lim:             limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*models.Satellite, 0, len(rows))
	for _, row := range rows {
		out = append(out, mappers.SatelliteFromRow(row))
	}
	return out, nil
}

// UpdateTLE writes new TLE lines.
func (r *SatelliteRepository) UpdateTLE(ctx context.Context, id uuid.UUID, line1, line2, updatedBy string) (*models.Satellite, error) {
	row, err := r.q.UpdateTLE(ctx, satmissiondb.UpdateTLEParams{
		ID:        mappers.PgUUID(id),
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
	return mappers.SatelliteFromRow(row), nil
}

// UpdateOrbitalState writes a new orbital state snapshot.
func (r *SatelliteRepository) UpdateOrbitalState(ctx context.Context, id uuid.UUID, state models.OrbitalState, updatedBy string) (*models.Satellite, error) {
	rx := state.RxKm
	ry := state.RyKm
	rz := state.RzKm
	vx := state.VxKmS
	vy := state.VyKmS
	vz := state.VzKmS
	row, err := r.q.UpdateOrbitalState(ctx, satmissiondb.UpdateOrbitalStateParams{
		ID:                mappers.PgUUID(id),
		LastStateRxKm:     &rx,
		LastStateRyKm:     &ry,
		LastStateRzKm:     &rz,
		LastStateVxKmS:    &vx,
		LastStateVyKmS:    &vy,
		LastStateVzKmS:    &vz,
		LastStateEpoch:    mappers.PgTimestampPtr(&state.Epoch),
		UpdatedBy:         updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mappers.SatelliteFromRow(row), nil
}

// SetMode updates the current mode.
func (r *SatelliteRepository) SetMode(ctx context.Context, id uuid.UUID, mode models.SatelliteMode, updatedBy string) (*models.Satellite, error) {
	row, err := r.q.SetMode(ctx, satmissiondb.SetModeParams{
		ID:          mappers.PgUUID(id),
		CurrentMode: int32(mode),
		UpdatedBy:   updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mappers.SatelliteFromRow(row), nil
}
