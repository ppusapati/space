// Package service holds the eo-catalog business logic. Handlers
// translate proto requests to service calls; the service layer owns
// validation, ID generation, default-handling, and repository wiring.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/ppusapati/space/pkg/errs"
	"github.com/ppusapati/space/services/eo-catalog/internal/models"
	"github.com/ppusapati/space/services/eo-catalog/internal/repository"
)

// Catalog is the service-layer facade.
type Catalog struct {
	repo *repository.Repository
	// IDFn is the UUID generator. Tests override it for determinism.
	IDFn func() uuid.UUID
	// NowFn is the wall-clock source. Tests override it for determinism.
	NowFn func() time.Time
}

// New constructs a Catalog with the given repository.
func New(repo *repository.Repository) *Catalog {
	return &Catalog{
		repo:  repo,
		IDFn:  uuid.New,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// CreateCollectionInput is the input to [Catalog.CreateCollection].
type CreateCollectionInput struct {
	TenantID      uuid.UUID
	Slug          string
	Title         string
	Description   string
	License       string
	SpatialExtent models.BoundingBox
	TemporalStart *time.Time
	TemporalEnd   *time.Time
	CreatedBy     string
}

// CreateCollection persists a new Collection.
func (s *Catalog) CreateCollection(ctx context.Context, in CreateCollectionInput) (*models.Collection, error) {
	if in.Slug == "" || in.Title == "" || in.TenantID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id, slug, title required")
	}
	if in.TemporalStart != nil && in.TemporalEnd != nil && in.TemporalStart.After(*in.TemporalEnd) {
		return nil, errs.New(errs.DomainInvalidArgument, "temporal_start must be ≤ temporal_end")
	}
	c := &models.Collection{
		ID:            s.IDFn(),
		TenantID:      in.TenantID,
		Slug:          in.Slug,
		Title:         in.Title,
		Description:   in.Description,
		License:       in.License,
		SpatialExtent: in.SpatialExtent,
		TemporalStart: in.TemporalStart,
		TemporalEnd:   in.TemporalEnd,
		CreatedBy:     defaultIfEmpty(in.CreatedBy, "system"),
	}
	return s.repo.Collections.Create(ctx, c)
}

// GetCollection fetches by id.
func (s *Catalog) GetCollection(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
	c, err := s.repo.Collections.Get(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errs.New(errs.DomainNotFound, "collection %s", id)
	}
	return c, err
}

// ListCollections returns one page of collections.
func (s *Catalog) ListCollections(
	ctx context.Context, tenantID uuid.UUID, cursorTS *time.Time, cursorID uuid.UUID, limit int32,
) ([]*models.Collection, error) {
	if tenantID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id required")
	}
	return s.repo.Collections.ListAfter(ctx, tenantID, cursorTS, cursorID, limit)
}

// CreateItemInput is the input to [Catalog.CreateItem].
type CreateItemInput struct {
	TenantID        uuid.UUID
	CollectionID    uuid.UUID
	Mission         string
	Platform        string
	Instrument      string
	Datetime        time.Time
	BBox            models.BoundingBox
	GeometryGeoJSON string
	CloudCover      float64
	PropertiesJSON  string
	Assets          []models.Asset
	CreatedBy       string
}

// CreateItem persists a new Item with its initial Assets.
func (s *Catalog) CreateItem(ctx context.Context, in CreateItemInput) (*models.Item, error) {
	if in.TenantID == uuid.Nil || in.CollectionID == uuid.Nil || in.Mission == "" {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id, collection_id, mission required")
	}
	if in.CloudCover < 0 || in.CloudCover > 100 {
		return nil, errs.New(errs.DomainInvalidArgument, "cloud_cover must be in [0, 100]")
	}
	if in.Datetime.IsZero() {
		return nil, errs.New(errs.DomainInvalidArgument, "datetime required")
	}
	if in.PropertiesJSON == "" {
		in.PropertiesJSON = "{}"
	}
	it := &models.Item{
		ID:              s.IDFn(),
		TenantID:        in.TenantID,
		CollectionID:    in.CollectionID,
		Mission:         in.Mission,
		Platform:        in.Platform,
		Instrument:      in.Instrument,
		Datetime:        in.Datetime,
		BBox:            in.BBox,
		GeometryGeoJSON: in.GeometryGeoJSON,
		CloudCover:      in.CloudCover,
		PropertiesJSON:  in.PropertiesJSON,
		Assets:          in.Assets,
		CreatedBy:       defaultIfEmpty(in.CreatedBy, "system"),
	}
	return s.repo.Items.CreateWithAssets(ctx, it)
}

// GetItem fetches by id.
func (s *Catalog) GetItem(ctx context.Context, id uuid.UUID) (*models.Item, error) {
	it, err := s.repo.Items.Get(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errs.New(errs.DomainNotFound, "item %s", id)
	}
	return it, err
}

// SearchItemsInput is the input to [Catalog.SearchItems].
type SearchItemsInput struct {
	TenantID       uuid.UUID
	CollectionID   *uuid.UUID
	DatetimeStart  time.Time
	DatetimeEnd    time.Time
	BBox           *models.BoundingBox
	MaxCloudCover  *float64
	CursorDatetime *time.Time
	CursorID       uuid.UUID
	Limit          int32
}

// SearchItems runs the temporal + spatial + cloud-cover filter.
func (s *Catalog) SearchItems(ctx context.Context, in SearchItemsInput) ([]*models.Item, error) {
	if in.TenantID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id required")
	}
	if in.DatetimeStart.IsZero() || in.DatetimeEnd.IsZero() {
		return nil, errs.New(errs.DomainInvalidArgument, "datetime range required")
	}
	if in.DatetimeStart.After(in.DatetimeEnd) {
		return nil, errs.New(errs.DomainInvalidArgument, "datetime_start must be ≤ datetime_end")
	}
	return s.repo.Items.Search(ctx, repository.SearchParams{
		TenantID:       in.TenantID,
		CollectionID:   in.CollectionID,
		DatetimeStart:  in.DatetimeStart,
		DatetimeEnd:    in.DatetimeEnd,
		BBox:           in.BBox,
		MaxCloudCover:  in.MaxCloudCover,
		CursorDatetime: in.CursorDatetime,
		CursorID:       in.CursorID,
		Limit:          in.Limit,
	})
}

// RecordQualityInput is the input to [Catalog.RecordQuality].
type RecordQualityInput struct {
	ItemID             uuid.UUID
	CloudCover         float64
	RadiometricRMSE    float64
	GeometricAccuracyM float64
	Notes              string
}

// RecordQuality persists a new QualityResult.
func (s *Catalog) RecordQuality(ctx context.Context, in RecordQualityInput) (*models.QualityResult, error) {
	if in.ItemID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "item_id required")
	}
	if in.CloudCover < 0 || in.CloudCover > 1 {
		return nil, errs.New(errs.DomainInvalidArgument, "cloud_cover must be in [0, 1]")
	}
	if in.RadiometricRMSE < 0 || in.GeometricAccuracyM < 0 {
		return nil, errs.New(errs.DomainInvalidArgument, "RMSE and accuracy must be ≥ 0")
	}
	q := &models.QualityResult{
		ID:                 s.IDFn(),
		ItemID:             in.ItemID,
		CloudCover:         in.CloudCover,
		RadiometricRMSE:    in.RadiometricRMSE,
		GeometricAccuracyM: in.GeometricAccuracyM,
		Notes:              in.Notes,
	}
	return s.repo.Quality.Record(ctx, q)
}

// ListQuality returns QA results for an item, newest first.
func (s *Catalog) ListQuality(
	ctx context.Context, itemID uuid.UUID, cursorTS *time.Time, cursorID uuid.UUID, limit int32,
) ([]*models.QualityResult, error) {
	if itemID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "item_id required")
	}
	return s.repo.Quality.ListForItem(ctx, itemID, cursorTS, cursorID, limit)
}

func defaultIfEmpty(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
