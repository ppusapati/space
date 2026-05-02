// Package models holds eo-analytics domain types.
package models

import (
	"time"

	"p9e.in/chetana/packages/ulid"
)

// InferenceTask mirrors eoanalyticsv1.InferenceTask.
type InferenceTask int32

// Task constants.
const (
	TaskUnspecified    InferenceTask = 0
	TaskDetection      InferenceTask = 1
	TaskSegmentation   InferenceTask = 2
	TaskClassification InferenceTask = 3
)

// InferenceJobStatus mirrors eoanalyticsv1.InferenceJobStatus.
type InferenceJobStatus int32

// Status constants.
const (
	StatusUnspecified InferenceJobStatus = 0
	StatusQueued      InferenceJobStatus = 1
	StatusRunning     InferenceJobStatus = 2
	StatusSucceeded   InferenceJobStatus = 3
	StatusFailed      InferenceJobStatus = 4
)

// Model is a registered ML model.
type Model struct {
	ID           ulid.ID
	TenantID     ulid.ID
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

// InferenceJob is a single inference task.
type InferenceJob struct {
	ID           ulid.ID
	TenantID     ulid.ID
	ModelID      ulid.ID
	ItemID       ulid.ID
	Status       InferenceJobStatus
	OutputURI    string
	ErrorMessage string
	StartedAt    time.Time
	FinishedAt   time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    string
	UpdatedBy    string
}

// Page describes server-side pagination state.
type Page struct {
	TotalCount int32
	PageOffset int32
	PageSize   int32
	HasNext    bool
}
