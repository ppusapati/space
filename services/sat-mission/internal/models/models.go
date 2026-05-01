// Package models holds sat-mission domain types.
package models

import (
	"time"

	"github.com/google/uuid"
)

// SatelliteMode mirrors satv1.SatelliteMode.
type SatelliteMode int

// Mode constants.
const (
	ModeUnspecified    SatelliteMode = 0
	ModeNadir          SatelliteMode = 1
	ModeSunPointing    SatelliteMode = 2
	ModeTargetTracking SatelliteMode = 3
	ModeInertialHold   SatelliteMode = 4
	ModeSafe           SatelliteMode = 5
)

// OrbitalState is a position+velocity snapshot in ECI at an epoch.
type OrbitalState struct {
	RxKm    float64
	RyKm    float64
	RzKm    float64
	VxKmS   float64
	VyKmS   float64
	VzKmS   float64
	Epoch   time.Time
	Valid   bool
}

// Satellite is the central registry entity.
type Satellite struct {
	ID                       uuid.UUID
	TenantID                 uuid.UUID
	Name                     string
	NORADID                  string
	InternationalDesignator  string
	TLELine1                 string
	TLELine2                 string
	CurrentMode              SatelliteMode
	LastState                OrbitalState
	ConfigJSON               string
	Active                   bool
	CreatedAt                time.Time
	UpdatedAt                time.Time
	CreatedBy                string
	UpdatedBy                string
}
