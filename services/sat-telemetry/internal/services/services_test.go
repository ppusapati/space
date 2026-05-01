package services_test

import (
	"context"
	"strings"
	"testing"
	"time"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	"github.com/ppusapati/space/services/sat-telemetry/internal/models"
	"github.com/ppusapati/space/services/sat-telemetry/internal/services"
)

func TestDefineChannelRequiresIDs(t *testing.T) {
	s := services.New(nil)
	_, err := s.DefineChannel(context.Background(), services.DefineChannelInput{})
	if err == nil {
		t.Fatal("expected error for missing ids")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestDefineChannelRejectsUnspecifiedType(t *testing.T) {
	s := services.New(nil)
	_, err := s.DefineChannel(context.Background(), services.DefineChannelInput{
		TenantID:    ulid.New(),
		SatelliteID: ulid.New(),
		Subsystem:   "eps",
		Name:        "battery_voltage",
	})
	if err == nil {
		t.Fatal("expected error for unspecified value type")
	}
}

func TestDefineChannelRejectsInvertedLimits(t *testing.T) {
	s := services.New(nil)
	_, err := s.DefineChannel(context.Background(), services.DefineChannelInput{
		TenantID:    ulid.New(),
		SatelliteID: ulid.New(),
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
	s := services.New(nil)
	_, _, err := s.IngestFrame(context.Background(), services.IngestFrameInput{
		TenantID:      ulid.New(),
		SatelliteID:   ulid.New(),
		SatTime:       time.Now(),
		PayloadSHA256: strings.Repeat("z", 64),
		FrameType:     "HK",
	})
	if err == nil {
		t.Fatal("expected error for non-hex sha")
	}
}

func TestIngestFrameRejectsZeroSatTime(t *testing.T) {
	s := services.New(nil)
	_, _, err := s.IngestFrame(context.Background(), services.IngestFrameInput{
		TenantID:      ulid.New(),
		SatelliteID:   ulid.New(),
		PayloadSHA256: strings.Repeat("a", 64),
		FrameType:     "HK",
	})
	if err == nil {
		t.Fatal("expected error for zero sat_time")
	}
}

func TestIngestFrameRejectsEmptyFrameType(t *testing.T) {
	s := services.New(nil)
	_, _, err := s.IngestFrame(context.Background(), services.IngestFrameInput{
		TenantID:      ulid.New(),
		SatelliteID:   ulid.New(),
		SatTime:       time.Now(),
		PayloadSHA256: strings.Repeat("a", 64),
	})
	if err == nil {
		t.Fatal("expected error for empty frame_type")
	}
}

func TestQuerySamplesRequiresTenantAndChannel(t *testing.T) {
	s := services.New(nil)
	if _, err := s.QuerySamples(context.Background(), services.QuerySamplesInput{Limit: 10}); err == nil {
		t.Fatal("expected error for missing ids")
	}
}

func TestQuerySamplesRejectsBadLimit(t *testing.T) {
	s := services.New(nil)
	_, err := s.QuerySamples(context.Background(), services.QuerySamplesInput{
		TenantID: ulid.New(), ChannelID: ulid.New(), Limit: 0,
	})
	if err == nil {
		t.Fatal("expected error for zero limit")
	}
	_, err = s.QuerySamples(context.Background(), services.QuerySamplesInput{
		TenantID: ulid.New(), ChannelID: ulid.New(), Limit: 200000,
	})
	if err == nil {
		t.Fatal("expected error for over-large limit")
	}
}

func TestQuerySamplesRejectsInvertedTimeRange(t *testing.T) {
	s := services.New(nil)
	t1 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := s.QuerySamples(context.Background(), services.QuerySamplesInput{
		TenantID: ulid.New(), ChannelID: ulid.New(), Limit: 10,
		TimeStart: t1, TimeEnd: t0,
	})
	if err == nil {
		t.Fatal("expected error for end < start")
	}
}

func TestListsRequireTenant(t *testing.T) {
	s := services.New(nil)
	if _, _, err := s.ListChannelsForTenant(context.Background(), services.ListChannelsInput{}); err == nil {
		t.Fatal("expected error for nil tenant")
	}
	if _, _, err := s.ListFramesForTenant(context.Background(), services.ListFramesInput{}); err == nil {
		t.Fatal("expected error for nil tenant")
	}
}
