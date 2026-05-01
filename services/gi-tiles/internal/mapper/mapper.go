// Package mapper converts proto / domain / sqlc types for gi-tiles.
package mapper

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"p9e.in/samavaya/packages/api/v1/fields"
	"p9e.in/samavaya/packages/api/v1/pagination"
	"p9e.in/samavaya/packages/ulid"

	pbti "github.com/ppusapati/space/services/gi-tiles/api"
	gitidb "github.com/ppusapati/space/services/gi-tiles/db/generated"
	"github.com/ppusapati/space/services/gi-tiles/internal/models"
)

func PgUUID(id ulid.ID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}
func FromPgUUID(p pgtype.UUID) ulid.ID {
	if !p.Valid {
		return ulid.Zero
	}
	return ulid.ID(p.Bytes)
}

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

func PageResponse(p models.Page) *pagination.PaginationResponse {
	return &pagination.PaginationResponse{
		TotalCount: p.TotalCount,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
		HasNext:    p.HasNext,
	}
}

func FieldsToProto(id ulid.ID, createdBy string, createdAt time.Time, updatedBy string, updatedAt time.Time) *fields.Fields {
	out := &fields.Fields{Uuid: id.String(), IsActive: true}
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

func TileSetFromRow(row gitidb.TileSet) *models.TileSet {
	return &models.TileSet{
		ID:          FromPgUUID(row.ID),
		TenantID:    FromPgUUID(row.TenantID),
		Slug:        row.Slug,
		Name:        row.Name,
		Description: row.Description,
		Format:      models.TileFormat(row.Format),
		Projection:  row.Projection,
		MinZoom:     row.MinZoom,
		MaxZoom:     row.MaxZoom,
		SourceURI:   row.SourceUri,
		Attribution: row.Attribution,
		Active:      row.Active,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
		CreatedBy:   row.CreatedBy,
		UpdatedBy:   row.UpdatedBy,
	}
}

func TileSetToProto(t *models.TileSet) *pbti.TileSet {
	if t == nil {
		return nil
	}
	return &pbti.TileSet{
		Id:          t.ID.String(),
		TenantId:    t.TenantID.String(),
		Slug:        t.Slug,
		Name:        t.Name,
		Description: t.Description,
		Format:      pbti.TileFormat(t.Format),
		Projection:  t.Projection,
		MinZoom:     t.MinZoom,
		MaxZoom:     t.MaxZoom,
		SourceUri:   t.SourceURI,
		Attribution: t.Attribution,
		Active:      t.Active,
		Fields:      FieldsToProto(t.ID, t.CreatedBy, t.CreatedAt, t.UpdatedBy, t.UpdatedAt),
	}
}
