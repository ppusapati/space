// Package repository wraps the gs-scheduler sqlc layer.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/samavaya/packages/ulid"

	gsschdb "github.com/ppusapati/space/services/gs-scheduler/db/generated"
	"github.com/ppusapati/space/services/gs-scheduler/internal/mapper"
	"github.com/ppusapati/space/services/gs-scheduler/internal/models"
)

var ErrNotFound = errors.New("repository: not found")

type Repo struct {
	q    *gsschdb.Queries
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: gsschdb.New(pool), pool: pool}
}

// ----- ContactPass --------------------------------------------------------

type InsertContactPassParams struct {
	ID              ulid.ID
	TenantID        ulid.ID
	StationID       ulid.ID
	SatelliteID     ulid.ID
	AOSTime         time.Time
	TCATime         time.Time
	LOSTime         time.Time
	MaxElevationDeg float64
	AOSAzimuthDeg   float64
	LOSAzimuthDeg   float64
	Source          string
	CreatedBy       string
}

func (r *Repo) InsertContactPass(ctx context.Context, p InsertContactPassParams) (*models.ContactPass, error) {
	row, err := r.q.InsertContactPass(ctx, gsschdb.InsertContactPassParams{
		ID:              mapper.PgUUID(p.ID),
		TenantID:        mapper.PgUUID(p.TenantID),
		StationID:       mapper.PgUUID(p.StationID),
		SatelliteID:     mapper.PgUUID(p.SatelliteID),
		AosTime:         mapper.PgTimestamp(p.AOSTime),
		TcaTime:         mapper.PgTimestamp(p.TCATime),
		LosTime:         mapper.PgTimestamp(p.LOSTime),
		MaxElevationDeg: p.MaxElevationDeg,
		AosAzimuthDeg:   p.AOSAzimuthDeg,
		LosAzimuthDeg:   p.LOSAzimuthDeg,
		Source:          p.Source,
		CreatedBy:       p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.ContactPassFromRow(row), nil
}

func (r *Repo) GetContactPass(ctx context.Context, id ulid.ID) (*models.ContactPass, error) {
	row, err := r.q.GetContactPass(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.ContactPassFromRow(row), nil
}

type ListContactPassesParams struct {
	TenantID        ulid.ID
	StationID       *ulid.ID
	SatelliteID     *ulid.ID
	AOSStart        time.Time
	AOSEnd          time.Time
	MinElevationDeg *float64
	PageOffset      int32
	PageSize        int32
}

func (r *Repo) ListContactPassesForTenant(ctx context.Context, p ListContactPassesParams) ([]*models.ContactPass, int32, error) {
	var stationPg, satellitePg pgtype.UUID
	if p.StationID != nil {
		stationPg = mapper.PgUUID(*p.StationID)
	}
	if p.SatelliteID != nil {
		satellitePg = mapper.PgUUID(*p.SatelliteID)
	}
	total, err := r.q.CountContactPassesForTenant(ctx, gsschdb.CountContactPassesForTenantParams{
		TenantID:        mapper.PgUUID(p.TenantID),
		StationID:       stationPg,
		SatelliteID:     satellitePg,
		AosStart:        mapper.PgTimestampOrNull(p.AOSStart),
		AosEnd:          mapper.PgTimestampOrNull(p.AOSEnd),
		MinElevationDeg: p.MinElevationDeg,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListContactPassesForTenant(ctx, gsschdb.ListContactPassesForTenantParams{
		TenantID:        mapper.PgUUID(p.TenantID),
		StationID:       stationPg,
		SatelliteID:     satellitePg,
		AosStart:        mapper.PgTimestampOrNull(p.AOSStart),
		AosEnd:          mapper.PgTimestampOrNull(p.AOSEnd),
		MinElevationDeg: p.MinElevationDeg,
		PageOffset:      p.PageOffset,
		PageSize:        p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.ContactPass, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.ContactPassFromRow(row))
	}
	return out, int32(total), nil
}

// ----- Booking ------------------------------------------------------------

type RequestBookingParams struct {
	ID        ulid.ID
	TenantID  ulid.ID
	PassID    ulid.ID
	Priority  int32
	Status    models.BookingStatus
	Purpose   string
	Notes     string
	CreatedBy string
}

func (r *Repo) RequestBooking(ctx context.Context, p RequestBookingParams) (*models.Booking, error) {
	row, err := r.q.RequestBooking(ctx, gsschdb.RequestBookingParams{
		ID:        mapper.PgUUID(p.ID),
		TenantID:  mapper.PgUUID(p.TenantID),
		PassID:    mapper.PgUUID(p.PassID),
		Priority:  p.Priority,
		Status:    int32(p.Status),
		Purpose:   p.Purpose,
		Notes:     p.Notes,
		CreatedBy: p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.BookingFromRow(row), nil
}

func (r *Repo) GetBooking(ctx context.Context, id ulid.ID) (*models.Booking, error) {
	row, err := r.q.GetBooking(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.BookingFromRow(row), nil
}

type ListBookingsParams struct {
	TenantID   ulid.ID
	Status     *models.BookingStatus
	PageOffset int32
	PageSize   int32
}

func (r *Repo) ListBookingsForTenant(ctx context.Context, p ListBookingsParams) ([]*models.Booking, int32, error) {
	var statusPtr *int32
	if p.Status != nil {
		v := int32(*p.Status)
		statusPtr = &v
	}
	total, err := r.q.CountBookingsForTenant(ctx, gsschdb.CountBookingsForTenantParams{
		TenantID: mapper.PgUUID(p.TenantID),
		Status:   statusPtr,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListBookingsForTenant(ctx, gsschdb.ListBookingsForTenantParams{
		TenantID:   mapper.PgUUID(p.TenantID),
		Status:     statusPtr,
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.Booking, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.BookingFromRow(row))
	}
	return out, int32(total), nil
}

func (r *Repo) UpdateBookingStatus(
	ctx context.Context, id ulid.ID, status models.BookingStatus, errorMessage, updatedBy string,
) (*models.Booking, error) {
	row, err := r.q.UpdateBookingStatus(ctx, gsschdb.UpdateBookingStatusParams{
		ID:           mapper.PgUUID(id),
		Status:       int32(status),
		ErrorMessage: errorMessage,
		UpdatedBy:    updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.BookingFromRow(row), nil
}
