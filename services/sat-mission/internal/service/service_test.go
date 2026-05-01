package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ppusapati/space/pkg/errs"
	"github.com/ppusapati/space/services/sat-mission/internal/models"
	"github.com/ppusapati/space/services/sat-mission/internal/service"
)

func TestRegisterSatelliteRequiresTenantAndName(t *testing.T) {
	m := service.New(nil)
	_, err := m.RegisterSatellite(context.Background(), service.RegisterSatelliteInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	var de *errs.E
	if !errors.As(err, &de) || de.Domain != errs.DomainInvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}

	_, err = m.RegisterSatellite(context.Background(), service.RegisterSatelliteInput{
		TenantID: uuid.New(),
	})
	if err == nil || !errors.As(err, &de) || de.Domain != errs.DomainInvalidArgument {
		t.Fatalf("expected InvalidArgument for missing name, got %v", err)
	}
}

func TestUpdateTLERejectsBadLineLength(t *testing.T) {
	m := service.New(nil)
	_, err := m.UpdateTLE(context.Background(), uuid.New(), "short", "alsobad", "alice")
	if err == nil {
		t.Fatal("expected error for short TLE")
	}
	var de *errs.E
	if !errors.As(err, &de) || de.Domain != errs.DomainInvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestUpdateTLERejectsWrongLineNumbers(t *testing.T) {
	// Two 69-char strings whose first chars are '3' and '4'.
	line1 := "3" + strings.Repeat(" ", 68)
	line2 := "4" + strings.Repeat(" ", 68)
	m := service.New(nil)
	_, err := m.UpdateTLE(context.Background(), uuid.New(), line1, line2, "alice")
	if err == nil {
		t.Fatal("expected error for wrong line numbers")
	}
	var de *errs.E
	if !errors.As(err, &de) || de.Domain != errs.DomainInvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestUpdateOrbitalStateRejectsInvalidState(t *testing.T) {
	m := service.New(nil)
	_, err := m.UpdateOrbitalState(context.Background(), uuid.New(), models.OrbitalState{}, "alice")
	if err == nil {
		t.Fatal("expected error for invalid state")
	}
	_, err = m.UpdateOrbitalState(context.Background(), uuid.New(), models.OrbitalState{Valid: true}, "alice")
	if err == nil {
		t.Fatal("expected error for zero epoch")
	}
	_ = time.Now
}

func TestSetModeRejectsUnspecified(t *testing.T) {
	m := service.New(nil)
	_, err := m.SetMode(context.Background(), uuid.New(), models.ModeUnspecified, "alice")
	if err == nil {
		t.Fatal("expected error for unspecified mode")
	}
	var de *errs.E
	if !errors.As(err, &de) || de.Domain != errs.DomainInvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestListSatellitesRequiresTenant(t *testing.T) {
	m := service.New(nil)
	_, err := m.ListSatellites(context.Background(), uuid.Nil, nil, uuid.Nil, 10)
	if err == nil {
		t.Fatal("expected error for nil tenant")
	}
}
