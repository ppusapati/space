package services_test

import (
	"context"
	"testing"
	"time"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	"github.com/ppusapati/space/services/gs-rf/internal/services"
)

func TestCreateLinkBudgetRejectsZeroIDs(t *testing.T) {
	s := services.New(nil)
	_, err := s.CreateLinkBudget(context.Background(), services.CreateLinkBudgetInput{})
	if err == nil {
		t.Fatal("expected error for zero ids")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestCreateLinkBudgetRejectsZeroFreq(t *testing.T) {
	s := services.New(nil)
	_, err := s.CreateLinkBudget(context.Background(), services.CreateLinkBudgetInput{
		TenantID: ulid.New(), PassID: ulid.New(), StationID: ulid.New(),
		AntennaID: ulid.New(), SatelliteID: ulid.New(),
	})
	if err == nil {
		t.Fatal("expected error for zero carrier_freq_hz")
	}
}

func TestRecordMeasurementRejectsZeroIDs(t *testing.T) {
	s := services.New(nil)
	_, err := s.RecordMeasurement(context.Background(), services.RecordMeasurementInput{})
	if err == nil {
		t.Fatal("expected error for zero ids")
	}
}

func TestRecordMeasurementRejectsZeroSampledAt(t *testing.T) {
	s := services.New(nil)
	_, err := s.RecordMeasurement(context.Background(), services.RecordMeasurementInput{
		TenantID: ulid.New(), PassID: ulid.New(),
		StationID: ulid.New(), AntennaID: ulid.New(),
	})
	if err == nil {
		t.Fatal("expected error for zero sampled_at")
	}
}

func TestRecordMeasurementRejectsBadBER(t *testing.T) {
	s := services.New(nil)
	_, err := s.RecordMeasurement(context.Background(), services.RecordMeasurementInput{
		TenantID: ulid.New(), PassID: ulid.New(),
		StationID: ulid.New(), AntennaID: ulid.New(),
		SampledAt: time.Now(), BER: 1.5,
	})
	if err == nil {
		t.Fatal("expected error for ber > 1")
	}
}

func TestListsRequireTenant(t *testing.T) {
	s := services.New(nil)
	if _, _, err := s.ListLinkBudgetsForTenant(context.Background(), services.ListLinkBudgetsInput{}); err == nil {
		t.Fatal("expected error for nil tenant on budgets")
	}
	if _, _, err := s.ListMeasurementsForTenant(context.Background(), services.ListMeasurementsInput{}); err == nil {
		t.Fatal("expected error for nil tenant on measurements")
	}
}

func TestListMeasurementsRejectsInvertedTimeRange(t *testing.T) {
	s := services.New(nil)
	t1 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	_, _, err := s.ListMeasurementsForTenant(context.Background(), services.ListMeasurementsInput{
		TenantID: ulid.New(), TimeStart: t1, TimeEnd: t0,
	})
	if err == nil {
		t.Fatal("expected error for end < start")
	}
}
