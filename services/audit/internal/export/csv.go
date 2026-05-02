// csv.go — CSV export with the same signed envelope as json.go,
// emitted as a leading comment-shaped header line.
//
// Output shape:
//
//	# envelope: {... signed JSON ...}
//	id,tenant_id,event_time,actor_user_id,...
//	1,...
//	2,...
//
// CSV consumers strip the `#` line; chain-attestation tooling
// reads the same line back and verifies the envelope.

package export

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ppusapati/space/services/audit/internal/search"
)

// CSVExporter writes audit events as CSV with a signed envelope
// comment header.
type CSVExporter struct {
	search *search.Service
	pool   *pgxpool.Pool
	clk    func() time.Time
}

// NewCSVExporter wraps the search service + pool. clock=nil → time.Now.
func NewCSVExporter(s *search.Service, pool *pgxpool.Pool, clock func() time.Time) (*CSVExporter, error) {
	if s == nil {
		return nil, errors.New("export: nil search service")
	}
	if pool == nil {
		return nil, errors.New("export: nil pool")
	}
	if clock == nil {
		clock = time.Now
	}
	return &CSVExporter{search: s, pool: pool, clk: clock}, nil
}

// Export streams the events matching `q` to `w` as CSV with a
// signed envelope comment header.
func (c *CSVExporter) Export(ctx context.Context, q search.Query, w io.Writer) (*Envelope, error) {
	if q.TenantID == "" {
		return nil, errors.New("export: tenant_id is required")
	}
	tipSeq, tipHash, err := c.search.ChainTipFor(ctx, q.TenantID)
	if err != nil {
		return nil, fmt.Errorf("export: chain tip: %w", err)
	}
	// Reuse the first-last-count helper on JSONExporter via a
	// small adapter to avoid duplicating the SQL.
	helper := &JSONExporter{search: c.search, pool: c.pool, clk: c.clk}
	first, last, count, err := helper.firstLastForRange(ctx, q)
	if err != nil {
		return nil, err
	}

	env := &Envelope{
		Format:        "csv",
		TenantID:      q.TenantID,
		ExportedAt:    c.clk().UTC(),
		RowCount:      count,
		FirstChainSeq: first.ChainSeq,
		FirstRowHash:  first.RowHash,
		LastChainSeq:  last.ChainSeq,
		LastRowHash:   last.RowHash,
		ChainTipSeq:   tipSeq,
		ChainTipHash:  tipHash,
	}
	if err := env.Sign(); err != nil {
		return nil, err
	}

	envBytes, err := json.Marshal(env)
	if err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(w, "# envelope: %s\n", envBytes); err != nil {
		return nil, fmt.Errorf("export: write envelope header: %w", err)
	}

	cw := csv.NewWriter(w)
	defer cw.Flush()
	if err := cw.Write([]string{
		"id", "tenant_id", "event_time",
		"actor_user_id", "actor_session_id",
		"actor_client_ip", "actor_user_agent",
		"action", "resource", "decision", "reason",
		"matched_policy_id", "procedure", "classification",
		"metadata",
	}); err != nil {
		return nil, err
	}

	if err := c.search.Stream(ctx, q, func(h search.Hit) error {
		return cw.Write([]string{
			fmt.Sprintf("%d", h.ID),
			h.TenantID,
			h.EventTime.UTC().Format(time.RFC3339Nano),
			h.ActorUserID,
			h.ActorSessionID,
			h.ActorClientIP,
			h.ActorUserAgent,
			h.Action,
			h.Resource,
			h.Decision,
			h.Reason,
			h.MatchedPolicyID,
			h.Procedure,
			h.Classification,
			string(h.Metadata),
		})
	}); err != nil {
		return nil, err
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return nil, fmt.Errorf("export: csv flush: %w", err)
	}
	return env, nil
}
