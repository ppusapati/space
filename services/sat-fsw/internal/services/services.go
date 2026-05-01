// Package services holds sat-fsw business logic.
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	"github.com/ppusapati/space/services/sat-fsw/internal/models"
	"github.com/ppusapati/space/services/sat-fsw/internal/repository"
)

// FlightSoftware is the sat-fsw service-layer facade.
type FlightSoftware struct {
	repo  *repository.Repo
	IDFn  func() ulid.ID
	NowFn func() time.Time
}

// New constructs a FlightSoftware service.
func New(repo *repository.Repo) *FlightSoftware {
	return &FlightSoftware{
		repo:  repo,
		IDFn:  ulid.NewMonotonic,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// ----- FirmwareBuilds ------------------------------------------------------

// RegisterFirmwareBuildInput is the input for [FlightSoftware.RegisterFirmwareBuild].
type RegisterFirmwareBuildInput struct {
	TenantID          ulid.ID
	TargetPlatform    string
	Subsystem         string
	Version           string
	GitSHA            string
	ArtefactURI       string
	ArtefactSizeBytes uint64
	ArtefactSHA256    string
	Notes             string
	CreatedBy         string
}

// RegisterFirmwareBuild persists a new build with status BUILDING.
func (s *FlightSoftware) RegisterFirmwareBuild(ctx context.Context, in RegisterFirmwareBuildInput) (*models.FirmwareBuild, error) {
	if in.TenantID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	in.TargetPlatform = strings.TrimSpace(in.TargetPlatform)
	in.Subsystem = strings.TrimSpace(in.Subsystem)
	in.Version = strings.TrimSpace(in.Version)
	in.GitSHA = strings.ToLower(strings.TrimSpace(in.GitSHA))
	in.ArtefactURI = strings.TrimSpace(in.ArtefactURI)
	in.ArtefactSHA256 = strings.ToLower(strings.TrimSpace(in.ArtefactSHA256))
	if in.TargetPlatform == "" || in.Subsystem == "" || in.Version == "" || in.ArtefactURI == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "target_platform, subsystem, version, artefact_uri required")
	}
	if !isHex(in.GitSHA, 40) {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "git_sha must be 40-char hex")
	}
	if !isHex(in.ArtefactSHA256, 64) {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "artefact_sha256 must be 64-char hex")
	}
	if in.ArtefactSizeBytes == 0 {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "artefact_size_bytes must be > 0")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return s.repo.RegisterFirmwareBuild(ctx, repository.RegisterFirmwareBuildParams{
		ID:                s.IDFn(),
		TenantID:          in.TenantID,
		TargetPlatform:    in.TargetPlatform,
		Subsystem:         in.Subsystem,
		Version:           in.Version,
		GitSHA:            in.GitSHA,
		ArtefactURI:       in.ArtefactURI,
		ArtefactSizeBytes: in.ArtefactSizeBytes,
		ArtefactSHA256:    in.ArtefactSHA256,
		Status:            models.BuildBuilding,
		Notes:             in.Notes,
		CreatedBy:         createdBy,
	})
}

// GetFirmwareBuild fetches a build by id.
func (s *FlightSoftware) GetFirmwareBuild(ctx context.Context, id ulid.ID) (*models.FirmwareBuild, error) {
	b, err := s.repo.GetFirmwareBuild(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("FIRMWARE_BUILD_NOT_FOUND", "firmware_build "+id.String())
	}
	return b, err
}

// ListFirmwareBuildsInput is the input for [FlightSoftware.ListFirmwareBuildsForTenant].
type ListFirmwareBuildsInput struct {
	TenantID   ulid.ID
	Subsystem  string
	Status     *models.FirmwareBuildStatus
	PageOffset int32
	PageSize   int32
}

// ListFirmwareBuildsForTenant returns one page of builds.
func (s *FlightSoftware) ListFirmwareBuildsForTenant(ctx context.Context, in ListFirmwareBuildsInput) ([]*models.FirmwareBuild, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := s.repo.ListFirmwareBuildsForTenant(ctx, repository.ListFirmwareBuildsParams{
		TenantID:   in.TenantID,
		Subsystem:  in.Subsystem,
		Status:     in.Status,
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

// UpdateFirmwareBuildStatus enforces the build-status transition graph.
//
//	BUILDING   -> READY | REJECTED
//	READY      -> DEPRECATED
//	REJECTED, DEPRECATED — terminal.
func (s *FlightSoftware) UpdateFirmwareBuildStatus(
	ctx context.Context, id ulid.ID, status models.FirmwareBuildStatus, notes, updatedBy string,
) (*models.FirmwareBuild, error) {
	if status == models.BuildUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "status required")
	}
	current, err := s.GetFirmwareBuild(ctx, id)
	if err != nil {
		return nil, err
	}
	if !validBuildTransition(current.Status, status) {
		return nil, pkgerrors.New(412, "ILLEGAL_TRANSITION",
			"illegal firmware_build status transition")
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	updated, err := s.repo.UpdateFirmwareBuildStatus(ctx, id, status, notes, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("FIRMWARE_BUILD_NOT_FOUND", "firmware_build "+id.String())
	}
	return updated, err
}

func validBuildTransition(from, to models.FirmwareBuildStatus) bool {
	switch from {
	case models.BuildBuilding:
		return to == models.BuildReady || to == models.BuildRejected
	case models.BuildReady:
		return to == models.BuildDeprecated
	case models.BuildRejected, models.BuildDeprecated:
		return false
	default:
		return false
	}
}

// ----- DeploymentManifests -------------------------------------------------

// CreateDeploymentManifestInput is the input for
// [FlightSoftware.CreateDeploymentManifest].
type CreateDeploymentManifestInput struct {
	TenantID        ulid.ID
	SatelliteID     ulid.ID
	ManifestVersion string
	Assignments     map[string]string
	Notes           string
	CreatedBy       string
}

// CreateDeploymentManifest persists a manifest in DRAFT status. It validates
// each assigned firmware_build:
//   - exists
//   - belongs to the same tenant
//   - targets the named subsystem
//   - is in READY status
func (s *FlightSoftware) CreateDeploymentManifest(ctx context.Context, in CreateDeploymentManifestInput) (*models.DeploymentManifest, error) {
	if in.TenantID.IsZero() || in.SatelliteID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id and satellite_id required")
	}
	in.ManifestVersion = strings.TrimSpace(in.ManifestVersion)
	if in.ManifestVersion == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "manifest_version required")
	}
	if len(in.Assignments) == 0 {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "assignments required")
	}
	for subsystem, buildIDStr := range in.Assignments {
		subsystem = strings.TrimSpace(subsystem)
		if subsystem == "" {
			return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "assignment subsystem must not be empty")
		}
		buildID, err := ulid.Parse(strings.TrimSpace(buildIDStr))
		if err != nil {
			return nil, pkgerrors.BadRequest("INVALID_ARGUMENT",
				"assignment value for subsystem "+subsystem+" must be a ULID")
		}
		b, err := s.repo.GetFirmwareBuild(ctx, buildID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, pkgerrors.BadRequest("FIRMWARE_BUILD_NOT_FOUND",
					"firmware_build "+buildID.String()+" missing for subsystem "+subsystem)
			}
			return nil, err
		}
		if b.TenantID != in.TenantID {
			return nil, pkgerrors.BadRequest("INVALID_ARGUMENT",
				"firmware_build tenant mismatch for subsystem "+subsystem)
		}
		if b.Subsystem != subsystem {
			return nil, pkgerrors.BadRequest("INVALID_ARGUMENT",
				"firmware_build subsystem mismatch for "+subsystem)
		}
		if b.Status != models.BuildReady {
			return nil, pkgerrors.New(412, "FIRMWARE_BUILD_NOT_READY",
				"firmware_build for subsystem "+subsystem+" is not READY")
		}
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return s.repo.CreateDeploymentManifest(ctx, repository.CreateDeploymentManifestParams{
		ID:              s.IDFn(),
		TenantID:        in.TenantID,
		SatelliteID:     in.SatelliteID,
		ManifestVersion: in.ManifestVersion,
		Status:          models.DeploymentDraft,
		Assignments:     in.Assignments,
		Notes:           in.Notes,
		CreatedBy:       createdBy,
	})
}

// GetDeploymentManifest fetches a manifest by id.
func (s *FlightSoftware) GetDeploymentManifest(ctx context.Context, id ulid.ID) (*models.DeploymentManifest, error) {
	m, err := s.repo.GetDeploymentManifest(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("DEPLOYMENT_MANIFEST_NOT_FOUND", "deployment_manifest "+id.String())
	}
	return m, err
}

// ListDeploymentManifestsInput is the input for
// [FlightSoftware.ListDeploymentManifestsForTenant].
type ListDeploymentManifestsInput struct {
	TenantID    ulid.ID
	SatelliteID *ulid.ID
	Status      *models.DeploymentStatus
	PageOffset  int32
	PageSize    int32
}

// ListDeploymentManifestsForTenant returns one page of manifests.
func (s *FlightSoftware) ListDeploymentManifestsForTenant(ctx context.Context, in ListDeploymentManifestsInput) ([]*models.DeploymentManifest, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := s.repo.ListDeploymentManifestsForTenant(ctx, repository.ListDeploymentManifestsParams{
		TenantID:    in.TenantID,
		SatelliteID: in.SatelliteID,
		Status:      in.Status,
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

// UpdateDeploymentManifestStatus enforces the deployment-status transition graph:
//
//	DRAFT       -> APPROVED
//	APPROVED    -> DEPLOYED
//	DEPLOYED    -> ROLLED_BACK
//	ROLLED_BACK — terminal.
func (s *FlightSoftware) UpdateDeploymentManifestStatus(
	ctx context.Context, id ulid.ID, status models.DeploymentStatus, notes, updatedBy string,
) (*models.DeploymentManifest, error) {
	if status == models.DeploymentUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "status required")
	}
	current, err := s.GetDeploymentManifest(ctx, id)
	if err != nil {
		return nil, err
	}
	if !validDeploymentTransition(current.Status, status) {
		return nil, pkgerrors.New(412, "ILLEGAL_TRANSITION",
			"illegal deployment_manifest status transition")
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	updated, err := s.repo.UpdateDeploymentManifestStatus(ctx, id, status, notes, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("DEPLOYMENT_MANIFEST_NOT_FOUND", "deployment_manifest "+id.String())
	}
	return updated, err
}

func validDeploymentTransition(from, to models.DeploymentStatus) bool {
	switch from {
	case models.DeploymentDraft:
		return to == models.DeploymentApproved
	case models.DeploymentApproved:
		return to == models.DeploymentDeployed
	case models.DeploymentDeployed:
		return to == models.DeploymentRolledBack
	case models.DeploymentRolledBack:
		return false
	default:
		return false
	}
}

// isHex returns true when s is exactly n lowercase hex characters.
func isHex(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		ok := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')
		if !ok {
			return false
		}
	}
	return true
}
