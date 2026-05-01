package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/ppusapati/space/pkg/errs"
	"github.com/ppusapati/space/services/eo-pipeline/internal/models"
	"github.com/ppusapati/space/services/eo-pipeline/internal/service"
)

func TestSubmitJobRejectsEmptyIdentifiers(t *testing.T) {
	p := service.New(nil)
	_, err := p.SubmitJob(context.Background(), service.SubmitJobInput{
		Stage: models.StageRadiometric,
	})
	if err == nil {
		t.Fatal("expected error for missing tenant_id / item_id")
	}
	var de *errs.E
	if !errors.As(err, &de) || de.Domain != errs.DomainInvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestSubmitJobRejectsUnspecifiedStage(t *testing.T) {
	p := service.New(nil)
	_, err := p.SubmitJob(context.Background(), service.SubmitJobInput{
		TenantID: uuid.New(), ItemID: uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error for unspecified stage")
	}
}
