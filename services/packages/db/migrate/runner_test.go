package migrate

import (
	"strings"
	"testing"
	"testing/fstest"
)

// TestReadMigrations_OrdersFilesLexicographically verifies that the
// embedded directory enumeration is deterministic and that Atlas
// housekeeping files (atlas.sum) are skipped.
func TestReadMigrations_OrdersFilesLexicographically(t *testing.T) {
	fs := fstest.MapFS{
		"0002_b.sql":   {Data: []byte("SELECT 2;")},
		"0001_a.sql":   {Data: []byte("SELECT 1;")},
		"0003_c.sql":   {Data: []byte("SELECT 3;")},
		"atlas.sum":    {Data: []byte("h1:abc=\n0001_a.sql h1:xyz=")},
		"notes.txt":    {Data: []byte("ignored — wrong extension")},
	}

	got, err := readMigrations(fs)
	if err != nil {
		t.Fatalf("readMigrations: %v", err)
	}

	want := []string{"0001_a", "0002_b", "0003_c"}
	if len(got) != len(want) {
		t.Fatalf("got %d migrations, want %d", len(got), len(want))
	}
	for i, m := range got {
		if m.Version != want[i] {
			t.Errorf("position %d: got %q want %q", i, m.Version, want[i])
		}
	}
}

// TestReadMigrations_DetectsTxModeNoneDirective verifies that files
// declaring the autocommit directive are flagged correctly.
func TestReadMigrations_DetectsTxModeNoneDirective(t *testing.T) {
	fs := fstest.MapFS{
		"0001_extensions.sql": {Data: []byte(
			`-- atlas:txmode none
CREATE EXTENSION IF NOT EXISTS timescaledb;`)},
		"0002_tables.sql": {Data: []byte(
			`CREATE TABLE foo (id int);`)},
	}

	got, err := readMigrations(fs)
	if err != nil {
		t.Fatalf("readMigrations: %v", err)
	}
	if !got[0].TxNone {
		t.Errorf("0001_extensions: expected TxNone=true (autocommit)")
	}
	if got[1].TxNone {
		t.Errorf("0002_tables: expected TxNone=false (transactional)")
	}
}

// TestReadMigrations_ChecksumsAreStable verifies that re-reading the
// same file produces the same checksum (essential for the drift-detection
// branch in EnsureUp).
func TestReadMigrations_ChecksumsAreStable(t *testing.T) {
	fs := fstest.MapFS{
		"0001_x.sql": {Data: []byte("CREATE TABLE x (id int);")},
	}
	a, err := readMigrations(fs)
	if err != nil {
		t.Fatalf("readMigrations a: %v", err)
	}
	b, err := readMigrations(fs)
	if err != nil {
		t.Fatalf("readMigrations b: %v", err)
	}
	if a[0].Checksum != b[0].Checksum {
		t.Errorf("checksum unstable: a=%s b=%s", a[0].Checksum, b[0].Checksum)
	}
	if a[0].Checksum == "" {
		t.Error("checksum empty")
	}
}

// TestSplitSimpleStatements_HandlesAutocommitMigration verifies that the
// statement splitter correctly cuts the extensions migration into one
// statement per CREATE EXTENSION call. This path is only exercised by
// migrations carrying the atlas:txmode none directive.
func TestSplitSimpleStatements_HandlesAutocommitMigration(t *testing.T) {
	sql := `-- atlas:txmode none
CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
`
	stmts := splitSimpleStatements(sql)
	if len(stmts) != 3 {
		t.Fatalf("got %d statements, want 3 — output: %q", len(stmts), stmts)
	}
	for i, s := range stmts {
		trim := strings.TrimSpace(s)
		if !strings.HasPrefix(trim, "CREATE EXTENSION") {
			t.Errorf("stmt %d does not start with CREATE EXTENSION: %q", i, trim)
		}
		if !strings.HasSuffix(trim, ";") {
			t.Errorf("stmt %d missing terminating semicolon: %q", i, trim)
		}
	}
}

// TestMigrationTxModeLabel verifies the human-readable label used in
// log lines.
func TestMigrationTxModeLabel(t *testing.T) {
	cases := []struct {
		tx    bool
		label string
	}{
		{true, "autocommit"},
		{false, "transactional"},
	}
	for _, c := range cases {
		got := migration{TxNone: c.tx}.TxModeLabel()
		if got != c.label {
			t.Errorf("TxNone=%v: got label %q want %q", c.tx, got, c.label)
		}
	}
}
