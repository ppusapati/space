// Package repository wraps the gi-tiles sqlc layer.
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/samavaya/packages/ulid"

	gitidb "github.com/ppusapati/space/services/gi-tiles/db/generated"
	"github.com/ppusapati/space/services/gi-tiles/internal/mapper"
	"github.com/ppusapati/space/services/gi-tiles/internal/models"
)

var ErrNotFound = errors.New("repository: not found")

type Repo struct {
	q    *gitidb.Queries
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: gitidb.New(pool), pool: pool}
}

type CreateTileSetParams struct {
	ID          ulid.ID
	TenantID    ulid.ID
	Slug        string
	Name        string
	Description string
	Format      models.TileFormat
	Projection  string
	MinZoom     int32
	MaxZoom     int32
	SourceURI   string
	Attribution string
	CreatedBy   string
}

func (r *Repo) CreateTileSet(ctx context.Context, p CreateTileSetParams) (*models.TileSet, error) {
	row, err := r.q.CreateTileSet(ctx, gitidb.CreateTileSetParams{
		ID:          mapper.PgUUID(p.ID),
		TenantID:    mapper.PgUUID(p.TenantID),
		Slug:        p.Slug,
		Name:        p.Name,
		Description: p.Description,
		Format:      int32(p.Format),
		Projection:  p.Projection,
		MinZoom:     p.MinZoom,
		MaxZoom:     p.MaxZoom,
		SourceUri:   p.SourceURI,
		Attribution: p.Attribution,
		CreatedBy:   p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.TileSetFromRow(row), nil
}

func (r *Repo) GetTileSet(ctx context.Context, id ulid.ID) (*models.TileSet, error) {
	row, err := r.q.GetTileSet(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.TileSetFromRow(row), nil
}

type ListTileSetsParams struct {
	TenantID   ulid.ID
	Format     *models.TileFormat
	PageOffset int32
	PageSize   int32
}

func (r *Repo) ListTileSetsForTenant(ctx context.Context, p ListTileSetsParams) ([]*models.TileSet, int32, error) {
	var formatPtr *int32
	if p.Format != nil {
		v := int32(*p.Format)
		formatPtr = &v
	}
	total, err := r.q.CountTileSetsForTenant(ctx, gitidb.CountTileSetsForTenantParams{
		TenantID: mapper.PgUUID(p.TenantID),
		Format:   formatPtr,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListTileSetsForTenant(ctx, gitidb.ListTileSetsForTenantParams{
		TenantID:   mapper.PgUUID(p.TenantID),
		Format:     formatPtr,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.TileSet, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.TileSetFromRow(row))
	}
	return out, int32(total), nil
}

func (r *Repo) DeprecateTileSet(ctx context.Context, id ulid.ID, updatedBy string) (*models.TileSet, error) {
	row, err := r.q.DeprecateTileSet(ctx, gitidb.DeprecateTileSetParams{
		ID:        mapper.PgUUID(id),
		UpdatedBy: updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.TileSetFromRow(row), nil
}
