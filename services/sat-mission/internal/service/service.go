// Package service holds sat-mission business logic.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/ppusapati/space/pkg/errs"
	"github.com/ppusapati/space/services/sat-mission/internal/models"
	"github.com/ppusapati/space/services/sat-mission/internal/repository"
)

// Mission is the service-layer facade.
type Mission struct {
	repo  *repository.SatelliteRepository
	IDFn  func() uuid.UUID
	NowFn func() time.Time
}

// New constructs a Mission service.
func New(repo *repository.SatelliteRepository) *Mission {
	return &Mission{
		repo:  repo,
		IDFn:  uuid.New,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// RegisterSatelliteInput is the input to [Mission.RegisterSatellite].
type RegisterSatelliteInput struct {
	TenantID                uuid.UUID
	Name                    string
	NORADID                 string
	InternationalDesignator string
	ConfigJSON              string
	CreatedBy               string
}

// RegisterSatellite persists a new Satellite.
func (m *Mission) RegisterSatellite(ctx context.Context, in RegisterSatelliteInput) (*models.Satellite, error) {
	if in.TenantID == uuid.Nil || in.Name == "" {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id and name required")
	}
	if in.ConfigJSON == "" {
		in.ConfigJSON = "{}"
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	s := &models.Satellite{
		ID:                      m.IDFn(),
		TenantID:                in.TenantID,
		Name:                    in.Name,
		NORADID:                 in.NORADID,
		InternationalDesignator: in.InternationalDesignator,
		ConfigJSON:              in.ConfigJSON,
		Active:                  true,
		CurrentMode:             models.ModeSafe,
		CreatedBy:               createdBy,
	}
	return m.repo.Register(ctx, s)
}

// GetSatellite fetches by id.
func (m *Mission) GetSatellite(ctx context.Context, id uuid.UUID) (*models.Satellite, error) {
	s, err := m.repo.Get(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errs.New(errs.DomainNotFound, "satellite %s", id)
	}
	return s, err
}

// ListSatellites returns one page of satellites.
func (m *Mission) ListSatellites(
	ctx context.Context, tenantID uuid.UUID, cursorTS *time.Time, cursorID uuid.UUID, limit int32,
) ([]*models.Satellite, error) {
	if tenantID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id required")
	}
	return m.repo.List(ctx, tenantID, cursorTS, cursorID, limit)
}

// UpdateTLE replaces the active TLE pair.
func (m *Mission) UpdateTLE(ctx context.Context, id uuid.UUID, line1, line2, updatedBy string) (*models.Satellite, error) {
	if len(line1) != 69 || len(line2) != 69 {
		return nil, errs.New(errs.DomainInvalidArgument, "TLE lines must be 69 characters")
	}
	if line1[0] != '1' || line2[0] != '2' {
		return nil, errs.New(errs.DomainInvalidArgument, "TLE line numbers must be '1' and '2'")
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	s, err := m.repo.UpdateTLE(ctx, id, line1, line2, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errs.New(errs.DomainNotFound, "satellite %s", id)
	}
	return s, err
}

// UpdateOrbitalState writes a new state snapshot.
func (m *Mission) UpdateOrbitalState(ctx context.Context, id uuid.UUID, state models.OrbitalState, updatedBy string) (*models.Satellite, error) {
	if !state.Valid {
		return nil, errs.New(errs.DomainInvalidArgument, "orbital state required")
	}
	if state.Epoch.IsZero() {
		return nil, errs.New(errs.DomainInvalidArgument, "epoch required")
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	s, err := m.repo.UpdateOrbitalState(ctx, id, state, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errs.New(errs.DomainNotFound, "satellite %s", id)
	}
	return s, err
}

// SetMode updates the current mode.
func (m *Mission) SetMode(ctx context.Context, id uuid.UUID, mode models.SatelliteMode, updatedBy string) (*models.Satellite, error) {
	if mode == models.ModeUnspecified {
		return nil, errs.New(errs.DomainInvalidArgument, "mode required")
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	s, err := m.repo.SetMode(ctx, id, mode, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errs.New(errs.DomainNotFound, "satellite %s", id)
	}
	return s, err
}
