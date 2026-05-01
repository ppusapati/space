// Package handlers wires ConnectRPC RPCs to the sat-mission service.
package handlers

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	commonv1 "github.com/ppusapati/space/api/p9e/space/common/v1"
	satv1 "github.com/ppusapati/space/api/p9e/space/satsubsys/v1"
	"github.com/ppusapati/space/api/p9e/space/satsubsys/v1/satsubsysv1connect"
	"github.com/ppusapati/space/pkg/errs"
	"github.com/ppusapati/space/pkg/pagination"
	"github.com/ppusapati/space/pkg/validation"
	"github.com/ppusapati/space/services/sat-mission/internal/mappers"
	"github.com/ppusapati/space/services/sat-mission/internal/models"
	"github.com/ppusapati/space/services/sat-mission/internal/service"
)

// MissionHandler implements satsubsysv1connect.MissionServiceHandler.
type MissionHandler struct {
	satsubsysv1connect.UnimplementedMissionServiceHandler
	svc          *service.Mission
	cursorSecret []byte
}

// NewMissionHandler returns a handler.
func NewMissionHandler(svc *service.Mission, cursorSecret []byte) *MissionHandler {
	return &MissionHandler{svc: svc, cursorSecret: cursorSecret}
}

// RegisterSatellite implements the proto RPC.
func (h *MissionHandler) RegisterSatellite(
	ctx context.Context, req *connect.Request[satv1.RegisterSatelliteRequest],
) (*connect.Response[satv1.RegisterSatelliteResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	tid, err := uuid.Parse(req.Msg.GetTenantId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	s, err := h.svc.RegisterSatellite(ctx, service.RegisterSatelliteInput{
		TenantID:                tid,
		Name:                    req.Msg.GetName(),
		NORADID:                 req.Msg.GetNoradId(),
		InternationalDesignator: req.Msg.GetInternationalDesignator(),
		ConfigJSON:              req.Msg.GetConfigJson(),
	})
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&satv1.RegisterSatelliteResponse{Satellite: mappers.SatelliteToProto(s)}), nil
}

// GetSatellite implements the proto RPC.
func (h *MissionHandler) GetSatellite(
	ctx context.Context, req *connect.Request[satv1.GetSatelliteRequest],
) (*connect.Response[satv1.GetSatelliteResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	s, err := h.svc.GetSatellite(ctx, id)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&satv1.GetSatelliteResponse{Satellite: mappers.SatelliteToProto(s)}), nil
}

// ListSatellites implements the proto RPC.
func (h *MissionHandler) ListSatellites(
	ctx context.Context, req *connect.Request[satv1.ListSatellitesRequest],
) (*connect.Response[satv1.ListSatellitesResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	tid, err := uuid.Parse(req.Msg.GetTenantId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	cursor, err := pagination.Decode(h.cursorSecret, req.Msg.GetPage().GetPageToken())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	limit := int32(pagination.ClampPageSize(int(req.Msg.GetPage().GetPageSize()))) + 1
	var cursorTS *time.Time
	if !cursor.CreatedAt.IsZero() {
		t := cursor.CreatedAt
		cursorTS = &t
	}
	rows, err := h.svc.ListSatellites(ctx, tid, cursorTS, cursor.ID, limit)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	resp := &satv1.ListSatellitesResponse{Page: &commonv1.PageResponse{}}
	page := limit - 1
	if int32(len(rows)) > page {
		next := rows[page-1]
		resp.Page.NextPageToken = pagination.Encode(h.cursorSecret, pagination.Cursor{
			CreatedAt: next.CreatedAt, ID: next.ID,
		})
		rows = rows[:page]
	}
	for _, s := range rows {
		resp.Satellites = append(resp.Satellites, mappers.SatelliteToProto(s))
	}
	return connect.NewResponse(resp), nil
}

// UpdateTLE implements the proto RPC.
func (h *MissionHandler) UpdateTLE(
	ctx context.Context, req *connect.Request[satv1.UpdateTLERequest],
) (*connect.Response[satv1.UpdateTLEResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	s, err := h.svc.UpdateTLE(ctx, id, req.Msg.GetTleLine1(), req.Msg.GetTleLine2(), "")
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&satv1.UpdateTLEResponse{Satellite: mappers.SatelliteToProto(s)}), nil
}

// UpdateOrbitalState implements the proto RPC.
func (h *MissionHandler) UpdateOrbitalState(
	ctx context.Context, req *connect.Request[satv1.UpdateOrbitalStateRequest],
) (*connect.Response[satv1.UpdateOrbitalStateResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if req.Msg.GetState() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errs.New(errs.DomainInvalidArgument, "state required"))
	}
	state := mappers.OrbitalStateFromProto(req.Msg.GetState())
	s, err := h.svc.UpdateOrbitalState(ctx, id, state, "")
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&satv1.UpdateOrbitalStateResponse{Satellite: mappers.SatelliteToProto(s)}), nil
}

// SetMode implements the proto RPC.
func (h *MissionHandler) SetMode(
	ctx context.Context, req *connect.Request[satv1.SetModeRequest],
) (*connect.Response[satv1.SetModeResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	s, err := h.svc.SetMode(ctx, id, models.SatelliteMode(req.Msg.GetMode()), "")
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&satv1.SetModeResponse{Satellite: mappers.SatelliteToProto(s)}), nil
}
