# Chetana Space Platform — Implementation Requirements (v1)

## 0. Document control

| Field | Value |
|---|---|
| Document | `plan/requirements.md` |
| Version | 1.0 |
| Status | Baseline — locked for v1 implementation |
| Owners | Platform Architecture |
| Inputs | `space_plan/docs/*` (immutable contracts), conversation decisions Q1–Q21 + compliance + GovCloud + ITAR posture |
| Companion docs | `plan/design.md` (how), `plan/todo.md` (when) |

This document compiles the binding contracts in `space_plan/docs/` into a single normative checklist for the v1 build. **It does not override `space_plan/`.** Where this document is silent on a requirement that appears in `space_plan/`, the `space_plan/` version applies. Where a backlog row in `space_plan/D6.2-product-backlog-*.md` is descoped from v1, that descoping is recorded here under §8.

Requirement IDs use the pattern `REQ-{CATEGORY}-{MODULE}-{NNN}`. Every requirement has:

- A binding statement (MUST / SHALL / MUST NOT).
- A traceability link to one or more `space_plan/` documents and/or a conversation decision.
- A target phase (Phase 0–6) and a verification method (test / inspection / audit / load benchmark).

---

## 1. Vision

Chetana is a **single-tenant, generic, real-spacecraft, real-RF, real-ground-network space technology platform** serving four user types simultaneously:

| User type | Primary surface | Auth posture |
|---|---|---|
| Mission ops engineer | Internal ops UI (`web/apps/shell` ops route group) | Authenticated, US-person attribute required for ITAR-controlled telemetry/command paths |
| EO analyst | Internal analyst UI (workspaces, AOIs, processing pipelines, change-detection) | Authenticated; classification-aware |
| GeoInt analyst | Internal analyst UI (workspaces, reports) | Authenticated; classification-aware |
| External imagery-as-a-service customer | Public portal route group + Public REST/STAC API | API key with scope `public-collections:read`; no access to ITAR-controlled data |

The platform is **architected for multi-region** (US-Gov, EU, India) but **deploys to a single AWS GovCloud (US-East) Kubernetes cluster in v1**. Multi-region rollout is staged in v1.x and v2.0.

---

## 2. Stakeholders & access tiers

### 2.1 Personas

- **Ops Engineer (US-person, cleared)**: schedules passes, executes commands, monitors live telemetry, plans maneuvers.
- **EO Analyst (cleared or non-cleared)**: ingests scenes, runs processing pipelines (orthorectification, pan-sharpen, change-detection), publishes products.
- **GeoInt Analyst**: works in spatial workspaces, draws AOIs, runs analyses (counting, tracking, terrain, heatmap, custom scripts), authors reports.
- **External Customer**: queries STAC catalog, subscribes to AOI deliveries, downloads scenes via presigned URLs, cites with DOI.
- **System Admin (US-person, cleared)**: manages users, roles, audit policies, retention, runs DR drills.
- **Compliance Officer**: reviews audit log, exports evidence packages, manages POA&M, runs DPIAs.

### 2.2 Authorization attributes (ABAC)

Every principal carries:

- `role[]` — RBAC role assignments.
- `is_us_person: bool` — required for ITAR-controlled resource access.
- `clearance_level: enum { public, internal, restricted, cui, itar }`.
- `nationality: ISO-3166-alpha-2`.
- `tenant_id` — single-tenant in v1; field exists for forward compatibility.
- `scopes[]` — API key scopes for external customers.

Every protected resource carries `data_classification ∈ { public, internal, restricted, cui, itar }`. Authorization decision = role permission AND classification clearance AND (if classification = itar then is_us_person).

---

## 3. Functional requirements

### 3.1 Common (REQ-FUNC-CMN)

Source: `space_plan/docs/D6.2-product-backlog-common.md` (US-CMN-001 … US-CMN-018). Phase 0–1.

| ID | Requirement | Trace | Phase |
|---|---|---|---|
| REQ-FUNC-CMN-001 | Every service MUST expose `/health` returning HTTP 200 with `{status: "ok", version, git_sha, uptime_s}`. | US-CMN-001 | 0 |
| REQ-FUNC-CMN-002 | Every service MUST expose `/ready` that performs live dependency checks against Postgres, Kafka, Redis (where used), and any required upstream services, with a 5-second result cache. | US-CMN-002 | 0 |
| REQ-FUNC-CMN-003 | Every service MUST expose `/metrics` (Prometheus format) on a separate port (default `:9090`) with at minimum: HTTP/RPC request counts and durations, Kafka consumer lag, DB connection-pool stats, build info. | US-CMN-004 | 0 |
| REQ-FUNC-CMN-004 | The platform MUST aggregate dependency health into a single read endpoint and emit alerts on flap or sustained failure (Slack + email + PagerDuty). | US-CMN-005, US-CMN-006 | 1 |
| REQ-FUNC-CMN-005 | The platform MUST provide an Export Service that accepts asynchronous bulk-export jobs, processes them via background workers, persists progress, generates 24-h presigned S3 URLs, and auto-cleans expired exports. | US-CMN-007 … US-CMN-012 | 1 |
| REQ-FUNC-CMN-006 | The platform MUST provide a distributed Scheduler service supporting cron expressions, manual triggers, Redis-based distributed locking, configurable timeouts, retry policies, execution history, and enable/disable toggles. | US-CMN-013 … US-CMN-018 | 1 |

### 3.2 Platform (REQ-FUNC-PLT)

Source: `space_plan/docs/D6.2-product-backlog-platform.md` (US-PLT-001 … US-PLT-045) and `space_plan/docs/D7.10-deepdive-authentication-authorization.md`. Phase 1.

#### 3.2.1 IAM service

| ID | Requirement | Trace | Phase |
|---|---|---|---|
| REQ-FUNC-PLT-IAM-001 | IAM MUST authenticate users with email + password using Argon2id (memory ≥ 64 MiB, iterations ≥ 3, parallelism ≥ 4). | US-PLT-001, D7.10 | 1 |
| REQ-FUNC-PLT-IAM-002 | IAM MUST issue access tokens (TTL 15 m) and refresh tokens (TTL 7 d) signed with FIPS 140-2/3 validated cryptographic modules using RSA-2048 (or ECDSA-P-256). Refresh tokens are single-use; reuse triggers session invalidation. | US-PLT-003, D7.10 | 1 |
| REQ-FUNC-PLT-IAM-003 | IAM MUST enforce login rate-limiting (10/min/IP, 5 failures/account) using Redis sliding window with progressive lockout (15 min → 1 h → 24 h). | US-PLT-002 | 1 |
| REQ-FUNC-PLT-IAM-004 | IAM MUST support TOTP MFA per RFC 6238 with 160-bit secrets, QR enrolment, ten 8-character backup codes (stored hashed), and ±1 time-step tolerance. | US-PLT-005, US-PLT-006 | 1 |
| REQ-FUNC-PLT-IAM-005 | IAM MUST support WebAuthn Level 2 registration and assertion with sign-count clone detection. | US-PLT-014, US-PLT-015 | 1 |
| REQ-FUNC-PLT-IAM-006 | IAM MUST support OIDC issuer endpoints: `/.well-known/openid-configuration`, `/oauth2/authorize` (PKCE S256 mandatory), `/oauth2/token` (auth-code, refresh, client-credentials grants), `/oauth2/userinfo`, `/.well-known/jwks.json`. | US-PLT-013, D7.10 | 1 |
| REQ-FUNC-PLT-IAM-007 | IAM MUST support SAML 2.0 SP-initiated SSO with JIT user provisioning and configurable attribute → role mapping. | US-PLT-013, D7.10 | 1 |
| REQ-FUNC-PLT-IAM-008 | IAM MUST issue tokens whose claims include `is_us_person`, `clearance_level`, `nationality`, `roles[]`, `scopes[]`, and `tenant_id`. | Conversation Q11, ITAR posture | 1 |
| REQ-FUNC-PLT-IAM-009 | IAM MUST enforce session limits (max 5 concurrent per user), idle timeout (1 h), absolute lifetime (24 h), and revocation. | D7.10 | 1 |
| REQ-FUNC-PLT-IAM-010 | IAM MUST support password reset with cryptographically random 256-bit tokens, hashed storage, 1-h expiry, 3/h rate limit, constant-time response. | US-PLT-008, US-PLT-009 | 1 |
| REQ-FUNC-PLT-IAM-011 | IAM MUST provide GDPR Subject Access Request (SAR) and deletion endpoints per Article 15 / Article 17. | GDPR | 1 |

#### 3.2.2 RBAC + ABAC

| ID | Requirement | Trace | Phase |
|---|---|---|---|
| REQ-FUNC-PLT-AUTHZ-001 | The platform MUST evaluate authorization as `(RBAC permission match) AND (classification clearance ≥ resource classification) AND (is_us_person if resource.classification = itar)`. | ITAR, conversation Q11 | 1 |
| REQ-FUNC-PLT-AUTHZ-002 | RBAC permissions MUST follow the format `{module}.{resource}.{action}` with wildcard support (e.g. `groundstation.pass.*`). | US-PLT-031 | 1 |
| REQ-FUNC-PLT-AUTHZ-003 | The platform MUST support custom roles with priority-ordered, deny-wins policy evaluation. | D7.10 | 1 |
| REQ-FUNC-PLT-AUTHZ-004 | Authorization decisions MUST emit a structured audit event (allow/deny + reason). | ISO 27001 A.8.3, D7.10 | 1 |

#### 3.2.3 Tenant management (single-tenant in v1, but service exists)

| ID | Requirement | Trace | Phase |
|---|---|---|---|
| REQ-FUNC-PLT-TENANT-001 | The platform MUST maintain a single tenant record with name, branding, security policy (MFA-required, session-timeout, password-policy), and quotas. | US-PLT-016 … US-PLT-025 | 1 |
| REQ-FUNC-PLT-TENANT-002 | All service-layer code MUST receive `tenant_id` from the principal context for forward-compatibility with multi-tenant in v1.x. | Conversation Q3 | 1 |
| REQ-FUNC-PLT-TENANT-003 | RLS MUST NOT be enabled on Postgres tables in v1 (single-tenant); the schema MUST include a `tenant_id` column with NOT NULL DEFAULT for forward compatibility. | Conversation Q3 | 1 |

#### 3.2.4 Audit service

| ID | Requirement | Trace | Phase |
|---|---|---|---|
| REQ-FUNC-PLT-AUDIT-001 | The platform MUST log every authentication, authorization, configuration change, command submission, command approval, command execution, telemetry export, AOI access, and report export to an append-only audit store. | US-PLT-036 … US-PLT-040, ISO 27001 A.8.15 | 1 |
| REQ-FUNC-PLT-AUDIT-002 | Audit events MUST form a SHA-256 hash chain (each row carries `prev_hash`); a verifier MUST be able to detect tampering. | US-PLT-039, ITAR | 1 |
| REQ-FUNC-PLT-AUDIT-003 | Audit retention MUST be ≥ 5 years online and ≥ 7 years cold (S3 Glacier / S3 Glacier Deep Archive equivalent in GovCloud). | ITAR (5 y), FedRAMP-Mod | 1 |
| REQ-FUNC-PLT-AUDIT-004 | Audit search MUST support time + actor + action + resource filters and full-text on JSONB payload. | US-PLT-037 | 1 |
| REQ-FUNC-PLT-AUDIT-005 | Audit export MUST support CSV/JSON with chain-signature attestation in the export envelope. | US-PLT-038 | 1 |
| REQ-FUNC-PLT-AUDIT-006 | The audit chain MUST NOT be writable from any service except via the audit-service interceptor; direct DB writes by other services are prohibited. | ISO 27001 A.8.15, ITAR | 1 |

#### 3.2.5 Notification service

| ID | Requirement | Trace | Phase |
|---|---|---|---|
| REQ-FUNC-PLT-NOTIFY-001 | The platform MUST send templated email (AWS SES via FIPS endpoint), SMS (AWS SNS via FIPS endpoint), and in-app notifications with preference-based routing. | US-PLT-041 … US-PLT-045 | 1 |
| REQ-FUNC-PLT-NOTIFY-002 | Templates MUST use a deterministic syntax (Handlebars or equivalent), be versioned, and support variable validation. | US-PLT-045 | 1 |
| REQ-FUNC-PLT-NOTIFY-003 | Security-classified notifications (login, MFA change, password reset) MUST be mandatory and not opt-out-able. | D7.10 | 1 |
| REQ-FUNC-PLT-NOTIFY-004 | SMS notifications MUST honour a 5/h/user rate limit. | US-PLT-044 | 1 |

### 3.3 Real-time gateway (REQ-FUNC-RT)

Source: `space_plan/docs/D7.3-deepdive-websocket-protocols.md`. Phase 1.

| ID | Requirement | Trace | Phase |
|---|---|---|---|
| REQ-FUNC-RT-001 | The platform MUST expose a WebSocket gateway (`wss://…/v1/rt`) that authenticates with a JWT, supports topic subscriptions, and routes messages from Kafka. | D7.3 | 1 |
| REQ-FUNC-RT-002 | Subscription topics MUST follow the hierarchy `{domain}.{entity_id}.{kind}` (e.g. `telemetry.satellite.<sat_id>.<subsystem>`, `pass.<pass_id>.state`, `alert.<severity>`, `command.<cmd_id>.state`). | D7.3 | 1 |
| REQ-FUNC-RT-003 | The gateway MUST enforce per-topic RBAC + ABAC. ITAR-classified topics (e.g. `telemetry.satellite.<sat_id>.itar.*`) MUST require `is_us_person = true`. | ITAR, D7.3 | 1 |
| REQ-FUNC-RT-004 | The gateway MUST send 30 s ping/pong heartbeats and close idle connections. | D7.3 | 1 |
| REQ-FUNC-RT-005 | The gateway MUST apply per-connection backpressure (max 1000 msg/s/topic, drop-oldest on buffer overflow). | D7.3 | 1 |
| REQ-FUNC-RT-006 | The gateway MUST distribute messages across cluster replicas via Redis Pub/Sub (or equivalent) for sticky-session-free horizontal scaling. | D7.3 | 1 |

### 3.4 Satellite & flight (REQ-FUNC-SAT)

Source: `space_plan/docs/D6.2-product-backlog-satellite.md`, `D7.5-deepdive-conjunction-assessment.md`, `D7.9-deepdive-command-execution-verification.md`, `D7.4-deepdive-tle-pass-prediction.md`. Phase 2 + 4.

| ID | Requirement | Trace | Phase |
|---|---|---|---|
| REQ-FUNC-SAT-001 | The platform MUST manage a generic spacecraft profile capturing `bus_type`, `bands[] ∈ {UHF, S, X}`, `modulations[] ∈ {BPSK, QPSK, OQPSK, 8PSK, GMSK}`, `ccsds_profile ∈ {TM_TF, TC_TF, AOS, USLP}`, link budget, safety modes, and subsystem catalog. Service code SHALL read profile to configure runtime behaviour. | Conversation Q3, US-SAT-001…SAT-002 | 2 |
| REQ-FUNC-SAT-002 | The platform MUST monitor subsystem health (power, ADCS, CDH, comms, thermal, propulsion, payload, structure) with rule-based alarming and rollup. | US-SAT-003 … US-SAT-005 | 4 |
| REQ-FUNC-SAT-003 | The platform MUST track power budget (solar input, battery SoC, total load, eclipse prediction), battery health (SoC, voltage, current, temp, cycle count, capacity fade), ADCS mode (attitude quaternion → Euler, wheel speeds, momentum), and thermal map. | US-SAT-006 … US-SAT-009 | 4 |
| REQ-FUNC-SAT-004 | The platform MUST visualise orbit state (ECI position/velocity, Keplerian elements), 2D ground track, 3D orbit (Cesium) with day/night terminator and 24-h propagation. | US-SAT-013, US-SAT-014, Cesium choice | 4 |
| REQ-FUNC-SAT-005 | The platform MUST propagate orbit using SGP4/SDP4 implemented in Rust (`compute/crates/orbit-prop`); the same crate MUST compile for `wasm32-unknown-unknown` for browser-side propagation. | D7.4, conversation Q7 | 2 |
| REQ-FUNC-SAT-006 | The platform MUST refresh TLEs from Space-Track every 6 h with retry/backoff, persist history, and expose freshness alerts. | D7.4 | 2 |
| REQ-FUNC-SAT-007 | The platform MUST ingest CCSDS 508.0-B-2 CDMs from Space-Track every 8 h, propagate covariance to TCA, compute Pc via the Foster method (or equivalent peer-reviewed algorithm), classify risk (green < 1e-7, yellow 1e-7…1e-5, orange 1e-5…1e-4, red ≥ 1e-4), and emit alerts. | D7.5 | 4 |
| REQ-FUNC-SAT-008 | The platform MUST plan collision-avoidance maneuvers (along/cross/radial options), check secondary conjunctions, and present trade-off visualisations including B-plane geometry. | US-SAT-027, US-SAT-028, D7.5 | 4 |
| REQ-FUNC-SAT-009 | Command execution MUST follow a 17-state finite-state machine (created → validating → queued → scheduled → uplinking → sent → acked → verifying → completed; plus failed, cancelled, rejected, rolled-back paths) per `D7.9`. Every transition MUST be logged. | D7.9 | 2 |
| REQ-FUNC-SAT-010 | Commands MUST be classified by hazard level (`safe` auto-approves; `caution` and `critical` require two-person approval where the approver is a different US-person principal with `command.approve` permission). | D7.9 | 2 |
| REQ-FUNC-SAT-011 | Commands MUST be encoded as CCSDS TC Transfer Frames (using `flight/crates/cdh-ccsds`), include sequence numbers, CRC-16-CCITT, and may carry HSM-encrypted payload (HSM integration deferred to Phase 6). | D7.9 | 2 |
| REQ-FUNC-SAT-012 | Command verification MUST correlate spacecraft ACK + telemetry-state-match within configurable timeout; timeout triggers `verification_failed` and optional re-queue. | D7.9 | 2 |
| REQ-FUNC-SAT-013 | The platform MUST execute spacecraft simulation via `services/sat-simulation` for end-to-end testing without flight hardware, supporting all profile combinations from REQ-FUNC-SAT-001. | Conversation Q2 | 2 |

### 3.5 Ground station (REQ-FUNC-GS)

Source: `space_plan/docs/D6.2-product-backlog-groundstation.md`, `D7.1-deepdive-telemetry-pipeline.md`, `D7.2-deepdive-pass-execution-state-machine.md`, `D7.4`, `space_plan/docs/README.md`. Phase 2.

#### 3.5.1 Service boundaries

| ID | Requirement | Trace | Phase |
|---|---|---|---|
| REQ-FUNC-GS-BOUNDARY-001 | The ground-station plane MUST present seven services aligned to `space_plan/docs/README.md`: `SatelliteService`, `GroundStationService`, `PassService`, `TelemetryService`, `CommandService`, `AnomalyService`, `AlertService`, totalling the 52 RPCs enumerated in the README. | Conversation Q10, README.md | 2 |
| REQ-FUNC-GS-BOUNDARY-002 | Existing services (`gs-ingest`, `gs-mc`, `gs-rf`, `gs-scheduler`, `sat-command`, `sat-mission`, `sat-telemetry`) MUST be refactored or split to land within the seven plan boundaries. The defense-side ITAR services live in `chetana-defense`. | Conversation Q10, ITAR | 2 |

#### 3.5.2 TLE & pass prediction

| ID | Requirement | Trace | Phase |
|---|---|---|---|
| REQ-FUNC-GS-PASS-001 | `gs-pass-pred` MUST predict passes (AOS / max elevation / LOS within ±1 s of true), Doppler shift, sky-plot trajectory, and ground-track over a configurable horizon (1–30 days). | D7.4 | 2 |
| REQ-FUNC-GS-PASS-002 | Pass scheduling MUST detect antenna conflicts, support priority-based resolution, and enforce maintenance windows. | US-GS-020 … US-GS-027 | 2 |
| REQ-FUNC-GS-PASS-003 | Pass execution MUST follow the FSM in D7.2: `SCHEDULED → PREPARING → READY → ACQUIRING → TRACKING → CLOSING → REPORTING → COMPLETED`, plus `FAILED / CANCELLED / ABORTED` paths, with per-state timeouts, guards, and side-effect actions (allocate antenna, start TM capture, release resources). | D7.2 | 2 |

#### 3.5.3 Telemetry pipeline

| ID | Requirement | Trace | Phase |
|---|---|---|---|
| REQ-FUNC-GS-TM-001 | Frame reception MUST validate sync word, CRC, and APID, decommutate by ICD packet definition, calibrate parameters (polynomial / point-pair / lookup), apply limit checks (red/yellow/green; rate-of-change), and publish to Kafka. | D7.1 | 2 |
| REQ-FUNC-GS-TM-002 | Telemetry MUST be stored in a TimescaleDB hypertable (`telemetry_samples`) with continuous aggregates at 1-min and 1-h resolution. | D7.1, conversation Q6 | 2 |
| REQ-FUNC-GS-TM-003 | Retention MUST be: raw 7 d (hot hypertable), 1-min aggregates 90 d, 1-h aggregates 5 y; archived to S3 Glacier in compliance with REQ-FUNC-PLT-AUDIT-003. | D7.1 | 2 |
| REQ-FUNC-GS-TM-004 | Limit violations MUST emit alerts via `AlertService` and publish to the real-time gateway. | D7.1 | 2 |
| REQ-FUNC-GS-TM-005 | Real-time TM end-to-end latency (from frame reception to browser-rendered) MUST be ≤ 100 ms p95. | NFR-P (D6.2 platform NFR), conversation Q20 | 2 |

#### 3.5.4 Hardware abstraction

| ID | Requirement | Trace | Phase |
|---|---|---|---|
| REQ-FUNC-GS-HW-001 | The platform MUST implement a `HardwareDriver` Go interface in `services/packages/hardware/` with adapters for: USRP via UHD, RTL-SDR via librtlsdr, and a custom-protocol adapter — all three implementations MUST be production-grade, not stubs. | Conversation Q2-hw, no-stubs rule | 2 |
| REQ-FUNC-GS-HW-002 | The platform MUST implement an `AntennaController` interface with adapters for: Hamlib `rotctld` (TCP), GS-232 (RS-232 / TCP serial), and a custom-protocol adapter — all three implementations MUST be production-grade. | Conversation Q2-hw, no-stubs rule | 2 |
| REQ-FUNC-GS-HW-003 | The platform MUST implement a `GroundNetworkProvider` interface with adapters for: own-dish (direct hardware), AWS Ground Station (replacing Azure Orbital due to its 2026-09 EOL — pending confirmation in §9). | Conversation Q16, plan Q open | 2 |
| REQ-FUNC-GS-HW-004 | The RF chain MUST support all three bands (UHF / S / X) with software-configurable filters and LNAs per band. | Conversation Q3 | 2 |
| REQ-FUNC-GS-HW-005 | Demodulation MUST support BPSK, QPSK, OQPSK, 8PSK, and GMSK via `compute/crates/gs-bit-sync` and related crates. | Conversation Q3 | 2 |
| REQ-FUNC-GS-HW-006 | The platform MUST support all four CCSDS link layers (TM Transfer Frame per CCSDS 132.0-B-2, TC Transfer Frame per CCSDS 232.0-B-3, AOS per CCSDS 732.0-B-3, USLP per CCSDS 732.1-B-2). | Conversation Q3 | 2 |

### 3.6 Earth Observation (REQ-FUNC-EO)

Source: `space_plan/docs/D6.2-product-backlog-eo.md`, `D7.6-deepdive-stac-catalog-search.md`, `D7.7-deepdive-ml-inference-deployment.md`, `D7.8-deepdive-change-detection-pipeline.md`. Phase 3.

#### 3.6.1 STAC catalog

| ID | Requirement | Trace | Phase |
|---|---|---|---|
| REQ-FUNC-EO-CAT-001 | `eo-catalog` MUST expose STAC API 1.0.0 + OGC API Features endpoints: `/`, `/conformance`, `/collections`, `/collections/{id}`, `/collections/{id}/items`, `/collections/{id}/items/{itemId}`, `POST /search`, `/queryables`. | D7.6, US-EO-001 … US-EO-011 | 3 |
| REQ-FUNC-EO-CAT-002 | Catalog search MUST support BBox, temporal (ISO 8601), and CQL2 filters with ≤ 200 ms p95 latency. | D7.6 | 3 |
| REQ-FUNC-EO-CAT-003 | Spatial indexing MUST use PostGIS GIST + H3 hexagonal cells (resolutions 4, 6, 8). | D7.6 | 3 |
| REQ-FUNC-EO-CAT-004 | STAC items MUST validate against STAC 1.0.0 JSON Schema and supported extensions: EO (bands, cloud cover), SAR, projection, view, processing. | D7.6 | 3 |
| REQ-FUNC-EO-CAT-005 | Pagination MUST be cursor-based (opaque token in SearchRequest). | D7.6 | 3 |

#### 3.6.2 Processing pipeline

| ID | Requirement | Trace | Phase |
|---|---|---|---|
| REQ-FUNC-EO-PIPE-001 | `eo-pipeline` MUST accept processing jobs (orthorectification, pan-sharpen, spectral indices, mosaic, change-detection) and orchestrate execution via Kafka-driven workers. | US-EO-013 … US-EO-022, D7.8 | 3 |
| REQ-FUNC-EO-PIPE-002 | Orthorectification MUST use RPC model with DEM correction implemented in `compute/crates/eo-geometric` (production-grade, no stubs). | US-EO-014 | 3 |
| REQ-FUNC-EO-PIPE-003 | Pan-sharpening MUST support Brovey, Gram-Schmidt, IHS, PCA, and histogram-matched variants in `compute/crates/eo-pansharpen`. | US-EO-015 | 3 |
| REQ-FUNC-EO-PIPE-004 | Spectral indices MUST include NDVI, NDWI, EVI, SAVI in `compute/crates/eo-indices`. | US-EO-016 | 3 |
| REQ-FUNC-EO-PIPE-005 | Mosaic MUST support most-recent, least-cloud, and median composition in `compute/crates/eo-mosaic`. | US-EO-019 | 3 |
| REQ-FUNC-EO-PIPE-006 | Change-detection MUST orchestrate scene-pair selection, co-registration, radiometric normalisation, cloud masking, detection (CVA / image-diff / OBIA / DL), polygon extraction, and STAC publishing. F1 ≥ 0.90 on validation set; per-tile-pair latency ≤ 5 min p95; 24-h end-to-end SLA. | D7.8 | 3 |
| REQ-FUNC-EO-PIPE-007 | Throughput MUST sustain ≥ 100 Sentinel-2 tile pairs per hour. | NFR-P, D7.8 | 3 |

#### 3.6.3 ML serving

| ID | Requirement | Trace | Phase |
|---|---|---|---|
| REQ-FUNC-EO-ML-001 | The platform MUST serve ML inference via NVIDIA Triton Inference Server + ONNX Runtime + TensorRT, configured for dynamic batching (`max_queue_delay_microseconds=100`). | D7.7, conversation Q8 | 3 |
| REQ-FUNC-EO-ML-002 | Inference latency MUST be ≤ 100 ms p95 for a single 256×256 tile; throughput ≥ 10 000 tiles/min batch; GPU utilisation ≥ 80 %. | D7.7, NFR-P | 3 |
| REQ-FUNC-EO-ML-003 | The platform MUST maintain an MLflow-style model registry with versioning, status lifecycle (draft → staging → canary → production → archived), traffic-weight-based A/B routing, canary deployment, shadow deployment, and rollback. | D7.7 | 3 |
| REQ-FUNC-EO-ML-004 | The platform MUST auto-convert PyTorch and TensorFlow checkpoints to ONNX with tensor I/O verification on registry intake. | D7.7 | 3 |
| REQ-FUNC-EO-ML-005 | Triton MUST run on a Kubernetes HPA tied to GPU utilisation > 80 % and queue depth thresholds. | D7.7, conversation Q20 | 3 |
| REQ-FUNC-EO-ML-006 | Model artifacts MUST carry an `export_classification` attribute; ITAR-classified models MUST be servable only to US-person principals. | ITAR | 3 |

### 3.7 GeoInt (REQ-FUNC-GI)

Source: `space_plan/docs/D6.2-product-backlog-geoint.md`. Phase 4.

| ID | Requirement | Trace | Phase |
|---|---|---|---|
| REQ-FUNC-GI-WS-001 | `gi-workspace` MUST manage workspaces with members (viewer/editor/admin/owner), saved views, and activity audit. | US-GEO-001 … US-GEO-010 | 4 |
| REQ-FUNC-GI-WS-002 | Map layers MUST be configurable per workspace and support WMS, WMTS, COG, MVT sources. | US-GEO-003 | 4 |
| REQ-FUNC-GI-WS-003 | Annotations MUST be stored as GeoJSON with undo/redo, version history, and concurrent-editing conflict resolution. | US-GEO-005 | 4 |
| REQ-FUNC-GI-WS-004 | Measurements MUST compute geodesic distance and area using `@turf/*` (browser) or PostGIS (server). | US-GEO-006 | 4 |
| REQ-FUNC-GI-AOI-001 | `gi-aoi` MUST manage AOI polygons in PostGIS with monitoring rules, AOI alerts, and timeline of imagery covering each AOI. | US-GEO-011 … US-GEO-020 | 4 |
| REQ-FUNC-GI-AN-001 | Spatial analysis MUST include object counting, object tracking, area measurement, terrain analysis (DEM-based: elevation profile, slope, aspect, viewshed), buffer/proximity, heatmap, spatial query, and sandboxed Python custom scripts. | US-GEO-021 … US-GEO-028 | 4 |
| REQ-FUNC-GI-RPT-001 | `gi-report` MUST provide a WYSIWYG report editor, template library, embedded map snapshots, embedded imagery, export to PDF/DOCX/PPTX/HTML, share links with access control, version history, and scheduled report generation. | US-GEO-029 … US-GEO-036 | 4 |
| REQ-FUNC-GI-DEM-001 | The platform MUST host a DEM tile cache (S3-backed) with WMTS-style serving and on-demand viewshed/slope/aspect computation. | US-GEO-024 | 4 |

### 3.8 Imagery-as-a-Service (REQ-FUNC-IAAS)

Source: `space_plan/docs/D6.2-product-backlog-eo.md` US-EO-031 … US-EO-038. Phase 5.

| ID | Requirement | Trace | Phase |
|---|---|---|---|
| REQ-FUNC-IAAS-001 | The platform MUST expose a public API gateway with API-key authentication, per-key rate limiting, and usage metering. | US-EO-035, US-EO-036 | 5 |
| REQ-FUNC-IAAS-002 | The public surface MUST expose only public-classified collections; ITAR / CUI / restricted data MUST NOT be reachable from the public surface. | ITAR | 5 |
| REQ-FUNC-IAAS-003 | Customers MUST be able to subscribe to AOIs and receive presigned-URL deliveries for matching scenes via email. | US-EO-033, US-EO-034 | 5 |
| REQ-FUNC-IAAS-004 | The platform MUST register DOIs and provide formatted citations for published collections. | US-EO-038 | 5 |
| REQ-FUNC-IAAS-005 | The customer portal MUST live in a separate route group within the same SvelteKit app (`web/apps/shell/src/routes/(public)`). | Conversation Q14 | 5 |

---

## 4. Non-functional requirements

### 4.1 Performance (REQ-NFR-PERF)

| ID | Requirement | Trace | Verification |
|---|---|---|---|
| REQ-NFR-PERF-001 | TM end-to-end latency ≤ 100 ms p95 (frame ingestion → browser render). | D7.1, conversation Q20 | k6 / browser performance API |
| REQ-NFR-PERF-002 | Sustained EO processing throughput ≥ 100 Sentinel-2 tile pairs/h. | D7.8 | Load test in Phase 3 |
| REQ-NFR-PERF-003 | ML inference ≤ 100 ms p95 per 256×256 tile; ≥ 10 000 tiles/min batch. | D7.7 | Load test in Phase 3 |
| REQ-NFR-PERF-004 | STAC search ≤ 200 ms p95. | D7.6 | Load test in Phase 3 |
| REQ-NFR-PERF-005 | IAM login ≤ 100 ms p95 at 1 000 req/s. | D7.10 | k6 in Phase 1 |
| REQ-NFR-PERF-006 | WebSocket gateway ≤ 500 ms p95 push latency at 10 000 concurrent connections. | D7.3 | k6 in Phase 1 |

### 4.2 Reliability (REQ-NFR-REL)

| ID | Requirement | Trace | Verification |
|---|---|---|---|
| REQ-NFR-REL-001 | Platform availability ≥ 99.9 % monthly. | Plan-level NFR | DR drill, monitoring |
| REQ-NFR-REL-002 | RPO ≤ 5 min, RTO ≤ 1 h for the primary database tier. | ISO 27001 A.5.30 | DR drill in Phase 6 |
| REQ-NFR-REL-003 | Every service MUST have a Helm `HorizontalPodAutoscaler` and `PodDisruptionBudget` with `minAvailable ≥ 1`. | Conversation Q20 | Helm template + e2e test |
| REQ-NFR-REL-004 | NetworkPolicy default-deny MUST be enforced at namespace level; per-service ingress rules explicitly allow required traffic. | ISO 27001 A.8.20 | Policy audit |

### 4.3 Security (REQ-NFR-SEC)

| ID | Requirement | Trace | Verification |
|---|---|---|---|
| REQ-NFR-SEC-001 | All cryptographic operations MUST use FIPS 140-2/3 validated modules (Go: GOEXPERIMENT=boringcrypto or BoringSSL-Go; Rust: rustls + FIPS provider; Postgres: FIPS-validated build; Kafka: FIPS-mode TLS; Web: Web Crypto API in FIPS browser posture). | FedRAMP-Mod, ITAR | Crypto audit |
| REQ-NFR-SEC-002 | TLS 1.3 MUST be enforced for all external traffic; mTLS MUST be enforced for inter-service traffic on the cluster. | ISO 27001 A.8.24 | Penetration test |
| REQ-NFR-SEC-003 | All secrets MUST be stored in AWS Secrets Manager (FIPS endpoint) and rotated per policy (≤ 90 d for service credentials). | ISO 27001 A.5.16, A.5.17 | Inspection |
| REQ-NFR-SEC-004 | Container images MUST be signed with cosign and verified at admission time (Kyverno or equivalent). | Supply-chain | CI evidence |
| REQ-NFR-SEC-005 | SBOM (CycloneDX) MUST be generated for every release and stored alongside the image attestation. | NIST SSDF, FedRAMP | CI evidence |
| REQ-NFR-SEC-006 | SAST (gosec, semgrep), DAST (zap baseline), and SCA (cargo-audit, npm audit, pip-audit, trivy) MUST run on every PR; merging is blocked on critical findings. | ISO 27001 A.8.28, FedRAMP | CI evidence |

### 4.4 Observability (REQ-NFR-OBS)

| ID | Requirement | Trace | Verification |
|---|---|---|---|
| REQ-NFR-OBS-001 | Every service MUST emit OpenTelemetry traces, metrics, and structured logs with trace correlation. | ISO 27001 A.8.15 | Inspection |
| REQ-NFR-OBS-002 | Trace spans MUST propagate across ConnectRPC, Kafka, HTTP, and WebSocket boundaries. | ISO 27001 A.8.15 | e2e test |
| REQ-NFR-OBS-003 | Metrics MUST be scraped by Prometheus; dashboards MUST be defined in Grafana with provisioned-from-code JSON. | Plan-level | Inspection |
| REQ-NFR-OBS-004 | Audit log integrity MUST be verifiable independently of any service via the chain verifier (REQ-FUNC-PLT-AUDIT-002). | ITAR | Compliance test |

### 4.5 Scalability (REQ-NFR-SCALE)

| ID | Requirement | Trace | Verification |
|---|---|---|---|
| REQ-NFR-SCALE-001 | Every stateful service MUST be horizontally scalable; no service-instance affinity except via documented sticky-session paths (currently: realtime-gw via Redis Pub/Sub fan-out). | Plan-level | Architecture review |
| REQ-NFR-SCALE-002 | Kafka topics MUST be partitioned to support ≥ 10× current throughput without re-keying. | D7.1 | Inspection |
| REQ-NFR-SCALE-003 | All services MUST be region-aware (`region` injected via env var; data plane reads/writes routed regionally), even though v1 deploys to one region. | Conversation Q19 (multi-region readiness) | Inspection |

---

## 5. Compliance requirements

Compliance certifications are staggered (conversation: confirmed). Architecture must support all five from day one; certifications stagger.

### 5.1 ISO 27001 (v1 — Phase 6)

| ID | Requirement | Trace |
|---|---|---|
| REQ-COMP-ISO-001 | The platform MUST maintain an ISMS with documented scope, statement of applicability, risk register, and 93 Annex A controls mapped to implementation evidence. | ISO 27001:2022 |
| REQ-COMP-ISO-002 | The platform MUST execute internal audit, management review, and remediation cycles on at least an annual cadence. | ISO 27001:2022 |
| REQ-COMP-ISO-003 | The platform MUST maintain SOPs for incident response, business continuity, supplier management, change management. | ISO 27001:2022 |

### 5.2 GDPR (v1 — Phase 6)

| ID | Requirement | Trace |
|---|---|---|
| REQ-COMP-GDPR-001 | The platform MUST implement Subject Access Request (Article 15), Erasure (Article 17), Portability (Article 20), and Rectification (Article 16) endpoints. | GDPR |
| REQ-COMP-GDPR-002 | The platform MUST maintain a Records of Processing Activities (ROPA) document at `compliance/ropa.md`. | GDPR Article 30 |
| REQ-COMP-GDPR-003 | The platform MUST execute and version-control DPIAs for high-risk processing (e.g. external-customer surface, ML inference on personal data). | GDPR Article 35 |
| REQ-COMP-GDPR-004 | Breach notification MUST trigger an internal pager within 1 h of detection; supervisory authority notification within 72 h. | GDPR Article 33 |
| REQ-COMP-GDPR-005 | An EU representative and DPO MUST be appointed and identified in privacy notices. | GDPR Articles 27, 37 |

### 5.3 SOC 2 Type II (v1.x — post-launch)

| ID | Requirement | Trace |
|---|---|---|
| REQ-COMP-SOC2-001 | The platform MUST satisfy the AICPA Trust Services Criteria for Security, Availability, and Confidentiality with ≥ 6 months of evidence by the time of the audit. | SOC 2 |

### 5.4 India CERT-In (v1.2 — post-launch + India region)

| ID | Requirement | Trace |
|---|---|---|
| REQ-COMP-CERTIN-001 | Cyber security incidents MUST be reportable to CERT-In within 6 h of detection. | CERT-In Directions 2022 |
| REQ-COMP-CERTIN-002 | Logs MUST be retained for ≥ 180 days within Indian jurisdiction (requires India region). | CERT-In Directions 2022 |
| REQ-COMP-CERTIN-003 | VPN session metadata MUST be retained per CERT-In requirements. | CERT-In Directions 2022 |

### 5.5 ITAR (v2.0 — post-launch + 2-repo posture from day 1)

| ID | Requirement | Trace |
|---|---|---|
| REQ-COMP-ITAR-001 | Defense technical data MUST reside only in the `chetana-defense` repo, accessible only by US-person principals (verified via I-9 + onboarding). | 22 CFR 120.15, 120.10 |
| REQ-COMP-ITAR-002 | ITAR-classified runtime services MUST run in a dedicated Kubernetes namespace (`chetana-itar`) with US-person operator-only access and US-region-only nodes. | 22 CFR 120.10 |
| REQ-COMP-ITAR-003 | Container images carrying ITAR code MUST be labelled `org.chetana.classification=itar` and signed; deployment MUST verify the label. | ITAR best practice |
| REQ-COMP-ITAR-004 | ITAR audit events MUST be retained ≥ 5 years online. | ITAR record-keeping |
| REQ-COMP-ITAR-005 | The platform MUST register with DDTC and maintain the technology control plan. | ITAR |

### 5.6 FedRAMP-Moderate (v2.1 — post-launch)

| ID | Requirement | Trace |
|---|---|---|
| REQ-COMP-FEDRAMP-001 | The platform MUST be hosted in AWS GovCloud (US) for FedRAMP-Mod. | Conversation Q-cloud |
| REQ-COMP-FEDRAMP-002 | NIST SP 800-53 Rev 5 Moderate baseline (~325 controls) MUST be implemented and documented. | FedRAMP |
| REQ-COMP-FEDRAMP-003 | A 3PAO assessment MUST be completed; ATO MUST be issued before federal data is processed. | FedRAMP |
| REQ-COMP-FEDRAMP-004 | Continuous monitoring (POA&M, monthly vulnerability scans, annual assessment) MUST be operational. | FedRAMP |

---

## 6. Constraints

| ID | Constraint | Source |
|---|---|---|
| REQ-CONST-001 | `space_plan/docs/*` are immutable contracts; functional requirements may not be relaxed without an explicit, documented change order. | Conversation Q21 |
| REQ-CONST-002 | Tech-stack overrides authorised vs the plan's tech recommendations: Kafka (not NATS), Rust (not Julia), Cesium-only (not Cesium + MapLibre + Leaflet). | Conversation Q5, Q7, Q9 |
| REQ-CONST-003 | First-region cluster: AWS GovCloud (US-East). | Conversation Q-cloud |
| REQ-CONST-004 | Repository topology: two repos — `chetana-platform` (this repo, post-extraction) and `chetana-defense` (private, US-persons only). | Conversation Q-multirepo |
| REQ-CONST-005 | Single SvelteKit app (`web/apps/shell`) with module-separated route registry; no MFE in v1. | Conversation Q14 |
| REQ-CONST-006 | Tauri desktop wrapper deferred from v1. | Conversation Q15 |
| REQ-CONST-007 | Single tenant in v1; multi-tenant deferred. | Conversation Q3 |
| REQ-CONST-008 | NPM scope: `@p9e.in/chetana`. Go module path: `p9e.in/chetana/...`. Cargo crate authors: `chetana`. Locale strings rebrand from `samavāya` to `chetana`. | Conversation, brand |
| REQ-CONST-009 | NFR gates (HPA / PDB / load tests passing the targets in §4.1) MUST land in Phase 1 for the foundation services and Phase 2/3 for domain services; a service MUST NOT ship to production without its gate passing. | Conversation Q20 |
| REQ-CONST-010 | Implementation MUST NOT contain stubs, placeholders, TODO-based fallbacks, or non-production-grade code. Where a feature is genuinely deferred, the absence is recorded in §8 (Out of v1) and the API surface returns an explicit "feature not enabled" error rather than a partial implementation. | Conversation directive |
| REQ-CONST-011 | No code duplication: shared logic MUST live in `services/packages/` (Go) or `compute/crates/` (Rust) or `web/packages/` (TS); duplicated logic MUST be detected by lint and refactored before merge. | Conversation directive |
| REQ-CONST-012 | Every bug fix MUST include a regression test that fails before the fix and passes after; the test MUST cover sibling code paths plausibly affected by the same root cause. | Conversation directive |
| REQ-CONST-013 | No assumptions or hallucinations: every requirement, design choice, file path, and identifier in `plan/*` traces to either `space_plan/`, the existing codebase, or a recorded conversation decision. Ambiguities MUST be flagged in §9 (Open questions), not papered over. | Conversation directive |

---

## 7. Tech-stack decisions (locked for v1)

| Layer | Choice | Trace |
|---|---|---|
| Backend RPC | ConnectRPC + h2c (cluster-internal); ConnectRPC + TLS 1.3 (external) | Existing |
| Backend language | Go 1.26+ for application services; Rust for compute and flight; Python for ML | Existing + conversation |
| Event bus | Apache Kafka (Amazon MSK in GovCloud) | Conversation Q5 |
| OLTP DB | PostgreSQL 16 (RDS or self-managed on EKS) | Existing |
| Time-series DB | TimescaleDB extension on PostgreSQL | Conversation Q6 |
| Object storage | Amazon S3 (GovCloud) | Conversation Q-cloud |
| Cache / pub-sub | Redis (ElastiCache or self-managed) | Existing + D7.3 |
| Search index | PostgreSQL FTS + GIN (no Elasticsearch in v1) | Conservative scope |
| ML serving | NVIDIA Triton + ONNX Runtime + TensorRT | Conversation Q8 |
| ML registry | MLflow-style schema implemented in `services/eo-analytics` | Conversation Q8, D7.7 |
| Frontend | SvelteKit 2 + Svelte 5 + Vite 7 + UnoCSS + Tailwind 4 | Existing |
| 3D visualisation | Cesium (`@cesium/engine`) | Conversation Q9 |
| 2D map | Cesium Columbus view (no MapLibre/Leaflet) | Conversation Q9 |
| Charts | ECharts (existing wrapper) + D3 modules where ECharts is insufficient (sky plot, B-plane) | Existing + needs |
| WASM | wasm-pack (Rust → wasm32-unknown-unknown) | Existing |
| IaC | Helm 3 charts + Terraform for AWS infra | Existing + needs |
| CI | GitHub Actions (platform repo) + self-hosted runners in GovCloud account (defense repo) | Conversation |
| Container registry | Amazon ECR (separate repos for platform vs ITAR-namespace) | Conversation Q-multirepo |
| Identity | Own IAM service (no Keycloak/Auth0) | Conversation Q12 |

---

## 8. Out of v1 scope

The following items are explicitly **not** delivered in v1 and are tracked as v1.x / v2.0 backlog. Their absence is observable as "feature not enabled" errors at the API surface, never as silent stubs.

| Item | Phase target | Rationale |
|---|---|---|
| Tauri desktop wrapper | v2.0 | Conversation Q15 |
| Multi-tenant RLS on Postgres | v1.x | Conversation Q3 |
| MapLibre and Leaflet | v2.0 | Conversation Q9 |
| Julia scientific compute | Permanently descoped | Conversation Q7 |
| FedRAMP 3PAO assessment | v2.1 | §5.6 staggering |
| ITAR DDTC registration & TCP | v2.0 | §5.5 staggering |
| India region cluster + CERT-In compliance | v1.2 | §5.4 staggering |
| EU region cluster | v1.x | After GDPR audit |
| AWS Ground Station integration | v1.x | Phase 2 implements own-dish; AWS GS adapter behind interface in v1.x |
| KSAT / SSC ground network adapters | v2.0 | After AWS GS ships |
| Custom ML model training (US-EO-030) | v2.0 | D6.2 EO backlog flags this v2 |
| WebAssembly browser-side image preview | v1.x | After core v1 ships |
| HSM / KGV-72 command encryption | v1.x or v2.0 | Tied to ITAR phase; pluggable interface lands in Phase 2 |

---

## 9. Open questions (must be resolved before relevant phase)

These items are unresolved and block the phases listed.

| ID | Question | Blocks | Owner | Notes |
|---|---|---|---|---|
| OQ-001 | Confirm: AWS Ground Station replaces Azure Orbital as the second `GroundNetworkProvider` (Azure Orbital EOL 2026-09). | Phase 2 (provider adapter) | Customer | If declined, propose KSAT or SatNOGS as alternative. |
| OQ-002 | Provision empty `chetana-defense` GitHub repo + US-persons team. | Phase 0 PR-A2 | Customer | Required before repo split lands. |
| OQ-003 | GitHub Enterprise vs Cloud (affects SAML SSO + audit log streaming + IP allowlists for ITAR). | Phase 0 PR-A2 | Customer | If Cloud only: hardware-MFA-mandatory paid Enterprise org for defense repo. |
| OQ-004 | Internal Go module proxy / Cargo registry / buf BSR org existence. | Phase 0 PR-A2 | Customer | If absent, bootstrap in PR-A2 (~3 days). |
| OQ-005 | Sanity-check `compliance/itar-paths.txt` (sat-telemetry classification model: all-defense vs split). | Phase 0 PR-A2 | Customer | Currently conservative all-defense. |
| OQ-006 | Spacecraft details (bus type, exact RF parameters, link budget, safety modes) for the first vehicle. | Phase 2 spacecraft profile | Mission | Generic profile system lands first; concrete profile loaded later. |
| OQ-007 | First-contact / launch date. | Phase 2 hardware procurement timeline | Mission | TBD per conversation. |
| OQ-008 | Hosting boundaries — single GovCloud cluster for v1 confirmed; cross-region active/standby topology for v1.x is open. | v1.x planning | Architecture | Architecture is region-aware; cluster topology decision deferred. |
| OQ-009 | Compliance staffing — does the team have a DPO and ITAR Empowered Official, or do we contract them? | Phase 6 (DPIA), v2.0 (ITAR registration) | Customer | Staffing affects v1.x v2.0 calendar. |
| OQ-010 | EU representative under GDPR Article 27. | Phase 6 (GDPR audit) | Customer | Must be appointed before GDPR readiness audit. |

---

## 10. Verification matrix (summary)

Every requirement above is verified by at least one of:

- **Unit / integration test** — code-level tests in the owning service.
- **End-to-end test** — full-stack tests in `services/<svc>/test/e2e/` and `web/apps/shell/tests/e2e/`.
- **Load benchmark** — k6 / criterion / pytest-benchmark scenario in `bench/`.
- **Inspection** — manual or automated review of code, config, or evidence.
- **Audit** — third-party or internal compliance audit.
- **DR drill** — scheduled exercise.

The verification method is recorded per-requirement above. `plan/todo.md` enumerates the test cases as concrete tasks.
