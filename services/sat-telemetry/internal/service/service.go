// Package service holds sat-telemetry business logic.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/ppusapati/space/pkg/errs"
	"github.com/ppusapati/space/services/sat-telemetry/internal/models"
	"github.com/ppusapati/space/services/sat-telemetry/internal/repository"
)

// Telemetry is the service-layer facade.
type Telemetry struct {
	repo  *repository.Repo
	IDFn  func() uuid.UUID
	NowFn func() time.Time
}

// New constructs a Telemetry service.
func New(repo *repository.Repo) *Telemetry {
	return &Telemetry{
		repo:  repo,
		IDFn:  uuid.New,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// ----- Channels ------------------------------------------------------------

// DefineChannelInput is the input to [Telemetry.DefineChannel].
type DefineChannelInput struct {
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

// DefineChannel persists a new channel.
func (s *Telemetry) DefineChannel(ctx context.Context, in DefineChannelInput) (*models.Channel, error) {
	if in.TenantID == uuid.Nil || in.SatelliteID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id and satellite_id required")
	}
	if in.Subsystem == "" || in.Name == "" {
		return nil, errs.New(errs.DomainInvalidArgument, "subsystem and name required")
	}
	if in.ValueType == models.ValueTypeUnspecified {
		return nil, errs.New(errs.DomainInvalidArgument, "value_type required")
	}
	if in.SampleRateHz < 0 {
		return nil, errs.New(errs.DomainInvalidArgument, "sample_rate_hz must be >= 0")
	}
	if in.MaxValue < in.MinValue {
		return nil, errs.New(errs.DomainInvalidArgument, "max_value must be >= min_value")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return s.repo.DefineChannel(ctx, repository.DefineChannelParams{
		ID:           s.IDFn(),
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
func (s *Telemetry) GetChannel(ctx context.Context, id uuid.UUID) (*models.Channel, error) {
	c, err := s.repo.GetChannel(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errs.New(errs.DomainNotFound, "channel %s", id)
	}
	return c, err
}

// ListChannelsInput is the input to [Telemetry.ListChannels].
type ListChannelsInput struct {
	TenantID      uuid.UUID
	SatelliteID   *uuid.UUID
	Subsystem     *string
	CursorCreated *time.Time
	CursorID      uuid.UUID
	Limit         int32
}

// ListChannels returns one page of channels.
func (s *Telemetry) ListChannels(ctx context.Context, in ListChannelsInput) ([]*models.Channel, error) {
	if in.TenantID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id required")
	}
	return s.repo.ListChannels(ctx, repository.ListChannelsParams{
		TenantID:      in.TenantID,
		SatelliteID:   in.SatelliteID,
		Subsystem:     in.Subsystem,
		CursorCreated: in.CursorCreated,
		CursorID:      in.CursorID,
		Limit:         in.Limit,
	})
}

// DeprecateChannel marks a channel inactive.
func (s *Telemetry) DeprecateChannel(ctx context.Context, id uuid.UUID, updatedBy string) (*models.Channel, error) {
	if updatedBy == "" {
		updatedBy = "system"
	}
	c, err := s.repo.DeprecateChannel(ctx, id, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errs.New(errs.DomainNotFound, "channel %s", id)
	}
	return c, err
}

// ----- Frames + Samples ----------------------------------------------------

// IngestFrameInput is the input to [Telemetry.IngestFrame].
type IngestFrameInput struct {
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
	Samples          []SampleInput
}

// SampleInput is one inline sample.
type SampleInput struct {
	ChannelID   uuid.UUID
	SampleTime  time.Time
	ValueDouble float64
	ValueInt    int64
	ValueBool   bool
	ValueText   string
}

// IngestFrame writes a frame and its samples atomically. The PayloadSHA256
// is normalised to lowercase. Each sample's channel must be active and
// belong to the same tenant + satellite as the frame; the service caches
// channel metadata per-call to avoid repeated lookups.
func (s *Telemetry) IngestFrame(ctx context.Context, in IngestFrameInput) (*models.Frame, int, error) {
	if in.TenantID == uuid.Nil || in.SatelliteID == uuid.Nil {
		return nil, 0, errs.New(errs.DomainInvalidArgument, "tenant_id and satellite_id required")
	}
	if in.SatTime.IsZero() {
		return nil, 0, errs.New(errs.DomainInvalidArgument, "sat_time required")
	}
	if len(in.PayloadSHA256) != 64 || !isHex(in.PayloadSHA256) {
		return nil, 0, errs.New(errs.DomainInvalidArgument, "payload_sha256 must be a 64-char hex string")
	}
	if in.FrameType == "" {
		return nil, 0, errs.New(errs.DomainInvalidArgument, "frame_type required")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	channels := make(map[uuid.UUID]*models.Channel, len(in.Samples))
	samples := make([]repository.IngestSampleParams, 0, len(in.Samples))
	for i, sm := range in.Samples {
		if sm.ChannelID == uuid.Nil {
			return nil, 0, errs.New(errs.DomainInvalidArgument, "samples[%d]: channel_id required", i)
		}
		if sm.SampleTime.IsZero() {
			return nil, 0, errs.New(errs.DomainInvalidArgument, "samples[%d]: sample_time required", i)
		}
		c, ok := channels[sm.ChannelID]
		if !ok {
			loaded, err := s.repo.GetChannel(ctx, sm.ChannelID)
			if err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return nil, 0, errs.New(errs.DomainInvalidArgument,
						"samples[%d]: channel %s not found", i, sm.ChannelID)
				}
				return nil, 0, err
			}
			c = loaded
			channels[sm.ChannelID] = c
		}
		if c.TenantID != in.TenantID || c.SatelliteID != in.SatelliteID {
			return nil, 0, errs.New(errs.DomainInvalidArgument,
				"samples[%d]: channel %s belongs to a different tenant or satellite", i, sm.ChannelID)
		}
		if !c.Active {
			return nil, 0, errs.New(errs.DomainPreconditionFailed,
				"samples[%d]: channel %s is deprecated", i, sm.ChannelID)
		}
		samples = append(samples, repository.IngestSampleParams{
			SampleID:    s.IDFn(),
			ChannelID:   sm.ChannelID,
			SampleTime:  sm.SampleTime,
			ValueDouble: sm.ValueDouble,
			ValueInt:    sm.ValueInt,
			ValueBool:   sm.ValueBool,
			ValueText:   sm.ValueText,
		})
	}
	frame, err := s.repo.IngestFrame(ctx, repository.IngestFrameParams{
		FrameID:          s.IDFn(),
		TenantID:         in.TenantID,
		SatelliteID:      in.SatelliteID,
		APID:             in.APID,
		VirtualChannel:   in.VirtualChannel,
		SequenceCount:    in.SequenceCount,
		SatTime:          in.SatTime,
		PayloadSizeBytes: in.PayloadSizeBytes,
		PayloadSHA256:    lowerHex(in.PayloadSHA256),
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
func (s *Telemetry) GetFrame(ctx context.Context, id uuid.UUID) (*models.Frame, error) {
	f, err := s.repo.GetFrame(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, errs.New(errs.DomainNotFound, "frame %s", id)
	}
	return f, err
}

// ListFramesInput is the input to [Telemetry.ListFrames].
type ListFramesInput struct {
	TenantID         uuid.UUID
	SatelliteID      *uuid.UUID
	FrameType        *string
	GroundTimeStart  *time.Time
	GroundTimeEnd    *time.Time
	CursorGroundTime *time.Time
	CursorID         uuid.UUID
	Limit            int32
}

// ListFrames returns one page of frames.
func (s *Telemetry) ListFrames(ctx context.Context, in ListFramesInput) ([]*models.Frame, error) {
	if in.TenantID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id required")
	}
	return s.repo.ListFrames(ctx, repository.ListFramesParams{
		TenantID:         in.TenantID,
		SatelliteID:      in.SatelliteID,
		FrameType:        in.FrameType,
		GroundTimeStart:  in.GroundTimeStart,
		GroundTimeEnd:    in.GroundTimeEnd,
		CursorGroundTime: in.CursorGroundTime,
		CursorID:         in.CursorID,
		Limit:            in.Limit,
	})
}

// QuerySamplesInput is the input to [Telemetry.QuerySamples].
type QuerySamplesInput struct {
	TenantID  uuid.UUID
	ChannelID uuid.UUID
	TimeStart *time.Time
	TimeEnd   *time.Time
	Limit     int32
}

// QuerySamples returns one window of samples for a channel.
func (s *Telemetry) QuerySamples(ctx context.Context, in QuerySamplesInput) ([]*models.Sample, error) {
	if in.TenantID == uuid.Nil || in.ChannelID == uuid.Nil {
		return nil, errs.New(errs.DomainInvalidArgument, "tenant_id and channel_id required")
	}
	if in.Limit <= 0 || in.Limit > 100000 {
		return nil, errs.New(errs.DomainInvalidArgument, "limit must be in (0, 100000]")
	}
	if in.TimeStart != nil && in.TimeEnd != nil && in.TimeEnd.Before(*in.TimeStart) {
		return nil, errs.New(errs.DomainInvalidArgument, "time_end must be >= time_start")
	}
	return s.repo.QuerySamples(ctx, repository.QuerySamplesParams{
		TenantID:  in.TenantID,
		ChannelID: in.ChannelID,
		TimeStart: in.TimeStart,
		TimeEnd:   in.TimeEnd,
		Limit:     in.Limit,
	})
}

func isHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		case c >= 'A' && c <= 'F':
		default:
			return false
		}
	}
	return true
}

func lowerHex(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'F' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
