package services_test

import (
	"context"
	"testing"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	"github.com/ppusapati/space/services/gi-reports/internal/models"
	"github.com/ppusapati/space/services/gi-reports/internal/services"
)

func TestCreateTemplateRejectsEmpty(t *testing.T) {
	r := services.New(nil)
	_, err := r.CreateTemplate(context.Background(), services.CreateTemplateInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestCreateTemplateRejectsUnspecifiedFormat(t *testing.T) {
	r := services.New(nil)
	_, err := r.CreateTemplate(context.Background(), services.CreateTemplateInput{
		TenantID:    ulid.New(),
		Slug:        "weekly",
		Name:        "Weekly Report",
		TemplateURI: "s3://x",
	})
	if err == nil {
		t.Fatal("expected error for unspecified format")
	}
}

func TestGenerateReportRejectsZeroIDs(t *testing.T) {
	r := services.New(nil)
	_, err := r.GenerateReport(context.Background(), services.GenerateReportInput{})
	if err == nil {
		t.Fatal("expected error for zero ids")
	}
}

func TestUpdateReportStatusRejectsUnspecified(t *testing.T) {
	r := services.New(nil)
	_, err := r.UpdateReportStatus(context.Background(), ulid.New(),
		models.StatusUnspecified, "", "", "")
	if err == nil {
		t.Fatal("expected error for unspecified status")
	}
}

func TestListsRequireTenant(t *testing.T) {
	r := services.New(nil)
	if _, _, err := r.ListTemplatesForTenant(context.Background(), services.ListTemplatesInput{}); err == nil {
		t.Fatal("expected error for nil tenant on templates")
	}
	if _, _, err := r.ListReportsForTenant(context.Background(), services.ListReportsInput{}); err == nil {
		t.Fatal("expected error for nil tenant on reports")
	}
}
