// Package service holds sat-fsw business logic.
package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ppusapati/space/pkg/errs"
	"github.com/ppusapati/space/services/sat-fsw/internal/models"
	"github.com/ppusapati/space/services/sat-fsw/internal/repository"
)

// FSW is the service-layer facade.
type FSW struct {
	repo  *repository.Repo
	IDFn  func() uuid.UUID
	NowFn func() time.Time
}

// New constructs an FSW service.
func New(repo *repository.Repo) *FSW {
	return &FSW{
		repo:  repo,
		IDFn:  uuid.New,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// ----- FirmwareBuild -------------------------------------------------------

// RegisterFirmwareBuildInput is the input to [FSW.RegisterFirmwareBuild].
type RegisterFirmwareBuildInput struct {
	TenantID          uuid.UUID
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

// RegisterFirmwareBuild persists a new firmware build in BUILDING state.
func (s *FSW) RegisterFirmwareBuild(ctx context.Context, in RegisterFirmwareBuildInput) (*models.FirmwareBuild, error) {
	if in.TenantID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id required")
	}
	if in.TargetPlatform == "" || in.Subsystem == "" || in.Version == "" || in.ArtefactURI == "" {
		return nil, errs.New(errs.DomainInvalidArgument, "target_platform, subsystem, version and artefact_uri required")
	}
	if len(in.GitSHA) != 40 || !isHex(in.GitSHA) {
		return nil, errs.New(errs.DomainInvalidArgument, "git_sha must be a 40-char hex string")
	}
	if len(in.ArtefactSHA256) != 64 || !isHex(in.ArtefactSHA256) {
		return nil, errs.New(errs.DomainInvalidArgument, "artefact_sha256 must be a 64-char hex string")
	}
	if in.ArtefactSizeBytes == 0 {
		return nil, errs.New(errs.DomainInvalidArgument, "artefact_size_bytes must be > 0")
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
		GitSHA:            strings.ToLower(in.GitSHA),
		ArtefactURI:       in.ArtefactURI,
		ArtefactSizeBytes: in.ArtefactSizeBytes,
		ArtefactSHA256:    strings.ToLower(in.ArtefactSHA256),
		Status:            models.BuildStatusBuilding,
		Notes:             in.Notes,
		CreatedBy:         createdBy,
	})
}

// GetFirmwareBuild fetches a build by id.
func (s *FSW) GetFirmwareBuild(ctx context.Context, id uuid.UUID) (*models.FirmwareBuild, error) {
	b, err := s.repo.GetFirmwareBuild(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errs.New(errs.DomainNotFound, "firmware build %s", id)
	}
	return b, err
}

// ListFirmwareBuildsInput is the input to [FSW.ListFirmwareBuilds].
type ListFirmwareBuildsInput struct {
	TenantID      uuid.UUID
	Subsystem     *string
	Status        *models.FirmwareBuildStatus
	CursorCreated *time.Time
	CursorID      uuid.UUID
	Limit         int32
}

// ListFirmwareBuilds returns one page of builds.
func (s *FSW) ListFirmwareBuilds(ctx context.Context, in ListFirmwareBuildsInput) ([]*models.FirmwareBuild, error) {
	if in.TenantID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id required")
	}
	return s.repo.ListFirmwareBuilds(ctx, repository.ListFirmwareBuildsParams{
		TenantID:      in.TenantID,
		Subsystem:     in.Subsystem,
		Status:        in.Status,
		CursorCreated: in.CursorCreated,
		CursorID:      in.CursorID,
		Limit:         in.Limit,
	})
}

// UpdateFirmwareBuildStatus transitions a build to a new status. Transitions
// are validated against the current status to prevent illegal state changes.
func (s *FSW) UpdateFirmwareBuildStatus(
	ctx context.Context, id uuid.UUID, status models.FirmwareBuildStatus, notes, updatedBy string,
) (*models.FirmwareBuild, error) {
	if status == models.BuildStatusUnspecified {
		return nil, errs.New(errs.DomainInvalidArgument, "status required")
	}
	current, err := s.repo.GetFirmwareBuild(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errs.New(errs.DomainNotFound, "firmware build %s", id)
		}
		return nil, err
	}
	if !validBuildTransition(current.Status, status) {
		return nil, errs.New(errs.DomainPreconditionFailed,
			"illegal build status transition: %d -> %d", current.Status, status)
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	return s.repo.UpdateFirmwareBuildStatus(ctx, id, status, notes, updatedBy)
}

func validBuildTransition(from, to models.FirmwareBuildStatus) bool {
	switch from {
	case models.BuildStatusBuilding:
		return to == models.BuildStatusReady || to == models.BuildStatusRejected
	case models.BuildStatusReady:
		return to == models.BuildStatusDeprecated
	case models.BuildStatusRejected, models.BuildStatusDeprecated:
		return false
	default:
		return false
	}
}

// ----- DeploymentManifest --------------------------------------------------

// CreateDeploymentManifestInput is the input to [FSW.CreateDeploymentManifest].
type CreateDeploymentManifestInput struct {
	TenantID        uuid.UUID
	SatelliteID     uuid.UUID
	ManifestVersion string
	Assignments     map[string]string
	Notes           string
	CreatedBy       string
}

// CreateDeploymentManifest creates a manifest in DRAFT state. Each assigned
// firmware_build_id must be a parseable UUID and refer to a READY build owned
// by the same tenant.
func (s *FSW) CreateDeploymentManifest(ctx context.Context, in CreateDeploymentManifestInput) (*models.DeploymentManifest, error) {
	if in.TenantID == uuid.Nil || in.SatelliteID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id and satellite_id required")
	}
	if in.ManifestVersion == "" {
		return nil, errs.New(errs.DomainInvalidArgument, "manifest_version required")
	}
	if len(in.Assignments) == 0 {
		return nil, errs.New(errs.DomainInvalidArgument, "at least one subsystem assignment required")
	}
	for sub, idStr := range in.Assignments {
		if sub == "" {
			return nil, errs.New(errs.DomainInvalidArgument, "subsystem key must not be empty")
		}
		bid, err := uuid.Parse(idStr)
		if err != nil {
			return nil, errs.New(errs.DomainInvalidArgument, "assignment %q: %v", sub, err)
		}
		build, err := s.repo.GetFirmwareBuild(ctx, bid)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, errs.New(errs.DomainInvalidArgument, "assignment %q: build %s not found", sub, bid)
			}
			return nil, err
		}
		if build.TenantID != in.TenantID {
			return nil, errs.New(errs.DomainInvalidArgument,
				"assignment %q: build %s belongs to a different tenant", sub, bid)
		}
		if build.Status != models.BuildStatusReady {
			return nil, errs.New(errs.DomainPreconditionFailed,
				"assignment %q: build %s is not READY (status=%d)", sub, bid, build.Status)
		}
		if build.Subsystem != sub {
			return nil, errs.New(errs.DomainInvalidArgument,
				"assignment %q: build %s is for subsystem %q", sub, bid, build.Subsystem)
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
		Status:          models.DeploymentStatusDraft,
		Assignments:     in.Assignments,
		Notes:           in.Notes,
		CreatedBy:       createdBy,
	})
}

// GetDeploymentManifest fetches a manifest by id.
func (s *FSW) GetDeploymentManifest(ctx context.Context, id uuid.UUID) (*models.DeploymentManifest, error) {
	m, err := s.repo.GetDeploymentManifest(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errs.New(errs.DomainNotFound, "deployment manifest %s", id)
	}
	return m, err
}

// ListDeploymentManifestsInput is the input to [FSW.ListDeploymentManifests].
type ListDeploymentManifestsInput struct {
	TenantID      uuid.UUID
	SatelliteID   *uuid.UUID
	Status        *models.DeploymentStatus
	CursorCreated *time.Time
	CursorID      uuid.UUID
	Limit         int32
}

// ListDeploymentManifests returns one page of manifests.
func (s *FSW) ListDeploymentManifests(ctx context.Context, in ListDeploymentManifestsInput) ([]*models.DeploymentManifest, error) {
	if in.TenantID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id required")
	}
	return s.repo.ListDeploymentManifests(ctx, repository.ListDeploymentManifestsParams{
		TenantID:      in.TenantID,
		SatelliteID:   in.SatelliteID,
		Status:        in.Status,
		CursorCreated: in.CursorCreated,
		CursorID:      in.CursorID,
		Limit:         in.Limit,
	})
}

// UpdateDeploymentManifestStatus transitions a manifest to a new status.
func (s *FSW) UpdateDeploymentManifestStatus(
	ctx context.Context, id uuid.UUID, status models.DeploymentStatus, notes, updatedBy string,
) (*models.DeploymentManifest, error) {
	if status == models.DeploymentStatusUnspecified {
		return nil, errs.New(errs.DomainInvalidArgument, "status required")
	}
	current, err := s.repo.GetDeploymentManifest(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errs.New(errs.DomainNotFound, "deployment manifest %s", id)
		}
		return nil, err
	}
	if !validDeploymentTransition(current.Status, status) {
		return nil, errs.New(errs.DomainPreconditionFailed,
			"illegal deployment status transition: %d -> %d", current.Status, status)
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	return s.repo.UpdateDeploymentManifestStatus(ctx, id, status, notes, updatedBy)
}

func validDeploymentTransition(from, to models.DeploymentStatus) bool {
	switch from {
	case models.DeploymentStatusDraft:
		return to == models.DeploymentStatusApproved
	case models.DeploymentStatusApproved:
		return to == models.DeploymentStatusDeployed || to == models.DeploymentStatusDraft
	case models.DeploymentStatusDeployed:
		return to == models.DeploymentStatusRolledBack
	case models.DeploymentStatusRolledBack:
		return false
	default:
		return false
	}
}

func isHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		case c >= 'A' && c <= 'F':
		default:
			return false
		}
	}
	return true
}
