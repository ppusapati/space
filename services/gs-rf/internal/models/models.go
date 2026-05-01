// Package models holds gs-rf domain types.
package models

import (
	"time"

	"p9e.in/samavaya/packages/ulid"
)

// LinkBudget is a pre-pass RF calculation.
type LinkBudget struct {
	ID                 ulid.ID
	TenantID           ulid.ID
	PassID             ulid.ID
	StationID          ulid.ID
	AntennaID          ulid.ID
	SatelliteID        ulid.ID
	CarrierFreqHz      uint64
	TxPowerDBM         float64
	TxGainDBI          float64
	RxGainDBI          float64
	RxNoiseTempK       float64
	BandwidthHz        float64
	SlantRangeKm       float64
	FreeSpaceLossDB    float64
	AtmosphericLossDB  float64
	PolarizationLossDB float64
	PointingLossDB     float64
	PredictedEbN0DB    float64
	PredictedSNRDB     float64
	LinkMarginDB       float64
	Notes              string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CreatedBy          string
	UpdatedBy          string
}

// LinkMeasurement is one measured RF sample during a pass.
type LinkMeasurement struct {
	ID              ulid.ID
	TenantID        ulid.ID
	PassID          ulid.ID
	StationID       ulid.ID
	AntennaID       ulid.ID
	SampledAt       time.Time
	RSSIDBM         float64
	SNRDB           float64
	BER             float64
	FER             float64
	FrequencyHz     uint64
	DopplerShiftHz  float64
	CreatedAt       time.Time
	CreatedBy       string
}

// Page describes server-side pagination state.
type Page struct {
	TotalCount int32
	PageOffset int32
	PageSize   int32
	HasNext    bool
}
