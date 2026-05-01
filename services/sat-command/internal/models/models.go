// Package models holds sat-command domain types.
package models

import (
	"time"

	"p9e.in/samavaya/packages/ulid"
)

// UplinkStatus mirrors satcommandv1.UplinkStatus.
type UplinkStatus int32

// Status constants.
const (
	StatusUnspecified UplinkStatus = 0
	StatusQueued      UplinkStatus = 1
	StatusReleased    UplinkStatus = 2
	StatusAcked       UplinkStatus = 3
	StatusExecuted    UplinkStatus = 4
	StatusFailed      UplinkStatus = 5
	StatusCanceled    UplinkStatus = 6
)

// CommandDef is the definition of a satellite command.
type CommandDef struct {
	ID               ulid.ID
	TenantID         ulid.ID
	SatelliteID      ulid.ID // ulid.Zero = tenant-wide
	Subsystem        string
	Name             string
	Opcode           uint32
	ParametersSchema string
	Description      string
	Active           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CreatedBy        string
	UpdatedBy        string
}

// UplinkRequest is one scheduled release of a command.
type UplinkRequest struct {
	ID               ulid.ID
	TenantID         ulid.ID
	SatelliteID      ulid.ID
	CommandDefID     ulid.ID
	ParametersJSON   string
	ScheduledRelease time.Time
	Status           UplinkStatus
	SequenceNumber   uint64
	GatewayID        string
	SubmittedAt      time.Time
	ReleasedAt       time.Time
	AckedAt          time.Time
	CompletedAt      time.Time
	ErrorMessage     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CreatedBy        string
	UpdatedBy        string
}

// Page describes server-side pagination state.
type Page struct {
	TotalCount int32
	PageOffset int32
	PageSize   int32
	HasNext    bool
}
