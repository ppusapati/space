// bench/k6/scaffold.bench.js — TASK-P0-INFRA-001 acceptance criterion #5.
//
// This bench targets the example service in
// services/packages/observability/serverobs/example/. It is a smoke
// check, not a real NFR gate — it proves the bench harness wires up
// end-to-end (k6 binary, _lib/checks.js, Taskfile recipe, p95 reporting)
// before any phase-specific bench is authored.
//
// Run:
//   task -t bench/Taskfile.yml bench:scaffold
//   # or directly:
//   k6 run -e CHETANA_BENCH_NOAUTH=true bench/k6/scaffold.bench.js
//
// Expected: p95 < 50ms (the example service is a no-op /health
// handler, so any latency above that points at the harness or the
// developer workstation, not the code).

import http from 'k6/http';
import { check, sleep } from 'k6';
import { smokeThresholds } from './_lib/checks.js';

const baseURL = __ENV.CHETANA_BENCH_BASE_URL || 'http://localhost:8080';

export const options = {
  // 10 VUs ramping over 30s. Total iterations ~ 6 000 — enough samples
  // for a stable p95.
  stages: [
    { duration: '5s',  target: 5  },
    { duration: '20s', target: 10 },
    { duration: '5s',  target: 0  },
  ],
  thresholds: smokeThresholds(),
  // Tag every request so the JSON output groups cleanly.
  tags: { bench: 'scaffold' },
  // Fail the whole run if any threshold breaks; CI relies on the exit
  // code to mark the bench gate.
  noConnectionReuse: false,
  discardResponseBodies: true,
};

export default function () {
  // Three endpoints exercised in rotation — mirrors a real probe loop
  // (orchestrator hits /health + /ready, scraper hits /metrics).
  const endpoints = ['/health', '/ready', '/metrics'];
  for (const path of endpoints) {
    const res = http.get(baseURL + path, { tags: { name: path } });
    check(res, {
      [`${path} status 200`]: (r) => r.status === 200,
    });
  }
  // Tiny inter-iteration pause so 10 VUs don't saturate the loopback
  // socket pool on the developer machine.
  sleep(0.05);
}

// k6 calls handleSummary once at the end of the run. We emit a
// minimal JSON file so the Taskfile recipe can grep for the p95.
export function handleSummary(data) {
  const p95 = data.metrics?.http_req_duration?.values?.['p(95)'];
  const checks = data.metrics?.checks?.values?.rate;
  const failed = data.metrics?.http_req_failed?.values?.rate;
  return {
    'bench/results/phase0/scaffold.json': JSON.stringify(
      {
        bench: 'scaffold',
        p95_ms: p95,
        check_pass_rate: checks,
        request_failure_rate: failed,
        budget_p95_ms: 50,
        budget_error_rate: 0.01,
        thresholdsPassed: !data.metrics?.http_req_duration?.thresholds?.['p(95)<50']?.failed,
      },
      null,
      2,
    ),
    stdout: textSummary(data),
  };
}

// textSummary is k6's built-in summariser; we re-export so the default
// console block still renders.
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.2/index.js';
