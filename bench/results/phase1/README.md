# Phase-1 NFR baseline results

TASK-P1-NFR-001 acceptance #1: both benches green for two consecutive runs against ephemeral clusters with the documented hardware profile.

## Hardware profile

The phase-1 NFR gate (`.github/workflows/nfr-phase1.yml`) runs both benches against an ephemeral AWS cluster brought up by the workflow's `bring-up` job (placeholder until `tools/bench-cluster/up.sh` lands; see TASK-P1-CLUSTER-001, future). The profile encoded in the workflow:

| Component       | Instance type     | Count | Notes |
|-----------------|-------------------|-------|-------|
| IAM service     | `m6i.xlarge`      | 3     | 4 vCPU, 16 GB; behind ALB |
| realtime-gw     | `c6i.2xlarge`     | 3     | 8 vCPU, 16 GB; behind NLB |
| Postgres (RDS)  | `db.r6g.xlarge`   | 1     | 4 vCPU, 32 GB |
| Redis (ElastiCache) | `cache.r6g.large` | 1 | per-replica session-affinity disabled |
| Kafka (MSK)     | `m5.large`        | 3     | replication factor 3 |
| k6 runner       | `c6i.2xlarge`     | 1     | sized so the runner is NOT the bottleneck at 1k req/s + 10k WS |

## Acceptance gates

| Bench | Requirement | Threshold | Source |
|-------|-------------|-----------|--------|
| `iam-login.bench.js` | REQ-NFR-PERF-005 | p95 < 100 ms at 1 000 req/s sustained over 5 min | `bench/k6/iam-login.bench.js` |
| `realtime-fanout.bench.js` | REQ-NFR-PERF-006 | p95 push latency < 500 ms at 10 000 concurrent WS connections | `bench/k6/realtime-fanout.bench.js` |

The IAM bench's per-tag p95 budgets:

| Tag | Budget | Why |
|-----|--------|-----|
| `login-ok`  | p95 < 100 ms | The dominant flow's gate |
| `webauthn`  | p95 < 100 ms | Passkey assertion path |
| `refresh`   | p95 < 100 ms | Refresh-token rotation |
| `login-bad` | 200 ms ≤ p95 < 350 ms | The constant-time floor (REQ-FUNC-PLT-IAM-010) means `login-bad` MUST sit above 200 ms; we cap the upper bound at 350 ms to catch regressions in the constant-time mechanism (e.g. an early return that makes wrong-password fast = enumerable). |

The realtime bench measures **end-to-end push latency**: the producer harness stamps `server_emitted_at_ms` into every event payload; each k6 VU computes `Date.now() - server_emitted_at_ms` per received frame. The p95 budget covers producer → Kafka → realtime-gw Redis fan-out → backpressure buffer → WebSocket frame.

## How to read a result file

`task -t bench/Taskfile.yml phase1` and the workflow both produce two files per bench run under this directory:

```
iam-login.run1.summary.json     # k6 --summary-export — per-metric aggregates
iam-login.run1.json             # k6 --out json       — every sample (large)
iam-login.run2.summary.json
iam-login.run2.json
realtime-fanout.run1.summary.json
realtime-fanout.run1.json
realtime-fanout.run2.summary.json
realtime-fanout.run2.json
```

The summary JSON's `metrics.<name>.values` block carries the `p(95)`, `rate`, `count`, etc. fields the k6 thresholds key off. A passing run reports `metrics.<name>.thresholds: {<expr>: {ok: true}}` for every threshold in the bench script. A failing run flips `ok: false` AND k6 exits non-zero — that exit code is what the workflow's `gate` job reads.

## Baseline (TBD)

The first green pair of runs gets archived here as `baseline-2026-Q2.md` with:

- the cluster spec (resolved AMI ids, RDS engine version, k6 version)
- the exact bench environment vars (so a reviewer can repro)
- the four summary JSON files
- a one-page narrative covering anything noteworthy (cold-start p95 vs steady-state, per-tag breakdown, regression-against-prior-baseline if applicable)

Until the bring-up job is wired (TASK-P1-CLUSTER-001), this section is a placeholder and the workflow's `gate` job intentionally fails when the cluster URLs are empty — better to fail loudly than to silently green-light a missing measurement.

## Local invocation

For developer-local sanity (a chetana platform stood up via docker-compose):

```bash
# 1. Bring up the chetana platform locally with bench-friendly seed
#    data (see deploy/docker/docker-compose.bench.yaml — future).
# 2. Run both benches.
task -t bench/Taskfile.yml phase1
# 3. Read the verdict.
task -t bench/Taskfile.yml report:phase1
```

The local profile won't hit the production thresholds (a single laptop CPU isn't going to sustain 1k req/s for 5 min); the local task is for catching gross regressions during dev. The gate-of-record is the GitHub Actions workflow.
