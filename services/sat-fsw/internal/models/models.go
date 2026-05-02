// Package models holds sat-fsw domain types.
package models

import (
	"time"

	"p9e.in/chetana/packages/ulid"
)

// FirmwareBuildStatus mirrors satfswv1.FirmwareBuildStatus.
type FirmwareBuildStatus int32

// Build status constants.
const (
	BuildUnspecified FirmwareBuildStatus = 0
	BuildBuilding    FirmwareBuildStatus = 1
	BuildReady       FirmwareBuildStatus = 2
	BuildRejected    FirmwareBuildStatus = 3
	BuildDeprecated  FirmwareBuildStatus = 4
)

// DeploymentStatus mirrors satfswv1.DeploymentStatus.
type DeploymentStatus int32

// Deployment status constants.
const (
	DeploymentUnspecified DeploymentStatus = 0
	DeploymentDraft       DeploymentStatus = 1
	DeploymentApproved    DeploymentStatus = 2
	DeploymentDeployed    DeploymentStatus = 3
	DeploymentRolledBack  DeploymentStatus = 4
)

// FirmwareBuild is a single firmware artefact.
type FirmwareBuild struct {
	ID                ulid.ID
	TenantID          ulid.ID
	TargetPlatform    string
	Subsystem         string
	Version           string
	GitSHA            string
	ArtefactURI       string
	ArtefactSizeBytes uint64
	ArtefactSHA256    string
	Status            FirmwareBuildStatus
	Notes             string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CreatedBy         string
	UpdatedBy         string
}

// DeploymentManifest is one assignment of firmware builds onto a satellite.
type DeploymentManifest struct {
	ID              ulid.ID
	TenantID        ulid.ID
	SatelliteID     ulid.ID
	ManifestVersion string
	Status          DeploymentStatus
	Assignments     map[string]string
	Notes           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedBy       string
	UpdatedBy       string
}

// Page describes server-side pagination state.
type Page struct {
	TotalCount int32
	PageOffset int32
	PageSize   int32
	HasNext    bool
}
