package services_test

import (
	"context"
	"testing"
	"time"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	"github.com/ppusapati/space/services/sat-command/internal/models"
	"github.com/ppusapati/space/services/sat-command/internal/services"
)

func TestDefineCommandRejectsEmptyTenant(t *testing.T) {
	c := services.New(nil)
	_, err := c.DefineCommand(context.Background(), services.DefineCommandInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestDefineCommandRejectsEmptySubsystemOrName(t *testing.T) {
	c := services.New(nil)
	_, err := c.DefineCommand(context.Background(), services.DefineCommandInput{
		TenantID: ulid.New(),
	})
	if err == nil {
		t.Fatal("expected error for empty subsystem/name")
	}
}

func TestEnqueueUplinkRejectsZeroIDs(t *testing.T) {
	c := services.New(nil)
	_, err := c.EnqueueUplink(context.Background(), services.EnqueueUplinkInput{})
	if err == nil {
		t.Fatal("expected error for zero ids")
	}
}

func TestEnqueueUplinkRejectsZeroScheduledRelease(t *testing.T) {
	c := services.New(nil)
	_, err := c.EnqueueUplink(context.Background(), services.EnqueueUplinkInput{
		TenantID:     ulid.New(),
		SatelliteID:  ulid.New(),
		CommandDefID: ulid.New(),
	})
	if err == nil {
		t.Fatal("expected error for zero scheduled_release")
	}
}

func TestUpdateUplinkStatusRejectsUnspecified(t *testing.T) {
	c := services.New(nil)
	_, err := c.UpdateUplinkStatus(context.Background(), ulid.New(), models.StatusUnspecified, "", "")
	if err == nil {
		t.Fatal("expected error for unspecified status")
	}
}

func TestListsRequireTenant(t *testing.T) {
	c := services.New(nil)
	if _, _, err := c.ListCommandsForTenant(context.Background(), services.ListCommandsInput{}); err == nil {
		t.Fatal("expected error for nil tenant on commands")
	}
	if _, _, err := c.ListUplinksForTenant(context.Background(), services.ListUplinksInput{}); err == nil {
		t.Fatal("expected error for nil tenant on uplinks")
	}
}

func TestEnqueueValidationOrder(t *testing.T) {
	// Provide valid IDs but zero schedule — ensures schedule check fires.
	c := services.New(nil)
	_, err := c.EnqueueUplink(context.Background(), services.EnqueueUplinkInput{
		TenantID:         ulid.New(),
		SatelliteID:      ulid.New(),
		CommandDefID:     ulid.New(),
		ScheduledRelease: time.Time{},
	})
	if err == nil {
		t.Fatal("expected schedule-required error")
	}
}
