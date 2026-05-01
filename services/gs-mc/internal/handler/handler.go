// Package handler wires ConnectRPC RPCs to the gs-mc service.
package handler

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	pbmc "github.com/ppusapati/space/services/gs-mc/api"
	"github.com/ppusapati/space/services/gs-mc/api/gsmcv1connect"
	"github.com/ppusapati/space/services/gs-mc/internal/mapper"
	"github.com/ppusapati/space/services/gs-mc/internal/models"
	"github.com/ppusapati/space/services/gs-mc/internal/services"
)

// MCHandler implements gsmcv1connect.MissionControlServiceHandler.
type MCHandler struct {
	gsmcv1connect.UnimplementedMissionControlServiceHandler
	svc       *services.MissionControl
	validator protovalidate.Validator
}

func NewMCHandler(svc *services.MissionControl) (*MCHandler, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &MCHandler{svc: svc, validator: v}, nil
}

func (h *MCHandler) validate(msg proto.Message) error {
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

// ----- GroundStation RPCs -------------------------------------------------

func (h *MCHandler) CreateGroundStation(
	ctx context.Context, req *connect.Request[pbmc.CreateGroundStationRequest],
) (*connect.Response[pbmc.CreateGroundStationResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	s, err := h.svc.CreateGroundStation(ctx, services.CreateGroundStationInput{
		TenantID:     tid,
		Slug:         req.Msg.GetSlug(),
		Name:         req.Msg.GetName(),
		CountryCode:  req.Msg.GetCountryCode(),
		LatitudeDeg:  req.Msg.GetLatitudeDeg(),
		LongitudeDeg: req.Msg.GetLongitudeDeg(),
		AltitudeM:    req.Msg.GetAltitudeM(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbmc.CreateGroundStationResponse{Station: mapper.GroundStationToProto(s)}), nil
}

func (h *MCHandler) GetGroundStation(
	ctx context.Context, req *connect.Request[pbmc.GetGroundStationRequest],
) (*connect.Response[pbmc.GetGroundStationResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	s, err := h.svc.GetGroundStation(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbmc.GetGroundStationResponse{Station: mapper.GroundStationToProto(s)}), nil
}

func (h *MCHandler) ListGroundStations(
	ctx context.Context, req *connect.Request[pbmc.ListGroundStationsRequest],
) (*connect.Response[pbmc.ListGroundStationsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	rows, page, err := h.svc.ListGroundStationsForTenant(ctx, tid, offset, size)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbmc.ListGroundStationsResponse{Page: mapper.PageResponse(page)}
	for _, s := range rows {
		resp.Stations = append(resp.Stations, mapper.GroundStationToProto(s))
	}
	return connect.NewResponse(resp), nil
}

func (h *MCHandler) DeprecateGroundStation(
	ctx context.Context, req *connect.Request[pbmc.DeprecateGroundStationRequest],
) (*connect.Response[pbmc.DeprecateGroundStationResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	s, err := h.svc.DeprecateGroundStation(ctx, id, "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbmc.DeprecateGroundStationResponse{Station: mapper.GroundStationToProto(s)}), nil
}

// ----- Antenna RPCs -------------------------------------------------------

func (h *MCHandler) CreateAntenna(
	ctx context.Context, req *connect.Request[pbmc.CreateAntennaRequest],
) (*connect.Response[pbmc.CreateAntennaResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	stid, err := parseULID(req.Msg.GetStationId())
	if err != nil {
		return nil, err
	}
	a, err := h.svc.CreateAntenna(ctx, services.CreateAntennaInput{
		TenantID:        tid,
		StationID:       stid,
		Slug:            req.Msg.GetSlug(),
		Name:            req.Msg.GetName(),
		Band:            models.FrequencyBand(req.Msg.GetBand()),
		MinFreqHz:       req.Msg.GetMinFreqHz(),
		MaxFreqHz:       req.Msg.GetMaxFreqHz(),
		Polarization:    models.Polarization(req.Msg.GetPolarization()),
		GainDBI:         req.Msg.GetGainDbi(),
		SlewRateDegPerS: req.Msg.GetSlewRateDegPerS(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbmc.CreateAntennaResponse{Antenna: mapper.AntennaToProto(a)}), nil
}

func (h *MCHandler) GetAntenna(
	ctx context.Context, req *connect.Request[pbmc.GetAntennaRequest],
) (*connect.Response[pbmc.GetAntennaResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	a, err := h.svc.GetAntenna(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbmc.GetAntennaResponse{Antenna: mapper.AntennaToProto(a)}), nil
}

func (h *MCHandler) ListAntennas(
	ctx context.Context, req *connect.Request[pbmc.ListAntennasRequest],
) (*connect.Response[pbmc.ListAntennasResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListAntennasInput{
		TenantID:   tid,
		PageOffset: offset,
		PageSize:   size,
	}
	if s := req.Msg.GetStationId(); s != "" {
		stid, err := parseULID(s)
		if err != nil {
			return nil, err
		}
		in.StationID = &stid
	}
	if b := req.Msg.Band; b != nil {
		v := models.FrequencyBand(*b)
		in.Band = &v
	}
	rows, page, err := h.svc.ListAntennasForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbmc.ListAntennasResponse{Page: mapper.PageResponse(page)}
	for _, a := range rows {
		resp.Antennas = append(resp.Antennas, mapper.AntennaToProto(a))
	}
	return connect.NewResponse(resp), nil
}

func (h *MCHandler) DeprecateAntenna(
	ctx context.Context, req *connect.Request[pbmc.DeprecateAntennaRequest],
) (*connect.Response[pbmc.DeprecateAntennaResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	a, err := h.svc.DeprecateAntenna(ctx, id, "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbmc.DeprecateAntennaResponse{Antenna: mapper.AntennaToProto(a)}), nil
}
