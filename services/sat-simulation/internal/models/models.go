// Package models holds sat-simulation domain types.
package models

import (
	"time"

	"p9e.in/samavaya/packages/ulid"
)

// SimulationMode mirrors satsimulationv1.SimulationMode.
type SimulationMode int32

// Mode constants.
const (
	ModeUnspecified SimulationMode = 0
	ModeSITL        SimulationMode = 1
	ModeHITL        SimulationMode = 2
)

// RunStatus mirrors satsimulationv1.RunStatus.
type RunStatus int32

// Run status constants.
const (
	StatusUnspecified RunStatus = 0
	StatusQueued      RunStatus = 1
	StatusRunning     RunStatus = 2
	StatusCompleted   RunStatus = 3
	StatusFailed      RunStatus = 4
	StatusCanceled    RunStatus = 5
)

// Scenario is a reusable test specification.
type Scenario struct {
	ID          ulid.ID
	TenantID    ulid.ID
	Slug        string
	Title       string
	Description string
	SpecJSON    string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   string
	UpdatedBy   string
}

// SimulationRun is one execution of a scenario against a satellite.
type SimulationRun struct {
	ID             ulid.ID
	TenantID       ulid.ID
	SatelliteID    ulid.ID
	ScenarioID     ulid.ID
	Mode           SimulationMode
	Status         RunStatus
	ParametersJSON string
	LogURI         string
	TelemetryURI   string
	ResultsJSON    string
	Score          float64
	StartedAt      time.Time
	FinishedAt     time.Time
	ErrorMessage   string
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
