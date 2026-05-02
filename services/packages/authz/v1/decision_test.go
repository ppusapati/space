package authzv1

import (
	"errors"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------
// matchPermission — segment-aligned wildcard matcher
// ----------------------------------------------------------------------

func TestMatchPermission(t *testing.T) {
	cases := []struct {
		pattern, requested string
		want               bool
	}{
		// exact
		{"groundstation.pass.read", "groundstation.pass.read", true},
		{"groundstation.pass.read", "groundstation.pass.write", false},
		// trailing wildcard
		{"groundstation.pass.*", "groundstation.pass.read", true},
		{"groundstation.pass.*", "groundstation.pass.write", true},
		{"groundstation.pass.*", "groundstation.frame.read", false},
		// middle wildcard
		{"groundstation.*.read", "groundstation.pass.read", true},
		{"groundstation.*.read", "groundstation.frame.read", true},
		{"groundstation.*.read", "groundstation.frame.write", false},
		// leading wildcard
		{"*.pass.*", "groundstation.pass.read", true},
		{"*.pass.*", "ops.pass.write", true},
		{"*.pass.*", "groundstation.frame.read", false},
		// global wildcard
		{"*", "anything.at.all", true},
		// segment-count mismatch
		{"a.b", "a.b.c", false},
		{"a.b.c.d", "a.b.c", false},
	}
	for _, tc := range cases {
		t.Run(tc.pattern+"_vs_"+tc.requested, func(t *testing.T) {
			if got := matchPermission(tc.pattern, tc.requested); got != tc.want {
				t.Errorf("matchPermission(%q, %q) = %v, want %v", tc.pattern, tc.requested, got, tc.want)
			}
		})
	}
}

func TestValidPermissionPattern(t *testing.T) {
	good := []string{"*", "iam.user.read", "*.pass.*", "module.*.action", "a_b.c-d.e"}
	for _, p := range good {
		if !validPermissionPattern(p) {
			t.Errorf("expected valid: %q", p)
		}
	}
	bad := []string{"", "two.parts", "a..b", ".missing.first", "missing.last.", "iam.user.read!"}
	for _, p := range bad {
		if validPermissionPattern(p) {
			t.Errorf("expected invalid: %q", p)
		}
	}
}

// ----------------------------------------------------------------------
// PolicySet validation + sort
// ----------------------------------------------------------------------

func TestNewPolicySet_RejectsBadRules(t *testing.T) {
	_, err := NewPolicySet([]Policy{{Permission: "iam.user.read", Effect: "allow"}})
	if err == nil || !strings.Contains(err.Error(), "id is required") {
		t.Errorf("missing id: %v", err)
	}
	_, err = NewPolicySet([]Policy{{ID: "x", Effect: "allow"}})
	if err == nil || !strings.Contains(err.Error(), "permission is required") {
		t.Errorf("missing permission: %v", err)
	}
	_, err = NewPolicySet([]Policy{{ID: "x", Permission: "iam.user.read"}})
	if err == nil || !strings.Contains(err.Error(), "effect is required") {
		t.Errorf("missing effect: %v", err)
	}
	_, err = NewPolicySet([]Policy{{ID: "x", Permission: "iam.user.read", Effect: "deny", MinClearance: "ultra"}})
	if err == nil || !strings.Contains(err.Error(), "unknown min_clearance") {
		t.Errorf("bad clearance: %v", err)
	}
	_, err = NewPolicySet([]Policy{{ID: "x", Permission: "two.parts", Effect: "allow"}})
	if err == nil || !strings.Contains(err.Error(), "valid") {
		t.Errorf("bad permission: %v", err)
	}
}

func TestNewPolicySet_SortsByPriorityDescDenyFirst(t *testing.T) {
	in := []Policy{
		{ID: "a", Permission: "*", Effect: "allow", Priority: 10},
		{ID: "b", Permission: "*", Effect: "deny", Priority: 10},
		{ID: "c", Permission: "*", Effect: "allow", Priority: 100},
	}
	set, err := NewPolicySet(in)
	if err != nil {
		t.Fatalf("set: %v", err)
	}
	rules := set.Rules()
	if rules[0].ID != "c" {
		t.Errorf("priority: %v", rules)
	}
	if rules[1].ID != "b" {
		t.Errorf("deny-first within tie: %v", rules)
	}
}

// ----------------------------------------------------------------------
// LoadPoliciesYAML
// ----------------------------------------------------------------------

func TestLoadPoliciesYAML(t *testing.T) {
	body := []byte(`
policies:
  - id: r1
    description: operators read passes
    effect: allow
    priority: 50
    permission: groundstation.pass.read
    roles: [operator, mission_lead]
  - id: r2
    description: itar read requires US person
    effect: deny
    priority: 200
    permission: '*.pass.read'
    require_us_person: true
    min_clearance: itar
`)
	set, err := LoadPoliciesYAML(body)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if set.Len() != 2 {
		t.Fatalf("expected 2 rules, got %d", set.Len())
	}
}

// ----------------------------------------------------------------------
// Decide — exhaustive truth-table
// ----------------------------------------------------------------------

// principalSpec is a compact constructor for test principals so the
// table rows stay readable.
type principalSpec struct {
	user, tenant, clearance string
	roles                   []string
	usPerson                bool
}

func (p principalSpec) Build() *Principal {
	return &Principal{
		UserID:         p.user,
		TenantID:       p.tenant,
		Roles:          append([]string{}, p.roles...),
		ClearanceLevel: p.clearance,
		IsUSPerson:     p.usPerson,
	}
}

func TestDecide_TruthTable(t *testing.T) {
	rules := []Policy{
		// 1. Operator may read passes.
		{ID: "ops-pass-read", Effect: EffectAllow, Priority: 50,
			Permission: "groundstation.pass.read", Roles: []string{"operator"}},
		// 2. Mission lead can write passes.
		{ID: "ml-pass-write", Effect: EffectAllow, Priority: 60,
			Permission: "groundstation.pass.write", Roles: []string{"mission_lead"},
			MinClearance: "restricted"},
		// 3. ITAR-classified resources: deny unless principal is
		//    a US person at ITAR clearance. Permission pattern
		//    targets only the `itar.*` module so the deny does
		//    not blanket every pass-resource access.
		{ID: "itar-deny", Effect: EffectDeny, Priority: 1000,
			Permission: "itar.*.*", MinClearance: "itar", RequireUSPerson: true},
		// 4. Super-admin: full wildcard.
		{ID: "super", Effect: EffectAllow, Priority: 9999,
			Permission: "*", Roles: []string{"admin"}},
		// 5. Always-deny on a sensitive command.
		{ID: "deny-launch", Effect: EffectDeny, Priority: 5000,
			Permission: "groundstation.command.launch"},
	}
	set, err := NewPolicySet(rules)
	if err != nil {
		t.Fatalf("set: %v", err)
	}

	cases := []struct {
		name       string
		principal  principalSpec
		permission string
		want       Effect
		wantReason string
		wantPolicy string
	}{
		{
			name:       "operator reads pass",
			principal:  principalSpec{user: "u1", roles: []string{"operator"}, clearance: "internal"},
			permission: "groundstation.pass.read",
			want:       EffectAllow, wantPolicy: "ops-pass-read", wantReason: ReasonAllowedByRule,
		},
		{
			name:       "operator forbidden to write pass",
			principal:  principalSpec{user: "u1", roles: []string{"operator"}, clearance: "internal"},
			permission: "groundstation.pass.write",
			want:       EffectDeny, wantReason: ReasonNoMatchingAllow,
		},
		{
			name:       "mission lead with restricted writes pass",
			principal:  principalSpec{user: "u2", roles: []string{"mission_lead"}, clearance: "restricted"},
			permission: "groundstation.pass.write",
			want:       EffectAllow, wantPolicy: "ml-pass-write", wantReason: ReasonAllowedByRule,
		},
		{
			name:       "mission lead lacking restricted clearance is denied",
			principal:  principalSpec{user: "u2", roles: []string{"mission_lead"}, clearance: "internal"},
			permission: "groundstation.pass.write",
			want:       EffectDeny, wantReason: ReasonNoMatchingAllow,
		},
		{
			name:       "non-US-person hitting ITAR resource",
			principal:  principalSpec{user: "u3", roles: []string{"operator", "admin"}, clearance: "itar", usPerson: false},
			permission: "itar.payload.read",
			want:       EffectDeny, wantPolicy: "itar-deny", wantReason: ReasonExplicitDeny,
		},
		{
			name:       "US person below ITAR clearance hits ITAR resource",
			principal:  principalSpec{user: "u3", roles: []string{"operator", "admin"}, clearance: "cui", usPerson: true},
			permission: "itar.payload.read",
			want:       EffectDeny, wantPolicy: "itar-deny", wantReason: ReasonExplicitDeny,
		},
		{
			name:       "US person at ITAR clearance reads ITAR resource",
			principal:  principalSpec{user: "u3", roles: []string{"admin"}, clearance: "itar", usPerson: true},
			permission: "itar.payload.read",
			want:       EffectAllow, wantPolicy: "super", wantReason: ReasonAllowedByRule,
		},
		{
			name:       "admin global wildcard wins",
			principal:  principalSpec{user: "u4", roles: []string{"admin"}, clearance: "cui"},
			permission: "iam.user.delete",
			want:       EffectAllow, wantPolicy: "super", wantReason: ReasonAllowedByRule,
		},
		{
			name:       "deny-launch beats super-admin (deny-wins)",
			principal:  principalSpec{user: "u4", roles: []string{"admin"}, clearance: "cui"},
			permission: "groundstation.command.launch",
			want:       EffectDeny, wantPolicy: "deny-launch", wantReason: ReasonExplicitDeny,
		},
		{
			name:       "no role match → no allow → default deny",
			principal:  principalSpec{user: "u5", roles: []string{"viewer"}, clearance: "internal"},
			permission: "groundstation.pass.read",
			want:       EffectDeny, wantReason: ReasonNoMatchingAllow,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Decide(tc.principal.Build(), Request{Permission: tc.permission}, set)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if got.Effect != tc.want {
				t.Errorf("effect: got %q want %q", got.Effect, tc.want)
			}
			if tc.wantPolicy != "" && got.MatchedPolicyID != tc.wantPolicy {
				t.Errorf("matched policy: got %q want %q", got.MatchedPolicyID, tc.wantPolicy)
			}
			if tc.wantReason != "" && got.Reason != tc.wantReason {
				t.Errorf("reason: got %q want %q", got.Reason, tc.wantReason)
			}
		})
	}
}

func TestDecide_NilPrincipal(t *testing.T) {
	set, _ := NewPolicySet(nil)
	_, err := Decide(nil, Request{Permission: "x.y.z"}, set)
	if !errors.Is(err, ErrNoPrincipal) {
		t.Errorf("got %v want ErrNoPrincipal", err)
	}
}

func TestDecide_EmptyPermission(t *testing.T) {
	set, _ := NewPolicySet(nil)
	_, err := Decide(&Principal{UserID: "u"}, Request{}, set)
	if !errors.Is(err, ErrNoPermission) {
		t.Errorf("got %v want ErrNoPermission", err)
	}
}

func TestDecide_NilPolicySet(t *testing.T) {
	_, err := Decide(&Principal{UserID: "u"}, Request{Permission: "x.y.z"}, nil)
	if !errors.Is(err, ErrNoPolicySet) {
		t.Errorf("got %v want ErrNoPolicySet", err)
	}
}

// REQ-FUNC-PLT-AUTHZ-001 default deny: empty policy set → deny.
func TestDecide_EmptyPolicySetDeniesAll(t *testing.T) {
	set, _ := NewPolicySet(nil)
	got, err := Decide(&Principal{UserID: "u", Roles: []string{"admin"}}, Request{Permission: "x.y.z"}, set)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.Effect != EffectDeny || got.Reason != ReasonNoMatchingAllow {
		t.Errorf("got %+v", got)
	}
}

// Tenant-scoped allow does not leak across tenants.
func TestDecide_TenantScoping(t *testing.T) {
	set, _ := NewPolicySet([]Policy{
		{ID: "t1-only", Effect: EffectAllow, Priority: 10,
			Permission: "iam.user.read", Tenant: "tenant-1"},
	})
	allowed := principalSpec{user: "u", tenant: "tenant-1"}.Build()
	denied := principalSpec{user: "u", tenant: "tenant-2"}.Build()

	got, _ := Decide(allowed, Request{Permission: "iam.user.read"}, set)
	if got.Effect != EffectAllow {
		t.Errorf("tenant-1 should allow: %+v", got)
	}
	got, _ = Decide(denied, Request{Permission: "iam.user.read"}, set)
	if got.Effect != EffectDeny {
		t.Errorf("tenant-2 should deny: %+v", got)
	}
}

// Wildcard tenant matches every principal.
func TestDecide_WildcardTenantMatches(t *testing.T) {
	set, _ := NewPolicySet([]Policy{
		{ID: "any-tenant", Effect: EffectAllow, Priority: 10,
			Permission: "iam.user.read", Tenant: "*"},
	})
	got, _ := Decide(principalSpec{user: "u", tenant: "tenant-x"}.Build(),
		Request{Permission: "iam.user.read"}, set)
	if got.Effect != EffectAllow {
		t.Errorf("wildcard tenant should match: %+v", got)
	}
}
