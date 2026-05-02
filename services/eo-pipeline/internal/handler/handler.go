// Package handler wires ConnectRPC RPCs to the eo-pipeline service.
package handler

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	pbpipe "github.com/ppusapati/space/services/eo-pipeline/api"
	"github.com/ppusapati/space/services/eo-pipeline/api/eopipelinev1connect"
	"github.com/ppusapati/space/services/eo-pipeline/internal/mapper"
	"github.com/ppusapati/space/services/eo-pipeline/internal/models"
	"github.com/ppusapati/space/services/eo-pipeline/internal/services"
)

// PipelineHandler implements eopipelinev1connect.PipelineServiceHandler.
type PipelineHandler struct {
	eopipelinev1connect.UnimplementedPipelineServiceHandler
	svc       *services.Pipeline
	validator protovalidate.Validator
}

// NewPipelineHandler returns a handler.
func NewPipelineHandler(svc *services.Pipeline) (*PipelineHandler, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &PipelineHandler{svc: svc, validator: v}, nil
}

func (h *PipelineHandler) validate(msg proto.Message) error {
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

// SubmitJob implements the proto RPC.
func (h *PipelineHandler) SubmitJob(
	ctx context.Context, req *connect.Request[pbpipe.SubmitJobRequest],
) (*connect.Response[pbpipe.SubmitJobResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	iid, err := parseULID(req.Msg.GetItemId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.SubmitJob(ctx, services.SubmitJobInput{
		TenantID:       tid,
		ItemID:         iid,
		Stage:          models.JobStage(req.Msg.GetStage()),
		ParametersJSON: req.Msg.GetParametersJson(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbpipe.SubmitJobResponse{Job: mapper.JobToProto(j)}), nil
}

// GetJob implements the proto RPC.
func (h *PipelineHandler) GetJob(
	ctx context.Context, req *connect.Request[pbpipe.GetJobRequest],
) (*connect.Response[pbpipe.GetJobResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.GetJob(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbpipe.GetJobResponse{Job: mapper.JobToProto(j)}), nil
}

// ListJobs implements the proto RPC.
func (h *PipelineHandler) ListJobs(
	ctx context.Context, req *connect.Request[pbpipe.ListJobsRequest],
) (*connect.Response[pbpipe.ListJobsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListJobsInput{
		TenantID:   tid,
		PageOffset: offset,
		PageSize:   size,
	}
	if s := req.Msg.Status; s != nil {
		v := models.JobStatus(*s)
		in.Status = &v
	}
	if s := req.Msg.Stage; s != nil {
		v := models.JobStage(*s)
		in.Stage = &v
	}
	rows, page, err := h.svc.ListJobsForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbpipe.ListJobsResponse{Page: mapper.PageResponse(page)}
	for _, j := range rows {
		resp.Jobs = append(resp.Jobs, mapper.JobToProto(j))
	}
	return connect.NewResponse(resp), nil
}

// UpdateJobStatus implements the proto RPC.
func (h *PipelineHandler) UpdateJobStatus(
	ctx context.Context, req *connect.Request[pbpipe.UpdateJobStatusRequest],
) (*connect.Response[pbpipe.UpdateJobStatusResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.UpdateJobStatus(ctx, id,
		models.JobStatus(req.Msg.GetStatus()),
		req.Msg.GetOutputUri(),
		req.Msg.GetErrorMessage(),
		"")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbpipe.UpdateJobStatusResponse{Job: mapper.JobToProto(j)}), nil
}

// CancelJob implements the proto RPC.
func (h *PipelineHandler) CancelJob(
	ctx context.Context, req *connect.Request[pbpipe.CancelJobRequest],
) (*connect.Response[pbpipe.CancelJobResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	j, err := h.svc.CancelJob(ctx, id, "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbpipe.CancelJobResponse{Job: mapper.JobToProto(j)}), nil
}
