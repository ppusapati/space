// Package handler wires ConnectRPC RPCs to the gi-analytics service.
package handler

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	pban "github.com/ppusapati/space/services/gi-analytics/api"
	"github.com/ppusapati/space/services/gi-analytics/api/gianalyticsv1connect"
	"github.com/ppusapati/space/services/gi-analytics/internal/mapper"
	"github.com/ppusapati/space/services/gi-analytics/internal/models"
	"github.com/ppusapati/space/services/gi-analytics/internal/services"
)

type AnalyticsHandler struct {
	gianalyticsv1connect.UnimplementedAnalyticsServiceHandler
	svc       *services.Analytics
	validator protovalidate.Validator
}

func NewAnalyticsHandler(svc *services.Analytics) (*AnalyticsHandler, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &AnalyticsHandler{svc: svc, validator: v}, nil
}

func (h *AnalyticsHandler) validate(msg proto.Message) error {
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

func (h *AnalyticsHandler) SubmitAnalysisJob(
	ctx context.Context, req *connect.Request[pban.SubmitAnalysisJobRequest],
) (*connect.Response[pban.SubmitAnalysisJobResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.SubmitAnalysisJob(ctx, services.SubmitAnalysisJobInput{
		TenantID:       tid,
		Type:           models.AnalysisType(req.Msg.GetType()),
		InputURIs:      req.Msg.GetInputUris(),
		ParametersJSON: req.Msg.GetParametersJson(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pban.SubmitAnalysisJobResponse{Job: mapper.AnalysisJobToProto(j)}), nil
}

func (h *AnalyticsHandler) GetAnalysisJob(
	ctx context.Context, req *connect.Request[pban.GetAnalysisJobRequest],
) (*connect.Response[pban.GetAnalysisJobResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.GetAnalysisJob(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pban.GetAnalysisJobResponse{Job: mapper.AnalysisJobToProto(j)}), nil
}

func (h *AnalyticsHandler) ListAnalysisJobs(
	ctx context.Context, req *connect.Request[pban.ListAnalysisJobsRequest],
) (*connect.Response[pban.ListAnalysisJobsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListAnalysisJobsInput{TenantID: tid, PageOffset: offset, PageSize: size}
	if s := req.Msg.Status; s != nil {
		v := models.AnalysisStatus(*s)
		in.Status = &v
	}
	if t := req.Msg.Type; t != nil {
		v := models.AnalysisType(*t)
		in.Type = &v
	}
	rows, page, err := h.svc.ListAnalysisJobsForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pban.ListAnalysisJobsResponse{Page: mapper.PageResponse(page)}
	for _, j := range rows {
		resp.Jobs = append(resp.Jobs, mapper.AnalysisJobToProto(j))
	}
	return connect.NewResponse(resp), nil
}

func (h *AnalyticsHandler) UpdateAnalysisJobStatus(
	ctx context.Context, req *connect.Request[pban.UpdateAnalysisJobStatusRequest],
) (*connect.Response[pban.UpdateAnalysisJobStatusResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.UpdateAnalysisJobStatus(ctx, services.UpdateAnalysisJobStatusInput{
		ID:                 id,
		Status:             models.AnalysisStatus(req.Msg.GetStatus()),
		OutputURI:          req.Msg.GetOutputUri(),
		ResultsSummaryJSON: req.Msg.GetResultsSummaryJson(),
		ErrorMessage:       req.Msg.GetErrorMessage(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pban.UpdateAnalysisJobStatusResponse{Job: mapper.AnalysisJobToProto(j)}), nil
}

func (h *AnalyticsHandler) CancelAnalysisJob(
	ctx context.Context, req *connect.Request[pban.CancelAnalysisJobRequest],
) (*connect.Response[pban.CancelAnalysisJobResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.CancelAnalysisJob(ctx, id, "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pban.CancelAnalysisJobResponse{Job: mapper.AnalysisJobToProto(j)}), nil
}
