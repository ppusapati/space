package services_test

import (
	"context"
	"testing"
	"time"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	"github.com/ppusapati/space/services/gs-scheduler/internal/models"
	"github.com/ppusapati/space/services/gs-scheduler/internal/services"
)

func TestInsertContactPassRejectsZeroIDs(t *testing.T) {
	s := services.New(nil)
	_, err := s.InsertContactPass(context.Background(), services.InsertContactPassInput{})
	if err == nil {
		t.Fatal("expected error for zero ids")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestInsertContactPassRejectsBadAOSLOSOrder(t *testing.T) {
	s := services.New(nil)
	t1 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	t0 := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
	_, err := s.InsertContactPass(context.Background(), services.InsertContactPassInput{
		TenantID: ulid.New(), StationID: ulid.New(), SatelliteID: ulid.New(),
		AOSTime: t1, LOSTime: t0,
	})
	if err == nil {
		t.Fatal("expected error for los_time <= aos_time")
	}
}

func TestRequestBookingRejectsZeroIDs(t *testing.T) {
	s := services.New(nil)
	_, err := s.RequestBooking(context.Background(), services.RequestBookingInput{})
	if err == nil {
		t.Fatal("expected error for zero ids")
	}
}

func TestRequestBookingRejectsBadPriority(t *testing.T) {
	s := services.New(nil)
	_, err := s.RequestBooking(context.Background(), services.RequestBookingInput{
		TenantID: ulid.New(),
		PassID:   ulid.New(),
		Priority: 200,
		Purpose:  "downlink",
	})
	if err == nil {
		t.Fatal("expected error for priority > 100")
	}
}

func TestUpdateBookingStatusRejectsUnspecified(t *testing.T) {
	s := services.New(nil)
	_, err := s.UpdateBookingStatus(context.Background(), ulid.New(), models.StatusUnspecified, "", "")
	if err == nil {
		t.Fatal("expected error for unspecified status")
	}
}

func TestListsRequireTenant(t *testing.T) {
	s := services.New(nil)
	if _, _, err := s.ListContactPassesForTenant(context.Background(), services.ListContactPassesInput{}); err == nil {
		t.Fatal("expected error for nil tenant on passes")
	}
	if _, _, err := s.ListBookingsForTenant(context.Background(), services.ListBookingsInput{}); err == nil {
		t.Fatal("expected error for nil tenant on bookings")
	}
}
