// Package models holds sat-mission domain types.
package models

import (
	"time"

	"p9e.in/samavaya/packages/ulid"
)

// SatelliteMode mirrors satmissionv1.SatelliteMode.
type SatelliteMode int32

// Mode constants.
const (
	ModeUnspecified     SatelliteMode = 0
	ModeNadir           SatelliteMode = 1
	ModeSunPointing     SatelliteMode = 2
	ModeTargetTracking  SatelliteMode = 3
	ModeInertialHold    SatelliteMode = 4
	ModeSafe            SatelliteMode = 5
)

// OrbitalState is the ECI position+velocity at an epoch.
type OrbitalState struct {
	RxKm   float64
	RyKm   float64
	RzKm   float64
	VxKmS  float64
	VyKmS  float64
	VzKmS  float64
	Epoch  time.Time
	Valid  bool
}

// Satellite is one tenant-owned satellite.
type Satellite struct {
	ID                      ulid.ID
	TenantID                ulid.ID
	Name                    string
	NoradID                 string
	InternationalDesignator string
	TLELine1                string
	TLELine2                string
	CurrentMode             SatelliteMode
	LastState               OrbitalState
	ConfigJSON              string
	Active                  bool
	CreatedAt               time.Time
	UpdatedAt               time.Time
	CreatedBy               string
	UpdatedBy               string
}

// Page describes server-side pagination state.
type Page struct {
	TotalCount int32
	PageOffset int32
	PageSize   int32
	HasNext    bool
}
