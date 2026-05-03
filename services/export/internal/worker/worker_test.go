package worker

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ppusapati/space/services/export/internal/queue"
	"github.com/ppusapati/space/services/export/internal/s3"
)

func TestRegistry_Register_Lookup(t *testing.T) {
	r := NewRegistry()
	called := false
	r.Register("test", ProcessFunc(func(ctx context.Context, j *queue.Job) (ProcessOutput, error) {
		called = true
		return ProcessOutput{}, nil
	}))
	p, err := r.Lookup("test")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	_, _ = p.Process(context.Background(), &queue.Job{})
	if !called {
		t.Error("processor not invoked")
	}
}

func TestRegistry_LookupUnknown(t *testing.T) {
	r := NewRegistry()
	_, err := r.Lookup("missing")
	if !errors.Is(err, ErrNoProcessor) {
		t.Errorf("got %v want ErrNoProcessor", err)
	}
}

func TestNew_RejectsMissingDeps(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*Config)
	}{
		{"no store", func(c *Config) { c.Store = nil }},
		{"no uploader", func(c *Config) { c.Uploader = nil }},
		{"no registry", func(c *Config) { c.Registry = nil }},
		{"no bucket", func(c *Config) { c.Bucket = "" }},
	}
	good := Config{
		Store: &queue.Store{}, Uploader: &s3.NopUploader{},
		Registry: NewRegistry(), Bucket: "b",
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := good
			c.mut(&cfg)
			if _, err := New(cfg); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestNew_AppliesDefaults(t *testing.T) {
	w, err := New(Config{
		Store:    &queue.Store{},
		Uploader: &s3.NopUploader{},
		Registry: NewRegistry(),
		Bucket:   "b",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if w.id == "" {
		t.Error("id default")
	}
	if w.leaseTTL == 0 {
		t.Error("lease ttl default")
	}
	if w.pollInterval == 0 {
		t.Error("poll interval default")
	}
	if w.presignedFor != 24*time.Hour {
		t.Errorf("presigned default: %v", w.presignedFor)
	}
}

func TestComposeKey_Stable(t *testing.T) {
	job := &queue.Job{
		ID: "job-123", TenantID: "tenant-1", Kind: "audit_csv",
		EnqueuedAt: time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC),
	}
	got := composeKey(job, "events.csv")
	want := "exports/tenant-1/audit_csv/2026/05/job-123-events.csv"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestComposeKey_DefaultFilename(t *testing.T) {
	job := &queue.Job{ID: "j", TenantID: "t", Kind: "k", EnqueuedAt: time.Now().UTC()}
	got := composeKey(job, "")
	if !strings.HasSuffix(got, "j-output.bin") {
		t.Errorf("default filename: %q", got)
	}
}

func TestSanitiseKind(t *testing.T) {
	cases := map[string]string{
		"audit_csv":     "audit_csv",
		"audit-csv":     "audit-csv",
		"audit/csv":     "audit_csv",
		"audit csv":     "audit_csv",
		"audit.csv.v2":  "audit_csv_v2",
		"":              "unknown",
	}
	for in, want := range cases {
		if got := sanitiseKind(in); got != want {
			t.Errorf("sanitiseKind(%q): %q want %q", in, got, want)
		}
	}
}
