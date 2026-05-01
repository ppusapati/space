// Package models holds gi-fusion domain types.
package models

import (
	"time"

	"p9e.in/samavaya/packages/ulid"
)

// FusionMethod mirrors gifusionv1.FusionMethod.
type FusionMethod int32

const (
	MethodUnspecified FusionMethod = 0
	MethodPanSharpen  FusionMethod = 1
	MethodTimeSeries  FusionMethod = 2
	MethodMultimodal  FusionMethod = 3
	MethodVector      FusionMethod = 4
	MethodCustom      FusionMethod = 5
)

// FusionStatus mirrors gifusionv1.FusionStatus.
type FusionStatus int32

const (
	StatusUnspecified FusionStatus = 0
	StatusQueued      FusionStatus = 1
	StatusRunning     FusionStatus = 2
	StatusCompleted   FusionStatus = 3
	StatusFailed      FusionStatus = 4
	StatusCanceled    FusionStatus = 5
)

// FusionJob is a fusion-job record.
type FusionJob struct {
	ID             ulid.ID
	TenantID       ulid.ID
	Method         FusionMethod
	Status         FusionStatus
	InputURIs      []string
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
