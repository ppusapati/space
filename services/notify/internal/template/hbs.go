// Package template implements the Handlebars-style renderer the
// notify service uses for email + SMS + in-app payloads.
//
// → REQ-FUNC-PLT-NOTIFY-001 acceptance #1: missing variable →
//   400 with the variable name (NEVER an empty rendered field).
// → design.md §4.7.
//
// Templates live in the `notification_templates` table. Each row
// carries:
//
//   • body              — the Handlebars source.
//   • variables_schema  — JSONB list of required variable names
//                          (e.g. ["user_email", "reset_link"]).
//   • mandatory bool    — when true, the user's notification
//                          preferences cannot opt out of this
//                          template (REQ-FUNC-PLT-NOTIFY-003).
//
// At render time we validate the supplied variables against the
// schema BEFORE invoking the renderer. A missing variable returns
// MissingVariableError naming the offender; an unexpected
// variable is allowed (templates can ignore extra context).

package template

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/aymerick/raymond"
)

// VariableSchema is the validated shape of a template's required
// variables. Stored as JSONB in `notification_templates.variables_schema`.
type VariableSchema struct {
	Required []string `json:"required"`
}

// Template is the in-memory shape of one row.
type Template struct {
	ID        string
	Version   int
	Channel   string // "email" | "sms" | "inapp"
	Body      string
	Variables VariableSchema
	Mandatory bool
}

// Renderer compiles + renders templates.
type Renderer struct {
	cache map[string]*raymond.Template
}

// NewRenderer returns a Renderer with an empty compile cache.
func NewRenderer() *Renderer {
	return &Renderer{cache: make(map[string]*raymond.Template)}
}

// Render validates `vars` against the template's schema and runs
// the Handlebars expansion. Returns the rendered body or an error.
//
// The schema check is BEFORE the render so a missing variable
// surfaces as a typed error naming the offender — not as an empty
// rendered field that an attacker could exploit (acceptance #1).
func (r *Renderer) Render(t *Template, vars map[string]any) (string, error) {
	if t == nil {
		return "", errors.New("template: nil template")
	}
	if vars == nil {
		vars = map[string]any{}
	}

	if missing := MissingVariables(t.Variables.Required, vars); len(missing) > 0 {
		return "", &MissingVariableError{
			TemplateID: t.ID,
			Missing:    missing,
		}
	}

	cacheKey := fmt.Sprintf("%s@%d", t.ID, t.Version)
	tmpl, ok := r.cache[cacheKey]
	if !ok {
		compiled, err := raymond.Parse(t.Body)
		if err != nil {
			return "", fmt.Errorf("template: parse %q: %w", t.ID, err)
		}
		tmpl = compiled
		r.cache[cacheKey] = tmpl
	}

	out, err := tmpl.Exec(vars)
	if err != nil {
		return "", fmt.Errorf("template: exec %q: %w", t.ID, err)
	}
	return out, nil
}

// MissingVariables returns the names in `required` that have no
// non-empty value in `vars`. Empty string + nil + missing key all
// count as missing — we want acceptance #1's "never an empty
// rendered field" guarantee to hold even when the caller passes
// `vars["foo"] = ""`.
func MissingVariables(required []string, vars map[string]any) []string {
	var missing []string
	for _, k := range required {
		v, ok := vars[k]
		if !ok || v == nil {
			missing = append(missing, k)
			continue
		}
		s, isStr := v.(string)
		if isStr && strings.TrimSpace(s) == "" {
			missing = append(missing, k)
		}
	}
	sort.Strings(missing)
	return missing
}

// MissingVariableError is returned from Render when one or more
// required variables are absent. Implements `error` and exposes
// the offending names so the caller can produce a 400 with the
// variable name in the body (acceptance #1).
type MissingVariableError struct {
	TemplateID string
	Missing    []string
}

// Error implements error.
func (e *MissingVariableError) Error() string {
	return fmt.Sprintf("template %q: missing required variable(s): %s",
		e.TemplateID, strings.Join(e.Missing, ", "))
}
