package saml

import (
	"reflect"
	"testing"
)

func TestProjectRoles_GroupMapping(t *testing.T) {
	m := AttributeMapping{
		GroupsAttribute: "http://schemas.xmlsoap.org/claims/Group",
		GroupRoleMap: map[string]string{
			"chetana-operators":     "operator",
			"chetana-admins":        "admin",
			"chetana-mission-leads": "mission_lead",
		},
		DefaultRoles: []string{"viewer"},
	}
	attrs := map[string][]string{
		"http://schemas.xmlsoap.org/claims/Group": {
			"chetana-operators",
			"chetana-mission-leads",
			"unmapped-group",
		},
	}
	got := projectRoles(m, attrs)
	want := []string{"operator", "mission_lead", "viewer"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestProjectRoles_NoGroupsAttribute(t *testing.T) {
	m := AttributeMapping{
		DefaultRoles: []string{"viewer"},
	}
	got := projectRoles(m, map[string][]string{
		"some-other-attr": {"x", "y"},
	})
	if !reflect.DeepEqual(got, []string{"viewer"}) {
		t.Errorf("got %v", got)
	}
}

func TestProjectRoles_DeduplicatesUnion(t *testing.T) {
	m := AttributeMapping{
		GroupsAttribute: "groups",
		GroupRoleMap:    map[string]string{"engineers": "engineer"},
		DefaultRoles:    []string{"engineer", "viewer"},
	}
	got := projectRoles(m, map[string][]string{
		"groups": {"engineers", "engineers"},
	})
	want := []string{"engineer", "viewer"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestRequireEmail(t *testing.T) {
	m := AttributeMapping{EmailAttribute: "email"}
	got, err := requireEmail(m, map[string][]string{"email": {"user@example.com"}})
	if err != nil || got != "user@example.com" {
		t.Errorf("got (%q,%v) want (user@example.com, nil)", got, err)
	}

	// Multi-valued: first non-empty wins.
	got, _ = requireEmail(m, map[string][]string{"email": {"  ", "user@example.com"}})
	if got != "user@example.com" {
		t.Errorf("multi-valued: %q", got)
	}

	// Missing attribute.
	if _, err := requireEmail(m, map[string][]string{"other": {"x"}}); err != ErrMissingEmail {
		t.Errorf("missing: got %v", err)
	}

	// Empty mapping config.
	if _, err := requireEmail(AttributeMapping{}, map[string][]string{}); err != ErrMissingEmail {
		t.Errorf("no email attr config: got %v", err)
	}
}

func TestFirstAttribute(t *testing.T) {
	if got := firstAttribute("display", map[string][]string{"display": {"User Example"}}); got != "User Example" {
		t.Errorf("got %q", got)
	}
	if got := firstAttribute("display", map[string][]string{"display": {"  ", "User"}}); got != "User" {
		t.Errorf("trim: got %q", got)
	}
	if got := firstAttribute("", map[string][]string{}); got != "" {
		t.Errorf("empty name: got %q", got)
	}
	if got := firstAttribute("missing", map[string][]string{"other": {"x"}}); got != "" {
		t.Errorf("missing: got %q", got)
	}
}

func TestDisplayOrEmail(t *testing.T) {
	if got := displayOrEmail("User Example", "u@x"); got != "User Example" {
		t.Errorf("got %q", got)
	}
	if got := displayOrEmail("", "u@x"); got != "u@x" {
		t.Errorf("fallback: got %q", got)
	}
}
