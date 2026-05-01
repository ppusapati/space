// Package mappers converts between proto, domain (models), and sqlc
// generated types. Each mapper is pure: no I/O, no logging, just shape
// conversion.
package mappers

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	eov1 "github.com/ppusapati/space/api/p9e/space/earthobs/v1"
	commonv1 "github.com/ppusapati/space/api/p9e/space/common/v1"
	eocatalogdb "github.com/ppusapati/space/services/eo-catalog/db/generated"
	"github.com/ppusapati/space/services/eo-catalog/internal/models"
)

// ---------- low-level helpers ----------------------------------------

// PgUUID converts a uuid.UUID to pgtype.UUID with Valid=true.
func PgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// FromPgUUID returns the uuid.UUID portion of a pgtype.UUID. Invalid
// values become uuid.Nil.
func FromPgUUID(p pgtype.UUID) uuid.UUID {
	if !p.Valid {
		return uuid.Nil
	}
	return p.Bytes
}

// PgTimestampPtr converts a *time.Time to pgtype.Timestamptz.
func PgTimestampPtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// PgTimestamp converts a time.Time to a Valid pgtype.Timestamptz.
func PgTimestamp(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// FromPgTimestampPtr returns a pointer to a time.Time, or nil when invalid.
func FromPgTimestampPtr(p pgtype.Timestamptz) *time.Time {
	if !p.Valid {
		return nil
	}
	t := p.Time
	return &t
}

// PtrFloat64 returns a non-nil pointer to v when set, else nil.
func PtrFloat64(set bool, v float64) *float64 {
	if !set {
		return nil
	}
	return &v
}

// FromPtrFloat64 returns *p or 0; the second return is the validity.
func FromPtrFloat64(p *float64) (float64, bool) {
	if p == nil {
		return 0, false
	}
	return *p, true
}

// ---------- BoundingBox -----------------------------------------------

// BBoxFromProto converts the proto bbox to domain. nil ⇒ zero value.
func BBoxFromProto(p *eov1.BoundingBox) models.BoundingBox {
	if p == nil {
		return models.BoundingBox{}
	}
	return models.BoundingBox{
		LonMin: p.LonMin, LatMin: p.LatMin,
		LonMax: p.LonMax, LatMax: p.LatMax,
		Valid: true,
	}
}

// BBoxToProto converts domain → proto.
func BBoxToProto(b models.BoundingBox) *eov1.BoundingBox {
	if !b.Valid {
		return nil
	}
	return &eov1.BoundingBox{
		LonMin: b.LonMin, LatMin: b.LatMin,
		LonMax: b.LonMax, LatMax: b.LatMax,
	}
}

// ---------- AuditFields -----------------------------------------------

// AuditToProto packs the audit timestamps and principals into the
// shared common.v1.AuditFields message.
func AuditToProto(createdAt, updatedAt time.Time, createdBy, updatedBy string) *commonv1.AuditFields {
	return &commonv1.AuditFields{
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
		CreatedBy: createdBy,
		UpdatedBy: updatedBy,
	}
}

// ---------- Collection ------------------------------------------------

// CollectionFromRow converts the sqlc row to a domain Collection.
func CollectionFromRow(row eocatalogdb.Collection) *models.Collection {
	return &models.Collection{
		ID:       FromPgUUID(row.ID),
		TenantID: FromPgUUID(row.TenantID),
		Slug:     row.Slug,
		Title:    row.Title,
		Description: row.Description,
		License:     row.License,
		SpatialExtent: models.BoundingBox{
			LonMin: zero(row.BboxLonMin), LatMin: zero(row.BboxLatMin),
			LonMax: zero(row.BboxLonMax), LatMax: zero(row.BboxLatMax),
			Valid: row.BboxLonMin != nil,
		},
		TemporalStart: FromPgTimestampPtr(row.TemporalStart),
		TemporalEnd:   FromPgTimestampPtr(row.TemporalEnd),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		CreatedBy:     row.CreatedBy,
		UpdatedBy:     row.UpdatedBy,
	}
}

// CollectionToProto converts a domain Collection to the proto type.
func CollectionToProto(c *models.Collection) *eov1.Collection {
	if c == nil {
		return nil
	}
	out := &eov1.Collection{
		Id:          c.ID.String(),
		TenantId:    c.TenantID.String(),
		Slug:        c.Slug,
		Title:       c.Title,
		Description: c.Description,
		License:     c.License,
		Audit:       AuditToProto(c.CreatedAt, c.UpdatedAt, c.CreatedBy, c.UpdatedBy),
	}
	if c.SpatialExtent.Valid {
		out.SpatialExtent = BBoxToProto(c.SpatialExtent)
	}
	if c.TemporalStart != nil {
		out.TemporalStart = timestamppb.New(*c.TemporalStart)
	}
	if c.TemporalEnd != nil {
		out.TemporalEnd = timestamppb.New(*c.TemporalEnd)
	}
	return out
}

// ---------- Item ------------------------------------------------------

// ItemFromRow converts the sqlc row to a domain Item (without assets).
func ItemFromRow(row eocatalogdb.Item) *models.Item {
	return &models.Item{
		ID:           FromPgUUID(row.ID),
		TenantID:     FromPgUUID(row.TenantID),
		CollectionID: FromPgUUID(row.CollectionID),
		Mission:      row.Mission,
		Platform:     row.Platform,
		Instrument:   row.Instrument,
		Datetime:     row.Datetime.Time,
		BBox: models.BoundingBox{
			LonMin: zero(row.BboxLonMin), LatMin: zero(row.BboxLatMin),
			LonMax: zero(row.BboxLonMax), LatMax: zero(row.BboxLatMax),
			Valid: row.BboxLonMin != nil,
		},
		GeometryGeoJSON: row.GeometryGeojson,
		CloudCover:      row.CloudCover,
		PropertiesJSON:  string(row.PropertiesJson),
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		CreatedBy:       row.CreatedBy,
		UpdatedBy:       row.UpdatedBy,
	}
}

// ItemToProto converts a domain Item to the proto type, including assets.
func ItemToProto(it *models.Item) *eov1.Item {
	if it == nil {
		return nil
	}
	out := &eov1.Item{
		Id:              it.ID.String(),
		TenantId:        it.TenantID.String(),
		CollectionId:    it.CollectionID.String(),
		Mission:         it.Mission,
		Platform:        it.Platform,
		Instrument:      it.Instrument,
		Datetime:        timestamppb.New(it.Datetime),
		Bbox:            BBoxToProto(it.BBox),
		GeometryGeojson: it.GeometryGeoJSON,
		CloudCover:      it.CloudCover,
		PropertiesJson:  it.PropertiesJSON,
		Audit:           AuditToProto(it.CreatedAt, it.UpdatedAt, it.CreatedBy, it.UpdatedBy),
	}
	for _, a := range it.Assets {
		out.Assets = append(out.Assets, AssetToProto(a))
	}
	return out
}

// ---------- Asset -----------------------------------------------------

// AssetFromRow converts a sqlc Asset row to a domain Asset.
func AssetFromRow(row eocatalogdb.Asset) models.Asset {
	return models.Asset{
		ID:        FromPgUUID(row.ID),
		ItemID:    FromPgUUID(row.ItemID),
		Key:       row.Key,
		Href:      row.Href,
		MediaType: row.MediaType,
		Title:     row.Title,
		Roles:     append([]string(nil), row.Roles...),
	}
}

// AssetFromProto converts the proto Asset to a domain value.
func AssetFromProto(p *eov1.Asset) models.Asset {
	if p == nil {
		return models.Asset{}
	}
	return models.Asset{
		Key:       p.Key,
		Href:      p.Href,
		MediaType: p.MediaType,
		Title:     p.Title,
		Roles:     append([]string(nil), p.Roles...),
	}
}

// AssetToProto converts a domain Asset to the proto type.
func AssetToProto(a models.Asset) *eov1.Asset {
	return &eov1.Asset{
		Key:       a.Key,
		Href:      a.Href,
		MediaType: a.MediaType,
		Title:     a.Title,
		Roles:     append([]string(nil), a.Roles...),
	}
}

// ---------- QualityResult ---------------------------------------------

// QualityResultFromRow converts the sqlc row to a domain value.
func QualityResultFromRow(row eocatalogdb.QualityResult) *models.QualityResult {
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

// QualityResultToProto converts domain → proto.
func QualityResultToProto(q *models.QualityResult) *eov1.QualityResult {
	if q == nil {
		return nil
	}
	return &eov1.QualityResult{
		Id:                 q.ID.String(),
		ItemId:             q.ItemID.String(),
		CloudCover:         q.CloudCover,
		RadiometricRmse:    q.RadiometricRMSE,
		GeometricAccuracyM: q.GeometricAccuracyM,
		Notes:              q.Notes,
		ComputedAt:         timestamppb.New(q.ComputedAt),
	}
}

func zero(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
