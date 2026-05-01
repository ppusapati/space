// Package handlers wires ConnectRPC RPCs to the sat-telemetry service.
package handlers

import (
	"context"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	commonv1 "github.com/ppusapati/space/api/p9e/space/common/v1"
	satv1 "github.com/ppusapati/space/api/p9e/space/satsubsys/v1"
	"github.com/ppusapati/space/api/p9e/space/satsubsys/v1/satsubsysv1connect"
	"github.com/ppusapati/space/pkg/errs"
	"github.com/ppusapati/space/pkg/pagination"
	"github.com/ppusapati/space/pkg/validation"
	"github.com/ppusapati/space/services/sat-telemetry/internal/mappers"
	"github.com/ppusapati/space/services/sat-telemetry/internal/models"
	"github.com/ppusapati/space/services/sat-telemetry/internal/service"
)

// TelemetryHandler implements satsubsysv1connect.TelemetryServiceHandler.
type TelemetryHandler struct {
	satsubsysv1connect.UnimplementedTelemetryServiceHandler
	svc          *service.Telemetry
	cursorSecret []byte
}

// NewTelemetryHandler returns a handler.
func NewTelemetryHandler(svc *service.Telemetry, cursorSecret []byte) *TelemetryHandler {
	return &TelemetryHandler{svc: svc, cursorSecret: cursorSecret}
}

// ----- Channels -----------------------------------------------------------

// DefineChannel implements the proto RPC.
func (h *TelemetryHandler) DefineChannel(
	ctx context.Context, req *connect.Request[satv1.DefineChannelRequest],
) (*connect.Response[satv1.DefineChannelResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	tid, err := uuid.Parse(req.Msg.GetTenantId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	sid, err := uuid.Parse(req.Msg.GetSatelliteId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	c, err := h.svc.DefineChannel(ctx, service.DefineChannelInput{
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
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&satv1.DefineChannelResponse{Channel: mappers.ChannelToProto(c)}), nil
}

// GetChannel implements the proto RPC.
func (h *TelemetryHandler) GetChannel(
	ctx context.Context, req *connect.Request[satv1.GetChannelRequest],
) (*connect.Response[satv1.GetChannelResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	c, err := h.svc.GetChannel(ctx, id)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&satv1.GetChannelResponse{Channel: mappers.ChannelToProto(c)}), nil
}

// ListChannels implements the proto RPC.
func (h *TelemetryHandler) ListChannels(
	ctx context.Context, req *connect.Request[satv1.ListChannelsRequest],
) (*connect.Response[satv1.ListChannelsResponse], error) {
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
	in := service.ListChannelsInput{
		TenantID: tid,
		Limit:    int32(pagination.ClampPageSize(int(req.Msg.GetPage().GetPageSize()))) + 1,
		CursorID: cursor.ID,
	}
	if !cursor.CreatedAt.IsZero() {
		t := cursor.CreatedAt
		in.CursorCreated = &t
	}
	if sid := req.Msg.GetSatelliteId(); sid != "" {
		parsed, err := uuid.Parse(sid)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		in.SatelliteID = &parsed
	}
	if sub := req.Msg.GetSubsystem(); sub != "" {
		in.Subsystem = &sub
	}
	rows, err := h.svc.ListChannels(ctx, in)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	resp := &satv1.ListChannelsResponse{Page: &commonv1.PageResponse{}}
	page := in.Limit - 1
	if int32(len(rows)) > page {
		next := rows[page-1]
		resp.Page.NextPageToken = pagination.Encode(h.cursorSecret, pagination.Cursor{
			CreatedAt: next.CreatedAt, ID: next.ID,
		})
		rows = rows[:page]
	}
	for _, c := range rows {
		resp.Channels = append(resp.Channels, mappers.ChannelToProto(c))
	}
	return connect.NewResponse(resp), nil
}

// DeprecateChannel implements the proto RPC.
func (h *TelemetryHandler) DeprecateChannel(
	ctx context.Context, req *connect.Request[satv1.DeprecateChannelRequest],
) (*connect.Response[satv1.DeprecateChannelResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	c, err := h.svc.DeprecateChannel(ctx, id, "")
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&satv1.DeprecateChannelResponse{Channel: mappers.ChannelToProto(c)}), nil
}

// ----- Frames + Samples ----------------------------------------------------

// IngestTelemetryFrame implements the proto RPC.
func (h *TelemetryHandler) IngestTelemetryFrame(
	ctx context.Context, req *connect.Request[satv1.IngestTelemetryFrameRequest],
) (*connect.Response[satv1.IngestTelemetryFrameResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	tid, err := uuid.Parse(req.Msg.GetTenantId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	sid, err := uuid.Parse(req.Msg.GetSatelliteId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	in := service.IngestFrameInput{
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
	for i, s := range req.Msg.GetSamples() {
		cid, err := uuid.Parse(s.GetChannelId())
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		sample := service.SampleInput{
			ChannelID:   cid,
			ValueDouble: s.GetValueDouble(),
			ValueInt:    s.GetValueInt(),
			ValueBool:   s.GetValueBool(),
			ValueText:   s.GetValueText(),
		}
		if t := s.GetSampleTime(); t != nil {
			sample.SampleTime = t.AsTime()
		}
		_ = i
		in.Samples = append(in.Samples, sample)
	}
	frame, count, err := h.svc.IngestFrame(ctx, in)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&satv1.IngestTelemetryFrameResponse{
		Frame:       mappers.FrameToProto(frame),
		SampleCount: uint32(count),
	}), nil
}

// GetFrame implements the proto RPC.
func (h *TelemetryHandler) GetFrame(
	ctx context.Context, req *connect.Request[satv1.GetFrameRequest],
) (*connect.Response[satv1.GetFrameResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	f, err := h.svc.GetFrame(ctx, id)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&satv1.GetFrameResponse{Frame: mappers.FrameToProto(f)}), nil
}

// ListFrames implements the proto RPC.
func (h *TelemetryHandler) ListFrames(
	ctx context.Context, req *connect.Request[satv1.ListFramesRequest],
) (*connect.Response[satv1.ListFramesResponse], error) {
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
	in := service.ListFramesInput{
		TenantID: tid,
		Limit:    int32(pagination.ClampPageSize(int(req.Msg.GetPage().GetPageSize()))) + 1,
		CursorID: cursor.ID,
	}
	if !cursor.CreatedAt.IsZero() {
		t := cursor.CreatedAt
		in.CursorGroundTime = &t
	}
	if sid := req.Msg.GetSatelliteId(); sid != "" {
		parsed, err := uuid.Parse(sid)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		in.SatelliteID = &parsed
	}
	if ft := req.Msg.GetFrameType(); ft != "" {
		in.FrameType = &ft
	}
	if t := req.Msg.GetGroundTimeStart(); t != nil {
		v := t.AsTime()
		in.GroundTimeStart = &v
	}
	if t := req.Msg.GetGroundTimeEnd(); t != nil {
		v := t.AsTime()
		in.GroundTimeEnd = &v
	}
	rows, err := h.svc.ListFrames(ctx, in)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	resp := &satv1.ListFramesResponse{Page: &commonv1.PageResponse{}}
	page := in.Limit - 1
	if int32(len(rows)) > page {
		next := rows[page-1]
		resp.Page.NextPageToken = pagination.Encode(h.cursorSecret, pagination.Cursor{
			CreatedAt: next.GroundTime, ID: next.ID,
		})
		rows = rows[:page]
	}
	for _, f := range rows {
		resp.Frames = append(resp.Frames, mappers.FrameToProto(f))
	}
	return connect.NewResponse(resp), nil
}

// QuerySamples implements the proto RPC.
func (h *TelemetryHandler) QuerySamples(
	ctx context.Context, req *connect.Request[satv1.QuerySamplesRequest],
) (*connect.Response[satv1.QuerySamplesResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	tid, err := uuid.Parse(req.Msg.GetTenantId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	cid, err := uuid.Parse(req.Msg.GetChannelId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	in := service.QuerySamplesInput{
		TenantID:  tid,
		ChannelID: cid,
		Limit:     int32(req.Msg.GetLimit()),
	}
	if in.Limit == 0 {
		in.Limit = 1000
	}
	if t := req.Msg.GetTimeStart(); t != nil {
		v := t.AsTime()
		in.TimeStart = &v
	}
	if t := req.Msg.GetTimeEnd(); t != nil {
		v := t.AsTime()
		in.TimeEnd = &v
	}
	rows, err := h.svc.QuerySamples(ctx, in)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	resp := &satv1.QuerySamplesResponse{}
	for _, s := range rows {
		resp.Samples = append(resp.Samples, mappers.SampleToProto(s))
	}
	return connect.NewResponse(resp), nil
}
