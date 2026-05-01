// Package handler wires ConnectRPC RPCs to the sat-fsw service.
package handler

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	pbfsw "github.com/ppusapati/space/services/sat-fsw/api"
	"github.com/ppusapati/space/services/sat-fsw/api/satfswv1connect"
	"github.com/ppusapati/space/services/sat-fsw/internal/mapper"
	"github.com/ppusapati/space/services/sat-fsw/internal/models"
	"github.com/ppusapati/space/services/sat-fsw/internal/services"
)

// FSWHandler implements satfswv1connect.FlightSoftwareServiceHandler.
type FSWHandler struct {
	satfswv1connect.UnimplementedFlightSoftwareServiceHandler
	svc       *services.FlightSoftware
	validator protovalidate.Validator
}

// NewFSWHandler returns a handler.
func NewFSWHandler(svc *services.FlightSoftware) (*FSWHandler, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &FSWHandler{svc: svc, validator: v}, nil
}

func (h *FSWHandler) validate(msg proto.Message) error {
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

// ----- FirmwareBuild RPCs --------------------------------------------------

// RegisterFirmwareBuild implements the proto RPC.
func (h *FSWHandler) RegisterFirmwareBuild(
	ctx context.Context, req *connect.Request[pbfsw.RegisterFirmwareBuildRequest],
) (*connect.Response[pbfsw.RegisterFirmwareBuildResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	b, err := h.svc.RegisterFirmwareBuild(ctx, services.RegisterFirmwareBuildInput{
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
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbfsw.RegisterFirmwareBuildResponse{Build: mapper.FirmwareBuildToProto(b)}), nil
}

// GetFirmwareBuild implements the proto RPC.
func (h *FSWHandler) GetFirmwareBuild(
	ctx context.Context, req *connect.Request[pbfsw.GetFirmwareBuildRequest],
) (*connect.Response[pbfsw.GetFirmwareBuildResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	b, err := h.svc.GetFirmwareBuild(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbfsw.GetFirmwareBuildResponse{Build: mapper.FirmwareBuildToProto(b)}), nil
}

// ListFirmwareBuilds implements the proto RPC.
func (h *FSWHandler) ListFirmwareBuilds(
	ctx context.Context, req *connect.Request[pbfsw.ListFirmwareBuildsRequest],
) (*connect.Response[pbfsw.ListFirmwareBuildsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListFirmwareBuildsInput{
		TenantID:   tid,
		Subsystem:  req.Msg.GetSubsystem(),
		PageOffset: offset,
		PageSize:   size,
	}
	if s := req.Msg.Status; s != nil {
		v := models.FirmwareBuildStatus(*s)
		in.Status = &v
	}
	rows, page, err := h.svc.ListFirmwareBuildsForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbfsw.ListFirmwareBuildsResponse{Page: mapper.PageResponse(page)}
	for _, b := range rows {
		resp.Builds = append(resp.Builds, mapper.FirmwareBuildToProto(b))
	}
	return connect.NewResponse(resp), nil
}

// UpdateFirmwareBuildStatus implements the proto RPC.
func (h *FSWHandler) UpdateFirmwareBuildStatus(
	ctx context.Context, req *connect.Request[pbfsw.UpdateFirmwareBuildStatusRequest],
) (*connect.Response[pbfsw.UpdateFirmwareBuildStatusResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	b, err := h.svc.UpdateFirmwareBuildStatus(ctx, id,
		models.FirmwareBuildStatus(req.Msg.GetStatus()), req.Msg.GetNotes(), "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbfsw.UpdateFirmwareBuildStatusResponse{Build: mapper.FirmwareBuildToProto(b)}), nil
}

// ----- DeploymentManifest RPCs ---------------------------------------------

// CreateDeploymentManifest implements the proto RPC.
func (h *FSWHandler) CreateDeploymentManifest(
	ctx context.Context, req *connect.Request[pbfsw.CreateDeploymentManifestRequest],
) (*connect.Response[pbfsw.CreateDeploymentManifestResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	sid, err := parseULID(req.Msg.GetSatelliteId())
	if err != nil {
		return nil, err
	}
	m, err := h.svc.CreateDeploymentManifest(ctx, services.CreateDeploymentManifestInput{
		TenantID:        tid,
		SatelliteID:     sid,
		ManifestVersion: req.Msg.GetManifestVersion(),
		Assignments:     req.Msg.GetAssignments(),
		Notes:           req.Msg.GetNotes(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbfsw.CreateDeploymentManifestResponse{Manifest: mapper.DeploymentManifestToProto(m)}), nil
}

// GetDeploymentManifest implements the proto RPC.
func (h *FSWHandler) GetDeploymentManifest(
	ctx context.Context, req *connect.Request[pbfsw.GetDeploymentManifestRequest],
) (*connect.Response[pbfsw.GetDeploymentManifestResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	m, err := h.svc.GetDeploymentManifest(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbfsw.GetDeploymentManifestResponse{Manifest: mapper.DeploymentManifestToProto(m)}), nil
}

// ListDeploymentManifests implements the proto RPC.
func (h *FSWHandler) ListDeploymentManifests(
	ctx context.Context, req *connect.Request[pbfsw.ListDeploymentManifestsRequest],
) (*connect.Response[pbfsw.ListDeploymentManifestsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListDeploymentManifestsInput{
		TenantID:   tid,
		PageOffset: offset,
		PageSize:   size,
	}
	if s := req.Msg.GetSatelliteId(); s != "" {
		sid, err := parseULID(s)
		if err != nil {
			return nil, err
		}
		in.SatelliteID = &sid
	}
	if s := req.Msg.Status; s != nil {
		v := models.DeploymentStatus(*s)
		in.Status = &v
	}
	rows, page, err := h.svc.ListDeploymentManifestsForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbfsw.ListDeploymentManifestsResponse{Page: mapper.PageResponse(page)}
	for _, m := range rows {
		resp.Manifests = append(resp.Manifests, mapper.DeploymentManifestToProto(m))
	}
	return connect.NewResponse(resp), nil
}

// UpdateDeploymentManifestStatus implements the proto RPC.
func (h *FSWHandler) UpdateDeploymentManifestStatus(
	ctx context.Context, req *connect.Request[pbfsw.UpdateDeploymentManifestStatusRequest],
) (*connect.Response[pbfsw.UpdateDeploymentManifestStatusResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	m, err := h.svc.UpdateDeploymentManifestStatus(ctx, id,
		models.DeploymentStatus(req.Msg.GetStatus()), req.Msg.GetNotes(), "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbfsw.UpdateDeploymentManifestStatusResponse{Manifest: mapper.DeploymentManifestToProto(m)}), nil
}
