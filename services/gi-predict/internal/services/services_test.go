package services_test

import (
	"context"
	"testing"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	"github.com/ppusapati/space/services/gi-predict/internal/models"
	"github.com/ppusapati/space/services/gi-predict/internal/services"
)

func TestSubmitForecastJobRejectsEmpty(t *testing.T) {
	p := services.New(nil)
	_, err := p.SubmitForecastJob(context.Background(), services.SubmitForecastJobInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestSubmitForecastJobRejectsBadHorizon(t *testing.T) {
	p := services.New(nil)
	_, err := p.SubmitForecastJob(context.Background(), services.SubmitForecastJobInput{
		TenantID:    ulid.New(),
		Type:        models.TypeNDVITrend,
		HorizonDays: 0,
		InputURIs:   []string{"s3://x"},
	})
	if err == nil {
		t.Fatal("expected error for zero horizon_days")
	}
	_, err = p.SubmitForecastJob(context.Background(), services.SubmitForecastJobInput{
		TenantID:    ulid.New(),
		Type:        models.TypeNDVITrend,
		HorizonDays: 9999,
		InputURIs:   []string{"s3://x"},
	})
	if err == nil {
		t.Fatal("expected error for horizon > 3650")
	}
}

func TestSubmitForecastJobRejectsEmptyInputs(t *testing.T) {
	p := services.New(nil)
	_, err := p.SubmitForecastJob(context.Background(), services.SubmitForecastJobInput{
		TenantID:    ulid.New(),
		Type:        models.TypeNDVITrend,
		HorizonDays: 7,
	})
	if err == nil {
		t.Fatal("expected error for empty inputs")
	}
}

func TestUpdateForecastJobStatusRejectsUnspecified(t *testing.T) {
	p := services.New(nil)
	_, err := p.UpdateForecastJobStatus(context.Background(), services.UpdateForecastJobStatusInput{
		ID: ulid.New(), Status: models.StatusUnspecified,
	})
	if err == nil {
		t.Fatal("expected error for unspecified status")
	}
}

func TestListRequiresTenant(t *testing.T) {
	p := services.New(nil)
	if _, _, err := p.ListForecastJobsForTenant(context.Background(), services.ListForecastJobsInput{}); err == nil {
		t.Fatal("expected error for nil tenant")
	}
}
