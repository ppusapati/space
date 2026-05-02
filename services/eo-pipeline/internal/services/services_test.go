package services_test

import (
	"context"
	"testing"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	"github.com/ppusapati/space/services/eo-pipeline/internal/models"
	"github.com/ppusapati/space/services/eo-pipeline/internal/services"
)

func TestSubmitJobRejectsEmptyIDs(t *testing.T) {
	p := services.New(nil)
	_, err := p.SubmitJob(context.Background(), services.SubmitJobInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestSubmitJobRejectsUnspecifiedStage(t *testing.T) {
	p := services.New(nil)
	_, err := p.SubmitJob(context.Background(), services.SubmitJobInput{
		TenantID: ulid.New(),
		ItemID:   ulid.New(),
	})
	if err == nil {
		t.Fatal("expected error for unspecified stage")
	}
}

func TestUpdateJobStatusRejectsUnspecified(t *testing.T) {
	p := services.New(nil)
	_, err := p.UpdateJobStatus(context.Background(), ulid.New(), models.StatusUnspecified, "", "", "")
	if err == nil {
		t.Fatal("expected error for unspecified status")
	}
}

func TestListJobsRequiresTenant(t *testing.T) {
	p := services.New(nil)
	_, _, err := p.ListJobsForTenant(context.Background(), services.ListJobsInput{})
	if err == nil {
		t.Fatal("expected error for nil tenant")
	}
}
