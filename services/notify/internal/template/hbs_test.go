package template

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestRender_HappyPath(t *testing.T) {
	r := NewRenderer()
	tmpl := &Template{
		ID:      "x",
		Version: 1,
		Body:    "Hello {{name}}!",
		Variables: VariableSchema{
			Required: []string{"name"},
		},
	}
	got, err := r.Render(tmpl, map[string]any{"name": "World"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != "Hello World!" {
		t.Errorf("got %q", got)
	}
}

// REQ-FUNC-PLT-NOTIFY-001 acceptance #1: missing variable surfaces
// as a typed error naming the offender.
func TestRender_MissingVariable_NamesOffender(t *testing.T) {
	r := NewRenderer()
	tmpl := &Template{
		ID:        "x",
		Body:      "Hello {{name}}",
		Variables: VariableSchema{Required: []string{"name", "code"}},
	}
	_, err := r.Render(tmpl, map[string]any{"name": "World"}) // missing "code"
	if err == nil {
		t.Fatal("expected error")
	}
	var missing *MissingVariableError
	if !errors.As(err, &missing) {
		t.Fatalf("expected MissingVariableError, got %T", err)
	}
	if !reflect.DeepEqual(missing.Missing, []string{"code"}) {
		t.Errorf("missing list: %v", missing.Missing)
	}
	if !strings.Contains(missing.Error(), "code") {
		t.Errorf("error message missing 'code': %v", missing.Error())
	}
}

func TestRender_EmptyStringCountsAsMissing(t *testing.T) {
	r := NewRenderer()
	tmpl := &Template{
		ID:        "x",
		Body:      "Hello {{name}}",
		Variables: VariableSchema{Required: []string{"name"}},
	}
	_, err := r.Render(tmpl, map[string]any{"name": "   "})
	if err == nil {
		t.Fatal("expected error for whitespace-only value")
	}
	var missing *MissingVariableError
	if !errors.As(err, &missing) {
		t.Fatalf("type: %T", err)
	}
}

func TestRender_NilVarsRejected(t *testing.T) {
	r := NewRenderer()
	tmpl := &Template{
		ID:        "x",
		Body:      "x",
		Variables: VariableSchema{Required: []string{"name"}},
	}
	_, err := r.Render(tmpl, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRender_NilTemplateRejected(t *testing.T) {
	r := NewRenderer()
	if _, err := r.Render(nil, map[string]any{}); err == nil {
		t.Error("expected error")
	}
}

func TestRender_CompileCacheReused(t *testing.T) {
	r := NewRenderer()
	tmpl := &Template{ID: "x", Version: 7, Body: "{{a}}", Variables: VariableSchema{Required: []string{"a"}}}
	for i := 0; i < 3; i++ {
		if _, err := r.Render(tmpl, map[string]any{"a": "b"}); err != nil {
			t.Fatalf("iter %d: %v", i, err)
		}
	}
	if len(r.cache) != 1 {
		t.Errorf("cache should be hot: %d entries", len(r.cache))
	}
}

func TestRender_VersionedCacheKey(t *testing.T) {
	r := NewRenderer()
	tmpl1 := &Template{ID: "x", Version: 1, Body: "{{a}}", Variables: VariableSchema{Required: []string{"a"}}}
	tmpl2 := &Template{ID: "x", Version: 2, Body: "{{a}} v2", Variables: VariableSchema{Required: []string{"a"}}}
	_, _ = r.Render(tmpl1, map[string]any{"a": "x"})
	_, _ = r.Render(tmpl2, map[string]any{"a": "x"})
	if len(r.cache) != 2 {
		t.Errorf("expected 2 cache entries: got %d", len(r.cache))
	}
}

func TestMissingVariables_SortedOutput(t *testing.T) {
	got := MissingVariables([]string{"z", "a", "m"}, map[string]any{})
	want := []string{"a", "m", "z"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}
