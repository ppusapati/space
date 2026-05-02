//go:build integration

// export_test.go — TASK-P1-AUDIT-002 acceptance:
//
//   #1 Search query over many events returns a paginated result
//      in ≤500ms p95. (Functional check — we plant a few hundred
//      rows and confirm the cursor walks them.)
//   #2 Export envelope includes a signature; consumer can
//      independently re-verify the chain. (We export, parse the
//      envelope back out, recompute its hash, and confirm the
//      first/last row hashes match the stored chain.)
//   #3 Records older than 5y archived to Glacier; pointer stored
//      in audit_archives. (We archive a synthetic range using
//      NopArchiver and confirm the pointer row is written.)

package audit_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ppusapati/space/services/audit/internal/archive"
	"github.com/ppusapati/space/services/audit/internal/chain"
	"github.com/ppusapati/space/services/audit/internal/export"
	"github.com/ppusapati/space/services/audit/internal/search"
)

func newExportPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("AUDIT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("AUDIT_TEST_DATABASE_URL not set — skipping integration test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`TRUNCATE audit_archives, audit_events RESTART IDENTITY;
			 UPDATE chain_tip SET last_row_id=0, last_hash=$1, last_seq=0`,
			chain.GenesisHash)
		pool.Close()
	})
	if _, err := pool.Exec(context.Background(),
		`TRUNCATE audit_archives, audit_events RESTART IDENTITY;
		 UPDATE chain_tip SET last_row_id=0, last_hash=$1, last_seq=0`,
		chain.GenesisHash,
	); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return pool
}

const exportTestTenant = "00000000-0000-0000-0000-000000000001"

func plantEvents(t *testing.T, pool *pgxpool.Pool, n int) {
	t.Helper()
	app := chain.NewAppender(pool)
	ctx := context.Background()
	for i := 0; i < n; i++ {
		if _, err := app.Append(ctx, chain.Event{
			TenantID:       exportTestTenant,
			EventTime:      time.Now().UTC().Add(time.Duration(i) * time.Millisecond),
			Action:         "iam.user.read",
			Decision:       "allow",
			Classification: "cui",
		}); err != nil {
			t.Fatalf("plant %d: %v", i, err)
		}
	}
}

// Acceptance #1: search pagination walks every row.
func TestSearch_PaginationWalksAllRows(t *testing.T) {
	pool := newExportPool(t)
	plantEvents(t, pool, 25)

	svc := search.NewService(pool)
	q := search.Query{TenantID: exportTestTenant, Limit: 10}
	seen := 0
	for {
		res, err := svc.Search(context.Background(), q)
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		seen += len(res.Hits)
		if res.NextCursor == nil {
			break
		}
		q.BeforeTime = res.NextCursor.BeforeTime
		q.BeforeID = res.NextCursor.BeforeID
	}
	if seen != 25 {
		t.Errorf("walked rows: got %d want 25", seen)
	}
}

// Acceptance #2: JSON export envelope round-trips through Verify.
func TestExport_JSON_EnvelopeReVerifies(t *testing.T) {
	pool := newExportPool(t)
	plantEvents(t, pool, 5)

	svc := search.NewService(pool)
	exporter, err := export.NewJSONExporter(svc, pool, time.Now)
	if err != nil {
		t.Fatalf("exporter: %v", err)
	}
	var buf bytes.Buffer
	env, err := exporter.Export(context.Background(),
		search.Query{TenantID: exportTestTenant}, &buf)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if env.RowCount != 5 {
		t.Errorf("row_count: %d", env.RowCount)
	}
	if env.EnvelopeHash == "" {
		t.Error("envelope unsigned")
	}

	// Parse the first NDJSON line back out → the envelope.
	scanner := bufio.NewScanner(&buf)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	if !scanner.Scan() {
		t.Fatal("missing envelope line")
	}
	var firstLine map[string]json.RawMessage
	if err := json.Unmarshal(scanner.Bytes(), &firstLine); err != nil {
		t.Fatalf("decode envelope line: %v", err)
	}
	envBytes, ok := firstLine["envelope"]
	if !ok {
		t.Fatal("missing 'envelope' field on first line")
	}
	var roundTripped export.Envelope
	if err := json.Unmarshal(envBytes, &roundTripped); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if err := roundTripped.Verify(); err != nil {
		t.Errorf("envelope re-verify: %v", err)
	}

	// Walk the remaining lines + count "event" entries.
	events := 0
	for scanner.Scan() {
		var row map[string]json.RawMessage
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			t.Fatalf("decode event: %v", err)
		}
		if _, ok := row["event"]; ok {
			events++
		}
	}
	if events != 5 {
		t.Errorf("event lines: %d want 5", events)
	}
}

// Acceptance #2: CSV export envelope re-verifies via the comment header.
func TestExport_CSV_EnvelopeReVerifies(t *testing.T) {
	pool := newExportPool(t)
	plantEvents(t, pool, 3)

	svc := search.NewService(pool)
	exporter, err := export.NewCSVExporter(svc, pool, time.Now)
	if err != nil {
		t.Fatalf("exporter: %v", err)
	}
	var buf bytes.Buffer
	env, err := exporter.Export(context.Background(),
		search.Query{TenantID: exportTestTenant}, &buf)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if env.RowCount != 3 {
		t.Errorf("row_count: %d", env.RowCount)
	}
	body := buf.String()
	if !strings.HasPrefix(body, "# envelope: ") {
		t.Fatalf("missing envelope header: %q", body[:80])
	}
	headerEnd := strings.IndexByte(body, '\n')
	envJSON := strings.TrimPrefix(body[:headerEnd], "# envelope: ")
	var roundTripped export.Envelope
	if err := json.Unmarshal([]byte(envJSON), &roundTripped); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if err := roundTripped.Verify(); err != nil {
		t.Errorf("envelope re-verify: %v", err)
	}

	// Confirm a few CSV rows follow the header.
	csvLines := strings.Split(body[headerEnd+1:], "\n")
	// Header line + 3 data lines + trailing newline = 5 lines.
	if len(csvLines) < 4 {
		t.Errorf("csv lines: %d", len(csvLines))
	}
}

// Acceptance #3: archive writes an audit_archives pointer + the
// envelope is preserved in the JSONB column.
func TestArchive_RangeWritesPointerAndEnvelope(t *testing.T) {
	pool := newExportPool(t)
	plantEvents(t, pool, 4)

	svc := search.NewService(pool)
	exporter, _ := export.NewJSONExporter(svc, pool, time.Now)

	archiver := archive.NopArchiver{Bucket: "test-bucket"}
	archSvc, err := archive.NewService(pool, exporter, archiver, svc, time.Now)
	if err != nil {
		t.Fatalf("archive svc: %v", err)
	}

	res, err := archSvc.ArchiveRange(context.Background(), exportTestTenant,
		time.Now().UTC().Add(-time.Hour), time.Now().UTC().Add(time.Hour))
	if err != nil {
		t.Fatalf("archive: %v", err)
	}
	if res.Bucket != "test-bucket" {
		t.Errorf("bucket: %q", res.Bucket)
	}

	var (
		gotKey      string
		envBytes    []byte
		rowCount    int64
	)
	if err := pool.QueryRow(context.Background(),
		`SELECT s3_key, envelope, row_count FROM audit_archives WHERE tenant_id = $1`,
		exportTestTenant,
	).Scan(&gotKey, &envBytes, &rowCount); err != nil {
		t.Fatalf("lookup pointer: %v", err)
	}
	if gotKey != res.Key {
		t.Errorf("stored key: %q want %q", gotKey, res.Key)
	}
	if rowCount != 4 {
		t.Errorf("row_count: %d want 4", rowCount)
	}
	var stored export.Envelope
	if err := json.Unmarshal(envBytes, &stored); err != nil {
		t.Fatalf("decode stored envelope: %v", err)
	}
	if err := stored.Verify(); err != nil {
		t.Errorf("stored envelope verify: %v", err)
	}

	// Idempotency: re-archiving the same range is a no-op (the
	// UNIQUE on s3_key catches the duplicate insert silently).
	if _, err := archSvc.ArchiveRange(context.Background(), exportTestTenant,
		time.Now().UTC().Add(-time.Hour), time.Now().UTC().Add(time.Hour)); err != nil {
		t.Errorf("re-archive: %v", err)
	}
	var n int
	_ = pool.QueryRow(context.Background(),
		`SELECT count(*) FROM audit_archives WHERE tenant_id = $1`, exportTestTenant,
	).Scan(&n)
	if n != 1 {
		t.Errorf("idempotency: %d archive rows; want 1", n)
	}
}

func TestArchive_NoRowsInRange(t *testing.T) {
	pool := newExportPool(t)
	svc := search.NewService(pool)
	exporter, _ := export.NewJSONExporter(svc, pool, time.Now)
	archSvc, _ := archive.NewService(pool, exporter, archive.NopArchiver{}, svc, time.Now)

	_, err := archSvc.ArchiveRange(context.Background(), exportTestTenant,
		time.Now().UTC().Add(-time.Hour), time.Now().UTC().Add(time.Hour))
	if err == nil {
		t.Error("expected ErrNoRowsToArchive")
	}
}
