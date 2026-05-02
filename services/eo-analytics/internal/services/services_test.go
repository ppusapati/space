package services_test

import (
	"context"
	"errors"
	"testing"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	"github.com/ppusapati/space/services/eo-analytics/internal/models"
	"github.com/ppusapati/space/services/eo-analytics/internal/services"
)

func TestRegisterModelRejectsEmptyTenant(t *testing.T) {
	a := services.New(nil)
	_, err := a.RegisterModel(context.Background(), services.RegisterModelInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestRegisterModelRejectsMissingTask(t *testing.T) {
	a := services.New(nil)
	_, err := a.RegisterModel(context.Background(), services.RegisterModelInput{
		TenantID:    ulid.New(),
		Name:        "yolov8",
		Version:     "1.0.0",
		Framework:   "onnx",
		ArtefactURI: "s3://bucket/yolov8.onnx",
	})
	if err == nil {
		t.Fatal("expected error for unspecified task")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestSubmitInferenceJobRejectsZeroIDs(t *testing.T) {
	a := services.New(nil)
	_, err := a.SubmitInferenceJob(context.Background(), services.SubmitInferenceJobInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestUpdateInferenceJobStatusRejectsUnspecified(t *testing.T) {
	a := services.New(nil)
	_, err := a.UpdateInferenceJobStatus(context.Background(), ulid.New(), models.StatusUnspecified, "", "", "")
	if err == nil {
		t.Fatal("expected error for unspecified status")
	}
}

func TestListModelsRequiresTenant(t *testing.T) {
	a := services.New(nil)
	_, _, err := a.ListModelsForTenant(context.Background(), services.ListModelsInput{})
	if err == nil {
		t.Fatal("expected error for nil tenant")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestListInferenceJobsRequiresTenant(t *testing.T) {
	a := services.New(nil)
	_, _, err := a.ListInferenceJobsForTenant(context.Background(), services.ListInferenceJobsInput{})
	if err == nil {
		t.Fatal("expected error for nil tenant")
	}
}

func TestDeactivateModelRejectsZeroID(t *testing.T) {
	a := services.New(nil)
	_, err := a.DeactivateModel(context.Background(), ulid.Zero, "user")
	if err == nil {
		t.Fatal("expected error for zero id")
	}
}

func TestErrorTypePassesThrough(t *testing.T) {
	a := services.New(nil)
	_, err := a.RegisterModel(context.Background(), services.RegisterModelInput{})
	var pe *pkgerrors.Error
	if !errors.As(err, &pe) {
		t.Fatalf("expected packages/errors.Error, got %T %v", err, err)
	}
}
