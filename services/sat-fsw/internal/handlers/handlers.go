// Package handlers wires ConnectRPC RPCs to the sat-fsw service.
package handlers

import (
	"context"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	commonv1 "github.com/ppusapati/space/api/p9e/space/common/v1"
	satv1 "github.com/ppusapati/space/api/p9e/space/satsubsys/v1"
	"github.com/ppusapati/space/api/p9e/space/satsubsys/v1/satsubsysv1connect"
	"github.com/ppusapati/space/pkg/errs"
	"github.com/ppusapati/space/pkg/pagination"
	"github.com/ppusapati/space/pkg/validation"
	"github.com/ppusapati/space/services/sat-fsw/internal/mappers"
	"github.com/ppusapati/space/services/sat-fsw/internal/models"
	"github.com/ppusapati/space/services/sat-fsw/internal/service"
)

// FSWHandler implements satsubsysv1connect.FlightSoftwareServiceHandler.
type FSWHandler struct {
	satsubsysv1connect.UnimplementedFlightSoftwareServiceHandler
	svc          *service.FSW
	cursorSecret []byte
}

// NewFSWHandler returns a handler.
func NewFSWHandler(svc *service.FSW, cursorSecret []byte) *FSWHandler {
	return &FSWHandler{svc: svc, cursorSecret: cursorSecret}
}

// ----- FirmwareBuild -------------------------------------------------------

// RegisterFirmwareBuild implements the proto RPC.
func (h *FSWHandler) RegisterFirmwareBuild(
	ctx context.Context, req *connect.Request[satv1.RegisterFirmwareBuildRequest],
) (*connect.Response[satv1.RegisterFirmwareBuildResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	tid, err := uuid.Parse(req.Msg.GetTenantId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	b, err := h.svc.RegisterFirmwareBuild(ctx, service.RegisterFirmwareBuildInput{
		TenantID:          tid,
		TargetPlatform:    req.Msg.GetTargetPlatform(),
		Subsystem:         req.Msg.GetSubsystem(),
		Version:           req.Msg.GetVersion(),
		GitSHA:            req.Msg.GetGitSha(),
		ArtefactURI:       req.Msg.GetArtefactUri(),
		ArtefactSizeBytes: req.Msg.GetArtefactSizeBytes(),
		ArtefactSHA256:    req.Msg.GetArtefactSha256(),
		Notes:             req.Msg.GetNotes(),
	})
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&satv1.RegisterFirmwareBuildResponse{Build: mappers.FirmwareBuildToProto(b)}), nil
}

// GetFirmwareBuild implements the proto RPC.
func (h *FSWHandler) GetFirmwareBuild(
	ctx context.Context, req *connect.Request[satv1.GetFirmwareBuildRequest],
) (*connect.Response[satv1.GetFirmwareBuildResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	b, err := h.svc.GetFirmwareBuild(ctx, id)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&satv1.GetFirmwareBuildResponse{Build: mappers.FirmwareBuildToProto(b)}), nil
}

// ListFirmwareBuilds implements the proto RPC.
func (h *FSWHandler) ListFirmwareBuilds(
	ctx context.Context, req *connect.Request[satv1.ListFirmwareBuildsRequest],
) (*connect.Response[satv1.ListFirmwareBuildsResponse], error) {
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
	in := service.ListFirmwareBuildsInput{
		TenantID: tid,
		Limit:    int32(pagination.ClampPageSize(int(req.Msg.GetPage().GetPageSize()))) + 1,
		CursorID: cursor.ID,
	}
	if !cursor.CreatedAt.IsZero() {
		t := cursor.CreatedAt
		in.CursorCreated = &t
	}
	if sub := req.Msg.GetSubsystem(); sub != "" {
		in.Subsystem = &sub
	}
	if st := req.Msg.Status; st != nil {
		v := models.FirmwareBuildStatus(*st)
		in.Status = &v
	}
	rows, err := h.svc.ListFirmwareBuilds(ctx, in)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	resp := &satv1.ListFirmwareBuildsResponse{Page: &commonv1.PageResponse{}}
	page := in.Limit - 1
	if int32(len(rows)) > page {
		next := rows[page-1]
		resp.Page.NextPageToken = pagination.Encode(h.cursorSecret, pagination.Cursor{
			CreatedAt: next.CreatedAt, ID: next.ID,
		})
		rows = rows[:page]
	}
	for _, b := range rows {
		resp.Builds = append(resp.Builds, mappers.FirmwareBuildToProto(b))
	}
	return connect.NewResponse(resp), nil
}

// UpdateFirmwareBuildStatus implements the proto RPC.
func (h *FSWHandler) UpdateFirmwareBuildStatus(
	ctx context.Context, req *connect.Request[satv1.UpdateFirmwareBuildStatusRequest],
) (*connect.Response[satv1.UpdateFirmwareBuildStatusResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	b, err := h.svc.UpdateFirmwareBuildStatus(ctx, id,
		models.FirmwareBuildStatus(req.Msg.GetStatus()), req.Msg.GetNotes(), "")
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&satv1.UpdateFirmwareBuildStatusResponse{Build: mappers.FirmwareBuildToProto(b)}), nil
}

// ----- DeploymentManifest --------------------------------------------------

// CreateDeploymentManifest implements the proto RPC.
func (h *FSWHandler) CreateDeploymentManifest(
	ctx context.Context, req *connect.Request[satv1.CreateDeploymentManifestRequest],
) (*connect.Response[satv1.CreateDeploymentManifestResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	tid, err := uuid.Parse(req.Msg.GetTenantId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	sid, err := uuid.Parse(req.Msg.GetSatelliteId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	m, err := h.svc.CreateDeploymentManifest(ctx, service.CreateDeploymentManifestInput{
		TenantID:        tid,
		SatelliteID:     sid,
		ManifestVersion: req.Msg.GetManifestVersion(),
		Assignments:     req.Msg.GetAssignments(),
		Notes:           req.Msg.GetNotes(),
	})
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&satv1.CreateDeploymentManifestResponse{Manifest: mappers.DeploymentManifestToProto(m)}), nil
}

// GetDeploymentManifest implements the proto RPC.
func (h *FSWHandler) GetDeploymentManifest(
	ctx context.Context, req *connect.Request[satv1.GetDeploymentManifestRequest],
) (*connect.Response[satv1.GetDeploymentManifestResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	m, err := h.svc.GetDeploymentManifest(ctx, id)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&satv1.GetDeploymentManifestResponse{Manifest: mappers.DeploymentManifestToProto(m)}), nil
}

// ListDeploymentManifests implements the proto RPC.
func (h *FSWHandler) ListDeploymentManifests(
	ctx context.Context, req *connect.Request[satv1.ListDeploymentManifestsRequest],
) (*connect.Response[satv1.ListDeploymentManifestsResponse], error) {
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
	in := service.ListDeploymentManifestsInput{
		TenantID: tid,
		Limit:    int32(pagination.ClampPageSize(int(req.Msg.GetPage().GetPageSize()))) + 1,
		CursorID: cursor.ID,
	}
	if !cursor.CreatedAt.IsZero() {
		t := cursor.CreatedAt
		in.CursorCreated = &t
	}
	if sid := req.Msg.GetSatelliteId(); sid != "" {
		parsed, err := uuid.Parse(sid)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		in.SatelliteID = &parsed
	}
	if st := req.Msg.Status; st != nil {
		v := models.DeploymentStatus(*st)
		in.Status = &v
	}
	rows, err := h.svc.ListDeploymentManifests(ctx, in)
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	resp := &satv1.ListDeploymentManifestsResponse{Page: &commonv1.PageResponse{}}
	page := in.Limit - 1
	if int32(len(rows)) > page {
		next := rows[page-1]
		resp.Page.NextPageToken = pagination.Encode(h.cursorSecret, pagination.Cursor{
			CreatedAt: next.CreatedAt, ID: next.ID,
		})
		rows = rows[:page]
	}
	for _, m := range rows {
		resp.Manifests = append(resp.Manifests, mappers.DeploymentManifestToProto(m))
	}
	return connect.NewResponse(resp), nil
}

// UpdateDeploymentManifestStatus implements the proto RPC.
func (h *FSWHandler) UpdateDeploymentManifestStatus(
	ctx context.Context, req *connect.Request[satv1.UpdateDeploymentManifestStatusRequest],
) (*connect.Response[satv1.UpdateDeploymentManifestStatusResponse], error) {
	if err := validation.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	m, err := h.svc.UpdateDeploymentManifestStatus(ctx, id,
		models.DeploymentStatus(req.Msg.GetStatus()), req.Msg.GetNotes(), "")
	if err != nil {
		return nil, errs.ToConnect(err)
	}
	return connect.NewResponse(&satv1.UpdateDeploymentManifestStatusResponse{Manifest: mappers.DeploymentManifestToProto(m)}), nil
}
