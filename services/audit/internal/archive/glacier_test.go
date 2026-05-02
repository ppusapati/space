package archive

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNopArchiver_RejectsEmptyTenant(t *testing.T) {
	if _, err := (NopArchiver{}).Upload(context.Background(), UploadInput{}); err == nil {
		t.Error("expected error for empty tenant_id")
	}
}

func TestNopArchiver_DeterministicKey(t *testing.T) {
	in := UploadInput{
		TenantID:   "11111111-1111-1111-1111-111111111111",
		RangeStart: time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
		Body:       []byte("payload"),
	}
	a, err := (NopArchiver{}).Upload(context.Background(), in)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	b, err := (NopArchiver{}).Upload(context.Background(), in)
	if err != nil {
		t.Fatalf("upload 2: %v", err)
	}
	if a.Key != b.Key {
		t.Errorf("nondeterministic: %q vs %q", a.Key, b.Key)
	}
	if !strings.Contains(a.Key, "11111111") {
		t.Errorf("key missing tenant: %q", a.Key)
	}
	if !strings.HasSuffix(a.Key, ".json") {
		t.Errorf("key suffix: %q", a.Key)
	}
	if a.StorageClass != "GLACIER" {
		t.Errorf("storage class: %q", a.StorageClass)
	}
	if a.BytesCompressed != int64(len(in.Body)) {
		t.Errorf("bytes: %d want %d", a.BytesCompressed, len(in.Body))
	}
}

func TestNopArchiver_BucketDefaultsToChetanaAuditCold(t *testing.T) {
	res, _ := (NopArchiver{}).Upload(context.Background(), UploadInput{
		TenantID: "u", RangeStart: time.Now(),
	})
	if res.Bucket != "chetana-audit-cold" {
		t.Errorf("default bucket: %q", res.Bucket)
	}
}

func TestNopArchiver_RespectsConfiguredBucket(t *testing.T) {
	res, _ := (NopArchiver{Bucket: "my-bucket"}).Upload(context.Background(), UploadInput{
		TenantID: "u", RangeStart: time.Now(),
	})
	if res.Bucket != "my-bucket" {
		t.Errorf("bucket: %q", res.Bucket)
	}
}
