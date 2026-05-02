// Package models holds sat-telemetry domain types.
package models

import (
	"time"

	"p9e.in/chetana/packages/ulid"
)

// ChannelValueType mirrors sattelemetryv1.ChannelValueType.
type ChannelValueType int32

// Value type constants.
const (
	ValueTypeUnspecified ChannelValueType = 0
	ValueTypeFloat       ChannelValueType = 1
	ValueTypeInt         ChannelValueType = 2
	ValueTypeBool        ChannelValueType = 3
	ValueTypeEnum        ChannelValueType = 4
)

// Channel is a named, typed measurement point on a satellite subsystem.
type Channel struct {
	ID           ulid.ID
	TenantID     ulid.ID
	SatelliteID  ulid.ID
	Subsystem    string
	Name         string
	Units        string
	ValueType    ChannelValueType
	MinValue     float64
	MaxValue     float64
	SampleRateHz float64
	Active       bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    string
	UpdatedBy    string
}

// Frame is a single CCSDS-style packet recorded at the ground.
type Frame struct {
	ID               ulid.ID
	TenantID         ulid.ID
	SatelliteID     ulid.ID
	APID             uint32
	VirtualChannel   uint32
	SequenceCount    uint64
	SatTime          time.Time
	GroundTime       time.Time
	PayloadSizeBytes uint64
	PayloadSHA256    string
	FrameType        string
	CreatedBy        string
}

// Sample is a single measurement of a channel at a point in time.
type Sample struct {
	ID          ulid.ID
	TenantID    ulid.ID
	SatelliteID ulid.ID
	FrameID     ulid.ID // ulid.Zero when standalone
	ChannelID   ulid.ID
	SampleTime  time.Time
	ValueDouble float64
	ValueInt    int64
	ValueBool   bool
	ValueText   string
	IngestedAt  time.Time
}

// Page describes server-side pagination state.
type Page struct {
	TotalCount int32
	PageOffset int32
	PageSize   int32
	HasNext    bool
}
