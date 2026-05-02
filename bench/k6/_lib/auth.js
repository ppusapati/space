// bench/k6/_lib/auth.js — shared k6 helper for IAM token acquisition.
//
// Phase-1+ benches that target authenticated endpoints import getToken
// once per VU, cache the result for the test run, and stamp the
// Authorization header on every request.
//
// Phase-0 (this PR) benches do not need real IAM yet — getToken returns
// a stub bearer when CHETANA_BENCH_NOAUTH=true so the scaffold bench
// can exercise the example service without IAM standing up. As soon as
// the real IAM lands (TASK-P1-IAM-002) we flip the default and add the
// expected env vars to bench/Taskfile.yml.

import http from 'k6/http';
import { check, fail } from 'k6';

const noAuth = (__ENV.CHETANA_BENCH_NOAUTH || '').toLowerCase() === 'true';
const tokenURL =
  __ENV.CHETANA_IAM_TOKEN_URL ||
  'https://iam.chetana.p9e.in/oauth2/token';
const clientID = __ENV.CHETANA_BENCH_CLIENT_ID || '';
const clientSecret = __ENV.CHETANA_BENCH_CLIENT_SECRET || '';

// In-memory cache keyed by (clientID). Single-VU benches use one entry;
// multi-VU benches share the cache so they don't all hammer /oauth2/token.
const cache = {};

/**
 * getToken returns a bearer token for the configured client. The first
 * call per (clientID) hits the IAM service; subsequent calls return
 * the cached value. When CHETANA_BENCH_NOAUTH=true a stub bearer is
 * returned without any HTTP call.
 */
export function getToken() {
  if (noAuth) return 'bench-noauth-stub-token';
  if (cache[clientID]) return cache[clientID];

  if (!clientID || !clientSecret) {
    fail('CHETANA_BENCH_CLIENT_ID / CHETANA_BENCH_CLIENT_SECRET unset and CHETANA_BENCH_NOAUTH not enabled');
  }

  const res = http.post(
    tokenURL,
    {
      grant_type: 'client_credentials',
      client_id: clientID,
      client_secret: clientSecret,
    },
    { tags: { name: 'bench-iam-token' } },
  );
  check(res, {
    'IAM token 200': (r) => r.status === 200,
    'IAM token has access_token': (r) => !!(r.json() || {}).access_token,
  }) || fail(`IAM token acquisition failed: status=${res.status} body=${res.body}`);

  cache[clientID] = res.json().access_token;
  return cache[clientID];
}

/**
 * authHeaders returns the canonical Authorization headers map.
 */
export function authHeaders() {
  return { Authorization: `Bearer ${getToken()}` };
}
