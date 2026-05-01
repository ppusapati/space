// Package services holds gs-scheduler business logic.
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	"github.com/ppusapati/space/services/gs-scheduler/internal/models"
	"github.com/ppusapati/space/services/gs-scheduler/internal/repository"
)

// Scheduler is the gs-scheduler service-layer facade.
type Scheduler struct {
	repo  *repository.Repo
	IDFn  func() ulid.ID
	NowFn func() time.Time
}

func New(repo *repository.Repo) *Scheduler {
	return &Scheduler{
		repo:  repo,
		IDFn:  ulid.NewMonotonic,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// ----- ContactPass --------------------------------------------------------

type InsertContactPassInput struct {
	TenantID        ulid.ID
	StationID       ulid.ID
	SatelliteID     ulid.ID
	AOSTime         time.Time
	TCATime         time.Time
	LOSTime         time.Time
	MaxElevationDeg float64
	AOSAzimuthDeg   float64
	LOSAzimuthDeg   float64
	Source          string
	CreatedBy       string
}

func (s *Scheduler) InsertContactPass(ctx context.Context, in InsertContactPassInput) (*models.ContactPass, error) {
	if in.TenantID.IsZero() || in.StationID.IsZero() || in.SatelliteID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT",
			"tenant_id, station_id, satellite_id required")
	}
	if in.AOSTime.IsZero() || in.LOSTime.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "aos_time and los_time required")
	}
	if !in.LOSTime.After(in.AOSTime) {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "los_time must be > aos_time")
	}
	if in.TCATime.IsZero() {
		// allow TCA inferred as midpoint when not supplied
		in.TCATime = in.AOSTime.Add(in.LOSTime.Sub(in.AOSTime) / 2)
	} else if in.TCATime.Before(in.AOSTime) || in.TCATime.After(in.LOSTime) {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tca_time must be within [aos, los]")
	}
	in.Source = strings.TrimSpace(in.Source)
	if in.Source == "" {
		in.Source = "predicted"
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return s.repo.InsertContactPass(ctx, repository.InsertContactPassParams{
		ID:              s.IDFn(),
		TenantID:        in.TenantID,
		StationID:       in.StationID,
		SatelliteID:     in.SatelliteID,
		AOSTime:         in.AOSTime,
		TCATime:         in.TCATime,
		LOSTime:         in.LOSTime,
		MaxElevationDeg: in.MaxElevationDeg,
		AOSAzimuthDeg:   in.AOSAzimuthDeg,
		LOSAzimuthDeg:   in.LOSAzimuthDeg,
		Source:          in.Source,
		CreatedBy:       createdBy,
	})
}

func (s *Scheduler) GetContactPass(ctx context.Context, id ulid.ID) (*models.ContactPass, error) {
	p, err := s.repo.GetContactPass(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("PASS_NOT_FOUND", "contact_pass "+id.String())
	}
	return p, err
}

type ListContactPassesInput struct {
	TenantID        ulid.ID
	StationID       *ulid.ID
	SatelliteID     *ulid.ID
	AOSStart        time.Time
	AOSEnd          time.Time
	MinElevationDeg *float64
	PageOffset      int32
	PageSize        int32
}

func (s *Scheduler) ListContactPassesForTenant(ctx context.Context, in ListContactPassesInput) ([]*models.ContactPass, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	if !in.AOSStart.IsZero() && !in.AOSEnd.IsZero() && in.AOSEnd.Before(in.AOSStart) {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "aos_end must be >= aos_start")
	}
	rows, total, err := s.repo.ListContactPassesForTenant(ctx, repository.ListContactPassesParams{
		TenantID:        in.TenantID,
		StationID:       in.StationID,
		SatelliteID:     in.SatelliteID,
		AOSStart:        in.AOSStart,
		AOSEnd:          in.AOSEnd,
		MinElevationDeg: in.MinElevationDeg,
		PageOffset:      in.PageOffset,
		PageSize:        in.PageSize,
	})
	if err != nil {
		return nil, models.Page{}, err
	}
	return rows, models.Page{
		TotalCount: total,
		PageOffset: in.PageOffset,
		PageSize:   in.PageSize,
		HasNext:    in.PageOffset+int32(len(rows)) < total,
	}, nil
}

// ----- Booking ------------------------------------------------------------

type RequestBookingInput struct {
	TenantID  ulid.ID
	PassID    ulid.ID
	Priority  int32
	Purpose   string
	Notes     string
	CreatedBy string
}

func (s *Scheduler) RequestBooking(ctx context.Context, in RequestBookingInput) (*models.Booking, error) {
	if in.TenantID.IsZero() || in.PassID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id and pass_id required")
	}
	if in.Priority < 0 || in.Priority > 100 {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "priority must be in [0, 100]")
	}
	in.Purpose = strings.TrimSpace(in.Purpose)
	if in.Purpose == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "purpose required")
	}
	// Verify pass exists and belongs to same tenant.
	pass, err := s.repo.GetContactPass(ctx, in.PassID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, pkgerrors.BadRequest("PASS_NOT_FOUND",
				"contact_pass "+in.PassID.String())
		}
		return nil, err
	}
	if pass.TenantID != in.TenantID {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "pass tenant mismatch")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return s.repo.RequestBooking(ctx, repository.RequestBookingParams{
		ID:        s.IDFn(),
		TenantID:  in.TenantID,
		PassID:    in.PassID,
		Priority:  in.Priority,
		Status:    models.StatusRequested,
		Purpose:   in.Purpose,
		Notes:     in.Notes,
		CreatedBy: createdBy,
	})
}

func (s *Scheduler) GetBooking(ctx context.Context, id ulid.ID) (*models.Booking, error) {
	b, err := s.repo.GetBooking(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("BOOKING_NOT_FOUND", "booking "+id.String())
	}
	return b, err
}

type ListBookingsInput struct {
	TenantID   ulid.ID
	Status     *models.BookingStatus
	PageOffset int32
	PageSize   int32
}

func (s *Scheduler) ListBookingsForTenant(ctx context.Context, in ListBookingsInput) ([]*models.Booking, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := s.repo.ListBookingsForTenant(ctx, repository.ListBookingsParams{
		TenantID:   in.TenantID,
		Status:     in.Status,
		PageOffset: in.PageOffset,
		PageSize:   in.PageSize,
	})
	if err != nil {
		return nil, models.Page{}, err
	}
	return rows, models.Page{
		TotalCount: total,
		PageOffset: in.PageOffset,
		PageSize:   in.PageSize,
		HasNext:    in.PageOffset+int32(len(rows)) < total,
	}, nil
}

// CancelBooking marks a booking CANCELED.
func (s *Scheduler) CancelBooking(ctx context.Context, id ulid.ID, reason, updatedBy string) (*models.Booking, error) {
	msg := strings.TrimSpace(reason)
	if msg == "" {
		msg = "canceled by user"
	}
	return s.UpdateBookingStatus(ctx, id, models.StatusCanceled, msg, updatedBy)
}

// UpdateBookingStatus enforces the booking transition graph:
//
//	REQUESTED -> APPROVED | CANCELED
//	APPROVED  -> SCHEDULED | CANCELED
//	SCHEDULED -> ACTIVE | CANCELED
//	ACTIVE    -> COMPLETED | FAILED | CANCELED
//	COMPLETED, FAILED, CANCELED — terminal.
func (s *Scheduler) UpdateBookingStatus(
	ctx context.Context, id ulid.ID, status models.BookingStatus, errorMessage, updatedBy string,
) (*models.Booking, error) {
	if status == models.StatusUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "status required")
	}
	current, err := s.GetBooking(ctx, id)
	if err != nil {
		return nil, err
	}
	if !validBookingTransition(current.Status, status) {
		return nil, pkgerrors.New(412, "ILLEGAL_TRANSITION",
			"illegal booking status transition")
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	updated, err := s.repo.UpdateBookingStatus(ctx, id, status, errorMessage, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("BOOKING_NOT_FOUND", "booking "+id.String())
	}
	return updated, err
}

func validBookingTransition(from, to models.BookingStatus) bool {
	switch from {
	case models.StatusRequested:
		return to == models.StatusApproved || to == models.StatusCanceled
	case models.StatusApproved:
		return to == models.StatusScheduled || to == models.StatusCanceled
	case models.StatusScheduled:
		return to == models.StatusActive || to == models.StatusCanceled
	case models.StatusActive:
		return to == models.StatusCompleted || to == models.StatusFailed || to == models.StatusCanceled
	case models.StatusCompleted, models.StatusFailed, models.StatusCanceled:
		return false
	default:
		return false
	}
}
