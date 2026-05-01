package services_test

import (
	"context"
	"testing"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	"github.com/ppusapati/space/services/gi-analytics/internal/models"
	"github.com/ppusapati/space/services/gi-analytics/internal/services"
)

func TestSubmitAnalysisJobRejectsEmpty(t *testing.T) {
	a := services.New(nil)
	_, err := a.SubmitAnalysisJob(context.Background(), services.SubmitAnalysisJobInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestSubmitAnalysisJobRejectsUnspecifiedType(t *testing.T) {
	a := services.New(nil)
	_, err := a.SubmitAnalysisJob(context.Background(), services.SubmitAnalysisJobInput{
		TenantID:  ulid.New(),
		InputURIs: []string{"s3://x"},
	})
	if err == nil {
		t.Fatal("expected error for unspecified type")
	}
}

func TestSubmitAnalysisJobRejectsEmptyInputs(t *testing.T) {
	a := services.New(nil)
	_, err := a.SubmitAnalysisJob(context.Background(), services.SubmitAnalysisJobInput{
		TenantID: ulid.New(),
		Type:     models.TypeNDVITimeSeries,
	})
	if err == nil {
		t.Fatal("expected error for empty inputs")
	}
}

func TestUpdateAnalysisJobStatusRejectsUnspecified(t *testing.T) {
	a := services.New(nil)
	_, err := a.UpdateAnalysisJobStatus(context.Background(), services.UpdateAnalysisJobStatusInput{
		ID: ulid.New(), Status: models.StatusUnspecified,
	})
	if err == nil {
		t.Fatal("expected error for unspecified status")
	}
}

func TestListRequiresTenant(t *testing.T) {
	a := services.New(nil)
	if _, _, err := a.ListAnalysisJobsForTenant(context.Background(), services.ListAnalysisJobsInput{}); err == nil {
		t.Fatal("expected error for nil tenant")
	}
}
