// Package repository wraps the sat-command sqlc layer.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/samavaya/packages/ulid"

	satcmddb "github.com/ppusapati/space/services/sat-command/db/generated"
	"github.com/ppusapati/space/services/sat-command/internal/mapper"
	"github.com/ppusapati/space/services/sat-command/internal/models"
)

// ErrNotFound is returned when no row matches.
var ErrNotFound = errors.New("repository: not found")

// Repo persists CommandDefs and UplinkRequests.
type Repo struct {
	q    *satcmddb.Queries
	pool *pgxpool.Pool
}

// New constructs a Repo.
func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: satcmddb.New(pool), pool: pool}
}

// ----- CommandDef ----------------------------------------------------------

// DefineCommandParams holds the input for [Repo.DefineCommand].
type DefineCommandParams struct {
	ID               ulid.ID
	TenantID         ulid.ID
	SatelliteID      ulid.ID
	Subsystem        string
	Name             string
	Opcode           uint32
	ParametersSchema string
	Description      string
	CreatedBy        string
}

// DefineCommand inserts a new command_defs row.
func (r *Repo) DefineCommand(ctx context.Context, p DefineCommandParams) (*models.CommandDef, error) {
	row, err := r.q.DefineCommand(ctx, satcmddb.DefineCommandParams{
		ID:               mapper.PgUUID(p.ID),
		TenantID:         mapper.PgUUID(p.TenantID),
		SatelliteID:      mapper.PgUUIDOrNull(p.SatelliteID),
		Subsystem:        p.Subsystem,
		Name:             p.Name,
		Opcode:           int64(p.Opcode),
		ParametersSchema: p.ParametersSchema,
		Description:      p.Description,
		CreatedBy:        p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapper.CommandDefFromRow(row), nil
}

// GetCommand returns a command_def by id.
func (r *Repo) GetCommand(ctx context.Context, id ulid.ID) (*models.CommandDef, error) {
	row, err := r.q.GetCommand(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.CommandDefFromRow(row), nil
}

// ListCommandsParams holds the input for [Repo.ListCommandsForTenant].
type ListCommandsParams struct {
	TenantID    ulid.ID
	SatelliteID *ulid.ID
	Subsystem   string
	PageOffset  int32
	PageSize    int32
}

// ListCommandsForTenant returns one page of command defs.
func (r *Repo) ListCommandsForTenant(ctx context.Context, p ListCommandsParams) ([]*models.CommandDef, int32, error) {
	var subsystemPtr *string
	if p.Subsystem != "" {
		v := p.Subsystem
		subsystemPtr = &v
	}
	var satellitePg pgtype.UUID
	if p.SatelliteID != nil {
		satellitePg = mapper.PgUUID(*p.SatelliteID)
	}
	total, err := r.q.CountCommandsForTenant(ctx, satcmddb.CountCommandsForTenantParams{
		TenantID:    mapper.PgUUID(p.TenantID),
		SatelliteID: satellitePg,
		Subsystem:   subsystemPtr,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListCommandsForTenant(ctx, satcmddb.ListCommandsForTenantParams{
		TenantID:    mapper.PgUUID(p.TenantID),
		SatelliteID: satellitePg,
		Subsystem:   subsystemPtr,
		PageOffset:  p.PageOffset,
		PageSize:    p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.CommandDef, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.CommandDefFromRow(row))
	}
	return out, int32(total), nil
}

// DeprecateCommand marks a command_def inactive.
func (r *Repo) DeprecateCommand(ctx context.Context, id ulid.ID, updatedBy string) (*models.CommandDef, error) {
	row, err := r.q.DeprecateCommand(ctx, satcmddb.DeprecateCommandParams{
		ID:        mapper.PgUUID(id),
		UpdatedBy: updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.CommandDefFromRow(row), nil
}

// ----- Uplink --------------------------------------------------------------

// EnqueueUplinkParams holds the input for [Repo.EnqueueUplink].
type EnqueueUplinkParams struct {
	ID               ulid.ID
	TenantID         ulid.ID
	SatelliteID      ulid.ID
	CommandDefID     ulid.ID
	ParametersJSON   string
	ScheduledRelease time.Time
	Status           models.UplinkStatus
	GatewayID        string
	CreatedBy        string
}

// EnqueueUplink atomically allocates the next satellite-scoped sequence and
// inserts an uplink_requests row.
func (r *Repo) EnqueueUplink(ctx context.Context, p EnqueueUplinkParams) (*models.UplinkRequest, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.q.WithTx(tx)
	seq, err := q.NextUplinkSequence(ctx, mapper.PgUUID(p.SatelliteID))
	if err != nil {
		return nil, err
	}
	row, err := q.EnqueueUplink(ctx, satcmddb.EnqueueUplinkParams{
		ID:               mapper.PgUUID(p.ID),
		TenantID:         mapper.PgUUID(p.TenantID),
		SatelliteID:      mapper.PgUUID(p.SatelliteID),
		CommandDefID:     mapper.PgUUID(p.CommandDefID),
		ParametersJson:   p.ParametersJSON,
		ScheduledRelease: mapper.PgTimestamp(p.ScheduledRelease),
		Status:           int32(p.Status),
		SequenceNumber:   seq,
		GatewayID:        p.GatewayID,
		CreatedBy:        p.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return mapper.UplinkFromRow(row), nil
}

// GetUplink returns an uplink request by id.
func (r *Repo) GetUplink(ctx context.Context, id ulid.ID) (*models.UplinkRequest, error) {
	row, err := r.q.GetUplink(ctx, mapper.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return mapper.UplinkFromRow(row), nil
}

// ListUplinksParams holds the input for [Repo.ListUplinksForTenant].
type ListUplinksParams struct {
	TenantID     ulid.ID
	SatelliteID  *ulid.ID
	Status       *models.UplinkStatus
	ReleaseStart time.Time
	ReleaseEnd   time.Time
	PageOffset   int32
	PageSize     int32
}

// ListUplinksForTenant returns one page of uplink requests.
func (r *Repo) ListUplinksForTenant(ctx context.Context, p ListUplinksParams) ([]*models.UplinkRequest, int32, error) {
	var statusPtr *int32
	if p.Status != nil {
		v := int32(*p.Status)
		statusPtr = &v
	}
	var satellitePg pgtype.UUID
	if p.SatelliteID != nil {
		satellitePg = mapper.PgUUID(*p.SatelliteID)
	}
	total, err := r.q.CountUplinksForTenant(ctx, satcmddb.CountUplinksForTenantParams{
		TenantID:     mapper.PgUUID(p.TenantID),
		SatelliteID:  satellitePg,
		Status:       statusPtr,
		ReleaseStart: mapper.PgTimestampOrNull(p.ReleaseStart),
		ReleaseEnd:   mapper.PgTimestampOrNull(p.ReleaseEnd),
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.ListUplinksForTenant(ctx, satcmddb.ListUplinksForTenantParams{
		TenantID:     mapper.PgUUID(p.TenantID),
		SatelliteID:  satellitePg,
		Status:       statusPtr,
		ReleaseStart: mapper.PgTimestampOrNull(p.ReleaseStart),
		ReleaseEnd:   mapper.PgTimestampOrNull(p.ReleaseEnd),
		PageOffset:   p.PageOffset,
		PageSize:     p.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*models.UplinkRequest, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapper.UplinkFromRow(row))
	}
	return out, int32(total), nil
}

// UpdateUplinkStatus transitions an uplink to a new status.
func (r *Repo) UpdateUplinkStatus(
	ctx context.Context, id ulid.ID, status models.UplinkStatus, errorMessage, updatedBy string,
) (*models.UplinkRequest, error) {
	row, err := r.q.UpdateUplinkStatus(ctx, satcmddb.UpdateUplinkStatusParams{
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
	return mapper.UplinkFromRow(row), nil
}
