// Package handler wires ConnectRPC RPCs to the sat-mission service.
package handler

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	pbsm "github.com/ppusapati/space/services/sat-mission/api"
	"github.com/ppusapati/space/services/sat-mission/api/satmissionv1connect"
	"github.com/ppusapati/space/services/sat-mission/internal/mapper"
	"github.com/ppusapati/space/services/sat-mission/internal/models"
	"github.com/ppusapati/space/services/sat-mission/internal/services"
)

// MissionHandler implements satmissionv1connect.MissionServiceHandler.
type MissionHandler struct {
	satmissionv1connect.UnimplementedMissionServiceHandler
	svc       *services.Mission
	validator protovalidate.Validator
}

// NewMissionHandler returns a handler.
func NewMissionHandler(svc *services.Mission) (*MissionHandler, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &MissionHandler{svc: svc, validator: v}, nil
}

func (h *MissionHandler) validate(msg proto.Message) error {
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

// RegisterSatellite implements the proto RPC.
func (h *MissionHandler) RegisterSatellite(
	ctx context.Context, req *connect.Request[pbsm.RegisterSatelliteRequest],
) (*connect.Response[pbsm.RegisterSatelliteResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	s, err := h.svc.RegisterSatellite(ctx, services.RegisterSatelliteInput{
		TenantID:                tid,
		Name:                    req.Msg.GetName(),
		NoradID:                 req.Msg.GetNoradId(),
		InternationalDesignator: req.Msg.GetInternationalDesignator(),
		ConfigJSON:              req.Msg.GetConfigJson(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsm.RegisterSatelliteResponse{Satellite: mapper.SatelliteToProto(s)}), nil
}

// GetSatellite implements the proto RPC.
func (h *MissionHandler) GetSatellite(
	ctx context.Context, req *connect.Request[pbsm.GetSatelliteRequest],
) (*connect.Response[pbsm.GetSatelliteResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	s, err := h.svc.GetSatellite(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsm.GetSatelliteResponse{Satellite: mapper.SatelliteToProto(s)}), nil
}

// ListSatellites implements the proto RPC.
func (h *MissionHandler) ListSatellites(
	ctx context.Context, req *connect.Request[pbsm.ListSatellitesRequest],
) (*connect.Response[pbsm.ListSatellitesResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	rows, page, err := h.svc.ListSatellitesForTenant(ctx, tid, offset, size)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbsm.ListSatellitesResponse{Page: mapper.PageResponse(page)}
	for _, s := range rows {
		resp.Satellites = append(resp.Satellites, mapper.SatelliteToProto(s))
	}
	return connect.NewResponse(resp), nil
}

// UpdateTLE implements the proto RPC.
func (h *MissionHandler) UpdateTLE(
	ctx context.Context, req *connect.Request[pbsm.UpdateTLERequest],
) (*connect.Response[pbsm.UpdateTLEResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	s, err := h.svc.UpdateTLE(ctx, id, req.Msg.GetTleLine1(), req.Msg.GetTleLine2(), "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsm.UpdateTLEResponse{Satellite: mapper.SatelliteToProto(s)}), nil
}

// UpdateOrbitalState implements the proto RPC.
func (h *MissionHandler) UpdateOrbitalState(
	ctx context.Context, req *connect.Request[pbsm.UpdateOrbitalStateRequest],
) (*connect.Response[pbsm.UpdateOrbitalStateResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	if req.Msg.GetState() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			pkgerrors.BadRequest("INVALID_ARGUMENT", "state required"))
	}
	state := mapper.OrbitalStateFromProto(req.Msg.GetState())
	s, err := h.svc.UpdateOrbitalState(ctx, id, state, "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsm.UpdateOrbitalStateResponse{Satellite: mapper.SatelliteToProto(s)}), nil
}

// SetMode implements the proto RPC.
func (h *MissionHandler) SetMode(
	ctx context.Context, req *connect.Request[pbsm.SetModeRequest],
) (*connect.Response[pbsm.SetModeResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	s, err := h.svc.SetMode(ctx, id, models.SatelliteMode(req.Msg.GetMode()), "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsm.SetModeResponse{Satellite: mapper.SatelliteToProto(s)}), nil
}
