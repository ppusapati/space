// Package handlers wires ConnectRPC RPCs to the eo-pipeline service.
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
	"github.com/ppusapati/space/services/eo-pipeline/internal/mappers"
	"github.com/ppusapati/space/services/eo-pipeline/internal/models"
	"github.com/ppusapati/space/services/eo-pipeline/internal/service"
)

// PipelineHandler implements earthobsv1connect.PipelineServiceHandler.
type PipelineHandler struct {
	earthobsv1connect.UnimplementedPipelineServiceHandler
	svc          *service.Pipeline
	cursorSecret []byte
}

// NewPipelineHandler returns a new handler.
func NewPipelineHandler(svc *service.Pipeline, cursorSecret []byte) *PipelineHandler {
	return &PipelineHandler{svc: svc, cursorSecret: cursorSecret}
}

// SubmitJob ----------------------------------------------------------

func (h *PipelineHandler) SubmitJob(
	ctx context.Context, req *connect.Request[eov1.SubmitJobRequest],
) (*connect.Response[eov1.SubmitJobResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	tid, err := uuid.Parse(req.Msg.GetTenantId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	iid, err := uuid.Parse(req.Msg.GetItemId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	job, err := h.svc.SubmitJob(ctx, service.SubmitJobInput{
		TenantID:       tid,
		ItemID:         iid,
		Stage:          models.JobStage(req.Msg.GetStage()),
		ParametersJSON: req.Msg.GetParametersJson(),
	})
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&eov1.SubmitJobResponse{Job: mappers.JobToProto(job)}), nil
}

// GetJob -------------------------------------------------------------

func (h *PipelineHandler) GetJob(
	ctx context.Context, req *connect.Request[eov1.GetJobRequest],
) (*connect.Response[eov1.GetJobResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	job, err := h.svc.GetJob(ctx, id)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&eov1.GetJobResponse{Job: mappers.JobToProto(job)}), nil
}

// ListJobs -----------------------------------------------------------

func (h *PipelineHandler) ListJobs(
	ctx context.Context, req *connect.Request[eov1.ListJobsRequest],
) (*connect.Response[eov1.ListJobsResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	tid, err := uuid.Parse(req.Msg.GetTenantId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	in := service.ListJobsInput{
		TenantID: tid,
		Limit:    int32(pagination.ClampPageSize(int(req.Msg.GetPage().GetPageSize()))) + 1,
	}
	if s := req.Msg.Status; s != nil {
		v := models.JobStatus(*s)
		in.Status = &v
	}
	if s := req.Msg.Stage; s != nil {
		v := models.JobStage(*s)
		in.Stage = &v
	}
	cursor, err := pagination.Decode(h.cursorSecret, req.Msg.GetPage().GetPageToken())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if !cursor.CreatedAt.IsZero() {
		in.CursorCreated = &cursor.CreatedAt
		in.CursorID = cursor.ID
	}
	rows, err := h.svc.ListJobs(ctx, in)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	resp := &eov1.ListJobsResponse{Page: &commonv1.PageResponse{}}
	limit := in.Limit - 1
	if int32(len(rows)) > limit {
		next := rows[limit-1]
		resp.Page.NextPageToken = pagination.Encode(h.cursorSecret, pagination.Cursor{
			CreatedAt: next.CreatedAt, ID: next.ID,
		})
		rows = rows[:limit]
	}
	for _, j := range rows {
		resp.Jobs = append(resp.Jobs, mappers.JobToProto(j))
	}
	return connect.NewResponse(resp), nil
}

// UpdateJobStatus ----------------------------------------------------

func (h *PipelineHandler) UpdateJobStatus(
	ctx context.Context, req *connect.Request[eov1.UpdateJobStatusRequest],
) (*connect.Response[eov1.UpdateJobStatusResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	job, err := h.svc.UpdateStatus(ctx, id,
		models.JobStatus(req.Msg.GetStatus()),
		req.Msg.GetOutputUri(),
		req.Msg.GetErrorMessage(),
		"",
	)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&eov1.UpdateJobStatusResponse{Job: mappers.JobToProto(job)}), nil
}

// CancelJob ----------------------------------------------------------

func (h *PipelineHandler) CancelJob(
	ctx context.Context, req *connect.Request[eov1.CancelJobRequest],
) (*connect.Response[eov1.CancelJobResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	job, err := h.svc.Cancel(ctx, id, "")
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&eov1.CancelJobResponse{Job: mappers.JobToProto(job)}), nil
}
