// Package models holds eo-pipeline domain types.
package models

import (
	"time"

	"p9e.in/samavaya/packages/ulid"
)

// JobStage mirrors eopipelinev1.JobStage.
type JobStage int32

// Stage constants.
const (
	StageUnspecified     JobStage = 0
	StageRadiometric     JobStage = 1
	StageGeometric       JobStage = 2
	StageAtmospheric     JobStage = 3
	StagePanSharpen      JobStage = 4
	StageMosaic          JobStage = 5
	StageSARSpeckle      JobStage = 6
	StageSARTerrain      JobStage = 7
	StageSARPolarimetric JobStage = 8
)

// JobStatus mirrors eopipelinev1.JobStatus.
type JobStatus int32

// Status constants.
const (
	StatusUnspecified JobStatus = 0
	StatusPending     JobStatus = 1
	StatusRunning     JobStatus = 2
	StatusSucceeded   JobStatus = 3
	StatusFailed      JobStatus = 4
	StatusCancelled   JobStatus = 5
)

// Job is a single processing job.
type Job struct {
	ID             ulid.ID
	TenantID       ulid.ID
	ItemID         ulid.ID
	Stage          JobStage
	Status         JobStatus
	ParametersJSON string
	OutputURI      string
	ErrorMessage   string
	StartedAt      time.Time
	FinishedAt     time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CreatedBy      string
	UpdatedBy      string
}

// Page describes server-side pagination state.
type Page struct {
	TotalCount int32
	PageOffset int32
	PageSize   int32
	HasNext    bool
}
