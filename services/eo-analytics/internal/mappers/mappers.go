// Package mappers converts proto / domain / sqlc types.
package mappers

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/ppusapati/space/api/p9e/space/common/v1"
	eov1 "github.com/ppusapati/space/api/p9e/space/earthobs/v1"
	eoanalyticsdb "github.com/ppusapati/space/services/eo-analytics/db/generated"
	"github.com/ppusapati/space/services/eo-analytics/internal/models"
)

// PgUUID wraps uuid.UUID for sqlc.
func PgUUID(id uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: id, Valid: true} }

// FromPgUUID returns the uuid.UUID portion of a pgtype.UUID.
func FromPgUUID(p pgtype.UUID) uuid.UUID {
	if !p.Valid {
		return uuid.Nil
	}
	return p.Bytes
}

// FromPgTimestampPtr returns *time.Time from a pgtype.Timestamptz.
func FromPgTimestampPtr(p pgtype.Timestamptz) *time.Time {
	if !p.Valid {
		return nil
	}
	t := p.Time
	return &t
}

// PgTimestampPtr converts *time.Time to pgtype.Timestamptz.
func PgTimestampPtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// ----- Model -------------------------------------------------------

// ModelFromRow converts a sqlc Model row to a domain Model.
func ModelFromRow(row eoanalyticsdb.Model) *models.Model {
	return &models.Model{
		ID:           FromPgUUID(row.ID),
		TenantID:     FromPgUUID(row.TenantID),
		Name:         row.Name,
		Version:      row.Version,
		Task:         models.InferenceTask(row.Task),
		Framework:    row.Framework,
		ArtefactURI:  row.ArtefactUri,
		MetadataJSON: string(row.MetadataJson),
		Active:       row.Active,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
		CreatedBy:    row.CreatedBy,
		UpdatedBy:    row.UpdatedBy,
	}
}

// ModelToProto converts domain → proto.
func ModelToProto(m *models.Model) *eov1.Model {
	if m == nil {
		return nil
	}
	return &eov1.Model{
		Id:           m.ID.String(),
		TenantId:     m.TenantID.String(),
		Name:         m.Name,
		Version:      m.Version,
		Task:         eov1.InferenceTask(m.Task),
		Framework:    m.Framework,
		ArtefactUri:  m.ArtefactURI,
		MetadataJson: m.MetadataJSON,
		Active:       m.Active,
		Audit: &commonv1.AuditFields{
			CreatedAt: timestamppb.New(m.CreatedAt),
			UpdatedAt: timestamppb.New(m.UpdatedAt),
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
}

// ----- InferenceJob ------------------------------------------------

// InferenceJobFromRow converts a sqlc row to a domain InferenceJob.
func InferenceJobFromRow(row eoanalyticsdb.InferenceJob) *models.InferenceJob {
	return &models.InferenceJob{
		ID:           FromPgUUID(row.ID),
		TenantID:     FromPgUUID(row.TenantID),
		ModelID:      FromPgUUID(row.ModelID),
		ItemID:       FromPgUUID(row.ItemID),
		Status:       models.InferenceJobStatus(row.Status),
		OutputURI:    row.OutputUri,
		ErrorMessage: row.ErrorMessage,
		StartedAt:    FromPgTimestampPtr(row.StartedAt),
		FinishedAt:   FromPgTimestampPtr(row.FinishedAt),
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
		CreatedBy:    row.CreatedBy,
		UpdatedBy:    row.UpdatedBy,
	}
}

// InferenceJobToProto converts domain → proto.
func InferenceJobToProto(j *models.InferenceJob) *eov1.InferenceJob {
	if j == nil {
		return nil
	}
	out := &eov1.InferenceJob{
		Id:           j.ID.String(),
		TenantId:     j.TenantID.String(),
		ModelId:      j.ModelID.String(),
		ItemId:       j.ItemID.String(),
		Status:       eov1.InferenceJobStatus(j.Status),
		OutputUri:    j.OutputURI,
		ErrorMessage: j.ErrorMessage,
		Audit: &commonv1.AuditFields{
			CreatedAt: timestamppb.New(j.CreatedAt),
			UpdatedAt: timestamppb.New(j.UpdatedAt),
			CreatedBy: j.CreatedBy,
			UpdatedBy: j.UpdatedBy,
		},
	}
	if j.StartedAt != nil {
		out.StartedAt = timestamppb.New(*j.StartedAt)
	}
	if j.FinishedAt != nil {
		out.FinishedAt = timestamppb.New(*j.FinishedAt)
	}
	return out
}
