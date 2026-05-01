// Package models holds the eo-pipeline domain types.
package models

import (
	"time"

	"github.com/google/uuid"
)

// JobStage matches eov1.JobStage.
type JobStage int

// Stage constants — must match the proto enum.
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

// JobStatus matches eov1.JobStatus.
type JobStatus int

// Status constants.
const (
	StatusUnspecified JobStatus = 0
	StatusPending     JobStatus = 1
	StatusRunning     JobStatus = 2
	StatusSucceeded   JobStatus = 3
	StatusFailed      JobStatus = 4
	StatusCancelled   JobStatus = 5
)

// Job is one processing-job record.
type Job struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	ItemID         uuid.UUID
	Stage          JobStage
	Status         JobStatus
	ParametersJSON string
	OutputURI      string
	ErrorMessage   string
	StartedAt      *time.Time
	FinishedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CreatedBy      string
	UpdatedBy      string
}

// IsTerminal reports whether the status is one of Succeeded / Failed /
// Cancelled.
func (j *Job) IsTerminal() bool {
	return j.Status == StatusSucceeded || j.Status == StatusFailed || j.Status == StatusCancelled
}
