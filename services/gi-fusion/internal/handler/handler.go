// Package handler wires ConnectRPC RPCs to the gi-fusion service.
package handler

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	pbfu "github.com/ppusapati/space/services/gi-fusion/api"
	"github.com/ppusapati/space/services/gi-fusion/api/gifusionv1connect"
	"github.com/ppusapati/space/services/gi-fusion/internal/mapper"
	"github.com/ppusapati/space/services/gi-fusion/internal/models"
	"github.com/ppusapati/space/services/gi-fusion/internal/services"
)

type FusionHandler struct {
	gifusionv1connect.UnimplementedFusionServiceHandler
	svc       *services.Fusion
	validator protovalidate.Validator
}

func NewFusionHandler(svc *services.Fusion) (*FusionHandler, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &FusionHandler{svc: svc, validator: v}, nil
}

func (h *FusionHandler) validate(msg proto.Message) error {
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

func (h *FusionHandler) SubmitFusionJob(
	ctx context.Context, req *connect.Request[pbfu.SubmitFusionJobRequest],
) (*connect.Response[pbfu.SubmitFusionJobResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.SubmitFusionJob(ctx, services.SubmitFusionJobInput{
		TenantID:       tid,
		Method:         models.FusionMethod(req.Msg.GetMethod()),
		InputURIs:      req.Msg.GetInputUris(),
		ParametersJSON: req.Msg.GetParametersJson(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbfu.SubmitFusionJobResponse{Job: mapper.FusionJobToProto(j)}), nil
}

func (h *FusionHandler) GetFusionJob(
	ctx context.Context, req *connect.Request[pbfu.GetFusionJobRequest],
) (*connect.Response[pbfu.GetFusionJobResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.GetFusionJob(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbfu.GetFusionJobResponse{Job: mapper.FusionJobToProto(j)}), nil
}

func (h *FusionHandler) ListFusionJobs(
	ctx context.Context, req *connect.Request[pbfu.ListFusionJobsRequest],
) (*connect.Response[pbfu.ListFusionJobsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListFusionJobsInput{TenantID: tid, PageOffset: offset, PageSize: size}
	if s := req.Msg.Status; s != nil {
		v := models.FusionStatus(*s)
		in.Status = &v
	}
	if m := req.Msg.Method; m != nil {
		v := models.FusionMethod(*m)
		in.Method = &v
	}
	rows, page, err := h.svc.ListFusionJobsForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbfu.ListFusionJobsResponse{Page: mapper.PageResponse(page)}
	for _, j := range rows {
		resp.Jobs = append(resp.Jobs, mapper.FusionJobToProto(j))
	}
	return connect.NewResponse(resp), nil
}

func (h *FusionHandler) UpdateFusionJobStatus(
	ctx context.Context, req *connect.Request[pbfu.UpdateFusionJobStatusRequest],
) (*connect.Response[pbfu.UpdateFusionJobStatusResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.UpdateFusionJobStatus(ctx, id,
		models.FusionStatus(req.Msg.GetStatus()),
		req.Msg.GetOutputUri(), req.Msg.GetErrorMessage(), "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbfu.UpdateFusionJobStatusResponse{Job: mapper.FusionJobToProto(j)}), nil
}

func (h *FusionHandler) CancelFusionJob(
	ctx context.Context, req *connect.Request[pbfu.CancelFusionJobRequest],
) (*connect.Response[pbfu.CancelFusionJobResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.CancelFusionJob(ctx, id, "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbfu.CancelFusionJobResponse{Job: mapper.FusionJobToProto(j)}), nil
}
