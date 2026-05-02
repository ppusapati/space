// Package models holds gi-analytics domain types.
package models

import (
	"time"

	"p9e.in/chetana/packages/ulid"
)

// AnalysisType mirrors gianalyticsv1.AnalysisType.
type AnalysisType int32

const (
	TypeUnspecified     AnalysisType = 0
	TypeNDVITimeSeries  AnalysisType = 1
	TypeChangeDetection AnalysisType = 2
	TypeObjectTracking  AnalysisType = 3
	TypeClassification  AnalysisType = 4
	TypeFloodRisk       AnalysisType = 5
	TypeDroughtIndex    AnalysisType = 6
	TypeLandcover       AnalysisType = 7
)

// AnalysisStatus mirrors gianalyticsv1.AnalysisStatus.
type AnalysisStatus int32

const (
	StatusUnspecified AnalysisStatus = 0
	StatusQueued      AnalysisStatus = 1
	StatusRunning     AnalysisStatus = 2
	StatusCompleted   AnalysisStatus = 3
	StatusFailed      AnalysisStatus = 4
	StatusCanceled    AnalysisStatus = 5
)

// AnalysisJob is one analysis job record.
type AnalysisJob struct {
	ID                 ulid.ID
	TenantID           ulid.ID
	Type               AnalysisType
	Status             AnalysisStatus
	InputURIs          []string
	ParametersJSON     string
	OutputURI          string
	ResultsSummaryJSON string
	ErrorMessage       string
	StartedAt          time.Time
	FinishedAt         time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CreatedBy          string
	UpdatedBy          string
}

// Page describes server-side pagination state.
type Page struct {
	TotalCount int32
	PageOffset int32
	PageSize   int32
	HasNext    bool
}
