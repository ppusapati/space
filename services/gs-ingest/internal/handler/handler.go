// Package handler wires ConnectRPC RPCs to the gs-ingest service.
package handler

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	pbing "github.com/ppusapati/space/services/gs-ingest/api"
	"github.com/ppusapati/space/services/gs-ingest/api/gsingestv1connect"
	"github.com/ppusapati/space/services/gs-ingest/internal/mapper"
	"github.com/ppusapati/space/services/gs-ingest/internal/models"
	"github.com/ppusapati/space/services/gs-ingest/internal/services"
)

type IngestHandler struct {
	gsingestv1connect.UnimplementedIngestServiceHandler
	svc       *services.Ingest
	validator protovalidate.Validator
}

func NewIngestHandler(svc *services.Ingest) (*IngestHandler, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &IngestHandler{svc: svc, validator: v}, nil
}

func (h *IngestHandler) validate(msg proto.Message) error {
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

// ----- IngestSession RPCs -------------------------------------------------

func (h *IngestHandler) StartIngestSession(
	ctx context.Context, req *connect.Request[pbing.StartIngestSessionRequest],
) (*connect.Response[pbing.StartIngestSessionResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	bid, err := parseULID(req.Msg.GetBookingId())
	if err != nil {
		return nil, err
	}
	pid, err := parseULID(req.Msg.GetPassId())
	if err != nil {
		return nil, err
	}
	stid, err := parseULID(req.Msg.GetStationId())
	if err != nil {
		return nil, err
	}
	satid, err := parseULID(req.Msg.GetSatelliteId())
	if err != nil {
		return nil, err
	}
	s, err := h.svc.StartIngestSession(ctx, services.StartIngestSessionInput{
		TenantID: tid, BookingID: bid, PassID: pid,
		StationID: stid, SatelliteID: satid,
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbing.StartIngestSessionResponse{Session: mapper.IngestSessionToProto(s)}), nil
}

func (h *IngestHandler) GetIngestSession(
	ctx context.Context, req *connect.Request[pbing.GetIngestSessionRequest],
) (*connect.Response[pbing.GetIngestSessionResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	s, err := h.svc.GetIngestSession(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbing.GetIngestSessionResponse{Session: mapper.IngestSessionToProto(s)}), nil
}

func (h *IngestHandler) ListIngestSessions(
	ctx context.Context, req *connect.Request[pbing.ListIngestSessionsRequest],
) (*connect.Response[pbing.ListIngestSessionsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListIngestSessionsInput{
		TenantID: tid, PageOffset: offset, PageSize: size,
	}
	if v := req.Msg.GetStationId(); v != "" {
		x, err := parseULID(v)
		if err != nil {
			return nil, err
		}
		in.StationID = &x
	}
	if v := req.Msg.GetSatelliteId(); v != "" {
		x, err := parseULID(v)
		if err != nil {
			return nil, err
		}
		in.SatelliteID = &x
	}
	if s := req.Msg.Status; s != nil {
		v := models.IngestStatus(*s)
		in.Status = &v
	}
	rows, page, err := h.svc.ListIngestSessionsForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbing.ListIngestSessionsResponse{Page: mapper.PageResponse(page)}
	for _, s := range rows {
		resp.Sessions = append(resp.Sessions, mapper.IngestSessionToProto(s))
	}
	return connect.NewResponse(resp), nil
}

func (h *IngestHandler) UpdateIngestStatus(
	ctx context.Context, req *connect.Request[pbing.UpdateIngestStatusRequest],
) (*connect.Response[pbing.UpdateIngestStatusResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	s, err := h.svc.UpdateIngestStatus(ctx, id,
		models.IngestStatus(req.Msg.GetStatus()), req.Msg.GetErrorMessage(), "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbing.UpdateIngestStatusResponse{Session: mapper.IngestSessionToProto(s)}), nil
}

// ----- DownlinkFrame RPCs -------------------------------------------------

func (h *IngestHandler) RecordDownlinkFrame(
	ctx context.Context, req *connect.Request[pbing.RecordDownlinkFrameRequest],
) (*connect.Response[pbing.RecordDownlinkFrameResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	sid, err := parseULID(req.Msg.GetSessionId())
	if err != nil {
		return nil, err
	}
	in := services.RecordDownlinkFrameInput{
		TenantID:         tid,
		SessionID:        sid,
		APID:             req.Msg.GetApid(),
		VirtualChannel:   req.Msg.GetVirtualChannel(),
		SequenceCount:    req.Msg.GetSequenceCount(),
		PayloadSizeBytes: req.Msg.GetPayloadSizeBytes(),
		PayloadSHA256:    req.Msg.GetPayloadSha256(),
		PayloadURI:       req.Msg.GetPayloadUri(),
		FrameType:        req.Msg.GetFrameType(),
	}
	if t := req.Msg.GetGroundTime(); t != nil {
		in.GroundTime = t.AsTime()
	}
	f, err := h.svc.RecordDownlinkFrame(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbing.RecordDownlinkFrameResponse{Frame: mapper.DownlinkFrameToProto(f)}), nil
}

func (h *IngestHandler) GetDownlinkFrame(
	ctx context.Context, req *connect.Request[pbing.GetDownlinkFrameRequest],
) (*connect.Response[pbing.GetDownlinkFrameResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	f, err := h.svc.GetDownlinkFrame(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbing.GetDownlinkFrameResponse{Frame: mapper.DownlinkFrameToProto(f)}), nil
}

func (h *IngestHandler) ListDownlinkFrames(
	ctx context.Context, req *connect.Request[pbing.ListDownlinkFramesRequest],
) (*connect.Response[pbing.ListDownlinkFramesResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListDownlinkFramesInput{
		TenantID:   tid,
		FrameType:  req.Msg.GetFrameType(),
		PageOffset: offset,
		PageSize:   size,
	}
	if v := req.Msg.GetSessionId(); v != "" {
		x, err := parseULID(v)
		if err != nil {
			return nil, err
		}
		in.SessionID = &x
	}
	if t := req.Msg.GetTimeStart(); t != nil {
		in.TimeStart = t.AsTime()
	}
	if t := req.Msg.GetTimeEnd(); t != nil {
		in.TimeEnd = t.AsTime()
	}
	rows, page, err := h.svc.ListDownlinkFramesForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbing.ListDownlinkFramesResponse{Page: mapper.PageResponse(page)}
	for _, f := range rows {
		resp.Frames = append(resp.Frames, mapper.DownlinkFrameToProto(f))
	}
	return connect.NewResponse(resp), nil
}
