// Package repository wraps the sat-fsw sqlc layer.
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/chetana/packages/ulid"

	satfswdb "github.com/ppusapati/space/services/sat-fsw/db/generated"
	"github.com/ppusapati/space/services/sat-fsw/internal/mapper"
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

// ----- FirmwareBuilds ------------------------------------------------------

// RegisterFirmwareBuildParams holds the input for [Repo.RegisterFirmwareBuild].
type RegisterFirmwareBuildParams struct {
	ID                ulid.ID
	TenantID          ulid.ID
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

// RegisterFirmwareBuild inserts a new firmware_build row.
func (r *Repo) RegisterFirmwareBuild(ctx context.Context, p RegisterFirmwareBuildParams) (*models.FirmwareBuild, error) {
	row, err := r.q.RegisterFirmwareBuild(ctx, satfswdb.RegisterFirmwareBuildParams{
		ID:                mapper.PgUUID(p.ID),
		TenantID:          mapper.PgUUID(p.TenantID),
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
	return mapper.FirmwareBuildFromRow(row), nil
}

// GetFirmwareBuild returns a build by id.
func (r *Repo) GetFirmwareBuild(ctx context.Context, id ulid.ID) (*models.FirmwareBuild, error) {
	row, err := r.q.GetFirmwareBuild(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.FirmwareBuildFromRow(row), nil
}

// ListFirmwareBuildsParams holds the input for [Repo.ListFirmwareBuildsForTenant].
type ListFirmwareBuildsParams struct {
	TenantID   ulid.ID
	Subsystem  string
	Status     *models.FirmwareBuildStatus
	PageOffset int32
	PageSize   int32
}

// ListFirmwareBuildsForTenant returns one page of builds.
func (r *Repo) ListFirmwareBuildsForTenant(ctx context.Context, p ListFirmwareBuildsParams) ([]*models.FirmwareBuild, int32, error) {
	var statusPtr *int32
	if p.Status != nil {
		v := int32(*p.Status)
		statusPtr = &v
	}
	var subsystemPtr *string
	if p.Subsystem != "" {
		v := p.Subsystem
		subsystemPtr = &v
	}
	total, err := r.q.CountFirmwareBuildsForTenant(ctx, satfswdb.CountFirmwareBuildsForTenantParams{
		TenantID:  mapper.PgUUID(p.TenantID),
		Subsystem: subsystemPtr,
		Status:    statusPtr,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListFirmwareBuildsForTenant(ctx, satfswdb.ListFirmwareBuildsForTenantParams{
		TenantID:   mapper.PgUUID(p.TenantID),
		Subsystem:  subsystemPtr,
		Status:     statusPtr,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.FirmwareBuild, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.FirmwareBuildFromRow(row))
	}
	return out, int32(total), nil
}

// UpdateFirmwareBuildStatus updates the status of a build.
func (r *Repo) UpdateFirmwareBuildStatus(
	ctx context.Context, id ulid.ID, status models.FirmwareBuildStatus, notes, updatedBy string,
) (*models.FirmwareBuild, error) {
	row, err := r.q.UpdateFirmwareBuildStatus(ctx, satfswdb.UpdateFirmwareBuildStatusParams{
		ID:        mapper.PgUUID(id),
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
	return mapper.FirmwareBuildFromRow(row), nil
}

// ----- DeploymentManifests -------------------------------------------------

// CreateDeploymentManifestParams holds the input for
// [Repo.CreateDeploymentManifest].
type CreateDeploymentManifestParams struct {
	ID              ulid.ID
	TenantID        ulid.ID
	SatelliteID     ulid.ID
	ManifestVersion string
	Status          models.DeploymentStatus
	Assignments     map[string]string
	Notes           string
	CreatedBy       string
}

// CreateDeploymentManifest inserts a new deployment_manifest row.
func (r *Repo) CreateDeploymentManifest(ctx context.Context, p CreateDeploymentManifestParams) (*models.DeploymentManifest, error) {
	row, err := r.q.CreateDeploymentManifest(ctx, satfswdb.CreateDeploymentManifestParams{
		ID:              mapper.PgUUID(p.ID),
		TenantID:        mapper.PgUUID(p.TenantID),
		SatelliteID:     mapper.PgUUID(p.SatelliteID),
		ManifestVersion: p.ManifestVersion,
		Status:          int32(p.Status),
		AssignmentsJson: mapper.AssignmentsToJSON(p.Assignments),
		Notes:           p.Notes,
		CreatedBy:       p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.DeploymentManifestFromRow(row), nil
}

// GetDeploymentManifest returns a manifest by id.
func (r *Repo) GetDeploymentManifest(ctx context.Context, id ulid.ID) (*models.DeploymentManifest, error) {
	row, err := r.q.GetDeploymentManifest(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.DeploymentManifestFromRow(row), nil
}

// ListDeploymentManifestsParams holds the input for
// [Repo.ListDeploymentManifestsForTenant].
type ListDeploymentManifestsParams struct {
	TenantID    ulid.ID
	SatelliteID *ulid.ID
	Status      *models.DeploymentStatus
	PageOffset  int32
	PageSize    int32
}

// ListDeploymentManifestsForTenant returns one page of manifests.
func (r *Repo) ListDeploymentManifestsForTenant(ctx context.Context, p ListDeploymentManifestsParams) ([]*models.DeploymentManifest, int32, error) {
	var statusPtr *int32
	if p.Status != nil {
		v := int32(*p.Status)
		statusPtr = &v
	}
	var satellitePg pgtype.UUID
	if p.SatelliteID != nil {
		satellitePg = mapper.PgUUID(*p.SatelliteID)
	}
	total, err := r.q.CountDeploymentManifestsForTenant(ctx, satfswdb.CountDeploymentManifestsForTenantParams{
		TenantID:    mapper.PgUUID(p.TenantID),
		SatelliteID: satellitePg,
		Status:      statusPtr,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListDeploymentManifestsForTenant(ctx, satfswdb.ListDeploymentManifestsForTenantParams{
		TenantID:    mapper.PgUUID(p.TenantID),
		SatelliteID: satellitePg,
		Status:      statusPtr,
		PageOffset:  p.PageOffset,
		PageSize:    p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.DeploymentManifest, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.DeploymentManifestFromRow(row))
	}
	return out, int32(total), nil
}

// UpdateDeploymentManifestStatus updates the status of a manifest.
func (r *Repo) UpdateDeploymentManifestStatus(
	ctx context.Context, id ulid.ID, status models.DeploymentStatus, notes, updatedBy string,
) (*models.DeploymentManifest, error) {
	row, err := r.q.UpdateDeploymentManifestStatus(ctx, satfswdb.UpdateDeploymentManifestStatusParams{
		ID:        mapper.PgUUID(id),
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
	return mapper.DeploymentManifestFromRow(row), nil
}
