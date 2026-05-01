// Package mappers converts between proto, domain, and sqlc types.
package mappers

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/ppusapati/space/api/p9e/space/common/v1"
	eov1 "github.com/ppusapati/space/api/p9e/space/earthobs/v1"
	eopipelinedb "github.com/ppusapati/space/services/eo-pipeline/db/generated"
	"github.com/ppusapati/space/services/eo-pipeline/internal/models"
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

// FromPgTimestampPtr returns a *time.Time, or nil when not Valid.
func FromPgTimestampPtr(p pgtype.Timestamptz) *time.Time {
	if !p.Valid {
		return nil
	}
	t := p.Time
	return &t
}

// JobFromRow converts a sqlc Job row to a domain Job.
func JobFromRow(row eopipelinedb.Job) *models.Job {
	return &models.Job{
		ID:             FromPgUUID(row.ID),
		TenantID:       FromPgUUID(row.TenantID),
		ItemID:         FromPgUUID(row.ItemID),
		Stage:          models.JobStage(row.Stage),
		Status:         models.JobStatus(row.Status),
		ParametersJSON: string(row.ParametersJson),
		OutputURI:      row.OutputUri,
		ErrorMessage:   row.ErrorMessage,
		StartedAt:      FromPgTimestampPtr(row.StartedAt),
		FinishedAt:     FromPgTimestampPtr(row.FinishedAt),
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
		CreatedBy:      row.CreatedBy,
		UpdatedBy:      row.UpdatedBy,
	}
}

// JobToProto converts a domain Job to its proto representation.
func JobToProto(j *models.Job) *eov1.Job {
	if j == nil {
		return nil
	}
	out := &eov1.Job{
		Id:             j.ID.String(),
		TenantId:       j.TenantID.String(),
		ItemId:         j.ItemID.String(),
		Stage:          eov1.JobStage(j.Stage),
		Status:         eov1.JobStatus(j.Status),
		ParametersJson: j.ParametersJSON,
		OutputUri:      j.OutputURI,
		ErrorMessage:   j.ErrorMessage,
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
