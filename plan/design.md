# Chetana Space Platform — Implementation Design (v1)

## 0. Document control

| Field | Value |
|---|---|
| Document | `plan/design.md` |
| Version | 1.0 |
| Status | Baseline — locked for v1 implementation |
| Companion docs | `plan/requirements.md`, `plan/todo.md` |

This document specifies **how** the platform is built. Every section traces to one or more requirement IDs in `plan/requirements.md` (notation `→ REQ-…`). No design choice in this document is permitted without a backing requirement; conversely, every requirement maps to at least one section here.

---

## 1. Architecture overview

```
                         ┌────────────────────────────────────────────────┐
                         │         AWS GovCloud (US-East) — v1            │
                         │  region-aware code, single cluster v1          │
                         │                                                │
   ┌──────────────┐      │  ┌─────────────────────────────────────────┐   │
   │  Browser     │──────┼─▶│ chetana-platform Kubernetes namespace   │   │
   │ (SvelteKit)  │ TLS1.3│  │  iam, audit, notify, export, realtime, │   │
   │  + Cesium    │ FIPS │  │  eo-*, gi-*, gs-station, gs-pass-pred,  │   │
   │  + WASM      │      │  │  gs-scheduler                            │   │
   └──────────────┘      │  └────────────┬────────────────────────────┘   │
         │               │               │  Kafka (MSK FIPS)              │
         │ wss://        │               ▼                                │
         │ realtime      │  ┌─────────────────────────────────────────┐   │
         │               │  │ chetana-itar Kubernetes namespace        │   │
         │               │  │  sat-command, sat-conjunction, sat-fsw, │   │
         │               │  │  sat-mission, sat-simulation,           │   │
         │               │  │  sat-telemetry, gs-rf                    │   │
         │               │  │  US-persons-only nodegroup              │   │
         │               │  └────────────┬────────────────────────────┘   │
         │               │               │                                │
         │               │  ┌─────────────────────────────────────────┐   │
         │               │  │ Data plane: Postgres+TimescaleDB,       │   │
         │               │  │ S3 (FIPS), KMS (FIPS), MSK (FIPS)       │   │
         │               │  └─────────────────────────────────────────┘   │
         │               └────────────────────────────────────────────────┘
         │
         │ Public API (mTLS not required for /v1/public/*)
         ▼
   ┌──────────────┐
   │ External     │
   │ customer     │
   │ (API key)    │
   └──────────────┘

   ┌──────────────┐                ┌──────────────────────┐
   │ Real         │ Hardware       │ AWS Ground Station   │
   │ ground       │ Driver         │ (v1.x, GovCloud)     │
   │ dish         │ interface      │                      │
   │ (USRP/RTL/   │                │                      │
   │  custom)     │                │                      │
   └──────────────┘                └──────────────────────┘
```

→ REQ-CONST-003, REQ-CONST-004, REQ-NFR-SEC-001, REQ-FUNC-RT-*.

---

## 2. Repository topology

### 2.1 `chetana-platform` (this repo, post-rebrand + post-extraction)

→ REQ-CONST-004, REQ-CONST-008.

```
chetana-platform/
├── services/
│   ├── proto/                       # buf workspace; non-defense protos (source of truth)
│   ├── packages/                    # shared Go libs; module path p9e.in/chetana/packages
│   ├── iam/                         # NEW
│   ├── audit/                       # NEW
│   ├── notify/                      # NEW
│   ├── export/                      # NEW
│   ├── export-worker/               # NEW
│   ├── realtime-gw/                 # NEW
│   ├── platform-tenants/            # NEW (single-tenant runtime in v1)
│   ├── eo-catalog/                  # extended (STAC HTTP)
│   ├── eo-pipeline/                 # extended (orchestrator + change detection)
│   ├── eo-analytics/                # extended (ML registry + Triton control)
│   ├── gi-{analytics,fusion,predict,reports,tiles}/  # extended
│   ├── gi-workspace/                # NEW
│   ├── gi-aoi/                      # NEW
│   ├── gi-report/                   # NEW
│   ├── gs-station/                  # NEW (station registry, capabilities)
│   ├── gs-scheduler/                # extended (pass FSM, scheduling)
│   ├── gs-pass-pred/                # NEW (TLE manager + SGP4 FFI)
│   ├── gs-ingest/                   # extended (frame ingest fan-out only; demod is in defense gs-rf)
│   └── go.work
├── compute/
│   ├── crates/
│   │   ├── eo-* (atmos-corr, geometric, indices, mosaic, pansharpen, radiometric, sar)
│   │   ├── gi-* (abi, export, fusion-eng, report-render, tile-render)
│   │   ├── gs-antenna/              # generic pointing math
│   │   ├── sim-harness, sim-physics
│   └── Cargo.toml
├── ml/
│   ├── packages/{ml_serving, eo_ml, gi_ml}/
│   └── workers/
├── web/
│   ├── apps/shell/
│   ├── packages/{ui, core, stores, api, proto, wasm, design-tokens, i18n, utility, types, testing, configs}/
│   ├── pnpm-workspace.yaml + turbo.json
├── deploy/
│   ├── helm/chetana-platform/
│   ├── helm/chetana-itar-overlay/   # references defense images by digest
│   ├── docker/                      # local dev
│   ├── terraform/                   # AWS GovCloud infra
├── compliance/
│   ├── classification.yaml
│   ├── itar-paths.txt
│   ├── controls/{iso27001.csv, gdpr.csv, soc2.csv, certin.csv, itar.csv, fedramp-mod.csv}
│   ├── dpia/
│   ├── ropa.md
│   ├── policies/
│   ├── sbom/
│   └── itar-readme.md
├── bench/
│   ├── iam/, realtime/, telemetry/, eo-pipeline/, ml/, stac/
│   └── README.md
├── space_plan/                      # immutable contract docs (input)
├── plan/                            # implementation plan (output)
├── .github/workflows/
├── Taskfile.yml
└── README.md
```

### 2.2 `chetana-defense` (NEW, private, US-persons-only)

→ REQ-COMP-ITAR-001, REQ-COMP-ITAR-002.

```
chetana-defense/
├── services/
│   ├── proto/                       # defense-only protos
│   ├── sat-command/                 # 17-state FSM (D7.9), 2-person, CCSDS encoding
│   ├── sat-conjunction/             # CDM ingest, Pc, screening, maneuver planner
│   ├── sat-fsw/                     # flight software gateway
│   ├── sat-mission/                 # subsystem health, power, thermal, ADCS
│   ├── sat-simulation/              # high-fidelity simulator
│   ├── sat-telemetry/               # decommutator, calibration, limit checker
│   ├── gs-rf/                       # RF link control, waveform selection, demod orchestration
│   └── go.work
├── compute/
│   └── crates/
│       ├── orbit-prop/, gs-pass-pred/, gs-tle/
│       ├── gs-rf-driver/{uhd,librtlsdr,custom}/
│       ├── gs-fec/, gs-bit-sync/, gs-doppler/
├── flight/
│   └── crates/                      # adcs-*, cdh-*, eps-*
├── deploy/
│   └── helm/chetana-itar/           # ITAR-namespace charts
├── compliance/itar-controls/
│   ├── usml-mapping.md
│   ├── us-persons-roster.csv        # encrypted; access-controlled
│   ├── tcp.md                       # technology control plan
│   └── export-records/
├── bench/itar/
├── .github/workflows/               # self-hosted runners in GovCloud
└── Taskfile.yml
```

### 2.3 Cross-repo contracts

→ REQ-CONST-004.

| Boundary | Mechanism | Versioning |
|---|---|---|
| Platform protos consumed by defense | Buf BSR module `buf.build/chetana/platform`; published from `chetana-platform/services/proto/` on every release tag. Defense `services/proto/buf.yaml` declares it as a dep. | SemVer; breaking changes go through deprecation cycle. |
| Defense protos consumed by platform | Buf BSR module `buf.build/chetana/defense-iface`; **interface-only** — RPC method names, Kafka topic names, JWT claim names, NO restricted message bodies. Restricted bodies stay private to defense. | SemVer. |
| Shared Go packages | `p9e.in/chetana/packages` published via internal Go module proxy from `chetana-platform/services/packages/`. Defense pins versions in `go.mod`. | SemVer. |
| Public-knowledge Rust crates | `compute/crates/sim-physics`, `compute/crates/sim-harness`, the generic `flight/crates/cdh-ccsds` framing helpers (the protocol is a published CCSDS standard) live in **platform**, published via internal Cargo registry. | SemVer. |
| ITAR Rust crates | Stay in defense, never published. | Internal git deps. |
| Container images | Platform → ECR `chetana-platform/*`; defense → ECR `chetana-defense/*` in the US-persons-only AWS account. | SHA-pinned in Helm. |
| Helm umbrella | Lives in `chetana-platform/deploy/helm/`. ITAR overlay adds defense images by digest, deploys to `chetana-itar` namespace. | SemVer per chart. |
| Local dev (cleared engineer) | Sibling clones; defense `Taskfile.yml dev-fullstack` uses Go `replace` / Cargo `[patch]` / pnpm `link:` to point at the local platform clone. | n/a |
| Local dev (non-cleared engineer) | Clones platform only; pulls pre-built defense images from ECR for compose stack. | n/a |
| Integration tests across the line | Run in defense CI (which has access to both); platform CI runs platform-only + smoke against last-released defense images. | n/a |
| Release coordination | Compatibility matrix in defense `RELEASES.md`; each platform release pins a defense version range and vice versa. | n/a |

### 2.4 ITAR subtree manifest

→ REQ-COMP-ITAR-001, OQ-005.

`compliance/itar-paths.txt` (proposed; awaiting OQ-005 sanity check):

```
# USML Cat XV (Spacecraft Systems & Associated Equipment) controlled
# After repo split these paths live entirely in chetana-defense; this
# file in chetana-platform documents what was extracted and serves as
# a CI guard that no ITAR-controlled code is reintroduced into the
# platform repo.

flight/**
services/sat-command/**
services/sat-conjunction/**
services/sat-fsw/**
services/sat-mission/**
services/sat-simulation/**
services/sat-telemetry/internal/itar/**
compute/crates/orbit-prop/**
compute/crates/gs-pass-pred/**
compute/crates/gs-tle/**
compute/crates/gs-rf-driver/**
compute/crates/gs-fec/**
compute/crates/gs-bit-sync/**
compute/crates/gs-doppler/**
services/proto/space/satellite/**
services/proto/space/conjunction/**
deploy/helm/chetana-itar/**
```

CI guard in `chetana-platform`: a workflow step diffs every PR against this file and fails if any path matches.

---

## 3. Service catalog

### 3.1 Platform plane (lives in `chetana-platform`)

| Service | Phase | Plan trace | Key responsibilities |
|---|---|---|---|
| `iam` | 1 | D7.10, US-PLT-001..045 | Login, MFA (TOTP, WebAuthn), OIDC issuer, OAuth2, RBAC + ABAC engine, sessions, API keys, SAML, password reset, GDPR SAR/erasure, JWT signing with FIPS-validated keys |
| `audit` | 1 | US-PLT-036..040, ITAR | Append-only `audit_events` with SHA-256 hash chain, search, signed export, integrity verifier, hot/warm/cold tiers (5y online + 7y archived) |
| `notify` | 1 | US-PLT-041..045 | Email (SES FIPS), SMS (SNS FIPS), in-app push via Kafka → realtime-gw, Handlebars templates with versioning, preferences |
| `export` | 1 | US-CMN-007..012 | Async export job API, presigned URL generation, auto-cleanup |
| `export-worker` | 1 | US-CMN-007..012 | Background processor: query → S3 multipart → mark complete |
| `realtime-gw` | 1 | D7.3, US-GS-022/032/041, US-EO-017 | WebSocket gateway: JWT auth, RBAC + ABAC per topic, Kafka → WS fan-out via Redis Pub/Sub, backpressure, heartbeat |
| `platform-tenants` | 1 | US-PLT-016..025 | Tenant config, branding, security policy, quotas (single-tenant runtime in v1; data model already supports multi) |
| `eo-catalog` | 3 | D7.6, US-EO-001..012 | STAC 1.0.0 + OGC API Features REST + ConnectRPC; CQL2 parser; H3 spatial index |
| `eo-pipeline` | 3 | D7.8, US-EO-013..022 | Job orchestrator, scene-pair selection, pipeline stages calling `compute/crates/eo-*`, change-detection workflow |
| `eo-analytics` | 3 | D7.7, US-EO-023..030 | ML registry (MLflow-style schema), Triton control plane, canary/shadow deployment, ONNX conversion intake |
| `gi-tiles` | 4 | US-GEO-003 | WMS / WMTS / MVT tile server using `compute/crates/gi-tile-render` |
| `gi-fusion` | 4 | US-GEO-021..028 | Spatial fusion engine for analyses |
| `gi-predict` | 4 | US-GEO-019, ML | Forecasting service (crop, risk, urban) wrapping `ml/packages/gi_ml` |
| `gi-reports` | 4 | US-GEO-029..036 | (legacy name; `gi-report` is the new authoritative name) |
| `gi-report` | 4 | US-GEO-029..036 | WYSIWYG report editor backend, PDF/DOCX/PPTX/HTML export, share links, version history |
| `gi-workspace` | 4 | US-GEO-001..010 | Workspaces, members, layers, saved views, annotations, activity audit |
| `gi-aoi` | 4 | US-GEO-011..020 | AOI polygons in PostGIS, monitoring rules, AOI alerts, timeline |
| `gi-analytics` | 4 | US-GEO-021..028 | Analysis dispatcher (counting, tracking, terrain, heatmap, spatial query, custom scripts) |
| `gs-station` | 2 | US-GS-009..016 | Station registry, antenna config, location, capabilities, maintenance windows, health rollup |
| `gs-pass-pred` | 2 | D7.4, US-GS-005..008/017..028 | TLE manager (Space-Track + Celestrak), FFI to `compute/crates/orbit-prop`, pass + Doppler RPCs |
| `gs-scheduler` | 2 | D7.2, US-GS-019..028 | Pass scheduling, conflict resolver, antenna reservation, pass FSM |
| `gs-ingest` | 2 | D7.1, US-GS-039..043 | Frame fan-out from Kafka topic `telemetry.frames` to TimescaleDB; cross-namespace Kafka consumer (the demod that produces frames lives in defense `gs-rf`) |

### 3.2 Defense plane (lives in `chetana-defense`)

| Service | Phase | Plan trace | Key responsibilities |
|---|---|---|---|
| `sat-command` | 2 | D7.9, US-GS-049..058 | 17-state command FSM, 2-person approval, CCSDS TC encoding, ACK + verification, retry, audit chain, hazard classification |
| `sat-conjunction` | 4 | D7.5, US-SAT-023..032 | CDM ingest from Space-Track, screening, Pc, encounter geometry, B-plane, maneuver planner, alerts |
| `sat-fsw` | 2 | US-SAT-001..010 | Flight software gateway (host-side); HK telemetry packetisation; uplink/downlink session control |
| `sat-mission` | 4 | US-SAT-001..012 | Subsystem catalog, health rule engine, power budget, battery model, ADCS mode, thermal, mission timeline, anomaly tracking |
| `sat-simulation` | 2 | US-SAT-013, REQ-FUNC-SAT-013 | High-fidelity 6-DOF simulator with all profile combos; replay support |
| `sat-telemetry` | 2 | D7.1, US-GS-039..048 | Decommutation, calibration, limit checker, anomaly publisher; writes to `telemetry_samples` hypertable; emits to Kafka `telemetry.params` |
| `gs-rf` | 2 | US-GS-029..038 | RF link control: tunes SDR via `gs-rf-driver`, configures demod chain (`gs-bit-sync` + `gs-fec` + `gs-doppler`), produces `telemetry.frames` Kafka events |

### 3.3 Plan-boundary alignment for the seven ground-station services

→ REQ-FUNC-GS-BOUNDARY-001, REQ-FUNC-GS-BOUNDARY-002, conversation Q10, `space_plan/docs/README.md`.

| Plan service | RPC count (plan) | Implemented as |
|---|---|---|
| `SatelliteService` | 8 | Service implementation in `chetana-defense/services/sat-mission` (catalog + TLE bookkeeping); facade thin client in `chetana-platform/services/proto/space/satellite/v1/satellite.proto` reflecting unrestricted RPCs |
| `GroundStationService` | 7 | `chetana-platform/services/gs-station` |
| `PassService` | 9 | `chetana-platform/services/gs-scheduler` (split into `pass-pred` for prediction + `pass-exec-fsm` for execution; same service binary, two RPC groupings) |
| `TelemetryService` | 6 | `chetana-defense/services/sat-telemetry` (defense-side, ITAR namespace) + `chetana-platform/services/gs-ingest` (write-side fan-out only) |
| `CommandService` | 8 | `chetana-defense/services/sat-command` |
| `AnomalyService` | 6 | New facade in `chetana-platform/services/gs-station/internal/anomaly/`; back-end logic in `sat-telemetry` and `sat-mission` |
| `AlertService` | 8 | New facade in `chetana-platform/services/notify/internal/alert/` (because alerts are routed through the notify pipeline) |
| **Total** | **52 RPCs** | |

---

## 4. Cross-cutting concerns

### 4.1 Identity & authorization

→ REQ-FUNC-PLT-IAM-*, REQ-FUNC-PLT-AUTHZ-*, REQ-NFR-SEC-001.

#### 4.1.1 Token model

JWT (RS256) with claims:

```jsonc
{
  "iss": "https://iam.chetana.p9e.in",
  "sub": "<user_id>",
  "aud": "<service_name | api>",
  "exp": "<unix>", "iat": "<unix>", "jti": "<uuid>",
  "tenant_id": "<tenant>",
  "is_us_person": true,
  "clearance_level": "internal | restricted | cui | itar",
  "nationality": "US",
  "roles": ["ops.admin", "telemetry.viewer"],
  "scopes": ["telemetry:read", "command:submit"],
  "session_id": "<uuid>",
  "amr": ["pwd", "totp"]
}
```

#### 4.1.2 ABAC decision

A request is authorised iff **all four** are true:

1. The principal's roles grant a permission matching `{module}.{resource}.{action}` (with wildcard).
2. `principal.clearance_level` ≥ `resource.data_classification` (ordered: `public < internal < restricted < cui < itar`).
3. If `resource.data_classification == "itar"` then `principal.is_us_person == true`.
4. No deny policy applies (deny-wins).

The decision logic lives in `services/packages/authz/decision.go` and is the **single** path for authorization across all services. Every service interceptor calls it; no service implements its own check.

→ REQ-CONST-011 (no duplication).

#### 4.1.3 FIPS crypto

| Subsystem | Choice | File |
|---|---|---|
| Go services | `GOEXPERIMENT=boringcrypto` build mode + restricted `crypto/tls` cipher list | `services/packages/crypto/fips.go`; build constraint `//go:build boringcrypto` |
| Rust services | `rustls` with FIPS provider (`aws-lc-rs`) | `Cargo.toml` features `rustls-fips` |
| Python (ML) | OpenSSL FIPS module via system OpenSSL | container base image |
| JS (browser) | Web Crypto API only; no `node-forge` or non-FIPS libs | `web/packages/wasm/crates/crypto/` for crypto primitives |
| Postgres | FIPS-validated PostgreSQL build (or RDS Postgres in FIPS mode) | Helm chart parameter |
| Kafka | MSK with TLS 1.2+ FIPS-mode | Terraform |
| KMS | AWS KMS FIPS endpoint | service config |

A startup self-check in every Go binary asserts FIPS mode (`boring.Enabled()` returns true). Failure → exit 1.

### 4.2 Audit

→ REQ-FUNC-PLT-AUDIT-001..006.

```sql
-- chetana-platform/services/audit/db/schema/0001_init.sql
CREATE TABLE audit_events (
  id BIGSERIAL PRIMARY KEY,
  event_time TIMESTAMPTZ NOT NULL DEFAULT now(),
  actor_user_id UUID,
  actor_principal JSONB NOT NULL,           -- snapshot of principal claims
  action TEXT NOT NULL,                     -- e.g. "command.submit"
  resource_kind TEXT NOT NULL,
  resource_id TEXT NOT NULL,
  data_classification TEXT NOT NULL,        -- public|internal|restricted|cui|itar
  outcome TEXT NOT NULL,                    -- allowed|denied|error
  reason TEXT,
  payload JSONB,                            -- domain-specific event detail
  prev_hash BYTEA NOT NULL,
  this_hash BYTEA NOT NULL,
  ingest_node TEXT NOT NULL,
  region TEXT NOT NULL
);
CREATE INDEX audit_events_event_time_idx ON audit_events (event_time DESC);
CREATE INDEX audit_events_actor_idx ON audit_events (actor_user_id);
CREATE INDEX audit_events_action_idx ON audit_events (action);
CREATE INDEX audit_events_resource_idx ON audit_events (resource_kind, resource_id);
CREATE INDEX audit_events_payload_gin ON audit_events USING gin (payload jsonb_path_ops);
```

Hash chain: `this_hash = SHA256(prev_hash || canonicalize(row_without_hashes))`.

The audit service is the **only** writer; other services emit `audit.events.v1` Kafka messages. Direct DB writes to `audit_events` are revoked at the Postgres role level for all other services. Verifier tool (`services/audit/cmd/verify`) walks the chain and reports the first broken link.

Retention tiering: monthly partition rotation; partitions older than 5 y are detached and uploaded to S3 Glacier with a manifest. A separate verifier validates archived partitions.

### 4.3 Real-time gateway

→ REQ-FUNC-RT-001..006.

```
            Kafka topics                Redis Pub/Sub                Browser WS
            -----------                 -------------                ----------

services ─▶ telemetry.params ─┐
services ─▶ pass.state ───────┤        rt.fanout.<topic>
services ─▶ alert.* ──────────┼─▶ realtime-gw ─▶ Redis ─▶ realtime-gw replicas ─▶ wss://
services ─▶ command.state ────┘
```

`realtime-gw` design:

- One pod per cluster zone; sticky sessions are **not** required because Redis Pub/Sub fans out to all replicas.
- JWT auth on connect (query string or `Authorization` header).
- Subscribe message: `{ "id": "<seq>", "type": "subscribe", "topic": "telemetry.satellite.42.power" }`.
- ABAC check on subscribe per REQ-FUNC-RT-003. Reject with code `4001` (auth) or `4002` (forbidden).
- Per-connection ring buffer; drop-oldest at 1000 msg/s/topic; overflow event emitted to client as `{type: "overflow", dropped_count: N, topic: "..."}`.
- Heartbeat: 30 s ping; disconnect on missed pong.
- Backpressure metrics emitted to Prometheus: `realtime_gw_queue_depth_bytes`, `realtime_gw_drops_total`.

### 4.4 Hardware abstraction

→ REQ-FUNC-GS-HW-001..006.

Three-layer abstraction:

```
┌────────────────────────────────────────────┐
│  services/gs-rf  (Go, defense plane)       │
│  ─ business logic, scheduling, mode mgmt   │
└────────┬───────────────────────────────────┘
         │ Go interface in services/packages/hardware/
         │   HardwareDriver: Tune, SetGain, RxIQ, TxIQ, TxStream
         │   AntennaController: SetAzEl, GetAzEl, SetTrack
         │   GroundNetworkProvider: AllocateContact, ReleaseContact
┌────────▼───────────────────────────────────┐
│  Adapter package (Go shim around Rust FFI) │
│  services/packages/hardware/{uhd,rtl,…}/   │
└────────┬───────────────────────────────────┘
         │ CGO or out-of-process gRPC sidecar
┌────────▼───────────────────────────────────┐
│ Rust adapter crates (defense plane):       │
│  compute/crates/gs-rf-driver/{uhd,         │
│    librtlsdr, custom}                      │
│  compute/crates/gs-antenna/{hamlib,        │
│    gs232, custom}                          │
└────────────────────────────────────────────┘
```

All three SDR adapters and all three rotator adapters MUST be production-grade — no stubs. → REQ-CONST-010.

`GroundNetworkProvider` adapters:

- `own-dish` — direct hardware (uses the SDR + rotator adapters above).
- `aws-gs` — AWS Ground Station (replaces Azure Orbital pending OQ-001).
- KSAT / SSC adapters are v2.0; the interface is in v1 but the implementations are not.

### 4.5 Spacecraft profile system

→ REQ-FUNC-SAT-001.

```protobuf
// services/proto/space/satellite/v1/profile.proto
syntax = "proto3";
package chetana.satellite.v1;

message SpacecraftProfile {
  string profile_id = 1;                      // ULID
  string spacecraft_id = 2;                   // NORAD or internal ID
  string bus_type = 3;                        // free text; e.g. "LEO-100kg-3U"
  repeated Band bands = 4;
  repeated Modulation modulations = 5;
  repeated CcsdsProfile ccsds_profiles = 6;
  LinkBudget link_budget = 7;
  repeated SafetyMode safety_modes = 8;
  repeated Subsystem subsystems = 9;
  string itar_classification = 10;            // "public" or "itar"
  string version = 11;                        // SemVer
  google.protobuf.Timestamp effective_at = 12;
}

enum Band { BAND_UNSPECIFIED = 0; UHF = 1; S = 2; X = 3; }
enum Modulation { MOD_UNSPECIFIED = 0; BPSK = 1; QPSK = 2; OQPSK = 3; PSK_8 = 4; GMSK = 5; }
enum CcsdsProfile { CCSDS_UNSPECIFIED = 0; TM_TF = 1; TC_TF = 2; AOS = 3; USLP = 4; }

message LinkBudget { /* eirp_dbw, antenna_gain_db, noise_temp_k, … */ }
message SafetyMode { /* name, criteria_proto, recovery_action */ }
message Subsystem { /* power, adcs, cdh, comms, thermal, propulsion, payload, structure */ }
```

Loaded by `chetana-defense/services/sat-mission/internal/profile/loader.go` and consumed by all defense services. The non-restricted subset is replicated into platform-side `gs-pass-pred`, `gs-station`, and the web UI.

### 4.6 Classification & data labeling

→ REQ-CONST-013, REQ-COMP-ITAR-002.

Every Postgres table: `data_classification TEXT NOT NULL DEFAULT 'internal' CHECK (data_classification IN ('public','internal','restricted','cui','itar'))`. Default trigger sets it from the principal's clearance on insert.

Every Kafka topic name: prefixed by classification when restricted, e.g. `telemetry.satellite.<id>.itar.power`. ACLs in MSK match the prefix.

Every container image: `LABEL org.chetana.classification=...`. Admission controller (Kyverno) verifies images deployed to `chetana-itar` namespace carry `=itar`.

Every API response JSON: top-level field `_meta.classification` for the most-restrictive datum in the payload. Serializers in `services/packages/api/` enforce this.

### 4.7 FIPS crypto policy

See §4.1.3.

### 4.8 Multi-region readiness

→ REQ-NFR-SCALE-003, REQ-CONST-009.

Even though v1 is single-region:

- Every service reads `CHETANA_REGION` from env. Default `us-gov-east-1`.
- Database connections, S3 buckets, Kafka clusters are addressed by region — `region.PostgresDSN()`, `region.S3Bucket(prefix)`, `region.KafkaBootstrap()` helpers in `services/packages/region/`.
- Kafka topics carry a region prefix `<region>.<topic>` only when cross-region replication is enabled (v1.x).
- Helm overlays exist for `us-gov-east-1` (active in v1), `eu-central-1` (template exists; not deployed), `ap-south-1` (template; not deployed).
- Audit events carry a `region` column.

Adding a region in v1.x = `terraform apply` of a new VPC + EKS cluster + Helm install + cross-region replication setup. No code changes.

---

## 5. Data architecture

### 5.1 Database per service

→ Existing convention.

Each service has its own logical Postgres database; physical Postgres instance shared. Migrations executed by Atlas (chosen for declarative HCL + drift detection); migration runner is a pre-deploy Helm hook job.

TimescaleDB extension is enabled cluster-wide. Hypertables created by services that need them:

| Service | Hypertable | Time column | Partition interval | Continuous aggregates |
|---|---|---|---|---|
| `sat-telemetry` | `telemetry_samples` | `sample_time` | 1 day | 1-min, 1-h |
| `audit` | `audit_events` | `event_time` | 1 month | none (audit kept raw) |
| `eo-pipeline` | `processing_job_events` | `event_time` | 1 day | 1-h (job-rate dashboard) |
| `realtime-gw` | (no DB; ephemeral) | — | — | — |

### 5.2 Object storage

Single S3 bucket per region with prefix-per-domain:

```
s3://chetana-data-us-gov-east-1/
├── stac/items/{collection_id}/{item_id}/...
├── stac/assets/{collection_id}/{item_id}/{asset}.tif
├── exports/{export_id}/{filename}
├── audit-archive/{partition_yyyymm}.parquet
├── ml-models/{model_id}/{version}/{onnx,trt,metadata.json}
├── tiles/{tileset_id}/{z}/{x}/{y}.png
└── reports/{report_id}/{version}.{pdf,docx,pptx,html}
```

Bucket policy: TLS-only; SSE with KMS-FIPS; bucket-key enabled; versioning on; MFA-delete on `audit-archive/*`.

### 5.3 Kafka topic taxonomy

| Topic | Producer | Consumers | Partitions | Retention | Classification |
|---|---|---|---|---|---|
| `telemetry.frames` | `gs-rf` (defense) | `sat-telemetry`, `gs-ingest` | 32 | 7 d | itar |
| `telemetry.params` | `sat-telemetry` | `realtime-gw`, `gs-ingest`, anomaly engines | 32 | 7 d | itar |
| `telemetry.alerts` | `sat-telemetry` | `notify`, `realtime-gw` | 8 | 30 d | itar |
| `pass.state` | `gs-scheduler` | `realtime-gw`, audit | 16 | 30 d | restricted |
| `command.state` | `sat-command` | `realtime-gw`, audit | 16 | 30 d | itar |
| `audit.events.v1` | every service | `audit` | 32 | 90 d (then raw archived) | varies (carries label) |
| `notify.outbound` | every service | `notify` | 16 | 7 d | varies |
| `eo.items.created` | `eo-catalog` | `gi-aoi`, subscribers | 16 | 30 d | varies |
| `eo.jobs.events` | `eo-pipeline` | `realtime-gw`, audit | 16 | 30 d | internal |
| `ml.inference.requests` | `eo-analytics` | Triton (via FastAPI gateway) | 32 | 1 d | varies |
| `conjunction.events` | `sat-conjunction` | `notify`, `realtime-gw`, audit | 8 | 90 d | itar |

ACLs enforced in MSK: defense services and `realtime-gw`/`audit`/`notify` only for ITAR-classified topics.

### 5.4 Retention matrix

| Data class | Hot | Warm | Cold |
|---|---|---|---|
| Audit (REQ-FUNC-PLT-AUDIT-003) | 5 y in `audit_events` (TimescaleDB) | — | 7 y in S3 Glacier as Parquet |
| Telemetry raw (REQ-FUNC-GS-TM-003) | 7 d in hypertable | 90 d (1-min) | 5 y (1-h) |
| STAC items | unlimited (bucket-versioned) | — | — |
| Processing-job events | 90 d in hot | — | — |
| Realtime messages | 0 (ephemeral) | — | — |
| ML model artifacts | unlimited (versioned) | — | — |
| Reports | unlimited | — | — |

---

## 6. Web architecture

→ REQ-CONST-005, REQ-CONST-008, conversation Q14, web audit findings.

### 6.1 Single SvelteKit app

`web/apps/shell/` is the only app. The `dev:all` script in `web/package.json` is rewritten to reference only this app; the previous Samavāya MFE references (`@samavāya/identity`, `@samavāya/masters`, `@samavāya/finance`) are removed.

### 6.2 Route registry

The existing `[domain]/[entity]` generic route pattern (`web/apps/shell/src/routes/(app)/[domain]/[entity]/+page.svelte`) is kept. The `DOMAIN_MODULES` registry in `web/apps/shell/src/lib/modules/index.ts` is **replaced**:

```ts
// Removed (ERP):
// asset, audit-erp, banking, budget, communication, data, finance,
// fulfillment, hr, inventory, manufacturing, masters, projects,
// purchase, sales, workflow

// Kept (renamed, generic platform):
import { identity } from './identity/index.js';
import { audit } from './audit/index.js';
import { notifications } from './notifications/index.js';
import { platform } from './platform/index.js';

// Added (space):
import { mission } from './mission/index.js';
import { satellite } from './satellite/index.js';
import { groundstation } from './groundstation/index.js';
import { pass } from './pass/index.js';
import { telemetry } from './telemetry/index.js';
import { command } from './command/index.js';
import { conjunction } from './conjunction/index.js';
import { eo } from './eo/index.js';
import { gi } from './gi/index.js';
import { aoi } from './aoi/index.js';
import { report } from './report/index.js';
import { alert } from './alert/index.js';

export const DOMAIN_MODULES = {
  // platform
  identity, audit, notifications, platform,
  // satellite
  mission, satellite, conjunction,
  // ground station
  groundstation, pass, telemetry, command,
  // earth observation
  eo,
  // geoint
  gi, aoi, report,
  // cross-cutting
  alert,
};
```

### 6.3 Route groups

```
web/apps/shell/src/routes/
├── (app)/                   # authenticated ops + analyst surface
│   ├── [domain]/[entity]/   # generic schema-driven CRUD (kept)
│   ├── [domain]/            # domain landing redirect (kept)
│   ├── dashboard/
│   ├── globe/               # Cesium 3D globe + ground tracks
│   ├── pass/[passId]/       # live pass execution view
│   ├── command/queue/       # command queue + 2-person approval
│   ├── telemetry/[satId]/   # telemetry strip charts, limit overlays
│   ├── catalog/             # STAC search bar, footprint browser
│   ├── pipeline/[jobId]/    # processing job detail
│   ├── workspace/[wsId]/    # GeoInt workspace canvas
│   ├── aoi/[aoiId]/         # AOI detail + timeline
│   ├── conjunction/         # conjunction list + B-plane viz
│   ├── report/[reportId]/   # report editor
│   └── settings/            # user, MFA, sessions, API keys
├── (auth)/                  # unauthenticated
│   ├── login/
│   ├── mfa/
│   ├── reset-password/
│   └── sso-callback/
└── (public)/                # public customer portal (Phase 5)
    ├── catalog/             # public STAC search (read-only)
    ├── subscribe/           # AOI subscription
    └── docs/                # public API docs + DOI citation
```

### 6.4 Cesium integration

→ REQ-FUNC-SAT-004.

- Dependency: `@cesium/engine` (no `cesium` umbrella package — engine-only is leaner).
- Cesium ion access token NOT used (FedRAMP/ITAR concern); host Cesium World Terrain mirror in S3 (or use built-in WGS84 ellipsoid only if terrain isn't required for ITAR-cleared work).
- Components in `web/packages/ui/src/space/orbit/`: `<CesiumGlobe/>`, `<GroundTrack2D/>`, `<GroundTrack3D/>`, `<OrbitVisualizer/>`, `<FootprintOverlay/>`.
- 2D map needs are met by Cesium Columbus view (`SceneMode.COLUMBUS_VIEW`).
- Bundle splitting: Cesium loaded only on routes that need it (`/globe`, `/pass/*`, `/conjunction`, `/aoi/*`).

### 6.5 WASM kernels

→ REQ-FUNC-SAT-005, web audit findings.

`web/packages/wasm/crates/` retains:

- `core` — decimal, date, JSON
- `crypto` — HMAC, SHA-256, JWT verification only (FIPS-via-WebCrypto for actual signing)
- `i18n` — date/number formatting
- `offline` — IndexedDB queue
- `validation` — generic validators

Removed (move to `archived/` or separate repo):

- `barcode`, `bom`, `compliance`, `depreciation`, `ledger`, `payroll`, `pricing`, `tax-engine`

Added in Phase 2/3:

- `sgp4` — wraps `compute/crates/orbit-prop` for `wasm32-unknown-unknown` target
- `coordinates` — ECEF↔geodetic, TEME↔J2000, horizon coords
- `image-preview` — COG tile slice + JPEG/PNG decode + histogram

The `compute/` crates that target both `host` and `wasm32` are configured via Cargo features:

```toml
# compute/crates/orbit-prop/Cargo.toml
[features]
default = ["host"]
host = []
wasm = ["wasm-bindgen"]
```

### 6.6 Module-by-module UI scope

(Detailed component inventory in `plan/todo.md` Phase 2/3/4 task lists.)

---

## 7. Deployment topology

→ REQ-CONST-003, REQ-NFR-REL-003.

### 7.1 Single GovCloud cluster v1

Region: `us-gov-east-1`. EKS cluster `chetana-prod-east1`. Two node groups:

- `general` — t3.large / m6i.large, taints none, label `tier=general`.
- `itar-eligible` — m6i.xlarge / g6.xlarge (for Triton), taint `itar-only=true:NoSchedule`, label `itar-eligible=true`. Operator IAM access restricted to US-persons admin group.

Two namespaces:

- `chetana-platform` — runs all platform-plane services. Deployments tolerate `tier=general` only.
- `chetana-itar` — runs all defense-plane services. Deployments require `itar-only=true` toleration and `itar-eligible=true` node selector.

NetworkPolicies: default-deny per namespace; explicit allow rules per service. ITAR-namespace egress restricted to MSK (FIPS), RDS (FIPS), KMS, Secrets Manager.

### 7.2 Helm umbrella

`deploy/helm/chetana/`:

- One umbrella chart, parameterised per service.
- Each service template: Deployment, Service, ConfigMap, ServiceAccount, HPA, PDB, NetworkPolicy, ServiceMonitor (Prometheus), PodMonitor (if needed).
- ITAR overlay: `deploy/helm/chetana-itar-overlay/values.yaml` adds defense-image references by digest, namespace override, taint toleration.

### 7.3 ITAR overlay

ITAR overlay applies on top of base umbrella:

```yaml
# deploy/helm/chetana-itar-overlay/values.yaml
namespace: chetana-itar
nodeSelector:
  itar-eligible: "true"
tolerations:
  - key: itar-only
    operator: Equal
    value: "true"
    effect: NoSchedule
images:
  sat-command: <ECR-defense>:<digest>
  sat-conjunction: <ECR-defense>:<digest>
  # ... etc.
networkPolicy:
  itarNamespaceOnly: true
```

### 7.4 Multi-region readiness hooks

- Helm value `region` consumed by every chart.
- Terraform module `terraform/modules/region/` parameterised by region; instantiating a new region = new module call.
- Kafka MirrorMaker 2 config templated; not active in v1.

---

## 8. CI / CD pipeline

→ REQ-NFR-SEC-004, REQ-NFR-SEC-005, REQ-NFR-SEC-006.

### 8.1 Lint + test + build matrix (every PR)

In `chetana-platform/.github/workflows/ci.yml`:

| Stage | Tool | Failure policy |
|---|---|---|
| Lint Go | `gofmt`, `goimports`, `golangci-lint` (gosec enabled) | block merge |
| Lint Rust | `cargo fmt --check`, `cargo clippy -- -D warnings` | block merge |
| Lint TS | `prettier --check`, `eslint`, `svelte-check` | block merge |
| Lint Python | `ruff check`, `mypy --strict` | block merge |
| Lint Helm | `helm lint`, `kubeval`, `polaris` | block merge |
| Unit tests Go | `go test ./...` (race + cover) | block merge; coverage ≥ 80% required (REQ verify) |
| Unit tests Rust | `cargo test --workspace --all-targets` | block merge |
| Unit tests TS | `vitest run --coverage` | block merge; coverage ≥ 80% |
| Unit tests Python | `pytest --cov` | block merge; coverage ≥ 80% |
| Buf lint + breaking | `buf lint`, `buf breaking --against '...^1'` | block merge |
| ITAR path guard | diff against `compliance/itar-paths.txt`; fail if matched | block merge |
| Build images | `docker buildx build` per service; SBOM via `syft`; sign via `cosign` | block merge |
| SAST | `gosec`, `semgrep ci`, `cargo-audit` | block merge on critical |
| SCA | `trivy fs`, `npm audit`, `pip-audit` | block merge on critical |
| DAST (nightly) | `zap-baseline` against staging | block deploy on critical |

### 8.2 Cross-repo coordination

- BSR: platform proto module published on tag; defense proto module published on tag.
- Internal Go module proxy: hosted on a private Athens or `goproxy` instance in GovCloud; both repos consume.
- Internal Cargo registry: hosted on `cargo-quickpub` or AWS CodeArtifact.
- Compatibility matrix in `chetana-defense/RELEASES.md`: each defense release pins a `chetana-platform` SemVer range.

### 8.3 Release process

1. Tag `vX.Y.Z` on `main` of either repo.
2. CI publishes:
   - Buf modules (proto)
   - Go module (packages)
   - Cargo crates (public-knowledge ones)
   - Container images (signed, with SBOM and Cosign attestation)
   - Helm charts (signed)
3. Release manager updates `RELEASES.md` with the matrix entry.
4. Deploy via ArgoCD or `helm upgrade`; ArgoCD verifies signatures.

---

## 9. Compliance design

### 9.1 ISO 27001 control mapping

Lives in `compliance/controls/iso27001.csv` with columns `control_id, title, status, evidence, owner, last_reviewed`. Examples:

- `A.5.16 Identity management` → IAM service implementation + audit log.
- `A.8.15 Logging` → audit hash chain + retention.
- `A.8.24 Use of cryptography` → FIPS policy.
- `A.5.30 ICT readiness for business continuity` → DR drill record.

### 9.2 GDPR controls

- `compliance/dpia/<surface>.md` — one DPIA per high-risk processing surface (the public customer surface; ML inference; AOI monitoring).
- `compliance/ropa.md` — Records of Processing Activities.
- IAM `/v1/me/data-export` (Article 15 + 20) and `/v1/me/erasure` (Article 17) endpoints.
- Breach pager wired to alerting; 1-h internal, 72-h authority.

### 9.3 ITAR controls

- USML categorisation in `chetana-defense/compliance/itar-controls/usml-mapping.md`.
- US-persons roster in `chetana-defense/compliance/itar-controls/us-persons-roster.csv` (encrypted at rest; access logged).
- TCP (Technology Control Plan) in `chetana-defense/compliance/itar-controls/tcp.md`.
- Export records under `chetana-defense/compliance/itar-controls/export-records/<yyyy-mm-dd>-<recipient>.md`.
- Runtime: ITAR namespace + image labels + admission control.
- DDTC registration is v2.0 (REQ-COMP-ITAR-005); the operational controls are v1.

### 9.4 Audit evidence flow

```
service ─emits→ Kafka audit.events.v1 ─consumed by→ audit service
                                                       │
                                                       ▼
                                          Postgres audit_events (hash-chained)
                                                       │
                                          ┌────────────┼─────────────┐
                                          ▼            ▼             ▼
                                  search API     export API     archive job
                                                 (signed)        (Glacier)
```

---

## 10. Coding standards

→ REQ-CONST-010, REQ-CONST-011, REQ-CONST-012, REQ-CONST-013.

### 10.1 No code duplication

- Shared logic lives in `services/packages/` (Go), `compute/crates/` (Rust shared), or `web/packages/` (TS).
- `golangci-lint` enabled for `dupl`, `gocognit`, `funlen`.
- `cargo clippy -- -D warnings` enabled for `clippy::similar_names`, `clippy::cognitive_complexity`.
- Pre-commit hook: `tools/check-duplication.sh` runs `dupl` and fails on any duplicate block ≥ 10 lines.
- Reviewers reject PRs that paste logic instead of factor it.

### 10.2 Bug-fix regression policy

For every bug fix:

1. Add a regression test that **fails** before the fix and **passes** after. The test name MUST reference the bug ID or the original symptom.
2. Identify the **bug class** (e.g. "off-by-one in pagination") and add tests for sibling code paths plausibly affected by the same root cause.
3. PR description includes: symptom, root cause, fix, regression-test list, sibling-paths-checked list.
4. CI fails if a PR labelled `bug-fix` does not add at least one new test file.

### 10.3 Test coverage minimums

- Per-package coverage ≥ 80% (Go, Rust, TS, Python). CI fails below 80% for changed packages.
- Critical paths (auth, command FSM, audit chain, FIPS) ≥ 95%.

### 10.4 Lint rules

- Go: `golangci-lint.yml` enables `gofmt, govet, staticcheck, errcheck, gocritic, gosec, dupl, unused, ineffassign, errorlint, contextcheck, gocognit, funlen, lll, goimports, revive, exhaustive, prealloc`. No allow-comments without justification.
- Rust: `clippy::pedantic` baseline; documented exceptions only.
- TS: `eslint-plugin-svelte`, `eslint-plugin-import`, `@typescript-eslint/recommended`. `no-console` enforced; structured logger only.
- Python: `ruff` (`E, F, B, I, N, UP, S, ANN`), `mypy --strict`.
- Helm: `helm lint`, `polaris`, `kubeval`, `kubeconform`.
- Markdown: `markdownlint` for `plan/`, `compliance/`, `README.md`, etc.

### 10.5 Documentation expectations

- Every Go package has a package comment explaining purpose.
- Every public Rust item has a `///` doc.
- Every TS package has a README.
- Every service has `docs/runbook.md` and `docs/sli-slo.md`.
- ADRs in `docs/adr/NNNN-title.md` for architecturally significant decisions.

---

## 11. Risk register

| ID | Risk | Likelihood | Impact | Mitigation | Owner |
|---|---|---|---|---|---|
| R-AZURE-EOL | Azure Orbital EOL 2026-09 | High | Medium | Switch to AWS GS as second provider (OQ-001) | Architecture |
| R-FIPS-LIB | Some Go/Rust libs lack FIPS-mode crypto path | Medium | High | Phase 0 inventory + substitution; Reject FIPS check at startup | Platform |
| R-US-PERSON-OPS | Insufficient US-person staff for ITAR namespace operations | Medium | High | Bifurcated ops model; US-persons-only on-call rotation for ITAR | Customer |
| R-TRITON | Triton config + dynamic-batching tuning is non-trivial | Medium | Medium | Phase 1 spike (week 8) with dummy ONNX; Phase 3 starts with muscle memory | ML team |
| R-CESIUM-BUNDLE | Cesium bundle size impacts initial load | Medium | Low | Code-split per route; lazy-load | Web |
| R-PLAN-DRIFT | `space_plan/` doc duplicates (`D6.1.md` vs `D6.1 (1).md`) | Low | Low | Identical-by-diff today; freeze canonical names | Platform |
| R-PROTO-DRIFT | Cross-repo proto drift | Medium | High | BSR breaking-change detection in CI | Platform |
| R-HW-DELAY | Real spacecraft hardware delays | High | High | `sat-simulation` keeps software exercised | Mission |
| R-COMP-LOAD | Five-regime compliance overload | High | High | Staggered cert sequence (REQ §5.1–§5.6) | Compliance |
| R-CDM-LATENCY | Space-Track CDM 8–24 h lag | Certain | Medium | Build for it: cache, retry, recompute Pc as TLE freshens | Platform |

---

## 12. Traceability summary

Each requirement in `plan/requirements.md` is implemented by at least one design section above. The mapping is:

| Requirement category | Design sections |
|---|---|
| REQ-FUNC-CMN-* | §3.1, §4.2, §4.3, §5.1 |
| REQ-FUNC-PLT-IAM-* | §4.1, §3.1 (`iam`) |
| REQ-FUNC-PLT-AUDIT-* | §4.2, §3.1 (`audit`) |
| REQ-FUNC-PLT-NOTIFY-* | §3.1 (`notify`) |
| REQ-FUNC-PLT-AUTHZ-* | §4.1.2 |
| REQ-FUNC-PLT-TENANT-* | §3.1 (`platform-tenants`), §4.6 |
| REQ-FUNC-RT-* | §4.3, §3.1 (`realtime-gw`) |
| REQ-FUNC-SAT-001 | §4.5 |
| REQ-FUNC-SAT-005..006 | §3.2 (`gs-pass-pred`), §6.5 |
| REQ-FUNC-SAT-007..008 | §3.2 (`sat-conjunction`) |
| REQ-FUNC-SAT-009..012 | §3.2 (`sat-command`) |
| REQ-FUNC-GS-BOUNDARY-* | §3.3 |
| REQ-FUNC-GS-TM-* | §3.2 (`sat-telemetry`), §5.1, §5.3 |
| REQ-FUNC-GS-HW-* | §4.4 |
| REQ-FUNC-EO-CAT-* | §3.1 (`eo-catalog`) |
| REQ-FUNC-EO-PIPE-* | §3.1 (`eo-pipeline`) |
| REQ-FUNC-EO-ML-* | §3.1 (`eo-analytics`), §5.2 |
| REQ-FUNC-GI-* | §3.1 (`gi-*`) |
| REQ-FUNC-IAAS-* | §6.3 (public route group), §3.1 |
| REQ-NFR-PERF-* | §8 (load-test stages), §11 (NFR risks) |
| REQ-NFR-REL-* | §7 (HPA/PDB/NetworkPolicy in chart) |
| REQ-NFR-SEC-* | §4.1.3, §4.6, §8 (CI stages), §9 |
| REQ-NFR-OBS-* | §4.2, §4.3, §8 |
| REQ-NFR-SCALE-* | §4.8, §7 |
| REQ-COMP-ISO-* | §9.1 |
| REQ-COMP-GDPR-* | §9.2 |
| REQ-COMP-ITAR-* | §2.2, §2.4, §4.6, §7.3, §9.3 |
| REQ-COMP-FEDRAMP-* | §4.1.3, §7, §9 |
| REQ-CONST-* | §1, §2, §6, §10 |
