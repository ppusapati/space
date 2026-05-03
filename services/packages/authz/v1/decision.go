// decision.go — single source of truth for chetana authorisation
// decisions (REQ-CONST-011: every service interceptor calls this
// one function — no service implements its own check).
//
// The decision is the AND of:
//
//   1. RBAC matches (the principal carries a role mentioned in
//      a matching allow-rule).
//   2. Clearance gate (principal's clearance >= rule's MinClearance).
//   3. ITAR gate (when the rule sets RequireUSPerson, the
//      principal's IsUSPerson flag must be true).
//   4. Deny-wins (any matching deny-rule overrides every allow).
//
// The walk is linear over priority-sorted rules and short-circuits
// on the first matching deny. Decision latency stays well under
// 1ms p99 even at 10k policies (REQ-FUNC-PLT-AUTHZ-004 — the
// micro-bench in decision_bench_test.go pins this).
//
// Every Decide call returns a Result that the caller's interceptor
// stamps onto the audit chain (REQ-FUNC-PLT-AUTHZ-004) — including
// the rule id that produced the verdict + the reason code so an
// auditor can replay the decision deterministically.

package authzv1

import (
	"errors"
	"fmt"
)

// Request is the per-call input to Decide.
type Request struct {
	Permission string // "{module}.{resource}.{action}"
	TenantID   string // optional; matched against Policy.Tenant
}

// Decision is the verdict shape Decide returns. The caller's
// interceptor lifts MatchedPolicyID + Reason into the audit chain.
type Decision struct {
	Effect          Effect
	MatchedPolicyID string
	Reason          string
}

// IsAllowed reports whether the decision permits the request.
func (d Decision) IsAllowed() bool { return d.Effect == EffectAllow }

// Reason codes — surfaced in audit events + interceptor error
// translation. Stable strings: do NOT rename without coordinating
// with the audit consumer + the SOC2 evidence tooling.
const (
	ReasonAllowedByRule   = "allowed_by_rule"
	ReasonNoMatchingAllow = "no_matching_allow"
	ReasonExplicitDeny    = "explicit_deny"
	ReasonClearance       = "insufficient_clearance"
	ReasonITAR            = "itar_us_person_required"
	ReasonNoPrincipal     = "no_principal"
	ReasonNoPermission    = "no_permission"
	ReasonNoPolicySet     = "no_policy_set"
)

// defaultDeny is the verdict when no rule matches. Default-deny
// is the chetana posture (REQ-FUNC-PLT-AUTHZ-001).
var defaultDeny = Decision{
	Effect: EffectDeny,
	Reason: ReasonNoMatchingAllow,
}

// Decide runs the policy walk and returns the verdict. The
// returned `err` is non-nil only for malformed input (nil
// principal, empty permission, nil policy set); known outcomes —
// allow / deny / no-match — are surfaced via Decision.Effect with
// a nil error so callers can handle them with normal control flow.
func Decide(principal *Principal, req Request, policies *PolicySet) (Decision, error) {
	if principal == nil {
		return Decision{Effect: EffectDeny, Reason: ReasonNoPrincipal}, ErrNoPrincipal
	}
	if req.Permission == "" {
		return Decision{Effect: EffectDeny, Reason: ReasonNoPermission}, ErrNoPermission
	}
	if policies == nil {
		return Decision{Effect: EffectDeny, Reason: ReasonNoPolicySet}, ErrNoPolicySet
	}

	principalClearance, _ := clearanceLevel(principal.ClearanceLevel)

	// Walk in priority order (already deny-first within ties).
	// The first matching DENY short-circuits.
	// For ALLOW we keep the highest-priority match and look ahead
	// only as long as a higher-priority deny might still match.
	var bestAllow *Decision
	for i := range policies.rules {
		rule := &policies.rules[i]

		// Tenant gate (cheap; do first to avoid the permission
		// match cost on the wrong tenant).
		if rule.Tenant != "" && rule.Tenant != "*" {
			if rule.Tenant != principal.TenantID {
				continue
			}
		}
		if !matchPermission(rule.Permission, req.Permission) {
			continue
		}

		switch rule.Effect {
		case EffectDeny:
			// REQ-FUNC-PLT-AUTHZ-002 deny-wins. Even if an
			// allow has already been seen at a higher priority,
			// we have to keep walking equal-or-higher priority
			// denies — but the priority sort guarantees we have
			// already seen them before any equal-priority allow.
			fires, denyReason := denyFires(rule, principal, principalClearance)
			if fires {
				return Decision{
					Effect:          EffectDeny,
					MatchedPolicyID: rule.ID,
					Reason:          denyReason,
				}, nil
			}

		case EffectAllow:
			// Skip if we already locked in an allow at a higher
			// priority — only equal-or-higher priority denies
			// can still flip the verdict.
			if bestAllow != nil {
				continue
			}
			// RBAC + clearance + ITAR gates.
			if !rbacRoleMatch(rule.Roles, principal.Roles) {
				continue
			}
			if !clearancePasses(rule, principalClearance) {
				continue
			}
			if !itarPasses(rule, principal) {
				continue
			}
			matched := Decision{
				Effect:          EffectAllow,
				MatchedPolicyID: rule.ID,
				Reason:          ReasonAllowedByRule,
			}
			bestAllow = &matched
			// We CANNOT return immediately because a
			// lower-priority deny rule that matches the same
			// permission still wins (deny-wins). Continue the
			// walk; the next iterations only short-circuit on
			// matching denies.
		}
	}

	if bestAllow != nil {
		return *bestAllow, nil
	}
	return defaultDeny, nil
}

// denyFires returns whether the deny rule applies to the principal
// AND a precise reason code when it does. Reason mapping:
//
//   • RequireUSPerson gate failed (non-US-person) → ReasonITAR
//   • MinClearance gate failed (insufficient)     → ReasonClearance
//   • Unconditional deny (no gates)               → ReasonExplicitDeny
//
// This lets the realtime gateway (and any other interceptor) emit
// a typed close code that distinguishes ITAR violations from
// generic policy denials per REQ-FUNC-RT-002.
func denyFires(rule *Policy, principal *Principal, principalClearance int) (bool, string) {
	if len(rule.Roles) > 0 && !rbacRoleMatch(rule.Roles, principal.Roles) {
		return false, ""
	}

	hasClearanceGate := rule.MinClearance != ""
	hasITARGate := rule.RequireUSPerson

	if !hasClearanceGate && !hasITARGate {
		return true, ReasonExplicitDeny
	}
	if hasITARGate && !principal.IsUSPerson {
		return true, ReasonITAR
	}
	if hasClearanceGate {
		req, _ := clearanceLevel(rule.MinClearance)
		if principalClearance < req {
			return true, ReasonClearance
		}
	}
	return false, ""
}

// rulePassesAttributes evaluates whether a deny rule should fire
// for the given principal. Semantics:
//
//   • Roles (when set) scope the deny: if the principal does NOT
//     hold any of the listed roles, the deny does NOT apply.
//
//   • MinClearance + RequireUSPerson are PROTECTION GATES on the
//     deny — when set, the deny fires ONLY if the principal
//     FAILS the gate. e.g. a deny with RequireUSPerson=true
//     fires only against non-US-persons; a US-person principal
//     is NOT denied by it.
//
//   • A deny with NO clearance/ITAR gates is an unconditional
//     deny on the matched permission (e.g. "deny launch
//     command"); applies whenever permission + role-scope match.
//
// This means an ITAR-style rule (`deny *.pass.* min_clearance=itar
// require_us_person=true`) reads as: "deny any pass-resource
// access UNLESS the principal is a US person at ITAR clearance."
func rulePassesAttributes(rule *Policy, principal *Principal, principalClearance int) bool {
	// Role scope.
	if len(rule.Roles) > 0 && !rbacRoleMatch(rule.Roles, principal.Roles) {
		return false
	}

	hasClearanceGate := rule.MinClearance != ""
	hasITARGate := rule.RequireUSPerson

	// Unconditional deny — no gates → fires on any matching
	// permission + role-scope.
	if !hasClearanceGate && !hasITARGate {
		return true
	}

	// Gates are present: the deny fires only if the principal
	// fails ANY gate (i.e. the deny PROTECTS the resource against
	// principals who don't meet the listed attributes).
	if hasClearanceGate {
		req, _ := clearanceLevel(rule.MinClearance)
		if principalClearance < req {
			return true
		}
	}
	if hasITARGate && !principal.IsUSPerson {
		return true
	}
	// All present gates passed → the principal IS authorised in
	// the eyes of this deny → deny does NOT fire.
	return false
}

// rbacRoleMatch returns true when at least one of the rule's
// roles is in the principal's role set, OR the rule has no role
// constraint (any-role).
func rbacRoleMatch(ruleRoles []string, principalRoles []string) bool {
	if len(ruleRoles) == 0 {
		return true
	}
	for _, r := range ruleRoles {
		for _, pr := range principalRoles {
			if r == pr {
				return true
			}
		}
	}
	return false
}

// clearancePasses returns true when the rule does NOT require a
// minimum clearance OR the principal meets it.
func clearancePasses(rule *Policy, principalClearance int) bool {
	if rule.MinClearance == "" {
		return true
	}
	req, _ := clearanceLevel(rule.MinClearance)
	return principalClearance >= req
}

// itarPasses returns true when the rule does NOT require a US
// person OR the principal IS a US person.
func itarPasses(rule *Policy, principal *Principal) bool {
	if !rule.RequireUSPerson {
		return true
	}
	return principal.IsUSPerson
}

// String returns a compact debug form of a Decision.
func (d Decision) String() string {
	return fmt.Sprintf("Decision{Effect:%q PolicyID:%q Reason:%q}",
		d.Effect, d.MatchedPolicyID, d.Reason)
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrNoPrincipal is returned when Decide is called with a nil
// principal. The interceptor maps this to 401 Unauthorized.
var ErrNoPrincipal = errors.New("authz: no principal supplied")

// ErrNoPermission is returned when Decide is called with an
// empty permission string.
var ErrNoPermission = errors.New("authz: no permission supplied")

// ErrNoPolicySet is returned when Decide is called with a nil
// policy set. Indicates a service started before the policy
// loader hydrated the cache.
var ErrNoPolicySet = errors.New("authz: no policy set loaded")
