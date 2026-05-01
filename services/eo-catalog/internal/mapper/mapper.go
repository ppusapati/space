// Package mapper converts proto / domain / sqlc types for eo-catalog.
package mapper

import (
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"p9e.in/samavaya/packages/api/v1/fields"
	"p9e.in/samavaya/packages/api/v1/pagination"
	"p9e.in/samavaya/packages/ulid"

	pbcat "github.com/ppusapati/space/services/eo-catalog/api"
	eocatalogdb "github.com/ppusapati/space/services/eo-catalog/db/generated"
	"github.com/ppusapati/space/services/eo-catalog/internal/models"
)

// PgUUID converts a ulid.ID into a pgtype.UUID payload (16 bytes).
func PgUUID(id ulid.ID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

// PgUUIDOrNull returns NULL when id is the zero ULID.
func PgUUIDOrNull(id ulid.ID) pgtype.UUID {
	if id.IsZero() {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

// FromPgUUID extracts the ulid.ID payload from a pgtype.UUID. Returns
// ulid.Zero when the column was NULL.
func FromPgUUID(p pgtype.UUID) ulid.ID {
	if !p.Valid {
		return ulid.Zero
	}
	return ulid.ID(p.Bytes)
}

// PgTimestamp converts time.Time to pgtype.Timestamptz.
func PgTimestamp(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// PgTimestampOrNull returns NULL when t is the zero value.
func PgTimestampOrNull(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// FloatPtr is the canonical helper for nilable float64 columns.
func FloatPtr(v float64, present bool) *float64 {
	if !present {
		return nil
	}
	return &v
}

// DerefFloat returns *p, or 0 when p is nil.
func DerefFloat(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

// ----- Pagination ----------------------------------------------------------

// PageRequest extracts (offset, size) from a PaginationRequest, applying
// defaults: page_size defaults to 50, capped at 500; offset defaults to 0.
func PageRequest(p *pagination.PaginationRequest) (offset, size int32) {
	if p == nil {
		return 0, 50
	}
	offset = p.GetPageOffset()
	if offset < 0 {
		offset = 0
	}
	size = p.GetPageSize()
	switch {
	case size <= 0:
		size = 50
	case size > 500:
		size = 500
	}
	return offset, size
}

// PageResponse builds a PaginationResponse from a Page.
func PageResponse(p models.Page) *pagination.PaginationResponse {
	return &pagination.PaginationResponse{
		TotalCount: p.TotalCount,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
		HasNext:    p.HasNext,
	}
}

// ----- Bounding box --------------------------------------------------------

// BBoxFromColumns builds a BoundingBox from four nilable float columns.
func BBoxFromColumns(lonMin, latMin, lonMax, latMax *float64) models.BoundingBox {
	if lonMin == nil || latMin == nil || lonMax == nil || latMax == nil {
		return models.BoundingBox{}
	}
	return models.BoundingBox{
		LonMin: *lonMin,
		LatMin: *latMin,
		LonMax: *lonMax,
		LatMax: *latMax,
		Valid:  true,
	}
}

// BBoxToProto converts a domain BoundingBox to its proto form. Returns nil
// when invalid (no extent).
func BBoxToProto(b models.BoundingBox) *pbcat.BoundingBox {
	if !b.Valid {
		return nil
	}
	return &pbcat.BoundingBox{
		LonMin: b.LonMin,
		LatMin: b.LatMin,
		LonMax: b.LonMax,
		LatMax: b.LatMax,
	}
}

// BBoxFromProto converts a proto BoundingBox to domain. nil ⇒ invalid.
func BBoxFromProto(p *pbcat.BoundingBox) models.BoundingBox {
	if p == nil {
		return models.BoundingBox{}
	}
	return models.BoundingBox{
		LonMin: p.GetLonMin(),
		LatMin: p.GetLatMin(),
		LonMax: p.GetLonMax(),
		LatMax: p.GetLatMax(),
		Valid:  true,
	}
}

// ----- Audit fields --------------------------------------------------------

// FieldsToProto builds a Fields message from a domain object's audit fields.
func FieldsToProto(id ulid.ID, createdBy string, createdAt time.Time, updatedBy string, updatedAt time.Time) *fields.Fields {
	out := &fields.Fields{
		Uuid:     id.String(),
		IsActive: true,
	}
	if createdBy != "" {
		out.CreatedBy = wrapperspb.String(createdBy)
	}
	if !createdAt.IsZero() {
		out.CreatedAt = timestamppb.New(createdAt)
	}
	if updatedBy != "" {
		out.UpdatedBy = wrapperspb.String(updatedBy)
	}
	if !updatedAt.IsZero() {
		out.UpdatedAt = timestamppb.New(updatedAt)
	}
	return out
}

// ----- Collection ----------------------------------------------------------

// CollectionFromRow converts a sqlc Collection row to the domain model.
func CollectionFromRow(row eocatalogdb.Collection) *models.Collection {
	c := &models.Collection{
		ID:            FromPgUUID(row.ID),
		TenantID:      FromPgUUID(row.TenantID),
		Slug:          row.Slug,
		Title:         row.Title,
		Description:   row.Description,
		License:       row.License,
		SpatialExtent: BBoxFromColumns(row.BboxLonMin, row.BboxLatMin, row.BboxLonMax, row.BboxLatMax),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		CreatedBy:     row.CreatedBy,
		UpdatedBy:     row.UpdatedBy,
	}
	if row.TemporalStart.Valid {
		c.TemporalStart = row.TemporalStart.Time
	}
	if row.TemporalEnd.Valid {
		c.TemporalEnd = row.TemporalEnd.Time
	}
	return c
}

// CollectionToProto converts a domain Collection to its proto form.
func CollectionToProto(c *models.Collection) *pbcat.Collection {
	if c == nil {
		return nil
	}
	out := &pbcat.Collection{
		Id:            c.ID.String(),
		TenantId:      c.TenantID.String(),
		Slug:          c.Slug,
		Title:         c.Title,
		Description:   c.Description,
		License:       c.License,
		SpatialExtent: BBoxToProto(c.SpatialExtent),
		Fields:        FieldsToProto(c.ID, c.CreatedBy, c.CreatedAt, c.UpdatedBy, c.UpdatedAt),
	}
	if !c.TemporalStart.IsZero() {
		out.TemporalStart = timestamppb.New(c.TemporalStart)
	}
	if !c.TemporalEnd.IsZero() {
		out.TemporalEnd = timestamppb.New(c.TemporalEnd)
	}
	return out
}

// ----- Item ----------------------------------------------------------------

// ItemFromRow converts a sqlc Item row to the domain model. Assets must be
// loaded separately and attached.
func ItemFromRow(row eocatalogdb.Item) *models.Item {
	return &models.Item{
		ID:              FromPgUUID(row.ID),
		TenantID:        FromPgUUID(row.TenantID),
		CollectionID:    FromPgUUID(row.CollectionID),
		Mission:         row.Mission,
		Platform:        row.Platform,
		Instrument:      row.Instrument,
		Datetime:        row.Datetime.Time,
		BBox:            BBoxFromColumns(row.BboxLonMin, row.BboxLatMin, row.BboxLonMax, row.BboxLatMax),
		GeometryGeoJSON: row.GeometryGeojson,
		CloudCover:      row.CloudCover,
		PropertiesJSON:  string(row.PropertiesJson),
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		CreatedBy:       row.CreatedBy,
		UpdatedBy:       row.UpdatedBy,
	}
}

// ItemToProto converts a domain Item to its proto form.
func ItemToProto(i *models.Item) *pbcat.Item {
	if i == nil {
		return nil
	}
	out := &pbcat.Item{
		Id:              i.ID.String(),
		TenantId:        i.TenantID.String(),
		CollectionId:    i.CollectionID.String(),
		Mission:         i.Mission,
		Platform:        i.Platform,
		Instrument:      i.Instrument,
		Datetime:        timestamppb.New(i.Datetime),
		Bbox:            BBoxToProto(i.BBox),
		GeometryGeojson: i.GeometryGeoJSON,
		CloudCover:      i.CloudCover,
		PropertiesJson:  i.PropertiesJSON,
		Fields:          FieldsToProto(i.ID, i.CreatedBy, i.CreatedAt, i.UpdatedBy, i.UpdatedAt),
	}
	for _, a := range i.Assets {
		out.Assets = append(out.Assets, AssetToProto(a))
	}
	return out
}

// ----- Asset ---------------------------------------------------------------

// AssetFromRow converts a sqlc Asset row to the domain model.
func AssetFromRow(row eocatalogdb.Asset) models.Asset {
	roles := append([]string(nil), row.Roles...)
	return models.Asset{
		ID:        FromPgUUID(row.ID),
		ItemID:    FromPgUUID(row.ItemID),
		Key:       row.Key,
		Href:      row.Href,
		MediaType: row.MediaType,
		Title:     row.Title,
		Roles:     roles,
	}
}

// AssetToProto converts a domain Asset to its proto form.
func AssetToProto(a models.Asset) *pbcat.Asset {
	return &pbcat.Asset{
		Key:       a.Key,
		Href:      a.Href,
		MediaType: a.MediaType,
		Title:     a.Title,
		Roles:     append([]string(nil), a.Roles...),
	}
}

// AssetFromProto converts a proto Asset into a domain Asset (key trimmed).
func AssetFromProto(p *pbcat.Asset) models.Asset {
	if p == nil {
		return models.Asset{}
	}
	return models.Asset{
		Key:       strings.TrimSpace(p.GetKey()),
		Href:      strings.TrimSpace(p.GetHref()),
		MediaType: p.GetMediaType(),
		Title:     p.GetTitle(),
		Roles:     append([]string(nil), p.GetRoles()...),
	}
}

// ----- QualityResult -------------------------------------------------------

// QualityFromRow converts a sqlc QualityResult row to the domain model.
func QualityFromRow(row eocatalogdb.QualityResult) *models.QualityResult {
	return &models.QualityResult{
		ID:                 FromPgUUID(row.ID),
		ItemID:             FromPgUUID(row.ItemID),
		CloudCover:         row.CloudCover,
		RadiometricRMSE:    row.RadiometricRmse,
		GeometricAccuracyM: row.GeometricAccuracyM,
		Notes:              row.Notes,
		ComputedAt:         row.ComputedAt.Time,
	}
}

// QualityToProto converts a domain QualityResult to its proto form.
func QualityToProto(q *models.QualityResult) *pbcat.QualityResult {
	if q == nil {
		return nil
	}
	return &pbcat.QualityResult{
		Id:                 q.ID.String(),
		ItemId:             q.ItemID.String(),
		CloudCover:         q.CloudCover,
		RadiometricRmse:    q.RadiometricRMSE,
		GeometricAccuracyM: q.GeometricAccuracyM,
		Notes:              q.Notes,
		ComputedAt:         timestamppb.New(q.ComputedAt),
	}
}
