// Package services holds eo-catalog business logic.
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	"github.com/ppusapati/space/services/eo-catalog/internal/models"
	"github.com/ppusapati/space/services/eo-catalog/internal/repository"
)

// Catalog is the eo-catalog service-layer facade.
type Catalog struct {
	repo  *repository.Repo
	IDFn  func() ulid.ID
	NowFn func() time.Time
}

// New constructs a Catalog service with monotonic ULIDs and UTC timestamps.
func New(repo *repository.Repo) *Catalog {
	return &Catalog{
		repo:  repo,
		IDFn:  ulid.NewMonotonic,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// ----- Collections ---------------------------------------------------------

// CreateCollectionInput is the input for [Catalog.CreateCollection].
type CreateCollectionInput struct {
	TenantID      ulid.ID
	Slug          string
	Title         string
	Description   string
	License       string
	SpatialExtent models.BoundingBox
	TemporalStart time.Time
	TemporalEnd   time.Time
	CreatedBy     string
}

// CreateCollection persists a new collection.
func (c *Catalog) CreateCollection(ctx context.Context, in CreateCollectionInput) (*models.Collection, error) {
	if in.TenantID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	in.Slug = strings.TrimSpace(in.Slug)
	in.Title = strings.TrimSpace(in.Title)
	if in.Slug == "" || in.Title == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "slug and title required")
	}
	if !in.TemporalStart.IsZero() && !in.TemporalEnd.IsZero() && in.TemporalEnd.Before(in.TemporalStart) {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "temporal_end must be >= temporal_start")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return c.repo.CreateCollection(ctx, repository.CreateCollectionParams{
		ID:            c.IDFn(),
		TenantID:      in.TenantID,
		Slug:          in.Slug,
		Title:         in.Title,
		Description:   in.Description,
		License:       in.License,
		SpatialExtent: in.SpatialExtent,
		TemporalStart: in.TemporalStart,
		TemporalEnd:   in.TemporalEnd,
		CreatedBy:     createdBy,
	})
}

// GetCollection fetches a collection by id, returning a domain not-found
// error when missing.
func (c *Catalog) GetCollection(ctx context.Context, id ulid.ID) (*models.Collection, error) {
	col, err := c.repo.GetCollection(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("COLLECTION_NOT_FOUND", "collection "+id.String())
	}
	return col, err
}

// DeleteCollection removes a collection.
func (c *Catalog) DeleteCollection(ctx context.Context, id ulid.ID) error {
	if err := c.repo.DeleteCollection(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return pkgerrors.NotFound("COLLECTION_NOT_FOUND", "collection "+id.String())
		}
		return err
	}
	return nil
}

// ListCollectionsForTenant returns one page of collections.
func (c *Catalog) ListCollectionsForTenant(
	ctx context.Context, tenantID ulid.ID, offset, size int32,
) ([]*models.Collection, models.Page, error) {
	if tenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := c.repo.ListCollectionsForTenant(ctx, tenantID, offset, size)
	if err != nil {
		return nil, models.Page{}, err
	}
	return rows, models.Page{
		TotalCount: total,
		PageOffset: offset,
		PageSize:   size,
		HasNext:    offset+int32(len(rows)) < total,
	}, nil
}

// ----- Items ---------------------------------------------------------------

// CreateItemInput is the input for [Catalog.CreateItem].
type CreateItemInput struct {
	TenantID        ulid.ID
	CollectionID    ulid.ID
	Mission         string
	Platform        string
	Instrument      string
	Datetime        time.Time
	BBox            models.BoundingBox
	GeometryGeoJSON string
	CloudCover      float64
	PropertiesJSON  string
	CreatedBy       string
}

// CreateItem persists a new item, returning a not-found error when the
// referenced collection does not exist.
func (c *Catalog) CreateItem(ctx context.Context, in CreateItemInput) (*models.Item, error) {
	if in.TenantID.IsZero() || in.CollectionID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id and collection_id required")
	}
	if strings.TrimSpace(in.Mission) == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "mission required")
	}
	if in.Datetime.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "datetime required")
	}
	if in.CloudCover < 0 || in.CloudCover > 100 {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "cloud_cover must be in [0, 100]")
	}
	col, err := c.repo.GetCollection(ctx, in.CollectionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, pkgerrors.BadRequest("COLLECTION_NOT_FOUND",
				"collection "+in.CollectionID.String())
		}
		return nil, err
	}
	if col.TenantID != in.TenantID {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT",
			"collection tenant mismatch")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return c.repo.CreateItem(ctx, repository.CreateItemParams{
		ID:              c.IDFn(),
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
		CreatedBy:       createdBy,
	})
}

// GetItem fetches an item by id with its assets.
func (c *Catalog) GetItem(ctx context.Context, id ulid.ID) (*models.Item, error) {
	item, err := c.repo.GetItem(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("ITEM_NOT_FOUND", "item "+id.String())
	}
	return item, err
}

// ListItemsForTenant returns one page of items.
func (c *Catalog) ListItemsForTenant(
	ctx context.Context, tenantID ulid.ID, collectionID *ulid.ID, offset, size int32,
) ([]*models.Item, models.Page, error) {
	if tenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := c.repo.ListItemsForTenant(ctx, repository.ListItemsParams{
		TenantID:     tenantID,
		CollectionID: collectionID,
		PageOffset:   offset,
		PageSize:     size,
	})
	if err != nil {
		return nil, models.Page{}, err
	}
	return rows, models.Page{
		TotalCount: total,
		PageOffset: offset,
		PageSize:   size,
		HasNext:    offset+int32(len(rows)) < total,
	}, nil
}

// AddAsset attaches a new asset to an existing item.
func (c *Catalog) AddAsset(ctx context.Context, itemID ulid.ID, a models.Asset) (*models.Item, error) {
	if itemID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "item_id required")
	}
	a.Key = strings.TrimSpace(a.Key)
	a.Href = strings.TrimSpace(a.Href)
	if a.Key == "" || a.Href == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "asset.key and asset.href required")
	}
	// Verify item exists.
	if _, err := c.GetItem(ctx, itemID); err != nil {
		return nil, err
	}
	return c.repo.AddAsset(ctx, c.IDFn(), itemID, a)
}

// ----- Quality -------------------------------------------------------------

// RecordQualityResultInput is the input for [Catalog.RecordQualityResult].
type RecordQualityResultInput struct {
	ItemID             ulid.ID
	CloudCover         float64
	RadiometricRMSE    float64
	GeometricAccuracyM float64
	Notes              string
}

// RecordQualityResult persists a new quality result.
func (c *Catalog) RecordQualityResult(ctx context.Context, in RecordQualityResultInput) (*models.QualityResult, error) {
	if in.ItemID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "item_id required")
	}
	if in.CloudCover < 0 || in.CloudCover > 1 {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "cloud_cover must be in [0, 1]")
	}
	// Ensure the parent item exists for referential clarity.
	if _, err := c.GetItem(ctx, in.ItemID); err != nil {
		return nil, err
	}
	return c.repo.RecordQuality(ctx, repository.RecordQualityParams{
		ID:                 c.IDFn(),
		ItemID:             in.ItemID,
		CloudCover:         in.CloudCover,
		RadiometricRMSE:    in.RadiometricRMSE,
		GeometricAccuracyM: in.GeometricAccuracyM,
		Notes:              in.Notes,
	})
}
