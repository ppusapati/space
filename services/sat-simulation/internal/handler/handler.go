// Package handler wires ConnectRPC RPCs to the sat-simulation service.
package handler

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	pbsim "github.com/ppusapati/space/services/sat-simulation/api"
	"github.com/ppusapati/space/services/sat-simulation/api/satsimulationv1connect"
	"github.com/ppusapati/space/services/sat-simulation/internal/mapper"
	"github.com/ppusapati/space/services/sat-simulation/internal/models"
	"github.com/ppusapati/space/services/sat-simulation/internal/services"
)

// SimulationHandler implements satsimulationv1connect.SimulationServiceHandler.
type SimulationHandler struct {
	satsimulationv1connect.UnimplementedSimulationServiceHandler
	svc       *services.Simulation
	validator protovalidate.Validator
}

// NewSimulationHandler returns a handler.
func NewSimulationHandler(svc *services.Simulation) (*SimulationHandler, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &SimulationHandler{svc: svc, validator: v}, nil
}

func (h *SimulationHandler) validate(msg proto.Message) error {
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

// ----- Scenario RPCs ------------------------------------------------------

// CreateScenario implements the proto RPC.
func (h *SimulationHandler) CreateScenario(
	ctx context.Context, req *connect.Request[pbsim.CreateScenarioRequest],
) (*connect.Response[pbsim.CreateScenarioResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	sc, err := h.svc.CreateScenario(ctx, services.CreateScenarioInput{
		TenantID:    tid,
		Slug:        req.Msg.GetSlug(),
		Title:       req.Msg.GetTitle(),
		Description: req.Msg.GetDescription(),
		SpecJSON:    req.Msg.GetSpecJson(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsim.CreateScenarioResponse{Scenario: mapper.ScenarioToProto(sc)}), nil
}

// GetScenario implements the proto RPC.
func (h *SimulationHandler) GetScenario(
	ctx context.Context, req *connect.Request[pbsim.GetScenarioRequest],
) (*connect.Response[pbsim.GetScenarioResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	sc, err := h.svc.GetScenario(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsim.GetScenarioResponse{Scenario: mapper.ScenarioToProto(sc)}), nil
}

// ListScenarios implements the proto RPC.
func (h *SimulationHandler) ListScenarios(
	ctx context.Context, req *connect.Request[pbsim.ListScenariosRequest],
) (*connect.Response[pbsim.ListScenariosResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	rows, page, err := h.svc.ListScenariosForTenant(ctx, tid, offset, size)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbsim.ListScenariosResponse{Page: mapper.PageResponse(page)}
	for _, sc := range rows {
		resp.Scenarios = append(resp.Scenarios, mapper.ScenarioToProto(sc))
	}
	return connect.NewResponse(resp), nil
}

// DeprecateScenario implements the proto RPC.
func (h *SimulationHandler) DeprecateScenario(
	ctx context.Context, req *connect.Request[pbsim.DeprecateScenarioRequest],
) (*connect.Response[pbsim.DeprecateScenarioResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	sc, err := h.svc.DeprecateScenario(ctx, id, "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsim.DeprecateScenarioResponse{Scenario: mapper.ScenarioToProto(sc)}), nil
}

// ----- Run RPCs -----------------------------------------------------------

// StartRun implements the proto RPC.
func (h *SimulationHandler) StartRun(
	ctx context.Context, req *connect.Request[pbsim.StartRunRequest],
) (*connect.Response[pbsim.StartRunResponse], error) {
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
	scid, err := parseULID(req.Msg.GetScenarioId())
	if err != nil {
		return nil, err
	}
	r, err := h.svc.StartRun(ctx, services.StartRunInput{
		TenantID:       tid,
		SatelliteID:    sid,
		ScenarioID:     scid,
		Mode:           models.SimulationMode(req.Msg.GetMode()),
		ParametersJSON: req.Msg.GetParametersJson(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsim.StartRunResponse{Run: mapper.RunToProto(r)}), nil
}

// GetRun implements the proto RPC.
func (h *SimulationHandler) GetRun(
	ctx context.Context, req *connect.Request[pbsim.GetRunRequest],
) (*connect.Response[pbsim.GetRunResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	r, err := h.svc.GetRun(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsim.GetRunResponse{Run: mapper.RunToProto(r)}), nil
}

// ListRuns implements the proto RPC.
func (h *SimulationHandler) ListRuns(
	ctx context.Context, req *connect.Request[pbsim.ListRunsRequest],
) (*connect.Response[pbsim.ListRunsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListRunsInput{
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
	if s := req.Msg.GetScenarioId(); s != "" {
		scid, err := parseULID(s)
		if err != nil {
			return nil, err
		}
		in.ScenarioID = &scid
	}
	if s := req.Msg.Status; s != nil {
		v := models.RunStatus(*s)
		in.Status = &v
	}
	if m := req.Msg.Mode; m != nil {
		v := models.SimulationMode(*m)
		in.Mode = &v
	}
	rows, page, err := h.svc.ListRunsForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbsim.ListRunsResponse{Page: mapper.PageResponse(page)}
	for _, r := range rows {
		resp.Runs = append(resp.Runs, mapper.RunToProto(r))
	}
	return connect.NewResponse(resp), nil
}

// UpdateRunStatus implements the proto RPC.
func (h *SimulationHandler) UpdateRunStatus(
	ctx context.Context, req *connect.Request[pbsim.UpdateRunStatusRequest],
) (*connect.Response[pbsim.UpdateRunStatusResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	r, err := h.svc.UpdateRunStatus(ctx, services.UpdateRunStatusInput{
		ID:           id,
		Status:       models.RunStatus(req.Msg.GetStatus()),
		LogURI:       req.Msg.GetLogUri(),
		TelemetryURI: req.Msg.GetTelemetryUri(),
		ResultsJSON:  req.Msg.GetResultsJson(),
		Score:        req.Msg.GetScore(),
		ErrorMessage: req.Msg.GetErrorMessage(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsim.UpdateRunStatusResponse{Run: mapper.RunToProto(r)}), nil
}

// CancelRun implements the proto RPC.
func (h *SimulationHandler) CancelRun(
	ctx context.Context, req *connect.Request[pbsim.CancelRunRequest],
) (*connect.Response[pbsim.CancelRunResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	r, err := h.svc.CancelRun(ctx, id, req.Msg.GetReason(), "")
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbsim.CancelRunResponse{Run: mapper.RunToProto(r)}), nil
}
