// Package models holds eo-analytics domain types.
package models

import (
	"time"

	"github.com/google/uuid"
)

// InferenceTask matches eov1.InferenceTask.
type InferenceTask int

// Task constants.
const (
	TaskUnspecified  InferenceTask = 0
	TaskDetection    InferenceTask = 1
	TaskSegmentation InferenceTask = 2
	TaskClassification InferenceTask = 3
)

// InferenceJobStatus matches eov1.InferenceJobStatus.
type InferenceJobStatus int

// Status constants.
const (
	StatusUnspecified InferenceJobStatus = 0
	StatusPending     InferenceJobStatus = 1
	StatusRunning     InferenceJobStatus = 2
	StatusSucceeded   InferenceJobStatus = 3
	StatusFailed      InferenceJobStatus = 4
)

// Model is a registered ML model.
type Model struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Name         string
	Version      string
	Task         InferenceTask
	Framework    string
	ArtefactURI  string
	MetadataJSON string
	Active       bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    string
	UpdatedBy    string
}

// InferenceJob is one inference request against a model.
type InferenceJob struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	ModelID      uuid.UUID
	ItemID       uuid.UUID
	Status       InferenceJobStatus
	OutputURI    string
	ErrorMessage string
	StartedAt    *time.Time
	FinishedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    string
	UpdatedBy    string
}

// IsTerminal reports whether the status is Succeeded or Failed.
func (j *InferenceJob) IsTerminal() bool {
	return j.Status == StatusSucceeded || j.Status == StatusFailed
}
