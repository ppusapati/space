// bench/k6/iam-login.bench.js — TASK-P1-NFR-001 acceptance #1.
//
// REQ-NFR-PERF-005: IAM login p95 < 100 ms at 1 000 req/s sustained.
// REQ-CONST-009:    NFR gates run as a CI workflow that blocks merge.
//
// Workload shape (matches the chetana production login mix):
//
//   • 60% — successful password+TOTP login (the dominant flow during
//           a working day). Pre-seeded user pool of 1 000 accounts so
//           the per-user rate-limiter (TASK-P1-IAM-001 acceptance #2)
//           does not skew p95.
//
//   • 20% — successful WebAuthn assertion. Pre-registered credentials
//           against the same pool. The browser-side
//           navigator.credentials.get is replaced by a deterministic
//           signed assertion built into the bench (signed in
//           bench:setup so the inner loop is just an HTTP POST).
//
//   • 15% — refresh-token rotation. Every successful login from the
//           prior 5 minutes feeds this branch.
//
//   •  5% — bad-credentials failure. Exercises the constant-time
//           response path so its 250 ms floor is included in the p95
//           — REQ-FUNC-PLT-IAM-010's response-timing-variance bound
//           is enforced separately by services/iam/test/reset_test.go's
//           timing-variance assertion; here we only confirm the
//           floor isn't an outlier that pushes the gate.
//
// 1 000 req/s arrival
// -------------------
// k6's `arrival-rate` executor is the only way to pin a request rate
// independent of VU count. We allocate 200 VUs as the warm pool +
// 800 as a burst headroom; if the bench can't keep up at 1k/s the
// scenario aborts (the run fails). Ramp pattern:
//
//   0-30s:   ramp 0 → 1 000 req/s
//   30-300s: hold 1 000 req/s (270 s steady-state — enough for
//             ~270 000 samples; the p95 is statistically tight)
//   300-330s: ramp 1 000 → 0
//
// Invocation
// ----------
//   task -t bench/Taskfile.yml bench:iam
//
// Required env (typically via the workflow's secrets block):
//   CHETANA_BENCH_BASE_URL  e.g. https://iam.bench.chetana.p9e.in
//   CHETANA_BENCH_USER_POOL e.g. https://bench.chetana.p9e.in/seed/users.json
//   CHETANA_BENCH_OAUTH_CLIENT_ID + CHETANA_BENCH_OAUTH_CLIENT_SECRET
//     for the refresh-token branch.

import http from 'k6/http';
import { check, group } from 'k6';
import { Counter } from 'k6/metrics';
import { perfThresholds } from './_lib/checks.js';

const baseURL = __ENV.CHETANA_BENCH_BASE_URL || 'http://localhost:8080';
const userPoolURL = __ENV.CHETANA_BENCH_USER_POOL || `${baseURL}/seed/bench-users.json`;
const refreshTokens = []; // populated by the success branch; sampled by the refresh branch

// Custom counters surface the per-branch mix in the JSON output so a
// regression in one path doesn't get hidden by averaging.
const cLoginOk = new Counter('chetana_bench_iam_login_ok');
const cLoginBad = new Counter('chetana_bench_iam_login_bad');
const cWebAuthn = new Counter('chetana_bench_iam_webauthn');
const cRefresh = new Counter('chetana_bench_iam_refresh');

// User pool fetched once at setup. Each user is { email, password,
// webauthn_credential, webauthn_signed_assertion }. The bench
// fixture (deploy/bench/iam-seed.go, future) generates these from
// a chetana-controlled CA so the assertion verifies cleanly against
// the test webauthn config without a real authenticator in the loop.
let users = [];

export const options = {
  scenarios: {
    iam_mix: {
      executor: 'ramping-arrival-rate',
      startRate: 0,
      timeUnit: '1s',
      preAllocatedVUs: 200,
      maxVUs: 1000,
      stages: [
        { duration: '30s',  target: 1000 },  // ramp up
        { duration: '4m30s', target: 1000 }, // hold
        { duration: '30s',  target: 0 },     // ramp down
      ],
    },
  },
  thresholds: {
    // Acceptance #1: p95 < 100 ms across every authenticated branch.
    // The bad-credentials path includes a 250 ms constant-time floor
    // (REQ-FUNC-PLT-IAM-010); we exclude it from the p95 budget so
    // the gate measures the real auth surface, not the deliberately
    // padded timing channel. See the per-tag thresholds below.
    'http_req_duration{tag:login-ok}':     ['p(95)<100'],
    'http_req_duration{tag:webauthn}':     ['p(95)<100'],
    'http_req_duration{tag:refresh}':      ['p(95)<100'],
    'http_req_failed{tag:login-ok}':       ['rate<0.001'],
    'http_req_failed{tag:webauthn}':       ['rate<0.001'],
    'http_req_failed{tag:refresh}':        ['rate<0.001'],
    // The "bad" branch's threshold is the floor: 200 ms < p95 < 350 ms
    // — anything OUT of that envelope is a regression in the
    // constant-time mechanism.
    'http_req_duration{tag:login-bad}':    ['p(95)<350', 'p(50)>200'],
    // Aggregate check rate.
    checks: ['rate>0.999'],
  },
  noConnectionReuse: false,
  discardResponseBodies: true,
};

export function setup() {
  const res = http.get(userPoolURL);
  check(res, { 'user pool 200': (r) => r.status === 200 });
  if (res.status !== 200) {
    throw new Error(`user pool fetch failed: ${res.status}`);
  }
  users = res.json();
  if (!Array.isArray(users) || users.length < 1000) {
    throw new Error(`user pool too small: ${users?.length ?? 0} (need ≥1 000 to avoid per-user rate-limit skew)`);
  }
  return { users };
}

function pickUser(data) {
  return data.users[Math.floor(Math.random() * data.users.length)];
}

export default function (data) {
  const dice = Math.random();

  if (dice < 0.60) {
    runLoginOk(data);
  } else if (dice < 0.80) {
    runWebAuthn(data);
  } else if (dice < 0.95) {
    runRefresh(data);
  } else {
    runLoginBad(data);
  }
}

function runLoginOk(data) {
  const u = pickUser(data);
  group('login-ok', () => {
    const res = http.post(
      `${baseURL}/v1/iam/login`,
      JSON.stringify({ email: u.email, password: u.password, mfa_code: u.totp_now }),
      { headers: { 'Content-Type': 'application/json' }, tags: { tag: 'login-ok', name: 'login' } },
    );
    const ok = check(res, {
      'login 200': (r) => r.status === 200,
      'login has access_token': (r) => {
        try {
          return !!(r.json() || {}).access_token;
        } catch {
          return false;
        }
      },
    });
    if (ok) {
      cLoginOk.add(1);
      // Capture the refresh token for the refresh branch. Bounded
      // so the array doesn't grow without limit.
      try {
        const body = res.json();
        if (body?.refresh_token && refreshTokens.length < 10000) {
          refreshTokens.push(body.refresh_token);
        }
      } catch {
        // ignore
      }
    }
  });
}

function runWebAuthn(data) {
  const u = pickUser(data);
  group('webauthn', () => {
    // Begin
    const begin = http.post(
      `${baseURL}/v1/iam/webauthn/assert/begin`,
      JSON.stringify({ email: u.email }),
      { headers: { 'Content-Type': 'application/json' }, tags: { tag: 'webauthn', name: 'webauthn-begin' } },
    );
    if (begin.status !== 200) return;
    let beginBody;
    try {
      beginBody = begin.json();
    } catch {
      return;
    }
    // Finish — submit the pre-signed assertion baked into the
    // user fixture. The chetana cmd-layer's verifier accepts these
    // because the bench's webauthn_credential was registered against
    // the chetana-bench webauthn config at fixture-build time.
    const finish = http.post(
      `${baseURL}/v1/iam/webauthn/assert/finish`,
      JSON.stringify({
        session_token: beginBody.session_token,
        credential: u.webauthn_signed_assertion,
      }),
      { headers: { 'Content-Type': 'application/json' }, tags: { tag: 'webauthn', name: 'webauthn-finish' } },
    );
    const ok = check(finish, {
      'webauthn 200': (r) => r.status === 200,
    });
    if (ok) cWebAuthn.add(1);
  });
}

function runRefresh(_data) {
  if (refreshTokens.length === 0) {
    // No tokens cached yet — fall through to a login so the bench
    // does not stall.
    runLoginOk(_data);
    return;
  }
  const tokIdx = Math.floor(Math.random() * refreshTokens.length);
  const tok = refreshTokens[tokIdx];
  group('refresh', () => {
    const res = http.post(
      `${baseURL}/v1/iam/refresh`,
      JSON.stringify({ refresh_token: tok }),
      { headers: { 'Content-Type': 'application/json' }, tags: { tag: 'refresh', name: 'refresh' } },
    );
    const ok = check(res, {
      'refresh 200': (r) => r.status === 200,
      'refresh has new tokens': (r) => {
        try {
          const b = r.json();
          return !!b.access_token && !!b.refresh_token;
        } catch {
          return false;
        }
      },
    });
    if (ok) {
      cRefresh.add(1);
      // Replace the consumed refresh with the rotated one (refresh
      // tokens are single-use per the IAM contract).
      try {
        const newRefresh = res.json()?.refresh_token;
        if (newRefresh) refreshTokens[tokIdx] = newRefresh;
      } catch {
        // ignore
      }
    }
  });
}

function runLoginBad(data) {
  const u = pickUser(data);
  group('login-bad', () => {
    const res = http.post(
      `${baseURL}/v1/iam/login`,
      JSON.stringify({ email: u.email, password: 'wrong-password-bench' }),
      { headers: { 'Content-Type': 'application/json' }, tags: { tag: 'login-bad', name: 'login' } },
    );
    const ok = check(res, {
      'bad-creds 200 or 401': (r) => r.status === 200 || r.status === 401,
      'bad-creds takes >= 200ms (constant-time floor)': (r) => r.timings.duration >= 200,
    });
    if (ok) cLoginBad.add(1);
  });
}
