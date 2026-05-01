// Package models holds the eo-catalog domain types — pure Go structs
// independent of protobuf and sqlc. Mappers convert between layers.
package models

import (
	"time"

	"github.com/google/uuid"
)

// BoundingBox is `(lon_min, lat_min, lon_max, lat_max)` in WGS84
// degrees. The zero value represents "no bbox".
type BoundingBox struct {
	LonMin, LatMin, LonMax, LatMax float64
	Valid                          bool
}

// Collection groups a set of Items.
type Collection struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	Slug          string
	Title         string
	Description   string
	License       string
	SpatialExtent BoundingBox
	TemporalStart *time.Time
	TemporalEnd   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CreatedBy     string
	UpdatedBy     string
}

// Asset is one file inside an Item — typically a per-band COG.
type Asset struct {
	ID        uuid.UUID
	ItemID    uuid.UUID
	Key       string
	Href      string
	MediaType string
	Title     string
	Roles     []string
}

// Item is one scene.
type Item struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	CollectionID    uuid.UUID
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

// QualityResult is a per-item QA record produced by the pipeline.
type QualityResult struct {
	ID                 uuid.UUID
	ItemID             uuid.UUID
	CloudCover         float64
	RadiometricRMSE    float64
	GeometricAccuracyM float64
	Notes              string
	ComputedAt         time.Time
}
