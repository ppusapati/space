// Package models holds sat-fsw domain types.
package models

import (
	"time"

	"github.com/google/uuid"
)

// FirmwareBuildStatus mirrors satv1.FirmwareBuildStatus.
type FirmwareBuildStatus int

// Build status constants.
const (
	BuildStatusUnspecified FirmwareBuildStatus = 0
	BuildStatusBuilding    FirmwareBuildStatus = 1
	BuildStatusReady       FirmwareBuildStatus = 2
	BuildStatusRejected    FirmwareBuildStatus = 3
	BuildStatusDeprecated  FirmwareBuildStatus = 4
)

// DeploymentStatus mirrors satv1.DeploymentStatus.
type DeploymentStatus int

// Deployment status constants.
const (
	DeploymentStatusUnspecified DeploymentStatus = 0
	DeploymentStatusDraft       DeploymentStatus = 1
	DeploymentStatusApproved    DeploymentStatus = 2
	DeploymentStatusDeployed    DeploymentStatus = 3
	DeploymentStatusRolledBack  DeploymentStatus = 4
)

// FirmwareBuild is a built firmware artefact for a flight subsystem.
type FirmwareBuild struct {
	ID                uuid.UUID
	TenantID          uuid.UUID
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

// DeploymentManifest pins a firmware build per subsystem on a satellite.
type DeploymentManifest struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	SatelliteID     uuid.UUID
	ManifestVersion string
	Status          DeploymentStatus
	Assignments     map[string]string
	Notes           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedBy       string
	UpdatedBy       string
}
