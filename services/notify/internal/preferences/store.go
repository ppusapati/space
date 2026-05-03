// Package preferences implements per-user notification opt-out.
//
// → REQ-FUNC-PLT-NOTIFY-003: mandatory templates ignore opt-outs.
//
// One row per (user_id, template_id) the user has explicitly
// opted out of. The absence of a row means "opted in by default."
//
// Mandatory templates (login, MFA change, password reset) are
// flagged on the template row itself; the IsAllowed check below
// short-circuits to true when the template's Mandatory bit is on.

package preferences

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store wraps a pgxpool.Pool with the preferences helpers.
type Store struct {
	pool *pgxpool.Pool
	clk  func() time.Time
}

// NewStore wraps a pool.
func NewStore(pool *pgxpool.Pool, clock func() time.Time) (*Store, error) {
	if pool == nil {
		return nil, errors.New("preferences: nil pool")
	}
	if clock == nil {
		clock = time.Now
	}
	return &Store{pool: pool, clk: clock}, nil
}

// IsAllowed reports whether `userID` has opted in to receiving
// `templateID`. Mandatory templates are ALWAYS allowed regardless
// of the preference row.
func (s *Store) IsAllowed(ctx context.Context, userID, templateID string, mandatory bool) (bool, error) {
	if mandatory {
		return true, nil
	}
	if userID == "" || templateID == "" {
		return false, errors.New("preferences: empty user_id / template_id")
	}
	var optedOut bool
	err := s.pool.QueryRow(ctx, `
SELECT opted_out FROM notification_preferences
WHERE user_id = $1 AND template_id = $2
`, userID, templateID).Scan(&optedOut)
	if err != nil {
		// pgx.ErrNoRows → no preference row → opted in by default.
		return true, nil
	}
	return !optedOut, nil
}

// SetOptOut writes (or updates) a preference row. UPSERTs on the
// (user_id, template_id) PK.
func (s *Store) SetOptOut(ctx context.Context, userID, templateID string, optedOut bool) error {
	if userID == "" || templateID == "" {
		return errors.New("preferences: empty user_id / template_id")
	}
	if _, err := s.pool.Exec(ctx, `
INSERT INTO notification_preferences (user_id, template_id, opted_out, updated_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, template_id) DO UPDATE SET
  opted_out = EXCLUDED.opted_out,
  updated_at = EXCLUDED.updated_at
`, userID, templateID, optedOut, s.clk().UTC()); err != nil {
		return fmt.Errorf("preferences: upsert: %w", err)
	}
	return nil
}
