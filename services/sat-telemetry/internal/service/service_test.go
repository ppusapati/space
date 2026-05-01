package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ppusapati/space/pkg/errs"
	"github.com/ppusapati/space/services/sat-telemetry/internal/models"
	"github.com/ppusapati/space/services/sat-telemetry/internal/service"
)

func TestDefineChannelRequiresIDs(t *testing.T) {
	s := service.New(nil)
	_, err := s.DefineChannel(context.Background(), service.DefineChannelInput{})
	if err == nil {
		t.Fatal("expected error for missing ids")
	}
	var de *errs.E
	if !errors.As(err, &de) || de.Domain != errs.DomainInvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestDefineChannelRejectsUnspecifiedType(t *testing.T) {
	s := service.New(nil)
	_, err := s.DefineChannel(context.Background(), service.DefineChannelInput{
		TenantID:    uuid.New(),
		SatelliteID: uuid.New(),
		Subsystem:   "eps",
		Name:        "battery_voltage",
	})
	if err == nil {
		t.Fatal("expected error for unspecified value type")
	}
}

func TestDefineChannelRejectsNegativeRate(t *testing.T) {
	s := service.New(nil)
	_, err := s.DefineChannel(context.Background(), service.DefineChannelInput{
		TenantID:     uuid.New(),
		SatelliteID:  uuid.New(),
		Subsystem:    "eps",
		Name:         "battery_voltage",
		ValueType:    models.ValueTypeFloat,
		SampleRateHz: -1,
	})
	if err == nil {
		t.Fatal("expected error for negative sample rate")
	}
}

func TestDefineChannelRejectsInvertedLimits(t *testing.T) {
	s := service.New(nil)
	_, err := s.DefineChannel(context.Background(), service.DefineChannelInput{
		TenantID:    uuid.New(),
		SatelliteID: uuid.New(),
		Subsystem:   "eps",
		Name:        "battery_voltage",
		ValueType:   models.ValueTypeFloat,
		MinValue:    10,
		MaxValue:    5,
	})
	if err == nil {
		t.Fatal("expected error for max < min")
	}
}

func TestIngestFrameRejectsBadSHA(t *testing.T) {
	s := service.New(nil)
	_, _, err := s.IngestFrame(context.Background(), service.IngestFrameInput{
		TenantID:      uuid.New(),
		SatelliteID:   uuid.New(),
		SatTime:       time.Now(),
		PayloadSHA256: strings.Repeat("z", 64),
		FrameType:     "HK",
	})
	if err == nil {
		t.Fatal("expected error for non-hex sha")
	}
}

func TestIngestFrameRejectsZeroSatTime(t *testing.T) {
	s := service.New(nil)
	_, _, err := s.IngestFrame(context.Background(), service.IngestFrameInput{
		TenantID:      uuid.New(),
		SatelliteID:   uuid.New(),
		PayloadSHA256: strings.Repeat("a", 64),
		FrameType:     "HK",
	})
	if err == nil {
		t.Fatal("expected error for zero sat_time")
	}
}

func TestIngestFrameRejectsEmptyFrameType(t *testing.T) {
	s := service.New(nil)
	_, _, err := s.IngestFrame(context.Background(), service.IngestFrameInput{
		TenantID:      uuid.New(),
		SatelliteID:   uuid.New(),
		SatTime:       time.Now(),
		PayloadSHA256: strings.Repeat("a", 64),
	})
	if err == nil {
		t.Fatal("expected error for empty frame_type")
	}
}

func TestQuerySamplesRequiresTenantAndChannel(t *testing.T) {
	s := service.New(nil)
	if _, err := s.QuerySamples(context.Background(), service.QuerySamplesInput{Limit: 10}); err == nil {
		t.Fatal("expected error for missing ids")
	}
}

func TestQuerySamplesRejectsBadLimit(t *testing.T) {
	s := service.New(nil)
	_, err := s.QuerySamples(context.Background(), service.QuerySamplesInput{
		TenantID: uuid.New(), ChannelID: uuid.New(), Limit: 0,
	})
	if err == nil {
		t.Fatal("expected error for zero limit")
	}
	_, err = s.QuerySamples(context.Background(), service.QuerySamplesInput{
		TenantID: uuid.New(), ChannelID: uuid.New(), Limit: 200000,
	})
	if err == nil {
		t.Fatal("expected error for over-large limit")
	}
}

func TestQuerySamplesRejectsInvertedTimeRange(t *testing.T) {
	s := service.New(nil)
	t1 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := s.QuerySamples(context.Background(), service.QuerySamplesInput{
		TenantID: uuid.New(), ChannelID: uuid.New(), Limit: 10,
		TimeStart: &t1, TimeEnd: &t0,
	})
	if err == nil {
		t.Fatal("expected error for end < start")
	}
}

func TestListsRequireTenant(t *testing.T) {
	s := service.New(nil)
	if _, err := s.ListChannels(context.Background(), service.ListChannelsInput{}); err == nil {
		t.Fatal("expected error for nil tenant")
	}
	if _, err := s.ListFrames(context.Background(), service.ListFramesInput{}); err == nil {
		t.Fatal("expected error for nil tenant")
	}
}
