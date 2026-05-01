// Package models holds eo-catalog domain types.
package models

import (
	"time"

	"p9e.in/samavaya/packages/ulid"
)

// BoundingBox is an axis-aligned geographic bounding box (degrees).
type BoundingBox struct {
	LonMin float64
	LatMin float64
	LonMax float64
	LatMax float64
	Valid  bool
}

// Collection is a STAC collection — a logical group of items.
type Collection struct {
	ID            ulid.ID
	TenantID      ulid.ID
	Slug          string
	Title         string
	Description   string
	License       string
	SpatialExtent BoundingBox
	TemporalStart time.Time
	TemporalEnd   time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CreatedBy     string
	UpdatedBy     string
}

// Asset is one file inside an Item.
type Asset struct {
	ID        ulid.ID
	ItemID    ulid.ID
	Key       string
	Href      string
	MediaType string
	Title     string
	Roles     []string
}

// Item is a STAC item — a single scene.
type Item struct {
	ID              ulid.ID
	TenantID        ulid.ID
	CollectionID    ulid.ID
	Mission         string
	Platform        string
	Instrument      string
	Datetime        time.Time
	BBox            BoundingBox
	GeometryGeoJSON string
	CloudCover      float64
	PropertiesJSON  string
	Assets          []Asset
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedBy       string
	UpdatedBy       string
}

// QualityResult is one QA assessment of an item.
type QualityResult struct {
	ID                 ulid.ID
	ItemID             ulid.ID
	CloudCover         float64
	RadiometricRMSE    float64
	GeometricAccuracyM float64
	Notes              string
	ComputedAt         time.Time
}

// Page describes a server-side view of a paginated query.
type Page struct {
	TotalCount int32
	PageOffset int32
	PageSize   int32
	HasNext    bool
}
