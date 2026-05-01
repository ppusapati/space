// Package handler wires ConnectRPC RPCs to the sat-telemetry service.
package handler

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	pbtlm "github.com/ppusapati/space/services/sat-telemetry/api"
	"github.com/ppusapati/space/services/sat-telemetry/api/sattelemetryv1connect"
	"github.com/ppusapati/space/services/sat-telemetry/internal/mapper"
	"github.com/ppusapati/space/services/sat-telemetry/internal/models"
	"github.com/ppusapati/space/services/sat-telemetry/internal/services"
)

// TelemetryHandler implements sattelemetryv1connect.TelemetryServiceHandler.
type TelemetryHandler struct {
	sattelemetryv1connect.UnimplementedTelemetryServiceHandler
	svc       *services.Telemetry
	validator protovalidate.Validator
}

// NewTelemetryHandler returns a handler.
func NewTelemetryHandler(svc *services.Telemetry) (*TelemetryHandler, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &TelemetryHandler{svc: svc, validator: v}, nil
}

func (h *TelemetryHandler) validate(msg proto.Message) error {
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

// ----- Channel RPCs --------------------------------------------------------

// DefineChannel implements the proto RPC.
func (h *TelemetryHandler) DefineChannel(
	ctx context.Context, req *connect.Request[pbtlm.DefineChannelRequest],
) (*connect.Response[pbtlm.DefineChannelResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	sid, err := parseULID(req.Msg.GetSatelliteId())
	if err != nil {
		return nil, err
	}
	c, err := h.svc.DefineChannel(ctx, services.DefineChannelInput{
		TenantID:     tid,
		SatelliteID:  sid,
		Subsystem:    req.Msg.GetSubsystem(),
		Name:         req.Msg.GetName(),
		Units:        req.Msg.GetUnits(),
		ValueType:    models.ChannelValueType(req.Msg.GetValueType()),
		MinValue:     req.Msg.GetMinValue(),
		MaxValue:     req.Msg.GetMaxValue(),
		SampleRateHz: req.Msg.GetSampleRateHz(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbtlm.DefineChannelResponse{Channel: mapper.ChannelToProto(c)}), nil
}

// GetChannel implements the proto RPC.
func (h *TelemetryHandler) GetChannel(
	ctx context.Context, req *connect.Request[pbtlm.GetChannelRequest],
) (*connect.Response[pbtlm.GetChannelResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	c, err := h.svc.GetChannel(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbtlm.GetChannelResponse{Channel: mapper.ChannelToProto(c)}), nil
}

// ListChannels implements the proto RPC.
func (h *TelemetryHandler) ListChannels(
	ctx context.Context, req *connect.Request[pbtlm.ListChannelsRequest],
) (*connect.Response[pbtlm.ListChannelsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListChannelsInput{
		TenantID:   tid,
		Subsystem:  req.Msg.GetSubsystem(),
		PageOffset: offset,
		PageSize:   size,
	}
	if s := req.Msg.GetSatelliteId(); s != "" {
		sid, err := parseULID(s)
		if err != nil {
			return nil, err
		}
		in.SatelliteID = &sid
	}
	rows, page, err := h.svc.ListChannelsForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbtlm.ListChannelsResponse{Page: mapper.PageResponse(page)}
	for _, c := range rows {
		resp.Channels = append(resp.Channels, mapper.ChannelToProto(c))
	}
	return connect.NewResponse(resp), nil
}

// DeprecateChannel implements the proto RPC.
func (h *TelemetryHandler) DeprecateChannel(
	ctx context.Context, req *connect.Request[pbtlm.DeprecateChannelRequest],
) (*connect.Response[pbtlm.DeprecateChannelResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	c, err := h.svc.DeprecateChannel(ctx, id, "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbtlm.DeprecateChannelResponse{Channel: mapper.ChannelToProto(c)}), nil
}

// ----- Frames + Samples ----------------------------------------------------

// IngestTelemetryFrame implements the proto RPC.
func (h *TelemetryHandler) IngestTelemetryFrame(
	ctx context.Context, req *connect.Request[pbtlm.IngestTelemetryFrameRequest],
) (*connect.Response[pbtlm.IngestTelemetryFrameResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	sid, err := parseULID(req.Msg.GetSatelliteId())
	if err != nil {
		return nil, err
	}
	in := services.IngestFrameInput{
		TenantID:         tid,
		SatelliteID:      sid,
		APID:             req.Msg.GetApid(),
		VirtualChannel:   req.Msg.GetVirtualChannel(),
		SequenceCount:    req.Msg.GetSequenceCount(),
		PayloadSizeBytes: req.Msg.GetPayloadSizeBytes(),
		PayloadSHA256:    req.Msg.GetPayloadSha256(),
		FrameType:        req.Msg.GetFrameType(),
	}
	if t := req.Msg.GetSatTime(); t != nil {
		in.SatTime = t.AsTime()
	}
	for _, s := range req.Msg.GetSamples() {
		cid, err := parseULID(s.GetChannelId())
		if err != nil {
			return nil, err
		}
		smp := services.SampleInput{
			ChannelID:   cid,
			ValueDouble: s.GetValueDouble(),
			ValueInt:    s.GetValueInt(),
			ValueBool:   s.GetValueBool(),
			ValueText:   s.GetValueText(),
		}
		if t := s.GetSampleTime(); t != nil {
			smp.SampleTime = t.AsTime()
		}
		in.Samples = append(in.Samples, smp)
	}
	frame, count, err := h.svc.IngestFrame(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbtlm.IngestTelemetryFrameResponse{
		Frame:       mapper.FrameToProto(frame),
		SampleCount: uint32(count),
	}), nil
}

// GetFrame implements the proto RPC.
func (h *TelemetryHandler) GetFrame(
	ctx context.Context, req *connect.Request[pbtlm.GetFrameRequest],
) (*connect.Response[pbtlm.GetFrameResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	f, err := h.svc.GetFrame(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbtlm.GetFrameResponse{Frame: mapper.FrameToProto(f)}), nil
}

// ListFrames implements the proto RPC.
func (h *TelemetryHandler) ListFrames(
	ctx context.Context, req *connect.Request[pbtlm.ListFramesRequest],
) (*connect.Response[pbtlm.ListFramesResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListFramesInput{
		TenantID:   tid,
		FrameType:  req.Msg.GetFrameType(),
		PageOffset: offset,
		PageSize:   size,
	}
	if s := req.Msg.GetSatelliteId(); s != "" {
		sid, err := parseULID(s)
		if err != nil {
			return nil, err
		}
		in.SatelliteID = &sid
	}
	if t := req.Msg.GetGroundTimeStart(); t != nil {
		in.TimeStart = t.AsTime()
	}
	if t := req.Msg.GetGroundTimeEnd(); t != nil {
		in.TimeEnd = t.AsTime()
	}
	rows, page, err := h.svc.ListFramesForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbtlm.ListFramesResponse{Page: mapper.PageResponse(page)}
	for _, f := range rows {
		resp.Frames = append(resp.Frames, mapper.FrameToProto(f))
	}
	return connect.NewResponse(resp), nil
}

// QuerySamples implements the proto RPC.
func (h *TelemetryHandler) QuerySamples(
	ctx context.Context, req *connect.Request[pbtlm.QuerySamplesRequest],
) (*connect.Response[pbtlm.QuerySamplesResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	cid, err := parseULID(req.Msg.GetChannelId())
	if err != nil {
		return nil, err
	}
	in := services.QuerySamplesInput{
		TenantID:  tid,
		ChannelID: cid,
		Limit:     int32(req.Msg.GetLimit()),
	}
	if in.Limit == 0 {
		in.Limit = 1000
	}
	if t := req.Msg.GetTimeStart(); t != nil {
		in.TimeStart = t.AsTime()
	}
	if t := req.Msg.GetTimeEnd(); t != nil {
		in.TimeEnd = t.AsTime()
	}
	rows, err := h.svc.QuerySamples(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbtlm.QuerySamplesResponse{}
	for _, s := range rows {
		resp.Samples = append(resp.Samples, mapper.SampleToProto(s))
	}
	return connect.NewResponse(resp), nil
}
