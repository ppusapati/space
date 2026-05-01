/**
 * Session Timeout Store
 *
 * Manages user session timeout and auto-logout:
 * - Track user activity (mouse, keyboard, touch)
 * - Configurable timeout duration
 * - Warning before logout
 * - Extension on activity
 * - Backend sync for session validation
 */

import { writable, derived, get } from 'svelte/store';

// ============================================================================
// Types
// ============================================================================

export interface SessionTimeoutConfig {
  /** Timeout duration in milliseconds (default: 30 minutes) */
  timeout: number;
  /** Warning time before logout in milliseconds (default: 5 minutes) */
  warningTime: number;
  /** Events to track for activity */
  activityEvents: string[];
  /** Whether to extend on activity */
  extendOnActivity: boolean;
  /** Callback when session expires */
  onTimeout?: () => void;
  /** Callback when warning is shown */
  onWarning?: (remainingTime: number) => void;
  /** Endpoint for session validation */
  validateEndpoint?: string;
  /** Interval to validate session with backend */
  validateInterval?: number;
}

export interface SessionTimeoutState {
  isActive: boolean;
  lastActivity: Date;
  expiresAt: Date | null;
  isWarning: boolean;
  remainingTime: number;
  isValidating: boolean;
}

// ============================================================================
// Constants
// ============================================================================

const DEFAULT_CONFIG: SessionTimeoutConfig = {
  timeout: 30 * 60 * 1000, // 30 minutes
  warningTime: 5 * 60 * 1000, // 5 minutes
  activityEvents: ['mousedown', 'mousemove', 'keydown', 'scroll', 'touchstart', 'click'],
  extendOnActivity: true,
};

// ============================================================================
// Store Implementation
// ============================================================================

function createSessionTimeoutStore() {
  const { subscribe, set, update } = writable<SessionTimeoutState>({
    isActive: false,
    lastActivity: new Date(),
    expiresAt: null,
    isWarning: false,
    remainingTime: 0,
    isValidating: false,
  });

  let config: SessionTimeoutConfig = { ...DEFAULT_CONFIG };
  let timeoutTimer: ReturnType<typeof setTimeout> | null = null;
  let warningTimer: ReturnType<typeof setTimeout> | null = null;
  let countdownTimer: ReturnType<typeof setInterval> | null = null;
  let validateTimer: ReturnType<typeof setInterval> | null = null;
  let activityThrottle: ReturnType<typeof setTimeout> | null = null;

  function clearTimers() {
    if (timeoutTimer) { clearTimeout(timeoutTimer); timeoutTimer = null; }
    if (warningTimer) { clearTimeout(warningTimer); warningTimer = null; }
    if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null; }
    if (validateTimer) { clearInterval(validateTimer); validateTimer = null; }
  }

  function startCountdown() {
    if (countdownTimer) clearInterval(countdownTimer);
    countdownTimer = setInterval(() => {
      const state = get({ subscribe });
      if (!state.expiresAt) return;
      const remaining = Math.max(0, state.expiresAt.getTime() - Date.now());
      update(s => ({ ...s, remainingTime: remaining }));
      if (remaining <= 0) {
        clearInterval(countdownTimer!);
        countdownTimer = null;
      }
    }, 1000);
  }

  function scheduleTimers() {
    clearTimers();
    const now = new Date();
    const expiresAt = new Date(now.getTime() + config.timeout);
    const warningAt = config.timeout - config.warningTime;

    update(s => ({ ...s, lastActivity: now, expiresAt, isWarning: false, remainingTime: config.timeout }));

    // Warning timer
    warningTimer = setTimeout(() => {
      update(s => ({ ...s, isWarning: true }));
      startCountdown();
      config.onWarning?.(config.warningTime);
    }, warningAt);

    // Timeout timer
    timeoutTimer = setTimeout(() => {
      handleTimeout();
    }, config.timeout);

    // Validation interval
    if (config.validateEndpoint && config.validateInterval) {
      validateTimer = setInterval(() => validateSession(), config.validateInterval);
    }
  }

  async function validateSession(): Promise<boolean> {
    if (!config.validateEndpoint) return true;
    update(s => ({ ...s, isValidating: true }));
    try {
      // 'omit' for Bearer-token auth — see DEPLOYMENT_READINESS.md item
      // 45 round 3 for the CORS wildcard-origin trap rationale.
      const response = await fetch(config.validateEndpoint, { method: 'POST', credentials: 'omit' });
      update(s => ({ ...s, isValidating: false }));
      if (!response.ok) { handleTimeout(); return false; }
      return true;
    } catch {
      update(s => ({ ...s, isValidating: false }));
      return false;
    }
  }

  function handleTimeout() {
    clearTimers();
    update(s => ({ ...s, isActive: false, expiresAt: null, isWarning: false, remainingTime: 0 }));
    removeActivityListeners();
    config.onTimeout?.();
  }

  function handleActivity() {
    if (activityThrottle) return;
    activityThrottle = setTimeout(() => { activityThrottle = null; }, 1000);

    const state = get({ subscribe });
    if (!state.isActive || !config.extendOnActivity) return;

    // Reset timers on activity (only if not in warning state, or always if extendOnActivity is true)
    scheduleTimers();
  }

  function addActivityListeners() {
    if (typeof window === 'undefined') return;
    config.activityEvents.forEach(event => window.addEventListener(event, handleActivity, { passive: true }));
  }

  function removeActivityListeners() {
    if (typeof window === 'undefined') return;
    config.activityEvents.forEach(event => window.removeEventListener(event, handleActivity));
  }

  return {
    subscribe,

    /**
     * Initialize and start session timeout tracking
     */
    start(options?: Partial<SessionTimeoutConfig>) {
      config = { ...DEFAULT_CONFIG, ...options };
      update(s => ({ ...s, isActive: true }));
      addActivityListeners();
      scheduleTimers();
    },

    /**
     * Stop session timeout tracking
     */
    stop() {
      clearTimers();
      removeActivityListeners();
      if (activityThrottle) { clearTimeout(activityThrottle); activityThrottle = null; }
      set({ isActive: false, lastActivity: new Date(), expiresAt: null, isWarning: false, remainingTime: 0, isValidating: false });
    },

    /**
     * Extend session by resetting timers
     */
    extend() {
      const state = get({ subscribe });
      if (!state.isActive) return;
      scheduleTimers();
    },

    /**
     * Force logout immediately
     */
    logout() {
      handleTimeout();
    },

    /**
     * Update configuration
     */
    configure(options: Partial<SessionTimeoutConfig>) {
      config = { ...config, ...options };
    },

    /**
     * Validate session with backend
     */
    validate: validateSession,

    /**
     * Get remaining time in readable format
     */
    getFormattedRemainingTime(): string {
      const state = get({ subscribe });
      const minutes = Math.floor(state.remainingTime / 60000);
      const seconds = Math.floor((state.remainingTime % 60000) / 1000);
      return `${minutes}:${seconds.toString().padStart(2, '0')}`;
    },
  };
}

export const sessionTimeoutStore = createSessionTimeoutStore();

// Derived stores
export const isSessionActive = derived(sessionTimeoutStore, $s => $s.isActive);
export const isSessionWarning = derived(sessionTimeoutStore, $s => $s.isWarning);
export const sessionRemainingTime = derived(sessionTimeoutStore, $s => $s.remainingTime);

// Convenience exports
export function startSessionTimeout(options?: Partial<SessionTimeoutConfig>) {
  sessionTimeoutStore.start(options);
}

export function stopSessionTimeout() {
  sessionTimeoutStore.stop();
}

export function extendSession() {
  sessionTimeoutStore.extend();
}
