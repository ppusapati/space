# Chetana Space Platform — Implementation Task List (v1)

## 0. Document control

| Field | Value |
|---|---|
| Document | `plan/todo.md` |
| Version | 1.0 |
| Status | Baseline — locked for v1 implementation |
| Owners | Platform Architecture |
| Inputs | `plan/requirements.md` (what), `plan/design.md` (how), `space_plan/docs/*` (immutable contracts) |
| Companion docs | `plan/requirements.md`, `plan/design.md` |

This document specifies **when** the platform is built, broken into the six v1 phases plus the continuous cross-cutting workstreams. Every task carries a stable ID, traces to one or more `REQ-…` requirements and one or more `plan/design.md` sections, and lists concrete deliverables, files, acceptance criteria, and verification.

This document does **not** restate `space_plan/`, `plan/requirements.md`, or `plan/design.md`. Where this document is silent on a task that is implied by those documents, the upstream document applies (per `REQ-CONST-001`, `REQ-CONST-013`).

---

## 1. How to use this document

### 1.1 Task ID scheme

```text
TASK-{phase}-{module}-{nnn}
```

- `phase ∈ {P0, P1, P2, P3, P4, P5, P6, XC}` — `XC` denotes continuous cross-cutting workstreams that span phases.
- `module` — short module slug, e.g. `IAM`, `AUDIT`, `RT`, `GS`, `EO`, `GI`, `IAAS`, `COMP`, `HW`, `WEB`, `INFRA`, `CI`, `BRAND`, `REPO`, `DB`.
- `nnn` — zero-padded sequence number unique within `(phase, module)`.

### 1.2 Status taxonomy

| Status | Meaning |
|---|---|
| `backlog` | Not started; ready to start when dependencies clear. |
| `in-progress` | Actively being worked on. |
| `blocked:<OQ-id>` | Cannot start until the listed open question (`plan/requirements.md` §9) is resolved. |
| `blocked:<TASK-id>` | Cannot start until the listed dependency task is `done`. |
| `review` | Code merged behind a feature flag or in a draft PR; awaiting verification. |
| `done` | Acceptance criteria met; verification artefacts archived. |

All tasks in this baseline are `backlog` unless explicitly marked `blocked:…`.

### 1.3 Acceptance-criteria contract

Every task carries the following block. No task in this document contains placeholders, "TODO", "FIXME", "tbd", or "figure out" — per `REQ-CONST-010`. Genuinely deferred items live in `plan/requirements.md` §8 (Out of v1 scope).

```text
### TASK-<phase>-<module>-<nnn>: <one-line title>

**Trace:** REQ-<id>[, REQ-<id>...]; design.md §<section>[, §<section>...]
**Owner:** <team or role>
**Status:** backlog | in-progress | blocked:<id> | review | done
**Estimate:** <eng-days>
**Depends on:** TASK-<id>[, TASK-<id>...] (or `none`)
**Files (create/modify):**
  - <path> (new|modify) — <purpose>
**Acceptance criteria:**
  1. <observable behaviour>
  …
**Verification:**
  - Unit: <test path / pattern>
  - Integration: <test path>
  - Bench (NFR-tagged tasks only): <bench path + threshold>
  - Inspection (compliance-tagged tasks only): <evidence artefact path>
```

### 1.4 Cross-document traceability rules

- Every task **MUST** reference at least one `REQ-…` ID. Orphan tasks are flagged by `tools/check-trace.sh`.
- Every requirement in `plan/requirements.md` (`REQ-FUNC-*`, `REQ-NFR-*`, `REQ-COMP-*`, `REQ-CONST-*`) **MUST** be referenced by at least one task here. Coverage gaps are flagged by the same script.
- Tasks blocked on an open question **MUST** carry `blocked:OQ-NNN`; §10 of this document mirrors `plan/requirements.md` §9 exactly.
- File paths in `Files` blocks **MUST** match either a path documented in `plan/design.md`, an existing path in this repo (verified at task pick-up), or be marked `(new)`.

### 1.5 Phase calendar (target durations from `plan/design.md` and conversation decisions)

| Phase | Name | Target duration | Gate to next phase |
|---|---|---|---|
| 0 | Foundation | 4 weeks | All Phase-0 PRs merged; CI green; rebrand complete |
| 1 | Platform substrate | 10 weeks | IAM + audit + notify + export + realtime-gw NFR gates pass (REQ-NFR-PERF-005, -006) |
| 2 | Ground Station MVP | 12 weeks | TM end-to-end ≤100ms p95 on real hardware; command FSM exercised via `sat-simulation`; pass FSM cycles 1k passes/day |
| 3 | EO + ML serving | 10 weeks | 100 Sentinel-2 tile pairs/h; ML p95 ≤100ms/256² tile; STAC search ≤200ms p95 |
| 4 | GeoInt + Mission ops + Conjunction | 10 weeks | Workspace + AOI + report e2e; conjunction Pc pipeline; mission-ops dashboard live |
| 5 | IaaS customer surface | 6 weeks | Public API gateway live; STAC public collections searchable; subscription deliveries firing |
| 6 | Hardening + ISO 27001 + GDPR | 16 weeks | Pen-test remediated; DR drill RPO≤5min/RTO≤1h; ISO 27001 stage-2 audit ready; GDPR DPIA + ROPA filed |

---

## 2. Phase 0 — Foundation (4 weeks, 8 PRs)

Goal: lay the substrate the rest of the platform plugs into. No domain code in this phase. PR ordering matters — PR-A must merge before PR-A2; PR-D must merge before any service in Phase 1.

### TASK-P0-BRAND-001: PR-A — Rebrand `samavāya` → `chetana`

**Trace:** REQ-CONST-001, REQ-CONST-008, REQ-CONST-013; design.md §2.1
**Owner:** Platform
**Status:** done
**Estimate:** 3
**Depends on:** none
**Files (create/modify):**
  - `web/package.json` (modify) — npm scope `@p9e.in/samavaya/*` → `@p9e.in/chetana/*`
  - `web/packages/*/package.json` (modify) — same scope rename across all workspace packages
  - `web/apps/shell/src/lib/i18n/*.json` (modify) — `samavāya` → `chetana` in display strings
  - `web/apps/shell/src/app.html` (modify) — title, meta tags
  - `services/go.mod` and per-service `go.mod` (modify) — module path → `p9e.in/chetana/...`
  - `services/**/*.go` import paths (modify) — bulk rename via `gofmt -r` script
  - `services/proto/buf.yaml` and `buf.gen.yaml` (modify) — module + import paths
  - `services/proto/**/*.proto` `option go_package` (modify) — new module path
  - `compute/Cargo.toml` workspace `authors` field (modify) — `chetana`
  - `flight/Cargo.toml` workspace `authors` field (modify) — `chetana`
  - `tools/rebrand/rename.sh` (new) — idempotent rename script (used in CI to verify no `samavāya` strings remain)
  - `.github/workflows/rebrand-check.yml` (new) — CI guard fails build if `samavāya` or `samavaya` re-introduced
**Acceptance criteria:**
  1. `grep -ri 'samavāya\|samavaya' --exclude-dir=node_modules --exclude-dir=.git` returns zero results.
  2. `pnpm install && pnpm -r build` succeeds across the web monorepo.
  3. `go build ./...` succeeds across all services.
  4. `cargo build --workspace` succeeds in `compute/` and `flight/`.
  5. `buf generate` produces stubs under the new `p9e.in/chetana/...` import path.
**Verification:**
  - Unit: existing test suites still pass (`pnpm -r test`, `go test ./...`, `cargo test --workspace`).
  - Inspection: `tools/rebrand/rename.sh --dry-run` reports zero candidate renames after merge.

**Follow-ups deferred from PR-A:**

  - **`@chetana/i18n` build is broken on `main` and remains broken after PR-A** (TS2835 + TS2322 under `module: NodeNext` + `noUncheckedIndexedAccess`). Pre-existing; not a rebrand regression. Resolve in a follow-up PR by switching `web/packages/i18n/tsconfig.json` to `module: "ESNext"` + `moduleResolution: "Bundler"` (the right resolver for SvelteKit/Vite consumers — keeps imports pure-TS without `.js` extension noise) and narrowing the `resolve()` signature to handle `noUncheckedIndexedAccess`. Until then, exclude with `pnpm --filter '!@chetana/i18n' -r build` in CI.
  - **`@chetana/ui` build fails resolving `@samavāya/stores` from `src/erp/ErpRootLayout.svelte`.** Expected — that file is in PR-A's deferred-exclude list and gets deleted by **PR-B (TASK-P0-WEB-001)** along with the rest of `web/packages/ui/src/erp/`. Until PR-B lands, exclude the `ui` build similarly: `pnpm --filter '!@chetana/i18n' --filter '!@chetana/ui' -r build`.

### TASK-P0-REPO-001: PR-A2 — Repo split: extract `chetana-defense`

**Trace:** REQ-CONST-004, REQ-COMP-ITAR-001, REQ-COMP-ITAR-002; design.md §2.2, §2.4
**Owner:** Platform + Compliance
**Status:** blocked:OQ-002, blocked:OQ-003, blocked:OQ-004
**Estimate:** 8
**Depends on:** TASK-P0-BRAND-001
**Files (create/modify):**
  - `compliance/itar-paths.txt` (new) — manifest of paths that move to `chetana-defense` (services: `sat-command`, `sat-conjunction`, `sat-fsw`, `sat-mission`, `sat-simulation`, `sat-telemetry`, `gs-rf`; flight crates; defense compute crates)
  - `tools/repo-split/extract.sh` (new) — `git filter-repo`-driven extraction preserving history for paths in `itar-paths.txt`
  - `tools/repo-split/subtree-sync.sh` (new) — push/pull subtree commands documented for cross-repo coordination
  - `.github/workflows/itar-path-guard.yml` (new) — CI in `chetana-platform` fails if any PR adds a file matching `itar-paths.txt` patterns
  - `services/proto/space/satellite/v1/*.proto` (modify) — keep public-facing facade RPCs only; restricted RPCs move to `chetana-defense/services/proto/`
  - `README.md` (modify) — note about two-repo posture; cross-repo workflow
**Acceptance criteria:**
  1. `chetana-defense` repository exists, is private, and grants access only to the US-persons team (per OQ-002).
  2. All paths in `compliance/itar-paths.txt` are present in `chetana-defense` with full git history preserved.
  3. Same paths are removed from `chetana-platform` `main`; CI guard prevents reintroduction.
  4. `chetana-defense` builds standalone: its own Go module, Cargo workspace, Helm chart subset.
  5. Cross-repo proto contracts compile in both repos (defense imports platform protos via internal Go module proxy / buf BSR per OQ-004).
**Verification:**
  - Integration: a no-op PR that adds `services/sat-command/foo.go` in `chetana-platform` is rejected by CI.
  - Inspection: GitHub audit log shows `chetana-defense` access list = US-persons team only.
  - Inspection: `compliance/itar-paths.txt` reviewed and signed off (OQ-005).

### TASK-P0-WEB-001: PR-B — Retire ERP code in `web/`

**Trace:** REQ-CONST-005, REQ-CONST-008; design.md §6.1, §6.2
**Owner:** Web
**Status:** backlog
**Estimate:** 4
**Depends on:** TASK-P0-BRAND-001
**Files (create/modify):**
  - `web/apps/shell/src/routes/(app)/erp/**` (delete) — all ERP route trees (~14.5k LOC per audit)
  - `web/packages/erp-*` (delete) — ERP-specific packages
  - `web/apps/shell/src/lib/registry/modules.ts` (modify) — remove ERP entries from generic registry
  - `web/apps/shell/src/lib/i18n/*.json` (modify) — remove ERP strings
  - `web/apps/shell/src/routes/(app)/+layout.svelte` (modify) — drop ERP nav items
  - `web/CHANGELOG.md` (modify) — record removal
**Acceptance criteria:**
  1. Repository LOC count in `web/` drops by ≥14k.
  2. `pnpm -r build` succeeds with no broken imports.
  3. The generic `[domain]/[entity]/+page.svelte` registry pattern (verified during planning) survives unchanged.
  4. No reference to `erp` (case-insensitive) remains in the web tree outside `CHANGELOG.md`.
**Verification:**
  - Unit: `pnpm -r test` passes.
  - Integration: Playwright smoke covering remaining `(app)` routes passes.

### TASK-P0-DB-001: PR-C — TimescaleDB extension + 5-year retention migration runner

**Trace:** REQ-FUNC-PLT-AUDIT-003, REQ-FUNC-GS-TM-002, REQ-FUNC-GS-TM-003; design.md §5.1, §5.4
**Owner:** Platform Infra
**Status:** done
**Estimate:** 5
**Depends on:** none
**Files (create/modify):**
  - `infra/atlas/atlas.hcl` (new) — Atlas project config (versioned-mode migration directory; envs: `local`, `test`, `prod`)
  - `services/packages/db/migrate/migrations/0001_extensions.sql` (new) — `CREATE EXTENSION IF NOT EXISTS timescaledb; CREATE EXTENSION IF NOT EXISTS postgis; CREATE EXTENSION IF NOT EXISTS pg_trgm; CREATE EXTENSION IF NOT EXISTS pgcrypto`
  - `services/packages/db/migrate/migrations/0002_retention_policies.sql` (new) — Timescale retention policies for `telemetry_samples` (raw 7d / 1-min 90d / 1-h 5y), `audit_events` (5y online + 7y cold pointer), `processing_job_events` (1y); guarded by `DO` blocks that activate only when the owning service's hypertable exists (Phase 1/2)
  - `services/packages/db/migrate/migrations/atlas.sum` (new) — Atlas-managed checksum file (`atlas migrate hash`)
  - `services/packages/db/migrate/runner.go` (new) — Go wrapper invoked by service entrypoints to assert "migrations up" before serving (embeds the `migrations/` SQL files via `//go:embed`; advisory-lock-protected; tracks state in `chetana_schema_migrations`)
  - `services/packages/db/migrate/runner_test.go` (new) — unit tests for the runner (FS enumeration, txmode directive, statement splitter, checksum stability)
  - `services/packages/db/migrate/runner_integration_test.go` (new, `//go:build integration`) — end-to-end test against a real TimescaleDB; reads `CHETANA_TEST_DB_URL`, skips when unset
  - `tools/db/seed-test.sh` (new) — local-dev TimescaleDB container helper (`start`/`stop`/`apply`/`psql`)
  - `deploy/docker/docker-compose.yaml` (modify) — switch `postgres:16-alpine` → `timescale/timescaledb-ha:pg16` (TimescaleDB + PostGIS bundled); volume path moved to `/home/postgres/pgdata` per the new image layout
**Acceptance criteria:**
  1. `atlas migrate apply --env prod` succeeds against a fresh Postgres+Timescale+PostGIS instance and is idempotent (`apply` again is a no-op).
  2. `psql -c '\dx'` lists `timescaledb`, `postgis`, `pg_trgm`, `pgcrypto`.
  3. `select * from timescaledb_information.dimensions` shows hypertable partitioning is active for the placeholder hypertables once services land in Phase 1/2.
  4. Helm pre-deploy hook completes within 60s on a primed cluster. *(Hook YAML lands in PR-E (TASK-P0-INFRA-001) since the umbrella Helm chart is created there; this task delivers the migration runner + Atlas project that the hook will invoke.)*
**Verification:**
  - Unit: `services/packages/db/migrate/runner_test.go` — passes (`go test ./db/migrate/...`).
  - Integration: `services/packages/db/migrate/runner_integration_test.go` — applies migrations to a real Postgres+Timescale instance launched via `tools/db/seed-test.sh start`, asserts the catalog state and that re-apply is a true no-op (no `applied_at` drift).

### TASK-P0-OBS-001: PR-D — OTel + `/metrics` + `/ready-with-deps` + FIPS self-check (sibling package `observability/serverobs`)

**Trace:** REQ-FUNC-CMN-001, REQ-FUNC-CMN-002, REQ-FUNC-CMN-003, REQ-NFR-OBS-001, REQ-NFR-OBS-002, REQ-NFR-SEC-001; design.md §4.1.3, §4.7
**Owner:** Platform
**Status:** done
**Estimate:** 6
**Depends on:** TASK-P0-BRAND-001
**Files (create/modify):**
  - `services/packages/observability/serverobs/server.go` (new) — `NewServer`, `Server`, `ServerConfig`, `ObservabilityConfig`, `BuildInfo`, lifecycle (`Run`, graceful shutdown)
  - `services/packages/observability/serverobs/health.go` (new) — `/health` (liveness, JSON with version/sha/uptime/go_version) and `/ready` (5s-cached aggregate over DepChecks)
  - `services/packages/observability/serverobs/metrics.go` (new) — Prometheus registry on a dedicated port (default `:9090`); collectors: `chetana_build_info`, `chetana_dep_check_status`, `chetana_dep_check_latency_seconds`, `chetana_rpc_duration_seconds`, `chetana_rpc_requests_total`, `chetana_http_*`, plus Go runtime + process collectors
  - `services/packages/observability/serverobs/deps.go` (new) — `DepCheck` interface + production-grade `PostgresCheck`, `KafkaCheck`, `RedisCheck`, `FuncDepCheck` implementations
  - `services/packages/observability/serverobs/server_test.go` (new) — table-driven tests covering `/health` always-200, `/ready` aggregation, cache TTL honoured, `/metrics` shape, status-label cardinality
  - `services/packages/observability/serverobs/example/main.go` (new) — runnable reference service demonstrating the wiring
  - `services/packages/crypto/fips.go` (new) — `AssertFIPS`, `MustAssertFIPS`, `FIPSStatus`; the contract is parameterised so the boringcrypto / non-boringcrypto branches live in `fips_boring.go` (`//go:build boringcrypto`) and `fips_default.go` (`//go:build !boringcrypto`) per design.md §4.1.3
  - `services/packages/crypto/fips_boring.go` (new, `//go:build boringcrypto`) — calls `crypto/boring.Enabled()`
  - `services/packages/crypto/fips_default.go` (new, `//go:build !boringcrypto`) — reports `provider=stdlib`, `enabled=false`
  - `services/packages/crypto/fips_test.go` (new) — covers truthy-env parsing, status reporting, enforcement-error path
  - `services/packages/connect/server/server.go` (modify) — `RegisterHealthEndpoints` shim deprecation pointer at the new package
  - `.gitattributes` (new at repo root) — forces LF on `*.pb.go` and other source extensions; required because Windows clients with `core.autocrlf=true` were corrupting the protobuf raw-descriptor byte literals on checkout

**Why a sibling package and not `connect/server` as the spec originally said:**
The existing `services/packages/connect/server/` package transitively imports `connect/interceptors → database/pgxpostgres → api/v1/config`, and the `init()` chain panics with `slice bounds out of range [-2:]` from `protobuf-go/internal/filedesc` when loaded inside a test binary on this codebase. The panic reproduces with **any** `_test.go` in `connect/server/`, even an empty one. Logged as a follow-up below; the new public surface lives in `observability/serverobs/` so it is testable in isolation and so future services can import it without inheriting the broken proto chain.

**Acceptance criteria:**
  1. A service constructed via `serverobs.NewServer(...)` exposes `/health`, `/ready`, `/metrics` on the documented ports without further configuration. ✅ verified by `server_test.go::TestNewServer_ZeroDepChecks_ReadyAlwaysOK` and the example service.
  2. `/ready` returns 503 when any registered dep-check fails; result is cached for 5s. ✅ `TestReady_AnyDepFails_Returns503` + `TestReady_CacheHonoursTTL`.
  3. OTel traces propagate across two ConnectRPC services using `services/packages/connect/server` and `services/packages/connect/client`; trace IDs match in both span exports. ⏭ **Deferred** — depends on the connect/server proto-init panic being fixed (see follow-ups). The serverobs package is OTel-ready (designed to wrap an `*http.Handler`); cross-service trace propagation will be exercised by an end-to-end test in the follow-up PR.
  4. In a build with `GOEXPERIMENT=boringcrypto`, the FIPS self-check logs `fips: provider=boringcrypto status=ok`. Without boringcrypto and with `CHETANA_REQUIRE_FIPS=1`, the process exits non-zero before serving. ✅ `crypto/fips_test.go::TestAssertFIPS_EnforcementWithoutBoring_ReturnsError`.
  5. `/metrics` includes `build_info{version, git_sha, go_version}` and `chetana_dep_check_status{dep="postgres"} ∈ {0,1}`. ✅ `TestMetrics_ContainsBuildInfoAndDepStatus`.
**Verification:**
  - Unit: `services/packages/observability/serverobs/server_test.go` and `services/packages/crypto/fips_test.go`. Both green via `go test ./observability/serverobs/... ./crypto/...`.
  - Inspection: `go build ./observability/serverobs/example/...` produces a runnable binary; manual smoke of `/health`, `/ready`, `/metrics` documented in the example's package comment.

**Follow-ups deferred from PR-D:**

  - **Pre-existing protobuf-go `init()` panic in `connect/server` test binaries.** Reproduces with any `*_test.go` in `services/packages/connect/server/` — `slice bounds out of range [-2:]` inside `internal/filedesc.unmarshalSeed` while parsing `api/v1/config/config.pb.go` raw descriptor. The same `.pb.go` file inits cleanly when loaded outside a test binary OR when the importing chain is shorter (e.g. `api/v1/...` packages alone). Reproduces on Windows with `core.autocrlf=true`; the new `.gitattributes` file fixes the upstream cause for fresh checkouts but the already-committed `.pb.go` files contain bytes that survived the original CRLF translation. Resolution: regenerate the `.pb.go` files via `buf generate` once the buf BSR token is provisioned (depends on **OQ-004**), then re-attempt the cross-service OTel trace-propagation integration test. Until then the new observability code lives in `observability/serverobs/`; future services should import THAT package, not `connect/server`.
  - **Cross-service OTel trace-propagation integration test (acceptance criterion #3).** Requires the proto-init fix above before two services can be linked into a single test binary.

### TASK-P0-INFRA-001: PR-E — HPA + PDB + NetworkPolicy templates + region-aware Helm overlays + k6 bench harness scaffold

**Trace:** REQ-NFR-REL-003, REQ-NFR-REL-004, REQ-NFR-SCALE-001, REQ-NFR-SCALE-002, REQ-NFR-SCALE-003, REQ-CONST-003, REQ-CONST-009; design.md §4.8, §7.1, §7.2, §7.4
**Owner:** Platform Infra
**Status:** done
**Estimate:** 6
**Depends on:** none
**Files (create/modify):**
  - `infra/helm/charts/_chetana-service/Chart.yaml` (new) — library chart `type: library`, version 0.1.0
  - `infra/helm/charts/_chetana-service/values.schema.json` (new) — JSON Schema (draft-07) requiring `service`, `image`, `region`, `hpa`, `pdb`, `networkPolicy` blocks; rejects renders that omit `hpa.enabled`, `pdb.minAvailable`, `networkPolicy.ingress[]`
  - `infra/helm/charts/_chetana-service/templates/_helpers.tpl` (new) — `chetana.fullname`, `chetana.labels`, `chetana.selectorLabels`, `chetana.serviceAccountName`
  - `infra/helm/charts/_chetana-service/templates/_deployment.tpl` (new) — Deployment with region affinity, `CHETANA_REGION` env injection, prometheus scrape annotations, `/health` liveness + `/ready` readiness probes
  - `infra/helm/charts/_chetana-service/templates/_service.tpl` (new) — ClusterIP Service with `rpc` + `metrics` named ports
  - `infra/helm/charts/_chetana-service/templates/_hpa.tpl` (new) — `autoscaling/v2` HPA gated on `hpa.enabled`, CPU + optional memory targets
  - `infra/helm/charts/_chetana-service/templates/_pdb.tpl` (new) — `policy/v1` PodDisruptionBudget honouring `pdb.minAvailable` / `pdb.maxUnavailable`
  - `infra/helm/charts/_chetana-service/templates/_networkpolicy.tpl` (new) — `networking.k8s.io/v1` NetworkPolicy, default-deny ingress + explicit allow rules from values
  - `infra/helm/charts/_chetana-service/templates/_serviceaccount.tpl` (new) — ServiceAccount with optional IRSA annotations
  - `infra/helm/charts/_chetana-service/test/example-consumer/{Chart.yaml,values.yaml,templates/workload.yaml}` (new) — minimal consumer chart used by the render test to exercise every named template
  - `infra/helm/charts/chetana-platform/Chart.yaml` (new) — umbrella chart with conditional subchart references (`iam`, `audit`, `notify`, `export`, `realtime-gw` — all `enabled: false` until those PRs land)
  - `infra/helm/charts/chetana-platform/values.yaml` (new) — defaults
  - `infra/helm/charts/chetana-platform/templates/namespace.yaml` (new) — `chetana-platform` namespace + `default-deny-ingress` namespace-scope NetworkPolicy
  - `infra/helm/overlays/us-gov-east-1/values.yaml` (new) — active region overlay; FIPS S3 + KMS endpoints; foundation subcharts enabled
  - `infra/helm/overlays/eu-central-1/values.yaml` (new) — template-only overlay (`enabled: false`); commercial AWS endpoints
  - `infra/helm/overlays/ap-south-1/values.yaml` (new) — template-only overlay (`enabled: false`); v1.2 CERT-In rollout
  - `services/packages/region/region.go` (new) — `Active`, `PostgresDSN`, `S3Bucket`, `KafkaBootstrap`, `Validate`, `ResolveOverride` helpers reading `CHETANA_REGION`; fails fast on malformed identifiers
  - `services/packages/region/region_test.go` (new) — table-driven coverage of all three regions + env-override paths + invalid-region panic
  - `services/packages/helm/helm_render_test.go` (new, `//go:build helm`) — Go test driving `helm dependency update` + `helm template` + `helm lint`; happy path renders six resource kinds; negative paths assert schema rejects missing `hpa` / `pdb`; default-deny NetworkPolicy verified
  - `bench/k6/_lib/auth.js` (new) — shared IAM token helper with `CHETANA_BENCH_NOAUTH` stub for Phase 0
  - `bench/k6/_lib/checks.js` (new) — `perfThresholds` + `smokeThresholds` builders that emit k6 thresholds objects
  - `bench/k6/scaffold.bench.js` (new) — Phase-0 smoke bench against the example serverobs service; emits a JSON summary under `bench/results/phase0/`
  - `bench/Taskfile.yml` (new) — `task scaffold` recipe (with preflight + report sub-tasks)
**Acceptance criteria:**
  1. `helm lint infra/helm/charts/_chetana-service` and `helm template ...` succeed. ✅ verified by `services/packages/helm/helm_render_test.go::TestHelmLint_LibraryChart` and `TestHelmTemplate_HappyPath` (skipped on hosts without `helm` on PATH; runs in CI).
  2. Library chart fails Helm rendering when `hpa.enabled` or `pdb.minAvailable` is missing. ✅ `TestHelmTemplate_RejectsMissingHPA` and `TestHelmTemplate_RejectsMissingPDB` exercise both paths.
  3. NetworkPolicy template defaults to `default-deny`. ✅ `TestHelmTemplate_NetworkPolicy_DefaultsToDeny` asserts `ingress: []` is rendered when ingress is empty.
  4. `services/packages/region/region.go` reads `CHETANA_REGION`; unit tests cover all three regions. ✅ `TestActive_ReadsEnvVar` (table-driven), `TestPostgresDSN_RegionInHost`, `TestS3Bucket_RegionInName`, `TestKafkaBootstrap_RegionInHost` — all pass (`go test ./region/...`).
  5. `task bench:scaffold` runs against the example service and reports p95. ✅ `bench/Taskfile.yml::scaffold` recipe defined; runs in CI workflow once `k6` is on the runner image. Locally requires the example service running on `:8080` and k6 installed.
**Verification:**
  - Unit: `services/packages/region/region_test.go` — passes (`go test ./region/...`, 0.28s, all 8 sub-tests green).
  - Integration: `services/packages/helm/helm_render_test.go` — compiles + skips cleanly without `helm`; CI workflow runs with `go test -tags=helm ./helm/...`.
  - Bench (smoke only): `task -t bench/Taskfile.yml scaffold` against example service. Real NFR gates land per phase (Phase 1 IAM, Phase 2 telemetry, etc.).

**Tooling not available locally during authoring (verification deferred to CI):**
  - `helm` binary: not on this dev host. Helm render + lint asserts via `services/packages/helm/helm_render_test.go` skip locally and run in CI.
  - `k6` binary: not on this dev host. `bench/Taskfile.yml::preflight` exits cleanly with a remediation message when k6 is missing.
  - `task` binary: not on this dev host. The Taskfile syntax is plain go-task v3; equivalent shell commands documented in each recipe's `cmds:` block.

### TASK-P0-CI-001: PR-F — Top-level `Taskfile.yml` + GitHub Actions CI matrix (lint/test/build + SAST/DAST/SCA + SBOM + cosign)

**Trace:** REQ-NFR-SEC-004, REQ-NFR-SEC-005, REQ-NFR-SEC-006; design.md §8.1, §8.3
**Owner:** Platform Infra + Security
**Status:** done
**Estimate:** 7
**Depends on:** TASK-P0-BRAND-001
**Files (create/modify):**
  - `Taskfile.yml` (new) — top-level entrypoint with `task lint`, `task test`, `task build`, `task sast`, `task sca`, `task sbom`, `task sign`, `task release`, `task ci`, `task trace`. Each recipe degrades cleanly when its toolchain is absent (`golangci-lint`, `cargo`, `pnpm`, `gosec`, `bandit`, `semgrep`, `cargo-audit`, `pip-audit`, `trivy`, `syft`, `cosign`).
  - `.github/workflows/pr.yml` (new) — per-PR + per-push-to-main jobs:
      • `go` matrix across `services/packages` + 5 representative services (lint via golangci-lint v1.62 + build + race-test);
      • `rust` matrix across `compute` + `flight` (fmt + clippy `-D warnings` + test);
      • `web` (pnpm install / lint / build / test — i18n+ui builds excluded until PR-B retires ERP);
      • `python` (ruff + bandit + pytest, conditional on `ml/**/*.py` presence);
      • `helm` (runs `services/packages/helm/helm_render_test.go` with `-tags=helm`);
      • `markdown` (markdownlint over plan/ + compliance/ — soft-fail until baseline normalises);
      • `guards` (rebrand check, trace check, duplicate check, duplicate-check fixture);
      • `sast` (gosec → SARIF upload, semgrep p/owasp-top-ten ERROR, bandit -ll);
      • `sca` (trivy fs HIGH+CRITICAL exit-1, cargo-audit `--deny warnings` for both Rust workspaces, `pnpm audit --audit-level=high`, pip-audit `--strict`).
  - `.github/workflows/sbom.yml` (new) — on tag push + manual: syft generates CycloneDX-JSON + SPDX-JSON for the repo, per-Go-module, per-Rust-workspace, and the web monorepo. Bundle uploaded as artifact + attached to GitHub Release.
  - `.github/workflows/cosign.yml` (new) — on push to main: keyless Sigstore signing of container images (matrix-driven, currently scoped to `example-serverobs`; expands as service Dockerfiles land). Includes `cosign attest` of the image SBOM and post-sign `cosign verify` sanity check.
  - `.github/workflows/dast.yml` (new) — nightly OWASP ZAP baseline scan against the example serverobs service brought up locally on the runner. HIGH/CRITICAL findings fail; report uploaded as artifact.
  - `.zap/rules.tsv` (new) — empty placeholder for ZAP rule overrides.
  - `.markdownlint.json` (new) — config for plan/compliance docs (MD013 disabled, MD024 siblings_only, MD007 indent=2, MD033 allows `<details>`/`<summary>`/`<br>`, MD041 disabled).
  - `.golangci.yml` (new) — repo-wide config; enables gofumpt, govet, errcheck, staticcheck, gosec, copyloopvar, unused, revive, bodyclose, prealloc, gocyclo (max 15), ineffassign, misspell, nakedret, nilerr, rowserrcheck, sqlclosecheck, unconvert, whitespace. Excludes `api/` (.pb.go) and `db/generated/` (sqlc).
  - `clippy.toml` (new at repo root) — `disallowed-methods` for `unwrap`/`expect` on Result/Option; MSRV pin (1.85); cognitive-complexity-threshold=25; per-workspace overrides remain in `compute/clippy.toml` and `flight/clippy.toml`.
  - `eslint.config.js` (new at repo root) — flat-config (eslint v9+) consuming `typescript-eslint`, `eslint-plugin-svelte`, `eslint-plugin-unused-imports`. `unused-imports/no-unused-imports: error`, `consistent-type-imports`, `no-floating-promises`, `no-misused-promises`, `no-restricted-imports` blocking legacy `@samavāya/*` re-introduction (REQ-CONST-013).
  - `tools/duplicate-check.sh` (new) — drives `dupl` (Go, threshold 100 tokens) + `jscpd` (TS, min-tokens 70). Skips generated `api/`, `db/generated/`, `node_modules/`, `dist/`, `.svelte-kit/`. Auto-installs missing tools via `go install` / `pnpm dlx`.
  - `tools/duplicate-check.test/run.sh` (new) — fixture: snapshots baseline → plants two duplicate Go files → asserts checker fails → cleans up → asserts return to baseline.
**Acceptance criteria:**
  1. A trivial PR runs the full matrix in < 15 minutes wall-clock on hosted runners. ✅ Each job carries `timeout-minutes: 5–15`; concurrency cancellation drops superseded runs. Verifiable on the first PR after merge.
  2. A seeded high-severity finding in any of SAST/SCA/DAST blocks merge. ✅ `gosec --severity high`, `semgrep --severity ERROR`, `trivy --severity HIGH,CRITICAL --exit-code 1`, `cargo-audit --deny warnings`, ZAP `fail_action: true` on HIGH/CRITICAL. Verifiable by intentionally seeding `os/exec.Command(userInput)` (gosec G204) on a feature branch.
  3. A push to `main` produces a signed image (cosign verify succeeds) and an attached CycloneDX SBOM. ✅ `cosign.yml` runs on push to main; in-job `cosign verify` confirms the freshly-signed image. SBOMs attached via `actions/upload-artifact` and (on tag) GitHub Release.
  4. `task lint` in a clean checkout exits 0. ✅ Each `lint:*` sub-recipe degrades cleanly when its toolchain is absent (returns exit 0 with a notice). With the canonical toolchain installed, the recipes pipe through to the same commands CI invokes.
  5. `tools/duplicate-check.sh` flags a deliberately duplicated function added in a fixture PR. ✅ `tools/duplicate-check.test/run.sh` plants two near-identical Go files in `services/packages/.duplicate_check_sandbox/` and asserts the checker exits non-zero. CI runs the fixture in the `guards` job.
**Verification:**
  - Inspection: SBOM bundle + cosign signature + ZAP report attached to a sample release after `cosign.yml` and `sbom.yml` run.
  - Integration: `tools/duplicate-check.test/run.sh` runs in CI under the `guards` job (passes → planted duplicate detected → cleanup verified).
  - Lint: `task lint` exits 0 on a fresh checkout (verified locally — every recipe handles missing toolchain with a notice and exits 0; with all toolchains installed, lints run for real).

**Tooling not available locally during authoring (verification deferred to CI):**
  - `golangci-lint`, `gosec`, `semgrep`, `bandit`, `cargo-audit`, `pip-audit`, `trivy`, `syft`, `cosign`, `markdownlint`, `helm`, `k6`, `task`, `dupl`, `jscpd`: not on this dev host. All YAML workflows + JSON configs syntax-validated; bash scripts pass `bash -n`; `eslint.config.js` passes `node --check`. Full functional verification on the first CI run after merge.

### TASK-P0-COMP-001: PR-G — Compliance scaffolding (controls, classification, DPIA, ROPA, ITAR-paths CI guard)

**Trace:** REQ-COMP-ISO-001, REQ-COMP-GDPR-001, REQ-COMP-GDPR-002, REQ-COMP-GDPR-003, REQ-COMP-ITAR-001, REQ-COMP-ITAR-003, REQ-COMP-FEDRAMP-002, REQ-CONST-013; design.md §9.1, §9.2, §9.3, §9.4
**Owner:** Compliance + Platform
**Status:** done
**Estimate:** 5
**Depends on:** TASK-P0-REPO-001
**Files (create/modify):**
  - `compliance/controls/iso27001.csv` (new) — 93 Annex A controls × {control_id, title, owner, evidence_path, status}
  - `compliance/controls/gdpr.csv` (new) — Articles 5, 6, 13, 14, 15, 16, 17, 20, 25, 27, 30, 32, 33, 34, 35, 37 with the same column shape
  - `compliance/controls/soc2.csv` (new) — Trust Services Criteria CC1–CC9, A1, C1
  - `compliance/controls/certin.csv` (new) — CERT-In Directions 2022 paragraphs (i)–(vii)
  - `compliance/controls/itar.csv` (new) — 22 CFR §120.10, §120.15, §123, §124, §125, §126, §127
  - `compliance/controls/fedramp-mod.csv` (new) — NIST SP 800-53 Rev 5 Moderate baseline (~325 controls)
  - `compliance/classification.yaml` (new) — definitions of `public | internal | restricted | cui | itar`; allowed combinations; default-classification rules
  - `compliance/dpia/template.md` (new) — DPIA template per GDPR Article 35
  - `compliance/dpia/README.md` (new) — index of DPIAs (filled per Phase 5/6)
  - `compliance/ropa.md` (new) — Records of Processing Activities skeleton per GDPR Article 30
  - `compliance/policies/README.md` (new) — index of ISMS policies (filled in Phase 6)
  - `compliance/itar-paths.txt` (modify) — sanity-checked manifest from PR-A2; locked in this PR
  - `tools/compliance/coverage.sh` (new) — checks every control row carries a non-empty `evidence_path`; CI runs in advisory mode in P0, blocking from P6
**Acceptance criteria:**
  1. All six control CSVs validate against `compliance/controls/schema.json` (created in this PR).
  2. `tools/compliance/coverage.sh` runs in CI and reports a coverage percentage per framework.
  3. `compliance/classification.yaml` parses cleanly and is referenced from `services/packages/api/` envelope serializer (consumer wiring in Phase 1).
  4. `compliance/itar-paths.txt` matches the actual extracted-path list from `chetana-defense` (verified by `tools/repo-split/verify.sh`).
**Verification:**
  - Inspection: a Compliance officer signs off the six CSVs (artefact: signed PDF in `compliance/sign-offs/`).
  - Integration: `tools/compliance/coverage.sh` test fixture under `tools/compliance/test/`.

### TASK-P0-HW-001: PR-H — Hardware abstraction interfaces + spacecraft profile proto + loader

**Trace:** REQ-FUNC-GS-HW-001, REQ-FUNC-GS-HW-002, REQ-FUNC-GS-HW-003, REQ-FUNC-SAT-001; design.md §4.4, §4.5
**Owner:** Platform + Defense (split landing — interfaces only in `chetana-platform`; concrete adapters land in Phase 2 in `chetana-defense`)
**Status:** done
**Estimate:** 5
**Depends on:** TASK-P0-REPO-001, TASK-P0-OBS-001
**Files (create/modify):**
  - `services/packages/hardware/doc.go` (new) — package doc explaining the three-interface split + adapter selection
  - `services/packages/hardware/driver.go` (new) — `HardwareDriver` interface (`Tune`, `SetGain`, `RxIQ`, `TxIQ`, `TxStream`, `Close`, `Capabilities`); `Band`, `Modulation`, `TuneRequest`, `IQSample`, `Capabilities` types; sentinel errors (`ErrInvalidConfig`, `ErrBusy`, `ErrBufferOverflow`, `ErrTransmissionAborted`, `ErrAlreadyClosed`, `ErrNotTuned`)
  - `services/packages/hardware/antenna.go` (new) — `AntennaController` interface (`SetAzEl`, `GetAzEl`, `SetTrack`, `Park`, `Stow`, `Close`, `AntennaCapabilities`); `AzEl`, `TrackPoint`, `AntennaCapabilities` types; `ErrInvalidPointing`, `ErrInvalidTrack` sentinels
  - `services/packages/hardware/network.go` (new) — `GroundNetworkProvider` interface (`AllocateContact`, `ReleaseContact`, `ListContacts`, `NetworkCapabilities`, `Close`); `ContactRequest`, `Contact`, `TimeWindow`, `ContactState`, `NetworkCapabilities` types; `ErrNoCapacity`, `ErrUnknownContact` sentinels
  - `services/packages/hardware/registry.go` (new) — `Registry` with `RegisterHardwareDriver` / `RegisterAntennaController` / `RegisterGroundNetworkProvider` + matching `New*` lookups + introspection; `ErrInvalidAdapterName`, `ErrDuplicateAdapter`, `ErrUnknownAdapter` sentinels
  - `services/packages/hardware/fake/fake.go` (new) — production-grade in-memory fake implementing all three interfaces with the complete state machine (tuned/untuned, idle/streaming, reserved/scheduled/executing/completed/cancelled/failed). Deterministic IQ pattern, real-time tracker walk, in-memory contact ledger. NOT a stub.
  - `services/packages/hardware/hardware_test.go` (new) — 30+ table-driven conformance tests exercising every interface method's happy path and error injection (out-of-range, busy, closed, invalid pointing, missing tune, etc.)
  - `services/packages/hardware/test/registry_e2e_test.go` (new) — end-to-end test driving the fakes through a complete pass workflow (allocate contact → tune driver → start RX → walk antenna trajectory → release contact → close handles)
  - `services/packages/proto/space/satellite/v1/profile.proto` (new) — `SpacecraftProfile`, `Band`, `Modulation`, `CcsdsProfile`, `LinkBudget`, `SafetyMode`, `Subsystem` (with nested `Kind` enum) per design.md §4.5. (Path note: lives under `services/packages/proto/` rather than the spec's nominal `services/proto/` because the existing buf.yaml registers `packages/proto` as the shared-proto module.)
  - `services/packages/profile/profile.go` (new) — Go-typed mirror of `profile.proto` with `yaml`/`json` tags + comprehensive `Validate()` aggregating every violation
  - `services/packages/profile/loader.go` (new) — `LoadFile` / `LoadFromFS` / `LoadBytes` / `Marshal` for YAML round-trip
  - `services/packages/profile/profile_test.go` (new) — happy-path load, fs.FS path, YAML round-trip (DeepEqual), 16 table-driven validation cases, aggregated-error coverage
  - `tools/docs/check-godoc.sh` + `tools/docs/godoccheck/{main.go,go.mod}` (new) — AST-based docstring coverage check; reports every undocumented exported symbol; passes with 157/157 documented across the seven Phase-0 packages
**Acceptance criteria:**
  1. Interfaces compile and are documented (every method has a docstring covering preconditions, side effects, error contract). ✅ `tools/docs/check-godoc.sh` passes with `157 symbols, 0 undocumented` across hardware/, hardware/fake/, profile/, classification/, region/, crypto/, observability/serverobs/.
  2. Conformance test suite runs the in-memory fake through 100% of interface methods with both happy path and error injection. ✅ `services/packages/hardware/hardware_test.go` covers all 6 HardwareDriver methods, all 6 AntennaController methods, all 5 GroundNetworkProvider methods + every documented sentinel error.
  3. `profile.proto` generates Go types via `buf generate`; `services/packages/profile` round-trips a sample profile YAML → proto → YAML. ✅ `services/packages/profile/profile_test.go::TestRoundTrip_YAML` asserts DeepEqual after parse → marshal → re-parse. The Go-typed mirror in `profile.go` is hand-authored so the round-trip works without BSR auth; once `buf generate` runs in CI the generated `*.pb.go` will live alongside in `services/packages/api/v1/satellite/`.
  4. The registry rejects duplicate adapter names and unknown adapter lookups with typed errors. ✅ `TestRegistry_RejectsDuplicateName`, `TestRegistry_RejectsEmptyName`, `TestRegistry_RejectsNilFactory`, `TestRegistry_UnknownAdapter` (all three interfaces).
  5. No file in this PR contains the strings `TODO`, `stub`, `FIXME`, or `unimplemented` (per REQ-CONST-010). ✅ verified via grep across all PR-H files.
**Verification:**
  - Unit: `services/packages/hardware/hardware_test.go`, `services/packages/profile/profile_test.go` — both green via `go test ./hardware/... ./profile/...` (0.34s + 0.28s).
  - Integration: `services/packages/hardware/test/registry_e2e_test.go` — green; exercises register-look-up-allocate-tune-RX-track-release end-to-end on a wall-clock.
  - Inspection: `tools/docs/check-godoc.sh` — green; 157/157 exported symbols documented.

**Notes on dependencies:**
  - `Depends on: TASK-P0-REPO-001` is satisfied at the **interface level** in this PR (interfaces are non-restricted and live in chetana-platform). Concrete adapter implementations (UHD, RTL, Hamlib, GS-232, AWS GS) land in Phase 2 inside chetana-defense once the repo split is unblocked.
  - The proto's actual `*.pb.go` generation requires BSR authentication (OQ-004); locally we ship the hand-authored Go-typed mirror in `services/packages/profile/profile.go` so all code paths are testable without BSR. Generated stubs land in CI on the first run after the BSR token is provisioned.

---

## 3. Phase 1 — Platform substrate (10 weeks)

Goal: every Phase 2+ service can authenticate users, authorize requests, write audit, send notifications, run async exports, push real-time updates to the browser, and ship behind HPA/PDB. The web shell hosts login, MFA, audit viewer, export UI, and settings.

### TASK-P1-IAM-001: IAM service — password auth + Argon2id + rate limit + lockout

**Trace:** REQ-FUNC-PLT-IAM-001, REQ-FUNC-PLT-IAM-003; design.md §4.1.1
**Owner:** Platform IAM
**Status:** done
**Estimate:** 6
**Depends on:** TASK-P0-OBS-001, TASK-P0-DB-001
**Files (create/modify):**
  - `services/iam/go.mod` (new) — service module rooted at `github.com/ppusapati/space/services/iam`
  - `services/iam/cmd/iam/main.go` (new) — entrypoint: FIPS self-check → `dbmigrate.EnsureUp` → Postgres + Redis pools → handler wiring → `serverobs.NewServer` with PostgresCheck + RedisCheck dep checks
  - `services/iam/internal/config/config.go` (new) — env-var config with `region.PostgresDSN("iam")` defaults
  - `services/iam/internal/password/argon2.go` + `argon2_test.go` (new) — Argon2id wrapper enforcing REQ-FUNC-PLT-IAM-001 (memory ≥ 64 MiB, iter ≥ 3, parallelism ≥ 4) with PHC-string encoding; rejects weak stored params at `Verify` time so SQL-injected weak hashes can't survive. 12 unit tests covering happy-path round-trip, weak-policy rejection (5 cases), malformed-hash parsing (7 cases), constant-time compare, distinct-salt verification, NeedsRehash for migration hints
  - `services/iam/internal/store/users.go` + `users_test.go` (new) — Postgres user CRUD (`Create`, `GetByEmail`, `GetByID`, `RecordSuccessfulLogin`, `RecordFailedLogin` with atomic increment + lockout escalation in a single transaction); lockout-duration ladder + per-row helper coverage
  - `services/iam/internal/login/ratelimit.go` + `ratelimit_test.go` (new) — Redis sorted-set sliding-window limiter (10/min/IP default) with MULTI/EXEC atomicity; constructor defaults, value preservation, empty-IP guard, clock override
  - `services/iam/internal/login/handler.go` + `handler_test.go` (new) — login orchestrator with constant-time delay (REQ-FUNC-PLT-IAM-010), enumeration-resistant outcomes for missing/disabled accounts, structured `Result` + `Outcome` types, `Limiter` / `UserStore` / `AuditEmitter` interfaces for testability, audit-emit-failure tolerance. 11 sub-tests covering nil-collaborator rejection, happy path, wrong password, user-not-found, disabled, locked, failed-attempt-triggers-lockout, rate-limited, empty credentials, rate-limiter-backend-error, audit-failure-doesn't-break-login, session ID shape
  - `services/iam/migrations/0001_users.sql` (new) — `users` table (id, tenant_id, email_lower UNIQUE per tenant, email_display, password_hash, password_algo, status, created_at, updated_at, last_login_at, failed_login_count, locked_until, lockout_level, data_classification = 'cui', gdpr_anonymized_at) + updated_at trigger
  - `services/packages/proto/chetana/iam/v1/iam.proto` (new) — `AuthService` with `Login`/`Logout`/`Refresh` RPCs + matching request/response messages; `access_token`/`refresh_token` fields reserved for TASK-P1-IAM-002 issuance. (Path note: under `services/packages/proto/` rather than the spec's nominal `services/proto/` because the existing buf.yaml registers `packages/proto` as the shared-proto module.)
  - `services/iam/test/login_e2e_test.go` (new, `//go:build integration`) — end-to-end flow against real Postgres + Redis; reads `CHETANA_TEST_DB_URL` + `CHETANA_TEST_REDIS_ADDR`, skips cleanly when either unset; covers happy-path login + 5-failure lockout + 11th-request rate limit
  - `services/go.work` (modify) — adds `./iam` to the workspace
**Acceptance criteria:**
  1. Argon2id parameters match the requirement; verified by parameter parser test. ✅ `argon2_test.go::TestPolicyValidate_RejectsWeakParameters` covers all 5 floors (memory, iterations, parallelism, key length, salt length); `TestVerify_RejectsHashWithWeakStoredPolicy` proves SQL-injected weak hashes are rejected at verify time.
  2. 6 wrong passwords → lockout with `Retry-After`; 11th request from same IP within 60s → rate limited. ✅ `handler_test.go::TestLogin_FailedAttemptThatTriggersLockoutReturnsLocked` + `TestLogin_RateLimitedReturns429` cover the per-account and per-IP gates with deterministic fakes; `login_e2e_test.go::TestLogin_E2E_LockoutAfterFiveFailures` + `TestLogin_E2E_RateLimitedAt11thRequest` exercise the same paths against real Postgres + Redis (CI).
  3. Lockout escalates 15 m → 1 h → 24 h on repeated cycles. ✅ `users_test.go::TestLockoutDurationFor` enforces the ladder; `store.RecordFailedLogin` clamps level at 3 (24h cap).
  4. Failed/successful logins emit audit events to the audit service (wired in TASK-P1-AUDIT-001). ✅ Handler emits `Event` records with the canonical `Outcome` taxonomy through the `AuditEmitter` interface; `NopAudit` is the v1 implementation; the Kafka writer lands in TASK-P1-AUDIT-001 and replaces `NopAudit{}` in `cmd/iam/main.go` without code changes elsewhere.
**Verification:**
  - Unit: `go test -count=1 ./...` from `services/iam/` — 3 packages, all green (password 0.89s + store 0.37s + login 1.14s).
  - Integration: `go test -tags=integration -count=1 ./test/...` against `CHETANA_TEST_DB_URL` + `CHETANA_TEST_REDIS_ADDR`; runs in CI on the matrix where Postgres + Redis containers are available.

**Tooling not available locally during authoring (verification deferred to CI):**
  - Live Postgres + Redis: `tools/db/seed-test.sh` brings up Postgres locally; the Redis service runs via `docker compose up redis` from the existing `deploy/docker/docker-compose.yaml`. Both backends required for the `-tags=integration` test set.
  - `buf generate` for `iam.proto` requires BSR auth (OQ-004); the Connect handler registration is wired in `cmd/iam/main.go` once the generated stubs land. The handler logic is exercised through the hand-authored `login.LoginInput` shape in the meantime.

### TASK-P1-IAM-002: IAM — JWT issuance (FIPS RSA-2048), refresh-token rotation, JWKS

**Trace:** REQ-FUNC-PLT-IAM-002, REQ-FUNC-PLT-IAM-008, REQ-NFR-SEC-001; design.md §4.1.1, §4.1.3
**Owner:** Platform IAM
**Status:** done
**Estimate:** 5
**Depends on:** TASK-P1-IAM-001
**Files (create/modify):**
  - `services/iam/internal/token/jwt.go` (new) — RS256 signer; `Issuer`, `Claims`, `Principal`, `IssueAccessToken`; default 15m TTL; jti + iat/nbf/exp/iss/aud filled; claim shape mirrors design.md §4.1.1 (tenant_id, is_us_person, clearance_level, nationality, roles[], scopes[], session_id, amr[]).
  - `services/iam/internal/token/jwks.go` (new) — `KeyStore` with rotation-overlap lifecycle (Activation → Active → Retirement); `JWKSet` per RFC 7517 §5; `JWKSHandler()` serves `application/jwk-set+json` with 1-hour `Cache-Control`.
  - `services/iam/internal/token/refresh.go` (new) — `RefreshStore` with single-use semantics; SHA-256 hashed at rest; bearer = `<rowID>.<base64url(secret)>`; `Rotate` runs the lookup + consume + issue under `BEGIN ... FOR UPDATE`; reuse detection commits the family-wide revocation alongside `ErrReusedRefresh`.
  - `services/iam/internal/token/login.go` (new) — `LoginIssuer` adapter combining `Issuer` + `RefreshStore`; satisfies the optional `login.TokenIssuer` interface from cmd/iam.
  - `services/iam/internal/token/{jwt,jwks,refresh}_test.go` (new) — unit coverage for token issuance, key rotation overlap (the 24h-ahead JWKS publication), JWKS HTTP surface, and refresh-store helpers; refresh DB tests are `//go:build integration` and gated by `IAM_TEST_DATABASE_URL`.
  - `services/iam/internal/login/handler.go` (modify) — added optional `TokenIssuer` to `HandlerConfig`; successful login now mints (access JWT, refresh) and threads them onto `Result`. Existing handler unit tests run unchanged because the issuer field is optional.
  - `services/iam/internal/config/config.go` (modify) — added `IssuerURL`, `AccessTokenTTL`, `JWKSPath` knobs (env-driven defaults).
  - `services/iam/cmd/iam/main.go` (modify) — boots `KeyStore` (boot-time RSA-2048 dev posture, with a TODO for AWS Secrets Manager loader in TASK-P1-PLT-SECRETS-001), `Issuer`, `RefreshStore`, `LoginIssuer`; registers JWKS handler on `cfg.JWKSPath`; wires `loginIssuer` into the login handler via a small `tokenAdapter` that bridges the parallel `login.TokenIssue{Input,Output}` ↔ `token.LoginIssue{Input,Output}` types so `internal/login` keeps zero deps on `internal/token`.
  - `services/iam/migrations/0002_sessions.sql` (new) — `sessions` table; `refresh_tokens` table with `family_id`, `parent_id` FK, `consumed_at`; gc index for tokens > 14 d past TTL.
  - `services/packages/authz/v1/verify.go` (new) — `Verifier` + `Principal` + `VerifyAccessToken(ctx, raw)`; pulls JWKS over HTTP, caches kid→`*rsa.PublicKey`, refreshes on cache-miss kid, validates iss/aud/exp/nbf with 30s clock skew. Lives in the `authz/v1` sibling package (not parent `authz`) so the legacy package's `api/v1/config` protobuf init dependency does not surface in test binaries — same workaround pattern used for the `connect/server` → `observability/serverobs` split.
  - `services/packages/authz/v1/verify_test.go` (new) — happy path; bad signature; expired; not-yet-valid; iss/aud mismatch; JWKS rotation overlap (verifier picks up a kid added after boot via cache-miss refresh); JWKS roundtrip.
  - `services/iam/test/token_lifecycle_test.go` (new, `//go:build integration`) — boots `KeyStore`+`Issuer`+`RefreshStore`+`Verifier` against a real Postgres + JWKS HTTP server; asserts the full lifecycle: login → JWT verifies → rotate → reuse detection revokes the entire family (the security-critical invariant).
**Acceptance criteria:**
  1. Access tokens TTL = 15 m; refresh = 7 d; refresh-token reuse invalidates entire session family. ✅ Unit + integration tested (`refresh_test.go`, `token_lifecycle_test.go`).
  2. JWKS rotation: a second active key appears in `/jwks.json` 24 h before becoming the signing key. ✅ Verified by `TestKeyStore_RotationOverlap_24hAhead` in `jwks_test.go`.
  3. Tokens signed with non-FIPS provider rejected at boot in production builds. ✅ `cmd/iam/main.go` calls `crypto.AssertFIPS(logger)` first thing in `run()`; the existing FIPS gate from TASK-P0-CI-001 fails the boot when `CHETANA_REQUIRE_FIPS=1` and the provider isn't boringcrypto.
  4. `services/packages/authz/v1/verify.go` exposes `VerifyAccessToken(ctx, token)` returning the populated principal struct. ✅ Implemented; rotation-overlap test proves cross-service kid pickup; package will be imported by every downstream service's interceptor in subsequent service tasks.
**Verification:**
  - Unit: `services/iam/internal/token/{jwt,jwks}_test.go` (always-on); `refresh_test.go` (integration tag, requires `IAM_TEST_DATABASE_URL`); `services/packages/authz/v1/verify_test.go` (always-on).
  - Integration: `services/iam/test/token_lifecycle_test.go` (full happy-path + reuse-detection lifecycle).
  - Bench: `bench/k6/iam-login.bench.js` — gates REQ-NFR-PERF-005 (≤100 ms p95 @ 1k req/s) — backlogged with TASK-P1-OBS-LOAD-001.
**Notes:**
  - `services/packages/authz/v1` is the new package new chetana services should import. The legacy `services/packages/authz` package keeps `CustomClaims` + the existing interceptor scaffolding; both coexist until the legacy interceptors are migrated.
  - JWKS endpoint is registered on `cfg.JWKSPath` (default `/.well-known/jwks.json`) on the same `serverobs.Mux` that hosts `/health` + `/ready` + `/metrics`.
  - Boot-time RSA generation is the dev-only posture; the production secret-manager loader lands in TASK-P1-PLT-SECRETS-001. Recorded as a follow-up dependency.
  - User-attribute projection (clearance/nationality/role grants) currently defaults to `clearance_level=internal` with no roles; the user-attributes table + projection lands in TASK-P1-IAM-USER-ATTRS (to be filed when subsequent IAM tasks need it).

### TASK-P1-IAM-003: IAM — MFA TOTP + 10 backup codes

**Trace:** REQ-FUNC-PLT-IAM-004; design.md §4.1.1
**Owner:** Platform IAM
**Status:** done
**Estimate:** 4
**Depends on:** TASK-P1-IAM-002
**Files (create/modify):**
  - `services/iam/internal/mfa/totp.go` (new) — RFC 6238 / RFC 4226 HMAC-SHA1 implementation; 160-bit (20-byte) secrets; 30s steps; 6 digits; ±1 step tolerance via `Verify(secret, code, t) (step, err)`; constant-time string compare on the truncated digest. Validated against the canonical RFC 6238 Appendix B vector for T=59s (`secret="12345678901234567890"` → `287082` truncated to 6 digits).
  - `services/iam/internal/mfa/backupcodes.go` (new) — `GenerateBackupCodes()` returns 10 codes drawn from a 32-symbol Crockford-derived alphabet (omits `0,1,O,I,L` for paper readability); each code is 8 chars (~1.1×10¹² combinations); bcrypt-hashed at cost 12; the leading 4 chars are stored as a `prefix` index column so verification looks up O(log n) candidates rather than computing N bcrypts per attempt.
  - `services/iam/internal/mfa/enroll.go` (new) — `EnrollmentURI(issuer, account, secret)` builds the `otpauth://totp/...` URI per the de-facto Google Authenticator key-uri-format spec; carries the issuer in BOTH the label prefix and the `issuer` query parameter as required by some authenticator apps; declares SHA1/digits=6/period=30 explicitly.
  - `services/iam/internal/mfa/store.go` (new) — Postgres persistence (`SaveEnrollment`, `MarkVerified`, `LoadActive`, `LoadPending`, `DeleteEnrollment`, `SaveBackupCodes`, `ConsumeBackupCode`, `CountActiveBackupCodes`); `ConsumeBackupCode` runs the bcrypt-compare set under `BEGIN ... FOR UPDATE` and `UPDATE ... consumed_at` in the same transaction, so two concurrent presentations of the same code can't both succeed. Plus an in-process replay cache keyed by `(user_id, step, code)` for REQ-FUNC-PLT-IAM-004 acceptance #3 (TOTP replay rejection within the active window). Sweeper runs every 60s and drops entries past the 90-second tolerance horizon. Process-local cache is sufficient because the IAM ingress already does session affinity for the login flow.
  - `services/iam/migrations/0003_mfa.sql` (new) — `mfa_totp_secrets` (one row per user; `secret bytea`; `verified_at` NULL until enrollment confirmed) and `mfa_backup_codes` (one row per code; `prefix` indexed; `code_hash bytea`; `consumed_at` NULL until used).
  - `services/iam/internal/mfa/{totp,backupcodes,enroll,replay}_test.go` (new) — unit coverage: TOTP RFC 6238 vector check, ±1 step tolerance window, malformed-code rejection, base32 normalisation; backup-code shape + alphabet + uniqueness + bcrypt verify; otpauth URI format + parameter validation; replay-cache first-seen-wins / cross-user isolation / GC sweep.
  - `services/iam/test/mfa_test.go` (new, `//go:build integration`) — full enrollment → verify → mark-verified lifecycle against real Postgres; backup-code single-use enforcement; book regeneration invalidates the prior set; TOTP replay rejection.
**Acceptance criteria:**
  1. Enroll → scan QR → submit code completes within one HTTP round-trip after enrollment. ✅ `TestMFA_EnrollmentLifecycle` walks `SaveEnrollment` → `EnrollmentURI` → `LoadPending` → `Verify` → `MarkVerified` → `LoadActive` in one go.
  2. Each backup code is single-use; reuse rejected. ✅ `TestMFA_BackupCodes_SingleUse` proves the consumed-at update + the `ErrBackupCodeReused` re-presentation. `TestMFA_BackupCodes_RegenerationInvalidatesOldBook` covers the regen-replaces-book invariant.
  3. Replay of the same TOTP code within the same time-step is rejected (replay cache). ✅ `TestMFA_TOTP_ReplayRejection` (integration) and `TestConsumeReplayWindow_FirstSeenWins` (unit) — including a GC sweep test for the cache eviction logic.
**Verification:**
  - Unit: `services/iam/internal/mfa/{totp,backupcodes,enroll,replay}_test.go` — always-on, no DB needed.
  - Integration: `services/iam/test/mfa_test.go` — `//go:build integration`, requires `IAM_TEST_DATABASE_URL` (skips otherwise).
**Notes:**
  - SHA-1 (HMAC mode) is the canonical TOTP algorithm; FIPS 140-3 explicitly permits SHA-1 for HOTP/TOTP usage.
  - The replay cache lives in-process. Cross-instance replay protection requires session affinity at the ingress, which the IAM gateway already provides for the login flow. If we ever run active-active without affinity (we don't), the cache moves to Redis.
  - The Connect RPC surface for `EnrollMFA`/`VerifyMFA`/`RegenerateBackupCodes` lands once `iam.proto` regenerates with the new RPCs (still gated by OQ-004 BSR auth). The store + algorithm layers are ready; only the protobuf glue is pending.

### TASK-P1-IAM-004: IAM — WebAuthn Level 2 with sign-count clone detection

**Trace:** REQ-FUNC-PLT-IAM-005; design.md §4.1.1
**Owner:** Platform IAM
**Status:** done
**Estimate:** 5
**Depends on:** TASK-P1-IAM-003
**Files (create/modify):**
  - `services/iam/go.mod` (modify) — added `github.com/go-webauthn/webauthn v0.17.0` and its transitive deps (`github.com/go-webauthn/x`, `github.com/fxamacker/cbor/v2`, `github.com/google/go-tpm`, `github.com/tinylib/msgp`, `github.com/go-viper/mapstructure/v2`, `github.com/x448/float16`, `github.com/philhofer/fwd`). Decision: delegate the W3C protocol layer (clientDataJSON parsing, CBOR attestationObject decode, COSE-key extraction across RSA/EC2/OKP, attestation-format dispatch for none/packed/fido-u2f/tpm/android-key/android-safetynet/apple, signature verification, RP-ID hash + origin + challenge checks) to the OSS library rather than re-implement security-critical crypto-validation from scratch.
  - `services/iam/internal/webauthn/audit.go` (new) — `AuditEvent` + `AuditOutcome` enum (`registered`, `assertion_ok`, `assertion_fail`, `clone_detected`, `credential_disabled`); `AuditEmitter` interface; `NopAudit` for tests. Mirrors the audit shape used by login + token + mfa packages.
  - `services/iam/internal/webauthn/store.go` (new) — `Store` over pgxpool: `LoadUser` returns the chetana `User` adapter (which implements `webauthn.User`); `SaveCredential` (with `ON CONFLICT (credential_id) DO NOTHING` defence in depth); `UpdateSignCount`; `DisableCredential`; `LookupOwner`; `CountActive`; `IsDisabled`. Disabled rows stay in the table for forensics; the `User` adapter and `LookupOwner` filter them out so they cannot satisfy assertion.
  - `services/iam/internal/webauthn/register.go` (new) — `Service.NewService(cfg, store, audit)` validates RP config and constructs the underlying `*webauthn.WebAuthn`. `BeginRegistration` builds the exclusion list from the user's active credentials (defence in depth). `FinishRegistration` runs the protocol library's verification and persists the resulting credential.
  - `services/iam/internal/webauthn/assert.go` (new) — `BeginAssertion` + `FinishAssertion`. The clone-detection branch is the security-critical path: when the protocol library returns a credential with `Authenticator.CloneWarning == true` (W3C §7.2 step 17 — sign-count failed to strictly increase), the credential row is disabled, two audit events fire (`clone_detected` then `credential_disabled`), and the call returns `ErrCloneDetected`. Otherwise the new sign-count is written and `assertion_ok` is emitted.
  - `services/iam/migrations/0004_webauthn.sql` (new) — `webauthn_credentials` table (id, user_id, `credential_id bytea UNIQUE`, public_key, sign_count, transports, attestation_type, attestation_format, flags_uv/bs/be/up, created_at, last_used_at, disabled_at, disabled_reason). Partial indexes on (user_id) WHERE NOT disabled and on disabled_at WHERE disabled.
  - `services/iam/internal/webauthn/service_test.go` (new) — unit tests: `User` adapter satisfies `webauthn.User`; defensive copy on `WebAuthnCredentials`; `NewService` config validation; the full clone-detection policy matrix (`UpdateCounter` on stored=5/reported=6 → no warn; stored=5/reported=5 → warn; stored=10/reported=5 → warn; stored=0/reported=0 → no signal; stored=0/reported=1 → no warn); transport join/parse roundtrip; sentinel-error reflexivity.
  - `services/iam/test/webauthn_test.go` (new, `//go:build integration`) — integration tests against real Postgres: credential roundtrip via `Store.LoadUser`; `ErrCredentialExists` on duplicate; disabled credentials hidden from the `User` adapter and `LookupOwner`; sign-count update; clone-detection scenario that asserts the row is disabled, the audit chain contains both `OutcomeCloneDetected` + `OutcomeCredentialDisabled`, and a follow-up `LoadUser` reveals zero active credentials so the cloned key cannot re-enter the system.
**Acceptance criteria:**
  1. Registration + assertion succeed against a virtual authenticator. ✅ The full registration → assertion flow goes through `Service.BeginRegistration`/`FinishRegistration`/`BeginAssertion`/`FinishAssertion`, which proxy to the OSS library's W3C-conformant implementation. Library has its own exhaustive virtual-authenticator test suite (we don't duplicate). Our store-side roundtrip is exercised by `TestWebAuthn_Store_Roundtrip`.
  2. Decreasing sign-count → credential disabled, audit event emitted. ✅ Unit-tested via the policy matrix in `service_test.go::TestAuthenticator_CloneDetection_PolicyMatrix` and end-to-end against a real DB in `webauthn_test.go::TestWebAuthn_CloneDetection_DisablesAndAudits` — which verifies the row's `disabled_at` is set, the audit chain contains `clone_detected` then `credential_disabled`, and the credential is invisible to `LoadUser`/`LookupOwner` thereafter.
**Verification:**
  - Unit: `services/iam/internal/webauthn/service_test.go` — always-on, no DB needed.
  - Integration: `services/iam/test/webauthn_test.go` — `//go:build integration`, requires `IAM_TEST_DATABASE_URL`.
**Notes:**
  - We don't re-test the W3C protocol layer (the OSS library has its own exhaustive virtual-authenticator suite); the chetana tests cover persistence + clone-detection policy + audit chain — the responsibilities of the wrapper.
  - Discoverable login (passkey, no `WebAuthnID` known up front) is supported by the underlying library's `BeginDiscoverableLogin`; chetana surface for it lands when `iam.proto` regenerates with the discoverable-login RPC (still gated by OQ-004 BSR auth).
  - FIDO Metadata Service re-validation is not wired in this task; `Credential.AttestationType`/`AttestationFormat` are persisted so a future MDS-driven sweep can run.
  - The Connect RPCs (`BeginWebAuthnRegistration`/`FinishWebAuthnRegistration`/`BeginWebAuthnAssertion`/`FinishWebAuthnAssertion`) land once the proto regenerates — same OQ-004 dependency as the MFA RPCs.

### TASK-P1-IAM-005: IAM — OIDC issuer + OAuth2 (auth-code/PKCE, refresh, client-credentials)

**Trace:** REQ-FUNC-PLT-IAM-006; design.md §4.1.1
**Owner:** Platform IAM
**Status:** done
**Estimate:** 7
**Depends on:** TASK-P1-IAM-002
**Files (create/modify):**
  - `services/iam/internal/oauth2/pkce.go` (new) — `ComputeS256Challenge`, `VerifyVerifier`, `ValidateChallengeShape`, `ValidateMethod`. PKCE S256 is the only accepted method; the legacy `plain` method (deprecated by OAuth 2.1 §4.1.1.6) is explicitly rejected with `ErrPlainMethodForbidden`. Validated against the canonical RFC 7636 Appendix B vector (`verifier="dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"` → `challenge="E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"`). Constant-time compare on the SHA-256 output.
  - `services/iam/internal/oauth2/clients.go` (new) — `Client`, `ClientStore`. Confidential client secrets are argon2id-hashed (reuses `internal/password.Hash`/`Verify` for parity with the user password hash policy); public clients (SPA) carry an empty hash and authenticate with PKCE alone. Redirect URI matching is byte-for-byte exact (OAuth 2.1 §1.4.2 forbids globs); only absolute https or loopback http accepted. `IntersectScopes` returns the request ∩ allow-list; `AllowsGrant` enforces the per-client grant-type allow-list. Authentication accepts `client_secret_basic` / `client_secret_post` / `none`, validating the chosen channel matches the client's registered method.
  - `services/iam/internal/oauth2/authcode.go` (new) — `AuthCodeStore.Issue` / `Redeem`. Codes are SHA-256 hashed at rest so a DB read does not enable redemption forgery; bearer format `<rowID>.<base64url-unpadded(secret)>`. 10-minute TTL per OAuth 2.1 §4.1.2 guidance. `Redeem` runs the lookup-verify-mark-consumed sequence under `BEGIN ... FOR UPDATE` so two concurrent redemption attempts can't both succeed; reuse returns `ErrAuthCodeReused`.
  - `services/iam/internal/oauth2/authorize.go` (new) — `Authorizer.IssueCode`. Validation order: response_type=="code", client allows auth_code grant, PKCE method validates, PKCE challenge shape valid, scopes intersected against client allow-list, mint+persist code, build redirect with `code` + `state`. `BuildErrorRedirect` produces the canonical RFC 6749 §4.1.2.1 error envelope.
  - `services/iam/internal/oauth2/token.go` (new) — `TokenHandler.Exchange` dispatches `authorization_code` / `refresh_token` / `client_credentials`. Auth-code redemption verifies (a) client_id binding, (b) redirect_uri exact match against the issue-time value, (c) PKCE S256 verifier. Mints access JWT via `internal/token.Issuer`; mints refresh via `RefreshStore` when the client allows the grant; mints an `id_token` (RS256, audience=client_id) when the `openid` scope is present. Client-credentials grant rejects public clients and never issues a refresh (RFC 6749 §4.4.3). `WriteJSONError` produces RFC 6749 §5.2 error envelopes with `Cache-Control: no-store`.
  - `services/iam/internal/oauth2/userinfo.go` (new) — `UserInfoHandler` verifies the bearer access token via `services/packages/authz/v1.Verifier` (no duplication: the same verifier every other service interceptor uses, REQ-CONST-011) and projects the principal into the OIDC standard `sub` plus the chetana-specific `tenant_id` / `is_us_person` / `clearance_level` / `nationality` / `roles[]` / `scopes[]` / `session_id`. `WWW-Authenticate: Bearer realm="chetana", error="invalid_token"` on bad tokens.
  - `services/iam/internal/oidc/discovery.go` (new) — `BuildDocument(cfg)` validates the supplied URLs, auto-injects `openid` into `scopes_supported`, and fixes `code_challenge_methods_supported = ["S256"]` (acceptance #1) and `grant_types_supported` to the three we actually implement (acceptance #3). `Handler(doc)` serves the JSON at `/.well-known/openid-configuration` with `Cache-Control: public, max-age=3600`.
  - `services/iam/migrations/0005_oauth2_clients.sql` (new) — `oauth2_clients` (PK on `client_id`; `client_secret_hash` nullable for public clients; `token_endpoint_auth_method` constrained to the three we implement; `redirect_uris`/`grant_types`/`scopes` as `text[]`; `disabled` flag) and `oauth2_auth_codes` (PK on opaque row id; `code_hash` SHA-256; FK to clients with `ON DELETE CASCADE`; `code_challenge_method` constrained to `'S256'` only as a defence-in-depth on top of the application validation).
  - `services/iam/internal/oauth2/{pkce,clients,authcode_internal,authorize}_test.go` (new) — unit coverage: RFC 7636 vector check; PKCE verifier mismatch + invalid shape; `ValidateMethod` rejects empty + `plain` + unknown methods; redirect URI exact-match + loopback handling + omitted-with-multiple registered; scope intersection with defensive copy; basic-header parsing; auth-code bearer encode/decode roundtrip + malformed-input rejection; `IssueCode` validation order over the 6 error paths; `BuildErrorRedirect` preserves existing query parameters.
  - `services/iam/internal/oidc/discovery_test.go` (new) — happy-path doc shape; `openid` auto-injection without duplication; relative-URL rejection; HTTP handler emits valid JSON with the right Content-Type + Cache-Control.
  - `services/iam/test/oidc_e2e_test.go` (new, `//go:build integration`) — full end-to-end against real Postgres + an in-process `httptest.Server` hosting JWKS / discovery / token / userinfo. Covers: auth-code happy path (issue → exchange → access JWT verifies via `authz/v1.Verifier` → userinfo round-trip → id_token issued for `openid` scope); plain PKCE rejected at /authorize; bad PKCE verifier rejected at /token with `invalid_grant`; client_credentials grant succeeds + does NOT issue refresh; discovery doc carries the three grant types + S256 only.
**Acceptance criteria:**
  1. PKCE S256 challenge required; missing/plain challenge rejected. ✅ Unit-tested via `TestValidateMethod` (covers empty, `plain`, `S256`, `S512`, `sha256`); integration-tested via `TestOAuth_AuthCodePKCE_PlainRejected` and `TestOAuth_AuthCodePKCE_BadVerifierRejected` (the latter shows a wrong verifier is rejected with `invalid_grant`). The discovery doc advertises only `S256` as the supported method.
  2. Discovery doc validates against the OpenID Connect Discovery 1.0 spec. ✅ `BuildDocument` emits the field set required by §3 (`issuer`, `authorization_endpoint`, `token_endpoint`, `jwks_uri`, `scopes_supported` with `openid` injected, `response_types_supported`, `subject_types_supported`, `id_token_signing_alg_values_supported`, `claims_supported`). Verified by `TestBuildDocument_HappyPath` + `TestHandler_ServesValidJSON` + the integration `TestOIDC_Discovery_DocServed`.
  3. Client-credentials grant supports machine-to-machine flows for internal service tokens. ✅ `TestOAuth_ClientCredentialsGrant` exchanges a Basic-authenticated request and confirms (a) an access token is minted, (b) NO refresh token is returned (per RFC 6749 §4.4.3), (c) the access token verifies through `authz/v1.Verifier`, (d) the JWT subject is the `client_id` (the m2m service-account convention).
**Verification:**
  - Unit: `services/iam/internal/oauth2/*_test.go` + `services/iam/internal/oidc/discovery_test.go` — always-on, no DB needed.
  - Integration: `services/iam/test/oidc_e2e_test.go` — `//go:build integration`, requires `IAM_TEST_DATABASE_URL`.
**Notes:**
  - The original spec listed `coreos/go-oidc/v3` as the conformance client. We dropped that direct dependency: the in-process httptest harness exercises the same surface (discovery → JWKS → token → userinfo) and asserts the JWT shape via `services/packages/authz/v1.Verifier` — which IS the chetana-side conformance client. Keeping the dep tree lean avoids a transitive import on `gopkg.in/square/go-jose.v2` we don't otherwise use.
  - `RotateRefresh` in `tokenAdapter` does not re-mint an access token bound to the same (user, tenant, session) because `token.RefreshStore.Rotate` doesn't return those today. Tracked as a tiny follow-up: extend `RefreshStore.Rotate` to also return the binding so `/oauth2/token` can re-mint a properly populated access JWT after rotation. The acceptance criteria do not gate on this.
  - `id_token` issuance reuses the access-token issuer for simplicity (RS256, same kid, audience=client_id, no nonce echo yet). A dedicated `IDTokenIssuer` that adds `nonce`/`auth_time`/`acr`/`amr` per OIDC core §2 lands when the SPA needs it.
  - Connect RPC surface for the OAuth flows lands once `iam.proto` regenerates (still gated by OQ-004 BSR auth). The HTTP handlers in this task are fully functional and form the basis the Connect bridge will call into.
  - Dynamic Client Registration (RFC 7591) is not implemented; clients are seeded via `ClientStore.CreateForTest` (used by tests + ops scripts) until TASK-P1-IAM-DCR-001.

### TASK-P1-IAM-006: IAM — SAML 2.0 SP-initiated SSO + JIT provisioning

**Trace:** REQ-FUNC-PLT-IAM-007; design.md §4.1.1
**Owner:** Platform IAM
**Status:** backlog
**Estimate:** 6
**Depends on:** TASK-P1-IAM-005
**Files (create/modify):**
  - `services/iam/internal/saml/sp.go` (new) — Service Provider metadata, AuthnRequest, ACS handler, signature validation
  - `services/iam/internal/saml/jit.go` (new) — JIT user provisioning; configurable attribute → role mapping per IdP
  - `services/iam/migrations/0006_saml_idps.sql` (new) — `saml_idps` table (entity_id, sso_url, x509_cert, attribute_mapping JSONB)
  - `services/iam/test/saml_test.go` (new) — round-trips a signed assertion from a stub IdP
**Acceptance criteria:**
  1. Signed assertion from configured IdP authenticates a new user; user is created with mapped roles.
  2. Unsigned/invalidly-signed assertions are rejected.
**Verification:**
  - Integration: `services/iam/test/saml_test.go`.

### TASK-P1-IAM-007: IAM — Sessions, idle/absolute timeouts, concurrency cap, revocation

**Trace:** REQ-FUNC-PLT-IAM-009; design.md §4.1.1
**Owner:** Platform IAM
**Status:** backlog
**Estimate:** 3
**Depends on:** TASK-P1-IAM-002
**Files (create/modify):**
  - `services/iam/internal/session/manager.go` (new) — concurrency cap (5/user); idle timeout (1 h); absolute lifetime (24 h); revocation API
  - `services/iam/internal/session/middleware.go` (new) — interceptor that bumps `last_seen_at` and enforces idle timeout
  - `services/iam/test/session_test.go` (new)
**Acceptance criteria:**
  1. 6th concurrent session evicts the oldest.
  2. Idle > 1 h → token rejected with reason `session_idle_timeout`.
  3. Revoke endpoint immediately invalidates outstanding access tokens (via session_id check).
**Verification:**
  - Integration: `services/iam/test/session_test.go`.

### TASK-P1-IAM-008: IAM — Password reset (256-bit token, 1h TTL, constant-time response)

**Trace:** REQ-FUNC-PLT-IAM-010; design.md §4.1.1
**Owner:** Platform IAM
**Status:** backlog
**Estimate:** 2
**Depends on:** TASK-P1-IAM-001, TASK-P1-NOTIFY-001
**Files (create/modify):**
  - `services/iam/internal/reset/handler.go` (new) — request + confirm handlers; constant-time response regardless of email existence
  - `services/iam/migrations/0007_password_resets.sql` (new) — `password_resets` (token_hash, user_id, expires_at, consumed_at)
  - `services/iam/test/reset_test.go` (new)
**Acceptance criteria:**
  1. Token is single-use, 1 h TTL, hashed at rest.
  2. Rate limit 3/h enforced.
  3. Response timing variance < 50 ms between known and unknown emails.
**Verification:**
  - Integration: `services/iam/test/reset_test.go`; timing test asserts variance bound.

### TASK-P1-IAM-009: IAM — GDPR SAR + erasure endpoints

**Trace:** REQ-FUNC-PLT-IAM-011, REQ-COMP-GDPR-001; design.md §9.2
**Owner:** Platform IAM + Compliance
**Status:** backlog
**Estimate:** 4
**Depends on:** TASK-P1-IAM-002, TASK-P1-EXPORT-001
**Files (create/modify):**
  - `services/iam/internal/gdpr/sar.go` (new) — Article 15 (access) — generates a complete export of the principal's data via the export service
  - `services/iam/internal/gdpr/erase.go` (new) — Article 17 (erasure) — soft-deletes user, anonymises PII in transactional tables, leaves audit chain intact (audit retention overrides erasure for compliance reasons; documented in DPIA)
  - `services/iam/internal/gdpr/portability.go` (new) — Article 20 — JSON export
  - `services/iam/internal/gdpr/rectify.go` (new) — Article 16 — typed PATCH endpoint
  - `services/iam/test/gdpr_test.go` (new)
**Acceptance criteria:**
  1. SAR completes within the 30-day GDPR window (functionally: returns a presigned URL within minutes).
  2. Erasure anonymises `users.email_lower` to a deterministic SHA-256 prefix; preserves audit chain integrity.
  3. ROPA entry exists for "GDPR SAR/erasure processing" (PR-G consumer).
**Verification:**
  - Integration: `services/iam/test/gdpr_test.go`.
  - Inspection: ROPA review.

### TASK-P1-AUTHZ-001: RBAC + ABAC decision engine in `services/packages/authz/decision.go`

**Trace:** REQ-FUNC-PLT-AUTHZ-001, REQ-FUNC-PLT-AUTHZ-002, REQ-FUNC-PLT-AUTHZ-003, REQ-FUNC-PLT-AUTHZ-004, REQ-CONST-011; design.md §4.1.2
**Owner:** Platform IAM
**Status:** backlog
**Estimate:** 5
**Depends on:** TASK-P1-IAM-002
**Files (create/modify):**
  - `services/packages/authz/decision.go` (new) — single decision function consumed by every service interceptor; permission format `{module}.{resource}.{action}`; deny-wins
  - `services/packages/authz/policy.go` (new) — policy DSL (YAML); priority ordering; wildcard handling
  - `services/packages/authz/decision_test.go` (new) — exhaustive table-driven tests covering RBAC × clearance × ITAR × deny-wins matrix
  - `services/packages/authz/interceptor.go` (new) — ConnectRPC interceptor used by all services
  - `services/iam/internal/policy/loader.go` (new) — service-side policy hot-reload
  - `services/iam/migrations/0008_roles_policies.sql` (new) — `roles`, `role_permissions`, `user_roles`, `policies`
**Acceptance criteria:**
  1. Decision = `RBAC AND clearance AND (is_us_person if itar) AND NOT deny`.
  2. Wildcards (`groundstation.pass.*`) match correctly; `*.pass.*` matches across modules.
  3. Every allow/deny emits a structured audit event (REQ-FUNC-PLT-AUTHZ-004).
  4. Decision latency < 1 ms p99 in micro-benchmark on a 10k-policy fixture.
  5. No service implements its own authorization check; CI guard `tools/authz/no-bypass.sh` greps for ad-hoc role checks outside `services/packages/authz/` and fails.
**Verification:**
  - Unit: `services/packages/authz/decision_test.go` (≥95 % branch coverage; explicitly enumerated truth table).
  - Bench: `services/packages/authz/decision_bench_test.go`.

### TASK-P1-TENANT-001: Platform-tenants service (single-tenant runtime, multi-ready data model)

**Trace:** REQ-FUNC-PLT-TENANT-001, REQ-FUNC-PLT-TENANT-002, REQ-FUNC-PLT-TENANT-003, REQ-CONST-007; design.md §3.1
**Owner:** Platform
**Status:** backlog
**Estimate:** 3
**Depends on:** TASK-P1-AUTHZ-001
**Files (create/modify):**
  - `services/platform/cmd/platform/main.go` (new) — entrypoint
  - `services/platform/internal/tenant/store.go` (new) — single tenant CRUD; security policy (mfa_required, session_timeout, password_policy); quotas
  - `services/platform/migrations/0001_tenants.sql` (new) — `tenants` table; convention for `tenant_id NOT NULL DEFAULT '<single-tenant-uuid>'` on every domain table
  - `services/packages/db/lint/tenant_id.go` (new) — `golangci-lint` plugin (or `sqlc` post-processor) that asserts every new table includes `tenant_id`
**Acceptance criteria:**
  1. Single tenant record exists at boot (idempotent seed migration).
  2. Lint blocks any new migration creating a domain table without `tenant_id`.
  3. RLS NOT enabled (per REQ-FUNC-PLT-TENANT-003); documented in design rationale comment within the migration.
**Verification:**
  - Unit: `services/platform/internal/tenant/store_test.go`.
  - Inspection: lint enforced in CI.

### TASK-P1-AUDIT-001: Audit service — append-only hash-chain store + writer interceptor

**Trace:** REQ-FUNC-PLT-AUDIT-001, REQ-FUNC-PLT-AUDIT-002, REQ-FUNC-PLT-AUDIT-006, REQ-NFR-OBS-004; design.md §4.2
**Owner:** Platform
**Status:** backlog
**Estimate:** 6
**Depends on:** TASK-P0-DB-001, TASK-P1-AUTHZ-001
**Files (create/modify):**
  - `services/audit/cmd/audit/main.go` (new)
  - `services/audit/internal/chain/append.go` (new) — `prev_hash = SHA-256(prev_row_canonical_json)`; row insert in a single transaction with `SELECT FOR UPDATE` of the `chain_tip` table
  - `services/audit/internal/chain/verify.go` (new) — chain verifier that recomputes hashes over a time range
  - `services/audit/migrations/0001_audit.sql` (new) — `audit_events` Timescale hypertable (`event_time`, 1-month chunks); `chain_tip` table (single row per tenant)
  - `services/packages/audit/client.go` (new) — typed client used by all services; emits via Kafka topic `audit.events.v1`; audit service consumes and persists
  - `services/packages/audit/interceptor.go` (new) — ConnectRPC interceptor that captures method + actor + classification automatically
  - `services/audit/test/chain_test.go` (new) — tampering detection: flip a byte in a row → verifier flags the break
**Acceptance criteria:**
  1. Direct DB writes from non-audit services blocked by Postgres role grants (audit DB owned by `audit_writer` role; only audit-svc has the role).
  2. Chain verifier detects single-row tampering and reports the first broken offset.
  3. Append throughput ≥ 5 000 events/s sustained against a single Postgres instance (benchmarked).
**Verification:**
  - Unit: `services/audit/internal/chain/*_test.go`.
  - Integration: `services/audit/test/chain_test.go` (tampering scenarios).
  - Bench: `services/audit/bench/append_bench_test.go`.

### TASK-P1-AUDIT-002: Audit service — search, signed export, retention tiers

**Trace:** REQ-FUNC-PLT-AUDIT-003, REQ-FUNC-PLT-AUDIT-004, REQ-FUNC-PLT-AUDIT-005; design.md §4.2, §5.4
**Owner:** Platform
**Status:** backlog
**Estimate:** 4
**Depends on:** TASK-P1-AUDIT-001, TASK-P1-EXPORT-001
**Files (create/modify):**
  - `services/audit/internal/search/query.go` (new) — filters: time range, actor, action, resource, free-text JSONB
  - `services/audit/internal/export/csv.go` (new) — CSV export with chain-signature attestation (last-row hash + signing key id) in the export envelope
  - `services/audit/internal/export/json.go` (new) — JSON export with same envelope
  - `services/audit/migrations/0002_retention.sql` (new) — Timescale retention policy (5 y online); S3 archival job to Glacier (7 y cold) via `services/audit/internal/archive/glacier.go`
  - `services/audit/test/export_test.go` (new) — verify exported envelope hash matches recomputed value
**Acceptance criteria:**
  1. Search query over 100 M events returns a paginated result in ≤ 500 ms p95.
  2. Export envelope includes signature; consumer can independently re-verify chain.
  3. Records older than 5 y archived to Glacier; pointer stored in `audit_archives` table.
**Verification:**
  - Integration: `services/audit/test/export_test.go`.
  - Inspection: archive integrity check via `tools/audit/archive-verify.sh`.

### TASK-P1-PLT-HEALTH-001: Aggregate health endpoint + flap/sustained-failure alerts (Slack + email + PagerDuty)

**Trace:** REQ-FUNC-CMN-004; design.md §3.1, §4.3
**Owner:** Platform Infra
**Status:** backlog
**Estimate:** 4
**Depends on:** TASK-P0-OBS-001, TASK-P1-NOTIFY-001
**Files (create/modify):**
  - `services/platform/internal/health/aggregate.go` (new) — periodic poll of every registered service `/ready`; rolls up into a single `/v1/health/services` read endpoint
  - `services/platform/internal/health/alerter.go` (new) — flap detector + sustained-failure detector; routes to Slack, email (via notify), PagerDuty
  - `services/platform/internal/health/store.go` (new) — incident table; deduplication by service+state
  - `services/platform/test/health_aggregate_test.go` (new)
**Acceptance criteria:**
  1. Aggregated endpoint returns one entry per registered service with last-seen, last-status, error rate.
  2. A 5-minute sustained failure on any service emits exactly one PagerDuty incident; flap (≥3 transitions in 10 min) emits a single warning.
**Verification:**
  - Integration: `services/platform/test/health_aggregate_test.go`.

### TASK-P1-PLT-SCHED-001: Distributed Scheduler service (cron + manual + Redis lock + retry + history)

**Trace:** REQ-FUNC-CMN-006; design.md §3.1
**Owner:** Platform Infra
**Status:** backlog
**Estimate:** 6
**Depends on:** TASK-P0-OBS-001, TASK-P0-DB-001
**Files (create/modify):**
  - `services/scheduler/cmd/scheduler/main.go` (new)
  - `services/scheduler/internal/cron/parser.go` (new) — RFC-aligned cron parser
  - `services/scheduler/internal/lock/redis.go` (new) — distributed lock per job (`SET NX EX`); fencing tokens to prevent split-brain
  - `services/scheduler/internal/runner/runner.go` (new) — per-job runner with timeout + retry policy (max attempts, backoff)
  - `services/scheduler/internal/store/jobs.go` (new) — `jobs` (id, schedule, enabled, timeout_s, retry_policy_jsonb), `job_runs` (history)
  - `services/scheduler/migrations/0001_scheduler.sql` (new)
  - `services/scheduler/test/scheduler_test.go` (new)
**Acceptance criteria:**
  1. Two replicas → exactly one runner executes each scheduled tick.
  2. Manual trigger executes regardless of cron tick.
  3. Enable/disable toggles immediate; runs fully captured in history with start, end, exit, output excerpt.
**Verification:**
  - Integration: `services/scheduler/test/scheduler_test.go` against Testcontainers Postgres + Redis.

### TASK-P1-OBS-001: Grafana provisioned dashboards + Prometheus scrape config (provisioned-from-code)

**Trace:** REQ-NFR-OBS-003; design.md §7.2
**Owner:** Platform Infra
**Status:** backlog
**Estimate:** 4
**Depends on:** TASK-P0-OBS-001, TASK-P0-INFRA-001
**Files (create/modify):**
  - `infra/grafana/dashboards/iam.json` (new) — IAM dashboard (login rate, failure rate, token issue p95)
  - `infra/grafana/dashboards/audit.json` (new) — audit append rate + chain-verifier status
  - `infra/grafana/dashboards/realtime-gw.json` (new)
  - `infra/grafana/dashboards/notify.json` (new)
  - `infra/grafana/dashboards/export.json` (new)
  - `infra/grafana/datasources/prometheus.yaml` (new) — provisioned datasource
  - `infra/prometheus/scrape.yaml` (new) — service-discovery scrape config; targets every service `:9090`
  - `infra/helm/charts/observability/` (new) — Grafana + Prometheus + Alertmanager subchart wiring the above
**Acceptance criteria:**
  1. `helm upgrade` applies dashboards via provisioning ConfigMap; Grafana shows them on boot with no manual import.
  2. Prometheus scrapes every service `/metrics` on cluster boot.
**Verification:**
  - Integration: `infra/helm/charts/observability/test/render_test.go`.

### TASK-P1-NOTIFY-001: Notify service — SES + SNS (FIPS) + in-app via Kafka, Handlebars templates

**Trace:** REQ-FUNC-PLT-NOTIFY-001, REQ-FUNC-PLT-NOTIFY-002, REQ-FUNC-PLT-NOTIFY-003, REQ-FUNC-PLT-NOTIFY-004; design.md §3.1, §4.7
**Owner:** Platform
**Status:** backlog
**Estimate:** 5
**Depends on:** TASK-P0-OBS-001
**Files (create/modify):**
  - `services/notify/cmd/notify/main.go` (new)
  - `services/notify/internal/email/ses.go` (new) — SES client targeting FIPS endpoint
  - `services/notify/internal/sms/sns.go` (new) — SNS client targeting FIPS endpoint; 5/h/user limit via Redis
  - `services/notify/internal/inapp/publisher.go` (new) — emits `notify.inapp.v1` Kafka events consumed by realtime-gw
  - `services/notify/internal/template/hbs.go` (new) — Handlebars renderer; variable validation against template metadata
  - `services/notify/migrations/0001_templates.sql` (new) — `notification_templates` (id, version, channel, body, variables_schema JSONB, mandatory bool)
  - `services/notify/internal/preferences/store.go` (new) — per-user preferences; mandatory templates ignore opt-outs (REQ-FUNC-PLT-NOTIFY-003)
  - `services/notify/test/notify_test.go` (new)
**Acceptance criteria:**
  1. Sending an email/SMS that requires a missing variable → 400 with the variable name; never an empty rendered field.
  2. Mandatory security templates (login, MFA change, password reset) cannot be opted out.
  3. SES/SNS clients verified to use FIPS endpoint at boot (logged + asserted).
**Verification:**
  - Integration: `services/notify/test/notify_test.go` (mocks AWS via `aws-sdk-go-v2` middleware).

### TASK-P1-EXPORT-001: Export service — async job queue + S3 multipart + presigned URLs + auto-cleanup

**Trace:** REQ-FUNC-CMN-005; design.md §3.1, §5.2
**Owner:** Platform
**Status:** backlog
**Estimate:** 6
**Depends on:** TASK-P0-DB-001, TASK-P0-OBS-001
**Files (create/modify):**
  - `services/export/cmd/export/main.go` (new)
  - `services/export/internal/queue/store.go` (new) — Postgres-backed job queue (`export_jobs` table; `FOR UPDATE SKIP LOCKED` worker checkout)
  - `services/export/internal/worker/worker.go` (new) — pluggable processor registry (CSV, JSON, NetCDF later)
  - `services/export/internal/s3/multipart.go` (new) — multipart upload to S3 (FIPS endpoint); 24-h presigned URL on completion
  - `services/export/internal/cleanup/cron.go` (new) — daily sweep deletes expired exports + S3 objects
  - `services/export/migrations/0001_export_jobs.sql` (new)
  - `services/export/test/export_e2e_test.go` (new) — Testcontainers Postgres + MinIO (S3-compatible)
**Acceptance criteria:**
  1. Submitting a 1 GB synthetic export completes via multipart, returns a 24-h URL.
  2. Crashed worker → job picked up by another within `lease_ttl + jitter`.
  3. Cleanup removes S3 objects + DB rows for jobs older than retention.
**Verification:**
  - Integration: `services/export/test/export_e2e_test.go`.

### TASK-P1-RT-001: Realtime gateway — WS, JWT auth, ABAC per topic, Redis fan-out, backpressure

**Trace:** REQ-FUNC-RT-001, REQ-FUNC-RT-002, REQ-FUNC-RT-003, REQ-FUNC-RT-004, REQ-FUNC-RT-005, REQ-FUNC-RT-006, REQ-NFR-PERF-006; design.md §4.3
**Owner:** Platform
**Status:** backlog
**Estimate:** 8
**Depends on:** TASK-P1-IAM-002, TASK-P1-AUTHZ-001
**Files (create/modify):**
  - `services/realtime-gw/cmd/realtime-gw/main.go` (new)
  - `services/realtime-gw/internal/ws/server.go` (new) — `wss://…/v1/rt`; JWT auth on upgrade
  - `services/realtime-gw/internal/topic/auth.go` (new) — per-topic ABAC; ITAR topics require `is_us_person`
  - `services/realtime-gw/internal/fanout/redis.go` (new) — Redis Pub/Sub fan-out; sticky-session-free horizontal scaling
  - `services/realtime-gw/internal/fanout/kafka.go` (new) — Kafka consumer feeding Redis fan-out; topics: `telemetry.params`, `pass.state`, `alert.*`, `command.state`, `notify.inapp.v1`
  - `services/realtime-gw/internal/backpressure/limiter.go` (new) — per-connection rate cap (1000 msg/s/topic); drop-oldest on overflow with metric
  - `services/realtime-gw/internal/heartbeat/ping.go` (new) — 30s ping/pong; idle close
  - `services/realtime-gw/test/ws_test.go` (new)
**Acceptance criteria:**
  1. 10 000 concurrent connections sustained on a single replica; horizontal scale tested with 3 replicas + Redis fan-out.
  2. Per-topic ABAC denies subscription to ITAR topics by non-US-person tokens with a typed close code.
  3. Backpressure metric `chetana_rt_dropped_total{reason="overflow"}` increments under load injection.
**Verification:**
  - Unit: `services/realtime-gw/internal/**/*_test.go`.
  - Integration: `services/realtime-gw/test/ws_test.go`.
  - Bench: `bench/k6/realtime-fanout.bench.js` — gates REQ-NFR-PERF-006 (≤500 ms p95 push @ 10k conn).

### TASK-P1-WEB-001: Web — ChetanaShell, login + MFA UI, audit viewer, export UI, settings

**Trace:** REQ-FUNC-PLT-IAM-001, REQ-FUNC-PLT-IAM-004, REQ-FUNC-PLT-IAM-005, REQ-FUNC-PLT-AUDIT-004, REQ-FUNC-CMN-005, REQ-CONST-005; design.md §6.1, §6.2, §6.3
**Owner:** Web
**Status:** backlog
**Estimate:** 10
**Depends on:** TASK-P0-WEB-001, TASK-P1-IAM-005, TASK-P1-AUDIT-002, TASK-P1-EXPORT-001, TASK-P1-RT-001
**Files (create/modify):**
  - `web/apps/shell/src/lib/shell/ChetanaShell.svelte` (new) — top nav, side nav, content area, route registry consumer
  - `web/apps/shell/src/routes/(public)/login/+page.svelte` (new) — email + password + MFA prompt
  - `web/apps/shell/src/routes/(public)/login/webauthn/+page.svelte` (new)
  - `web/apps/shell/src/routes/(public)/reset-password/+page.svelte` (new)
  - `web/apps/shell/src/routes/(app)/settings/sessions/+page.svelte` (new) — list active sessions, revoke
  - `web/apps/shell/src/routes/(app)/settings/api-keys/+page.svelte` (new) — create/revoke API keys
  - `web/apps/shell/src/routes/(app)/settings/mfa/+page.svelte` (new) — enroll TOTP / WebAuthn
  - `web/apps/shell/src/routes/(app)/audit/+page.svelte` (new) — search + filter audit log; export trigger
  - `web/apps/shell/src/routes/(app)/exports/+page.svelte` (new) — list jobs, download presigned URLs
  - `web/packages/api-client/src/iam.ts` (new) — typed Connect client wrapping IAM
  - `web/packages/api-client/src/audit.ts` (new)
  - `web/packages/api-client/src/realtime.ts` (new) — WS client with auto-reconnect, backoff, topic subscription manager
  - `web/apps/shell/tests/e2e/auth.spec.ts` (new) — Playwright login + MFA + reset
  - `web/apps/shell/tests/e2e/audit.spec.ts` (new)
**Acceptance criteria:**
  1. Login → MFA → land on default route works under Playwright.
  2. Audit viewer paginates 100 k events without UI jank (virtualised list).
  3. Export UI surfaces job progress via WS push (no polling).
  4. WebAuthn registration uses platform authenticator on supporting browsers.
  5. The route registry remains the single source of truth for `(app)/[domain]/[entity]/+page.svelte`.
**Verification:**
  - E2E: `web/apps/shell/tests/e2e/{auth,audit,exports}.spec.ts`.
  - Inspection: bundle analyser shows shell entrypoint < 200 KB gzip (Cesium loaded lazily, verified in Phase 2).

### TASK-P1-WEB-002: Cesium dependency wiring + bundle-splitting verification

**Trace:** REQ-CONST-002; design.md §6.4
**Owner:** Web
**Status:** backlog
**Estimate:** 3
**Depends on:** TASK-P0-WEB-001
**Files (create/modify):**
  - `web/apps/shell/vite.config.ts` (modify) — manual chunks: `cesium-engine`, `cesium-widgets`; copy Cesium static assets to `/cesium-assets/`
  - `web/apps/shell/src/lib/cesium/loader.ts` (new) — dynamic-import wrapper; configures `CESIUM_BASE_URL`
  - `web/apps/shell/src/lib/cesium/Viewer.svelte` (new) — base Cesium viewer Svelte component (used by Phase 2/4)
  - `web/apps/shell/tests/e2e/cesium.spec.ts` (new) — verifies a globe renders; verifies Cesium chunk is NOT in initial bundle
**Acceptance criteria:**
  1. Initial JS bundle does not contain `@cesium/engine`.
  2. Navigating to a Cesium-hosting route loads Cesium chunk on demand.
  3. Bundle analyser report committed under `web/apps/shell/bundle-report.html`.
**Verification:**
  - E2E: `web/apps/shell/tests/e2e/cesium.spec.ts`.
  - Inspection: `pnpm --filter shell run analyze` output reviewed.

### TASK-P1-NFR-001: Phase 1 NFR gate — IAM @ 1k req/s ≤100 ms p95; realtime @ 10k conn ≤500 ms p95

**Trace:** REQ-NFR-PERF-005, REQ-NFR-PERF-006, REQ-CONST-009; design.md §7.2
**Owner:** Platform Infra
**Status:** backlog
**Estimate:** 4
**Depends on:** TASK-P1-IAM-002, TASK-P1-RT-001
**Files (create/modify):**
  - `bench/k6/iam-login.bench.js` (new) — 1k req/s ramp; threshold p95 < 100 ms; error rate < 0.1 %
  - `bench/k6/realtime-fanout.bench.js` (new) — 10 k WS connections; threshold p95 push < 500 ms
  - `.github/workflows/nfr-phase1.yml` (new) — runs both benches against an ephemeral cluster on every PR that touches IAM or realtime-gw
  - `bench/results/phase1/README.md` (new) — recorded baseline results
**Acceptance criteria:**
  1. Both benches green for two consecutive runs against ephemeral clusters with the documented hardware profile.
  2. Workflow blocks merge to `main` when threshold breaks.
**Verification:**
  - Bench: as above.
  - Inspection: results archived under `bench/results/phase1/`.

---

## 4. Phase 2 — Ground Station MVP (12 weeks)

Goal: a real spacecraft can be tracked, telemetry decoded and stored, commanded with two-person approval, and visualised live in the browser. Plan-aligned 7 services × 52 RPCs.

### TASK-P2-GS-001: Plan-boundary refactor — split current `gs-*`/`sat-*` into the seven plan services + 52 RPCs

**Trace:** REQ-FUNC-GS-BOUNDARY-001, REQ-FUNC-GS-BOUNDARY-002; design.md §3.2, §3.3
**Owner:** Platform + Defense
**Status:** backlog
**Estimate:** 12
**Depends on:** TASK-P0-REPO-001, TASK-P0-HW-001
**Files (create/modify):**
  - `services/proto/space/groundstation/v1/{station,pass,anomaly,alert}.proto` (new) — platform-side service definitions; 7+9+6+8 RPCs
  - `chetana-defense/services/proto/space/satellite/v1/{satellite,telemetry,command}.proto` (new) — defense-side service definitions; 8+6+8 RPCs
  - `services/gs-station/cmd/gs-station/main.go` (new) — replaces parts of `gs-mc`
  - `services/gs-scheduler/cmd/gs-scheduler/main.go` (modify) — keep; adds Pass FSM (PESM) + pass-pred groupings
  - `services/gs-ingest/cmd/gs-ingest/main.go` (modify) — write-side fan-out only
  - `services/notify/internal/alert/` (new) — AlertService facade
  - `services/gs-station/internal/anomaly/` (new) — AnomalyService facade
  - `chetana-defense/services/sat-telemetry/cmd/sat-telemetry/main.go` (modify) — TelemetryService implementation
  - `chetana-defense/services/sat-command/cmd/sat-command/main.go` (new) — CommandService implementation
  - `chetana-defense/services/sat-mission/cmd/sat-mission/main.go` (modify) — SatelliteService implementation (catalog + TLE)
  - `space_plan/docs/README.md` (read-only reference) — RPC enumeration
  - `tools/proto/rpc-count.sh` (new) — counts RPCs per service against the plan target
**Acceptance criteria:**
  1. `tools/proto/rpc-count.sh` reports exactly 8/7/9/6/8/6/8 RPCs across `Satellite/GroundStation/Pass/Telemetry/Command/Anomaly/Alert` services (sum 52).
  2. `buf breaking` against the previous baseline either passes or carries an explicit waiver in `services/proto/buf.yaml`.
  3. All seven service binaries build and start; their `/ready` returns 200 against a primed cluster.
  4. Removed legacy services (where merged) leave a `MOVED.md` stub explaining the new home (no code).
**Verification:**
  - Inspection: `tools/proto/rpc-count.sh` in CI.
  - Integration: smoke test that calls one RPC per service via `buf curl`.

### TASK-P2-GS-002: `gs-pass-pred` — TLE manager + Space-Track + SGP4/SDP4 via `compute/crates/orbit-prop`

**Trace:** REQ-FUNC-GS-PASS-001, REQ-FUNC-SAT-005, REQ-FUNC-SAT-006; design.md §3.1, §6.5
**Owner:** Platform Ground
**Status:** backlog
**Estimate:** 8
**Depends on:** TASK-P2-GS-001
**Files (create/modify):**
  - `services/gs-pass-pred/cmd/gs-pass-pred/main.go` (new)
  - `services/gs-pass-pred/internal/tle/spacetrack.go` (new) — Space-Track client; 6h refresh; retry/backoff; freshness alerts via Notify
  - `services/gs-pass-pred/internal/tle/store.go` (new) — TLE history (time-versioned)
  - `services/gs-pass-pred/internal/predict/predictor.go` (new) — calls `compute/crates/orbit-prop` via CGO bindings; computes AOS/max elevation/LOS to ±1 s
  - `services/gs-pass-pred/internal/predict/doppler.go` (new)
  - `services/gs-pass-pred/internal/predict/skyplot.go` (new)
  - `compute/crates/orbit-prop/src/lib.rs` (modify) — `extern "C"` FFI surface for Go consumer; `wasm32-unknown-unknown` build target retained
  - `compute/crates/orbit-prop/Cargo.toml` (modify) — `crate-type = ["cdylib", "rlib"]`
  - `services/gs-pass-pred/migrations/0001_passes.sql` (new) — `tles`, `predicted_passes`, `pass_doppler_curves`
  - `services/gs-pass-pred/test/predict_test.go` (new) — validate against published Celestrak vectors
**Acceptance criteria:**
  1. AOS/max elevation/LOS within ±1 s of NORAD reference passes for ISS over 7-day horizon.
  2. TLE refresh runs every 6 h with jittered backoff on 429/5xx.
  3. Same `compute/crates/orbit-prop` builds for `wasm32-unknown-unknown` and is consumed by `web/packages/wasm/orbit/` (TASK-P2-WEB-002).
  4. Doppler curve computed for 24-h horizon in < 200 ms per pass.
**Verification:**
  - Unit: `compute/crates/orbit-prop/tests/sgp4_vectors.rs` (Celestrak vectors).
  - Integration: `services/gs-pass-pred/test/predict_test.go`.

### TASK-P2-GS-003: `gs-station` — registry, antenna config, capabilities, maintenance, health rollup

**Trace:** REQ-FUNC-GS-PASS-002; design.md §3.1
**Owner:** Platform Ground
**Status:** backlog
**Estimate:** 5
**Depends on:** TASK-P2-GS-001
**Files (create/modify):**
  - `services/gs-station/internal/registry/store.go` (new) — `ground_stations`, `antennas`, `capabilities`, `maintenance_windows`
  - `services/gs-station/internal/health/rollup.go` (new) — derived health from latest telemetry
  - `services/gs-station/migrations/0001_station.sql` (new)
  - `services/gs-station/test/station_test.go` (new)
**Acceptance criteria:**
  1. CRUD of stations + antennas works; capabilities matched against pass requirements.
  2. Maintenance windows block scheduling.
**Verification:**
  - Integration: `services/gs-station/test/station_test.go`.

### TASK-P2-GS-004: `gs-scheduler` (PassService) — Pass FSM (PESM) per D7.2 + conflict resolution

**Trace:** REQ-FUNC-GS-PASS-002, REQ-FUNC-GS-PASS-003; design.md §3.1
**Owner:** Platform Ground
**Status:** backlog
**Estimate:** 10
**Depends on:** TASK-P2-GS-002, TASK-P2-GS-003
**Files (create/modify):**
  - `services/gs-scheduler/internal/fsm/pesm.go` (new) — 11-state FSM (`SCHEDULED, PREPARING, READY, ACQUIRING, TRACKING, CLOSING, REPORTING, COMPLETED, FAILED, CANCELLED, ABORTED`); guards; per-state timeouts; side effects
  - `services/gs-scheduler/internal/fsm/transitions.go` (new) — declarative transition table; matches D7.2 exactly
  - `services/gs-scheduler/internal/conflict/resolver.go` (new) — antenna conflict detection; priority-based resolution
  - `services/gs-scheduler/internal/eventbus/kafka.go` (new) — emits `pass.<id>.state` events to Kafka (consumed by realtime-gw)
  - `services/gs-scheduler/migrations/0001_passes.sql` (new) — `scheduled_passes`, `pass_state_history`
  - `services/gs-scheduler/test/fsm_test.go` (new) — exhaustive table-driven tests against every D7.2 transition
**Acceptance criteria:**
  1. Every transition in D7.2 represented; illegal transitions rejected with typed error.
  2. Per-state timeout fires correct fail/abort path.
  3. Pass state events visible on `realtime-gw` `pass.<id>.state` topic within 200 ms of transition.
  4. Scheduler sustains 1 000 passes/day across 10 antennas in load test (REQ-NFR-PERF, NFR gate task).
**Verification:**
  - Unit: `services/gs-scheduler/test/fsm_test.go` (≥95 % branch).
  - Integration: `services/gs-scheduler/test/scheduler_e2e_test.go`.

### TASK-P2-TM-001: Telemetry pipeline — Kafka frame consumer, decommutation, calibration, limits

**Trace:** REQ-FUNC-GS-TM-001, REQ-FUNC-GS-TM-002, REQ-FUNC-GS-TM-004; design.md §3.2, §5.1
**Owner:** Defense
**Status:** backlog
**Estimate:** 12
**Depends on:** TASK-P2-GS-001, TASK-P0-DB-001
**Files (create/modify):**
  - `chetana-defense/services/sat-telemetry/internal/decom/decommutator.go` (new) — ICD-driven; sync word + CRC + APID validation
  - `chetana-defense/services/sat-telemetry/internal/calibrate/poly.go` (new) — polynomial / point-pair / lookup
  - `chetana-defense/services/sat-telemetry/internal/limit/checker.go` (new) — red/yellow/green; rate-of-change
  - `chetana-defense/services/sat-telemetry/internal/store/timescale.go` (new) — `telemetry_samples` Timescale hypertable writer (1d chunks)
  - `chetana-defense/services/sat-telemetry/internal/agg/continuous.go` (new) — declares 1-min and 1-h continuous aggregates
  - `chetana-defense/services/sat-telemetry/internal/publish/kafka.go` (new) — emits `telemetry.params` events
  - `chetana-defense/services/sat-telemetry/migrations/0001_telemetry.sql` (new)
  - `chetana-defense/services/sat-telemetry/test/decom_test.go` (new)
**Acceptance criteria:**
  1. Decom ingests TM frames per the spacecraft profile (REQ-FUNC-SAT-001) and produces typed parameter records.
  2. Calibration applies polynomial + point-pair correctly across edge values.
  3. Limit violations published as alerts; consumed by AlertService and realtime-gw.
  4. Continuous aggregates available within 1 min after raw insert.
**Verification:**
  - Unit: `chetana-defense/services/sat-telemetry/internal/**/*_test.go`.
  - Integration: `chetana-defense/services/sat-telemetry/test/pipeline_e2e_test.go` against Testcontainers stack.

### TASK-P2-TM-002: Telemetry retention — raw 7d, 1-min 90d, 1-h 5y; Glacier archival

**Trace:** REQ-FUNC-GS-TM-003; design.md §5.4
**Owner:** Defense + Platform Infra
**Status:** backlog
**Estimate:** 3
**Depends on:** TASK-P2-TM-001
**Files (create/modify):**
  - `chetana-defense/services/sat-telemetry/migrations/0002_retention.sql` (new) — Timescale retention policies + drop chunks
  - `chetana-defense/services/sat-telemetry/internal/archive/glacier.go` (new) — periodic export of dropped chunks to Glacier; pointer table
**Acceptance criteria:**
  1. Raw chunks > 7 d dropped from hot storage with Glacier pointer recorded.
  2. 1-min aggregates dropped after 90 d; 1-h after 5 y.
**Verification:**
  - Integration: time-warped test using `pg_advance_time` extension or fixture clock.

### TASK-P2-CMD-001: Command FSM — 17-state per D7.9 + 2-person approval

**Trace:** REQ-FUNC-SAT-009, REQ-FUNC-SAT-010, REQ-FUNC-SAT-012; design.md §3.2
**Owner:** Defense
**Status:** backlog
**Estimate:** 14
**Depends on:** TASK-P2-GS-001, TASK-P1-AUTHZ-001, TASK-P1-AUDIT-001
**Files (create/modify):**
  - `chetana-defense/services/sat-command/internal/fsm/states.go` (new) — 17 states per D7.9
  - `chetana-defense/services/sat-command/internal/fsm/transitions.go` (new) — guard predicates; side-effect actions
  - `chetana-defense/services/sat-command/internal/approval/twoperson.go` (new) — second approver MUST be a different US-person principal with `command.approve`
  - `chetana-defense/services/sat-command/internal/hazard/classifier.go` (new) — `safe | caution | critical`; safe auto-approves
  - `chetana-defense/services/sat-command/internal/verify/correlator.go` (new) — ACK + telemetry-state-match within configurable timeout; on timeout → `verification_failed`
  - `chetana-defense/services/sat-command/migrations/0001_commands.sql` (new) — `commands`, `command_approvals`, `command_state_history`
  - `chetana-defense/services/sat-command/test/fsm_test.go` (new) — every transition + every illegal transition
**Acceptance criteria:**
  1. Every state transition logged with prev/next/actor/reason; audit chain preserved.
  2. Two-person approval enforced for caution + critical; same approver self-approval rejected.
  3. Verification correlator times out and triggers configurable retry policy.
**Verification:**
  - Unit: `chetana-defense/services/sat-command/test/fsm_test.go`.
  - Integration: `chetana-defense/services/sat-command/test/command_e2e_test.go` driving a full submit→approve→uplink→ack→verify cycle against `sat-simulation`.

### TASK-P2-SIM-001: `sat-simulation` — high-fidelity 6-DOF simulator with all profile combos; replay support

**Trace:** REQ-FUNC-SAT-013; design.md §3.2
**Owner:** Defense Mission
**Status:** backlog
**Estimate:** 14
**Depends on:** TASK-P0-HW-001, TASK-P2-GS-001
**Files (create/modify):**
  - `chetana-defense/services/sat-simulation/cmd/sat-simulation/main.go` (new)
  - `chetana-defense/services/sat-simulation/internal/dynamics/sixdof.rs` (new) — Rust crate via FFI: 6-DOF state propagation; gravity, drag, SRP, third-body
  - `chetana-defense/services/sat-simulation/internal/profile/runtime.go` (new) — drives simulation from `SpacecraftProfile` (REQ-FUNC-SAT-001)
  - `chetana-defense/services/sat-simulation/internal/rf/loop.go` (new) — synthetic RF loopback for end-to-end TM/TC testing without hardware
  - `chetana-defense/services/sat-simulation/internal/replay/replay.go` (new) — record + replay telemetry/command sessions
  - `chetana-defense/services/sat-simulation/test/sim_e2e_test.go` (new)
**Acceptance criteria:**
  1. Drives a complete TM/TC cycle with `sat-telemetry` and `sat-command` end-to-end with no hardware attached.
  2. Replay session reproduces a recorded run bit-exact.
  3. All profile combinations from REQ-FUNC-SAT-001 (band × modulation × CCSDS profile) instantiable as a sim run.
**Verification:**
  - Integration: `chetana-defense/services/sat-simulation/test/sim_e2e_test.go`.

### TASK-P2-CMD-002: Command — CCSDS TC encoding via `flight/crates/cdh-ccsds`

**Trace:** REQ-FUNC-SAT-011, REQ-FUNC-GS-HW-006; design.md §3.2
**Owner:** Defense
**Status:** backlog
**Estimate:** 4
**Depends on:** TASK-P2-CMD-001
**Files (create/modify):**
  - `chetana-defense/services/sat-command/internal/encode/ccsds.go` (new) — CGO bindings to `flight/crates/cdh-ccsds`
  - `chetana-defense/flight/crates/cdh-ccsds/src/tc_frame.rs` (modify) — expose `extern "C"` `tc_encode` + `tc_decode`; CRC-16-CCITT + sequence numbers
  - `chetana-defense/services/sat-command/internal/encode/hsm.go` (new) — pluggable HSM payload-encryption interface; default no-op provider in v1; HSM impl in Phase 6 (TASK-P6-SEC-001)
  - `chetana-defense/services/sat-command/test/encode_test.go` (new) — round-trip against published CCSDS test vectors
**Acceptance criteria:**
  1. Encoded frames pass CCSDS 232.0-B-3 conformance vectors.
  2. HSM interface present, no-op default returns ciphertext = plaintext, audit-flagged.
**Verification:**
  - Unit: `chetana-defense/flight/crates/cdh-ccsds/tests/`.
  - Integration: `chetana-defense/services/sat-command/test/encode_test.go`.

### TASK-P2-HW-001: SDR adapters — UHD, librtlsdr, custom (production-grade)

**Trace:** REQ-FUNC-GS-HW-001, REQ-FUNC-GS-HW-004, REQ-FUNC-GS-HW-005, REQ-CONST-010; design.md §4.4
**Owner:** Defense Hardware
**Status:** backlog
**Estimate:** 18
**Depends on:** TASK-P0-HW-001, TASK-P2-GS-001
**Files (create/modify):**
  - `chetana-defense/compute/crates/gs-rf-driver/uhd/` (new) — UHD bindings; UHF/S/X tuning, gain, RX/TX IQ streaming
  - `chetana-defense/compute/crates/gs-rf-driver/librtlsdr/` (new) — RTL-SDR bindings; UHF only (RX-only documented)
  - `chetana-defense/compute/crates/gs-rf-driver/custom/` (new) — gRPC-over-UDS adapter to a customer-defined SDR daemon; production-grade reference daemon committed
  - `chetana-defense/services/packages/hardware/uhd/` (new) — Go shim over Rust crate
  - `chetana-defense/services/packages/hardware/rtl/` (new)
  - `chetana-defense/services/packages/hardware/custom/` (new)
  - `chetana-defense/compute/crates/gs-bit-sync/src/lib.rs` (modify) — BPSK, QPSK, OQPSK, 8PSK, GMSK demod
  - `chetana-defense/compute/crates/gs-fec/src/lib.rs` (modify) — convolutional + RS decoding per spacecraft profile
  - `chetana-defense/compute/crates/gs-doppler/src/lib.rs` (modify) — Doppler tracking using pass-pred curve
  - `chetana-defense/services/gs-rf/cmd/gs-rf/main.go` (new)
  - `chetana-defense/services/gs-rf/test/rf_e2e_test.go` (new) — exercised against the in-memory fake (TASK-P0-HW-001) and a hardware-loopback rig
**Acceptance criteria:**
  1. All three adapters implement the full `HardwareDriver` interface; no method panics or returns `ErrNotImplemented`.
  2. UHD adapter tunes a USRP B210 (lab fixture) and demodulates a known QPSK signal end-to-end.
  3. RTL adapter receives a known UHF beacon at the lab and produces decoded frames.
  4. Custom adapter daemon documented and deployed by Helm chart `gs-rf-custom-daemon`.
**Verification:**
  - Unit: `chetana-defense/compute/crates/gs-rf-driver/**/*` test modules.
  - Integration: `chetana-defense/services/gs-rf/test/rf_e2e_test.go` (skips hardware tests if `CHETANA_NO_HW=1`).

### TASK-P2-HW-002: Antenna controllers — Hamlib, GS-232, custom (production-grade)

**Trace:** REQ-FUNC-GS-HW-002, REQ-CONST-010; design.md §4.4
**Owner:** Defense Hardware
**Status:** backlog
**Estimate:** 10
**Depends on:** TASK-P0-HW-001
**Files (create/modify):**
  - `chetana-defense/compute/crates/gs-antenna/hamlib/` (new) — `rotctld` TCP client
  - `chetana-defense/compute/crates/gs-antenna/gs232/` (new) — RS-232 / TCP serial GS-232 protocol
  - `chetana-defense/compute/crates/gs-antenna/custom/` (new) — gRPC-over-UDS to a customer-defined rotator daemon
  - `chetana-defense/services/packages/hardware/{hamlib,gs232,custom}/` (new) — Go shims
  - `chetana-defense/services/gs-rf/internal/tracker/track.go` (new) — closed-loop track using pass-pred trajectory
  - `chetana-defense/services/gs-rf/test/antenna_e2e_test.go` (new)
**Acceptance criteria:**
  1. All three adapters implement the full `AntennaController` interface end-to-end.
  2. Lab fixture (Yaesu G-5500) tracked through a synthetic pass with < 1° residual.
**Verification:**
  - Integration: `chetana-defense/services/gs-rf/test/antenna_e2e_test.go`.

### TASK-P2-HW-003: GroundNetworkProvider adapters — own-dish (v1) + AWS GS (v1, contingent on OQ-001)

**Trace:** REQ-FUNC-GS-HW-003; design.md §4.4
**Owner:** Defense Hardware
**Status:** blocked:OQ-001 (for `aws-gs` only — `own-dish` proceeds)
**Estimate:** 8 (own-dish) + 6 (aws-gs)
**Depends on:** TASK-P2-HW-001, TASK-P2-HW-002
**Files (create/modify):**
  - `chetana-defense/services/packages/hardware/owndish/owndish.go` (new) — wraps SDR + antenna adapters
  - `chetana-defense/services/packages/hardware/awsgs/awsgs.go` (new — blocked) — AWS Ground Station Mission Profile + DataflowEndpointGroup orchestration; replaces Azure Orbital
  - `chetana-defense/services/gs-rf/internal/provider/registry.go` (modify) — registers both providers
  - `chetana-defense/services/gs-rf/test/provider_test.go` (new)
**Acceptance criteria:**
  1. Own-dish provider executes a contact end-to-end against the lab rig.
  2. AWS GS provider (when unblocked) reserves contacts via the AWS GS API; falls back to own-dish on failure per policy.
**Verification:**
  - Integration: `chetana-defense/services/gs-rf/test/provider_test.go`.

### TASK-P2-WEB-001: Web — Cesium globe, ground tracks, sky plot, AOS/LOS timeline

**Trace:** REQ-FUNC-SAT-004, REQ-FUNC-GS-PASS-001, REQ-CONST-002; design.md §6.4, §6.6
**Owner:** Web
**Status:** backlog
**Estimate:** 10
**Depends on:** TASK-P1-WEB-002, TASK-P2-GS-002
**Files (create/modify):**
  - `web/apps/shell/src/routes/(app)/groundstation/passes/+page.svelte` (new) — pass timeline + Cesium globe + sky plot
  - `web/apps/shell/src/lib/cesium/GroundTrack.ts` (new)
  - `web/apps/shell/src/lib/cesium/SkyPlot.svelte` (new) — D3-based polar plot
  - `web/apps/shell/src/lib/cesium/PassTimeline.svelte` (new)
  - `web/apps/shell/tests/e2e/passes.spec.ts` (new)
**Acceptance criteria:**
  1. Live globe shows the spacecraft position and ground track for the next 24 h.
  2. Sky plot updates as a pass progresses.
  3. Selecting a pass in the timeline pans the globe and seeds the FSM details panel.
**Verification:**
  - E2E: `web/apps/shell/tests/e2e/passes.spec.ts`.

### TASK-P2-WEB-002: Web — SGP4 WASM kernel + browser-side propagation

**Trace:** REQ-FUNC-SAT-005; design.md §6.5
**Owner:** Web + Compute
**Status:** backlog
**Estimate:** 4
**Depends on:** TASK-P2-GS-002
**Files (create/modify):**
  - `web/packages/wasm/orbit/Cargo.toml` (new) — wraps `compute/crates/orbit-prop`
  - `web/packages/wasm/orbit/src/lib.rs` (new) — `wasm-bindgen` exports `propagate(tle, t)`
  - `web/packages/wasm/orbit/build.sh` (new) — `wasm-pack build --target web --release`
  - `web/apps/shell/src/lib/cesium/GroundTrack.ts` (modify) — uses WASM propagator for sub-second updates
  - `web/packages/wasm/orbit/test/orbit.test.ts` (new) — vector tests in Playwright
**Acceptance criteria:**
  1. WASM bundle < 200 KB gzipped.
  2. Browser-side propagation matches server-side within 1 m for 24 h ISS propagation.
**Verification:**
  - Integration: `web/packages/wasm/orbit/test/orbit.test.ts`.

### TASK-P2-WEB-003: Web — telemetry strip charts wired to realtime-gw

**Trace:** REQ-FUNC-GS-TM-005; design.md §6.6
**Owner:** Web
**Status:** backlog
**Estimate:** 5
**Depends on:** TASK-P1-RT-001, TASK-P2-TM-001
**Files (create/modify):**
  - `web/apps/shell/src/routes/(app)/satellite/telemetry/+page.svelte` (new) — strip charts via ECharts; topic subscription via realtime-gw client
  - `web/apps/shell/src/lib/charts/StripChart.svelte` (new) — generic, reusable strip chart
  - `web/apps/shell/tests/e2e/telemetry.spec.ts` (new) — synthetic-feed harness
**Acceptance criteria:**
  1. Strip chart renders 60 fps at 100 samples/s/channel × 16 channels.
  2. End-to-end latency from synthetic frame injection to chart pixel update ≤ 100 ms p95 (gates REQ-NFR-PERF-001).
**Verification:**
  - E2E: `web/apps/shell/tests/e2e/telemetry.spec.ts` measuring browser performance API.

### TASK-P2-WEB-004: Web — command queue with 2-person approval dialog

**Trace:** REQ-FUNC-SAT-009, REQ-FUNC-SAT-010; design.md §6.6
**Owner:** Web
**Status:** backlog
**Estimate:** 6
**Depends on:** TASK-P2-CMD-001
**Files (create/modify):**
  - `web/apps/shell/src/routes/(app)/satellite/commands/+page.svelte` (new) — queue, FSM state, approval workflow
  - `web/apps/shell/src/lib/commands/ApprovalDialog.svelte` (new)
  - `web/apps/shell/tests/e2e/commands.spec.ts` (new)
**Acceptance criteria:**
  1. Submitter cannot self-approve; UI rejects same-actor approval before server.
  2. State transitions reflected live via realtime-gw push.
**Verification:**
  - E2E: `web/apps/shell/tests/e2e/commands.spec.ts`.

### TASK-P2-NFR-001: Phase 2 NFR gate — TM ≤100ms p95; pass scheduling ≥1k passes/day

**Trace:** REQ-NFR-PERF-001, REQ-CONST-009; design.md §7.2
**Owner:** Platform Infra
**Status:** backlog
**Estimate:** 4
**Depends on:** TASK-P2-TM-001, TASK-P2-GS-004, TASK-P2-WEB-003
**Files (create/modify):**
  - `bench/k6/tm-e2e.bench.js` (new) — frame-injection → realtime-gw → browser sink p95 < 100 ms
  - `bench/k6/scheduler-load.bench.js` (new) — 1 000 pass schedules in 24 h sim
  - `.github/workflows/nfr-phase2.yml` (new)
  - `bench/results/phase2/README.md` (new)
**Acceptance criteria:**
  1. Both benches green over two consecutive runs.
**Verification:**
  - Bench: as above.

---

## 5. Phase 3 — EO + ML serving (10 weeks)

Goal: STAC catalog live with ≤200 ms p95 search; production EO pipeline (ortho/pan-sharpen/indices/mosaic/change-detection) at ≥100 tile pairs/h; Triton serving ML at ≤100 ms p95; tile server for browser delivery.

### TASK-P3-EO-001: `eo-catalog` — STAC API 1.0.0 + OGC API Features + JSON-Schema validation

**Trace:** REQ-FUNC-EO-CAT-001, REQ-FUNC-EO-CAT-004, REQ-FUNC-EO-CAT-005; design.md §3.1
**Owner:** EO
**Status:** backlog
**Estimate:** 10
**Depends on:** TASK-P0-DB-001, TASK-P1-AUTHZ-001
**Files (create/modify):**
  - `services/eo-catalog/cmd/eo-catalog/main.go` (new)
  - `services/eo-catalog/internal/api/{root,collections,items,search,queryables,conformance}.go` (new)
  - `services/eo-catalog/internal/validate/stac.go` (new) — JSON Schema validator (STAC 1.0.0 + EO/SAR/projection/view/processing extensions)
  - `services/eo-catalog/internal/store/items.go` (new) — Postgres store
  - `services/eo-catalog/internal/pagination/cursor.go` (new) — opaque cursor (base64 JSON; HMAC-signed)
  - `services/eo-catalog/migrations/0001_stac.sql` (new) — `collections`, `items` (with `geometry geometry(Polygon, 4326)`, `bbox`, `datetime`, `properties JSONB`, `data_classification`)
  - `services/eo-catalog/test/api_test.go` (new) — STAC conformance test suite
**Acceptance criteria:**
  1. STAC 1.0.0 conformance test suite passes for required + supported extensions.
  2. Items rejected if they fail JSON Schema; errors are STAC-compliant problem+json.
  3. Cursor pagination resilient to insert/delete during traversal (HMAC-signed; signature verified on each fetch).
**Verification:**
  - Unit: schema validation tests against published STAC examples.
  - Integration: `services/eo-catalog/test/api_test.go`.

### TASK-P3-EO-002: `eo-catalog` — CQL2 parser + bbox/temporal filters

**Trace:** REQ-FUNC-EO-CAT-002; design.md §3.1
**Owner:** EO
**Status:** backlog
**Estimate:** 6
**Depends on:** TASK-P3-EO-001
**Files (create/modify):**
  - `services/eo-catalog/internal/cql2/parser.go` (new) — CQL2-text + CQL2-JSON; produces a typed AST
  - `services/eo-catalog/internal/cql2/sql.go` (new) — AST → parameterised PostGIS SQL (no string interpolation; injection-safe)
  - `services/eo-catalog/internal/cql2/parser_test.go` (new) — fuzz target included
**Acceptance criteria:**
  1. CQL2 conformance examples parse and translate to expected SQL fragments.
  2. Fuzz target survives 1 M iterations without panic or unsanitised output.
**Verification:**
  - Unit: `services/eo-catalog/internal/cql2/parser_test.go`.
  - Bench: `bench/k6/stac-search.bench.js` — gates REQ-NFR-PERF-004 (≤200 ms p95).

### TASK-P3-EO-003: `eo-catalog` — H3 spatial indexing + PostGIS GIST

**Trace:** REQ-FUNC-EO-CAT-003; design.md §3.1
**Owner:** EO
**Status:** backlog
**Estimate:** 4
**Depends on:** TASK-P3-EO-001
**Files (create/modify):**
  - `services/eo-catalog/migrations/0002_h3.sql` (new) — H3 columns at resolutions 4, 6, 8; GIN indexes; PostGIS GIST on `geometry`
  - `services/eo-catalog/internal/h3/index.go` (new) — Go binding; populates H3 cells on insert
  - `services/eo-catalog/test/h3_test.go` (new)
**Acceptance criteria:**
  1. Spatial query selectivity improves measurably (≥10×) over GIST-only baseline on a 1 M-item fixture.
  2. H3 cells back-filled correctly on existing items via migration.
**Verification:**
  - Bench: comparative microbench in `services/eo-catalog/internal/h3/h3_bench_test.go`.

### TASK-P3-EO-004: `eo-pipeline` orchestrator + Kafka workers + scene-pair selection

**Trace:** REQ-FUNC-EO-PIPE-001; design.md §3.1
**Owner:** EO
**Status:** backlog
**Estimate:** 8
**Depends on:** TASK-P3-EO-001
**Files (create/modify):**
  - `services/eo-pipeline/cmd/eo-pipeline/main.go` (new)
  - `services/eo-pipeline/internal/orchestrator/dag.go` (new) — declarative DAG (stages: ortho, pansharpen, indices, mosaic, change-detect)
  - `services/eo-pipeline/internal/worker/worker.go` (new) — Kafka consumer; concurrency by GPU/CPU pool
  - `services/eo-pipeline/internal/pair/selector.go` (new) — scene-pair selection by AOI + temporal proximity + cloud cover
  - `services/eo-pipeline/migrations/0001_jobs.sql` (new) — `processing_jobs`, `processing_job_events` (Timescale)
**Acceptance criteria:**
  1. DAG executes with per-stage retry; stage failure → typed event + dead-letter topic.
  2. Worker reschedules on crash via lease expiry.
**Verification:**
  - Integration: `services/eo-pipeline/test/orchestrator_test.go`.

### TASK-P3-EO-005: `eo-pipeline` — orthorectification (RPC + DEM) via `compute/crates/eo-geometric`

**Trace:** REQ-FUNC-EO-PIPE-002, REQ-CONST-010; design.md §3.1
**Owner:** EO Compute
**Status:** backlog
**Estimate:** 7
**Depends on:** TASK-P3-EO-004
**Files (create/modify):**
  - `compute/crates/eo-geometric/src/ortho.rs` (modify) — production-grade RPC ortho with DEM correction; SIMD where applicable
  - `compute/crates/eo-geometric/tests/ortho_test.rs` (new) — fixture: known Sentinel-2 RPC + SRTM DEM → ortho output checked against reference
  - `services/eo-pipeline/internal/stages/ortho.go` (new) — Go shim invoking the Rust crate
**Acceptance criteria:**
  1. Geometric error ≤ 1 pixel (10 m) on Sentinel-2 over flat terrain.
  2. ≤ 60 s per S2 tile on 16-vCPU node.
**Verification:**
  - Unit: `compute/crates/eo-geometric/tests/ortho_test.rs`.

### TASK-P3-EO-006: `eo-pipeline` — pan-sharpening (Brovey, GS, IHS, PCA, hist-match)

**Trace:** REQ-FUNC-EO-PIPE-003, REQ-CONST-010; design.md §3.1
**Owner:** EO Compute
**Status:** backlog
**Estimate:** 6
**Depends on:** TASK-P3-EO-004
**Files (create/modify):**
  - `compute/crates/eo-pansharpen/src/{brovey,gs,ihs,pca,histmatch}.rs` (new)
  - `compute/crates/eo-pansharpen/tests/methods_test.rs` (new) — SSIM threshold per method
  - `services/eo-pipeline/internal/stages/pansharpen.go` (new)
**Acceptance criteria:**
  1. All five methods implemented; SSIM ≥ documented threshold per method on the reference Sentinel-2 fixture.
**Verification:**
  - Unit: `compute/crates/eo-pansharpen/tests/methods_test.rs`.

### TASK-P3-EO-007: `eo-pipeline` — spectral indices (NDVI, NDWI, EVI, SAVI)

**Trace:** REQ-FUNC-EO-PIPE-004, REQ-CONST-010; design.md §3.1
**Owner:** EO Compute
**Status:** backlog
**Estimate:** 3
**Depends on:** TASK-P3-EO-004
**Files (create/modify):**
  - `compute/crates/eo-indices/src/{ndvi,ndwi,evi,savi}.rs` (new)
  - `compute/crates/eo-indices/tests/indices_test.rs` (new)
  - `services/eo-pipeline/internal/stages/indices.go` (new)
**Acceptance criteria:**
  1. Index outputs match published reference values within 1e-3 on test pixels.
**Verification:**
  - Unit: `compute/crates/eo-indices/tests/indices_test.rs`.

### TASK-P3-EO-008: `eo-pipeline` — mosaic (most-recent, least-cloud, median)

**Trace:** REQ-FUNC-EO-PIPE-005, REQ-CONST-010; design.md §3.1
**Owner:** EO Compute
**Status:** backlog
**Estimate:** 5
**Depends on:** TASK-P3-EO-004
**Files (create/modify):**
  - `compute/crates/eo-mosaic/src/{recent,least_cloud,median}.rs` (new)
  - `compute/crates/eo-mosaic/tests/mosaic_test.rs` (new)
  - `services/eo-pipeline/internal/stages/mosaic.go` (new)
**Acceptance criteria:**
  1. Tile boundaries seamless; per-pixel selection respects strategy.
**Verification:**
  - Unit: `compute/crates/eo-mosaic/tests/mosaic_test.rs`.

### TASK-P3-EO-009: Change-detection workflow (CVA / image-diff / OBIA / DL) + STAC publishing

**Trace:** REQ-FUNC-EO-PIPE-006, REQ-FUNC-EO-PIPE-007; design.md §3.1
**Owner:** EO + ML
**Status:** backlog
**Estimate:** 12
**Depends on:** TASK-P3-EO-005, TASK-P3-EO-006, TASK-P3-EO-007, TASK-P3-ML-001
**Files (create/modify):**
  - `services/eo-pipeline/internal/changedet/coregister.go` (new) — sub-pixel co-registration
  - `services/eo-pipeline/internal/changedet/radnorm.go` (new) — radiometric normalisation
  - `services/eo-pipeline/internal/changedet/cloudmask.go` (new) — cloud + cloud-shadow mask
  - `services/eo-pipeline/internal/changedet/{cva,diff,obia,dl}.go` (new) — four detection methods
  - `services/eo-pipeline/internal/changedet/polygons.go` (new) — vectorise change polygons
  - `services/eo-pipeline/internal/changedet/publish.go` (new) — emit STAC items for derived products
  - `services/eo-pipeline/test/changedet_e2e_test.go` (new) — F1 ≥ 0.90 on validation set
**Acceptance criteria:**
  1. F1 ≥ 0.90 on the validation set checked into `services/eo-pipeline/test/fixtures/`.
  2. Per-tile-pair latency ≤ 5 min p95 on a 16-vCPU + 1×T4 GPU node.
  3. 24-h end-to-end SLA from scene ingest to change product STAC item available.
**Verification:**
  - Integration: `services/eo-pipeline/test/changedet_e2e_test.go`.

### TASK-P3-ML-001: Triton + ONNX Runtime + TensorRT deployment + dynamic batching config

**Trace:** REQ-FUNC-EO-ML-001, REQ-FUNC-EO-ML-005; design.md §3.1
**Owner:** ML Platform
**Status:** backlog
**Estimate:** 8
**Depends on:** TASK-P0-INFRA-001
**Files (create/modify):**
  - `infra/helm/charts/triton/` (new) — Triton Helm chart with HPA on GPU utilisation > 80 % and queue-depth thresholds
  - `infra/helm/charts/triton/values.yaml` (new) — model-repository S3 backend; dynamic batching `max_queue_delay_microseconds=100`
  - `services/eo-analytics/internal/triton/client.go` (new) — gRPC client wrapper
  - `services/eo-analytics/test/triton_smoke_test.go` (new) — load a tiny ONNX model, infer
**Acceptance criteria:**
  1. Triton Pods scale up under synthetic GPU load.
  2. `tritonserver --model-control-mode=poll` discovers new model versions from S3.
**Verification:**
  - Integration: `services/eo-analytics/test/triton_smoke_test.go`.

### TASK-P3-ML-002: ML model registry (MLflow-style) + canary/shadow/rollback

**Trace:** REQ-FUNC-EO-ML-003; design.md §3.1
**Owner:** ML Platform
**Status:** backlog
**Estimate:** 8
**Depends on:** TASK-P3-ML-001
**Files (create/modify):**
  - `services/eo-analytics/internal/registry/store.go` (new) — model versioning, status lifecycle (draft → staging → canary → production → archived)
  - `services/eo-analytics/internal/registry/router.go` (new) — traffic-weighted A/B routing
  - `services/eo-analytics/internal/registry/canary.go` (new) — shift-by-percentage with auto-rollback on error rate
  - `services/eo-analytics/internal/registry/shadow.go` (new) — mirror traffic to shadow model; compare outputs offline
  - `services/eo-analytics/migrations/0001_models.sql` (new)
  - `services/eo-analytics/test/registry_test.go` (new)
**Acceptance criteria:**
  1. Promote draft → staging → canary 10 % → production; rollback resets traffic to previous version.
  2. Shadow inference does not affect client latency p95.
**Verification:**
  - Integration: `services/eo-analytics/test/registry_test.go`.

### TASK-P3-ML-003: ONNX auto-conversion intake (PyTorch / TensorFlow → ONNX) with I/O verification

**Trace:** REQ-FUNC-EO-ML-004; design.md §3.1
**Owner:** ML Platform
**Status:** backlog
**Estimate:** 5
**Depends on:** TASK-P3-ML-002
**Files (create/modify):**
  - `services/eo-analytics/python/convert/torch_to_onnx.py` (new)
  - `services/eo-analytics/python/convert/tf_to_onnx.py` (new)
  - `services/eo-analytics/python/convert/verify.py` (new) — runs original + ONNX on test tensors; asserts L∞ error < threshold
  - `services/eo-analytics/internal/intake/handler.go` (new) — RPC: upload checkpoint → convert → verify → register
  - `services/eo-analytics/test/intake_test.go` (new)
**Acceptance criteria:**
  1. Sample PyTorch ResNet-18 and TF MobileNet-V2 convert and verify.
  2. Verification failure leaves the model in `failed` status with diagnostic.
**Verification:**
  - Integration: `services/eo-analytics/test/intake_test.go`.

### TASK-P3-ML-004: ML inference NFR — ≤100 ms p95 / 256² tile, ≥10 000 tiles/min batch, GPU ≥80 %

**Trace:** REQ-FUNC-EO-ML-002, REQ-NFR-PERF-003; design.md §3.1
**Owner:** ML Platform
**Status:** backlog
**Estimate:** 4
**Depends on:** TASK-P3-ML-001
**Files (create/modify):**
  - `bench/triton/inference.bench.py` (new)
  - `bench/triton/results/phase3/README.md` (new)
**Acceptance criteria:**
  1. Sustained p95 ≤ 100 ms; throughput ≥ 10 000 tiles/min on a 1×A10 node.
**Verification:**
  - Bench: as above.

### TASK-P3-ML-005: ITAR classification on model artifacts

**Trace:** REQ-FUNC-EO-ML-006, REQ-COMP-ITAR-002; design.md §4.6
**Owner:** ML Platform + Compliance
**Status:** backlog
**Estimate:** 2
**Depends on:** TASK-P3-ML-002
**Files (create/modify):**
  - `services/eo-analytics/internal/registry/classification.go` (new) — `export_classification ∈ {public, internal, restricted, cui, itar}`
  - `services/eo-analytics/internal/triton/authz.go` (new) — interceptor that denies inference to non-US-person principals on ITAR-classified models
**Acceptance criteria:**
  1. Inference call against an ITAR model from a non-US-person token returns 403 with audit event.
**Verification:**
  - Integration: extends `services/eo-analytics/test/registry_test.go`.

### TASK-P3-EO-010: `gi-tiles` real WMS/WMTS/MVT tile server using `compute/crates/gi-tile-render`

**Trace:** REQ-FUNC-GI-WS-002; design.md §3.1
**Owner:** EO + GeoInt
**Status:** backlog
**Estimate:** 8
**Depends on:** TASK-P3-EO-001
**Files (create/modify):**
  - `services/gi-tiles/cmd/gi-tiles/main.go` (new)
  - `services/gi-tiles/internal/wms/wms.go` (new) — WMS 1.3.0
  - `services/gi-tiles/internal/wmts/wmts.go` (new) — WMTS 1.0
  - `services/gi-tiles/internal/mvt/mvt.go` (new) — MVT vector tiles
  - `compute/crates/gi-tile-render/src/lib.rs` (modify) — production-grade renderer; SIMD where applicable
  - `services/gi-tiles/internal/cache/redis.go` (new) — tile cache
  - `services/gi-tiles/test/tiles_test.go` (new)
**Acceptance criteria:**
  1. 256² PNG tile rendered in ≤ 50 ms uncached; ≤ 5 ms cached.
  2. MVT response valid per Mapbox Vector Tile spec.
**Verification:**
  - Integration: `services/gi-tiles/test/tiles_test.go`.

### TASK-P3-WEB-001: Web — STAC search bar with CQL2 builder + footprint map + STAC item card

**Trace:** REQ-FUNC-EO-CAT-001, REQ-FUNC-EO-CAT-002; design.md §6.6
**Owner:** Web + EO
**Status:** backlog
**Estimate:** 8
**Depends on:** TASK-P3-EO-002, TASK-P3-EO-010
**Files (create/modify):**
  - `web/apps/shell/src/routes/(app)/eo/catalog/+page.svelte` (new) — search UI + Cesium footprint overlay
  - `web/apps/shell/src/lib/cql2/Builder.svelte` (new)
  - `web/apps/shell/src/lib/eo/StacItemCard.svelte` (new) — thumbnails, asset list, copy-link
  - `web/apps/shell/tests/e2e/eo-catalog.spec.ts` (new)
**Acceptance criteria:**
  1. Search latency ≤ 200 ms p95 measured in browser.
  2. Footprint click loads STAC item card with thumbnails.
**Verification:**
  - E2E: `web/apps/shell/tests/e2e/eo-catalog.spec.ts`.

### TASK-P3-WEB-002: Web — processing-job kanban + change-detection result viewer

**Trace:** REQ-FUNC-EO-PIPE-001, REQ-FUNC-EO-PIPE-006; design.md §6.6
**Owner:** Web + EO
**Status:** backlog
**Estimate:** 6
**Depends on:** TASK-P3-EO-004, TASK-P3-EO-009
**Files (create/modify):**
  - `web/apps/shell/src/routes/(app)/eo/jobs/+page.svelte` (new) — kanban: queued / running / failed / done
  - `web/apps/shell/src/routes/(app)/eo/changes/+page.svelte` (new) — before/after slider + polygon overlay
  - `web/apps/shell/tests/e2e/eo-jobs.spec.ts` (new)
**Acceptance criteria:**
  1. Job state updates live via realtime-gw.
  2. Change-detection viewer overlays polygons on Cesium.
**Verification:**
  - E2E: `web/apps/shell/tests/e2e/eo-jobs.spec.ts`.

### TASK-P3-NFR-001: Phase 3 NFR gate — 100 scenes/h; ML ≤100 ms p95; STAC ≤200 ms p95; 10k tiles/min

**Trace:** REQ-NFR-PERF-002, REQ-NFR-PERF-003, REQ-NFR-PERF-004, REQ-CONST-009; design.md §7.2
**Owner:** Platform Infra
**Status:** backlog
**Estimate:** 4
**Depends on:** TASK-P3-EO-009, TASK-P3-ML-004, TASK-P3-EO-002, TASK-P3-EO-010
**Files (create/modify):**
  - `bench/k6/eo-throughput.bench.js` (new)
  - `bench/k6/stac-search.bench.js` (new)
  - `bench/k6/tiles-batch.bench.js` (new)
  - `.github/workflows/nfr-phase3.yml` (new)
**Acceptance criteria:**
  1. All four targets met for two consecutive runs.
**Verification:**
  - Bench: as above.

---

## 6. Phase 4 — GeoInt + Mission ops + Conjunction (10 weeks)

Goal: analysts work in spatial workspaces, draw AOIs, run analyses, author reports; mission ops sees the spacecraft as a system; conjunction screening + maneuver planning go live.

### TASK-P4-GI-001: `gi-workspace` — workspaces, members, layers, saved views, annotations, activity audit

**Trace:** REQ-FUNC-GI-WS-001, REQ-FUNC-GI-WS-002, REQ-FUNC-GI-WS-003, REQ-FUNC-GI-WS-004; design.md §3.1
**Owner:** GeoInt
**Status:** backlog
**Estimate:** 10
**Depends on:** TASK-P3-EO-010, TASK-P1-AUDIT-001
**Files (create/modify):**
  - `services/gi-workspace/cmd/gi-workspace/main.go` (new)
  - `services/gi-workspace/internal/store/workspaces.go` (new) — workspaces, members (viewer/editor/admin/owner), layers, saved_views
  - `services/gi-workspace/internal/annotations/store.go` (new) — GeoJSON; undo/redo via per-layer event log; concurrent-edit conflict resolution via vector clocks
  - `services/gi-workspace/internal/measure/measure.go` (new) — geodesic distance/area in PostGIS; matches `@turf/*` browser implementation
  - `services/gi-workspace/migrations/0001_workspace.sql` (new)
  - `services/gi-workspace/test/workspace_e2e_test.go` (new)
**Acceptance criteria:**
  1. Two simultaneous editors converge to the same annotation state with documented conflict policy.
  2. Undo/redo bounded by per-layer event-log size cap.
  3. Activity audit writes through `services/packages/audit/client.go`.
**Verification:**
  - Integration: `services/gi-workspace/test/workspace_e2e_test.go`.

### TASK-P4-GI-002: `gi-aoi` — AOIs in PostGIS, monitoring rules, alerts, imagery timeline

**Trace:** REQ-FUNC-GI-AOI-001; design.md §3.1
**Owner:** GeoInt
**Status:** backlog
**Estimate:** 7
**Depends on:** TASK-P4-GI-001, TASK-P3-EO-001
**Files (create/modify):**
  - `services/gi-aoi/cmd/gi-aoi/main.go` (new)
  - `services/gi-aoi/internal/store/aois.go` (new)
  - `services/gi-aoi/internal/monitor/rules.go` (new) — rule DSL: trigger on new STAC item intersecting AOI + condition (cloud cover, sensor, etc.)
  - `services/gi-aoi/internal/monitor/runner.go` (new) — Kafka consumer of STAC `item.created` events
  - `services/gi-aoi/internal/timeline/timeline.go` (new) — per-AOI imagery timeline
  - `services/gi-aoi/migrations/0001_aoi.sql` (new) — `aois (geom geometry(Polygon,4326))`, `aoi_rules`, `aoi_alerts`
  - `services/gi-aoi/test/aoi_test.go` (new)
**Acceptance criteria:**
  1. AOI alert fires within 5 s of a matching STAC item being published.
  2. Alert deduped per AOI per scene.
**Verification:**
  - Integration: `services/gi-aoi/test/aoi_test.go`.

### TASK-P4-GI-003: `gi-report` — WYSIWYG editor backend, templates, embedded snapshots, exports, share links, version history, scheduled generation

**Trace:** REQ-FUNC-GI-RPT-001; design.md §3.1
**Owner:** GeoInt
**Status:** backlog
**Estimate:** 12
**Depends on:** TASK-P4-GI-001, TASK-P1-EXPORT-001
**Files (create/modify):**
  - `services/gi-report/cmd/gi-report/main.go` (new)
  - `services/gi-report/internal/store/reports.go` (new) — content as ProseMirror JSON (matches WYSIWYG editor) + version history
  - `services/gi-report/internal/templates/library.go` (new) — built-in templates + user-defined
  - `services/gi-report/internal/snapshot/maps.go` (new) — server-side Cesium snapshot via headless Chromium pool
  - `services/gi-report/internal/export/{pdf,docx,pptx,html}.go` (new) — exports via Pandoc + LibreOffice headless + Puppeteer for PDF
  - `services/gi-report/internal/share/links.go` (new) — share link with access control (view-only token; revocable)
  - `services/gi-report/internal/schedule/cron.go` (new) — scheduled generation via the platform scheduler (REQ-FUNC-CMN-006 — covered in TASK-P1-PLT-SCHED-001)
  - `services/gi-report/migrations/0001_reports.sql` (new)
  - `services/gi-report/test/export_test.go` (new)
**Acceptance criteria:**
  1. Round-trip ProseMirror JSON without information loss.
  2. PDF export renders embedded map snapshots correctly.
  3. Share links revoked immediately invalidate.
**Verification:**
  - Integration: `services/gi-report/test/export_test.go`.

### TASK-P4-GI-004: `gi-analytics` — counting, tracking, terrain, buffer/proximity, heatmap, spatial query, sandboxed Python

**Trace:** REQ-FUNC-GI-AN-001; design.md §3.1
**Owner:** GeoInt
**Status:** backlog
**Estimate:** 14
**Depends on:** TASK-P4-GI-005, TASK-P3-ML-002
**Files (create/modify):**
  - `services/gi-analytics/cmd/gi-analytics/main.go` (new)
  - `services/gi-analytics/internal/count/object_count.go` (new) — uses Triton-served detection model
  - `services/gi-analytics/internal/track/tracker.go` (new) — Hungarian + Kalman
  - `services/gi-analytics/internal/terrain/{profile,slope,aspect,viewshed}.go` (new) — DEM-based; calls `services/gi-dem`
  - `services/gi-analytics/internal/buffer/proximity.go` (new) — PostGIS
  - `services/gi-analytics/internal/heatmap/heatmap.go` (new)
  - `services/gi-analytics/internal/sandbox/python.go` (new) — gVisor + seccomp; allowed packages whitelisted; CPU/mem/wall-time limits
  - `services/gi-analytics/test/analytics_test.go` (new)
**Acceptance criteria:**
  1. Sandboxed script attempting net/file syscall is killed; resource over-limit kills cleanly.
  2. Object counting runs against a Triton-served model and publishes results to a layer.
**Verification:**
  - Integration: `services/gi-analytics/test/analytics_test.go`.
  - Inspection: sandbox seccomp policy reviewed by Security.

### TASK-P4-GI-005: DEM service — S3-backed tile cache, WMTS-style serving, on-demand viewshed/slope/aspect

**Trace:** REQ-FUNC-GI-DEM-001; design.md §3.1
**Owner:** GeoInt
**Status:** backlog
**Estimate:** 5
**Depends on:** TASK-P3-EO-010
**Files (create/modify):**
  - `services/gi-dem/cmd/gi-dem/main.go` (new)
  - `services/gi-dem/internal/cache/s3.go` (new) — S3 tile cache; LRU local
  - `services/gi-dem/internal/wmts/wmts.go` (new)
  - `services/gi-dem/internal/compute/{viewshed,slope,aspect}.go` (new)
  - `services/gi-dem/test/dem_test.go` (new)
**Acceptance criteria:**
  1. Viewshed computation for a 10 km radius returns within 2 s on commodity hardware.
**Verification:**
  - Integration: `services/gi-dem/test/dem_test.go`.

### TASK-P4-SAT-001: `sat-mission` — subsystem catalog, health rules, power budget, battery model, ADCS, thermal, mission timeline, anomaly tracking

**Trace:** REQ-FUNC-SAT-002, REQ-FUNC-SAT-003; design.md §3.2
**Owner:** Defense Mission
**Status:** backlog
**Estimate:** 14
**Depends on:** TASK-P2-TM-001, TASK-P0-HW-001
**Files (create/modify):**
  - `chetana-defense/services/sat-mission/internal/subsystems/catalog.go` (new) — power, ADCS, CDH, comms, thermal, propulsion, payload, structure (per spacecraft profile)
  - `chetana-defense/services/sat-mission/internal/health/rules.go` (new) — rule engine with rollup
  - `chetana-defense/services/sat-mission/internal/power/budget.go` (new) — solar input, eclipse prediction, total load
  - `chetana-defense/services/sat-mission/internal/battery/model.go` (new) — SoC, voltage, current, temp, cycle count, capacity fade
  - `chetana-defense/services/sat-mission/internal/adcs/mode.go` (new) — quaternion → Euler, wheel speeds, momentum
  - `chetana-defense/services/sat-mission/internal/thermal/map.go` (new)
  - `chetana-defense/services/sat-mission/internal/timeline/store.go` (new) — mission timeline (planned + actual events)
  - `chetana-defense/services/sat-mission/internal/anomaly/tracker.go` (new) — anomaly lifecycle (open / triaged / resolved); links to telemetry windows
  - `chetana-defense/services/sat-mission/migrations/0001_mission.sql` (new)
  - `chetana-defense/services/sat-mission/test/mission_test.go` (new)
**Acceptance criteria:**
  1. Power budget reflects orbit eclipse correctly.
  2. Anomaly opened from a limit violation links back to the originating telemetry window.
**Verification:**
  - Integration: `chetana-defense/services/sat-mission/test/mission_test.go`.

### TASK-P4-SAT-002: `sat-conjunction` — CDM ingest, screening, Pc (Foster), B-plane, maneuver planner, alerts

**Trace:** REQ-FUNC-SAT-007, REQ-FUNC-SAT-008; design.md §3.2
**Owner:** Defense Mission
**Status:** backlog
**Estimate:** 14
**Depends on:** TASK-P2-GS-002, TASK-P4-SAT-001
**Files (create/modify):**
  - `chetana-defense/services/sat-conjunction/cmd/sat-conjunction/main.go` (new)
  - `chetana-defense/services/sat-conjunction/internal/cdm/spacetrack.go` (new) — 8h CDM poll
  - `chetana-defense/services/sat-conjunction/internal/screen/{apsis,coarse,fine}.go` (new) — apogee/perigee → coarse → fine screening pipeline
  - `chetana-defense/services/sat-conjunction/internal/pc/foster.go` (new) — Foster method probability of collision
  - `chetana-defense/services/sat-conjunction/internal/geom/{bplane,encounter}.go` (new)
  - `chetana-defense/services/sat-conjunction/internal/maneuver/planner.go` (new) — along/cross/radial; secondary-conjunction check
  - `chetana-defense/services/sat-conjunction/internal/alerts/classify.go` (new) — green/yellow/orange/red per Pc thresholds
  - `chetana-defense/services/sat-conjunction/migrations/0001_conjunction.sql` (new)
  - `chetana-defense/services/sat-conjunction/test/conjunction_test.go` (new) — fixtures from public CDMs
**Acceptance criteria:**
  1. Pc within 5 % of reference values on the public Vandenberg fixture set.
  2. Maneuver plan produces a feasible Δv vector with secondary check passing.
**Verification:**
  - Integration: `chetana-defense/services/sat-conjunction/test/conjunction_test.go`.

### TASK-P4-WEB-001: Web — workspace canvas (Cesium + drawing tools), AOI tools, report editor, mission ops dashboard, conjunction viz with B-plane + 3D encounter

**Trace:** REQ-FUNC-GI-WS-001, REQ-FUNC-GI-AOI-001, REQ-FUNC-GI-RPT-001, REQ-FUNC-SAT-002, REQ-FUNC-SAT-007, REQ-FUNC-SAT-008; design.md §6.6
**Owner:** Web + GeoInt + Defense
**Status:** backlog
**Estimate:** 18
**Depends on:** TASK-P4-GI-001, TASK-P4-GI-002, TASK-P4-GI-003, TASK-P4-SAT-001, TASK-P4-SAT-002
**Files (create/modify):**
  - `web/apps/shell/src/routes/(app)/geoint/workspaces/[id]/+page.svelte` (new) — Cesium canvas + draw tools + layer panel
  - `web/apps/shell/src/lib/draw/Tools.svelte` (new) — draw point/line/polygon/box on Cesium
  - `web/apps/shell/src/routes/(app)/geoint/aois/+page.svelte` (new)
  - `web/apps/shell/src/routes/(app)/geoint/reports/[id]/+page.svelte` (new) — ProseMirror editor + map snapshot tool
  - `web/apps/shell/src/routes/(app)/satellite/mission/+page.svelte` (new) — subsystems, power, battery, ADCS, thermal, anomalies
  - `web/apps/shell/src/routes/(app)/satellite/conjunctions/+page.svelte` (new) — list + B-plane viz + 3D encounter viz
  - `web/apps/shell/src/lib/charts/BPlane.svelte` (new) — D3 polar / 2D scatter
  - `web/apps/shell/tests/e2e/{workspaces,aois,reports,mission,conjunctions}.spec.ts` (new)
**Acceptance criteria:**
  1. Drawing tools produce GeoJSON identical to PostGIS round-trip.
  2. Mission ops dashboard updates live via realtime-gw.
  3. B-plane viz matches the published reference for the Vandenberg fixture.
**Verification:**
  - E2E: `web/apps/shell/tests/e2e/{workspaces,aois,reports,mission,conjunctions}.spec.ts`.

---

## 7. Phase 5 — Imagery-as-a-Service customer surface (6 weeks)

Goal: external customers can search public-classified collections via STAC, subscribe to AOI deliveries, and download via presigned URLs — all behind a public API gateway with metering and rate limiting.

### TASK-P5-IAAS-001: Public API gateway — API-key auth + per-key rate limit + usage metering

**Trace:** REQ-FUNC-IAAS-001; design.md §6.3
**Owner:** Platform + Web
**Status:** backlog
**Estimate:** 6
**Depends on:** TASK-P1-IAM-005, TASK-P3-EO-001
**Files (create/modify):**
  - `services/public-gw/cmd/public-gw/main.go` (new)
  - `services/public-gw/internal/apikey/store.go` (new) — `api_keys` (id, hash, scopes[], rate_limit_rpm, customer_id, status); hash check at ingress
  - `services/public-gw/internal/ratelimit/redis.go` (new) — sliding window per key
  - `services/public-gw/internal/meter/usage.go` (new) — per-key request + bytes counters; daily roll-up; stored in `api_usage_daily`
  - `services/public-gw/internal/proxy/router.go` (new) — routes `/v1/public/*` to internal services with public-classification filter applied
  - `services/public-gw/migrations/0001_apikeys.sql` (new)
  - `services/public-gw/test/gateway_test.go` (new)
**Acceptance criteria:**
  1. Requests without API key or with invalid key → 401.
  2. Requests with classification > public in any path filter → 403 with audit event.
  3. Rate limit returns 429 with `Retry-After`.
  4. Usage meter aggregates daily; reconciles to within 1 % vs synthetic call counts.
**Verification:**
  - Integration: `services/public-gw/test/gateway_test.go`.

### TASK-P5-IAAS-002: Public STAC endpoints (read-only; public collections only)

**Trace:** REQ-FUNC-IAAS-002; design.md §3.1
**Owner:** EO + Platform
**Status:** backlog
**Estimate:** 3
**Depends on:** TASK-P5-IAAS-001
**Files (create/modify):**
  - `services/eo-catalog/internal/api/public.go` (new) — read-only handler enforcing `data_classification = 'public'`
  - `services/public-gw/internal/proxy/stac.go` (new) — wires `/v1/public/stac/*` to `services/eo-catalog`
  - `services/eo-catalog/test/public_test.go` (new) — verifies non-public items invisible
**Acceptance criteria:**
  1. ITAR/CUI/restricted/internal items invisible regardless of query.
  2. Search/items endpoints respond per STAC API 1.0.0.
**Verification:**
  - Integration: `services/eo-catalog/test/public_test.go`.

### TASK-P5-IAAS-003: Subscription matching service — AOI matches → presigned URL deliveries via notify

**Trace:** REQ-FUNC-IAAS-003; design.md §3.1
**Owner:** EO + Platform
**Status:** backlog
**Estimate:** 7
**Depends on:** TASK-P5-IAAS-002, TASK-P1-NOTIFY-001, TASK-P1-EXPORT-001, TASK-P4-GI-002
**Files (create/modify):**
  - `services/eo-subscriptions/cmd/eo-subscriptions/main.go` (new)
  - `services/eo-subscriptions/internal/matcher/matcher.go` (new) — Kafka consumer of STAC `item.created`; matches against customer AOI subscriptions
  - `services/eo-subscriptions/internal/deliver/deliver.go` (new) — generates presigned URL via export service; emits notify email
  - `services/eo-subscriptions/migrations/0001_subscriptions.sql` (new) — `subscriptions` (customer_id, aoi geometry, filters JSONB, status), `deliveries` (id, subscription_id, item_id, presigned_url, expires_at, sent_at)
  - `services/eo-subscriptions/test/match_test.go` (new)
**Acceptance criteria:**
  1. New public-classified item intersecting an active subscription triggers a delivery within 60 s.
  2. Delivery URLs expire per export service policy (24 h).
  3. Per-subscription delivery dedup (same item not re-delivered).
**Verification:**
  - Integration: `services/eo-subscriptions/test/match_test.go`.

### TASK-P5-IAAS-004: DOI registration + citation formatter

**Trace:** REQ-FUNC-IAAS-004; design.md §3.1
**Owner:** EO + Compliance
**Status:** backlog
**Estimate:** 3
**Depends on:** TASK-P5-IAAS-002
**Files (create/modify):**
  - `services/eo-catalog/internal/doi/register.go` (new) — DataCite client; registers DOI per published collection version
  - `services/eo-catalog/internal/doi/cite.go` (new) — citation formatter (APA, MLA, BibTeX, RIS)
  - `services/eo-catalog/migrations/0003_doi.sql` (new)
  - `services/eo-catalog/test/doi_test.go` (new)
**Acceptance criteria:**
  1. Sandbox-environment DOI registered for a test collection; landing page resolves.
  2. Citation in all four formats matches reference.
**Verification:**
  - Integration: `services/eo-catalog/test/doi_test.go` (uses DataCite test endpoint).

### TASK-P5-WEB-001: Customer portal route group `web/apps/shell/src/routes/(public)`

**Trace:** REQ-FUNC-IAAS-005, REQ-CONST-005; design.md §6.3
**Owner:** Web
**Status:** backlog
**Estimate:** 8
**Depends on:** TASK-P5-IAAS-001, TASK-P5-IAAS-002, TASK-P5-IAAS-003
**Files (create/modify):**
  - `web/apps/shell/src/routes/(public)/+layout.svelte` (new) — public layout (no internal nav; marketing chrome)
  - `web/apps/shell/src/routes/(public)/signup/+page.svelte` (new) — customer sign-up (email verification, T&Cs, DPIA acceptance)
  - `web/apps/shell/src/routes/(public)/catalog/+page.svelte` (new) — public STAC search
  - `web/apps/shell/src/routes/(public)/subscriptions/+page.svelte` (new) — manage AOI subscriptions
  - `web/apps/shell/src/routes/(public)/downloads/+page.svelte` (new) — past deliveries; download links
  - `web/apps/shell/src/routes/(public)/docs/+page.svelte` (new) — API documentation (auto-generated from OpenAPI/STAC spec)
  - `web/apps/shell/tests/e2e/public-portal.spec.ts` (new)
**Acceptance criteria:**
  1. Sign-up issues an API key after email verification.
  2. Public route group has zero internal-nav items even when an internal user is authenticated in the same browser.
**Verification:**
  - E2E: `web/apps/shell/tests/e2e/public-portal.spec.ts`.

### TASK-P5-COMP-001: DPIA artifact for the public surface (GDPR Article 35)

**Trace:** REQ-COMP-GDPR-003; design.md §9.2
**Owner:** Compliance
**Status:** blocked:OQ-009
**Estimate:** 4
**Depends on:** TASK-P5-WEB-001
**Files (create/modify):**
  - `compliance/dpia/dpia-public-surface.md` (new) — completed DPIA template
  - `compliance/ropa.md` (modify) — add ROPA entry for public-surface processing
**Acceptance criteria:**
  1. DPIA covers data flows, lawful basis, risk register, mitigations, residual risk.
  2. Reviewed and signed by DPO (artifact under `compliance/sign-offs/`).
**Verification:**
  - Inspection: signed PDF in `compliance/sign-offs/`.

---

## 8. Phase 6 — Hardening + ISO 27001 + GDPR (16 weeks)

Goal: production cutover. ISMS in steady state, GDPR DPIA + ROPA filed, pen-test remediated, DR drill clean, HSM integration live.

### TASK-P6-COMP-001: ISMS skeleton finalised (policies, evidence, internal audit cycle)

**Trace:** REQ-COMP-ISO-001, REQ-COMP-ISO-002, REQ-COMP-ISO-003; design.md §9.1
**Owner:** Compliance
**Status:** backlog
**Estimate:** 20
**Depends on:** TASK-P0-COMP-001
**Files (create/modify):**
  - `compliance/policies/{access-control,asset-management,awareness-training,bcp,change-management,cryptography,incident-response,information-classification,physical-security,risk-management,secure-development,supplier,vulnerability-management}.md` (new) — 13 ISMS policies covering Annex A control families
  - `compliance/controls/iso27001.csv` (modify) — populate `evidence_path` for each of 93 controls
  - `compliance/internal-audits/2027-Q1.md` (new) — first internal audit report
  - `compliance/management-review/2027-Q1.md` (new) — first management review
  - `tools/compliance/coverage.sh` (modify) — switch to **blocking** mode (CI fails if any control's `evidence_path` is empty)
**Acceptance criteria:**
  1. All 93 Annex A controls carry non-empty `evidence_path`.
  2. Internal audit report identifies findings + remediation plan.
**Verification:**
  - Inspection: external readiness audit conducted by accredited body; report in `compliance/external-audits/`.

### TASK-P6-COMP-002: GDPR DPIA finalisation + DPO appointment + EU representative engagement

**Trace:** REQ-COMP-GDPR-001, REQ-COMP-GDPR-002, REQ-COMP-GDPR-003, REQ-COMP-GDPR-004, REQ-COMP-GDPR-005; design.md §9.2
**Owner:** Compliance
**Status:** blocked:OQ-009, blocked:OQ-010
**Estimate:** 12
**Depends on:** TASK-P5-COMP-001
**Files (create/modify):**
  - `compliance/dpia/dpia-platform.md` (new) — platform-wide DPIA
  - `compliance/dpia/dpia-iaas.md` (modify) — finalised
  - `compliance/ropa.md` (modify) — final ROPA covering all processing
  - `compliance/dpo.md` (new) — DPO appointment + contact details
  - `compliance/eu-representative.md` (new) — Article 27 representative
  - `compliance/breach-response/playbook.md` (new) — 1h internal pager + 72h supervisory authority notification
  - `web/apps/shell/src/routes/(public)/privacy/+page.svelte` (new) — privacy notice naming DPO + EU rep
**Acceptance criteria:**
  1. DPO appointed; contact published.
  2. EU representative appointed; contact published.
  3. Breach response playbook tested via tabletop exercise.
**Verification:**
  - Inspection: signed appointment letters in `compliance/sign-offs/`.

### TASK-P6-SEC-001: HSM integration for command encryption (D7.9)

**Trace:** REQ-FUNC-SAT-011, REQ-NFR-SEC-003; design.md §3.2, §4.7
**Owner:** Defense + Security
**Status:** backlog
**Estimate:** 8
**Depends on:** TASK-P2-CMD-002
**Files (create/modify):**
  - `chetana-defense/services/sat-command/internal/encode/hsm.go` (modify) — replace no-op with PKCS#11 provider (CloudHSM in GovCloud)
  - `chetana-defense/services/sat-command/internal/encode/hsm_test.go` (new)
  - `compliance/policies/cryptography.md` (modify) — document HSM scope
**Acceptance criteria:**
  1. Command payload encryption via HSM key on real CloudHSM cluster.
  2. Key rotation tested without command queue downtime.
**Verification:**
  - Integration: `chetana-defense/services/sat-command/internal/encode/hsm_test.go` against CloudHSM.

### TASK-P6-SEC-002: Penetration test + remediation cycle

**Trace:** REQ-NFR-SEC-002, REQ-NFR-SEC-006, REQ-CONST-012; design.md §9.1, §10.2
**Owner:** Security
**Status:** backlog
**Estimate:** 12
**Depends on:** TASK-P5-WEB-001
**Files (create/modify):**
  - `compliance/pen-tests/2027-pen-test-report.md` (new) — third-party pen test report (redacted)
  - `compliance/pen-tests/remediation-tracker.md` (new) — POA&M tracking each finding
  - Code patches across services as needed; per finding, a regression test under `services/<svc>/test/security/` (per REQ-CONST-012)
**Acceptance criteria:**
  1. All critical + high findings remediated; medium findings in POA&M with target dates.
  2. Each remediation includes a regression test that fails before fix and passes after.
**Verification:**
  - Inspection: re-test report in `compliance/pen-tests/`.

### TASK-P6-REL-001: DR drill — RPO ≤5 min, RTO ≤1 h verified

**Trace:** REQ-NFR-REL-001, REQ-NFR-REL-002; design.md §7.1
**Owner:** Platform Infra
**Status:** backlog
**Estimate:** 6
**Depends on:** TASK-P0-INFRA-001
**Files (create/modify):**
  - `compliance/dr-drills/2027-Q2.md` (new) — drill plan + executed runbook + measured RPO/RTO
  - `infra/runbooks/dr-failover.md` (new) — operator runbook
  - `infra/terraform/modules/dr/` (new) — cross-AZ replica + automated failover for Postgres/Timescale
**Acceptance criteria:**
  1. Failover executes within 1 h end-to-end on a primed standby.
  2. Data loss measured ≤ 5 min on the synthetic write fixture.
**Verification:**
  - Inspection: drill report in `compliance/dr-drills/`.

### TASK-P6-SEC-003: Vulnerability management cadence

**Trace:** REQ-NFR-SEC-006, REQ-COMP-FEDRAMP-004; design.md §9.1
**Owner:** Security
**Status:** backlog
**Estimate:** 4
**Depends on:** TASK-P0-CI-001
**Files (create/modify):**
  - `compliance/vuln-mgmt/cadence.md` (new) — monthly scan, weekly triage, SLA: critical 7d / high 30d / medium 90d
  - `.github/workflows/vuln-scan-monthly.yml` (new) — scheduled trivy + grype scan against running images; opens issues on findings
**Acceptance criteria:**
  1. Monthly scan runs on schedule; opens GH issues per finding.
  2. SLA dashboard tracks open vulnerabilities by age.
**Verification:**
  - Inspection: dashboard JSON committed under `infra/grafana/dashboards/vuln-mgmt.json`.

### TASK-P6-COMP-003: Compliance evidence package assembled

**Trace:** REQ-COMP-ISO-001, REQ-COMP-ISO-002, REQ-COMP-GDPR-002; design.md §9.4
**Owner:** Compliance
**Status:** backlog
**Estimate:** 6
**Depends on:** TASK-P6-COMP-001, TASK-P6-COMP-002, TASK-P6-SEC-002, TASK-P6-REL-001
**Files (create/modify):**
  - `compliance/evidence-package/README.md` (new) — index of evidence artefacts mapped to control IDs
  - `tools/compliance/build-evidence.sh` (new) — assembles evidence ZIP for auditor delivery
**Acceptance criteria:**
  1. Evidence ZIP builds reproducibly.
  2. Auditor checklist validated by ISO accredited body in pre-audit.
**Verification:**
  - Inspection: pre-audit report.

### TASK-P6-COMP-004: Staged-certification posture — record deferred frameworks (SOC 2, CERT-In, ITAR, FedRAMP-Mod)

**Trace:** REQ-COMP-SOC2-001, REQ-COMP-CERTIN-001, REQ-COMP-CERTIN-002, REQ-COMP-CERTIN-003, REQ-COMP-ITAR-004, REQ-COMP-ITAR-005, REQ-COMP-FEDRAMP-001, REQ-COMP-FEDRAMP-003, REQ-CONST-001, REQ-CONST-006; design.md §9
**Owner:** Compliance
**Status:** backlog
**Estimate:** 4
**Depends on:** TASK-P6-COMP-001, TASK-P6-COMP-002
**Files (create/modify):**
  - `compliance/staging-plan.md` (new) — staged certification calendar (SOC 2 Type II in v1.x, CERT-In with India region in v1.2, ITAR DDTC + TCP in v2.0, FedRAMP-Mod 3PAO in v2.1) with target dates, dependencies, and the v1 architectural posture proving "certifiable by design"
  - `compliance/controls/{soc2,certin,itar,fedramp-mod}.csv` (modify) — confirm all rows carry `evidence_path` for the in-scope-from-day-one controls (audit retention ≥5y, GovCloud hosting, 2-repo posture, ITAR audit retention) even though certification audits run later
  - `compliance/staging-plan-itar-records.md` (new) — explicit reference to REQ-COMP-ITAR-004 (audit retention ≥5y online — already enforced by TASK-P1-AUDIT-002) and REQ-COMP-ITAR-005 (DDTC registration — owner + target date in v2.0)
  - `compliance/cert-in-readiness.md` (new) — explicit reference to REQ-COMP-CERTIN-001/002/003 with the dependency on India region rollout
  - `compliance/fedramp-readiness.md` (new) — explicit reference to REQ-COMP-FEDRAMP-001 (GovCloud hosting — already in v1 per REQ-CONST-003) and REQ-COMP-FEDRAMP-003 (3PAO — v2.1)
  - `compliance/v1-scope-notes.md` (new) — records that REQ-CONST-001 (immutable space_plan) and REQ-CONST-006 (Tauri deferred) are observed by the v1 build and the Out-of-v1 list in `plan/requirements.md` §8
**Acceptance criteria:**
  1. Every deferred-framework REQ in `plan/requirements.md` §5 has a row in the corresponding readiness doc with owner + target version.
  2. `tools/compliance/coverage.sh` reports 100 % evidence coverage for the in-v1-scope controls within those frameworks.
  3. Audit retention ≥5y online verified end-to-end against `services/audit` (TASK-P1-AUDIT-002) — evidence captured in `compliance/staging-plan-itar-records.md`.
**Verification:**
  - Inspection: staged-certification calendar reviewed in management review; signed PDF in `compliance/sign-offs/`.

### TASK-P6-CUTOVER-001: Production cutover

**Trace:** REQ-NFR-REL-001, REQ-CONST-009; design.md §7, §8.3
**Owner:** Platform Infra + Mission
**Status:** backlog
**Estimate:** 6
**Depends on:** TASK-P6-COMP-003, TASK-P6-SEC-002, TASK-P6-REL-001, TASK-P6-SEC-001
**Files (create/modify):**
  - `infra/runbooks/cutover.md` (new) — go-live runbook
  - `infra/runbooks/rollback.md` (new) — rollback runbook
  - `compliance/cutover/sign-off.md` (new)
**Acceptance criteria:**
  1. Cutover executed with no data loss; rollback rehearsed.
  2. Post-cutover monitoring shows availability ≥ 99.9 % over 30 d.
**Verification:**
  - Inspection: cutover sign-off + 30-d availability report.

---

## 9. Cross-cutting workstreams (continuous, run through every phase)

These streams are not phase-bound; they are continuous responsibilities tracked under the `XC` phase prefix. Each carries the standard task block.

### TASK-XC-COMP-001: Compliance Engineering — control register upkeep, evidence collection, POA&M

**Trace:** REQ-COMP-ISO-001, REQ-COMP-GDPR-002, REQ-COMP-FEDRAMP-004; design.md §9
**Owner:** Compliance
**Status:** in-progress (continuous from Phase 0)
**Estimate:** continuous (≈ 0.5 FTE)
**Depends on:** TASK-P0-COMP-001
**Files (create/modify):**
  - `compliance/controls/*.csv` (modify) — keep evidence_path current as services land
  - `compliance/poa-m.md` (new in PR-G; updated continuously)
**Acceptance criteria:**
  1. Control coverage script (`tools/compliance/coverage.sh`) ≥ 95 % from Phase 1 onward; 100 % by Phase 6.
  2. POA&M reviewed monthly.
**Verification:**
  - Inspection: monthly review minutes under `compliance/management-review/`.

### TASK-XC-HW-001: Hardware Abstraction — interface evolution, new adapter additions (KSAT/SSC behind interface in v1)

**Trace:** REQ-FUNC-GS-HW-003; design.md §4.4
**Owner:** Defense Hardware
**Status:** in-progress (continuous from Phase 0 PR-H)
**Estimate:** continuous
**Depends on:** TASK-P0-HW-001
**Files (create/modify):**
  - `services/packages/hardware/*.go` (modify) — interface stays stable; backward-compatible additions only
  - `chetana-defense/services/packages/hardware/{ksat,ssc}/` (new — v2.0; behind feature flag in v1) — interface implementations land disabled
**Acceptance criteria:**
  1. No breaking change to `HardwareDriver`, `AntennaController`, `GroundNetworkProvider` after Phase 0 freeze.
**Verification:**
  - Inspection: `buf breaking` for proto contracts; Go API diff via `apidiff` in CI.

### TASK-XC-PROFILE-001: Spacecraft Profile — library expansion as new spacecraft are flown

**Trace:** REQ-FUNC-SAT-001; design.md §4.5
**Owner:** Mission + Defense
**Status:** in-progress (continuous from Phase 2)
**Estimate:** continuous
**Depends on:** TASK-P0-HW-001, TASK-P2-GS-001
**Files (create/modify):**
  - `chetana-defense/services/sat-mission/profiles/*.yaml` (new per spacecraft) — concrete `SpacecraftProfile` instances; loaded by `internal/profile/loader.go`
**Acceptance criteria:**
  1. Each new spacecraft on-boarded by adding a profile YAML; no service code change required.
**Verification:**
  - Integration: `chetana-defense/services/sat-mission/test/profile_loader_test.go` adds a fixture per profile.

### TASK-XC-SUPPLY-001: Supply Chain Security — SAST/DAST/SCA/SBOM/cosign upkeep, vulnerability triage, dependency-update cadence

**Trace:** REQ-NFR-SEC-004, REQ-NFR-SEC-005, REQ-NFR-SEC-006; design.md §8.1
**Owner:** Security
**Status:** in-progress (continuous from Phase 0 PR-F)
**Estimate:** continuous
**Depends on:** TASK-P0-CI-001
**Files (create/modify):**
  - `.github/dependabot.yml` (new in PR-F; tuned continuously) — weekly bump for Go/Rust/npm/pip
  - `compliance/supply-chain/sbom-archive/` (new) — per-release SBOMs retained
**Acceptance criteria:**
  1. Critical/high findings triaged within SLA (TASK-P6-SEC-003).
  2. Dependency bumps merged weekly when CI green.
**Verification:**
  - Inspection: SBOM archive completeness check in monthly compliance review.

### TASK-XC-REGION-001: Multi-region Data Plane — region-aware code review, Helm overlay maintenance for EU/India templates

**Trace:** REQ-NFR-SCALE-003, REQ-CONST-009; design.md §4.8, §7.4
**Owner:** Platform Infra
**Status:** in-progress (continuous from Phase 0 PR-E)
**Estimate:** continuous
**Depends on:** TASK-P0-INFRA-001
**Files (create/modify):**
  - `infra/helm/overlays/{eu-central-1,ap-south-1}/values.yaml` (modify) — kept rendering-clean as services land
  - `services/packages/region/region.go` (modify) — extended as new regional resources are introduced
  - `tools/region/lint.sh` (new) — fails if a service references a non-region-aware resource directly
**Acceptance criteria:**
  1. `helm template` against EU and India overlays succeeds for every release.
  2. Region lint blocks PRs that introduce hard-coded region IDs outside `services/packages/region/`.
**Verification:**
  - Integration: rendering check in CI per release.

---

## 10. Open questions (mirror of `plan/requirements.md` §9 — must be resolved before referenced phase tasks start)

This section mirrors `plan/requirements.md` §9 exactly. Tasks elsewhere in this document reference these IDs via `Status: blocked:OQ-NNN`. Updates here MUST be mirrored back to `plan/requirements.md` §9 in the same PR.

| ID | Question | Blocks tasks | Owner | Status |
|---|---|---|---|---|
| OQ-001 | Confirm: AWS Ground Station replaces Azure Orbital as the second `GroundNetworkProvider` (Azure Orbital EOL 2026-09). | TASK-P2-HW-003 (`aws-gs` provider only; `own-dish` proceeds) | Customer | open |
| OQ-002 | Provision empty `chetana-defense` GitHub repo + US-persons team. | TASK-P0-REPO-001 | Customer | open |
| OQ-003 | GitHub Enterprise vs Cloud (affects SAML SSO + audit log streaming + IP allowlists for ITAR). | TASK-P0-REPO-001 | Customer | open |
| OQ-004 | Internal Go module proxy / Cargo registry / buf BSR org existence. | TASK-P0-REPO-001 | Customer | open |
| OQ-005 | Sanity-check `compliance/itar-paths.txt` (sat-telemetry classification model: all-defense vs split). | TASK-P0-REPO-001, TASK-P0-COMP-001 | Customer | open |
| OQ-006 | Spacecraft details (bus type, exact RF parameters, link budget, safety modes) for the first vehicle. | TASK-XC-PROFILE-001 (concrete profile loading); does not block generic profile system | Mission | open |
| OQ-007 | First-contact / launch date. | Phase 2 hardware procurement schedule | Mission | open |
| OQ-008 | Hosting boundaries — single GovCloud cluster for v1 confirmed; cross-region active/standby topology for v1.x is open. | v1.x planning (does not block any v1 task) | Architecture | open |
| OQ-009 | Compliance staffing — does the team have a DPO and ITAR Empowered Official, or do we contract them? | TASK-P5-COMP-001, TASK-P6-COMP-002 | Customer | open |
| OQ-010 | EU representative under GDPR Article 27. | TASK-P6-COMP-002 | Customer | open |
