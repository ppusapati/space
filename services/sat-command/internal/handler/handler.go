// Package handler wires ConnectRPC RPCs to the sat-command service.
package handler

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	pbcmd "github.com/ppusapati/space/services/sat-command/api"
	"github.com/ppusapati/space/services/sat-command/api/satcommandv1connect"
	"github.com/ppusapati/space/services/sat-command/internal/mapper"
	"github.com/ppusapati/space/services/sat-command/internal/models"
	"github.com/ppusapati/space/services/sat-command/internal/services"
)

// CommandHandler implements satcommandv1connect.CommandServiceHandler.
type CommandHandler struct {
	satcommandv1connect.UnimplementedCommandServiceHandler
	svc       *services.Command
	validator protovalidate.Validator
}

// NewCommandHandler returns a handler.
func NewCommandHandler(svc *services.Command) (*CommandHandler, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &CommandHandler{svc: svc, validator: v}, nil
}

func (h *CommandHandler) validate(msg proto.Message) error {
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

// ----- CommandDef RPCs ----------------------------------------------------

// DefineCommand implements the proto RPC.
func (h *CommandHandler) DefineCommand(
	ctx context.Context, req *connect.Request[pbcmd.DefineCommandRequest],
) (*connect.Response[pbcmd.DefineCommandResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	in := services.DefineCommandInput{
		TenantID:         tid,
		Subsystem:        req.Msg.GetSubsystem(),
		Name:             req.Msg.GetName(),
		Opcode:           req.Msg.GetOpcode(),
		ParametersSchema: req.Msg.GetParametersSchema(),
		Description:      req.Msg.GetDescription(),
	}
	if s := req.Msg.GetSatelliteId(); s != "" {
		sid, err := parseULID(s)
		if err != nil {
			return nil, err
		}
		in.SatelliteID = sid
	}
	c, err := h.svc.DefineCommand(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbcmd.DefineCommandResponse{Command: mapper.CommandDefToProto(c)}), nil
}

// GetCommand implements the proto RPC.
func (h *CommandHandler) GetCommand(
	ctx context.Context, req *connect.Request[pbcmd.GetCommandRequest],
) (*connect.Response[pbcmd.GetCommandResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	c, err := h.svc.GetCommand(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbcmd.GetCommandResponse{Command: mapper.CommandDefToProto(c)}), nil
}

// ListCommands implements the proto RPC.
func (h *CommandHandler) ListCommands(
	ctx context.Context, req *connect.Request[pbcmd.ListCommandsRequest],
) (*connect.Response[pbcmd.ListCommandsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListCommandsInput{
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
	rows, page, err := h.svc.ListCommandsForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbcmd.ListCommandsResponse{Page: mapper.PageResponse(page)}
	for _, c := range rows {
		resp.Commands = append(resp.Commands, mapper.CommandDefToProto(c))
	}
	return connect.NewResponse(resp), nil
}

// DeprecateCommand implements the proto RPC.
func (h *CommandHandler) DeprecateCommand(
	ctx context.Context, req *connect.Request[pbcmd.DeprecateCommandRequest],
) (*connect.Response[pbcmd.DeprecateCommandResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	c, err := h.svc.DeprecateCommand(ctx, id, "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbcmd.DeprecateCommandResponse{Command: mapper.CommandDefToProto(c)}), nil
}

// ----- Uplink RPCs --------------------------------------------------------

// EnqueueUplink implements the proto RPC.
func (h *CommandHandler) EnqueueUplink(
	ctx context.Context, req *connect.Request[pbcmd.EnqueueUplinkRequest],
) (*connect.Response[pbcmd.EnqueueUplinkResponse], error) {
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
	cid, err := parseULID(req.Msg.GetCommandDefId())
	if err != nil {
		return nil, err
	}
	in := services.EnqueueUplinkInput{
		TenantID:       tid,
		SatelliteID:    sid,
		CommandDefID:   cid,
		ParametersJSON: req.Msg.GetParametersJson(),
		GatewayID:      req.Msg.GetGatewayId(),
	}
	if t := req.Msg.GetScheduledRelease(); t != nil {
		in.ScheduledRelease = t.AsTime()
	}
	u, err := h.svc.EnqueueUplink(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbcmd.EnqueueUplinkResponse{Uplink: mapper.UplinkToProto(u)}), nil
}

// GetUplink implements the proto RPC.
func (h *CommandHandler) GetUplink(
	ctx context.Context, req *connect.Request[pbcmd.GetUplinkRequest],
) (*connect.Response[pbcmd.GetUplinkResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	u, err := h.svc.GetUplink(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbcmd.GetUplinkResponse{Uplink: mapper.UplinkToProto(u)}), nil
}

// ListUplinks implements the proto RPC.
func (h *CommandHandler) ListUplinks(
	ctx context.Context, req *connect.Request[pbcmd.ListUplinksRequest],
) (*connect.Response[pbcmd.ListUplinksResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListUplinksInput{
		TenantID:   tid,
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
	if s := req.Msg.Status; s != nil {
		v := models.UplinkStatus(*s)
		in.Status = &v
	}
	if t := req.Msg.GetScheduledReleaseStart(); t != nil {
		in.ReleaseStart = t.AsTime()
	}
	if t := req.Msg.GetScheduledReleaseEnd(); t != nil {
		in.ReleaseEnd = t.AsTime()
	}
	rows, page, err := h.svc.ListUplinksForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbcmd.ListUplinksResponse{Page: mapper.PageResponse(page)}
	for _, u := range rows {
		resp.Uplinks = append(resp.Uplinks, mapper.UplinkToProto(u))
	}
	return connect.NewResponse(resp), nil
}

// CancelUplink implements the proto RPC.
func (h *CommandHandler) CancelUplink(
	ctx context.Context, req *connect.Request[pbcmd.CancelUplinkRequest],
) (*connect.Response[pbcmd.CancelUplinkResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	u, err := h.svc.CancelUplink(ctx, id, req.Msg.GetReason(), "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbcmd.CancelUplinkResponse{Uplink: mapper.UplinkToProto(u)}), nil
}

// UpdateUplinkStatus implements the proto RPC.
func (h *CommandHandler) UpdateUplinkStatus(
	ctx context.Context, req *connect.Request[pbcmd.UpdateUplinkStatusRequest],
) (*connect.Response[pbcmd.UpdateUplinkStatusResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	u, err := h.svc.UpdateUplinkStatus(ctx, id,
		models.UplinkStatus(req.Msg.GetStatus()), req.Msg.GetErrorMessage(), "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbcmd.UpdateUplinkStatusResponse{Uplink: mapper.UplinkToProto(u)}), nil
}
