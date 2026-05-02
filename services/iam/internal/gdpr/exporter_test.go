package gdpr

import (
	"context"
	"strings"
	"testing"
)

func TestNopExporter_RejectsEmptyUserID(t *testing.T) {
	if _, err := (NopExporter{}).EnqueueSAR(context.Background(), EnqueueSARInput{}); err == nil {
		t.Error("expected error for empty user_id")
	}
}

func TestNopExporter_ReturnsSyntheticJobID(t *testing.T) {
	id, err := (NopExporter{}).EnqueueSAR(context.Background(), EnqueueSARInput{
		UserID: "11111111-1111-1111-1111-111111111111",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.HasPrefix(string(id), "nop-") {
		t.Errorf("job id: %q", id)
	}
	if !strings.Contains(string(id), "11111111") {
		t.Errorf("job id missing user_id: %q", id)
	}
}
