# P9E Space — Microsystems Space Technology Platform

A multi-module backend for **Earth Observation analytics**, **Small Satellite
subsystems**, **Ground Station software**, and **Geospatial Intelligence**.
Compute is Rust (host-side workspace + a separate embedded-flight workspace),
ML is Python (ONNX Runtime + FastAPI), and the control-plane is **17
ConnectRPC services in Go**, all built from a single shared infrastructure
package (`p9e.in/samavaya/packages`).

## Repository layout

```
space/
├── compute/                  # Rust host-side workspace (data/RF/imagery crates)
│   └── crates/
│       ├── eo-*              # Earth-observation: indices, radiometric, geometric,
│       │                       atmospheric, pan-sharpen, mosaic, SAR
│       ├── gs-*              # Ground station: TLE, pass prediction, Doppler,
│       │                       FEC, bit-sync, antenna, RF driver
│       ├── gi-*              # GEOINT: fusion-engine, tile-render, report-render,
│       │                       ABI, export
│       ├── orbit-prop        # SGP4/SDP4 + RK4 + Dormand-Prince RK45
│       └── sim-*             # SITL/HITL physics + harness
├── flight/                   # Rust embedded workspace (cross-compiles to MCU/SoC)
│   └── crates/
│       ├── adcs-*            # ADCS: EKF, UKF, control (PID/LQR/SMC/MPC), actuators
│       ├── cdh-*             # CDH: CCSDS SPP, EDAC (Hamming SEC-DED), filesystem,
│       │                       scheduler
│       └── eps-*             # EPS: battery model, power management
├── ml/                       # Python ML packages and workers
│   ├── packages/
│   │   ├── ml_serving        # Predictor protocol, model registry, FastAPI, JobBus
│   │   ├── eo_ml             # Detector, segmenter, classifier, NMS, sliding-window
│   │   └── gi_ml             # flood_risk, drought_spi, forecast_ndvi_trend, ...
│   └── workers/
│       ├── eo-ml-worker      # Kafka-driven EO inference worker
│       └── gi-ml-worker      # Kafka-driven GEOINT worker
├── services/                 # Go ConnectRPC services (this is the control plane)
│   ├── go.work
│   ├── buf.yaml              # v2 workspace listing all proto modules
│   ├── packages/             # p9e.in/samavaya/packages — shared infra layer
│   ├── eo-catalog/           ┐
│   ├── eo-pipeline/          │  Earth Observation
│   ├── eo-analytics/         ┘
│   ├── sat-mission/          ┐
│   ├── sat-fsw/              │
│   ├── sat-telemetry/        │  Small Satellite Subsystems
│   ├── sat-command/          │
│   ├── sat-simulation/       ┘
│   ├── gs-mc/                ┐
│   ├── gs-rf/                │  Ground Station
│   ├── gs-scheduler/         │
│   ├── gs-ingest/            ┘
│   ├── gi-fusion/            ┐
│   ├── gi-analytics/         │
│   ├── gi-tiles/             │  Geospatial Intelligence (GEOINT)
│   ├── gi-reports/           │
│   └── gi-predict/           ┘
├── deploy/
│   ├── docker/               # docker-compose for the full local stack
│   └── helm/space/           # Umbrella Helm chart for Kubernetes
├── docs/
│   ├── PLAN.md
│   └── TODO.md
├── Makefile
├── requirements              # Original requirements document
└── README.md
```

## Service catalogue

All 17 Go services follow a single canonical template:

| Module       | Service          | Purpose                                                              |
|--------------|------------------|----------------------------------------------------------------------|
| EO           | eo-catalog       | STAC catalog: collections, items, assets, quality results            |
| EO           | eo-pipeline      | Processing-job orchestrator (radiometric, geometric, atmospheric…)   |
| EO           | eo-analytics     | ML-model registry + inference jobs                                   |
| Sat          | sat-mission      | Satellite registry: TLE, orbital state, mission config               |
| Sat          | sat-fsw          | Flight-software builds + deployment manifests (with cross-checks)    |
| Sat          | sat-telemetry    | Channels + frames + samples (transactional ingest)                   |
| Sat          | sat-command      | Command catalog + time-tagged uplink queue (per-sat sequence)        |
| Sat          | sat-simulation   | Reusable scenarios + SITL/HITL run management                        |
| GS           | gs-mc            | Ground stations + antennas (frequency band, polarization)            |
| GS           | gs-rf            | Pre-pass link budgets + live RF measurements (RSSI, SNR, BER)        |
| GS           | gs-scheduler     | Predicted contact passes + tenant bookings (priority, status)        |
| GS           | gs-ingest        | Per-pass ingest sessions + downlink frames (transactional)           |
| GI           | gi-fusion        | Multi-source fusion jobs (pan-sharpen, time-series, multimodal)      |
| GI           | gi-analytics     | GEOINT analysis (NDVI series, change detection, flood/drought)       |
| GI           | gi-tiles         | Tile-set catalog (PNG/MVT/COG/WEBP, EPSG projection, zoom range)     |
| GI           | gi-reports       | Report templates (PDF/HTML/DOCX/XLSX) + generated reports            |
| GI           | gi-predict       | Forecast jobs (NDVI trend, flood risk, drought, weather, custom)     |

Roughly **120 RPCs**, **110 service-level tests**, all using offset/limit
pagination + status state-machines + ULID identifiers.

## Architecture choices

- **Wire format:** ConnectRPC (HTTP/2 + h2c) speaks gRPC, gRPC-Web, and
  HTTP/JSON from the same handler.
- **Validation:** `buf.build/go/protovalidate` reads validation rules from
  the `.proto` definitions; no hand-written validators.
- **Identifiers:** [ULID](https://github.com/ulid/spec) (Crockford Base32,
  26 chars) everywhere; stored in Postgres `uuid` columns as 16-byte
  payloads.
- **Errors:** typed via `packages/errors` with HTTP-style codes
  (BadRequest 400, NotFound 404, Conflict 409, PreconditionFailed 412…),
  mapped to Connect codes at the handler edge.
- **Pagination:** `packages/proto/pagination.PaginationRequest/Response`
  (offset/limit + total_count via SQL window function).
- **Audit fields:** `packages/proto/fields.Fields` (uuid, is_active,
  created_by/at, updated_by/at).
- **Database:** PostgreSQL via `pgx/v5` + `sqlc` for type-safe queries.
  Each service owns one database; per-service schema migrations live in
  `services/<svc>/db/schema/`.
- **Configuration:** `services/<svc>/config/config.yaml` (non-secret
  defaults) + env-var overlay for secrets (`DATABASE_URL`,
  `ALLOWED_ORIGINS`). Loaded via `p9e.in/samavaya/packages/config`.
- **Container build:** distroless static images, each ~10–15 MB.
- **Per-service layout:**

  ```
  services/<svc>/
  ├── proto/<svc>.proto        # service-owned proto, one package per service
  ├── api/                     # buf-generated stubs
  ├── db/
  │   ├── sqlc.yaml
  │   ├── schema/              # forward-only migrations
  │   ├── queries/             # SQL with sqlc.arg()/sqlc.narg() params
  │   └── generated/           # sqlc output
  ├── cmd/main.go              # pgxpool + repo + svc + handler + pkgserver
  ├── internal/
  │   ├── handler/             # ConnectRPC handlers (protovalidate)
  │   ├── mapper/              # proto ↔ domain ↔ sqlc converters
  │   ├── models/              # domain types (ulid.ID, time.Time)
  │   ├── repository/          # sqlc wrapper (returns rows + total)
  │   ├── services/            # business logic, transition graphs
  │   └── config/              # packages/config loader
  ├── config/config.yaml       # non-secret defaults
  ├── buf.gen.yaml
  ├── Dockerfile               # distroless static, ENTRYPOINT
  ├── go.mod
  └── .gitignore
  ```

## Quick start

### Prerequisites

- Go **1.26.1+**
- [`buf`](https://buf.build/docs/installation) **v1.69+**
- [`sqlc`](https://docs.sqlc.dev/en/latest/overview/install.html) **v1.31+**
- Rust **1.79+** (for `compute/` and `flight/` crates)
- Python **3.11+** (for `ml/`)
- Docker (for the local stack)
- (optional) `helm` **3.12+** (for Kubernetes)

### Local stack (docker-compose)

```bash
cd deploy/docker
docker compose build
docker compose up -d
docker compose ps
curl http://localhost:18000/health   # eo-catalog
```

Each service is exposed on a unique host port in **18000–18016**; see
`deploy/README.md` for the full mapping.

### Build a single service from source

```bash
cd services/eo-catalog
go build ./...
go test ./...
```

The whole Go workspace lives under `services/go.work`, so:

```bash
cd services
for svc in eo-catalog eo-pipeline eo-analytics \
           sat-mission sat-fsw sat-telemetry sat-command sat-simulation \
           gs-mc gs-rf gs-scheduler gs-ingest \
           gi-fusion gi-analytics gi-tiles gi-reports gi-predict; do
  (cd $svc && go test ./...) || break
done
```

### Regenerate proto stubs and sqlc

From any service directory:

```bash
buf generate                 # regenerate api/<svc>.pb.go + .connect.go
(cd db && sqlc generate)     # regenerate db/generated/
```

The buf workspace at `services/buf.yaml` lets each service's proto
import the shared types from `services/packages/proto/`.

### Compute (Rust host-side)

```bash
cd compute
cargo build --workspace
cargo test --workspace
cargo clippy --workspace -- -D warnings
```

### Flight (Rust embedded)

```bash
cd flight
cargo build --workspace
cargo test --workspace
```

The `flight/` workspace is intentionally separate from `compute/` so it
can be cross-compiled to MCU targets (cortex-m, RISC-V) without dragging
host-only dependencies through the resolver.

### ML packages

```bash
cd ml/packages/ml_serving && pytest -q
cd ml/packages/eo_ml      && pytest -q
cd ml/packages/gi_ml      && pytest -q
```

Workers under `ml/workers/` are launched via Docker or `python -m`.

## Kubernetes deployment

```bash
kubectl create secret generic postgres-credentials \
  --from-literal=password=<password> \
  -n p9e-space

helm install space deploy/helm/space \
  --namespace p9e-space --create-namespace \
  --set global.imageRegistry=ghcr.io/yourorg \
  --set global.imageTag=0.1.0
```

The umbrella chart renders a Deployment + Service + ConfigMap for each
enabled entry under `services:` in `values.yaml`. See
`deploy/README.md` for `values.<env>.yaml` patterns and per-service
overrides (replicas, resources, CORS allowlist).

## Module ownership

| Module      | Lead language | Notes                                             |
|-------------|---------------|---------------------------------------------------|
| `compute/`  | Rust          | Host-side data crunching; deterministic, no_std-ready where possible |
| `flight/`   | Rust          | Embedded ADCS/CDH/EPS; cross-compiles to MCU      |
| `ml/`       | Python        | ONNX Runtime + FastAPI; Kafka-driven workers      |
| `services/` | Go            | Control-plane ConnectRPC services + sqlc          |
| `services/packages/` | Go    | Shared infrastructure: config, server, errors, ulid, p9log, observability, middleware, db, etc. |

## Status

All 17 services build and test green. The platform is ready for:
- **Local development** via `docker compose up`.
- **Kubernetes deployment** via the Helm umbrella chart.
- **Continued feature work** on top of the canonical service template
  (proto + sqlc + 5-layer internal/ + packages/connect/server).

See `docs/PLAN.md` for the architecture document and `docs/TODO.md` for
the remaining backlog.
