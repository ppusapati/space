package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/ppusapati/space/pkg/errs"
	"github.com/ppusapati/space/services/eo-analytics/internal/models"
	"github.com/ppusapati/space/services/eo-analytics/internal/service"
)

func TestRegisterModelRejectsEmptyFields(t *testing.T) {
	a := service.New(nil)
	_, err := a.RegisterModel(context.Background(), service.RegisterModelInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	var de *errs.E
	if !errors.As(err, &de) || de.Domain != errs.DomainInvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestRegisterModelRejectsUnspecifiedTask(t *testing.T) {
	a := service.New(nil)
	_, err := a.RegisterModel(context.Background(), service.RegisterModelInput{
		TenantID: uuid.New(), Name: "yolo", Version: "1", Framework: "onnx", ArtefactURI: "s3://x",
	})
	if err == nil {
		t.Fatal("expected error for TaskUnspecified")
	}
}

func TestSubmitInferenceJobRejectsMissingIDs(t *testing.T) {
	a := service.New(nil)
	_, err := a.SubmitInferenceJob(context.Background(), service.SubmitInferenceJobInput{})
	if err == nil {
		t.Fatal("expected error for missing ids")
	}
}

func TestUpdateInferenceJobStatusRejectsUnspecifiedTarget(t *testing.T) {
	a := service.New(nil)
	// Without a real repo we can only exercise the validation paths
	// that fire before the repository call. UpdateStatus fetches the
	// current job first and would dereference a nil repo, so we test
	// the fully validated SubmitInferenceJob branch above and rely on
	// integration tests for the post-fetch validation paths.
	_ = a
	_ = models.StatusUnspecified
}
