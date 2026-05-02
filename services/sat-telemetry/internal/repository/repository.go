// Package repository wraps the sat-telemetry sqlc layer.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/chetana/packages/ulid"

	sattlmdb "github.com/ppusapati/space/services/sat-telemetry/db/generated"
	"github.com/ppusapati/space/services/sat-telemetry/internal/mapper"
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

// DefineChannelParams holds the input for [Repo.DefineChannel].
type DefineChannelParams struct {
	ID           ulid.ID
	TenantID     ulid.ID
	SatelliteID  ulid.ID
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
		ID:           mapper.PgUUID(p.ID),
		TenantID:     mapper.PgUUID(p.TenantID),
		SatelliteID:  mapper.PgUUID(p.SatelliteID),
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
	return mapper.ChannelFromRow(row), nil
}

// GetChannel returns a channel by id.
func (r *Repo) GetChannel(ctx context.Context, id ulid.ID) (*models.Channel, error) {
	row, err := r.q.GetChannel(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.ChannelFromRow(row), nil
}

// ListChannelsParams holds the input for [Repo.ListChannelsForTenant].
type ListChannelsParams struct {
	TenantID    ulid.ID
	SatelliteID *ulid.ID
	Subsystem   string
	PageOffset  int32
	PageSize    int32
}

// ListChannelsForTenant returns one page of channels.
func (r *Repo) ListChannelsForTenant(ctx context.Context, p ListChannelsParams) ([]*models.Channel, int32, error) {
	var subsystemPtr *string
	if p.Subsystem != "" {
		v := p.Subsystem
		subsystemPtr = &v
	}
	var satellitePg = mapper.PgUUIDOrNull(ulid.Zero)
	if p.SatelliteID != nil {
		satellitePg = mapper.PgUUID(*p.SatelliteID)
	}
	total, err := r.q.CountChannelsForTenant(ctx, sattlmdb.CountChannelsForTenantParams{
		TenantID:    mapper.PgUUID(p.TenantID),
		SatelliteID: satellitePg,
		Subsystem:   subsystemPtr,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListChannelsForTenant(ctx, sattlmdb.ListChannelsForTenantParams{
		TenantID:    mapper.PgUUID(p.TenantID),
		SatelliteID: satellitePg,
		Subsystem:   subsystemPtr,
		PageOffset:  p.PageOffset,
		PageSize:    p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.Channel, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.ChannelFromRow(row))
	}
	return out, int32(total), nil
}

// DeprecateChannel marks a channel inactive.
func (r *Repo) DeprecateChannel(ctx context.Context, id ulid.ID, updatedBy string) (*models.Channel, error) {
	row, err := r.q.DeprecateChannel(ctx, sattlmdb.DeprecateChannelParams{
		ID:        mapper.PgUUID(id),
		UpdatedBy: updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.ChannelFromRow(row), nil
}

// ----- Frames + Samples ----------------------------------------------------

// IngestFrameParams holds the input for [Repo.IngestFrame].
type IngestFrameParams struct {
	FrameID          ulid.ID
	TenantID         ulid.ID
	SatelliteID      ulid.ID
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

// IngestSampleParams is one inline sample inside a frame ingest.
type IngestSampleParams struct {
	SampleID    ulid.ID
	ChannelID   ulid.ID
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
		ID:               mapper.PgUUID(p.FrameID),
		TenantID:         mapper.PgUUID(p.TenantID),
		SatelliteID:      mapper.PgUUID(p.SatelliteID),
		Apid:             int32(p.APID),
		VirtualChannel:   int32(p.VirtualChannel),
		SequenceCount:    int64(p.SequenceCount),
		SatTime:          mapper.PgTimestamp(p.SatTime),
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
			ID:          mapper.PgUUID(s.SampleID),
			TenantID:    mapper.PgUUID(p.TenantID),
			SatelliteID: mapper.PgUUID(p.SatelliteID),
			FrameID:     mapper.PgUUID(p.FrameID),
			ChannelID:   mapper.PgUUID(s.ChannelID),
			SampleTime:  mapper.PgTimestamp(s.SampleTime),
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
	return mapper.FrameFromRow(row), nil
}

// GetFrame returns a frame by id.
func (r *Repo) GetFrame(ctx context.Context, id ulid.ID) (*models.Frame, error) {
	row, err := r.q.GetFrame(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.FrameFromRow(row), nil
}

// ListFramesParams holds the input for [Repo.ListFramesForTenant].
type ListFramesParams struct {
	TenantID    ulid.ID
	SatelliteID *ulid.ID
	FrameType   string
	TimeStart   time.Time
	TimeEnd     time.Time
	PageOffset  int32
	PageSize    int32
}

// ListFramesForTenant returns one page of frames.
func (r *Repo) ListFramesForTenant(ctx context.Context, p ListFramesParams) ([]*models.Frame, int32, error) {
	var frameTypePtr *string
	if p.FrameType != "" {
		v := p.FrameType
		frameTypePtr = &v
	}
	var satellitePg = mapper.PgUUIDOrNull(ulid.Zero)
	if p.SatelliteID != nil {
		satellitePg = mapper.PgUUID(*p.SatelliteID)
	}
	total, err := r.q.CountFramesForTenant(ctx, sattlmdb.CountFramesForTenantParams{
		TenantID:    mapper.PgUUID(p.TenantID),
		SatelliteID: satellitePg,
		FrameType:   frameTypePtr,
		TimeStart:   mapper.PgTimestampOrNull(p.TimeStart),
		TimeEnd:     mapper.PgTimestampOrNull(p.TimeEnd),
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListFramesForTenant(ctx, sattlmdb.ListFramesForTenantParams{
		TenantID:    mapper.PgUUID(p.TenantID),
		SatelliteID: satellitePg,
		FrameType:   frameTypePtr,
		TimeStart:   mapper.PgTimestampOrNull(p.TimeStart),
		TimeEnd:     mapper.PgTimestampOrNull(p.TimeEnd),
		PageOffset:  p.PageOffset,
		PageSize:    p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.Frame, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.FrameFromRow(row))
	}
	return out, int32(total), nil
}

// QuerySamplesParams holds the input for [Repo.QuerySamples].
type QuerySamplesParams struct {
	TenantID  ulid.ID
	ChannelID ulid.ID
	TimeStart time.Time
	TimeEnd   time.Time
	Limit     int32
}

// QuerySamples returns one window of samples for a channel.
func (r *Repo) QuerySamples(ctx context.Context, p QuerySamplesParams) ([]*models.Sample, error) {
	rows, err := r.q.QuerySamples(ctx, sattlmdb.QuerySamplesParams{
		TenantID:  mapper.PgUUID(p.TenantID),
		ChannelID: mapper.PgUUID(p.ChannelID),
		TimeStart: mapper.PgTimestampOrNull(p.TimeStart),
		TimeEnd:   mapper.PgTimestampOrNull(p.TimeEnd),
		Lim:       p.Limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*models.Sample, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.SampleFromRow(row))
	}
	return out, nil
}
