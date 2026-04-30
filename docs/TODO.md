# Build Checklist — Space Technology Platform

Source plan: [`PLAN.md`](./PLAN.md). Implementation order is **compute-first**:
all Rust crates and Python ML packages are completed before any Go service is
written. Each item is enterprise-grade and contains **no stubs, placeholders,
or TODOs**.

## Legend
- `[ ]` not started · `[~]` in progress · `[x]` complete

## Phase 0 — Workspace plumbing
- [x] `go.work` (root)
- [x] `Makefile` (tools / proto / sqlc / build / test / vet)
- [x] `docs/PLAN.md`
- [x] `docs/TODO.md`
- [ ] `compute/Cargo.toml` (Rust workspace)
- [ ] `compute/rust-toolchain.toml`
- [ ] `compute/.cargo/config.toml`
- [ ] `compute/rustfmt.toml`
- [ ] `compute/clippy.toml`

## Phase 1 — Earth Observation compute (Rust)
- [ ] `compute/crates/eo-indices` — NDVI, EVI, SAVI, NDWI
- [ ] `compute/crates/eo-radiometric` — TOA / BOA reflectance
- [ ] `compute/crates/eo-pansharpen` — Brovey / IHS / PCA / Gram-Schmidt
- [ ] `compute/crates/eo-geometric` — DEM-based orthorectification
- [ ] `compute/crates/eo-mosaic` — seamline + blending
- [ ] `compute/crates/eo-sar` — speckle filter, terrain correction, polarimetric
- [ ] `compute/crates/eo-atmos-corr` — Sen2Cor / 6S / MODTRAN process wrapper

## Phase 2 — ADCS compute (Rust)
- [ ] `compute/crates/adcs-ekf` — Extended Kalman Filter (separate crate)
- [ ] `compute/crates/adcs-ukf` — Unscented Kalman Filter (separate crate)
- [ ] `compute/crates/adcs-control` — PID / LQR / SMC / MPC controllers
- [ ] `compute/crates/adcs-actuators` — RW / MTQ / thruster drivers
- [ ] `compute/crates/orbit-prop` — SGP4 / SDP4, RK4, RK78

## Phase 3 — Command & Data Handling (Rust)
- [ ] `compute/crates/cdh-ccsds` — Space Packet Protocol, TC/TM frames, COP-1
- [ ] `compute/crates/cdh-edac` — Hamming + RS scrubbing
- [ ] `compute/crates/cdh-fs` — log-structured on-board file system
- [ ] `compute/crates/cdh-scheduler` — time-tagged command execution

## Phase 4 — Electrical Power System (Rust)
- [ ] `compute/crates/eps-power` — power budget, MPPT, eclipse predictor
- [ ] `compute/crates/eps-battery` — SOC / SOH estimators

## Phase 5 — Simulation (Rust)
- [ ] `compute/crates/sim-physics` — 6-DOF dynamics, gravity, magnetic, drag, SRP
- [ ] `compute/crates/sim-harness` — SITL / HITL / Monte-Carlo runner

## Phase 6 — Ground Station compute (Rust)
- [ ] `compute/crates/gs-tle` — TLE parsing → ECI/ECEF
- [ ] `compute/crates/gs-pass-pred` — pass prediction with horizon mask
- [ ] `compute/crates/gs-doppler` — uplink/downlink frequency comp
- [ ] `compute/crates/gs-bit-sync` — bit & frame synchronization
- [ ] `compute/crates/gs-fec` — Reed-Solomon / LDPC / Turbo / convolutional
- [ ] `compute/crates/gs-antenna` — program / auto / step-track
- [ ] `compute/crates/gs-rf-driver` — SNMP / Modbus drivers

## Phase 7 — GEOINT compute (Rust)
- [ ] `compute/crates/gi-fusion-eng` — multi-layer raster fusion
- [ ] `compute/crates/gi-abi` — pattern-of-life / ABI engine
- [ ] `compute/crates/gi-tile-render` — MVT generator
- [ ] `compute/crates/gi-report-render` — report + PDF
- [ ] `compute/crates/gi-export` — GeoJSON / Shapefile / KML / GeoPackage

## Phase 8 — Worker daemons (Rust binaries)
- [ ] `compute/bins/eo-worker` — Kafka consumer for EO jobs
- [ ] `compute/bins/gs-worker` — Kafka consumer for GS jobs
- [ ] `compute/bins/gi-worker` — Kafka consumer for GI jobs

## Phase 9 — Python ML
- [ ] `ml/packages/ml_serving` — ONNX / Triton / TorchServe inference server
- [ ] `ml/packages/eo_ml` — YOLO / Faster-RCNN / U-Net / DeepLabV3+ / ViT
- [ ] `ml/packages/gi_ml` — risk maps, agricultural forecasting
- [ ] `ml/workers/eo-ml-worker`
- [ ] `ml/workers/gi-ml-worker`

## Phase 10 — Proto + shared Go infra
- [ ] `proto/p9e/space/common/v1`
- [ ] `proto/p9e/space/iam/v1`
- [ ] `proto/p9e/space/audit/v1`
- [ ] `proto/p9e/space/notify/v1`
- [ ] `pkg/config`
- [ ] `pkg/observability`
- [ ] `pkg/authz`
- [ ] `pkg/httpserver`
- [ ] `pkg/middleware`
- [ ] `pkg/db`
- [ ] `pkg/pagination`
- [ ] `pkg/validation`
- [ ] `pkg/errs`
- [ ] `pkg/ids`
- [ ] `pkg/timeutil`
- [ ] `pkg/kafka`
- [ ] `pkg/objectstore`
- [ ] `pkg/geo`
- [ ] `pkg/stac`
- [ ] `pkg/tle`
- [ ] `pkg/closer`

## Phase 11 — Go cross-cutting services
- [ ] `services/iam`
- [ ] `services/audit`
- [ ] `services/notify`

## Phase 12 — Earth Observation Go services
- [ ] `proto/p9e/space/earthobs/v1`
- [ ] `services/eo-catalog`
- [ ] `services/eo-pipeline`
- [ ] `services/eo-analytics`

## Phase 13 — Small Satellite Go services
- [ ] `proto/p9e/space/satsubsys/v1`
- [ ] `services/sat-mission`
- [ ] `services/sat-fsw`
- [ ] `services/sat-telemetry`
- [ ] `services/sat-command`
- [ ] `services/sat-simulation`

## Phase 14 — Ground Station Go services
- [ ] `proto/p9e/space/groundstation/v1`
- [ ] `services/gs-mc`
- [ ] `services/gs-rf`
- [ ] `services/gs-scheduler`
- [ ] `services/gs-ingest`

## Phase 15 — GEOINT Go services
- [ ] `proto/p9e/space/geoint/v1`
- [ ] `services/gi-fusion`
- [ ] `services/gi-analytics`
- [ ] `services/gi-tiles`
- [ ] `services/gi-reports`
- [ ] `services/gi-predict`

## Phase 16 — Deploy
- [ ] `deploy/docker/*` per service Dockerfiles
- [ ] `deploy/compose/docker-compose.yaml`
- [ ] `deploy/helm/*` per service Helm charts

## Per-crate Definition of Done
A crate is considered complete when **all** of the following hold:
- Public API documented with `///` rustdoc on every exported symbol.
- Unit tests covering nominal and edge-case paths.
- Integration test or runnable example demonstrating real usage.
- `cargo fmt` clean, `cargo clippy --all-targets --all-features -- -D warnings` clean.
- `cargo test --all-features` green.
- No `todo!()`, `unimplemented!()`, `unreachable!()` (except for proven-unreachable enum arms), `panic!()` in normal paths, `.unwrap()` outside tests, or "TBD"/"placeholder" comments.
- `Cargo.toml` dependencies pinned to exact versions verified against crates.io.

## Per-Go-service Definition of Done
A service is considered complete when **all** of the following hold:
- Proto file lints clean under `buf lint` and the generated stubs compile.
- `sqlc generate` succeeds; the schema migrates cleanly under Atlas.
- All RPCs validated via `protovalidate`.
- Handler / service / repository / mapper layers all populated with real code (no empty function bodies).
- Health, readiness, OTEL, and Prometheus endpoints wired in `main.go`.
- `go vet ./...` and `go test ./...` green.
- `Dockerfile` builds a distroless image.
