// bench/k6/_lib/checks.js — shared p95 / error-rate threshold helpers.
//
// Phase NFR gates (Phase 1 IAM, Phase 2 telemetry, Phase 3 EO/STAC,
// Phase 6 cutover) all assert the same shape of SLO: latency p95 below
// a budget AND error rate below a budget. This module exports the
// k6 `thresholds` object so individual benches express the SLO as
// data, not boilerplate.
//
// Usage (per-bench):
//
//   import { perfThresholds } from './_lib/checks.js';
//   export const options = {
//     thresholds: perfThresholds({ p95Ms: 100, errorRate: 0.001 }),
//     ...
//   };

/**
 * perfThresholds builds the standard k6 thresholds map enforcing p95
 * latency and error-rate budgets.
 *
 * @param {{ p95Ms: number, errorRate?: number }} budget
 * @returns {Record<string, string[]>}
 */
export function perfThresholds(budget) {
  if (typeof budget?.p95Ms !== 'number' || budget.p95Ms <= 0) {
    throw new Error('perfThresholds: budget.p95Ms must be a positive number');
  }
  const errorRate = budget.errorRate ?? 0.001; // 0.1% default
  return {
    // p95 latency budget. abortOnFail keeps long load tests honest.
    http_req_duration: [`p(95)<${budget.p95Ms}`],
    // Error-rate budget. checks=>rate aggregates pass/fail of every
    // `check(...)` in the script.
    checks: [`rate>${1 - errorRate}`],
    // Failed-request rate (k6 default: any non-2xx).
    http_req_failed: [`rate<${errorRate}`],
  };
}

/**
 * smokeThresholds is the looser budget used by the Phase-0 scaffold
 * bench: p95 < 50ms, error rate < 1%. Real Phase-N gates use
 * perfThresholds with their requirement-derived budgets.
 */
export function smokeThresholds() {
  return perfThresholds({ p95Ms: 50, errorRate: 0.01 });
}
