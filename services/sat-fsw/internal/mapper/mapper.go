// Package mapper converts proto / domain / sqlc types for sat-fsw.
package mapper

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"p9e.in/samavaya/packages/api/v1/fields"
	"p9e.in/samavaya/packages/api/v1/pagination"
	"p9e.in/samavaya/packages/ulid"

	pbfsw "github.com/ppusapati/space/services/sat-fsw/api"
	satfswdb "github.com/ppusapati/space/services/sat-fsw/db/generated"
	"github.com/ppusapati/space/services/sat-fsw/internal/models"
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

// ----- FirmwareBuild -------------------------------------------------------

// FirmwareBuildFromRow converts a sqlc row to the domain model.
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

// FirmwareBuildToProto converts a domain model to its proto form.
func FirmwareBuildToProto(b *models.FirmwareBuild) *pbfsw.FirmwareBuild {
	if b == nil {
		return nil
	}
	return &pbfsw.FirmwareBuild{
		Id:                b.ID.String(),
		TenantId:          b.TenantID.String(),
		TargetPlatform:    b.TargetPlatform,
		Subsystem:         b.Subsystem,
		Version:           b.Version,
		GitSha:            b.GitSHA,
		ArtefactUri:       b.ArtefactURI,
		ArtefactSizeBytes: b.ArtefactSizeBytes,
		ArtefactSha256:    b.ArtefactSHA256,
		Status:            pbfsw.FirmwareBuildStatus(b.Status),
		Notes:             b.Notes,
		Fields:            FieldsToProto(b.ID, b.CreatedBy, b.CreatedAt, b.UpdatedBy, b.UpdatedAt),
	}
}

// ----- DeploymentManifest --------------------------------------------------

// AssignmentsFromJSON unmarshals the assignments_json bytes into a map.
func AssignmentsFromJSON(b []byte) map[string]string {
	if len(b) == 0 {
		return map[string]string{}
	}
	out := map[string]string{}
	if err := json.Unmarshal(b, &out); err != nil {
		return map[string]string{}
	}
	return out
}

// AssignmentsToJSON marshals the assignments map. Returns "{}" for nil/empty.
func AssignmentsToJSON(m map[string]string) []byte {
	if len(m) == 0 {
		return []byte("{}")
	}
	b, err := json.Marshal(m)
	if err != nil {
		return []byte("{}")
	}
	return b
}

// DeploymentManifestFromRow converts a sqlc row to the domain model.
func DeploymentManifestFromRow(row satfswdb.DeploymentManifest) *models.DeploymentManifest {
	return &models.DeploymentManifest{
		ID:              FromPgUUID(row.ID),
		TenantID:        FromPgUUID(row.TenantID),
		SatelliteID:     FromPgUUID(row.SatelliteID),
		ManifestVersion: row.ManifestVersion,
		Status:          models.DeploymentStatus(row.Status),
		Assignments:     AssignmentsFromJSON(row.AssignmentsJson),
		Notes:           row.Notes,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		CreatedBy:       row.CreatedBy,
		UpdatedBy:       row.UpdatedBy,
	}
}

// DeploymentManifestToProto converts a domain model to its proto form.
func DeploymentManifestToProto(m *models.DeploymentManifest) *pbfsw.DeploymentManifest {
	if m == nil {
		return nil
	}
	out := &pbfsw.DeploymentManifest{
		Id:              m.ID.String(),
		TenantId:        m.TenantID.String(),
		SatelliteId:     m.SatelliteID.String(),
		ManifestVersion: m.ManifestVersion,
		Status:          pbfsw.DeploymentStatus(m.Status),
		Notes:           m.Notes,
		Fields:          FieldsToProto(m.ID, m.CreatedBy, m.CreatedAt, m.UpdatedBy, m.UpdatedAt),
	}
	if len(m.Assignments) > 0 {
		out.Assignments = make(map[string]string, len(m.Assignments))
		for k, v := range m.Assignments {
			out.Assignments[k] = v
		}
	}
	return out
}

// CleanString trims surrounding whitespace.
func CleanString(s string) string {
	return strings.TrimSpace(s)
}
