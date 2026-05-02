// Package repository wraps the eo-catalog sqlc layer.
package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/chetana/packages/ulid"

	eocatalogdb "github.com/ppusapati/space/services/eo-catalog/db/generated"
	"github.com/ppusapati/space/services/eo-catalog/internal/mapper"
	"github.com/ppusapati/space/services/eo-catalog/internal/models"
)

// ErrNotFound is returned when a row lookup misses.
var ErrNotFound = errors.New("repository: not found")

// Repo is the eo-catalog repository.
type Repo struct {
	q    *eocatalogdb.Queries
	pool *pgxpool.Pool
}

// New constructs a Repo from a pgx pool.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: eocatalogdb.New(pool), pool: pool}
}

// ----- Collections ---------------------------------------------------------

// CreateCollectionParams holds the input for [Repo.CreateCollection].
type CreateCollectionParams struct {
	ID            ulid.ID
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

// CreateCollection inserts a new collection.
func (r *Repo) CreateCollection(ctx context.Context, p CreateCollectionParams) (*models.Collection, error) {
	row, err := r.q.CreateCollection(ctx, eocatalogdb.CreateCollectionParams{
		ID:            mapper.PgUUID(p.ID),
		TenantID:      mapper.PgUUID(p.TenantID),
		Slug:          p.Slug,
		Title:         p.Title,
		Description:   p.Description,
		License:       p.License,
		BboxLonMin:    mapper.FloatPtr(p.SpatialExtent.LonMin, p.SpatialExtent.Valid),
		BboxLatMin:    mapper.FloatPtr(p.SpatialExtent.LatMin, p.SpatialExtent.Valid),
		BboxLonMax:    mapper.FloatPtr(p.SpatialExtent.LonMax, p.SpatialExtent.Valid),
		BboxLatMax:    mapper.FloatPtr(p.SpatialExtent.LatMax, p.SpatialExtent.Valid),
		TemporalStart: mapper.PgTimestampOrNull(p.TemporalStart),
		TemporalEnd:   mapper.PgTimestampOrNull(p.TemporalEnd),
		CreatedBy:     p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.CollectionFromRow(row), nil
}

// GetCollection fetches a collection by id.
func (r *Repo) GetCollection(ctx context.Context, id ulid.ID) (*models.Collection, error) {
	row, err := r.q.GetCollection(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.CollectionFromRow(row), nil
}

// DeleteCollection removes a collection. Returns ErrNotFound when no row matched.
func (r *Repo) DeleteCollection(ctx context.Context, id ulid.ID) error {
	if err := r.q.DeleteCollection(ctx, mapper.PgUUID(id)); err != nil {
		return err
	}
	return nil
}

// ListCollectionsForTenant returns one page of collections.
func (r *Repo) ListCollectionsForTenant(
	ctx context.Context, tenantID ulid.ID, offset, size int32,
) ([]*models.Collection, int32, error) {
	total, err := r.q.CountCollectionsForTenant(ctx, mapper.PgUUID(tenantID))
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListCollectionsForTenant(ctx, eocatalogdb.ListCollectionsForTenantParams{
		TenantID:   mapper.PgUUID(tenantID),
		PageOffset: offset,
		PageSize:   size,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.Collection, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.CollectionFromRow(row))
	}
	return out, int32(total), nil
}

// ----- Items ---------------------------------------------------------------

// CreateItemParams holds the input for [Repo.CreateItem].
type CreateItemParams struct {
	ID              ulid.ID
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

// CreateItem inserts a new item.
func (r *Repo) CreateItem(ctx context.Context, p CreateItemParams) (*models.Item, error) {
	props := strings.TrimSpace(p.PropertiesJSON)
	if props == "" {
		props = "{}"
	}
	row, err := r.q.CreateItem(ctx, eocatalogdb.CreateItemParams{
		ID:              mapper.PgUUID(p.ID),
		TenantID:        mapper.PgUUID(p.TenantID),
		CollectionID:    mapper.PgUUID(p.CollectionID),
		Mission:         p.Mission,
		Platform:        p.Platform,
		Instrument:      p.Instrument,
		Datetime:        mapper.PgTimestamp(p.Datetime),
		BboxLonMin:      mapper.FloatPtr(p.BBox.LonMin, p.BBox.Valid),
		BboxLatMin:      mapper.FloatPtr(p.BBox.LatMin, p.BBox.Valid),
		BboxLonMax:      mapper.FloatPtr(p.BBox.LonMax, p.BBox.Valid),
		BboxLatMax:      mapper.FloatPtr(p.BBox.LatMax, p.BBox.Valid),
		GeometryGeojson: p.GeometryGeoJSON,
		CloudCover:      p.CloudCover,
		PropertiesJson:  []byte(props),
		CreatedBy:       p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.ItemFromRow(row), nil
}

// GetItem fetches an item by id, including its assets.
func (r *Repo) GetItem(ctx context.Context, id ulid.ID) (*models.Item, error) {
	row, err := r.q.GetItem(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	item := mapper.ItemFromRow(row)
	assets, err := r.q.ListAssetsForItem(ctx, mapper.PgUUID(id))
	if err != nil {
		return nil, err
	}
	for _, a := range assets {
		item.Assets = append(item.Assets, mapper.AssetFromRow(a))
	}
	return item, nil
}

// ListItemsParams holds the input for [Repo.ListItemsForTenant].
type ListItemsParams struct {
	TenantID     ulid.ID
	CollectionID *ulid.ID // nil = any
	PageOffset   int32
	PageSize     int32
}

// ListItemsForTenant returns one page of items, with their assets attached.
func (r *Repo) ListItemsForTenant(ctx context.Context, p ListItemsParams) ([]*models.Item, int32, error) {
	var cid pgtype.UUID
	if p.CollectionID != nil {
		cid = mapper.PgUUID(*p.CollectionID)
	}
	total, err := r.q.CountItemsForTenant(ctx, eocatalogdb.CountItemsForTenantParams{
		TenantID:     mapper.PgUUID(p.TenantID),
		CollectionID: cid,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListItemsForTenant(ctx, eocatalogdb.ListItemsForTenantParams{
		TenantID:     mapper.PgUUID(p.TenantID),
		CollectionID: cid,
		PageOffset:   p.PageOffset,
		PageSize:     p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.Item, 0, len(rows))
	for _, row := range rows {
		item := mapper.ItemFromRow(row)
		assets, err := r.q.ListAssetsForItem(ctx, row.ID)
		if err != nil {
			return nil, 0, err
		}
		for _, a := range assets {
			item.Assets = append(item.Assets, mapper.AssetFromRow(a))
		}
		out = append(out, item)
	}
	return out, int32(total), nil
}

// ----- Assets --------------------------------------------------------------

// AddAsset inserts a new asset for an existing item.
func (r *Repo) AddAsset(ctx context.Context, id, itemID ulid.ID, a models.Asset) (*models.Item, error) {
	if _, err := r.q.CreateAsset(ctx, eocatalogdb.CreateAssetParams{
		ID:        mapper.PgUUID(id),
		ItemID:    mapper.PgUUID(itemID),
		Key:       a.Key,
		Href:      a.Href,
		MediaType: a.MediaType,
		Title:     a.Title,
		Roles:     a.Roles,
	}); err != nil {
		return nil, err
	}
	return r.GetItem(ctx, itemID)
}

// ----- Quality -------------------------------------------------------------

// RecordQualityParams holds the input for [Repo.RecordQuality].
type RecordQualityParams struct {
	ID                 ulid.ID
	ItemID             ulid.ID
	CloudCover         float64
	RadiometricRMSE    float64
	GeometricAccuracyM float64
	Notes              string
}

// RecordQuality inserts a new quality_results row.
func (r *Repo) RecordQuality(ctx context.Context, p RecordQualityParams) (*models.QualityResult, error) {
	row, err := r.q.RecordQuality(ctx, eocatalogdb.RecordQualityParams{
		ID:                 mapper.PgUUID(p.ID),
		ItemID:             mapper.PgUUID(p.ItemID),
		CloudCover:         p.CloudCover,
		RadiometricRmse:    p.RadiometricRMSE,
		GeometricAccuracyM: p.GeometricAccuracyM,
		Notes:              p.Notes,
	})
	if err != nil {
		return nil, err
	}
	return mapper.QualityFromRow(row), nil
}
