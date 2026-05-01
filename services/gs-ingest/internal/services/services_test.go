package services_test

import (
	"context"
	"strings"
	"testing"
	"time"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	"github.com/ppusapati/space/services/gs-ingest/internal/models"
	"github.com/ppusapati/space/services/gs-ingest/internal/services"
)

func TestStartIngestSessionRejectsZeroIDs(t *testing.T) {
	s := services.New(nil)
	_, err := s.StartIngestSession(context.Background(), services.StartIngestSessionInput{})
	if err == nil {
		t.Fatal("expected error for zero ids")
	}
	if pkgerrors.Code(err) != 400 {
		t.Fatalf("expected 400, got %d", pkgerrors.Code(err))
	}
}

func TestUpdateIngestStatusRejectsUnspecified(t *testing.T) {
	s := services.New(nil)
	_, err := s.UpdateIngestStatus(context.Background(), ulid.New(), models.StatusUnspecified, "", "")
	if err == nil {
		t.Fatal("expected error for unspecified status")
	}
}

func TestRecordDownlinkFrameRejectsZeroIDs(t *testing.T) {
	s := services.New(nil)
	_, err := s.RecordDownlinkFrame(context.Background(), services.RecordDownlinkFrameInput{})
	if err == nil {
		t.Fatal("expected error for zero ids")
	}
}

func TestRecordDownlinkFrameRejectsZeroSize(t *testing.T) {
	s := services.New(nil)
	_, err := s.RecordDownlinkFrame(context.Background(), services.RecordDownlinkFrameInput{
		TenantID:         ulid.New(),
		SessionID:        ulid.New(),
		PayloadSizeBytes: 0,
		PayloadSHA256:    strings.Repeat("a", 64),
		PayloadURI:       "s3://x",
		FrameType:        "TM",
	})
	if err == nil {
		t.Fatal("expected error for zero payload size")
	}
}

func TestRecordDownlinkFrameRejectsBadSHA(t *testing.T) {
	s := services.New(nil)
	_, err := s.RecordDownlinkFrame(context.Background(), services.RecordDownlinkFrameInput{
		TenantID:         ulid.New(),
		SessionID:        ulid.New(),
		PayloadSizeBytes: 1024,
		PayloadSHA256:    "z",
		PayloadURI:       "s3://x",
		FrameType:        "TM",
		GroundTime:       time.Now(),
	})
	if err == nil {
		t.Fatal("expected error for bad sha")
	}
}

func TestListsRequireTenant(t *testing.T) {
	s := services.New(nil)
	if _, _, err := s.ListIngestSessionsForTenant(context.Background(), services.ListIngestSessionsInput{}); err == nil {
		t.Fatal("expected error for nil tenant on sessions")
	}
	if _, _, err := s.ListDownlinkFramesForTenant(context.Background(), services.ListDownlinkFramesInput{}); err == nil {
		t.Fatal("expected error for nil tenant on frames")
	}
}
