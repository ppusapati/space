// Package mapper converts proto / domain / sqlc types for sat-mission.
package mapper

import (
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"p9e.in/chetana/packages/api/v1/fields"
	"p9e.in/chetana/packages/api/v1/pagination"
	"p9e.in/chetana/packages/ulid"

	pbsm "github.com/ppusapati/space/services/sat-mission/api"
	satmissiondb "github.com/ppusapati/space/services/sat-mission/db/generated"
	"github.com/ppusapati/space/services/sat-mission/internal/models"
)

// PgUUID converts a ulid.ID into a pgtype.UUID payload (16 bytes).
func PgUUID(id ulid.ID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

// FromPgUUID extracts ulid.ID from a pgtype.UUID. Returns ulid.Zero on NULL.
func FromPgUUID(p pgtype.UUID) ulid.ID {
	if !p.Valid {
		return ulid.Zero
	}
	return ulid.ID(p.Bytes)
}

// PgTimestamp converts time.Time to pgtype.Timestamptz.
func PgTimestamp(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// PageRequest extracts (offset, size) from a PaginationRequest.
func PageRequest(p *pagination.PaginationRequest) (offset, size int32) {
	if p == nil {
		return 0, 50
	}
	offset = p.GetPageOffset()
	if offset < 0 {
		offset = 0
	}
	size = p.GetPageSize()
	switch {
	case size <= 0:
		size = 50
	case size > 500:
		size = 500
	}
	return offset, size
}

// PageResponse builds a PaginationResponse.
func PageResponse(p models.Page) *pagination.PaginationResponse {
	return &pagination.PaginationResponse{
		TotalCount: p.TotalCount,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
		HasNext:    p.HasNext,
	}
}

// FieldsToProto builds an audit Fields message.
func FieldsToProto(id ulid.ID, createdBy string, createdAt time.Time, updatedBy string, updatedAt time.Time) *fields.Fields {
	out := &fields.Fields{Uuid: id.String(), IsActive: true}
	if createdBy != "" {
		out.CreatedBy = wrapperspb.String(createdBy)
	}
	if !createdAt.IsZero() {
		out.CreatedAt = timestamppb.New(createdAt)
	}
	if updatedBy != "" {
		out.UpdatedBy = wrapperspb.String(updatedBy)
	}
	if !updatedAt.IsZero() {
		out.UpdatedAt = timestamppb.New(updatedAt)
	}
	return out
}

// NormalizeJSON returns "{}" for empty / whitespace input, otherwise
// trims and returns as-is.
func NormalizeJSON(s string) string {
	t := strings.TrimSpace(s)
	if t == "" {
		return "{}"
	}
	return t
}

// ----- OrbitalState --------------------------------------------------------

// OrbitalStateFromColumns builds an OrbitalState from nilable float columns
// and an epoch timestamp.
func OrbitalStateFromColumns(rxKm, ryKm, rzKm, vxKmS, vyKmS, vzKmS *float64, epoch pgtype.Timestamptz) models.OrbitalState {
	if rxKm == nil || ryKm == nil || rzKm == nil || vxKmS == nil || vyKmS == nil || vzKmS == nil || !epoch.Valid {
		return models.OrbitalState{}
	}
	return models.OrbitalState{
		RxKm:  *rxKm,
		RyKm:  *ryKm,
		RzKm:  *rzKm,
		VxKmS: *vxKmS,
		VyKmS: *vyKmS,
		VzKmS: *vzKmS,
		Epoch: epoch.Time,
		Valid: true,
	}
}

// OrbitalStateToProto converts a domain OrbitalState to its proto form.
// Returns nil when the state is invalid.
func OrbitalStateToProto(s models.OrbitalState) *pbsm.OrbitalState {
	if !s.Valid {
		return nil
	}
	return &pbsm.OrbitalState{
		RxKm:   s.RxKm,
		RyKm:   s.RyKm,
		RzKm:   s.RzKm,
		VxKmS:  s.VxKmS,
		VyKmS:  s.VyKmS,
		VzKmS:  s.VzKmS,
		Epoch:  timestamppb.New(s.Epoch),
	}
}

// OrbitalStateFromProto converts a proto OrbitalState into the domain type.
func OrbitalStateFromProto(p *pbsm.OrbitalState) models.OrbitalState {
	if p == nil {
		return models.OrbitalState{}
	}
	out := models.OrbitalState{
		RxKm:  p.GetRxKm(),
		RyKm:  p.GetRyKm(),
		RzKm:  p.GetRzKm(),
		VxKmS: p.GetVxKmS(),
		VyKmS: p.GetVyKmS(),
		VzKmS: p.GetVzKmS(),
		Valid: true,
	}
	if e := p.GetEpoch(); e != nil {
		out.Epoch = e.AsTime()
	}
	return out
}

// ----- Satellite -----------------------------------------------------------

// SatelliteFromRow converts a sqlc Satellite row to the domain model.
func SatelliteFromRow(row satmissiondb.Satellite) *models.Satellite {
	return &models.Satellite{
		ID:                      FromPgUUID(row.ID),
		TenantID:                FromPgUUID(row.TenantID),
		Name:                    row.Name,
		NoradID:                 row.NoradID,
		InternationalDesignator: row.InternationalDesignator,
		TLELine1:                row.TleLine1,
		TLELine2:                row.TleLine2,
		CurrentMode:             models.SatelliteMode(row.CurrentMode),
		LastState: OrbitalStateFromColumns(
			row.LastStateRxKm, row.LastStateRyKm, row.LastStateRzKm,
			row.LastStateVxKmS, row.LastStateVyKmS, row.LastStateVzKmS,
			row.LastStateEpoch,
		),
		ConfigJSON: string(row.ConfigJson),
		Active:     row.Active,
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
		CreatedBy:  row.CreatedBy,
		UpdatedBy:  row.UpdatedBy,
	}
}

// SatelliteToProto converts a domain Satellite to its proto form.
func SatelliteToProto(s *models.Satellite) *pbsm.Satellite {
	if s == nil {
		return nil
	}
	return &pbsm.Satellite{
		Id:                      s.ID.String(),
		TenantId:                s.TenantID.String(),
		Name:                    s.Name,
		NoradId:                 s.NoradID,
		InternationalDesignator: s.InternationalDesignator,
		TleLine1:                s.TLELine1,
		TleLine2:                s.TLELine2,
		CurrentMode:             pbsm.SatelliteMode(s.CurrentMode),
		LastState:               OrbitalStateToProto(s.LastState),
		ConfigJson:              s.ConfigJSON,
		Active:                  s.Active,
		Fields:                  FieldsToProto(s.ID, s.CreatedBy, s.CreatedAt, s.UpdatedBy, s.UpdatedAt),
	}
}
