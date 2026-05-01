// Package models holds gi-tiles domain types.
package models

import (
	"time"

	"p9e.in/samavaya/packages/ulid"
)

// TileFormat mirrors gitilesv1.TileFormat.
type TileFormat int32

const (
	FormatUnspecified TileFormat = 0
	FormatPNG         TileFormat = 1
	FormatWebP        TileFormat = 2
	FormatJPEG        TileFormat = 3
	FormatMVT         TileFormat = 4
	FormatCOG         TileFormat = 5
	FormatGeoJSON     TileFormat = 6
)

// TileSet is one tile-set record.
type TileSet struct {
	ID          ulid.ID
	TenantID    ulid.ID
	Slug        string
	Name        string
	Description string
	Format      TileFormat
	Projection  string
	MinZoom     int32
	MaxZoom     int32
	SourceURI   string
	Attribution string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   string
	UpdatedBy   string
}

// Page describes server-side pagination state.
type Page struct {
	TotalCount int32
	PageOffset int32
	PageSize   int32
	HasNext    bool
}
