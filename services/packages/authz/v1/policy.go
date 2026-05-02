// policy.go — chetana RBAC + ABAC policy DSL.
//
// → REQ-FUNC-PLT-AUTHZ-001..004; design.md §4.1.2.
//
// Permission identifiers are dot-delimited triples:
//
//	{module}.{resource}.{action}
//
// e.g. "groundstation.pass.read", "iam.user.delete",
//      "telemetry.frame.write".
//
// Wildcards bind ONLY at segment boundaries; the matcher is greedy
// per segment but does not slide across segments. Examples:
//
//	"groundstation.pass.*"    matches every action on pass.
//	"groundstation.*.read"    matches read on every resource.
//	"*.pass.*"                matches pass-resource actions across modules.
//	"*.*.*"  / "*"            wildcards everything (reserved for the
//	                           super-admin role; deny rules with this
//	                           shape are flagged at policy-load time).
//
// A `Policy` is one rule. `PolicySet` holds the loaded rules in
// priority order (highest priority wins; ties broken by Effect with
// deny-wins). The decision engine walks the set, evaluates the
// match-by-permission + match-by-attributes predicates, and folds
// the surviving rules' effects with deny-wins semantics.

package authzv1

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Effect is the binary verdict a Policy assigns to a matching
// request.
type Effect string

// Canonical effects.
const (
	EffectAllow Effect = "allow"
	EffectDeny  Effect = "deny"
)

// Policy is one rule loaded from the YAML DSL.
//
// A request matches a Policy when ALL of:
//
//   • Permission matches the request's permission (wildcards
//     evaluated per-segment).
//   • Roles is empty OR the principal carries at least one of
//     the listed roles.
//   • MinClearance is empty OR the principal's clearance is
//     >= MinClearance per the canonical ladder
//     (public < internal < restricted < cui < itar).
//   • RequireUSPerson is false OR the principal's IsUSPerson
//     is true.
//   • Tenant matches (or is empty / wildcard).
type Policy struct {
	// ID is a stable identifier the audit chain references in
	// every allow/deny event so operators can trace the decision
	// back to the rule that produced it.
	ID string `yaml:"id"`

	// Description is free text shown in `chetanactl policy ls`.
	Description string `yaml:"description"`

	// Effect is "allow" or "deny".
	Effect Effect `yaml:"effect"`

	// Priority — higher wins. Defaults to 0.
	Priority int `yaml:"priority"`

	// Permission pattern. Required.
	Permission string `yaml:"permission"`

	// Roles the principal must carry (any-of). Empty = any role
	// (subject to the other predicates).
	Roles []string `yaml:"roles"`

	// MinClearance: principal must hold this or higher per the
	// canonical clearance ladder.
	MinClearance string `yaml:"min_clearance"`

	// RequireUSPerson: when true, denies non-US-person principals
	// regardless of effect. The most common use: ITAR data tagged
	// at restricted+/itar.
	RequireUSPerson bool `yaml:"require_us_person"`

	// Tenant scopes the rule to a specific tenant. Empty / "*"
	// matches every tenant (the single-tenant deployment ships
	// everything as "*").
	Tenant string `yaml:"tenant"`

	// Notes (optional). Surfaced in audit events when populated
	// so reviewers can see human context for an allow/deny.
	Notes string `yaml:"notes"`
}

// PolicySet is an immutable, priority-sorted bundle of policies.
type PolicySet struct {
	rules []Policy
}

// NewPolicySet validates + sorts the supplied rules. Returns an
// error when any rule is malformed.
func NewPolicySet(rules []Policy) (*PolicySet, error) {
	out := make([]Policy, 0, len(rules))
	for i, r := range rules {
		if err := validatePolicy(r); err != nil {
			return nil, fmt.Errorf("policy[%d] %q: %w", i, r.ID, err)
		}
		out = append(out, r)
	}
	// Highest priority first; ties broken by deny-first so the
	// linear scan can short-circuit on the first matching deny.
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Priority != out[j].Priority {
			return out[i].Priority > out[j].Priority
		}
		if out[i].Effect != out[j].Effect {
			return out[i].Effect == EffectDeny
		}
		return out[i].ID < out[j].ID
	})
	return &PolicySet{rules: out}, nil
}

// Rules returns a defensive copy of the loaded rules in their
// canonical sort order.
func (p *PolicySet) Rules() []Policy {
	out := make([]Policy, len(p.rules))
	copy(out, p.rules)
	return out
}

// Len returns the number of loaded rules.
func (p *PolicySet) Len() int { return len(p.rules) }

// LoadPoliciesYAML parses a YAML document containing a top-level
// `policies:` array.
func LoadPoliciesYAML(body []byte) (*PolicySet, error) {
	var doc struct {
		Policies []Policy `yaml:"policies"`
	}
	if err := yaml.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("authz: parse policies yaml: %w", err)
	}
	return NewPolicySet(doc.Policies)
}

// validatePolicy enforces the per-rule invariants.
func validatePolicy(p Policy) error {
	if p.ID == "" {
		return errors.New("id is required")
	}
	if p.Permission == "" {
		return errors.New("permission is required")
	}
	switch p.Effect {
	case EffectAllow, EffectDeny:
		// ok
	case "":
		return errors.New("effect is required (allow|deny)")
	default:
		return fmt.Errorf("unknown effect %q", p.Effect)
	}
	if !validPermissionPattern(p.Permission) {
		return fmt.Errorf("permission %q is not a valid {module}.{resource}.{action} triple", p.Permission)
	}
	if p.MinClearance != "" {
		if _, ok := clearanceLevel(p.MinClearance); !ok {
			return fmt.Errorf("unknown min_clearance %q", p.MinClearance)
		}
	}
	return nil
}

// matchPermission tests whether `requested` (concrete) matches
// `pattern` (may contain "*" segments). Both are split on "."; a
// "*" in the pattern matches any single segment value.
//
// A pattern of "*" alone matches every permission (used for
// the super-admin role).
func matchPermission(pattern, requested string) bool {
	if pattern == "*" || pattern == requested {
		return true
	}
	pp := strings.Split(pattern, ".")
	rp := strings.Split(requested, ".")
	if len(pp) != len(rp) {
		return false
	}
	for i := range pp {
		if pp[i] == "*" {
			continue
		}
		if pp[i] != rp[i] {
			return false
		}
	}
	return true
}

// validPermissionPattern checks the segment shape. Each segment
// must be either "*" or a non-empty alphanumeric+underscore
// identifier. The pattern itself must have either 1 segment ("*")
// or 3 segments.
func validPermissionPattern(p string) bool {
	if p == "*" {
		return true
	}
	parts := strings.Split(p, ".")
	if len(parts) != 3 {
		return false
	}
	for _, s := range parts {
		if s == "" {
			return false
		}
		if s == "*" {
			continue
		}
		for i := 0; i < len(s); i++ {
			c := s[i]
			switch {
			case c >= 'a' && c <= 'z',
				c >= 'A' && c <= 'Z',
				c >= '0' && c <= '9',
				c == '_', c == '-':
				// ok
			default:
				return false
			}
		}
	}
	return true
}

// clearanceLevel returns the integer rank of the named clearance.
// public=0, internal=1, restricted=2, cui=3, itar=4. The boolean
// is false for unknown levels.
func clearanceLevel(name string) (int, bool) {
	switch name {
	case "public":
		return 0, true
	case "internal":
		return 1, true
	case "restricted":
		return 2, true
	case "cui":
		return 3, true
	case "itar":
		return 4, true
	}
	return 0, false
}
