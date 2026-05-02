package services_test

import (
	"context"
	"testing"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	"github.com/ppusapati/space/services/gi-tiles/internal/models"
	"github.com/ppusapati/space/services/gi-tiles/internal/services"
)

func TestCreateTileSetRejectsEmpty(t *testing.T) {
	s := services.New(nil)
	_, err := s.CreateTileSet(context.Background(), services.CreateTileSetInput{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestCreateTileSetRejectsUnspecifiedFormat(t *testing.T) {
	s := services.New(nil)
	_, err := s.CreateTileSet(context.Background(), services.CreateTileSetInput{
		TenantID:   ulid.New(),
		Slug:       "ortho",
		Name:       "Orthophoto",
		Projection: "EPSG:3857",
		MinZoom:    0, MaxZoom: 18,
		SourceURI: "s3://x",
	})
	if err == nil {
		t.Fatal("expected error for unspecified format")
	}
}

func TestCreateTileSetRejectsBadZoomRange(t *testing.T) {
	s := services.New(nil)
	_, err := s.CreateTileSet(context.Background(), services.CreateTileSetInput{
		TenantID:   ulid.New(),
		Slug:       "ortho",
		Name:       "Orthophoto",
		Format:     models.FormatPNG,
		Projection: "EPSG:3857",
		MinZoom:    18, MaxZoom: 5,
		SourceURI: "s3://x",
	})
	if err == nil {
		t.Fatal("expected error for max < min zoom")
	}
}

func TestListRequiresTenant(t *testing.T) {
	s := services.New(nil)
	if _, _, err := s.ListTileSetsForTenant(context.Background(), services.ListTileSetsInput{}); err == nil {
		t.Fatal("expected error for nil tenant")
	}
}
