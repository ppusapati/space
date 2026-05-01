// Package services holds sat-telemetry business logic.
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	"github.com/ppusapati/space/services/sat-telemetry/internal/models"
	"github.com/ppusapati/space/services/sat-telemetry/internal/repository"
)

// Telemetry is the sat-telemetry service-layer facade.
type Telemetry struct {
	repo  *repository.Repo
	IDFn  func() ulid.ID
	NowFn func() time.Time
}

// New constructs a Telemetry service.
func New(repo *repository.Repo) *Telemetry {
	return &Telemetry{
		repo:  repo,
		IDFn:  ulid.NewMonotonic,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// ----- Channels ------------------------------------------------------------

// DefineChannelInput is the input for [Telemetry.DefineChannel].
type DefineChannelInput struct {
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

// DefineChannel persists a new channel.
func (t *Telemetry) DefineChannel(ctx context.Context, in DefineChannelInput) (*models.Channel, error) {
	if in.TenantID.IsZero() || in.SatelliteID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id and satellite_id required")
	}
	in.Subsystem = strings.TrimSpace(in.Subsystem)
	in.Name = strings.TrimSpace(in.Name)
	if in.Subsystem == "" || in.Name == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "subsystem and name required")
	}
	if in.ValueType == models.ValueTypeUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "value_type required")
	}
	if in.SampleRateHz < 0 {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "sample_rate_hz must be >= 0")
	}
	if in.MaxValue < in.MinValue {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "max_value must be >= min_value")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return t.repo.DefineChannel(ctx, repository.DefineChannelParams{
		ID:           t.IDFn(),
		TenantID:     in.TenantID,
		SatelliteID:  in.SatelliteID,
		Subsystem:    in.Subsystem,
		Name:         in.Name,
		Units:        in.Units,
		ValueType:    in.ValueType,
		MinValue:     in.MinValue,
		MaxValue:     in.MaxValue,
		SampleRateHz: in.SampleRateHz,
		CreatedBy:    createdBy,
	})
}

// GetChannel fetches a channel by id.
func (t *Telemetry) GetChannel(ctx context.Context, id ulid.ID) (*models.Channel, error) {
	c, err := t.repo.GetChannel(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("CHANNEL_NOT_FOUND", "channel "+id.String())
	}
	return c, err
}

// ListChannelsInput is the input for [Telemetry.ListChannelsForTenant].
type ListChannelsInput struct {
	TenantID    ulid.ID
	SatelliteID *ulid.ID
	Subsystem   string
	PageOffset  int32
	PageSize    int32
}

// ListChannelsForTenant returns one page of channels.
func (t *Telemetry) ListChannelsForTenant(ctx context.Context, in ListChannelsInput) ([]*models.Channel, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := t.repo.ListChannelsForTenant(ctx, repository.ListChannelsParams{
		TenantID:    in.TenantID,
		SatelliteID: in.SatelliteID,
		Subsystem:   in.Subsystem,
		PageOffset:  in.PageOffset,
		PageSize:    in.PageSize,
	})
	if err != nil {
		return nil, models.Page{}, err
	}
	return rows, models.Page{
		TotalCount: total,
		PageOffset: in.PageOffset,
		PageSize:   in.PageSize,
		HasNext:    in.PageOffset+int32(len(rows)) < total,
	}, nil
}

// DeprecateChannel marks a channel inactive.
func (t *Telemetry) DeprecateChannel(ctx context.Context, id ulid.ID, updatedBy string) (*models.Channel, error) {
	if updatedBy == "" {
		updatedBy = "system"
	}
	c, err := t.repo.DeprecateChannel(ctx, id, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("CHANNEL_NOT_FOUND", "channel "+id.String())
	}
	return c, err
}

// ----- Frames + Samples ----------------------------------------------------

// IngestFrameInput is the input for [Telemetry.IngestFrame].
type IngestFrameInput struct {
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
	Samples          []SampleInput
}

// SampleInput is one inline sample.
type SampleInput struct {
	ChannelID   ulid.ID
	SampleTime  time.Time
	ValueDouble float64
	ValueInt    int64
	ValueBool   bool
	ValueText   string
}

// IngestFrame writes a frame and its samples atomically. Each sample's
// channel must exist, belong to the same tenant + satellite, and be active.
// Returns the created frame plus the sample count.
func (t *Telemetry) IngestFrame(ctx context.Context, in IngestFrameInput) (*models.Frame, int, error) {
	if in.TenantID.IsZero() || in.SatelliteID.IsZero() {
		return nil, 0, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id and satellite_id required")
	}
	if in.SatTime.IsZero() {
		return nil, 0, pkgerrors.BadRequest("INVALID_ARGUMENT", "sat_time required")
	}
	if !isHex(in.PayloadSHA256, 64) {
		return nil, 0, pkgerrors.BadRequest("INVALID_ARGUMENT", "payload_sha256 must be 64-char hex")
	}
	if strings.TrimSpace(in.FrameType) == "" {
		return nil, 0, pkgerrors.BadRequest("INVALID_ARGUMENT", "frame_type required")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	channels := make(map[ulid.ID]*models.Channel, len(in.Samples))
	samples := make([]repository.IngestSampleParams, 0, len(in.Samples))
	for i, sm := range in.Samples {
		if sm.ChannelID.IsZero() {
			return nil, 0, pkgerrors.BadRequest("INVALID_ARGUMENT",
				"samples: channel_id required")
		}
		if sm.SampleTime.IsZero() {
			return nil, 0, pkgerrors.BadRequest("INVALID_ARGUMENT",
				"samples: sample_time required")
		}
		c, ok := channels[sm.ChannelID]
		if !ok {
			loaded, err := t.repo.GetChannel(ctx, sm.ChannelID)
			if err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return nil, 0, pkgerrors.BadRequest("CHANNEL_NOT_FOUND",
						"samples: channel "+sm.ChannelID.String()+" not found")
				}
				return nil, 0, err
			}
			c = loaded
			channels[sm.ChannelID] = c
		}
		if c.TenantID != in.TenantID || c.SatelliteID != in.SatelliteID {
			return nil, 0, pkgerrors.BadRequest("INVALID_ARGUMENT",
				"samples: channel "+sm.ChannelID.String()+" tenant/satellite mismatch")
		}
		if !c.Active {
			return nil, 0, pkgerrors.New(412, "CHANNEL_DEPRECATED",
				"samples: channel "+sm.ChannelID.String()+" is deprecated")
		}
		samples = append(samples, repository.IngestSampleParams{
			SampleID:    t.IDFn(),
			ChannelID:   sm.ChannelID,
			SampleTime:  sm.SampleTime,
			ValueDouble: sm.ValueDouble,
			ValueInt:    sm.ValueInt,
			ValueBool:   sm.ValueBool,
			ValueText:   sm.ValueText,
		})
		_ = i
	}
	frame, err := t.repo.IngestFrame(ctx, repository.IngestFrameParams{
		FrameID:          t.IDFn(),
		TenantID:         in.TenantID,
		SatelliteID:      in.SatelliteID,
		APID:             in.APID,
		VirtualChannel:   in.VirtualChannel,
		SequenceCount:    in.SequenceCount,
		SatTime:          in.SatTime,
		PayloadSizeBytes: in.PayloadSizeBytes,
		PayloadSHA256:    strings.ToLower(in.PayloadSHA256),
		FrameType:        in.FrameType,
		CreatedBy:        createdBy,
		Samples:          samples,
	})
	if err != nil {
		return nil, 0, err
	}
	return frame, len(samples), nil
}

// GetFrame fetches a frame by id.
func (t *Telemetry) GetFrame(ctx context.Context, id ulid.ID) (*models.Frame, error) {
	f, err := t.repo.GetFrame(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("FRAME_NOT_FOUND", "frame "+id.String())
	}
	return f, err
}

// ListFramesInput is the input for [Telemetry.ListFramesForTenant].
type ListFramesInput struct {
	TenantID    ulid.ID
	SatelliteID *ulid.ID
	FrameType   string
	TimeStart   time.Time
	TimeEnd     time.Time
	PageOffset  int32
	PageSize    int32
}

// ListFramesForTenant returns one page of frames.
func (t *Telemetry) ListFramesForTenant(ctx context.Context, in ListFramesInput) ([]*models.Frame, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := t.repo.ListFramesForTenant(ctx, repository.ListFramesParams{
		TenantID:    in.TenantID,
		SatelliteID: in.SatelliteID,
		FrameType:   in.FrameType,
		TimeStart:   in.TimeStart,
		TimeEnd:     in.TimeEnd,
		PageOffset:  in.PageOffset,
		PageSize:    in.PageSize,
	})
	if err != nil {
		return nil, models.Page{}, err
	}
	return rows, models.Page{
		TotalCount: total,
		PageOffset: in.PageOffset,
		PageSize:   in.PageSize,
		HasNext:    in.PageOffset+int32(len(rows)) < total,
	}, nil
}

// QuerySamplesInput is the input for [Telemetry.QuerySamples].
type QuerySamplesInput struct {
	TenantID  ulid.ID
	ChannelID ulid.ID
	TimeStart time.Time
	TimeEnd   time.Time
	Limit     int32
}

// QuerySamples returns one window of samples for a channel.
func (t *Telemetry) QuerySamples(ctx context.Context, in QuerySamplesInput) ([]*models.Sample, error) {
	if in.TenantID.IsZero() || in.ChannelID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id and channel_id required")
	}
	if in.Limit <= 0 || in.Limit > 100000 {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "limit must be in (0, 100000]")
	}
	if !in.TimeStart.IsZero() && !in.TimeEnd.IsZero() && in.TimeEnd.Before(in.TimeStart) {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "time_end must be >= time_start")
	}
	return t.repo.QuerySamples(ctx, repository.QuerySamplesParams{
		TenantID:  in.TenantID,
		ChannelID: in.ChannelID,
		TimeStart: in.TimeStart,
		TimeEnd:   in.TimeEnd,
		Limit:     in.Limit,
	})
}

func isHex(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		ok := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
		if !ok {
			return false
		}
	}
	return true
}
