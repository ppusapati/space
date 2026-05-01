// Package handler wires ConnectRPC RPCs to the gs-scheduler service.
package handler

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	pbsch "github.com/ppusapati/space/services/gs-scheduler/api"
	"github.com/ppusapati/space/services/gs-scheduler/api/gsschedulerv1connect"
	"github.com/ppusapati/space/services/gs-scheduler/internal/mapper"
	"github.com/ppusapati/space/services/gs-scheduler/internal/models"
	"github.com/ppusapati/space/services/gs-scheduler/internal/services"
)

type SchedulerHandler struct {
	gsschedulerv1connect.UnimplementedSchedulerServiceHandler
	svc       *services.Scheduler
	validator protovalidate.Validator
}

func NewSchedulerHandler(svc *services.Scheduler) (*SchedulerHandler, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &SchedulerHandler{svc: svc, validator: v}, nil
}

func (h *SchedulerHandler) validate(msg proto.Message) error {
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

// ----- ContactPass RPCs ---------------------------------------------------

func (h *SchedulerHandler) InsertContactPass(
	ctx context.Context, req *connect.Request[pbsch.InsertContactPassRequest],
) (*connect.Response[pbsch.InsertContactPassResponse], error) {
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
	satid, err := parseULID(req.Msg.GetSatelliteId())
	if err != nil {
		return nil, err
	}
	in := services.InsertContactPassInput{
		TenantID:        tid,
		StationID:       stid,
		SatelliteID:     satid,
		MaxElevationDeg: req.Msg.GetMaxElevationDeg(),
		AOSAzimuthDeg:   req.Msg.GetAosAzimuthDeg(),
		LOSAzimuthDeg:   req.Msg.GetLosAzimuthDeg(),
		Source:          req.Msg.GetSource(),
	}
	if t := req.Msg.GetAosTime(); t != nil {
		in.AOSTime = t.AsTime()
	}
	if t := req.Msg.GetTcaTime(); t != nil {
		in.TCATime = t.AsTime()
	}
	if t := req.Msg.GetLosTime(); t != nil {
		in.LOSTime = t.AsTime()
	}
	p, err := h.svc.InsertContactPass(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsch.InsertContactPassResponse{Pass: mapper.ContactPassToProto(p)}), nil
}

func (h *SchedulerHandler) GetContactPass(
	ctx context.Context, req *connect.Request[pbsch.GetContactPassRequest],
) (*connect.Response[pbsch.GetContactPassResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	p, err := h.svc.GetContactPass(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsch.GetContactPassResponse{Pass: mapper.ContactPassToProto(p)}), nil
}

func (h *SchedulerHandler) ListContactPasses(
	ctx context.Context, req *connect.Request[pbsch.ListContactPassesRequest],
) (*connect.Response[pbsch.ListContactPassesResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListContactPassesInput{
		TenantID:   tid,
		PageOffset: offset,
		PageSize:   size,
	}
	if s := req.Msg.GetStationId(); s != "" {
		v, err := parseULID(s)
		if err != nil {
			return nil, err
		}
		in.StationID = &v
	}
	if s := req.Msg.GetSatelliteId(); s != "" {
		v, err := parseULID(s)
		if err != nil {
			return nil, err
		}
		in.SatelliteID = &v
	}
	if t := req.Msg.GetAosStart(); t != nil {
		in.AOSStart = t.AsTime()
	}
	if t := req.Msg.GetAosEnd(); t != nil {
		in.AOSEnd = t.AsTime()
	}
	if v := req.Msg.GetMinElevationDeg(); v > 0 {
		in.MinElevationDeg = &v
	}
	rows, page, err := h.svc.ListContactPassesForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbsch.ListContactPassesResponse{Page: mapper.PageResponse(page)}
	for _, p := range rows {
		resp.Passes = append(resp.Passes, mapper.ContactPassToProto(p))
	}
	return connect.NewResponse(resp), nil
}

// ----- Booking RPCs -------------------------------------------------------

func (h *SchedulerHandler) RequestBooking(
	ctx context.Context, req *connect.Request[pbsch.RequestBookingRequest],
) (*connect.Response[pbsch.RequestBookingResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	pid, err := parseULID(req.Msg.GetPassId())
	if err != nil {
		return nil, err
	}
	b, err := h.svc.RequestBooking(ctx, services.RequestBookingInput{
		TenantID: tid,
		PassID:   pid,
		Priority: req.Msg.GetPriority(),
		Purpose:  req.Msg.GetPurpose(),
		Notes:    req.Msg.GetNotes(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsch.RequestBookingResponse{Booking: mapper.BookingToProto(b)}), nil
}

func (h *SchedulerHandler) GetBooking(
	ctx context.Context, req *connect.Request[pbsch.GetBookingRequest],
) (*connect.Response[pbsch.GetBookingResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	b, err := h.svc.GetBooking(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsch.GetBookingResponse{Booking: mapper.BookingToProto(b)}), nil
}

func (h *SchedulerHandler) ListBookings(
	ctx context.Context, req *connect.Request[pbsch.ListBookingsRequest],
) (*connect.Response[pbsch.ListBookingsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListBookingsInput{
		TenantID:   tid,
		PageOffset: offset,
		PageSize:   size,
	}
	if s := req.Msg.Status; s != nil {
		v := models.BookingStatus(*s)
		in.Status = &v
	}
	rows, page, err := h.svc.ListBookingsForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbsch.ListBookingsResponse{Page: mapper.PageResponse(page)}
	for _, b := range rows {
		resp.Bookings = append(resp.Bookings, mapper.BookingToProto(b))
	}
	return connect.NewResponse(resp), nil
}

func (h *SchedulerHandler) UpdateBookingStatus(
	ctx context.Context, req *connect.Request[pbsch.UpdateBookingStatusRequest],
) (*connect.Response[pbsch.UpdateBookingStatusResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	b, err := h.svc.UpdateBookingStatus(ctx, id,
		models.BookingStatus(req.Msg.GetStatus()), req.Msg.GetErrorMessage(), "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsch.UpdateBookingStatusResponse{Booking: mapper.BookingToProto(b)}), nil
}

func (h *SchedulerHandler) CancelBooking(
	ctx context.Context, req *connect.Request[pbsch.CancelBookingRequest],
) (*connect.Response[pbsch.CancelBookingResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	b, err := h.svc.CancelBooking(ctx, id, req.Msg.GetReason(), "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsch.CancelBookingResponse{Booking: mapper.BookingToProto(b)}), nil
}
