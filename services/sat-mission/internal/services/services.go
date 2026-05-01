// Package services holds sat-mission business logic.
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	"github.com/ppusapati/space/services/sat-mission/internal/models"
	"github.com/ppusapati/space/services/sat-mission/internal/repository"
)

// Mission is the sat-mission service-layer facade.
type Mission struct {
	repo  *repository.Repo
	IDFn  func() ulid.ID
	NowFn func() time.Time
}

// New constructs a Mission service.
func New(repo *repository.Repo) *Mission {
	return &Mission{
		repo:  repo,
		IDFn:  ulid.NewMonotonic,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// RegisterSatelliteInput is the input for [Mission.RegisterSatellite].
type RegisterSatelliteInput struct {
	TenantID                ulid.ID
	Name                    string
	NoradID                 string
	InternationalDesignator string
	ConfigJSON              string
	CreatedBy               string
}

// RegisterSatellite persists a new satellite.
func (m *Mission) RegisterSatellite(ctx context.Context, in RegisterSatelliteInput) (*models.Satellite, error) {
	if in.TenantID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "name required")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return m.repo.RegisterSatellite(ctx, repository.RegisterSatelliteParams{
		ID:                      m.IDFn(),
		TenantID:                in.TenantID,
		Name:                    in.Name,
		NoradID:                 in.NoradID,
		InternationalDesignator: in.InternationalDesignator,
		ConfigJSON:              in.ConfigJSON,
		CreatedBy:               createdBy,
	})
}

// GetSatellite fetches a satellite by id.
func (m *Mission) GetSatellite(ctx context.Context, id ulid.ID) (*models.Satellite, error) {
	s, err := m.repo.GetSatellite(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("SATELLITE_NOT_FOUND", "satellite "+id.String())
	}
	return s, err
}

// ListSatellitesForTenant returns one page of satellites.
func (m *Mission) ListSatellitesForTenant(
	ctx context.Context, tenantID ulid.ID, offset, size int32,
) ([]*models.Satellite, models.Page, error) {
	if tenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := m.repo.ListSatellitesForTenant(ctx, tenantID, offset, size)
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

// UpdateTLE validates the TLE and updates the satellite. Each line must be
// exactly 69 characters and start with the line-number prefix '1' or '2'.
func (m *Mission) UpdateTLE(ctx context.Context, id ulid.ID, line1, line2, updatedBy string) (*models.Satellite, error) {
	if id.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "id required")
	}
	if err := validateTLE(line1, line2); err != nil {
		return nil, err
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	s, err := m.repo.UpdateTLE(ctx, id, line1, line2, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("SATELLITE_NOT_FOUND", "satellite "+id.String())
	}
	return s, err
}

// validateTLE enforces TLE format invariants.
func validateTLE(line1, line2 string) error {
	if len(line1) != 69 || len(line2) != 69 {
		return pkgerrors.BadRequest("INVALID_ARGUMENT", "tle lines must each be 69 chars")
	}
	if line1[0] != '1' {
		return pkgerrors.BadRequest("INVALID_ARGUMENT", "tle_line1 must start with '1 '")
	}
	if line2[0] != '2' {
		return pkgerrors.BadRequest("INVALID_ARGUMENT", "tle_line2 must start with '2 '")
	}
	return nil
}

// UpdateOrbitalState updates the last orbital state on a satellite.
func (m *Mission) UpdateOrbitalState(ctx context.Context, id ulid.ID, s models.OrbitalState, updatedBy string) (*models.Satellite, error) {
	if id.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "id required")
	}
	if !s.Valid || s.Epoch.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "orbital state and epoch required")
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	out, err := m.repo.UpdateOrbitalState(ctx, id, s, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("SATELLITE_NOT_FOUND", "satellite "+id.String())
	}
	return out, err
}

// SetMode updates the satellite's current mode.
func (m *Mission) SetMode(ctx context.Context, id ulid.ID, mode models.SatelliteMode, updatedBy string) (*models.Satellite, error) {
	if id.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "id required")
	}
	if mode == models.ModeUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "mode required")
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	out, err := m.repo.SetMode(ctx, id, mode, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("SATELLITE_NOT_FOUND", "satellite "+id.String())
	}
	return out, err
}
