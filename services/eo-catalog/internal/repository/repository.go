// Package repository wraps the sqlc-generated layer with domain-typed
// methods. The four repositories are colocated because they share the
// same pgxpool and an unscoped Repository facade simplifies wiring.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	eocatalogdb "github.com/ppusapati/space/services/eo-catalog/db/generated"
	"github.com/ppusapati/space/services/eo-catalog/internal/mappers"
	"github.com/ppusapati/space/services/eo-catalog/internal/models"
)

// ErrNotFound is returned when a row does not exist.
var ErrNotFound = errors.New("repository: not found")

// Repository aggregates all four entity repositories.
type Repository struct {
	Collections *CollectionRepository
	Items       *ItemRepository
	Assets      *AssetRepository
	Quality     *QualityRepository
}

// New constructs a Repository against the given pgxpool.
func New(pool *pgxpool.Pool) *Repository {
	q := eocatalogdb.New(pool)
	return &Repository{
		Collections: &CollectionRepository{q: q, pool: pool},
		Items:       &ItemRepository{q: q, pool: pool},
		Assets:      &AssetRepository{q: q, pool: pool},
		Quality:     &QualityRepository{q: q, pool: pool},
	}
}

// ----- Collections ---------------------------------------------------

// CollectionRepository persists Collections.
type CollectionRepository struct {
	q    *eocatalogdb.Queries
	pool *pgxpool.Pool
}

// Create inserts a new Collection.
func (r *CollectionRepository) Create(ctx context.Context, c *models.Collection) (*models.Collection, error) {
	row, err := r.q.CreateCollection(ctx, eocatalogdb.CreateCollectionParams{
		ID:            mappers.PgUUID(c.ID),
		TenantID:      mappers.PgUUID(c.TenantID),
		Slug:          c.Slug,
		Title:         c.Title,
		Description:   c.Description,
		License:       c.License,
		BboxLonMin:    mappers.PtrFloat64(c.SpatialExtent.Valid, c.SpatialExtent.LonMin),
		BboxLatMin:    mappers.PtrFloat64(c.SpatialExtent.Valid, c.SpatialExtent.LatMin),
		BboxLonMax:    mappers.PtrFloat64(c.SpatialExtent.Valid, c.SpatialExtent.LonMax),
		BboxLatMax:    mappers.PtrFloat64(c.SpatialExtent.Valid, c.SpatialExtent.LatMax),
		TemporalStart: mappers.PgTimestampPtr(c.TemporalStart),
		TemporalEnd:   mappers.PgTimestampPtr(c.TemporalEnd),
		CreatedBy:     c.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mappers.CollectionFromRow(row), nil
}

// Get fetches by id.
func (r *CollectionRepository) Get(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
	row, err := r.q.GetCollection(ctx, mappers.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mappers.CollectionFromRow(row), nil
}

// ListAfter returns collections older than the cursor in (created_at, id) DESC order.
func (r *CollectionRepository) ListAfter(
	ctx context.Context, tenantID uuid.UUID, cursorTS *time.Time, cursorID uuid.UUID, limit int32,
) ([]*models.Collection, error) {
	rows, err := r.q.ListCollectionsForTenant(ctx, eocatalogdb.ListCollectionsForTenantParams{
		TenantID:        mappers.PgUUID(tenantID),
		CursorCreatedAt: mappers.PgTimestampPtr(cursorTS),
		CursorID:        mappers.PgUUID(cursorID),
		Lim:             limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*models.Collection, 0, len(rows))
	for _, row := range rows {
		out = append(out, mappers.CollectionFromRow(row))
	}
	return out, nil
}

// ----- Items ---------------------------------------------------------

// ItemRepository persists Items and (transactionally) their Assets.
type ItemRepository struct {
	q    *eocatalogdb.Queries
	pool *pgxpool.Pool
}

// CreateWithAssets inserts an Item plus its initial Assets in a single
// transaction.
func (r *ItemRepository) CreateWithAssets(
	ctx context.Context, it *models.Item,
) (*models.Item, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.q.WithTx(tx)
	row, err := q.CreateItem(ctx, eocatalogdb.CreateItemParams{
		ID:              mappers.PgUUID(it.ID),
		TenantID:        mappers.PgUUID(it.TenantID),
		CollectionID:    mappers.PgUUID(it.CollectionID),
		Mission:         it.Mission,
		Platform:        it.Platform,
		Instrument:      it.Instrument,
		Datetime:        mappers.PgTimestamp(it.Datetime),
		BboxLonMin:      mappers.PtrFloat64(it.BBox.Valid, it.BBox.LonMin),
		BboxLatMin:      mappers.PtrFloat64(it.BBox.Valid, it.BBox.LatMin),
		BboxLonMax:      mappers.PtrFloat64(it.BBox.Valid, it.BBox.LonMax),
		BboxLatMax:      mappers.PtrFloat64(it.BBox.Valid, it.BBox.LatMax),
		GeometryGeojson: it.GeometryGeoJSON,
		CloudCover:      it.CloudCover,
		PropertiesJson:  []byte(it.PropertiesJSON),
		CreatedBy:       it.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	created := mappers.ItemFromRow(row)
	for _, a := range it.Assets {
		assetRow, err := q.CreateAsset(ctx, eocatalogdb.CreateAssetParams{
			ID:        mappers.PgUUID(uuid.New()),
			ItemID:    mappers.PgUUID(created.ID),
			Key:       a.Key,
			Href:      a.Href,
			MediaType: a.MediaType,
			Title:     a.Title,
			Roles:     a.Roles,
		})
		if err != nil {
			return nil, err
		}
		created.Assets = append(created.Assets, mappers.AssetFromRow(assetRow))
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return created, nil
}

// Get fetches an Item plus its Assets.
func (r *ItemRepository) Get(ctx context.Context, id uuid.UUID) (*models.Item, error) {
	row, err := r.q.GetItem(ctx, mappers.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	it := mappers.ItemFromRow(row)
	assets, err := r.q.ListAssetsForItem(ctx, mappers.PgUUID(id))
	if err != nil {
		return nil, err
	}
	for _, a := range assets {
		it.Assets = append(it.Assets, mappers.AssetFromRow(a))
	}
	return it, nil
}

// SearchParams groups SearchItems inputs.
type SearchParams struct {
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

// Search runs the bounded-bbox + datetime + cloud-cover scan.
func (r *ItemRepository) Search(ctx context.Context, p SearchParams) ([]*models.Item, error) {
	var collectionID pgtype.UUID
	if p.CollectionID != nil {
		collectionID = mappers.PgUUID(*p.CollectionID)
	}
	var bboxLonMin, bboxLonMax, bboxLatMin, bboxLatMax *float64
	if p.BBox != nil && p.BBox.Valid {
		v := *p.BBox
		bboxLonMin, bboxLonMax = &v.LonMin, &v.LonMax
		bboxLatMin, bboxLatMax = &v.LatMin, &v.LatMax
	}
	rows, err := r.q.SearchItems(ctx, eocatalogdb.SearchItemsParams{
		TenantID:       mappers.PgUUID(p.TenantID),
		CollectionID:   collectionID,
		DatetimeStart:  mappers.PgTimestamp(p.DatetimeStart),
		DatetimeEnd:    mappers.PgTimestamp(p.DatetimeEnd),
		MaxCloudCover:  p.MaxCloudCover,
		BboxLonMin:     bboxLonMin,
		BboxLonMax:     bboxLonMax,
		BboxLatMin:     bboxLatMin,
		BboxLatMax:     bboxLatMax,
		CursorDatetime: mappers.PgTimestampPtr(p.CursorDatetime),
		CursorID:       mappers.PgUUID(p.CursorID),
		Lim:            p.Limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*models.Item, 0, len(rows))
	for _, row := range rows {
		out = append(out, mappers.ItemFromRow(row))
	}
	return out, nil
}

// ----- Assets --------------------------------------------------------

// AssetRepository persists Assets.
type AssetRepository struct {
	q    *eocatalogdb.Queries
	pool *pgxpool.Pool
}

// ListForItem returns the Assets attached to an Item ordered by key.
func (r *AssetRepository) ListForItem(ctx context.Context, itemID uuid.UUID) ([]models.Asset, error) {
	rows, err := r.q.ListAssetsForItem(ctx, mappers.PgUUID(itemID))
	if err != nil {
		return nil, err
	}
	out := make([]models.Asset, 0, len(rows))
	for _, row := range rows {
		out = append(out, mappers.AssetFromRow(row))
	}
	return out, nil
}

// ----- Quality -------------------------------------------------------

// QualityRepository persists QualityResult rows.
type QualityRepository struct {
	q    *eocatalogdb.Queries
	pool *pgxpool.Pool
}

// Record inserts a new QualityResult.
func (r *QualityRepository) Record(ctx context.Context, q *models.QualityResult) (*models.QualityResult, error) {
	row, err := r.q.RecordQuality(ctx, eocatalogdb.RecordQualityParams{
		ID:                 mappers.PgUUID(q.ID),
		ItemID:             mappers.PgUUID(q.ItemID),
		CloudCover:         q.CloudCover,
		RadiometricRmse:    q.RadiometricRMSE,
		GeometricAccuracyM: q.GeometricAccuracyM,
		Notes:              q.Notes,
	})
	if err != nil {
		return nil, err
	}
	return mappers.QualityResultFromRow(row), nil
}

// ListForItem returns QualityResults ordered by computed_at DESC.
func (r *QualityRepository) ListForItem(
	ctx context.Context, itemID uuid.UUID, cursorTS *time.Time, cursorID uuid.UUID, limit int32,
) ([]*models.QualityResult, error) {
	rows, err := r.q.ListQualityForItem(ctx, eocatalogdb.ListQualityForItemParams{
		ItemID:           mappers.PgUUID(itemID),
		CursorComputedAt: mappers.PgTimestampPtr(cursorTS),
		CursorID:         mappers.PgUUID(cursorID),
		Lim:              limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*models.QualityResult, 0, len(rows))
	for _, row := range rows {
		out = append(out, mappers.QualityResultFromRow(row))
	}
	return out, nil
}
