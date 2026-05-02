// Package handler wires ConnectRPC RPCs to the gi-tiles service.
package handler

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	pbti "github.com/ppusapati/space/services/gi-tiles/api"
	"github.com/ppusapati/space/services/gi-tiles/api/gitilesv1connect"
	"github.com/ppusapati/space/services/gi-tiles/internal/mapper"
	"github.com/ppusapati/space/services/gi-tiles/internal/models"
	"github.com/ppusapati/space/services/gi-tiles/internal/services"
)

type TilesHandler struct {
	gitilesv1connect.UnimplementedTilesServiceHandler
	svc       *services.Tiles
	validator protovalidate.Validator
}

func NewTilesHandler(svc *services.Tiles) (*TilesHandler, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &TilesHandler{svc: svc, validator: v}, nil
}

func (h *TilesHandler) validate(msg proto.Message) error {
	if err := h.validator.Validate(msg); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	return nil
}

func parseULID(s string) (ulid.ID, error) {
	id, err := ulid.Parse(s)
	if err != nil {
		return ulid.Zero, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return id, nil
}

func toConnect(err error) error {
	if err == nil {
		return nil
	}
	switch pkgerrors.Code(err) {
	case 400:
		return connect.NewError(connect.CodeInvalidArgument, err)
	case 401:
		return connect.NewError(connect.CodeUnauthenticated, err)
	case 403:
		return connect.NewError(connect.CodePermissionDenied, err)
	case 404:
		return connect.NewError(connect.CodeNotFound, err)
	case 409:
		return connect.NewError(connect.CodeAlreadyExists, err)
	case 412:
		return connect.NewError(connect.CodeFailedPrecondition, err)
	case 429:
		return connect.NewError(connect.CodeResourceExhausted, err)
	case 503:
		return connect.NewError(connect.CodeUnavailable, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}

func (h *TilesHandler) CreateTileSet(
	ctx context.Context, req *connect.Request[pbti.CreateTileSetRequest],
) (*connect.Response[pbti.CreateTileSetResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	t, err := h.svc.CreateTileSet(ctx, services.CreateTileSetInput{
		TenantID:    tid,
		Slug:        req.Msg.GetSlug(),
		Name:        req.Msg.GetName(),
		Description: req.Msg.GetDescription(),
		Format:      models.TileFormat(req.Msg.GetFormat()),
		Projection:  req.Msg.GetProjection(),
		MinZoom:     req.Msg.GetMinZoom(),
		MaxZoom:     req.Msg.GetMaxZoom(),
		SourceURI:   req.Msg.GetSourceUri(),
		Attribution: req.Msg.GetAttribution(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbti.CreateTileSetResponse{TileSet: mapper.TileSetToProto(t)}), nil
}

func (h *TilesHandler) GetTileSet(
	ctx context.Context, req *connect.Request[pbti.GetTileSetRequest],
) (*connect.Response[pbti.GetTileSetResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	t, err := h.svc.GetTileSet(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbti.GetTileSetResponse{TileSet: mapper.TileSetToProto(t)}), nil
}

func (h *TilesHandler) ListTileSets(
	ctx context.Context, req *connect.Request[pbti.ListTileSetsRequest],
) (*connect.Response[pbti.ListTileSetsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListTileSetsInput{TenantID: tid, PageOffset: offset, PageSize: size}
	if f := req.Msg.Format; f != nil {
		v := models.TileFormat(*f)
		in.Format = &v
	}
	rows, page, err := h.svc.ListTileSetsForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbti.ListTileSetsResponse{Page: mapper.PageResponse(page)}
	for _, t := range rows {
		resp.TileSets = append(resp.TileSets, mapper.TileSetToProto(t))
	}
	return connect.NewResponse(resp), nil
}

func (h *TilesHandler) DeprecateTileSet(
	ctx context.Context, req *connect.Request[pbti.DeprecateTileSetRequest],
) (*connect.Response[pbti.DeprecateTileSetResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	t, err := h.svc.DeprecateTileSet(ctx, id, "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbti.DeprecateTileSetResponse{TileSet: mapper.TileSetToProto(t)}), nil
}
