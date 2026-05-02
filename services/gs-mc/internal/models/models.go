// Package models holds gs-mc domain types.
package models

import (
	"time"

	"p9e.in/chetana/packages/ulid"
)

// FrequencyBand mirrors gsmcv1.FrequencyBand.
type FrequencyBand int32

const (
	BandUnspecified FrequencyBand = 0
	BandVHF         FrequencyBand = 1
	BandUHF         FrequencyBand = 2
	BandL           FrequencyBand = 3
	BandS           FrequencyBand = 4
	BandC           FrequencyBand = 5
	BandX           FrequencyBand = 6
	BandKu          FrequencyBand = 7
	BandK           FrequencyBand = 8
	BandKa          FrequencyBand = 9
)

// Polarization mirrors gsmcv1.Polarization.
type Polarization int32

const (
	PolUnspecified Polarization = 0
	PolRHCP        Polarization = 1
	PolLHCP        Polarization = 2
	PolLinearH     Polarization = 3
	PolLinearV     Polarization = 4
	PolDual        Polarization = 5
)

// GroundStation is a tenant-scoped ground-station definition.
type GroundStation struct {
	ID            ulid.ID
	TenantID      ulid.ID
	Slug          string
	Name          string
	CountryCode   string
	LatitudeDeg   float64
	LongitudeDeg  float64
	AltitudeM     float64
	Active        bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CreatedBy     string
	UpdatedBy     string
}

// Antenna is one antenna on a ground station.
type Antenna struct {
	ID                ulid.ID
	TenantID          ulid.ID
	StationID         ulid.ID
	Slug              string
	Name              string
	Band              FrequencyBand
	MinFreqHz         uint64
	MaxFreqHz         uint64
	Polarization      Polarization
	GainDBI           float64
	SlewRateDegPerS   float64
	Active            bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CreatedBy         string
	UpdatedBy         string
}

// Page describes server-side pagination state.
type Page struct {
	TotalCount int32
	PageOffset int32
	PageSize   int32
	HasNext    bool
}
