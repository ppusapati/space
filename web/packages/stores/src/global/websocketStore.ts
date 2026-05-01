/**
 * WebSocket Store
 *
 * Real-time communication for live updates:
 * - Automatic reconnection with exponential backoff
 * - Message queuing when disconnected
 * - Subscription-based channels
 * - Heartbeat/ping-pong
 * - Authentication support
 */

import { writable, derived, get } from 'svelte/store';

// ============================================================================
// Types
// ============================================================================

export type WebSocketStatus = 'disconnected' | 'connecting' | 'connected' | 'reconnecting' | 'error';

export interface WebSocketMessage<T = unknown> {
  type: string;
  channel?: string;
  payload: T;
  timestamp?: number;
  id?: string;
}

export interface WebSocketConfig {
  /** WebSocket URL */
  url: string;
  /** Auto reconnect (default: true) */
  autoReconnect: boolean;
  /** Max reconnect attempts (default: 10) */
  maxReconnectAttempts: number;
  /** Initial reconnect delay in ms (default: 1000) */
  reconnectDelay: number;
  /** Max reconnect delay in ms (default: 30000) */
  maxReconnectDelay: number;
  /** Heartbeat interval in ms (default: 30000, 0 to disable) */
  heartbeatInterval: number;
  /** Auth token or function to get token */
  auth?: string | (() => string | Promise<string>);
  /** Protocols */
  protocols?: string | string[];
  /** Queue messages when disconnected */
  queueOffline: boolean;
  /** Max queue size (default: 100) */
  maxQueueSize: number;
}

export interface WebSocketState {
  status: WebSocketStatus;
  lastConnectedAt: Date | null;
  lastDisconnectedAt: Date | null;
  reconnectAttempts: number;
  error: string | null;
  queuedMessages: WebSocketMessage[];
  subscriptions: Set<string>;
}

type MessageHandler<T = unknown> = (message: WebSocketMessage<T>) => void;

// ============================================================================
// Store Implementation
// ============================================================================

const DEFAULT_CONFIG: Omit<WebSocketConfig, 'url'> = {
  autoReconnect: true,
  maxReconnectAttempts: 10,
  reconnectDelay: 1000,
  maxReconnectDelay: 30000,
  heartbeatInterval: 30000,
  queueOffline: true,
  maxQueueSize: 100,
};

function createWebSocketStore() {
  const { subscribe, set, update } = writable<WebSocketState>({
    status: 'disconnected',
    lastConnectedAt: null,
    lastDisconnectedAt: null,
    reconnectAttempts: 0,
    error: null,
    queuedMessages: [],
    subscriptions: new Set(),
  });

  let socket: WebSocket | null = null;
  let config: WebSocketConfig | null = null;
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  let heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  const handlers = new Map<string, Set<MessageHandler>>();
  const channelHandlers = new Map<string, Set<MessageHandler>>();

  function clearTimers() {
    if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null; }
    if (heartbeatTimer) { clearInterval(heartbeatTimer); heartbeatTimer = null; }
  }

  function startHeartbeat() {
    if (!config?.heartbeatInterval) return;
    heartbeatTimer = setInterval(() => {
      if (socket?.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({ type: 'ping', timestamp: Date.now() }));
      }
    }, config.heartbeatInterval);
  }

  function flushQueue() {
    const state = get({ subscribe });
    if (socket?.readyState !== WebSocket.OPEN || state.queuedMessages.length === 0) return;

    for (const message of state.queuedMessages) {
      socket.send(JSON.stringify(message));
    }
    update(s => ({ ...s, queuedMessages: [] }));
  }

  function scheduleReconnect() {
    if (!config?.autoReconnect) return;

    const state = get({ subscribe });
    if (state.reconnectAttempts >= config.maxReconnectAttempts) {
      update(s => ({ ...s, status: 'error', error: 'Max reconnection attempts reached' }));
      return;
    }

    const delay = Math.min(
      config.reconnectDelay * Math.pow(2, state.reconnectAttempts),
      config.maxReconnectDelay
    );

    update(s => ({ ...s, status: 'reconnecting', reconnectAttempts: s.reconnectAttempts + 1 }));

    reconnectTimer = setTimeout(() => connect(), delay);
  }

  async function connect() {
    if (!config) return;
    if (socket?.readyState === WebSocket.OPEN) return;

    update(s => ({ ...s, status: 'connecting', error: null }));

    try {
      let url = config.url;

      // Add auth token if provided
      if (config.auth) {
        const token = typeof config.auth === 'function' ? await config.auth() : config.auth;
        const separator = url.includes('?') ? '&' : '?';
        url = `${url}${separator}token=${encodeURIComponent(token)}`;
      }

      socket = new WebSocket(url, config.protocols);

      socket.onopen = () => {
        update(s => ({
          ...s,
          status: 'connected',
          lastConnectedAt: new Date(),
          reconnectAttempts: 0,
          error: null,
        }));
        startHeartbeat();
        flushQueue();

        // Resubscribe to channels
        const state = get({ subscribe });
        for (const channel of state.subscriptions) {
          socket?.send(JSON.stringify({ type: 'subscribe', channel }));
        }
      };

      socket.onclose = (event) => {
        clearTimers();
        update(s => ({ ...s, status: 'disconnected', lastDisconnectedAt: new Date() }));

        if (!event.wasClean && config?.autoReconnect) {
          scheduleReconnect();
        }
      };

      socket.onerror = () => {
        update(s => ({ ...s, status: 'error', error: 'WebSocket error occurred' }));
      };

      socket.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);

          // Handle pong
          if (message.type === 'pong') return;

          // Call type handlers
          const typeHandlers = handlers.get(message.type);
          if (typeHandlers) {
            for (const handler of typeHandlers) handler(message);
          }

          // Call channel handlers
          if (message.channel) {
            const chHandlers = channelHandlers.get(message.channel);
            if (chHandlers) {
              for (const handler of chHandlers) handler(message);
            }
          }

          // Call wildcard handlers
          const wildcardHandlers = handlers.get('*');
          if (wildcardHandlers) {
            for (const handler of wildcardHandlers) handler(message);
          }
        } catch { /* ignore parse errors */ }
      };
    } catch (error) {
      update(s => ({ ...s, status: 'error', error: (error as Error).message }));
      scheduleReconnect();
    }
  }

  return {
    subscribe,

    /** Connect to WebSocket server */
    connect(options: WebSocketConfig) {
      config = { ...DEFAULT_CONFIG, ...options };
      connect();
    },

    /** Disconnect from server */
    disconnect() {
      clearTimers();
      if (socket) {
        socket.close(1000, 'Client disconnect');
        socket = null;
      }
      update(s => ({ ...s, status: 'disconnected', subscriptions: new Set() }));
    },

    /** Send a message */
    send<T>(type: string, payload: T, channel?: string): boolean {
      const message: WebSocketMessage<T> = {
        type,
        payload,
        channel,
        timestamp: Date.now(),
        id: `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      };

      if (socket?.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(message));
        return true;
      }

      if (config?.queueOffline) {
        update(s => {
          const queue = [...s.queuedMessages, message];
          if (queue.length > (config?.maxQueueSize || 100)) queue.shift();
          return { ...s, queuedMessages: queue };
        });
      }
      return false;
    },

    /** Subscribe to a channel */
    subscribeToChannel(channel: string, handler?: MessageHandler) {
      update(s => {
        const subs = new Set(s.subscriptions);
        subs.add(channel);
        return { ...s, subscriptions: subs };
      });

      if (socket?.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({ type: 'subscribe', channel }));
      }

      if (handler) {
        if (!channelHandlers.has(channel)) channelHandlers.set(channel, new Set());
        channelHandlers.get(channel)!.add(handler);
      }

      return () => this.unsubscribeFromChannel(channel, handler);
    },

    /** Unsubscribe from a channel */
    unsubscribeFromChannel(channel: string, handler?: MessageHandler) {
      update(s => {
        const subs = new Set(s.subscriptions);
        subs.delete(channel);
        return { ...s, subscriptions: subs };
      });

      if (socket?.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({ type: 'unsubscribe', channel }));
      }

      if (handler) {
        channelHandlers.get(channel)?.delete(handler);
      } else {
        channelHandlers.delete(channel);
      }
    },

    /** Add message type handler */
    on<T = unknown>(type: string, handler: MessageHandler<T>): () => void {
      if (!handlers.has(type)) handlers.set(type, new Set());
      handlers.get(type)!.add(handler as MessageHandler);
      return () => this.off(type, handler);
    },

    /** Remove message type handler */
    off<T = unknown>(type: string, handler: MessageHandler<T>) {
      handlers.get(type)?.delete(handler as MessageHandler);
    },

    /** Clear all handlers */
    clearHandlers() {
      handlers.clear();
      channelHandlers.clear();
    },

    /** Force reconnect */
    reconnect() {
      this.disconnect();
      if (config) connect();
    },

    /** Clear message queue */
    clearQueue() {
      update(s => ({ ...s, queuedMessages: [] }));
    },

    /** Destroy and cleanup */
    destroy() {
      this.disconnect();
      this.clearHandlers();
    },
  };
}

export const websocketStore = createWebSocketStore();

// Derived stores
export const wsStatus = derived(websocketStore, $s => $s.status);
export const wsConnected = derived(websocketStore, $s => $s.status === 'connected');
export const wsQueueSize = derived(websocketStore, $s => $s.queuedMessages.length);

// Convenience exports
export function connectWebSocket(options: WebSocketConfig) {
  websocketStore.connect(options);
}

export function disconnectWebSocket() {
  websocketStore.disconnect();
}

export function sendMessage<T>(type: string, payload: T, channel?: string) {
  return websocketStore.send(type, payload, channel);
}
