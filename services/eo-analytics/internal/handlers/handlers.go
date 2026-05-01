// Package handlers wires ConnectRPC RPCs to the eo-analytics service.
package handlers

import (
	"context"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	commonv1 "github.com/ppusapati/space/api/p9e/space/common/v1"
	eov1 "github.com/ppusapati/space/api/p9e/space/earthobs/v1"
	"github.com/ppusapati/space/api/p9e/space/earthobs/v1/earthobsv1connect"
	"github.com/ppusapati/space/pkg/errs"
	"github.com/ppusapati/space/pkg/pagination"
	"github.com/ppusapati/space/pkg/validation"
	"github.com/ppusapati/space/services/eo-analytics/internal/mappers"
	"github.com/ppusapati/space/services/eo-analytics/internal/models"
	"github.com/ppusapati/space/services/eo-analytics/internal/service"
)

// AnalyticsHandler implements earthobsv1connect.AnalyticsServiceHandler.
type AnalyticsHandler struct {
	earthobsv1connect.UnimplementedAnalyticsServiceHandler
	svc          *service.Analytics
	cursorSecret []byte
}

// NewAnalyticsHandler returns a handler.
func NewAnalyticsHandler(svc *service.Analytics, cursorSecret []byte) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc, cursorSecret: cursorSecret}
}

// ----- Models --------------------------------------------------------

// RegisterModel implements the proto RPC.
func (h *AnalyticsHandler) RegisterModel(
	ctx context.Context, req *connect.Request[eov1.RegisterModelRequest],
) (*connect.Response[eov1.RegisterModelResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	tid, err := uuid.Parse(req.Msg.GetTenantId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	m, err := h.svc.RegisterModel(ctx, service.RegisterModelInput{
		TenantID:     tid,
		Name:         req.Msg.GetName(),
		Version:      req.Msg.GetVersion(),
		Task:         models.InferenceTask(req.Msg.GetTask()),
		Framework:    req.Msg.GetFramework(),
		ArtefactURI:  req.Msg.GetArtefactUri(),
		MetadataJSON: req.Msg.GetMetadataJson(),
	})
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&eov1.RegisterModelResponse{Model: mappers.ModelToProto(m)}), nil
}

// GetModel implements the proto RPC.
func (h *AnalyticsHandler) GetModel(
	ctx context.Context, req *connect.Request[eov1.GetModelRequest],
) (*connect.Response[eov1.GetModelResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	m, err := h.svc.GetModel(ctx, id)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&eov1.GetModelResponse{Model: mappers.ModelToProto(m)}), nil
}

// ListModels implements the proto RPC.
func (h *AnalyticsHandler) ListModels(
	ctx context.Context, req *connect.Request[eov1.ListModelsRequest],
) (*connect.Response[eov1.ListModelsResponse], error) {
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
	in := service.ListModelsInput{
		TenantID: tid,
		Limit:    int32(pagination.ClampPageSize(int(req.Msg.GetPage().GetPageSize()))) + 1,
		CursorID: cursor.ID,
	}
	if !cursor.CreatedAt.IsZero() {
		in.CursorCreated = &cursor.CreatedAt
	}
	if t := req.Msg.Task; t != nil {
		v := models.InferenceTask(*t)
		in.Task = &v
	}
	rows, err := h.svc.ListModels(ctx, in)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	resp := &eov1.ListModelsResponse{Page: &commonv1.PageResponse{}}
	limit := in.Limit - 1
	if int32(len(rows)) > limit {
		next := rows[limit-1]
		resp.Page.NextPageToken = pagination.Encode(h.cursorSecret, pagination.Cursor{
			CreatedAt: next.CreatedAt, ID: next.ID,
		})
		rows = rows[:limit]
	}
	for _, m := range rows {
		resp.Models = append(resp.Models, mappers.ModelToProto(m))
	}
	return connect.NewResponse(resp), nil
}

// DeactivateModel implements the proto RPC.
func (h *AnalyticsHandler) DeactivateModel(
	ctx context.Context, req *connect.Request[eov1.DeactivateModelRequest],
) (*connect.Response[eov1.DeactivateModelResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	m, err := h.svc.DeactivateModel(ctx, id, "")
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&eov1.DeactivateModelResponse{Model: mappers.ModelToProto(m)}), nil
}

// ----- Inference Jobs ------------------------------------------------

// SubmitInferenceJob implements the proto RPC.
func (h *AnalyticsHandler) SubmitInferenceJob(
	ctx context.Context, req *connect.Request[eov1.SubmitInferenceJobRequest],
) (*connect.Response[eov1.SubmitInferenceJobResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	tid, err := uuid.Parse(req.Msg.GetTenantId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	mid, err := uuid.Parse(req.Msg.GetModelId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	iid, err := uuid.Parse(req.Msg.GetItemId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	j, err := h.svc.SubmitInferenceJob(ctx, service.SubmitInferenceJobInput{
		TenantID: tid, ModelID: mid, ItemID: iid,
	})
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&eov1.SubmitInferenceJobResponse{Job: mappers.InferenceJobToProto(j)}), nil
}

// GetInferenceJob implements the proto RPC.
func (h *AnalyticsHandler) GetInferenceJob(
	ctx context.Context, req *connect.Request[eov1.GetInferenceJobRequest],
) (*connect.Response[eov1.GetInferenceJobResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	j, err := h.svc.GetInferenceJob(ctx, id)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&eov1.GetInferenceJobResponse{Job: mappers.InferenceJobToProto(j)}), nil
}

// ListInferenceJobs implements the proto RPC.
func (h *AnalyticsHandler) ListInferenceJobs(
	ctx context.Context, req *connect.Request[eov1.ListInferenceJobsRequest],
) (*connect.Response[eov1.ListInferenceJobsResponse], error) {
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
	in := service.ListInferenceJobsInput{
		TenantID: tid,
		Limit:    int32(pagination.ClampPageSize(int(req.Msg.GetPage().GetPageSize()))) + 1,
		CursorID: cursor.ID,
	}
	if !cursor.CreatedAt.IsZero() {
		in.CursorCreated = &cursor.CreatedAt
	}
	if s := req.Msg.Status; s != nil {
		v := models.InferenceJobStatus(*s)
		in.Status = &v
	}
	rows, err := h.svc.ListInferenceJobs(ctx, in)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	resp := &eov1.ListInferenceJobsResponse{Page: &commonv1.PageResponse{}}
	limit := in.Limit - 1
	if int32(len(rows)) > limit {
		next := rows[limit-1]
		resp.Page.NextPageToken = pagination.Encode(h.cursorSecret, pagination.Cursor{
			CreatedAt: next.CreatedAt, ID: next.ID,
		})
		rows = rows[:limit]
	}
	for _, j := range rows {
		resp.Jobs = append(resp.Jobs, mappers.InferenceJobToProto(j))
	}
	return connect.NewResponse(resp), nil
}

// UpdateInferenceJobStatus implements the proto RPC.
func (h *AnalyticsHandler) UpdateInferenceJobStatus(
	ctx context.Context, req *connect.Request[eov1.UpdateInferenceJobStatusRequest],
) (*connect.Response[eov1.UpdateInferenceJobStatusResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	j, err := h.svc.UpdateInferenceJobStatus(ctx, id,
		models.InferenceJobStatus(req.Msg.GetStatus()),
		req.Msg.GetOutputUri(),
		req.Msg.GetErrorMessage(),
		"",
	)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&eov1.UpdateInferenceJobStatusResponse{Job: mappers.InferenceJobToProto(j)}), nil
}
