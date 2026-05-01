// Package handler wires ConnectRPC RPCs to the gi-predict service.
package handler

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	pbpr "github.com/ppusapati/space/services/gi-predict/api"
	"github.com/ppusapati/space/services/gi-predict/api/gipredictv1connect"
	"github.com/ppusapati/space/services/gi-predict/internal/mapper"
	"github.com/ppusapati/space/services/gi-predict/internal/models"
	"github.com/ppusapati/space/services/gi-predict/internal/services"
)

type PredictHandler struct {
	gipredictv1connect.UnimplementedForecastServiceHandler
	svc       *services.Predict
	validator protovalidate.Validator
}

func NewPredictHandler(svc *services.Predict) (*PredictHandler, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &PredictHandler{svc: svc, validator: v}, nil
}

func (h *PredictHandler) validate(msg proto.Message) error {
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

func (h *PredictHandler) SubmitForecastJob(
	ctx context.Context, req *connect.Request[pbpr.SubmitForecastJobRequest],
) (*connect.Response[pbpr.SubmitForecastJobResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	in := services.SubmitForecastJobInput{
		TenantID:       tid,
		Type:           models.ForecastType(req.Msg.GetType()),
		InputURIs:      req.Msg.GetInputUris(),
		HorizonDays:    req.Msg.GetHorizonDays(),
		ParametersJSON: req.Msg.GetParametersJson(),
	}
	if v := req.Msg.GetModelId(); v != "" {
		mid, err := parseULID(v)
		if err != nil {
			return nil, err
		}
		in.ModelID = mid
	}
	j, err := h.svc.SubmitForecastJob(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbpr.SubmitForecastJobResponse{Job: mapper.ForecastJobToProto(j)}), nil
}

func (h *PredictHandler) GetForecastJob(
	ctx context.Context, req *connect.Request[pbpr.GetForecastJobRequest],
) (*connect.Response[pbpr.GetForecastJobResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.GetForecastJob(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbpr.GetForecastJobResponse{Job: mapper.ForecastJobToProto(j)}), nil
}

func (h *PredictHandler) ListForecastJobs(
	ctx context.Context, req *connect.Request[pbpr.ListForecastJobsRequest],
) (*connect.Response[pbpr.ListForecastJobsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListForecastJobsInput{TenantID: tid, PageOffset: offset, PageSize: size}
	if s := req.Msg.Status; s != nil {
		v := models.ForecastStatus(*s)
		in.Status = &v
	}
	if t := req.Msg.Type; t != nil {
		v := models.ForecastType(*t)
		in.Type = &v
	}
	rows, page, err := h.svc.ListForecastJobsForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbpr.ListForecastJobsResponse{Page: mapper.PageResponse(page)}
	for _, j := range rows {
		resp.Jobs = append(resp.Jobs, mapper.ForecastJobToProto(j))
	}
	return connect.NewResponse(resp), nil
}

func (h *PredictHandler) UpdateForecastJobStatus(
	ctx context.Context, req *connect.Request[pbpr.UpdateForecastJobStatusRequest],
) (*connect.Response[pbpr.UpdateForecastJobStatusResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.UpdateForecastJobStatus(ctx, services.UpdateForecastJobStatusInput{
		ID:                 id,
		Status:             models.ForecastStatus(req.Msg.GetStatus()),
		OutputURI:          req.Msg.GetOutputUri(),
		ResultsSummaryJSON: req.Msg.GetResultsSummaryJson(),
		ErrorMessage:       req.Msg.GetErrorMessage(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbpr.UpdateForecastJobStatusResponse{Job: mapper.ForecastJobToProto(j)}), nil
}

func (h *PredictHandler) CancelForecastJob(
	ctx context.Context, req *connect.Request[pbpr.CancelForecastJobRequest],
) (*connect.Response[pbpr.CancelForecastJobResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.CancelForecastJob(ctx, id, "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbpr.CancelForecastJobResponse{Job: mapper.ForecastJobToProto(j)}), nil
}
