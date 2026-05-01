// Package repository wraps the sat-telemetry sqlc layer.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	sattlmdb "github.com/ppusapati/space/services/sat-telemetry/db/generated"
	"github.com/ppusapati/space/services/sat-telemetry/internal/mappers"
	"github.com/ppusapati/space/services/sat-telemetry/internal/models"
)

// ErrNotFound is returned when no row matches.
var ErrNotFound = errors.New("repository: not found")

// Repo persists Channels, Frames, and Samples.
type Repo struct {
	q    *sattlmdb.Queries
	pool *pgxpool.Pool
}

// New constructs a Repo.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: sattlmdb.New(pool), pool: pool}
}

// ----- Channels ------------------------------------------------------------

// DefineChannelParams is the input to DefineChannel.
type DefineChannelParams struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	SatelliteID  uuid.UUID
	Subsystem    string
	Name         string
	Units        string
	ValueType    models.ChannelValueType
	MinValue     float64
	MaxValue     float64
	SampleRateHz float64
	CreatedBy    string
}

// DefineChannel inserts a new channel row.
func (r *Repo) DefineChannel(ctx context.Context, p DefineChannelParams) (*models.Channel, error) {
	row, err := r.q.DefineChannel(ctx, sattlmdb.DefineChannelParams{
		ID:           mappers.PgUUID(p.ID),
		TenantID:     mappers.PgUUID(p.TenantID),
		SatelliteID:  mappers.PgUUID(p.SatelliteID),
		Subsystem:    p.Subsystem,
		Name:         p.Name,
		Units:        p.Units,
		ValueType:    int32(p.ValueType),
		MinValue:     p.MinValue,
		MaxValue:     p.MaxValue,
		SampleRateHz: p.SampleRateHz,
		CreatedBy:    p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mappers.ChannelFromRow(row), nil
}

// GetChannel returns a channel by id.
func (r *Repo) GetChannel(ctx context.Context, id uuid.UUID) (*models.Channel, error) {
	row, err := r.q.GetChannel(ctx, mappers.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mappers.ChannelFromRow(row), nil
}

// ListChannelsParams is the input to ListChannels.
type ListChannelsParams struct {
	TenantID      uuid.UUID
	SatelliteID   *uuid.UUID
	Subsystem     *string
	CursorCreated *time.Time
	CursorID      uuid.UUID
	Limit         int32
}

// ListChannels returns one page of channels.
func (r *Repo) ListChannels(ctx context.Context, p ListChannelsParams) ([]*models.Channel, error) {
	rows, err := r.q.ListChannels(ctx, sattlmdb.ListChannelsParams{
		TenantID:        mappers.PgUUID(p.TenantID),
		SatelliteID:     mappers.PgUUIDPtr(p.SatelliteID),
		Subsystem:       p.Subsystem,
		CursorCreatedAt: mappers.PgTimestampPtr(p.CursorCreated),
		CursorID:        mappers.PgUUID(p.CursorID),
		Lim:             p.Limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*models.Channel, 0, len(rows))
	for _, row := range rows {
		out = append(out, mappers.ChannelFromRow(row))
	}
	return out, nil
}

// DeprecateChannel sets active=false on a channel.
func (r *Repo) DeprecateChannel(ctx context.Context, id uuid.UUID, updatedBy string) (*models.Channel, error) {
	row, err := r.q.DeprecateChannel(ctx, sattlmdb.DeprecateChannelParams{
		ID:        mappers.PgUUID(id),
		UpdatedBy: updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mappers.ChannelFromRow(row), nil
}

// ----- Frames + Samples ----------------------------------------------------

// IngestFrameParams is the input to IngestFrame.
type IngestFrameParams struct {
	FrameID          uuid.UUID
	TenantID         uuid.UUID
	SatelliteID      uuid.UUID
	APID             uint32
	VirtualChannel   uint32
	SequenceCount    uint64
	SatTime          time.Time
	PayloadSizeBytes uint64
	PayloadSHA256    string
	FrameType        string
	CreatedBy        string
	Samples          []IngestSampleParams
}

// IngestSampleParams is one sample inside a frame ingest.
type IngestSampleParams struct {
	SampleID    uuid.UUID
	ChannelID   uuid.UUID
	SampleTime  time.Time
	ValueDouble float64
	ValueInt    int64
	ValueBool   bool
	ValueText   string
}

// IngestFrame writes a frame and all its samples in a single transaction.
func (r *Repo) IngestFrame(ctx context.Context, p IngestFrameParams) (*models.Frame, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.q.WithTx(tx)
	row, err := q.InsertTelemetryFrame(ctx, sattlmdb.InsertTelemetryFrameParams{
		ID:               mappers.PgUUID(p.FrameID),
		TenantID:         mappers.PgUUID(p.TenantID),
		SatelliteID:      mappers.PgUUID(p.SatelliteID),
		Apid:             int32(p.APID),
		VirtualChannel:   int32(p.VirtualChannel),
		SequenceCount:    int64(p.SequenceCount),
		SatTime:          mappers.PgTimestamp(p.SatTime),
		PayloadSizeBytes: int64(p.PayloadSizeBytes),
		PayloadSha256:    p.PayloadSHA256,
		FrameType:        p.FrameType,
		CreatedBy:        p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	for _, s := range p.Samples {
		if err := q.InsertSample(ctx, sattlmdb.InsertSampleParams{
			ID:          mappers.PgUUID(s.SampleID),
			TenantID:    mappers.PgUUID(p.TenantID),
			SatelliteID: mappers.PgUUID(p.SatelliteID),
			FrameID:     mappers.PgUUID(p.FrameID),
			ChannelID:   mappers.PgUUID(s.ChannelID),
			SampleTime:  mappers.PgTimestamp(s.SampleTime),
			ValueDouble: s.ValueDouble,
			ValueInt:    s.ValueInt,
			ValueBool:   s.ValueBool,
			ValueText:   s.ValueText,
		}); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return mappers.FrameFromRow(row), nil
}

// GetFrame returns a frame by id.
func (r *Repo) GetFrame(ctx context.Context, id uuid.UUID) (*models.Frame, error) {
	row, err := r.q.GetFrame(ctx, mappers.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mappers.FrameFromRow(row), nil
}

// ListFramesParams is the input to ListFrames.
type ListFramesParams struct {
	TenantID          uuid.UUID
	SatelliteID       *uuid.UUID
	FrameType         *string
	GroundTimeStart   *time.Time
	GroundTimeEnd     *time.Time
	CursorGroundTime  *time.Time
	CursorID          uuid.UUID
	Limit             int32
}

// ListFrames returns one page of frames.
func (r *Repo) ListFrames(ctx context.Context, p ListFramesParams) ([]*models.Frame, error) {
	rows, err := r.q.ListFrames(ctx, sattlmdb.ListFramesParams{
		TenantID:         mappers.PgUUID(p.TenantID),
		SatelliteID:      mappers.PgUUIDPtr(p.SatelliteID),
		FrameType:        p.FrameType,
		TimeStart:        mappers.PgTimestampPtr(p.GroundTimeStart),
		TimeEnd:          mappers.PgTimestampPtr(p.GroundTimeEnd),
		CursorGroundTime: mappers.PgTimestampPtr(p.CursorGroundTime),
		CursorID:         mappers.PgUUID(p.CursorID),
		Lim:              p.Limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*models.Frame, 0, len(rows))
	for _, row := range rows {
		out = append(out, mappers.FrameFromRow(row))
	}
	return out, nil
}

// QuerySamplesParams is the input to QuerySamples.
type QuerySamplesParams struct {
	TenantID  uuid.UUID
	ChannelID uuid.UUID
	TimeStart *time.Time
	TimeEnd   *time.Time
	Limit     int32
}

// QuerySamples returns one window of samples for a channel.
func (r *Repo) QuerySamples(ctx context.Context, p QuerySamplesParams) ([]*models.Sample, error) {
	rows, err := r.q.QuerySamples(ctx, sattlmdb.QuerySamplesParams{
		TenantID:  mappers.PgUUID(p.TenantID),
		ChannelID: mappers.PgUUID(p.ChannelID),
		TimeStart: mappers.PgTimestampPtr(p.TimeStart),
		TimeEnd:   mappers.PgTimestampPtr(p.TimeEnd),
		Lim:       p.Limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*models.Sample, 0, len(rows))
	for _, row := range rows {
		out = append(out, mappers.SampleFromQueryRow(row))
	}
	return out, nil
}
