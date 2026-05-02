// Package models holds gi-predict domain types.
package models

import (
	"time"

	"p9e.in/chetana/packages/ulid"
)

// ForecastType mirrors gipredictv1.ForecastType.
type ForecastType int32

const (
	TypeUnspecified      ForecastType = 0
	TypeNDVITrend        ForecastType = 1
	TypeFloodRisk        ForecastType = 2
	TypeDroughtIndex     ForecastType = 3
	TypeLandcoverChange  ForecastType = 4
	TypeWeather          ForecastType = 5
	TypeCustom           ForecastType = 6
)

// ForecastStatus mirrors gipredictv1.ForecastStatus.
type ForecastStatus int32

const (
	StatusUnspecified ForecastStatus = 0
	StatusQueued      ForecastStatus = 1
	StatusRunning     ForecastStatus = 2
	StatusCompleted   ForecastStatus = 3
	StatusFailed      ForecastStatus = 4
	StatusCanceled    ForecastStatus = 5
)

// ForecastJob is one forecast job record.
type ForecastJob struct {
	ID                 ulid.ID
	TenantID           ulid.ID
	Type               ForecastType
	Status             ForecastStatus
	ModelID            ulid.ID
	InputURIs          []string
	HorizonDays        int32
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
