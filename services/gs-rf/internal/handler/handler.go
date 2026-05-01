// Package handler wires ConnectRPC RPCs to the gs-rf service.
package handler

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	pbrf "github.com/ppusapati/space/services/gs-rf/api"
	"github.com/ppusapati/space/services/gs-rf/api/gsrfv1connect"
	"github.com/ppusapati/space/services/gs-rf/internal/mapper"
	"github.com/ppusapati/space/services/gs-rf/internal/services"
)

type RFHandler struct {
	gsrfv1connect.UnimplementedRFServiceHandler
	svc       *services.RF
	validator protovalidate.Validator
}

func NewRFHandler(svc *services.RF) (*RFHandler, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &RFHandler{svc: svc, validator: v}, nil
}

func (h *RFHandler) validate(msg proto.Message) error {
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

// ----- LinkBudget RPCs ----------------------------------------------------

func (h *RFHandler) CreateLinkBudget(
	ctx context.Context, req *connect.Request[pbrf.CreateLinkBudgetRequest],
) (*connect.Response[pbrf.CreateLinkBudgetResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	pid, err := parseULID(req.Msg.GetPassId())
	if err != nil {
		return nil, err
	}
	stid, err := parseULID(req.Msg.GetStationId())
	if err != nil {
		return nil, err
	}
	aid, err := parseULID(req.Msg.GetAntennaId())
	if err != nil {
		return nil, err
	}
	satid, err := parseULID(req.Msg.GetSatelliteId())
	if err != nil {
		return nil, err
	}
	b, err := h.svc.CreateLinkBudget(ctx, services.CreateLinkBudgetInput{
		TenantID:           tid,
		PassID:             pid,
		StationID:          stid,
		AntennaID:          aid,
		SatelliteID:        satid,
		CarrierFreqHz:      req.Msg.GetCarrierFreqHz(),
		TxPowerDBM:         req.Msg.GetTxPowerDbm(),
		TxGainDBI:          req.Msg.GetTxGainDbi(),
		RxGainDBI:          req.Msg.GetRxGainDbi(),
		RxNoiseTempK:       req.Msg.GetRxNoiseTempK(),
		BandwidthHz:        req.Msg.GetBandwidthHz(),
		SlantRangeKm:       req.Msg.GetSlantRangeKm(),
		FreeSpaceLossDB:    req.Msg.GetFreeSpaceLossDb(),
		AtmosphericLossDB:  req.Msg.GetAtmosphericLossDb(),
		PolarizationLossDB: req.Msg.GetPolarizationLossDb(),
		PointingLossDB:     req.Msg.GetPointingLossDb(),
		PredictedEbN0DB:    req.Msg.GetPredictedEbN0Db(),
		PredictedSNRDB:     req.Msg.GetPredictedSnrDb(),
		LinkMarginDB:       req.Msg.GetLinkMarginDb(),
		Notes:              req.Msg.GetNotes(),
	})
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbrf.CreateLinkBudgetResponse{Budget: mapper.LinkBudgetToProto(b)}), nil
}

func (h *RFHandler) GetLinkBudget(
	ctx context.Context, req *connect.Request[pbrf.GetLinkBudgetRequest],
) (*connect.Response[pbrf.GetLinkBudgetResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	b, err := h.svc.GetLinkBudget(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbrf.GetLinkBudgetResponse{Budget: mapper.LinkBudgetToProto(b)}), nil
}

func (h *RFHandler) ListLinkBudgets(
	ctx context.Context, req *connect.Request[pbrf.ListLinkBudgetsRequest],
) (*connect.Response[pbrf.ListLinkBudgetsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListLinkBudgetsInput{
		TenantID:   tid,
		PageOffset: offset,
		PageSize:   size,
	}
	if s := req.Msg.GetPassId(); s != "" {
		v, err := parseULID(s)
		if err != nil {
			return nil, err
		}
		in.PassID = &v
	}
	if s := req.Msg.GetStationId(); s != "" {
		v, err := parseULID(s)
		if err != nil {
			return nil, err
		}
		in.StationID = &v
	}
	if s := req.Msg.GetSatelliteId(); s != "" {
		v, err := parseULID(s)
		if err != nil {
			return nil, err
		}
		in.SatelliteID = &v
	}
	rows, page, err := h.svc.ListLinkBudgetsForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbrf.ListLinkBudgetsResponse{Page: mapper.PageResponse(page)}
	for _, b := range rows {
		resp.Budgets = append(resp.Budgets, mapper.LinkBudgetToProto(b))
	}
	return connect.NewResponse(resp), nil
}

// ----- LinkMeasurement RPCs -----------------------------------------------

func (h *RFHandler) RecordMeasurement(
	ctx context.Context, req *connect.Request[pbrf.RecordMeasurementRequest],
) (*connect.Response[pbrf.RecordMeasurementResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	pid, err := parseULID(req.Msg.GetPassId())
	if err != nil {
		return nil, err
	}
	stid, err := parseULID(req.Msg.GetStationId())
	if err != nil {
		return nil, err
	}
	aid, err := parseULID(req.Msg.GetAntennaId())
	if err != nil {
		return nil, err
	}
	in := services.RecordMeasurementInput{
		TenantID:       tid,
		PassID:         pid,
		StationID:      stid,
		AntennaID:      aid,
		RSSIDBM:        req.Msg.GetRssiDbm(),
		SNRDB:          req.Msg.GetSnrDb(),
		BER:            req.Msg.GetBer(),
		FER:            req.Msg.GetFer(),
		FrequencyHz:    req.Msg.GetFrequencyHz(),
		DopplerShiftHz: req.Msg.GetDopplerShiftHz(),
	}
	if t := req.Msg.GetSampledAt(); t != nil {
		in.SampledAt = t.AsTime()
	}
	m, err := h.svc.RecordMeasurement(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbrf.RecordMeasurementResponse{Measurement: mapper.MeasurementToProto(m)}), nil
}

func (h *RFHandler) GetMeasurement(
	ctx context.Context, req *connect.Request[pbrf.GetMeasurementRequest],
) (*connect.Response[pbrf.GetMeasurementResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	id, err := parseULID(req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	m, err := h.svc.GetMeasurement(ctx, id)
	if err != nil {
		return nil, toConnect(err)
	}
	return connect.NewResponse(&pbrf.GetMeasurementResponse{Measurement: mapper.MeasurementToProto(m)}), nil
}

func (h *RFHandler) ListMeasurements(
	ctx context.Context, req *connect.Request[pbrf.ListMeasurementsRequest],
) (*connect.Response[pbrf.ListMeasurementsResponse], error) {
	if err := h.validate(req.Msg); err != nil {
		return nil, err
	}
	tid, err := parseULID(req.Msg.GetTenantId())
	if err != nil {
		return nil, err
	}
	offset, size := mapper.PageRequest(req.Msg.GetPage())
	in := services.ListMeasurementsInput{
		TenantID:   tid,
		PageOffset: offset,
		PageSize:   size,
	}
	if s := req.Msg.GetPassId(); s != "" {
		v, err := parseULID(s)
		if err != nil {
			return nil, err
		}
		in.PassID = &v
	}
	if s := req.Msg.GetStationId(); s != "" {
		v, err := parseULID(s)
		if err != nil {
			return nil, err
		}
		in.StationID = &v
	}
	if t := req.Msg.GetTimeStart(); t != nil {
		in.TimeStart = t.AsTime()
	}
	if t := req.Msg.GetTimeEnd(); t != nil {
		in.TimeEnd = t.AsTime()
	}
	rows, page, err := h.svc.ListMeasurementsForTenant(ctx, in)
	if err != nil {
		return nil, toConnect(err)
	}
	resp := &pbrf.ListMeasurementsResponse{Page: mapper.PageResponse(page)}
	for _, m := range rows {
		resp.Measurements = append(resp.Measurements, mapper.MeasurementToProto(m))
	}
	return connect.NewResponse(resp), nil
}
