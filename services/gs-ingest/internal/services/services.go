// Package services holds gs-ingest business logic.
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	"github.com/ppusapati/space/services/gs-ingest/internal/models"
	"github.com/ppusapati/space/services/gs-ingest/internal/repository"
)

// Ingest is the gs-ingest service-layer facade.
type Ingest struct {
	repo  *repository.Repo
	IDFn  func() ulid.ID
	NowFn func() time.Time
}

func New(repo *repository.Repo) *Ingest {
	return &Ingest{
		repo:  repo,
		IDFn:  ulid.NewMonotonic,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// ----- IngestSession ------------------------------------------------------

type StartIngestSessionInput struct {
	TenantID    ulid.ID
	BookingID   ulid.ID
	PassID      ulid.ID
	StationID   ulid.ID
	SatelliteID ulid.ID
	CreatedBy   string
}

func (s *Ingest) StartIngestSession(ctx context.Context, in StartIngestSessionInput) (*models.IngestSession, error) {
	if in.TenantID.IsZero() || in.BookingID.IsZero() || in.PassID.IsZero() ||
		in.StationID.IsZero() || in.SatelliteID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT",
			"tenant_id, booking_id, pass_id, station_id, satellite_id required")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return s.repo.StartIngestSession(ctx, repository.StartIngestSessionParams{
		ID:          s.IDFn(),
		TenantID:    in.TenantID,
		BookingID:   in.BookingID,
		PassID:      in.PassID,
		StationID:   in.StationID,
		SatelliteID: in.SatelliteID,
		Status:      models.StatusActive,
		CreatedBy:   createdBy,
	})
}

func (s *Ingest) GetIngestSession(ctx context.Context, id ulid.ID) (*models.IngestSession, error) {
	sess, err := s.repo.GetIngestSession(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("SESSION_NOT_FOUND", "ingest_session "+id.String())
	}
	return sess, err
}

type ListIngestSessionsInput struct {
	TenantID    ulid.ID
	StationID   *ulid.ID
	SatelliteID *ulid.ID
	Status      *models.IngestStatus
	PageOffset  int32
	PageSize    int32
}

func (s *Ingest) ListIngestSessionsForTenant(ctx context.Context, in ListIngestSessionsInput) ([]*models.IngestSession, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := s.repo.ListIngestSessionsForTenant(ctx, repository.ListIngestSessionsParams{
		TenantID:    in.TenantID,
		StationID:   in.StationID,
		SatelliteID: in.SatelliteID,
		Status:      in.Status,
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

// UpdateIngestStatus enforces the session transition graph:
//
//	QUEUED -> ACTIVE | CANCELED
//	ACTIVE -> COMPLETED | FAILED | CANCELED
//	COMPLETED, FAILED, CANCELED — terminal.
func (s *Ingest) UpdateIngestStatus(
	ctx context.Context, id ulid.ID, status models.IngestStatus, errorMessage, updatedBy string,
) (*models.IngestSession, error) {
	if status == models.StatusUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "status required")
	}
	current, err := s.GetIngestSession(ctx, id)
	if err != nil {
		return nil, err
	}
	if !validIngestTransition(current.Status, status) {
		return nil, pkgerrors.New(412, "ILLEGAL_TRANSITION",
			"illegal ingest_session status transition")
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	updated, err := s.repo.UpdateIngestStatus(ctx, id, status, errorMessage, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("SESSION_NOT_FOUND", "ingest_session "+id.String())
	}
	return updated, err
}

func validIngestTransition(from, to models.IngestStatus) bool {
	switch from {
	case models.StatusQueued:
		return to == models.StatusActive || to == models.StatusCanceled
	case models.StatusActive:
		return to == models.StatusCompleted || to == models.StatusFailed || to == models.StatusCanceled
	case models.StatusCompleted, models.StatusFailed, models.StatusCanceled:
		return false
	default:
		return false
	}
}

// ----- DownlinkFrame ------------------------------------------------------

type RecordDownlinkFrameInput struct {
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

func (s *Ingest) RecordDownlinkFrame(ctx context.Context, in RecordDownlinkFrameInput) (*models.DownlinkFrame, error) {
	if in.TenantID.IsZero() || in.SessionID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id and session_id required")
	}
	if in.PayloadSizeBytes == 0 {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "payload_size_bytes must be > 0")
	}
	if !isHex(in.PayloadSHA256, 64) {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "payload_sha256 must be 64-char hex")
	}
	if strings.TrimSpace(in.PayloadURI) == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "payload_uri required")
	}
	if strings.TrimSpace(in.FrameType) == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "frame_type required")
	}
	if in.GroundTime.IsZero() {
		in.GroundTime = s.NowFn()
	}
	// Verify session exists, belongs to same tenant, and is ACTIVE.
	sess, err := s.repo.GetIngestSession(ctx, in.SessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, pkgerrors.BadRequest("SESSION_NOT_FOUND",
				"ingest_session "+in.SessionID.String())
		}
		return nil, err
	}
	if sess.TenantID != in.TenantID {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "session tenant mismatch")
	}
	if sess.Status != models.StatusActive {
		return nil, pkgerrors.New(412, "SESSION_NOT_ACTIVE",
			"ingest_session is not ACTIVE")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return s.repo.RecordDownlinkFrame(ctx, repository.RecordDownlinkFrameParams{
		ID:               s.IDFn(),
		TenantID:         in.TenantID,
		SessionID:        in.SessionID,
		APID:             in.APID,
		VirtualChannel:   in.VirtualChannel,
		SequenceCount:    in.SequenceCount,
		GroundTime:       in.GroundTime,
		PayloadSizeBytes: in.PayloadSizeBytes,
		PayloadSHA256:    strings.ToLower(in.PayloadSHA256),
		PayloadURI:       in.PayloadURI,
		FrameType:        in.FrameType,
		CreatedBy:        createdBy,
	})
}

func (s *Ingest) GetDownlinkFrame(ctx context.Context, id ulid.ID) (*models.DownlinkFrame, error) {
	f, err := s.repo.GetDownlinkFrame(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("FRAME_NOT_FOUND", "downlink_frame "+id.String())
	}
	return f, err
}

type ListDownlinkFramesInput struct {
	TenantID   ulid.ID
	SessionID  *ulid.ID
	FrameType  string
	TimeStart  time.Time
	TimeEnd    time.Time
	PageOffset int32
	PageSize   int32
}

func (s *Ingest) ListDownlinkFramesForTenant(ctx context.Context, in ListDownlinkFramesInput) ([]*models.DownlinkFrame, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := s.repo.ListDownlinkFramesForTenant(ctx, repository.ListDownlinkFramesParams{
		TenantID:   in.TenantID,
		SessionID:  in.SessionID,
		FrameType:  in.FrameType,
		TimeStart:  in.TimeStart,
		TimeEnd:    in.TimeEnd,
		PageOffset: in.PageOffset,
		PageSize:   in.PageSize,
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
