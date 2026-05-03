package cleanup

import (
	"testing"

	"github.com/ppusapati/space/services/export/internal/queue"
	"github.com/ppusapati/space/services/export/internal/s3"
)

func TestNew_RejectsMissingDeps(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Error("expected error for empty config")
	}
	if _, err := New(Config{Store: &queue.Store{}}); err == nil {
		t.Error("expected error for missing uploader")
	}
}

func TestNew_AppliesDefaults(t *testing.T) {
	s, err := New(Config{Store: &queue.Store{}, Uploader: &s3.NopUploader{}})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if s.limit != 100 {
		t.Errorf("limit default: %d", s.limit)
	}
}
