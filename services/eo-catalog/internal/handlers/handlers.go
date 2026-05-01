// Package handlers wires ConnectRPC requests to the service layer.
// Each handler validates the request body via protovalidate, calls
// the service, and converts results back to proto.
package handlers

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	commonv1 "github.com/ppusapati/space/api/p9e/space/common/v1"
	eov1 "github.com/ppusapati/space/api/p9e/space/earthobs/v1"
	"github.com/ppusapati/space/api/p9e/space/earthobs/v1/earthobsv1connect"
	"github.com/ppusapati/space/pkg/errs"
	"github.com/ppusapati/space/pkg/pagination"
	"github.com/ppusapati/space/pkg/validation"
	"github.com/ppusapati/space/services/eo-catalog/internal/mappers"
	"github.com/ppusapati/space/services/eo-catalog/internal/models"
	"github.com/ppusapati/space/services/eo-catalog/internal/service"
)

// CatalogHandler implements earthobsv1connect.CatalogServiceHandler.
type CatalogHandler struct {
	earthobsv1connect.UnimplementedCatalogServiceHandler
	svc          *service.Catalog
	cursorSecret []byte
}

// NewCatalogHandler returns a handler with the given service backend.
// The cursor secret is used to HMAC-sign pagination cursors and must be
// at least 16 bytes.
func NewCatalogHandler(svc *service.Catalog, cursorSecret []byte) *CatalogHandler {
	return &CatalogHandler{svc: svc, cursorSecret: cursorSecret}
}

// CreateCollection ----------------------------------------------------

func (h *CatalogHandler) CreateCollection(
	ctx context.Context, req *connect.Request[eov1.CreateCollectionRequest],
) (*connect.Response[eov1.CreateCollectionResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	tenantID, err := uuid.Parse(req.Msg.GetTenantId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	created, err := h.svc.CreateCollection(ctx, service.CreateCollectionInput{
		TenantID:      tenantID,
		Slug:          req.Msg.GetSlug(),
		Title:         req.Msg.GetTitle(),
		Description:   req.Msg.GetDescription(),
		License:       req.Msg.GetLicense(),
		SpatialExtent: mappers.BBoxFromProto(req.Msg.GetSpatialExtent()),
		TemporalStart: protoTime(req.Msg.GetTemporalStart()),
		TemporalEnd:   protoTime(req.Msg.GetTemporalEnd()),
	})
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&eov1.CreateCollectionResponse{
		Collection: mappers.CollectionToProto(created),
	}), nil
}

// GetCollection -------------------------------------------------------

func (h *CatalogHandler) GetCollection(
	ctx context.Context, req *connect.Request[eov1.GetCollectionRequest],
) (*connect.Response[eov1.GetCollectionResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	c, err := h.svc.GetCollection(ctx, id)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&eov1.GetCollectionResponse{
		Collection: mappers.CollectionToProto(c),
	}), nil
}

// ListCollections -----------------------------------------------------

func (h *CatalogHandler) ListCollections(
	ctx context.Context, req *connect.Request[eov1.ListCollectionsRequest],
) (*connect.Response[eov1.ListCollectionsResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	tenantID, err := uuid.Parse(req.Msg.GetTenantId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	cursor, err := pagination.Decode(h.cursorSecret, req.Msg.GetPage().GetPageToken())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	limit := int32(pagination.ClampPageSize(int(req.Msg.GetPage().GetPageSize())))
	cursorTS := optTime(cursor.CreatedAt)
	rows, err := h.svc.ListCollections(ctx, tenantID, cursorTS, cursor.ID, limit+1)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	resp := &eov1.ListCollectionsResponse{Page: &commonv1.PageResponse{}}
	if int32(len(rows)) > limit {
		next := rows[limit-1]
		resp.Page.NextPageToken = pagination.Encode(h.cursorSecret, pagination.Cursor{
			CreatedAt: next.CreatedAt, ID: next.ID,
		})
		rows = rows[:limit]
	}
	for _, c := range rows {
		resp.Collections = append(resp.Collections, mappers.CollectionToProto(c))
	}
	return connect.NewResponse(resp), nil
}

// CreateItem ----------------------------------------------------------

func (h *CatalogHandler) CreateItem(
	ctx context.Context, req *connect.Request[eov1.CreateItemRequest],
) (*connect.Response[eov1.CreateItemResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	tenantID, err := uuid.Parse(req.Msg.GetTenantId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	collectionID, err := uuid.Parse(req.Msg.GetCollectionId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	dt := req.Msg.GetDatetime()
	if dt == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("datetime required"))
	}
	assets := make([]models.Asset, 0, len(req.Msg.GetAssets()))
	for _, a := range req.Msg.GetAssets() {
		assets = append(assets, mappers.AssetFromProto(a))
	}
	created, err := h.svc.CreateItem(ctx, service.CreateItemInput{
		TenantID:        tenantID,
		CollectionID:    collectionID,
		Mission:         req.Msg.GetMission(),
		Platform:        req.Msg.GetPlatform(),
		Instrument:      req.Msg.GetInstrument(),
		Datetime:        dt.AsTime(),
		BBox:            mappers.BBoxFromProto(req.Msg.GetBbox()),
		GeometryGeoJSON: req.Msg.GetGeometryGeojson(),
		CloudCover:      req.Msg.GetCloudCover(),
		PropertiesJSON:  req.Msg.GetPropertiesJson(),
		Assets:          assets,
	})
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&eov1.CreateItemResponse{
		Item: mappers.ItemToProto(created),
	}), nil
}

// GetItem -------------------------------------------------------------

func (h *CatalogHandler) GetItem(
	ctx context.Context, req *connect.Request[eov1.GetItemRequest],
) (*connect.Response[eov1.GetItemResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	it, err := h.svc.GetItem(ctx, id)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&eov1.GetItemResponse{Item: mappers.ItemToProto(it)}), nil
}

// SearchItems ---------------------------------------------------------

func (h *CatalogHandler) SearchItems(
	ctx context.Context, req *connect.Request[eov1.SearchItemsRequest],
) (*connect.Response[eov1.SearchItemsResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	tenantID, err := uuid.Parse(req.Msg.GetTenantId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	in := service.SearchItemsInput{
		TenantID: tenantID,
		Limit:    int32(pagination.ClampPageSize(int(req.Msg.GetPage().GetPageSize()))) + 1,
	}
	if cid := req.Msg.GetCollectionId(); cid != "" {
		c, err := uuid.Parse(cid)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		in.CollectionID = &c
	}
	if t := req.Msg.GetDatetimeStart(); t != nil {
		in.DatetimeStart = t.AsTime()
	}
	if t := req.Msg.GetDatetimeEnd(); t != nil {
		in.DatetimeEnd = t.AsTime()
	}
	if b := req.Msg.GetBbox(); b != nil {
		bb := mappers.BBoxFromProto(b)
		in.BBox = &bb
	}
	if mcc := req.Msg.MaxCloudCover; mcc != nil {
		v := *mcc
		in.MaxCloudCover = &v
	}
	cursor, err := pagination.Decode(h.cursorSecret, req.Msg.GetPage().GetPageToken())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if !cursor.CreatedAt.IsZero() {
		in.CursorDatetime = &cursor.CreatedAt
		in.CursorID = cursor.ID
	}
	rows, err := h.svc.SearchItems(ctx, in)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	resp := &eov1.SearchItemsResponse{Page: &commonv1.PageResponse{}}
	limit := in.Limit - 1
	if int32(len(rows)) > limit {
		next := rows[limit-1]
		resp.Page.NextPageToken = pagination.Encode(h.cursorSecret, pagination.Cursor{
			CreatedAt: next.Datetime, ID: next.ID,
		})
		rows = rows[:limit]
	}
	for _, it := range rows {
		resp.Items = append(resp.Items, mappers.ItemToProto(it))
	}
	return connect.NewResponse(resp), nil
}

// RecordQuality -------------------------------------------------------

func (h *CatalogHandler) RecordQuality(
	ctx context.Context, req *connect.Request[eov1.RecordQualityRequest],
) (*connect.Response[eov1.RecordQualityResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	itemID, err := uuid.Parse(req.Msg.GetItemId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	q, err := h.svc.RecordQuality(ctx, service.RecordQualityInput{
		ItemID:             itemID,
		CloudCover:         req.Msg.GetCloudCover(),
		RadiometricRMSE:    req.Msg.GetRadiometricRmse(),
		GeometricAccuracyM: req.Msg.GetGeometricAccuracyM(),
		Notes:              req.Msg.GetNotes(),
	})
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&eov1.RecordQualityResponse{
		Result: mappers.QualityResultToProto(q),
	}), nil
}

// ListQualityForItem --------------------------------------------------

func (h *CatalogHandler) ListQualityForItem(
	ctx context.Context, req *connect.Request[eov1.ListQualityForItemRequest],
) (*connect.Response[eov1.ListQualityForItemResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	itemID, err := uuid.Parse(req.Msg.GetItemId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	cursor, err := pagination.Decode(h.cursorSecret, req.Msg.GetPage().GetPageToken())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	limit := int32(pagination.ClampPageSize(int(req.Msg.GetPage().GetPageSize())))
	cursorTS := optTime(cursor.CreatedAt)
	rows, err := h.svc.ListQuality(ctx, itemID, cursorTS, cursor.ID, limit+1)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	resp := &eov1.ListQualityForItemResponse{Page: &commonv1.PageResponse{}}
	if int32(len(rows)) > limit {
		next := rows[limit-1]
		resp.Page.NextPageToken = pagination.Encode(h.cursorSecret, pagination.Cursor{
			CreatedAt: next.ComputedAt, ID: next.ID,
		})
		rows = rows[:limit]
	}
	for _, q := range rows {
		resp.Results = append(resp.Results, mappers.QualityResultToProto(q))
	}
	return connect.NewResponse(resp), nil
}

// ----- helpers --------------------------------------------------------

func protoTime(t interface{ AsTime() time.Time }) *time.Time {
	if t == nil {
		return nil
	}
	v := t.AsTime()
	if v.IsZero() {
		return nil
	}
	return &v
}

func optTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
