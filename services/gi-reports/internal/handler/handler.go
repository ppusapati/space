// Package handler wires ConnectRPC RPCs to the gi-reports service.
package handler

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	pkgerrors "p9e.in/chetana/packages/errors"
	"p9e.in/chetana/packages/ulid"

	pbre "github.com/ppusapati/space/services/gi-reports/api"
	"github.com/ppusapati/space/services/gi-reports/api/gireportsv1connect"
	"github.com/ppusapati/space/services/gi-reports/internal/mapper"
	"github.com/ppusapati/space/services/gi-reports/internal/models"
	"github.com/ppusapati/space/services/gi-reports/internal/services"
)

type ReportsHandler struct {
	gireportsv1connect.UnimplementedReportsServiceHandler
	svc       *services.Reports
	validator protovalidate.Validator
}

func NewReportsHandler(svc *services.Reports) (*ReportsHandler, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &ReportsHandler{svc: svc, validator: v}, nil
}

func (h *ReportsHandler) validate(msg proto.Message) error {
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

// ----- Template RPCs ------------------------------------------------------

func (h *ReportsHandler) CreateTemplate(
	ctx context.Context, req *connect.Request[pbre.CreateTemplateRequest],
) (*connect.Response[pbre.CreateTemplateResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	t, err := h.svc.CreateTemplate(ctx, services.CreateTemplateInput{
		TenantID:         tid,
		Slug:             req.Msg.GetSlug(),
		Name:             req.Msg.GetName(),
		Description:      req.Msg.GetDescription(),
		TemplateURI:      req.Msg.GetTemplateUri(),
		Format:           models.ReportFormat(req.Msg.GetFormat()),
		ParametersSchema: req.Msg.GetParametersSchema(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbre.CreateTemplateResponse{Template: mapper.TemplateToProto(t)}), nil
}

func (h *ReportsHandler) GetTemplate(
	ctx context.Context, req *connect.Request[pbre.GetTemplateRequest],
) (*connect.Response[pbre.GetTemplateResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	t, err := h.svc.GetTemplate(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbre.GetTemplateResponse{Template: mapper.TemplateToProto(t)}), nil
}

func (h *ReportsHandler) ListTemplates(
	ctx context.Context, req *connect.Request[pbre.ListTemplatesRequest],
) (*connect.Response[pbre.ListTemplatesResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListTemplatesInput{TenantID: tid, PageOffset: offset, PageSize: size}
	if f := req.Msg.Format; f != nil {
		v := models.ReportFormat(*f)
		in.Format = &v
	}
	rows, page, err := h.svc.ListTemplatesForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbre.ListTemplatesResponse{Page: mapper.PageResponse(page)}
	for _, t := range rows {
		resp.Templates = append(resp.Templates, mapper.TemplateToProto(t))
	}
	return connect.NewResponse(resp), nil
}

func (h *ReportsHandler) DeprecateTemplate(
	ctx context.Context, req *connect.Request[pbre.DeprecateTemplateRequest],
) (*connect.Response[pbre.DeprecateTemplateResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	t, err := h.svc.DeprecateTemplate(ctx, id, "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbre.DeprecateTemplateResponse{Template: mapper.TemplateToProto(t)}), nil
}

// ----- Report RPCs --------------------------------------------------------

func (h *ReportsHandler) GenerateReport(
	ctx context.Context, req *connect.Request[pbre.GenerateReportRequest],
) (*connect.Response[pbre.GenerateReportResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	tpid, err := parseULID(req.Msg.GetTemplateId())
	if err != nil {
		return nil, err
	}
	r, err := h.svc.GenerateReport(ctx, services.GenerateReportInput{
		TenantID:       tid,
		TemplateID:     tpid,
		ParametersJSON: req.Msg.GetParametersJson(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbre.GenerateReportResponse{Report: mapper.ReportToProto(r)}), nil
}

func (h *ReportsHandler) GetReport(
	ctx context.Context, req *connect.Request[pbre.GetReportRequest],
) (*connect.Response[pbre.GetReportResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	r, err := h.svc.GetReport(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbre.GetReportResponse{Report: mapper.ReportToProto(r)}), nil
}

func (h *ReportsHandler) ListReports(
	ctx context.Context, req *connect.Request[pbre.ListReportsRequest],
) (*connect.Response[pbre.ListReportsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListReportsInput{TenantID: tid, PageOffset: offset, PageSize: size}
	if v := req.Msg.GetTemplateId(); v != "" {
		x, err := parseULID(v)
		if err != nil {
			return nil, err
		}
		in.TemplateID = &x
	}
	if s := req.Msg.Status; s != nil {
		v := models.ReportStatus(*s)
		in.Status = &v
	}
	rows, page, err := h.svc.ListReportsForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbre.ListReportsResponse{Page: mapper.PageResponse(page)}
	for _, r := range rows {
		resp.Reports = append(resp.Reports, mapper.ReportToProto(r))
	}
	return connect.NewResponse(resp), nil
}

func (h *ReportsHandler) UpdateReportStatus(
	ctx context.Context, req *connect.Request[pbre.UpdateReportStatusRequest],
) (*connect.Response[pbre.UpdateReportStatusResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	r, err := h.svc.UpdateReportStatus(ctx, id,
		models.ReportStatus(req.Msg.GetStatus()), req.Msg.GetOutputUri(), req.Msg.GetErrorMessage(), "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbre.UpdateReportStatusResponse{Report: mapper.ReportToProto(r)}), nil
}
