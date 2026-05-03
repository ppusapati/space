package s3

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestFIPSAsserts(t *testing.T) {
	good := []string{
		"https://s3-fips.us-east-1.amazonaws.com",
		"https://s3-fips.dualstack.us-east-2.amazonaws.com",
	}
	for _, ep := range good {
		if err := FIPSAsserts(ep); err != nil {
			t.Errorf("FIPSAsserts(%q): %v", ep, err)
		}
	}
	bad := []string{
		"",
		"https://s3.us-east-1.amazonaws.com",
		"https://example.com",
	}
	for _, ep := range bad {
		if err := FIPSAsserts(ep); err == nil {
			t.Errorf("FIPSAsserts(%q): expected error", ep)
		}
	}
}

func TestNopUploader_RoundtripUploadDelete(t *testing.T) {
	u := &NopUploader{Bucket: "test-bucket"}
	res, err := u.Upload(context.Background(), UploadInput{
		Key: "k", Body: []byte("hello"),
	})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if res.Bucket != "test-bucket" || res.Key != "k" || res.BytesTotal != 5 {
		t.Errorf("upload result: %+v", res)
	}
	if !strings.HasPrefix(res.ETag, "nop-") {
		t.Errorf("etag: %q", res.ETag)
	}
	if got := u.Stored["test-bucket/k"]; string(got) != "hello" {
		t.Errorf("stored: %s", got)
	}

	url, expiresAt, err := u.Presign(context.Background(), "test-bucket", "k", time.Hour)
	if err != nil {
		t.Fatalf("presign: %v", err)
	}
	if !strings.Contains(url, "test-bucket/k") {
		t.Errorf("presigned url: %q", url)
	}
	if !expiresAt.After(time.Now()) {
		t.Errorf("expiresAt: %v", expiresAt)
	}

	if err := u.Delete(context.Background(), "test-bucket", "k"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !u.Deleted["test-bucket/k"] {
		t.Error("delete record")
	}
	if _, ok := u.Stored["test-bucket/k"]; ok {
		t.Error("Stored should be purged after Delete")
	}
}

func TestNopUploader_RejectsEmptyBucket(t *testing.T) {
	u := &NopUploader{} // no default bucket
	if _, err := u.Upload(context.Background(), UploadInput{Key: "k"}); err == nil {
		t.Error("expected error for empty bucket")
	}
}

func TestNopUploader_PerCallBucketTakesPrecedence(t *testing.T) {
	u := &NopUploader{} // no default
	res, err := u.Upload(context.Background(), UploadInput{Bucket: "explicit", Key: "k", Body: []byte("x")})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if res.Bucket != "explicit" {
		t.Errorf("bucket: %q", res.Bucket)
	}
}

func TestNopUploader_ETagDeterministic(t *testing.T) {
	u := &NopUploader{Bucket: "b"}
	a, _ := u.Upload(context.Background(), UploadInput{Key: "k1", Body: []byte("x")})
	b, _ := u.Upload(context.Background(), UploadInput{Key: "k2", Body: []byte("x")})
	if a.ETag != b.ETag {
		t.Errorf("same body should produce same etag: %q vs %q", a.ETag, b.ETag)
	}
}
