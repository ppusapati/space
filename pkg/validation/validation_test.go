package validation

import (
	"testing"

	"google.golang.org/protobuf/types/known/emptypb"
)

func TestDefaultIsSingleton(t *testing.T) {
	v1, err := Default()
	if err != nil {
		t.Fatalf("Default: %v", err)
	}
	v2, err := Default()
	if err != nil {
		t.Fatalf("Default: %v", err)
	}
	if v1 != v2 {
		t.Fatal("expected same validator instance")
	}
}

func TestValidateOnUnconstrainedMessageSucceeds(t *testing.T) {
	if err := Validate(&emptypb.Empty{}); err != nil {
		t.Fatalf("Validate(&Empty{}) returned: %v", err)
	}
}

func TestIsRecognisesValidationError(t *testing.T) {
	if Is(nil) {
		t.Fatal("Is(nil) should be false")
	}
	ve := &ValidationError{Cause: nil}
	if !Is(ve) {
		t.Fatal("Is(*ValidationError) must be true")
	}
}
