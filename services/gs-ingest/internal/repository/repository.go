// Package repository wraps the gs-ingest sqlc layer.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/samavaya/packages/ulid"

	gsingdb "github.com/ppusapati/space/services/gs-ingest/db/generated"
	"github.com/ppusapati/space/services/gs-ingest/internal/mapper"
	"github.com/ppusapati/space/services/gs-ingest/internal/models"
)

var ErrNotFound = errors.New("repository: not found")

type Repo struct {
	q    *gsingdb.Queries
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: gsingdb.New(pool), pool: pool}
}

// ----- IngestSession ------------------------------------------------------

type StartIngestSessionParams struct {
	ID          ulid.ID
	TenantID    ulid.ID
	BookingID   ulid.ID
	PassID      ulid.ID
	StationID   ulid.ID
	SatelliteID ulid.ID
	Status      models.IngestStatus
	CreatedBy   string
}

func (r *Repo) StartIngestSession(ctx context.Context, p StartIngestSessionParams) (*models.IngestSession, error) {
	row, err := r.q.StartIngestSession(ctx, gsingdb.StartIngestSessionParams{
		ID:          mapper.PgUUID(p.ID),
		TenantID:    mapper.PgUUID(p.TenantID),
		BookingID:   mapper.PgUUID(p.BookingID),
		PassID:      mapper.PgUUID(p.PassID),
		StationID:   mapper.PgUUID(p.StationID),
		SatelliteID: mapper.PgUUID(p.SatelliteID),
		Status:      int32(p.Status),
		CreatedBy:   p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.IngestSessionFromRow(row), nil
}

func (r *Repo) GetIngestSession(ctx context.Context, id ulid.ID) (*models.IngestSession, error) {
	row, err := r.q.GetIngestSession(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.IngestSessionFromRow(row), nil
}

type ListIngestSessionsParams struct {
	TenantID    ulid.ID
	StationID   *ulid.ID
	SatelliteID *ulid.ID
	Status      *models.IngestStatus
	PageOffset  int32
	PageSize    int32
}

func (r *Repo) ListIngestSessionsForTenant(ctx context.Context, p ListIngestSessionsParams) ([]*models.IngestSession, int32, error) {
	var statusPtr *int32
	if p.Status != nil {
		v := int32(*p.Status)
		statusPtr = &v
	}
	var stationPg, satellitePg pgtype.UUID
	if p.StationID != nil {
		stationPg = mapper.PgUUID(*p.StationID)
	}
	if p.SatelliteID != nil {
		satellitePg = mapper.PgUUID(*p.SatelliteID)
	}
	total, err := r.q.CountIngestSessionsForTenant(ctx, gsingdb.CountIngestSessionsForTenantParams{
		TenantID:    mapper.PgUUID(p.TenantID),
		StationID:   stationPg,
		SatelliteID: satellitePg,
		Status:      statusPtr,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListIngestSessionsForTenant(ctx, gsingdb.ListIngestSessionsForTenantParams{
		TenantID:    mapper.PgUUID(p.TenantID),
		StationID:   stationPg,
		SatelliteID: satellitePg,
		Status:      statusPtr,
		PageOffset:  p.PageOffset,
		PageSize:    p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.IngestSession, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.IngestSessionFromRow(row))
	}
	return out, int32(total), nil
}

func (r *Repo) UpdateIngestStatus(
	ctx context.Context, id ulid.ID, status models.IngestStatus, errorMessage, updatedBy string,
) (*models.IngestSession, error) {
	row, err := r.q.UpdateIngestStatus(ctx, gsingdb.UpdateIngestStatusParams{
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
	return mapper.IngestSessionFromRow(row), nil
}

// ----- DownlinkFrame ------------------------------------------------------

type RecordDownlinkFrameParams struct {
	ID               ulid.ID
	TenantID         ulid.ID
	SessionID        ulid.ID
	APID             uint32
	VirtualChannel   uint32
	SequenceCount    uint64
	GroundTime       time.Time
	PayloadSizeBytes uint64
	PayloadSHA256    string
	PayloadURI       string
	FrameType        string
	CreatedBy        string
}

// RecordDownlinkFrame writes a frame and bumps session counters in one tx.
func (r *Repo) RecordDownlinkFrame(ctx context.Context, p RecordDownlinkFrameParams) (*models.DownlinkFrame, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.q.WithTx(tx)
	row, err := q.RecordDownlinkFrame(ctx, gsingdb.RecordDownlinkFrameParams{
		ID:               mapper.PgUUID(p.ID),
		TenantID:         mapper.PgUUID(p.TenantID),
		SessionID:        mapper.PgUUID(p.SessionID),
		Apid:             int32(p.APID),
		VirtualChannel:   int32(p.VirtualChannel),
		SequenceCount:    int64(p.SequenceCount),
		GroundTime:       mapper.PgTimestamp(p.GroundTime),
		PayloadSizeBytes: int64(p.PayloadSizeBytes),
		PayloadSha256:    p.PayloadSHA256,
		PayloadUri:       p.PayloadURI,
		FrameType:        p.FrameType,
		CreatedBy:        p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	if err := q.BumpIngestCounters(ctx, gsingdb.BumpIngestCountersParams{
		ID:    mapper.PgUUID(p.SessionID),
		Bytes: int64(p.PayloadSizeBytes),
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return mapper.DownlinkFrameFromRow(row), nil
}

func (r *Repo) GetDownlinkFrame(ctx context.Context, id ulid.ID) (*models.DownlinkFrame, error) {
	row, err := r.q.GetDownlinkFrame(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.DownlinkFrameFromRow(row), nil
}

type ListDownlinkFramesParams struct {
	TenantID   ulid.ID
	SessionID  *ulid.ID
	FrameType  string
	TimeStart  time.Time
	TimeEnd    time.Time
	PageOffset int32
	PageSize   int32
}

func (r *Repo) ListDownlinkFramesForTenant(ctx context.Context, p ListDownlinkFramesParams) ([]*models.DownlinkFrame, int32, error) {
	var frameTypePtr *string
	if p.FrameType != "" {
		v := p.FrameType
		frameTypePtr = &v
	}
	var sessionPg pgtype.UUID
	if p.SessionID != nil {
		sessionPg = mapper.PgUUID(*p.SessionID)
	}
	total, err := r.q.CountDownlinkFramesForTenant(ctx, gsingdb.CountDownlinkFramesForTenantParams{
		TenantID:  mapper.PgUUID(p.TenantID),
		SessionID: sessionPg,
		FrameType: frameTypePtr,
		TimeStart: mapper.PgTimestampOrNull(p.TimeStart),
		TimeEnd:   mapper.PgTimestampOrNull(p.TimeEnd),
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListDownlinkFramesForTenant(ctx, gsingdb.ListDownlinkFramesForTenantParams{
		TenantID:   mapper.PgUUID(p.TenantID),
		SessionID:  sessionPg,
		FrameType:  frameTypePtr,
		TimeStart:  mapper.PgTimestampOrNull(p.TimeStart),
		TimeEnd:    mapper.PgTimestampOrNull(p.TimeEnd),
		PageOffset: p.PageOffset,
		PageSize:   p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.DownlinkFrame, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.DownlinkFrameFromRow(row))
	}
	return out, int32(total), nil
}
