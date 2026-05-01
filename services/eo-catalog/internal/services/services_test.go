package services_test

import (
	"context"
	"errors"
	"testing"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	"github.com/ppusapati/space/services/eo-catalog/internal/services"
)

func TestCreateCollectionRejectsEmptyTenant(t *testing.T) {
	c := services.New(nil)
	_, err := c.CreateCollection(context.Background(), services.CreateCollectionInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestCreateCollectionRejectsEmptySlug(t *testing.T) {
	c := services.New(nil)
	_, err := c.CreateCollection(context.Background(), services.CreateCollectionInput{
		TenantID: ulid.New(),
	})
	if err == nil {
		t.Fatal("expected error for empty slug/title")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestCreateItemRejectsBadCloudCover(t *testing.T) {
	c := services.New(nil)
	_, err := c.CreateItem(context.Background(), services.CreateItemInput{
		TenantID:     ulid.New(),
		CollectionID: ulid.New(),
		Mission:      "sentinel-2",
		CloudCover:   -1,
	})
	if err == nil {
		t.Fatal("expected error for cloud_cover < 0")
	}
}

func TestCreateItemRejectsMissingFields(t *testing.T) {
	c := services.New(nil)
	_, err := c.CreateItem(context.Background(), services.CreateItemInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestRecordQualityRejectsBadCloudCover(t *testing.T) {
	c := services.New(nil)
	_, err := c.RecordQualityResult(context.Background(), services.RecordQualityResultInput{
		ItemID:     ulid.New(),
		CloudCover: 5.0,
	})
	if err == nil {
		t.Fatal("expected error for cloud_cover > 1")
	}
}

func TestRecordQualityRejectsZeroItem(t *testing.T) {
	c := services.New(nil)
	_, err := c.RecordQualityResult(context.Background(), services.RecordQualityResultInput{})
	if err == nil {
		t.Fatal("expected error for zero item id")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestListCollectionsRequiresTenant(t *testing.T) {
	c := services.New(nil)
	_, _, err := c.ListCollectionsForTenant(context.Background(), ulid.Zero, 0, 10)
	if err == nil {
		t.Fatal("expected error for zero tenant")
	}
}

func TestErrorTypePassesThrough(t *testing.T) {
	c := services.New(nil)
	_, err := c.CreateCollection(context.Background(), services.CreateCollectionInput{})
	var pe *pkgerrors.Error
	if !errors.As(err, &pe) {
		t.Fatalf("expected packages/errors.Error, got %T %v", err, err)
	}
}
