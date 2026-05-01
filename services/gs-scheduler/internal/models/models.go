// Package models holds gs-scheduler domain types.
package models

import (
	"time"

	"p9e.in/samavaya/packages/ulid"
)

// BookingStatus mirrors gsschedulerv1.BookingStatus.
type BookingStatus int32

const (
	StatusUnspecified BookingStatus = 0
	StatusRequested   BookingStatus = 1
	StatusApproved    BookingStatus = 2
	StatusScheduled   BookingStatus = 3
	StatusActive      BookingStatus = 4
	StatusCompleted   BookingStatus = 5
	StatusFailed      BookingStatus = 6
	StatusCanceled    BookingStatus = 7
)

// ContactPass is a predicted satellite-station contact opportunity.
type ContactPass struct {
	ID               ulid.ID
	TenantID         ulid.ID
	StationID        ulid.ID
	SatelliteID      ulid.ID
	AOSTime          time.Time
	TCATime          time.Time
	LOSTime          time.Time
	MaxElevationDeg  float64
	AOSAzimuthDeg    float64
	LOSAzimuthDeg    float64
	Source           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CreatedBy        string
	UpdatedBy        string
}

// Booking is a tenant's reservation of a contact pass.
type Booking struct {
	ID           ulid.ID
	TenantID     ulid.ID
	PassID       ulid.ID
	Priority     int32
	Status       BookingStatus
	Purpose      string
	Notes        string
	ScheduledAt  time.Time
	CompletedAt  time.Time
	ErrorMessage string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    string
	UpdatedBy    string
}

// Page describes server-side pagination state.
type Page struct {
	TotalCount int32
	PageOffset int32
	PageSize   int32
	HasNext    bool
}
