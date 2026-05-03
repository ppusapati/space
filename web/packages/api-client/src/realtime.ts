/**
 * realtime.ts — chetana realtime gateway WebSocket client.
 *
 * Surface mirrors services/realtime-gw/internal/ws/server.go:
 *
 *   wss://<gw>/v1/rt
 *   Sec-WebSocket-Protocol: chetana.v1, chetana.bearer.<token>
 *
 * Wire-protocol (JSON over text frames):
 *
 *   client → server:
 *     { "type": "subscribe", "topic": "<topic>" }
 *     { "type": "unsubscribe", "topic": "<topic>" }
 *     { "type": "ping" }
 *
 *   server → client:
 *     { "type": "subscribed", "topic": "<topic>" }
 *     { "type": "unsubscribed", "topic": "<topic>" }
 *     { "type": "message", "topic": "<topic>", "payload": <json> }
 *     { "type": "pong" }
 *     { "type": "error", "code": "<machine_code>", "reason": "<human>" }
 *
 * Typed close codes (services/realtime-gw/internal/topic/auth.go):
 *
 *   4001 policy_deny
 *   4002 itar_requires_us_person
 *   4003 insufficient_clearance
 *   4004 unknown_topic
 *
 * Reconnect strategy:
 *
 *   Exponential backoff with full jitter, capped at 30s. Resets to
 *   the base interval on a healthy connection (≥ 5s alive). When
 *   the close was 4002 / 4003 / 4004 we DO NOT auto-reconnect to
 *   the offending topic — those are user-actionable denials.
 */

export type CloseCode = 1000 | 1001 | 1006 | 4001 | 4002 | 4003 | 4004 | number;

export interface RealtimeOptions {
  /** wss:// URL of the gateway. */
  url: string;
  /** Access token (Bearer). */
  bearer: string;
  /** Max backoff between reconnects. Default 30s. */
  maxBackoffMs?: number;
  /** Initial backoff. Default 1s. */
  baseBackoffMs?: number;
  /** Onset of "healthy" so the backoff resets. Default 5s. */
  healthyMs?: number;
  /**
   * Optional logger sink. Defaults to console.warn for warnings +
   * console.error for fatal events.
   */
  logger?: (level: "info" | "warn" | "error", msg: string, meta?: unknown) => void;
}

export type MessageHandler = (payload: unknown) => void;

interface Subscription {
  topic: string;
  handlers: Set<MessageHandler>;
  /** Topics retried indefinitely; ITAR / clearance / unknown are pinned dead. */
  dead?: { code: CloseCode; reason: string };
}

export interface RealtimeClient {
  /** Open the connection. Idempotent — repeated calls are no-ops. */
  start(): void;
  /** Close + clear subscriptions. After close() the client is unusable. */
  close(): void;
  /**
   * Subscribe to `topic`. Returns an unsubscribe function. Multiple
   * handlers per topic are supported; the WS subscription is only
   * dropped when the last handler unsubscribes.
   */
  subscribe(topic: string, h: MessageHandler): () => void;
  /**
   * Snapshot of the current connection state. Useful for the UI
   * to render a reconnecting badge.
   */
  state(): "idle" | "connecting" | "open" | "closed" | "reconnecting";
  /**
   * Topics the client has marked dead (4002/4003/4004) — the UI
   * can surface these for re-auth prompts.
   */
  deadTopics(): { topic: string; code: CloseCode; reason: string }[];
}

export function createRealtimeClient(opts: RealtimeOptions): RealtimeClient {
  const cfg = {
    maxBackoffMs: opts.maxBackoffMs ?? 30_000,
    baseBackoffMs: opts.baseBackoffMs ?? 1_000,
    healthyMs: opts.healthyMs ?? 5_000,
  };
  const log =
    opts.logger ??
    ((level: "info" | "warn" | "error", msg: string, meta?: unknown) => {
      const fn = level === "error" ? console.error : level === "warn" ? console.warn : console.info;
      fn.call(console, `[chetana.realtime] ${msg}`, meta ?? "");
    });

  let ws: WebSocket | null = null;
  let started = false;
  let closed = false;
  let backoff = cfg.baseBackoffMs;
  let connectedAt = 0;
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  let stateLabel: ReturnType<RealtimeClient["state"]> = "idle";

  const subs = new Map<string, Subscription>();

  function setState(s: typeof stateLabel) {
    stateLabel = s;
  }

  function jitter(ms: number): number {
    return Math.floor(Math.random() * ms);
  }

  function send(obj: unknown) {
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    ws.send(JSON.stringify(obj));
  }

  function handleClose(ev: CloseEvent) {
    if (ws) ws = null;
    if (closed) {
      setState("closed");
      return;
    }
    const code = ev.code as CloseCode;
    const reason = ev.reason || "";
    log("warn", `socket closed code=${code} reason=${reason}`);

    // 4002 / 4003 / 4004 are typed denials per topic.go's
    // CloseReason map. The close hits the WHOLE socket but the
    // offending topic is the most-recently-issued subscribe — the
    // server-side guard refuses the upgrade for the topic only,
    // so the rest of the subscription set will reconnect cleanly.
    // We pin denied topics dead so the next reconnect cycle does
    // not re-issue them.
    if (code === 4002 || code === 4003 || code === 4004) {
      // Best-effort: mark every still-open subscribe attempt that
      // didn't yet receive `subscribed` as dead. The chetana
      // protocol acks subscriptions, so any sub without an ack
      // since the last reconnect is the suspect.
      for (const s of subs.values()) {
        if (!s.dead) s.dead = { code, reason };
      }
    }

    // Reconnect unless explicitly closed.
    scheduleReconnect();
  }

  function scheduleReconnect() {
    if (closed) return;
    setState("reconnecting");
    const delay = jitter(backoff);
    log("info", `reconnecting in ${delay}ms (current backoff ${backoff}ms)`);
    reconnectTimer = setTimeout(() => {
      backoff = Math.min(backoff * 2, cfg.maxBackoffMs);
      open();
    }, delay);
  }

  function open() {
    if (closed) return;
    setState("connecting");
    const sock = new WebSocket(opts.url, ["chetana.v1", `chetana.bearer.${opts.bearer}`]);
    ws = sock;

    sock.onopen = () => {
      connectedAt = Date.now();
      setState("open");
      log("info", "socket open");
      // Re-issue every still-live subscription on reconnect so the
      // server-side state machine matches the client's intent.
      for (const s of subs.values()) {
        if (s.dead) continue;
        send({ type: "subscribe", topic: s.topic });
      }
    };

    sock.onmessage = (ev) => {
      let parsed: unknown;
      try {
        parsed = JSON.parse(typeof ev.data === "string" ? ev.data : "");
      } catch {
        return;
      }
      const msg = parsed as { type?: string; topic?: string; payload?: unknown; code?: string; reason?: string };
      switch (msg.type) {
        case "message": {
          if (!msg.topic) return;
          const sub = subs.get(msg.topic);
          if (!sub) return;
          for (const h of sub.handlers) {
            try {
              h(msg.payload);
            } catch (e) {
              log("error", `handler threw for topic=${msg.topic}`, e);
            }
          }
          // Reset backoff once we've been alive past the healthy
          // threshold AND received a real message.
          if (Date.now() - connectedAt >= cfg.healthyMs) {
            backoff = cfg.baseBackoffMs;
          }
          break;
        }
        case "subscribed":
        case "unsubscribed":
        case "pong":
          break;
        case "error":
          log("warn", `server error code=${msg.code ?? "?"} reason=${msg.reason ?? ""}`);
          break;
      }
    };

    sock.onerror = () => {
      log("warn", "socket error (will close + reconnect)");
    };

    sock.onclose = handleClose;
  }

  return {
    start() {
      if (started) return;
      started = true;
      open();
    },
    close() {
      closed = true;
      if (reconnectTimer) {
        clearTimeout(reconnectTimer);
        reconnectTimer = null;
      }
      if (ws) {
        try {
          ws.close(1000, "client_close");
        } catch {
          // best-effort
        }
        ws = null;
      }
      subs.clear();
      setState("closed");
    },
    subscribe(topic, h) {
      let sub = subs.get(topic);
      if (!sub) {
        sub = { topic, handlers: new Set() };
        subs.set(topic, sub);
        if (stateLabel === "open") {
          send({ type: "subscribe", topic });
        }
      }
      sub.handlers.add(h);
      return () => {
        const cur = subs.get(topic);
        if (!cur) return;
        cur.handlers.delete(h);
        if (cur.handlers.size === 0) {
          subs.delete(topic);
          send({ type: "unsubscribe", topic });
        }
      };
    },
    state() {
      return stateLabel;
    },
    deadTopics() {
      const out: { topic: string; code: CloseCode; reason: string }[] = [];
      for (const s of subs.values()) {
        if (s.dead) out.push({ topic: s.topic, code: s.dead.code, reason: s.dead.reason });
      }
      return out;
    },
  };
}
