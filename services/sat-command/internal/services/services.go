// Package services holds sat-command business logic.
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	pkgerrors "p9e.in/samavaya/packages/errors"
	"p9e.in/samavaya/packages/ulid"

	"github.com/ppusapati/space/services/sat-command/internal/models"
	"github.com/ppusapati/space/services/sat-command/internal/repository"
)

// Command is the sat-command service-layer facade.
type Command struct {
	repo  *repository.Repo
	IDFn  func() ulid.ID
	NowFn func() time.Time
}

// New constructs a Command service.
func New(repo *repository.Repo) *Command {
	return &Command{
		repo:  repo,
		IDFn:  ulid.NewMonotonic,
		NowFn: func() time.Time { return time.Now().UTC() },
	}
}

// ----- CommandDef ----------------------------------------------------------

// DefineCommandInput is the input for [Command.DefineCommand].
type DefineCommandInput struct {
	TenantID         ulid.ID
	SatelliteID      ulid.ID // ulid.Zero = tenant-wide
	Subsystem        string
	Name             string
	Opcode           uint32
	ParametersSchema string
	Description      string
	CreatedBy        string
}

// DefineCommand persists a new command_def in active state.
func (c *Command) DefineCommand(ctx context.Context, in DefineCommandInput) (*models.CommandDef, error) {
	if in.TenantID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	in.Subsystem = strings.TrimSpace(in.Subsystem)
	in.Name = strings.TrimSpace(in.Name)
	if in.Subsystem == "" || in.Name == "" {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "subsystem and name required")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return c.repo.DefineCommand(ctx, repository.DefineCommandParams{
		ID:               c.IDFn(),
		TenantID:         in.TenantID,
		SatelliteID:      in.SatelliteID,
		Subsystem:        in.Subsystem,
		Name:             in.Name,
		Opcode:           in.Opcode,
		ParametersSchema: in.ParametersSchema,
		Description:      in.Description,
		CreatedBy:        createdBy,
	})
}

// GetCommand fetches a command_def by id.
func (c *Command) GetCommand(ctx context.Context, id ulid.ID) (*models.CommandDef, error) {
	cd, err := c.repo.GetCommand(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("COMMAND_NOT_FOUND", "command "+id.String())
	}
	return cd, err
}

// ListCommandsInput is the input for [Command.ListCommandsForTenant].
type ListCommandsInput struct {
	TenantID    ulid.ID
	SatelliteID *ulid.ID
	Subsystem   string
	PageOffset  int32
	PageSize    int32
}

// ListCommandsForTenant returns one page of command_defs.
func (c *Command) ListCommandsForTenant(ctx context.Context, in ListCommandsInput) ([]*models.CommandDef, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := c.repo.ListCommandsForTenant(ctx, repository.ListCommandsParams{
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

// DeprecateCommand marks a command_def inactive.
func (c *Command) DeprecateCommand(ctx context.Context, id ulid.ID, updatedBy string) (*models.CommandDef, error) {
	if updatedBy == "" {
		updatedBy = "system"
	}
	cd, err := c.repo.DeprecateCommand(ctx, id, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("COMMAND_NOT_FOUND", "command "+id.String())
	}
	return cd, err
}

// ----- Uplink --------------------------------------------------------------

// EnqueueUplinkInput is the input for [Command.EnqueueUplink].
type EnqueueUplinkInput struct {
	TenantID         ulid.ID
	SatelliteID      ulid.ID
	CommandDefID     ulid.ID
	ParametersJSON   string
	ScheduledRelease time.Time
	GatewayID        string
	CreatedBy        string
}

// EnqueueUplink persists a new uplink request in QUEUED status. Validates
// that the referenced command_def exists, belongs to the same tenant, is
// active, and either targets the satellite or is tenant-wide.
func (c *Command) EnqueueUplink(ctx context.Context, in EnqueueUplinkInput) (*models.UplinkRequest, error) {
	if in.TenantID.IsZero() || in.SatelliteID.IsZero() || in.CommandDefID.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT",
			"tenant_id, satellite_id, command_def_id required")
	}
	if in.ScheduledRelease.IsZero() {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "scheduled_release required")
	}
	cd, err := c.repo.GetCommand(ctx, in.CommandDefID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, pkgerrors.BadRequest("COMMAND_NOT_FOUND",
				"command "+in.CommandDefID.String())
		}
		return nil, err
	}
	if cd.TenantID != in.TenantID {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT",
			"command_def tenant mismatch")
	}
	if !cd.Active {
		return nil, pkgerrors.New(412, "COMMAND_DEPRECATED",
			"command_def is deprecated")
	}
	if !cd.SatelliteID.IsZero() && cd.SatelliteID != in.SatelliteID {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT",
			"command_def is bound to a different satellite")
	}
	createdBy := in.CreatedBy
	if createdBy == "" {
		createdBy = "system"
	}
	return c.repo.EnqueueUplink(ctx, repository.EnqueueUplinkParams{
		ID:               c.IDFn(),
		TenantID:         in.TenantID,
		SatelliteID:      in.SatelliteID,
		CommandDefID:     in.CommandDefID,
		ParametersJSON:   in.ParametersJSON,
		ScheduledRelease: in.ScheduledRelease,
		Status:           models.StatusQueued,
		GatewayID:        in.GatewayID,
		CreatedBy:        createdBy,
	})
}

// GetUplink fetches an uplink by id.
func (c *Command) GetUplink(ctx context.Context, id ulid.ID) (*models.UplinkRequest, error) {
	u, err := c.repo.GetUplink(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("UPLINK_NOT_FOUND", "uplink "+id.String())
	}
	return u, err
}

// ListUplinksInput is the input for [Command.ListUplinksForTenant].
type ListUplinksInput struct {
	TenantID     ulid.ID
	SatelliteID  *ulid.ID
	Status       *models.UplinkStatus
	ReleaseStart time.Time
	ReleaseEnd   time.Time
	PageOffset   int32
	PageSize     int32
}

// ListUplinksForTenant returns one page of uplinks.
func (c *Command) ListUplinksForTenant(ctx context.Context, in ListUplinksInput) ([]*models.UplinkRequest, models.Page, error) {
	if in.TenantID.IsZero() {
		return nil, models.Page{}, pkgerrors.BadRequest("INVALID_ARGUMENT", "tenant_id required")
	}
	rows, total, err := c.repo.ListUplinksForTenant(ctx, repository.ListUplinksParams{
		TenantID:     in.TenantID,
		SatelliteID:  in.SatelliteID,
		Status:       in.Status,
		ReleaseStart: in.ReleaseStart,
		ReleaseEnd:   in.ReleaseEnd,
		PageOffset:   in.PageOffset,
		PageSize:     in.PageSize,
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

// CancelUplink marks an uplink CANCELED, only legal from QUEUED or RELEASED.
func (c *Command) CancelUplink(ctx context.Context, id ulid.ID, reason, updatedBy string) (*models.UplinkRequest, error) {
	msg := strings.TrimSpace(reason)
	if msg == "" {
		msg = "canceled by user"
	}
	return c.UpdateUplinkStatus(ctx, id, models.StatusCanceled, msg, updatedBy)
}

// UpdateUplinkStatus transitions an uplink to a new status. Validates the
// transition graph:
//
//	QUEUED   -> RELEASED | CANCELED
//	RELEASED -> ACKED | FAILED | CANCELED
//	ACKED    -> EXECUTED | FAILED
//	EXECUTED, FAILED, CANCELED — terminal.
func (c *Command) UpdateUplinkStatus(
	ctx context.Context, id ulid.ID, status models.UplinkStatus, errorMessage, updatedBy string,
) (*models.UplinkRequest, error) {
	if status == models.StatusUnspecified {
		return nil, pkgerrors.BadRequest("INVALID_ARGUMENT", "status required")
	}
	current, err := c.GetUplink(ctx, id)
	if err != nil {
		return nil, err
	}
	if !validUplinkTransition(current.Status, status) {
		return nil, pkgerrors.New(412, "ILLEGAL_TRANSITION",
			"illegal uplink status transition")
	}
	if updatedBy == "" {
		updatedBy = "system"
	}
	updated, err := c.repo.UpdateUplinkStatus(ctx, id, status, errorMessage, updatedBy)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, pkgerrors.NotFound("UPLINK_NOT_FOUND", "uplink "+id.String())
	}
	return updated, err
}

func validUplinkTransition(from, to models.UplinkStatus) bool {
	switch from {
	case models.StatusQueued:
		return to == models.StatusReleased || to == models.StatusCanceled
	case models.StatusReleased:
		return to == models.StatusAcked || to == models.StatusFailed || to == models.StatusCanceled
	case models.StatusAcked:
		return to == models.StatusExecuted || to == models.StatusFailed
	case models.StatusExecuted, models.StatusFailed, models.StatusCanceled:
		return false
	default:
		return false
	}
}
