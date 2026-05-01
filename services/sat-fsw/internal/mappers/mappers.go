// Package mappers converts proto / domain / sqlc types for sat-fsw.
package mappers

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/ppusapati/space/api/p9e/space/common/v1"
	satv1 "github.com/ppusapati/space/api/p9e/space/satsubsys/v1"
	satfswdb "github.com/ppusapati/space/services/sat-fsw/db/generated"
	"github.com/ppusapati/space/services/sat-fsw/internal/models"
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

// PgUUIDPtr converts *uuid.UUID to pgtype.UUID (Valid=false when nil).
func PgUUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

// PgTimestampPtr converts *time.Time to pgtype.Timestamptz.
func PgTimestampPtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// FirmwareBuildFromRow converts a sqlc row to a domain FirmwareBuild.
func FirmwareBuildFromRow(row satfswdb.FirmwareBuild) *models.FirmwareBuild {
	return &models.FirmwareBuild{
		ID:                FromPgUUID(row.ID),
		TenantID:          FromPgUUID(row.TenantID),
		TargetPlatform:    row.TargetPlatform,
		Subsystem:         row.Subsystem,
		Version:           row.Version,
		GitSHA:            row.GitSha,
		ArtefactURI:       row.ArtefactUri,
		ArtefactSizeBytes: uint64(row.ArtefactSizeBytes),
		ArtefactSHA256:    row.ArtefactSha256,
		Status:            models.FirmwareBuildStatus(row.Status),
		Notes:             row.Notes,
		CreatedAt:         row.CreatedAt.Time,
		UpdatedAt:         row.UpdatedAt.Time,
		CreatedBy:         row.CreatedBy,
		UpdatedBy:         row.UpdatedBy,
	}
}

// FirmwareBuildToProto converts a domain FirmwareBuild to its proto form.
func FirmwareBuildToProto(b *models.FirmwareBuild) *satv1.FirmwareBuild {
	if b == nil {
		return nil
	}
	return &satv1.FirmwareBuild{
		Id:                b.ID.String(),
		TenantId:          b.TenantID.String(),
		TargetPlatform:    b.TargetPlatform,
		Subsystem:         b.Subsystem,
		Version:           b.Version,
		GitSha:            b.GitSHA,
		ArtefactUri:       b.ArtefactURI,
		ArtefactSizeBytes: b.ArtefactSizeBytes,
		ArtefactSha256:    b.ArtefactSHA256,
		Status:            satv1.FirmwareBuildStatus(b.Status),
		Notes:             b.Notes,
		Audit: &commonv1.AuditFields{
			CreatedAt: timestamppb.New(b.CreatedAt),
			UpdatedAt: timestamppb.New(b.UpdatedAt),
			CreatedBy: b.CreatedBy,
			UpdatedBy: b.UpdatedBy,
		},
	}
}

// DeploymentManifestFromRow converts a sqlc row to a domain DeploymentManifest.
func DeploymentManifestFromRow(row satfswdb.DeploymentManifest) (*models.DeploymentManifest, error) {
	assignments := map[string]string{}
	if len(row.AssignmentsJson) > 0 {
		if err := json.Unmarshal(row.AssignmentsJson, &assignments); err != nil {
			return nil, err
		}
	}
	return &models.DeploymentManifest{
		ID:              FromPgUUID(row.ID),
		TenantID:        FromPgUUID(row.TenantID),
		SatelliteID:     FromPgUUID(row.SatelliteID),
		ManifestVersion: row.ManifestVersion,
		Status:          models.DeploymentStatus(row.Status),
		Assignments:     assignments,
		Notes:           row.Notes,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		CreatedBy:       row.CreatedBy,
		UpdatedBy:       row.UpdatedBy,
	}, nil
}

// DeploymentManifestToProto converts a domain DeploymentManifest to its proto form.
func DeploymentManifestToProto(m *models.DeploymentManifest) *satv1.DeploymentManifest {
	if m == nil {
		return nil
	}
	out := &satv1.DeploymentManifest{
		Id:              m.ID.String(),
		TenantId:        m.TenantID.String(),
		SatelliteId:     m.SatelliteID.String(),
		ManifestVersion: m.ManifestVersion,
		Status:          satv1.DeploymentStatus(m.Status),
		Assignments:     map[string]string{},
		Notes:           m.Notes,
		Audit: &commonv1.AuditFields{
			CreatedAt: timestamppb.New(m.CreatedAt),
			UpdatedAt: timestamppb.New(m.UpdatedAt),
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
	for k, v := range m.Assignments {
		out.Assignments[k] = v
	}
	return out
}

// AssignmentsToJSON marshals a map to a JSON byte slice for storage.
func AssignmentsToJSON(m map[string]string) ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}
