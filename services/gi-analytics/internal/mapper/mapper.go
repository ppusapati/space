// Package mapper converts proto / domain / sqlc types for gi-analytics.
package mapper

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"p9e.in/samavaya/packages/api/v1/fields"
	"p9e.in/samavaya/packages/api/v1/pagination"
	"p9e.in/samavaya/packages/ulid"

	pban "github.com/ppusapati/space/services/gi-analytics/api"
	gianadb "github.com/ppusapati/space/services/gi-analytics/db/generated"
	"github.com/ppusapati/space/services/gi-analytics/internal/models"
)

func PgUUID(id ulid.ID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}
func FromPgUUID(p pgtype.UUID) ulid.ID {
	if !p.Valid {
		return ulid.Zero
	}
	return ulid.ID(p.Bytes)
}

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

func PageResponse(p models.Page) *pagination.PaginationResponse {
	return &pagination.PaginationResponse{
		TotalCount: p.TotalCount,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
		HasNext:    p.HasNext,
	}
}

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

func AnalysisJobFromRow(row gianadb.AnalysisJob) *models.AnalysisJob {
	j := &models.AnalysisJob{
		ID:                 FromPgUUID(row.ID),
		TenantID:           FromPgUUID(row.TenantID),
		Type:               models.AnalysisType(row.Type),
		Status:             models.AnalysisStatus(row.Status),
		InputURIs:          append([]string(nil), row.InputUris...),
		ParametersJSON:     string(row.ParametersJson),
		OutputURI:          row.OutputUri,
		ResultsSummaryJSON: string(row.ResultsSummaryJson),
		ErrorMessage:       row.ErrorMessage,
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
		CreatedBy:          row.CreatedBy,
		UpdatedBy:          row.UpdatedBy,
	}
	if row.StartedAt.Valid {
		j.StartedAt = row.StartedAt.Time
	}
	if row.FinishedAt.Valid {
		j.FinishedAt = row.FinishedAt.Time
	}
	return j
}

func AnalysisJobToProto(j *models.AnalysisJob) *pban.AnalysisJob {
	if j == nil {
		return nil
	}
	out := &pban.AnalysisJob{
		Id:                 j.ID.String(),
		TenantId:           j.TenantID.String(),
		Type:               pban.AnalysisType(j.Type),
		Status:             pban.AnalysisStatus(j.Status),
		InputUris:          append([]string(nil), j.InputURIs...),
		ParametersJson:     j.ParametersJSON,
		OutputUri:          j.OutputURI,
		ResultsSummaryJson: j.ResultsSummaryJSON,
		ErrorMessage:       j.ErrorMessage,
		Fields:             FieldsToProto(j.ID, j.CreatedBy, j.CreatedAt, j.UpdatedBy, j.UpdatedAt),
	}
	if !j.StartedAt.IsZero() {
		out.StartedAt = timestamppb.New(j.StartedAt)
	}
	if !j.FinishedAt.IsZero() {
		out.FinishedAt = timestamppb.New(j.FinishedAt)
	}
	return out
}
