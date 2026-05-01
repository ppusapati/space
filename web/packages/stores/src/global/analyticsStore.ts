/**
 * Analytics Store - User Behavior Tracking & Telemetry
 *
 * Features:
 * - Page view tracking
 * - Event tracking
 * - User session management
 * - Performance metrics
 * - Error tracking
 * - Custom dimensions
 * - Multiple provider support
 * - Privacy-aware (consent management)
 */

import { writable, derived, get } from 'svelte/store';

// ═══════════════════════════════════════════════════════════════════════════
// TYPES
// ═══════════════════════════════════════════════════════════════════════════

export interface AnalyticsEvent {
  /** Event name */
  name: string;
  /** Event category */
  category?: string;
  /** Event action */
  action?: string;
  /** Event label */
  label?: string;
  /** Event value (numeric) */
  value?: number;
  /** Custom properties */
  properties?: Record<string, unknown>;
  /** Timestamp */
  timestamp: number;
}

export interface PageView {
  /** Page path */
  path: string;
  /** Page title */
  title?: string;
  /** Referrer */
  referrer?: string;
  /** Custom dimensions */
  dimensions?: Record<string, string>;
  /** Timestamp */
  timestamp: number;
}

export interface UserProperties {
  /** User ID */
  userId?: string;
  /** Tenant ID */
  tenantId?: string;
  /** User email */
  email?: string;
  /** User roles */
  roles?: string[];
  /** Subscription plan */
  plan?: string;
  /** Custom properties */
  [key: string]: unknown;
}

export type UserSession = SessionInfo;

export interface SessionInfo {
  /** Session ID */
  sessionId: string;
  /** Session start time */
  startTime: number;
  /** Last activity time */
  lastActivity: number;
  /** Page views in session */
  pageViews: number;
  /** Events in session */
  events: number;
  /** Session duration (ms) */
  duration: number;
  /** Device info */
  device?: DeviceInfo;
}

export interface DeviceInfo {
  /** User agent */
  userAgent: string;
  /** Screen resolution */
  screenResolution: string;
  /** Viewport size */
  viewport: string;
  /** Device type */
  deviceType: 'desktop' | 'tablet' | 'mobile';
  /** Browser name */
  browser: string;
  /** Operating system */
  os: string;
  /** Language */
  language: string;
  /** Timezone */
  timezone: string;
}

export interface PerformanceMetric {
  /** Metric name */
  name: string;
  /** Metric value */
  value: number;
  /** Metric unit */
  unit?: string;
  /** Additional labels */
  labels?: Record<string, string>;
  /** Timestamp */
  timestamp: number;
}

export interface ErrorEvent {
  /** Error message */
  message: string;
  /** Error stack */
  stack?: string;
  /** Error type */
  type: 'error' | 'unhandledrejection' | 'network' | 'api';
  /** Error context */
  context?: Record<string, unknown>;
  /** Timestamp */
  timestamp: number;
}

export type ConsentCategory = 'necessary' | 'analytics' | 'marketing' | 'personalization';

export type UserConsent = ConsentState;

export interface ConsentState {
  /** Consent given timestamp */
  timestamp?: number;
  /** Categories consented to */
  categories: Record<ConsentCategory, boolean>;
}

export interface AnalyticsProvider {
  /** Provider name */
  name: string;
  /** Initialize provider */
  init: (config: unknown) => void;
  /** Track page view */
  trackPageView: (pageView: PageView) => void;
  /** Track event */
  trackEvent: (event: AnalyticsEvent) => void;
  /** Identify user */
  identify: (properties: UserProperties) => void;
  /** Track error */
  trackError?: (error: ErrorEvent) => void;
  /** Track performance */
  trackPerformance?: (metric: PerformanceMetric) => void;
}

export interface AnalyticsState {
  /** Whether analytics is enabled */
  enabled: boolean;
  /** Consent state */
  consent: ConsentState;
  /** User properties */
  user: UserProperties;
  /** Session info */
  session: SessionInfo | null;
  /** Event queue (for offline support) */
  queue: (AnalyticsEvent | PageView | ErrorEvent)[];
  /** Debug mode */
  debug: boolean;
  /** Registered providers */
  providers: string[];
}

// ═══════════════════════════════════════════════════════════════════════════
// HELPERS
// ═══════════════════════════════════════════════════════════════════════════

function generateSessionId(): string {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

function getDeviceInfo(): DeviceInfo {
  if (typeof window === 'undefined') {
    return {
      userAgent: '',
      screenResolution: '',
      viewport: '',
      deviceType: 'desktop',
      browser: '',
      os: '',
      language: '',
      timezone: '',
    };
  }

  const ua = navigator.userAgent;
  const width = window.screen.width;
  const height = window.screen.height;

  // Detect device type
  let deviceType: 'desktop' | 'tablet' | 'mobile' = 'desktop';
  if (/Mobi|Android/i.test(ua)) {
    deviceType = width > 768 ? 'tablet' : 'mobile';
  }

  // Detect browser
  let browser = 'Unknown';
  if (ua.includes('Firefox')) browser = 'Firefox';
  else if (ua.includes('Chrome')) browser = 'Chrome';
  else if (ua.includes('Safari')) browser = 'Safari';
  else if (ua.includes('Edge')) browser = 'Edge';

  // Detect OS
  let os = 'Unknown';
  if (ua.includes('Windows')) os = 'Windows';
  else if (ua.includes('Mac')) os = 'macOS';
  else if (ua.includes('Linux')) os = 'Linux';
  else if (ua.includes('Android')) os = 'Android';
  else if (ua.includes('iOS')) os = 'iOS';

  return {
    userAgent: ua,
    screenResolution: `${width}x${height}`,
    viewport: `${window.innerWidth}x${window.innerHeight}`,
    deviceType,
    browser,
    os,
    language: navigator.language,
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
  };
}

// ═══════════════════════════════════════════════════════════════════════════
// STORE
// ═══════════════════════════════════════════════════════════════════════════

const initialState: AnalyticsState = {
  enabled: false,
  consent: {
    categories: {
      necessary: true,
      analytics: false,
      marketing: false,
      personalization: false,
    },
  },
  user: {},
  session: null,
  queue: [],
  debug: false,
  providers: [],
};

const store = writable<AnalyticsState>(initialState);

// Registered providers
const providers: Map<string, AnalyticsProvider> = new Map();

// ═══════════════════════════════════════════════════════════════════════════
// PROVIDER MANAGEMENT
// ═══════════════════════════════════════════════════════════════════════════

/**
 * Register an analytics provider
 */
export function registerProvider(provider: AnalyticsProvider, config?: unknown): void {
  providers.set(provider.name, provider);
  provider.init(config);

  store.update((s) => ({
    ...s,
    providers: [...s.providers, provider.name],
  }));
}

/**
 * Unregister a provider
 */
export function unregisterProvider(name: string): void {
  providers.delete(name);

  store.update((s) => ({
    ...s,
    providers: s.providers.filter((p) => p !== name),
  }));
}

// ═══════════════════════════════════════════════════════════════════════════
// CONSENT MANAGEMENT
// ═══════════════════════════════════════════════════════════════════════════

/**
 * Update consent
 */
export function setConsent(categories: Partial<Record<ConsentCategory, boolean>>): void {
  store.update((s) => ({
    ...s,
    consent: {
      timestamp: Date.now(),
      categories: { ...s.consent.categories, ...categories },
    },
    enabled: categories.analytics ?? s.consent.categories.analytics,
  }));

  // Persist consent
  if (typeof localStorage !== 'undefined') {
    const state = get(store);
    localStorage.setItem('analytics_consent', JSON.stringify(state.consent));
  }
}

/**
 * Check if a consent category is granted
 */
export function hasConsent(category: ConsentCategory): boolean {
  return get(store).consent.categories[category];
}

// ═══════════════════════════════════════════════════════════════════════════
// SESSION MANAGEMENT
// ═══════════════════════════════════════════════════════════════════════════

/**
 * Start a new session
 */
export function startSession(): void {
  const now = Date.now();

  store.update((s) => ({
    ...s,
    session: {
      sessionId: generateSessionId(),
      startTime: now,
      lastActivity: now,
      pageViews: 0,
      events: 0,
      duration: 0,
      device: getDeviceInfo(),
    },
  }));
}

/**
 * End current session
 */
export function endSession(): void {
  store.update((s) => ({ ...s, session: null }));
}

/**
 * Update session activity
 */
function updateSessionActivity(): void {
  const now = Date.now();

  store.update((s) => {
    if (!s.session) return s;

    return {
      ...s,
      session: {
        ...s.session,
        lastActivity: now,
        duration: now - s.session.startTime,
      },
    };
  });
}

// ═══════════════════════════════════════════════════════════════════════════
// TRACKING FUNCTIONS
// ═══════════════════════════════════════════════════════════════════════════

/**
 * Track page view
 */
export function trackPageView(path: string, options?: Partial<Omit<PageView, 'path' | 'timestamp'>>): void {
  const state = get(store);

  if (!state.enabled || !state.consent.categories.analytics) {
    return;
  }

  const pageView: PageView = {
    path,
    title: options?.title ?? (typeof document !== 'undefined' ? document.title : undefined),
    referrer: options?.referrer ?? (typeof document !== 'undefined' ? document.referrer : undefined),
    dimensions: options?.dimensions,
    timestamp: Date.now(),
  };

  // Update session
  store.update((s) => ({
    ...s,
    session: s.session
      ? { ...s.session, pageViews: s.session.pageViews + 1, lastActivity: Date.now() }
      : s.session,
  }));

  // Debug log
  if (state.debug) {
    console.log('[Analytics] Page View:', pageView);
  }

  // Send to providers
  providers.forEach((provider) => {
    try {
      provider.trackPageView(pageView);
    } catch (error) {
      console.error(`[Analytics] Error in ${provider.name}:`, error);
    }
  });
}

/**
 * Track event
 */
export function trackEvent(
  name: string,
  properties?: Record<string, unknown>,
  options?: Partial<Omit<AnalyticsEvent, 'name' | 'properties' | 'timestamp'>>
): void {
  const state = get(store);

  if (!state.enabled || !state.consent.categories.analytics) {
    return;
  }

  const event: AnalyticsEvent = {
    name,
    category: options?.category,
    action: options?.action,
    label: options?.label,
    value: options?.value,
    properties,
    timestamp: Date.now(),
  };

  // Update session
  store.update((s) => ({
    ...s,
    session: s.session
      ? { ...s.session, events: s.session.events + 1, lastActivity: Date.now() }
      : s.session,
  }));

  // Debug log
  if (state.debug) {
    console.log('[Analytics] Event:', event);
  }

  // Send to providers
  providers.forEach((provider) => {
    try {
      provider.trackEvent(event);
    } catch (error) {
      console.error(`[Analytics] Error in ${provider.name}:`, error);
    }
  });
}

/**
 * Identify user
 */
export function identify(properties: UserProperties): void {
  store.update((s) => ({
    ...s,
    user: { ...s.user, ...properties },
  }));

  const state = get(store);

  if (!state.enabled) {
    return;
  }

  // Debug log
  if (state.debug) {
    console.log('[Analytics] Identify:', properties);
  }

  // Send to providers
  providers.forEach((provider) => {
    try {
      provider.identify(properties);
    } catch (error) {
      console.error(`[Analytics] Error in ${provider.name}:`, error);
    }
  });
}

/**
 * Track error
 */
export function trackError(
  message: string,
  options?: Partial<Omit<ErrorEvent, 'message' | 'timestamp'>>
): void {
  const state = get(store);

  if (!state.enabled || !state.consent.categories.analytics) {
    return;
  }

  const errorEvent: ErrorEvent = {
    message,
    stack: options?.stack,
    type: options?.type ?? 'error',
    context: options?.context,
    timestamp: Date.now(),
  };

  // Debug log
  if (state.debug) {
    console.log('[Analytics] Error:', errorEvent);
  }

  // Send to providers
  providers.forEach((provider) => {
    try {
      provider.trackError?.(errorEvent);
    } catch (error) {
      console.error(`[Analytics] Error in ${provider.name}:`, error);
    }
  });
}

/**
 * Track performance metric
 */
export function trackPerformance(
  name: string,
  value: number,
  options?: Partial<Omit<PerformanceMetric, 'name' | 'value' | 'timestamp'>>
): void {
  const state = get(store);

  if (!state.enabled || !state.consent.categories.analytics) {
    return;
  }

  const metric: PerformanceMetric = {
    name,
    value,
    unit: options?.unit,
    labels: options?.labels,
    timestamp: Date.now(),
  };

  // Debug log
  if (state.debug) {
    console.log('[Analytics] Performance:', metric);
  }

  // Send to providers
  providers.forEach((provider) => {
    try {
      provider.trackPerformance?.(metric);
    } catch (error) {
      console.error(`[Analytics] Error in ${provider.name}:`, error);
    }
  });
}

// ═══════════════════════════════════════════════════════════════════════════
// INITIALIZATION
// ═══════════════════════════════════════════════════════════════════════════

/**
 * Initialize analytics
 */
export function initAnalytics(options?: {
  enabled?: boolean;
  debug?: boolean;
  autoTrackPageViews?: boolean;
  autoTrackErrors?: boolean;
}): void {
  const { enabled = false, debug = false, autoTrackPageViews = true, autoTrackErrors = true } = options ?? {};

  // Load consent from localStorage
  let consent: ConsentState = initialState.consent;
  if (typeof localStorage !== 'undefined') {
    const saved = localStorage.getItem('analytics_consent');
    if (saved) {
      try {
        consent = JSON.parse(saved);
      } catch {
        // Ignore parse errors
      }
    }
  }

  store.set({
    ...initialState,
    enabled: enabled && consent.categories.analytics,
    consent,
    debug,
  });

  // Start session
  startSession();

  // Auto-track page views
  if (autoTrackPageViews && typeof window !== 'undefined') {
    // Track initial page view
    trackPageView(window.location.pathname);

    // Track navigation (for SPA)
    const originalPushState = history.pushState;
    history.pushState = function (...args) {
      originalPushState.apply(this, args);
      trackPageView(window.location.pathname);
    };

    window.addEventListener('popstate', () => {
      trackPageView(window.location.pathname);
    });
  }

  // Auto-track errors
  if (autoTrackErrors && typeof window !== 'undefined') {
    window.addEventListener('error', (event) => {
      trackError(event.message, {
        stack: event.error?.stack,
        type: 'error',
        context: {
          filename: event.filename,
          lineno: event.lineno,
          colno: event.colno,
        },
      });
    });

    window.addEventListener('unhandledrejection', (event) => {
      trackError(event.reason?.message || 'Unhandled Promise Rejection', {
        stack: event.reason?.stack,
        type: 'unhandledrejection',
      });
    });
  }

  // Track session activity
  if (typeof window !== 'undefined') {
    ['mousedown', 'keydown', 'scroll', 'touchstart'].forEach((event) => {
      window.addEventListener(event, updateSessionActivity, { passive: true });
    });
  }
}

/**
 * Enable/disable analytics
 */
export function setEnabled(enabled: boolean): void {
  store.update((s) => ({ ...s, enabled }));
}

/**
 * Enable/disable debug mode
 */
export function setDebug(debug: boolean): void {
  store.update((s) => ({ ...s, debug }));
}

// ═══════════════════════════════════════════════════════════════════════════
// DERIVED STORES
// ═══════════════════════════════════════════════════════════════════════════

export const enabled = derived(store, ($store) => $store.enabled);
export const consent = derived(store, ($store) => $store.consent);
export const session = derived(store, ($store) => $store.session);
export const user = derived(store, ($store) => $store.user);
export const debug = derived(store, ($store) => $store.debug);

// ═══════════════════════════════════════════════════════════════════════════
// EXPORT STORE
// ═══════════════════════════════════════════════════════════════════════════

export const analyticsStore = {
  subscribe: store.subscribe,
  registerProvider,
  unregisterProvider,
  setConsent,
  hasConsent,
  startSession,
  endSession,
  trackPageView,
  trackEvent,
  identify,
  trackError,
  trackPerformance,
  initAnalytics,
  setEnabled,
  setDebug,
};
