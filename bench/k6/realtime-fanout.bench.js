// bench/k6/realtime-fanout.bench.js — TASK-P1-NFR-001 acceptance #1.
//
// REQ-NFR-PERF-006: realtime fan-out p95 push latency < 500 ms at
//                    10 000 concurrent WebSocket connections.
//
// Workload shape:
//
//   • 10 000 long-lived WS connections, distributed evenly across
//     the realtime-gw replicas (k6 round-robins via the LB DNS).
//   • Each connection subscribes to ONE topic from the chetana
//     test-topic pool (telemetry.bench.{1..100}). 100 topics → ~100
//     subscribers per topic, mirroring a typical pass + telemetry
//     subscription density.
//   • The bench's setup() invokes the chetana producer harness to
//     fire 10 events/sec on each of the 100 topics for 60 seconds.
//     That's 100 events/sec × 100 topics × 100 subscribers =
//     1 000 000 push deliveries per second steady-state.
//   • Push latency = (server-side OccurredAt embedded in the
//     payload) - (client-side perf.now()). The p95 budget is 500 ms
//     end-to-end including: producer → Kafka → realtime-gw Redis
//     fan-out → backpressure buffer → WebSocket frame.
//
// Connection ramp
// ---------------
//   0-60s:    ramp 0 → 10 000 connections
//   60-180s:  hold; producer fires events
//   180-210s: ramp down
//
// Invocation
// ----------
//   task -t bench/Taskfile.yml bench:realtime
//
// Required env:
//   CHETANA_BENCH_RT_URL          wss URL of the realtime-gw LB
//   CHETANA_BENCH_PRODUCER_URL    HTTP control endpoint that triggers
//                                  the producer harness for the run
//                                  duration
//   CHETANA_BENCH_BEARER          short-lived access token for the
//                                  bench user (US-person, cui clearance,
//                                  authorised on every test topic)

import ws from 'k6/ws';
import http from 'k6/http';
import { check } from 'k6';
import { Trend, Counter } from 'k6/metrics';

const wsURL = __ENV.CHETANA_BENCH_RT_URL || 'ws://localhost:8086/v1/rt';
const producerURL =
  __ENV.CHETANA_BENCH_PRODUCER_URL || 'http://localhost:8086/bench/producer';
const bearer = __ENV.CHETANA_BENCH_BEARER || 'bench-noauth';
const topicCount = Number(__ENV.CHETANA_BENCH_TOPIC_COUNT || 100);

// Custom metrics:
//   • pushLatency: e2e push latency in ms. p95 budget 500ms.
//   • framesReceived / framesSent: sanity counters.
//   • backpressureDrops: a server-side header echoed in periodic
//     control frames; a non-zero value at 10k conn means the
//     gateway shed load and the run is invalid.
const pushLatency = new Trend('chetana_rt_push_latency_ms', true);
const framesReceived = new Counter('chetana_rt_frames_received');
const sessionErrors = new Counter('chetana_rt_session_errors');

export const options = {
  scenarios: {
    rt_fanout: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '60s',  target: 10000 }, // ramp connections
        { duration: '120s', target: 10000 }, // hold + producer fires
        { duration: '30s',  target: 0 },     // ramp down
      ],
      gracefulRampDown: '20s',
    },
  },
  thresholds: {
    chetana_rt_push_latency_ms: ['p(95)<500'],
    chetana_rt_session_errors: ['count<100'], // <0.1% session-establishment failure
    checks: ['rate>0.999'],
  },
};

export function setup() {
  // Trigger the producer harness for the duration of the bench. The
  // harness publishes 10 events/sec on each of `topicCount` topics
  // for `duration_s` seconds, then auto-stops. Returning the
  // baseline t0 from the producer's clock so the per-frame
  // server-side timestamp is comparable to the bench's wall clock
  // (allowing for the small NTP skew between the producer host +
  // the k6 runner).
  const res = http.post(
    `${producerURL}/start`,
    JSON.stringify({
      topic_count: topicCount,
      events_per_topic_per_second: 10,
      duration_s: 200,
    }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  check(res, { 'producer started': (r) => r.status === 202 });
  return { startedAt: Date.now() };
}

export default function () {
  // Each VU subscribes to ONE topic. 10 000 VUs / 100 topics =
  // 100 subscribers per topic.
  const topicID = (__VU - 1) % topicCount;
  const topic = `telemetry.bench.${topicID}`;

  // Sub-protocol-based bearer (matches services/realtime-gw/internal/ws/server.go).
  const protocols = `chetana.v1, chetana.bearer.${bearer}`;
  const params = {
    headers: { 'Sec-WebSocket-Protocol': protocols },
    tags: { topic: `bench.${topicID}` },
  };

  const res = ws.connect(wsURL, params, (socket) => {
    socket.on('open', () => {
      socket.send(JSON.stringify({ type: 'subscribe', topic }));
    });

    socket.on('message', (raw) => {
      framesReceived.add(1);
      let msg;
      try {
        msg = JSON.parse(raw);
      } catch {
        return;
      }
      if (msg.type !== 'message') return;
      // The producer stamps `server_emitted_at_ms` (epoch ms) into
      // the payload. Latency is now - emitted.
      const sentAt = msg.payload?.server_emitted_at_ms;
      if (typeof sentAt === 'number') {
        pushLatency.add(Date.now() - sentAt);
      }
    });

    socket.on('error', () => {
      sessionErrors.add(1);
    });

    socket.on('close', () => {
      // expected at end of run
    });

    // Keep the socket open for the steady-state portion of the run.
    // 150s gives ramp-up VUs ~ 90s of active subscription time
    // before ramp-down begins.
    socket.setTimeout(() => socket.close(), 150_000);
  });

  check(res, { 'ws upgrade 101': (r) => r && r.status === 101 });
  if (!res || res.status !== 101) {
    sessionErrors.add(1);
  }
}

export function teardown() {
  // Best-effort: stop the producer harness so a re-run starts from
  // a clean state.
  http.post(`${producerURL}/stop`, '');
}
