// Package mappers converts proto / domain / sqlc types for sat-mission.
package mappers

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/ppusapati/space/api/p9e/space/common/v1"
	satv1 "github.com/ppusapati/space/api/p9e/space/satsubsys/v1"
	satmissiondb "github.com/ppusapati/space/services/sat-mission/db/generated"
	"github.com/ppusapati/space/services/sat-mission/internal/models"
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

// PgTimestampPtr converts *time.Time to pgtype.Timestamptz.
func PgTimestampPtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// FloatPtr returns &v when the row column is set; nil otherwise.
func FloatPtr(set bool, v float64) *float64 {
	if !set {
		return nil
	}
	return &v
}

// SatelliteFromRow converts a sqlc row to a domain Satellite.
func SatelliteFromRow(row satmissiondb.Satellite) *models.Satellite {
	state := models.OrbitalState{}
	if row.LastStateRxKm != nil {
		state.RxKm = *row.LastStateRxKm
		state.RyKm = derefFloat(row.LastStateRyKm)
		state.RzKm = derefFloat(row.LastStateRzKm)
		state.VxKmS = derefFloat(row.LastStateVxKmS)
		state.VyKmS = derefFloat(row.LastStateVyKmS)
		state.VzKmS = derefFloat(row.LastStateVzKmS)
		if row.LastStateEpoch.Valid {
			state.Epoch = row.LastStateEpoch.Time
		}
		state.Valid = true
	}
	return &models.Satellite{
		ID:                      FromPgUUID(row.ID),
		TenantID:                FromPgUUID(row.TenantID),
		Name:                    row.Name,
		NORADID:                 row.NoradID,
		InternationalDesignator: row.InternationalDesignator,
		TLELine1:                row.TleLine1,
		TLELine2:                row.TleLine2,
		CurrentMode:             models.SatelliteMode(row.CurrentMode),
		LastState:               state,
		ConfigJSON:              string(row.ConfigJson),
		Active:                  row.Active,
		CreatedAt:               row.CreatedAt.Time,
		UpdatedAt:               row.UpdatedAt.Time,
		CreatedBy:               row.CreatedBy,
		UpdatedBy:               row.UpdatedBy,
	}
}

// SatelliteToProto converts a domain Satellite to its proto.
func SatelliteToProto(s *models.Satellite) *satv1.Satellite {
	if s == nil {
		return nil
	}
	out := &satv1.Satellite{
		Id:                      s.ID.String(),
		TenantId:                s.TenantID.String(),
		Name:                    s.Name,
		NoradId:                 s.NORADID,
		InternationalDesignator: s.InternationalDesignator,
		TleLine1:                s.TLELine1,
		TleLine2:                s.TLELine2,
		CurrentMode:             satv1.SatelliteMode(s.CurrentMode),
		ConfigJson:              s.ConfigJSON,
		Active:                  s.Active,
		Audit: &commonv1.AuditFields{
			CreatedAt: timestamppb.New(s.CreatedAt),
			UpdatedAt: timestamppb.New(s.UpdatedAt),
			CreatedBy: s.CreatedBy,
			UpdatedBy: s.UpdatedBy,
		},
	}
	if s.LastState.Valid {
		out.LastState = &satv1.OrbitalState{
			RxKm:   s.LastState.RxKm,
			RyKm:   s.LastState.RyKm,
			RzKm:   s.LastState.RzKm,
			VxKmS:  s.LastState.VxKmS,
			VyKmS:  s.LastState.VyKmS,
			VzKmS:  s.LastState.VzKmS,
			Epoch:  timestamppb.New(s.LastState.Epoch),
		}
	}
	return out
}

// OrbitalStateFromProto converts proto → domain. nil ⇒ invalid state.
func OrbitalStateFromProto(p *satv1.OrbitalState) models.OrbitalState {
	if p == nil {
		return models.OrbitalState{}
	}
	state := models.OrbitalState{
		RxKm:  p.RxKm,
		RyKm:  p.RyKm,
		RzKm:  p.RzKm,
		VxKmS: p.VxKmS,
		VyKmS: p.VyKmS,
		VzKmS: p.VzKmS,
		Valid: true,
	}
	if p.Epoch != nil {
		state.Epoch = p.Epoch.AsTime()
	}
	return state
}

func derefFloat(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
