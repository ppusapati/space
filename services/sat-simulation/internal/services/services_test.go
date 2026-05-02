package services_test

import (
	"context"
	"testing"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	"github.com/ppusapati/space/services/sat-simulation/internal/models"
	"github.com/ppusapati/space/services/sat-simulation/internal/services"
)

func TestCreateScenarioRejectsEmptyTenant(t *testing.T) {
	s := services.New(nil)
	_, err := s.CreateScenario(context.Background(), services.CreateScenarioInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestCreateScenarioRejectsEmptySlug(t *testing.T) {
	s := services.New(nil)
	_, err := s.CreateScenario(context.Background(), services.CreateScenarioInput{
		TenantID: ulid.New(),
	})
	if err == nil {
		t.Fatal("expected error for empty slug/title")
	}
}

func TestStartRunRejectsZeroIDs(t *testing.T) {
	s := services.New(nil)
	_, err := s.StartRun(context.Background(), services.StartRunInput{})
	if err == nil {
		t.Fatal("expected error for zero ids")
	}
}

func TestStartRunRejectsUnspecifiedMode(t *testing.T) {
	s := services.New(nil)
	_, err := s.StartRun(context.Background(), services.StartRunInput{
		TenantID:    ulid.New(),
		SatelliteID: ulid.New(),
		ScenarioID:  ulid.New(),
	})
	if err == nil {
		t.Fatal("expected error for unspecified mode")
	}
}

func TestUpdateRunStatusRejectsUnspecified(t *testing.T) {
	s := services.New(nil)
	_, err := s.UpdateRunStatus(context.Background(), services.UpdateRunStatusInput{
		ID:     ulid.New(),
		Status: models.StatusUnspecified,
	})
	if err == nil {
		t.Fatal("expected error for unspecified status")
	}
}

func TestListsRequireTenant(t *testing.T) {
	s := services.New(nil)
	if _, _, err := s.ListScenariosForTenant(context.Background(), ulid.Zero, 0, 10); err == nil {
		t.Fatal("expected error for nil tenant on scenarios")
	}
	if _, _, err := s.ListRunsForTenant(context.Background(), services.ListRunsInput{}); err == nil {
		t.Fatal("expected error for nil tenant on runs")
	}
}
