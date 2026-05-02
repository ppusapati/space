// Package handler wires ConnectRPC RPCs to the eo-catalog service.
package handler

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	pbcat "github.com/ppusapati/space/services/eo-catalog/api"
	"github.com/ppusapati/space/services/eo-catalog/api/eocatalogv1connect"
	"github.com/ppusapati/space/services/eo-catalog/internal/mapper"
	"github.com/ppusapati/space/services/eo-catalog/internal/services"
)

// CatalogHandler implements eocatalogv1connect.CatalogServiceHandler.
type CatalogHandler struct {
	eocatalogv1connect.UnimplementedCatalogServiceHandler
	svc       *services.Catalog
	validator protovalidate.Validator
}

// NewCatalogHandler returns a handler.
func NewCatalogHandler(svc *services.Catalog) (*CatalogHandler, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &CatalogHandler{svc: svc, validator: v}, nil
}

// validate runs protovalidate; returns a Connect InvalidArgument error on failure.
func (h *CatalogHandler) validate(msg proto.Message) error {
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

// toConnect maps a packages/errors.Error to the closest connect error code.
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
	case 429:
		return connect.NewError(connect.CodeResourceExhausted, err)
	case 503:
		return connect.NewError(connect.CodeUnavailable, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}

// ----- Collections ---------------------------------------------------------

// CreateCollection implements the proto RPC.
func (h *CatalogHandler) CreateCollection(
	ctx context.Context, req *connect.Request[pbcat.CreateCollectionRequest],
) (*connect.Response[pbcat.CreateCollectionResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	in := services.CreateCollectionInput{
		TenantID:      tid,
		Slug:          req.Msg.GetSlug(),
		Title:         req.Msg.GetTitle(),
		Description:   req.Msg.GetDescription(),
		License:       req.Msg.GetLicense(),
		SpatialExtent: mapper.BBoxFromProto(req.Msg.GetSpatialExtent()),
	}
	if t := req.Msg.GetTemporalStart(); t != nil {
		in.TemporalStart = t.AsTime()
	}
	if t := req.Msg.GetTemporalEnd(); t != nil {
		in.TemporalEnd = t.AsTime()
	}
	col, err := h.svc.CreateCollection(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbcat.CreateCollectionResponse{Collection: mapper.CollectionToProto(col)}), nil
}

// GetCollection implements the proto RPC.
func (h *CatalogHandler) GetCollection(
	ctx context.Context, req *connect.Request[pbcat.GetCollectionRequest],
) (*connect.Response[pbcat.GetCollectionResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	col, err := h.svc.GetCollection(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbcat.GetCollectionResponse{Collection: mapper.CollectionToProto(col)}), nil
}

// ListCollections implements the proto RPC.
func (h *CatalogHandler) ListCollections(
	ctx context.Context, req *connect.Request[pbcat.ListCollectionsRequest],
) (*connect.Response[pbcat.ListCollectionsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	rows, page, err := h.svc.ListCollectionsForTenant(ctx, tid, offset, size)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbcat.ListCollectionsResponse{Page: mapper.PageResponse(page)}
	for _, c := range rows {
		resp.Collections = append(resp.Collections, mapper.CollectionToProto(c))
	}
	return connect.NewResponse(resp), nil
}

// DeleteCollection implements the proto RPC.
func (h *CatalogHandler) DeleteCollection(
	ctx context.Context, req *connect.Request[pbcat.DeleteCollectionRequest],
) (*connect.Response[pbcat.DeleteCollectionResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	if err := h.svc.DeleteCollection(ctx, id); err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbcat.DeleteCollectionResponse{}), nil
}

// ----- Items ---------------------------------------------------------------

// CreateItem implements the proto RPC.
func (h *CatalogHandler) CreateItem(
	ctx context.Context, req *connect.Request[pbcat.CreateItemRequest],
) (*connect.Response[pbcat.CreateItemResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	cid, err := parseULID(req.Msg.GetCollectionId())
	if err != nil {
		return nil, err
	}
	in := services.CreateItemInput{
		TenantID:        tid,
		CollectionID:    cid,
		Mission:         req.Msg.GetMission(),
		Platform:        req.Msg.GetPlatform(),
		Instrument:      req.Msg.GetInstrument(),
		BBox:            mapper.BBoxFromProto(req.Msg.GetBbox()),
		GeometryGeoJSON: req.Msg.GetGeometryGeojson(),
		CloudCover:      req.Msg.GetCloudCover(),
		PropertiesJSON:  req.Msg.GetPropertiesJson(),
	}
	if t := req.Msg.GetDatetime(); t != nil {
		in.Datetime = t.AsTime()
	}
	item, err := h.svc.CreateItem(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbcat.CreateItemResponse{Item: mapper.ItemToProto(item)}), nil
}

// GetItem implements the proto RPC.
func (h *CatalogHandler) GetItem(
	ctx context.Context, req *connect.Request[pbcat.GetItemRequest],
) (*connect.Response[pbcat.GetItemResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	item, err := h.svc.GetItem(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbcat.GetItemResponse{Item: mapper.ItemToProto(item)}), nil
}

// ListItems implements the proto RPC.
func (h *CatalogHandler) ListItems(
	ctx context.Context, req *connect.Request[pbcat.ListItemsRequest],
) (*connect.Response[pbcat.ListItemsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	var cidPtr *ulid.ID
	if cidStr := req.Msg.GetCollectionId(); cidStr != "" {
		cid, err := parseULID(cidStr)
		if err != nil {
			return nil, err
		}
		cidPtr = &cid
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	rows, page, err := h.svc.ListItemsForTenant(ctx, tid, cidPtr, offset, size)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbcat.ListItemsResponse{Page: mapper.PageResponse(page)}
	for _, it := range rows {
		resp.Items = append(resp.Items, mapper.ItemToProto(it))
	}
	return connect.NewResponse(resp), nil
}

// AddAsset implements the proto RPC.
func (h *CatalogHandler) AddAsset(
	ctx context.Context, req *connect.Request[pbcat.AddAssetRequest],
) (*connect.Response[pbcat.AddAssetResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	itemID, err := parseULID(req.Msg.GetItemId())
	if err != nil {
		return nil, err
	}
	asset := mapper.AssetFromProto(req.Msg.GetAsset())
	if asset.Key == "" || asset.Href == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			pkgerrors.BadRequest("INVALID_ARGUMENT", "asset.key and asset.href required"))
	}
	updated, err := h.svc.AddAsset(ctx, itemID, asset)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbcat.AddAssetResponse{Item: mapper.ItemToProto(updated)}), nil
}

// RecordQualityResult implements the proto RPC.
func (h *CatalogHandler) RecordQualityResult(
	ctx context.Context, req *connect.Request[pbcat.RecordQualityResultRequest],
) (*connect.Response[pbcat.RecordQualityResultResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	itemID, err := parseULID(req.Msg.GetItemId())
	if err != nil {
		return nil, err
	}
	q, err := h.svc.RecordQualityResult(ctx, services.RecordQualityResultInput{
		ItemID:             itemID,
		CloudCover:         req.Msg.GetCloudCover(),
		RadiometricRMSE:    req.Msg.GetRadiometricRmse(),
		GeometricAccuracyM: req.Msg.GetGeometricAccuracyM(),
		Notes:              req.Msg.GetNotes(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbcat.RecordQualityResultResponse{Result: mapper.QualityToProto(q)}), nil
}
