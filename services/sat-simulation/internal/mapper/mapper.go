// Package mapper converts proto / domain / sqlc types for sat-simulation.
package mapper

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"p9e.in/chetana/packages/api/v1/fields"
	"p9e.in/chetana/packages/api/v1/pagination"
	"p9e.in/chetana/packages/ulid"

	pbsim "github.com/ppusapati/space/services/sat-simulation/api"
	satsimdb "github.com/ppusapati/space/services/sat-simulation/db/generated"
	"github.com/ppusapati/space/services/sat-simulation/internal/models"
)

// PgUUID converts a ulid.ID into a pgtype.UUID payload (16 bytes).
func PgUUID(id ulid.ID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

// PgUUIDOrNull treats ulid.Zero as NULL.
func PgUUIDOrNull(id ulid.ID) pgtype.UUID {
	if id.IsZero() {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

// FromPgUUID extracts ulid.ID from pgtype.UUID. Returns ulid.Zero on NULL.
func FromPgUUID(p pgtype.UUID) ulid.ID {
	if !p.Valid {
		return ulid.Zero
	}
	return ulid.ID(p.Bytes)
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

// ----- Scenario ------------------------------------------------------------

// ScenarioFromRow converts a sqlc row to the domain model.
func ScenarioFromRow(row satsimdb.Scenario) *models.Scenario {
	return &models.Scenario{
		ID:          FromPgUUID(row.ID),
		TenantID:    FromPgUUID(row.TenantID),
		Slug:        row.Slug,
		Title:       row.Title,
		Description: row.Description,
		SpecJSON:    string(row.SpecJson),
		Active:      row.Active,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
		CreatedBy:   row.CreatedBy,
		UpdatedBy:   row.UpdatedBy,
	}
}

// ScenarioToProto converts a domain Scenario to its proto form.
func ScenarioToProto(s *models.Scenario) *pbsim.Scenario {
	if s == nil {
		return nil
	}
	return &pbsim.Scenario{
		Id:          s.ID.String(),
		TenantId:    s.TenantID.String(),
		Slug:        s.Slug,
		Title:       s.Title,
		Description: s.Description,
		SpecJson:    s.SpecJSON,
		Active:      s.Active,
		Fields:      FieldsToProto(s.ID, s.CreatedBy, s.CreatedAt, s.UpdatedBy, s.UpdatedAt),
	}
}

// ----- SimulationRun -------------------------------------------------------

// RunFromRow converts a sqlc row to the domain model.
func RunFromRow(row satsimdb.SimulationRun) *models.SimulationRun {
	r := &models.SimulationRun{
		ID:             FromPgUUID(row.ID),
		TenantID:       FromPgUUID(row.TenantID),
		SatelliteID:    FromPgUUID(row.SatelliteID),
		ScenarioID:     FromPgUUID(row.ScenarioID),
		Mode:           models.SimulationMode(row.Mode),
		Status:         models.RunStatus(row.Status),
		ParametersJSON: string(row.ParametersJson),
		LogURI:         row.LogUri,
		TelemetryURI:   row.TelemetryUri,
		ResultsJSON:    string(row.ResultsJson),
		Score:          row.Score,
		ErrorMessage:   row.ErrorMessage,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
		CreatedBy:      row.CreatedBy,
		UpdatedBy:      row.UpdatedBy,
	}
	if row.StartedAt.Valid {
		r.StartedAt = row.StartedAt.Time
	}
	if row.FinishedAt.Valid {
		r.FinishedAt = row.FinishedAt.Time
	}
	return r
}

// RunToProto converts a domain SimulationRun to its proto form.
func RunToProto(r *models.SimulationRun) *pbsim.SimulationRun {
	if r == nil {
		return nil
	}
	out := &pbsim.SimulationRun{
		Id:             r.ID.String(),
		TenantId:       r.TenantID.String(),
		SatelliteId:    r.SatelliteID.String(),
		ScenarioId:     r.ScenarioID.String(),
		Mode:           pbsim.SimulationMode(r.Mode),
		Status:         pbsim.RunStatus(r.Status),
		ParametersJson: r.ParametersJSON,
		LogUri:         r.LogURI,
		TelemetryUri:   r.TelemetryURI,
		ResultsJson:    r.ResultsJSON,
		Score:          r.Score,
		ErrorMessage:   r.ErrorMessage,
		Fields:         FieldsToProto(r.ID, r.CreatedBy, r.CreatedAt, r.UpdatedBy, r.UpdatedAt),
	}
	if !r.StartedAt.IsZero() {
		out.StartedAt = timestamppb.New(r.StartedAt)
	}
	if !r.FinishedAt.IsZero() {
		out.FinishedAt = timestamppb.New(r.FinishedAt)
	}
	return out
}
