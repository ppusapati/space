// Package models holds gs-ingest domain types.
package models

import (
	"time"

	"p9e.in/chetana/packages/ulid"
)

// IngestStatus mirrors gsingestv1.IngestStatus.
type IngestStatus int32

const (
	StatusUnspecified IngestStatus = 0
	StatusQueued      IngestStatus = 1
	StatusActive      IngestStatus = 2
	StatusCompleted   IngestStatus = 3
	StatusFailed      IngestStatus = 4
	StatusCanceled    IngestStatus = 5
)

// IngestSession is one ingest session, typically scoped to a contact pass.
type IngestSession struct {
	ID             ulid.ID
	TenantID       ulid.ID
	BookingID      ulid.ID
	PassID         ulid.ID
	StationID      ulid.ID
	SatelliteID    ulid.ID
	Status         IngestStatus
	StartedAt      time.Time
	CompletedAt    time.Time
	FramesReceived uint64
	BytesReceived  uint64
	ErrorMessage   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CreatedBy      string
	UpdatedBy      string
}

// DownlinkFrame is one raw frame received during a session.
type DownlinkFrame struct {
	ID               ulid.ID
	TenantID         ulid.ID
	SessionID        ulid.ID
	APID             uint32
	VirtualChannel   uint32
	SequenceCount    uint64
	GroundTime       time.Time
	PayloadSizeBytes uint64
	PayloadSHA256    string
	PayloadURI       string
	FrameType        string
	CreatedAt        time.Time
	CreatedBy        string
}

// Page describes server-side pagination state.
type Page struct {
	TotalCount int32
	PageOffset int32
	PageSize   int32
	HasNext    bool
}
