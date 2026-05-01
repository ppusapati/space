// Package repository wraps the sat-fsw sqlc layer.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	satfswdb "github.com/ppusapati/space/services/sat-fsw/db/generated"
	"github.com/ppusapati/space/services/sat-fsw/internal/mappers"
	"github.com/ppusapati/space/services/sat-fsw/internal/models"
)

// ErrNotFound is returned when no row matches.
var ErrNotFound = errors.New("repository: not found")

// Repo persists FirmwareBuilds and DeploymentManifests.
type Repo struct {
	q    *satfswdb.Queries
	pool *pgxpool.Pool
}

// New constructs a Repo.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: satfswdb.New(pool), pool: pool}
}

// ----- FirmwareBuild -------------------------------------------------------

// RegisterFirmwareBuildParams is the input to RegisterFirmwareBuild.
type RegisterFirmwareBuildParams struct {
	ID                uuid.UUID
	TenantID          uuid.UUID
	TargetPlatform    string
	Subsystem         string
	Version           string
	GitSHA            string
	ArtefactURI       string
	ArtefactSizeBytes uint64
	ArtefactSHA256    string
	Status            models.FirmwareBuildStatus
	Notes             string
	CreatedBy         string
}

// RegisterFirmwareBuild inserts a new build row.
func (r *Repo) RegisterFirmwareBuild(ctx context.Context, p RegisterFirmwareBuildParams) (*models.FirmwareBuild, error) {
	row, err := r.q.RegisterFirmwareBuild(ctx, satfswdb.RegisterFirmwareBuildParams{
		ID:                mappers.PgUUID(p.ID),
		TenantID:          mappers.PgUUID(p.TenantID),
		TargetPlatform:    p.TargetPlatform,
		Subsystem:         p.Subsystem,
		Version:           p.Version,
		GitSha:            p.GitSHA,
		ArtefactUri:       p.ArtefactURI,
		ArtefactSizeBytes: int64(p.ArtefactSizeBytes),
		ArtefactSha256:    p.ArtefactSHA256,
		Status:            int32(p.Status),
		Notes:             p.Notes,
		CreatedBy:         p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mappers.FirmwareBuildFromRow(row), nil
}

// GetFirmwareBuild returns a build by id.
func (r *Repo) GetFirmwareBuild(ctx context.Context, id uuid.UUID) (*models.FirmwareBuild, error) {
	row, err := r.q.GetFirmwareBuild(ctx, mappers.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mappers.FirmwareBuildFromRow(row), nil
}

// ListFirmwareBuildsParams is the input to ListFirmwareBuilds.
type ListFirmwareBuildsParams struct {
	TenantID      uuid.UUID
	Subsystem     *string
	Status        *models.FirmwareBuildStatus
	CursorCreated *time.Time
	CursorID      uuid.UUID
	Limit         int32
}

// ListFirmwareBuilds returns one page of builds.
func (r *Repo) ListFirmwareBuilds(ctx context.Context, p ListFirmwareBuildsParams) ([]*models.FirmwareBuild, error) {
	var statusPtr *int32
	if p.Status != nil {
		v := int32(*p.Status)
		statusPtr = &v
	}
	rows, err := r.q.ListFirmwareBuilds(ctx, satfswdb.ListFirmwareBuildsParams{
		TenantID:        mappers.PgUUID(p.TenantID),
		Subsystem:       p.Subsystem,
		Status:          statusPtr,
		CursorCreatedAt: mappers.PgTimestampPtr(p.CursorCreated),
		CursorID:        mappers.PgUUID(p.CursorID),
		Lim:             p.Limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*models.FirmwareBuild, 0, len(rows))
	for _, row := range rows {
		out = append(out, mappers.FirmwareBuildFromRow(row))
	}
	return out, nil
}

// UpdateFirmwareBuildStatus updates the status field.
func (r *Repo) UpdateFirmwareBuildStatus(
	ctx context.Context, id uuid.UUID, status models.FirmwareBuildStatus, notes, updatedBy string,
) (*models.FirmwareBuild, error) {
	row, err := r.q.UpdateFirmwareBuildStatus(ctx, satfswdb.UpdateFirmwareBuildStatusParams{
		ID:        mappers.PgUUID(id),
		Status:    int32(status),
		Notes:     notes,
		UpdatedBy: updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mappers.FirmwareBuildFromRow(row), nil
}

// ----- DeploymentManifest --------------------------------------------------

// CreateDeploymentManifestParams is the input to CreateDeploymentManifest.
type CreateDeploymentManifestParams struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	SatelliteID     uuid.UUID
	ManifestVersion string
	Status          models.DeploymentStatus
	Assignments     map[string]string
	Notes           string
	CreatedBy       string
}

// CreateDeploymentManifest inserts a new manifest row.
func (r *Repo) CreateDeploymentManifest(ctx context.Context, p CreateDeploymentManifestParams) (*models.DeploymentManifest, error) {
	js, err := mappers.AssignmentsToJSON(p.Assignments)
	if err != nil {
		return nil, err
	}
	row, err := r.q.CreateDeploymentManifest(ctx, satfswdb.CreateDeploymentManifestParams{
		ID:              mappers.PgUUID(p.ID),
		TenantID:        mappers.PgUUID(p.TenantID),
		SatelliteID:     mappers.PgUUID(p.SatelliteID),
		ManifestVersion: p.ManifestVersion,
		Status:          int32(p.Status),
		AssignmentsJson: js,
		Notes:           p.Notes,
		CreatedBy:       p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mappers.DeploymentManifestFromRow(row)
}

// GetDeploymentManifest returns a manifest by id.
func (r *Repo) GetDeploymentManifest(ctx context.Context, id uuid.UUID) (*models.DeploymentManifest, error) {
	row, err := r.q.GetDeploymentManifest(ctx, mappers.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mappers.DeploymentManifestFromRow(row)
}

// ListDeploymentManifestsParams is the input to ListDeploymentManifests.
type ListDeploymentManifestsParams struct {
	TenantID      uuid.UUID
	SatelliteID   *uuid.UUID
	Status        *models.DeploymentStatus
	CursorCreated *time.Time
	CursorID      uuid.UUID
	Limit         int32
}

// ListDeploymentManifests returns one page of manifests.
func (r *Repo) ListDeploymentManifests(ctx context.Context, p ListDeploymentManifestsParams) ([]*models.DeploymentManifest, error) {
	var statusPtr *int32
	if p.Status != nil {
		v := int32(*p.Status)
		statusPtr = &v
	}
	rows, err := r.q.ListDeploymentManifests(ctx, satfswdb.ListDeploymentManifestsParams{
		TenantID:        mappers.PgUUID(p.TenantID),
		SatelliteID:     mappers.PgUUIDPtr(p.SatelliteID),
		Status:          statusPtr,
		CursorCreatedAt: mappers.PgTimestampPtr(p.CursorCreated),
		CursorID:        mappers.PgUUID(p.CursorID),
		Lim:             p.Limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*models.DeploymentManifest, 0, len(rows))
	for _, row := range rows {
		m, err := mappers.DeploymentManifestFromRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

// UpdateDeploymentManifestStatus updates the status field.
func (r *Repo) UpdateDeploymentManifestStatus(
	ctx context.Context, id uuid.UUID, status models.DeploymentStatus, notes, updatedBy string,
) (*models.DeploymentManifest, error) {
	row, err := r.q.UpdateDeploymentManifestStatus(ctx, satfswdb.UpdateDeploymentManifestStatusParams{
		ID:        mappers.PgUUID(id),
		Status:    int32(status),
		Notes:     notes,
		UpdatedBy: updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mappers.DeploymentManifestFromRow(row)
}
