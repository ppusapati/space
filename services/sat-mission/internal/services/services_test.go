package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	"github.com/ppusapati/space/services/sat-mission/internal/models"
	"github.com/ppusapati/space/services/sat-mission/internal/services"
)

func TestRegisterSatelliteRejectsEmptyTenant(t *testing.T) {
	m := services.New(nil)
	_, err := m.RegisterSatellite(context.Background(), services.RegisterSatelliteInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestRegisterSatelliteRejectsEmptyName(t *testing.T) {
	m := services.New(nil)
	_, err := m.RegisterSatellite(context.Background(), services.RegisterSatelliteInput{
		TenantID: ulid.New(),
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestUpdateTLERejectsWrongLength(t *testing.T) {
	m := services.New(nil)
	_, err := m.UpdateTLE(context.Background(), ulid.New(), "short", "also short", "user")
	if err == nil {
		t.Fatal("expected error for short TLE lines")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestUpdateTLERejectsBadLineNumber(t *testing.T) {
	// build two 69-char lines that start with wrong prefix
	bad1 := strings.Repeat("X", 69)
	bad2 := strings.Repeat("X", 69)
	m := services.New(nil)
	_, err := m.UpdateTLE(context.Background(), ulid.New(), bad1, bad2, "user")
	if err == nil {
		t.Fatal("expected error for wrong line prefixes")
	}
}

func TestUpdateOrbitalStateRejectsInvalid(t *testing.T) {
	m := services.New(nil)
	_, err := m.UpdateOrbitalState(context.Background(), ulid.New(), models.OrbitalState{}, "user")
	if err == nil {
		t.Fatal("expected error for invalid orbital state")
	}
}

func TestSetModeRejectsUnspecified(t *testing.T) {
	m := services.New(nil)
	_, err := m.SetMode(context.Background(), ulid.New(), models.ModeUnspecified, "user")
	if err == nil {
		t.Fatal("expected error for unspecified mode")
	}
}

func TestListSatellitesRequiresTenant(t *testing.T) {
	m := services.New(nil)
	_, _, err := m.ListSatellitesForTenant(context.Background(), ulid.Zero, 0, 10)
	if err == nil {
		t.Fatal("expected error for zero tenant")
	}
}

func TestSetModeRejectsZeroID(t *testing.T) {
	m := services.New(nil)
	_, err := m.SetMode(context.Background(), ulid.Zero, models.ModeNadir, "user")
	if err == nil {
		t.Fatal("expected error for zero id")
	}
}

func TestUpdateOrbitalStateRequiresEpoch(t *testing.T) {
	m := services.New(nil)
	state := models.OrbitalState{Valid: true} // missing epoch
	_, err := m.UpdateOrbitalState(context.Background(), ulid.New(), state, "user")
	if err == nil {
		t.Fatal("expected error when epoch is zero")
	}
}

func TestErrorTypePassesThrough(t *testing.T) {
	m := services.New(nil)
	_, err := m.RegisterSatellite(context.Background(), services.RegisterSatelliteInput{})
	var pe *pkgerrors.Error
	if !errors.As(err, &pe) {
		t.Fatalf("expected packages/errors.Error, got %T %v", err, err)
	}
	_ = time.Now() // keep import in case we expand
}
