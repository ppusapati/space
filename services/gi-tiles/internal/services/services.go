// Package services holds gi-tiles business logic.
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	"github.com/ppusapati/space/services/gi-tiles/internal/models"
	"github.com/ppusapati/space/services/gi-tiles/internal/repository"
)

type Tiles struct {
	repo  *repository.Repo
	IDFn  func() ulid.ID
	NowFn func() time.Time
}

func New(repo *repository.Repo) *Tiles {
	return &Tiles{
		repo:  repo,
		IDFn:  ulid.NewMonotonic,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

type CreateTileSetInput struct {
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

func (t *Tiles) CreateTileSet(ctx context.Context, in CreateTileSetInput) (*models.TileSet, error) {
	if in.TenantID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	in.Slug = strings.TrimSpace(in.Slug)
	in.Name = strings.TrimSpace(in.Name)
	if in.Slug == "" || in.Name == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "slug and name required")
	}
	if in.Format == models.FormatUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "format required")
	}
	if in.MinZoom < 0 || in.MaxZoom < 0 || in.MaxZoom < in.MinZoom {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "valid min_zoom <= max_zoom required")
	}
	in.Projection = strings.TrimSpace(in.Projection)
	if in.Projection == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "projection required")
	}
	if strings.TrimSpace(in.SourceURI) == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "source_uri required")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return t.repo.CreateTileSet(ctx, repository.CreateTileSetParams{
		ID:          t.IDFn(),
		TenantID:    in.TenantID,
		Slug:        in.Slug,
		Name:        in.Name,
		Description: in.Description,
		Format:      in.Format,
		Projection:  in.Projection,
		MinZoom:     in.MinZoom,
		MaxZoom:     in.MaxZoom,
		SourceURI:   in.SourceURI,
		Attribution: in.Attribution,
		CreatedBy:   createdBy,
	})
}

func (t *Tiles) GetTileSet(ctx context.Context, id ulid.ID) (*models.TileSet, error) {
	ts, err := t.repo.GetTileSet(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("TILE_SET_NOT_FOUND", "tile_set "+id.String())
	}
	return ts, err
}

type ListTileSetsInput struct {
	TenantID   ulid.ID
	Format     *models.TileFormat
	PageOffset int32
	PageSize   int32
}

func (t *Tiles) ListTileSetsForTenant(ctx context.Context, in ListTileSetsInput) ([]*models.TileSet, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := t.repo.ListTileSetsForTenant(ctx, repository.ListTileSetsParams{
		TenantID:   in.TenantID,
		Format:     in.Format,
		PageOffset: in.PageOffset,
		PageSize:   in.PageSize,
	})
	if err != nil {
		return nil, models.Page{}, err
	}
	return rows, models.Page{
		TotalCount: total,
		PageOffset: in.PageOffset,
		PageSize:   in.PageSize,
		HasNext:    in.PageOffset+int32(len(rows)) < total,
	}, nil
}

func (t *Tiles) DeprecateTileSet(ctx context.Context, id ulid.ID, updatedBy string) (*models.TileSet, error) {
	if updatedBy == "" {
		updatedBy = "system"
	}
	ts, err := t.repo.DeprecateTileSet(ctx, id, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("TILE_SET_NOT_FOUND", "tile_set "+id.String())
	}
	return ts, err
}
