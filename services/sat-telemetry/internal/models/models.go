// Package models holds sat-telemetry domain types.
package models

import (
	"time"

	"github.com/google/uuid"
)

// ChannelValueType mirrors satv1.ChannelValueType.
type ChannelValueType int

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
	ID           uuid.UUID
	TenantID     uuid.UUID
	SatelliteID  uuid.UUID
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
	ID               uuid.UUID
	TenantID         uuid.UUID
	SatelliteID      uuid.UUID
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
	ID          uuid.UUID
	TenantID    uuid.UUID
	SatelliteID uuid.UUID
	FrameID     uuid.UUID // uuid.Nil when standalone
	ChannelID   uuid.UUID
	SampleTime  time.Time
	ValueDouble float64
	ValueInt    int64
	ValueBool   bool
	ValueText   string
	IngestedAt  time.Time
}
