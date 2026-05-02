// Package handler wires ConnectRPC RPCs to the eo-analytics service.
package handler

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	pban "github.com/ppusapati/space/services/eo-analytics/api"
	"github.com/ppusapati/space/services/eo-analytics/api/eoanalyticsv1connect"
	"github.com/ppusapati/space/services/eo-analytics/internal/mapper"
	"github.com/ppusapati/space/services/eo-analytics/internal/models"
	"github.com/ppusapati/space/services/eo-analytics/internal/services"
)

// AnalyticsHandler implements eoanalyticsv1connect.AnalyticsServiceHandler.
type AnalyticsHandler struct {
	eoanalyticsv1connect.UnimplementedAnalyticsServiceHandler
	svc       *services.Analytics
	validator protovalidate.Validator
}

// NewAnalyticsHandler returns a handler.
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

// ----- Models --------------------------------------------------------------

// RegisterModel implements the proto RPC.
func (h *AnalyticsHandler) RegisterModel(
	ctx context.Context, req *connect.Request[pban.RegisterModelRequest],
) (*connect.Response[pban.RegisterModelResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	m, err := h.svc.RegisterModel(ctx, services.RegisterModelInput{
		TenantID:     tid,
		Name:         req.Msg.GetName(),
		Version:      req.Msg.GetVersion(),
		Task:         models.InferenceTask(req.Msg.GetTask()),
		Framework:    req.Msg.GetFramework(),
		ArtefactURI:  req.Msg.GetArtefactUri(),
		MetadataJSON: req.Msg.GetMetadataJson(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pban.RegisterModelResponse{Model: mapper.ModelToProto(m)}), nil
}

// GetModel implements the proto RPC.
func (h *AnalyticsHandler) GetModel(
	ctx context.Context, req *connect.Request[pban.GetModelRequest],
) (*connect.Response[pban.GetModelResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	m, err := h.svc.GetModel(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pban.GetModelResponse{Model: mapper.ModelToProto(m)}), nil
}

// ListModels implements the proto RPC.
func (h *AnalyticsHandler) ListModels(
	ctx context.Context, req *connect.Request[pban.ListModelsRequest],
) (*connect.Response[pban.ListModelsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListModelsInput{
		TenantID:   tid,
		PageOffset: offset,
		PageSize:   size,
	}
	if t := req.Msg.Task; t != nil {
		v := models.InferenceTask(*t)
		in.Task = &v
	}
	rows, page, err := h.svc.ListModelsForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pban.ListModelsResponse{Page: mapper.PageResponse(page)}
	for _, m := range rows {
		resp.Models = append(resp.Models, mapper.ModelToProto(m))
	}
	return connect.NewResponse(resp), nil
}

// DeactivateModel implements the proto RPC.
func (h *AnalyticsHandler) DeactivateModel(
	ctx context.Context, req *connect.Request[pban.DeactivateModelRequest],
) (*connect.Response[pban.DeactivateModelResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	m, err := h.svc.DeactivateModel(ctx, id, "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pban.DeactivateModelResponse{Model: mapper.ModelToProto(m)}), nil
}

// ----- InferenceJobs -------------------------------------------------------

// SubmitInferenceJob implements the proto RPC.
func (h *AnalyticsHandler) SubmitInferenceJob(
	ctx context.Context, req *connect.Request[pban.SubmitInferenceJobRequest],
) (*connect.Response[pban.SubmitInferenceJobResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	mid, err := parseULID(req.Msg.GetModelId())
	if err != nil {
		return nil, err
	}
	iid, err := parseULID(req.Msg.GetItemId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.SubmitInferenceJob(ctx, services.SubmitInferenceJobInput{
		TenantID: tid,
		ModelID:  mid,
		ItemID:   iid,
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pban.SubmitInferenceJobResponse{Job: mapper.InferenceJobToProto(j)}), nil
}

// GetInferenceJob implements the proto RPC.
func (h *AnalyticsHandler) GetInferenceJob(
	ctx context.Context, req *connect.Request[pban.GetInferenceJobRequest],
) (*connect.Response[pban.GetInferenceJobResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.GetInferenceJob(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pban.GetInferenceJobResponse{Job: mapper.InferenceJobToProto(j)}), nil
}

// ListInferenceJobs implements the proto RPC.
func (h *AnalyticsHandler) ListInferenceJobs(
	ctx context.Context, req *connect.Request[pban.ListInferenceJobsRequest],
) (*connect.Response[pban.ListInferenceJobsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListInferenceJobsInput{
		TenantID:   tid,
		PageOffset: offset,
		PageSize:   size,
	}
	if s := req.Msg.Status; s != nil {
		v := models.InferenceJobStatus(*s)
		in.Status = &v
	}
	rows, page, err := h.svc.ListInferenceJobsForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pban.ListInferenceJobsResponse{Page: mapper.PageResponse(page)}
	for _, j := range rows {
		resp.Jobs = append(resp.Jobs, mapper.InferenceJobToProto(j))
	}
	return connect.NewResponse(resp), nil
}

// UpdateInferenceJobStatus implements the proto RPC.
func (h *AnalyticsHandler) UpdateInferenceJobStatus(
	ctx context.Context, req *connect.Request[pban.UpdateInferenceJobStatusRequest],
) (*connect.Response[pban.UpdateInferenceJobStatusResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.UpdateInferenceJobStatus(ctx, id,
		models.InferenceJobStatus(req.Msg.GetStatus()),
		req.Msg.GetOutputUri(),
		req.Msg.GetErrorMessage(),
		"")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pban.UpdateInferenceJobStatusResponse{Job: mapper.InferenceJobToProto(j)}), nil
}
