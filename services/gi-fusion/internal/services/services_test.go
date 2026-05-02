package services_test

import (
	"context"
	"testing"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	"github.com/ppusapati/space/services/gi-fusion/internal/models"
	"github.com/ppusapati/space/services/gi-fusion/internal/services"
)

func TestSubmitFusionJobRejectsEmpty(t *testing.T) {
	f := services.New(nil)
	_, err := f.SubmitFusionJob(context.Background(), services.SubmitFusionJobInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestSubmitFusionJobRejectsUnspecifiedMethod(t *testing.T) {
	f := services.New(nil)
	_, err := f.SubmitFusionJob(context.Background(), services.SubmitFusionJobInput{
		TenantID:  ulid.New(),
		InputURIs: []string{"s3://x"},
	})
	if err == nil {
		t.Fatal("expected error for unspecified method")
	}
}

func TestSubmitFusionJobRejectsEmptyInputs(t *testing.T) {
	f := services.New(nil)
	_, err := f.SubmitFusionJob(context.Background(), services.SubmitFusionJobInput{
		TenantID: ulid.New(),
		Method:   models.MethodPanSharpen,
	})
	if err == nil {
		t.Fatal("expected error for empty inputs")
	}
}

func TestUpdateFusionJobStatusRejectsUnspecified(t *testing.T) {
	f := services.New(nil)
	_, err := f.UpdateFusionJobStatus(context.Background(), ulid.New(),
		models.StatusUnspecified, "", "", "")
	if err == nil {
		t.Fatal("expected error for unspecified status")
	}
}

func TestListRequiresTenant(t *testing.T) {
	f := services.New(nil)
	if _, _, err := f.ListFusionJobsForTenant(context.Background(), services.ListFusionJobsInput{}); err == nil {
		t.Fatal("expected error for nil tenant")
	}
}
