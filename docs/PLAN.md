# Space Technology Platform — Architecture Plan

**Document ID:** P9E-SPACE-PLAN-2026-001
**Version:** 1.0
**Status:** Active
**Source spec:** [`requirements`](../requirements) (P9E-SPACE-RS-2025-001)

This document is the implementation plan for the four-module Space Technology
Platform described in the requirements specification. It defines the service
inventory, compute-package inventory, folder structure, and ownership map
needed to deliver every functional and non-functional requirement.

The platform is delivered **without** integration with the P9e Chetana platform
(IAM/Policy/Analytics/Notify). Equivalent capabilities are owned by services in
this repository.

## 1. Top-level monorepo layout

```
space/
├── go.work                          Go workspace pinning all Go modules
├── Makefile                         tools, proto, sqlc, build, test, vet
├── README.md
├── requirements                     the source specification
│
├── proto/                           ALL .proto files (single source of truth)
│   └── p9e/space/
│       ├── common/v1/               pagination, errors, geo, time
│       ├── iam/v1/
│       ├── audit/v1/
│       ├── notify/v1/
│       ├── earthobs/v1/             eo_catalog, eo_pipeline, eo_analytics
│       ├── satsubsys/v1/            mission, fsw, telemetry, command, sim
│       ├── groundstation/v1/        mc, rf, scheduler, ingest
│       └── geoint/v1/               fusion, analytics, tiles, reports, predict
│
├── api/                             buf-generated Go (mirrors proto/)
│
├── pkg/                             shared Go libraries (one Go module)
│   ├── config/                      viper-based, env+file
│   ├── observability/               OTEL traces/metrics/logs, Prometheus
│   ├── authz/                       JWT verification, RBAC enforcement
│   ├── httpserver/                  ConnectRPC server, h2c, graceful shutdown
│   ├── middleware/                  auth, recovery, tenant, ratelimit, audit
│   ├── db/                          pgxpool, migration runner, txmgr
│   ├── pagination/                  cursor + offset helpers
│   ├── validation/                  protovalidate-go integration
│   ├── errs/                        domain → connect.Error mapping
│   ├── ids/                         ULID/UUIDv7 generators
│   ├── timeutil/                    CCSDS time, GPS time, UTC helpers
│   ├── kafka/                       producer/consumer
│   ├── objectstore/                 S3/MinIO client (presigned URLs)
│   ├── geo/                         PostGIS WKB / bbox helpers
│   ├── stac/                        STAC item/collection codec
│   ├── tle/                         TLE parsing/wire format
│   └── closer/                      graceful resource teardown
│
├── services/                        ALL Go ConnectRPC services (one module each)
│   ├── iam/
│   ├── audit/
│   ├── notify/
│   │
│   ├── eo-catalog/
│   ├── eo-pipeline/
│   ├── eo-analytics/
│   │
│   ├── sat-mission/
│   ├── sat-fsw/
│   ├── sat-telemetry/
│   ├── sat-command/
│   ├── sat-simulation/
│   │
│   ├── gs-mc/
│   ├── gs-rf/
│   ├── gs-scheduler/
│   ├── gs-ingest/
│   │
│   ├── gi-fusion/
│   ├── gi-analytics/
│   ├── gi-tiles/
│   ├── gi-reports/
│   └── gi-predict/
│
├── compute/                         Rust workspace (compute engines)
│   ├── Cargo.toml                   workspace
│   ├── crates/                      see § 4 for full list
│   └── bins/                        thin worker daemons (eo-worker, gs-worker, gi-worker)
│
├── ml/                              Python ML packages (Poetry workspace)
│   ├── packages/
│   │   ├── eo_ml/                   YOLO/Faster-RCNN/U-Net/DeepLabV3+/ViT
│   │   ├── gi_ml/                   risk maps, agricultural forecasting
│   │   └── ml_serving/              ONNX Runtime / Triton / TorchServe
│   └── workers/
│       ├── eo-ml-worker/            Kafka consumer for EO inference
│       └── gi-ml-worker/            Kafka consumer for GI inference
│
├── deploy/
│   ├── docker/                      Dockerfiles per service (distroless)
│   ├── compose/                     docker-compose for local dev
│   └── helm/                        Helm charts per service
│
└── docs/
    ├── PLAN.md                      this document
    ├── TODO.md                      build-order checklist
    └── adr/                         Architecture Decision Records
```

## 2. Per-service folder structure (uniform across ALL Go services)

Every Go service uses this exact layout:

```
services/<service>/
├── go.mod
├── sqlc.yaml
├── atlas.hcl                        Atlas migration config
├── Dockerfile
│
├── cmd/
│   └── <service>/
│       └── main.go                  bootstrap, DI, OTEL, signals
│
├── db/
│   ├── schema/                      *.sql DDL (Atlas-managed)
│   ├── queries/                     *.sql (sqlc input)
│   └── generated/                   sqlc output (committed)
│
└── internal/
    ├── config/                      service-specific config struct
    ├── handlers/                    ConnectRPC handlers (1 file per RPC group)
    ├── service/                     business logic
    ├── repository/                  repository interfaces + sqlc-backed impls
    ├── models/                      domain models (pure Go)
    └── mappers/                     proto ⇄ domain ⇄ sqlc converters
```

## 3. Service inventory (20 Go services)

### 3.1 Cross-cutting (3 services)
| Service  | Owns | Storage |
|----------|------|---------|
| `iam`    | OIDC issuer, tenants, users, roles, RBAC, MFA | PostgreSQL |
| `audit`  | Hash-chained tamper-evident audit log | PostgreSQL + WORM bucket |
| `notify` | Email / SMS / Webhook fan-out, templates | PostgreSQL + Kafka |

### 3.2 Earth Observation (3 services)
| Service        | Owns | Storage |
|----------------|------|---------|
| `eo-catalog`   | STAC collections/items/assets, ingestion lineage, QA | PostgreSQL + PostGIS, S3 |
| `eo-pipeline`  | Processing job orchestration | PostgreSQL, Kafka, S3 |
| `eo-analytics` | ML model registry, inference jobs, derived products | PostgreSQL + PostGIS, S3 |

### 3.3 Small Satellite Subsystems (5 services)
| Service           | Owns | Storage |
|-------------------|------|---------|
| `sat-mission`     | Satellite registry, mission config, modes, orbital state | PostgreSQL |
| `sat-fsw`         | Flight-software registry, deploy packages, manifests | PostgreSQL + S3 |
| `sat-telemetry`   | Housekeeping ingest, parameter catalog, retrieval | TimescaleDB |
| `sat-command`     | Command catalog, sequences, time-tagged uplink queue | PostgreSQL, Kafka |
| `sat-simulation`  | SITL/HITL run management, Monte-Carlo orchestration | PostgreSQL + S3 |

### 3.4 Ground Station (4 services)
| Service        | Owns | Storage |
|----------------|------|---------|
| `gs-mc`        | Live telemetry, dashboards, anomaly rules, procedures | TimescaleDB + PostgreSQL |
| `gs-rf`        | Antenna+RF equipment registry, signal-quality time-series | PostgreSQL + TimescaleDB |
| `gs-scheduler` | Pass prediction, multi-sat conflict resolution | PostgreSQL |
| `gs-ingest`    | Data reception orchestration, recording catalog | PostgreSQL + S3 |

### 3.5 Geospatial Intelligence (5 services)
| Service        | Owns | Storage |
|----------------|------|---------|
| `gi-fusion`    | Layer registry, fusion job orchestration, temporal stacking | PostgreSQL + PostGIS |
| `gi-analytics` | ABI patterns, infrastructure / environmental / maritime monitors | PostgreSQL + PostGIS |
| `gi-tiles`     | WMS / WMTS / MVT tile server, 3D-terrain manifests | PostgreSQL + S3 |
| `gi-reports`   | Templates, generated reports, annotations, export jobs | PostgreSQL + S3 |
| `gi-predict`   | Predictive model registry, inference jobs, scenario runs | PostgreSQL + PostGIS |

## 4. Rust compute crates (Cargo workspace at `compute/`)

### 4.1 Earth Observation compute
| Crate            | Capability |
|------------------|------------|
| `eo-radiometric` | TOA / BOA reflectance |
| `eo-geometric`   | Orthorectification (DEM-based) |
| `eo-atmos-corr`  | Sen2Cor / 6S / MODTRAN process wrapper |
| `eo-pansharpen`  | Brovey, IHS, PCA, Gram-Schmidt |
| `eo-mosaic`      | Seamline generation, blending |
| `eo-sar`         | Speckle filter, terrain correction, polarimetric decomposition |
| `eo-indices`     | NDVI, EVI, SAVI, NDWI |

### 4.2 Satellite flight software
| Crate            | Capability |
|------------------|------------|
| `adcs-ekf`       | Extended Kalman Filter (separate crate) |
| `adcs-ukf`       | Unscented Kalman Filter (separate crate) |
| `adcs-control`   | PID, LQR, sliding-mode, MPC controllers |
| `adcs-actuators` | Reaction wheel, magnetorquer, thruster drivers |
| `orbit-prop`     | SGP4 / SDP4, RK4, RK78 integrators |
| `cdh-ccsds`      | Space Packet Protocol, TC/TM frames, COP-1 |
| `cdh-fs`         | Log-structured on-board file system |
| `cdh-scheduler`  | Time-tagged command execution |
| `cdh-edac`       | Hamming + Reed-Solomon memory scrubbing |
| `eps-power`      | Power budget, MPPT, eclipse predictor |
| `eps-battery`    | State-of-charge / state-of-health estimators |
| `sim-physics`    | 6-DOF dynamics, gravity, magnetic, drag, SRP |
| `sim-harness`    | SITL/HITL harness, Monte-Carlo runner |

### 4.3 Ground-station DSP / drivers
| Crate          | Capability |
|----------------|------------|
| `gs-tle`       | TLE parsing → ECI/ECEF (uses `orbit-prop`) |
| `gs-pass-pred` | Pass prediction with horizon mask |
| `gs-doppler`   | Uplink/downlink frequency compensation |
| `gs-bit-sync`  | Bit & frame synchronization |
| `gs-fec`       | Reed-Solomon, LDPC, Turbo, convolutional codecs |
| `gs-antenna`   | Program / auto / step-track driver core |
| `gs-rf-driver` | SNMP / Modbus drivers for LNA / SSPA / synthesizer |

### 4.4 GEOINT compute
| Crate              | Capability |
|--------------------|------------|
| `gi-fusion-eng`    | Multi-layer raster fusion |
| `gi-abi`           | Pattern-of-life / activity-based intelligence |
| `gi-tile-render`   | MVT vector tile generator |
| `gi-report-render` | Report layout + PDF emission |
| `gi-export`        | GeoJSON / Shapefile / KML / GeoPackage writers |

### 4.5 Worker daemons (`compute/bins/`)
| Binary       | Purpose |
|--------------|---------|
| `eo-worker`  | Consumes EO job topics, runs EO crates, publishes results |
| `gs-worker`  | Consumes GS job topics, runs GS crates, publishes results |
| `gi-worker`  | Consumes GI job topics, runs GI crates, publishes results |

Each worker is a thin ConnectRPC + Kafka consumer linking the relevant crates.

### 4.6 Per-crate layout

```
compute/crates/<crate>/
├── Cargo.toml
├── src/
│   ├── lib.rs
│   └── ...
├── benches/                         Criterion benchmarks
├── tests/                           Integration tests
└── examples/                        Runnable usage examples
```

## 5. Python ML packages

| Package      | Purpose |
|--------------|---------|
| `eo_ml`      | YOLO, Faster R-CNN, U-Net, DeepLabV3+, Vision Transformers |
| `gi_ml`      | Predictive risk maps, agricultural forecasting, urban-growth models |
| `ml_serving` | Common inference server (ONNX Runtime / Triton / TorchServe) |

| Worker          | Purpose |
|-----------------|---------|
| `eo-ml-worker`  | Kafka consumer for EO inference jobs |
| `gi-ml-worker`  | Kafka consumer for GI inference jobs |

Per-package layout:
```
ml/packages/<pkg>/
├── pyproject.toml
├── src/<pkg>/
│   ├── __init__.py
│   ├── models/
│   ├── preprocess/
│   ├── postprocess/
│   └── pipeline.py
└── tests/
```

## 6. Requirements → ownership map

| Requirement IDs | Owning service / crate |
|-----------------|-------------------------|
| EO-FR-001 .. 005 | `eo-catalog` |
| EO-FR-010 .. 015 | `eo-pipeline` → `eo-radiometric` / `eo-geometric` / `eo-atmos-corr` / `eo-pansharpen` / `eo-mosaic` / `eo-sar` |
| EO-FR-020 .. 026 | `eo-analytics` → `eo_ml` (Python) + `eo-indices` |
| SAT-FR-001 .. 006 | `adcs-ekf` / `adcs-ukf` / `adcs-control` / `adcs-actuators` / `orbit-prop` (deployed via `sat-fsw`) |
| SAT-FR-010 .. 015 | `cdh-ccsds` / `cdh-fs` / `cdh-scheduler` / `cdh-edac` |
| SAT-FR-020 .. 024 | `eps-power` / `eps-battery` |
| SAT-FR-030 .. 033 | `sat-simulation` → `sim-physics` / `sim-harness` |
| GS-FR-001 .. 006 | `gs-mc` |
| GS-FR-010 .. 014 | `gs-rf` → `gs-antenna` / `gs-rf-driver` / `gs-doppler` |
| GS-FR-020 .. 024 | `gs-scheduler` → `gs-tle` / `gs-pass-pred` |
| GS-FR-030 .. 033 | `gs-ingest` → `gs-bit-sync` / `gs-fec` |
| GI-FR-001 .. 005 | `gi-fusion` → `gi-fusion-eng` |
| GI-FR-010 .. 015 | `gi-analytics` → `gi-abi` + `gi_ml` (Python) |
| GI-FR-020 .. 024 | `gi-tiles` / `gi-reports` → `gi-tile-render` / `gi-report-render` / `gi-export` |
| GI-FR-030 .. 033 | `gi-predict` → `gi_ml` (Python) |
| NFR-* | `pkg/*` shared infra + per-service config |

## 7. Cross-cutting decisions

1. **Auth without P9e:** `iam` issues short-lived RS256 JWTs from PostgreSQL-backed identities. Every service validates via JWKS exposed by `iam`.
2. **Multi-tenant:** every domain table carries `tenant_id`; PostgreSQL row-level security enforced.
3. **Time-series:** TimescaleDB for `sat-telemetry`, `gs-mc`, `gs-rf` signal-quality data.
4. **Object store:** S3 (MinIO in dev) for imagery, recordings, FSW packages, reports.
5. **Streaming:** Kafka topics for telemetry ingest, command queues, ML/compute job dispatch.
6. **Compute dispatch:** Go service writes a job row + emits Kafka event; Rust/Python worker consumes, executes, writes result + emits completion event. **No CGO in Go services.**
7. **Observability:** OpenTelemetry traces/metrics/logs everywhere; Prometheus scrape; Jaeger; Grafana.
8. **Migrations:** Atlas declarative; one schema per service.
9. **Validation:** `protovalidate` enforced on every RPC.
10. **Pagination:** cursor-based on every List RPC.

## 8. Build order (compute-first)

1. Save `PLAN.md` and `TODO.md` (this document, and the build checklist).
2. Set up Cargo workspace at `compute/`.
3. Implement Earth-Observation Rust crates: `eo-indices`, `eo-radiometric`, `eo-pansharpen`, `eo-geometric`, `eo-mosaic`, `eo-sar`, `eo-atmos-corr`.
4. Implement ADCS Rust crates: `adcs-ekf`, `adcs-ukf`, `adcs-control`, `adcs-actuators`, `orbit-prop`.
5. Implement C&DH Rust crates: `cdh-ccsds`, `cdh-edac`, `cdh-fs`, `cdh-scheduler`.
6. Implement EPS Rust crates: `eps-power`, `eps-battery`.
7. Implement simulation crates: `sim-physics`, `sim-harness`.
8. Implement GS Rust crates: `gs-tle`, `gs-pass-pred`, `gs-doppler`, `gs-bit-sync`, `gs-fec`, `gs-antenna`, `gs-rf-driver`.
9. Implement GEOINT Rust crates: `gi-fusion-eng`, `gi-abi`, `gi-tile-render`, `gi-report-render`, `gi-export`.
10. Build Python ML packages: `eo_ml`, `gi_ml`, `ml_serving`, plus workers.
11. Build Go cross-cutting services: `iam`, `audit`, `notify`.
12. Build Earth Observation Go services: `eo-catalog`, `eo-pipeline`, `eo-analytics`.
13. Build Small Satellite Go services: `sat-mission`, `sat-fsw`, `sat-telemetry`, `sat-command`, `sat-simulation`.
14. Build Ground Station Go services: `gs-mc`, `gs-rf`, `gs-scheduler`, `gs-ingest`.
15. Build GEOINT Go services: `gi-fusion`, `gi-analytics`, `gi-tiles`, `gi-reports`, `gi-predict`.
16. Build worker daemons: `compute/bins/eo-worker`, `gs-worker`, `gi-worker`.
17. Deploy artefacts: Dockerfiles, docker-compose, Helm charts.

## 9. Engineering guardrails

- **No stubs, placeholders, or TODOs in committed code.**
- **No hallucinated APIs**; every external library is verified to exist and the API used is the actual API of the version pinned.
- **No assumptions** about runtime behaviour without tests covering them.
- **Enterprise-grade quality bar:** structured errors, observability, validation, configuration, graceful shutdown, tests, benchmarks (where relevant), documentation comments on every exported symbol.

## 10. Out of scope (explicit)

- P9e Chetana platform integration (replaced by `iam` / `audit` / `notify`).
- Trained ML model weights — Python packages provide inference plumbing only.
- Hardware-specific HALs for satellite flight crates — flight crates are
  hardware-agnostic libraries; a board crate is added per mission.
- Sen2Cor / 6S / MODTRAN binaries themselves — the `eo-atmos-corr` crate
  invokes them as external processes.
