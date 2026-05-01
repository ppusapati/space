/**
 * Feature Flag Store - Gradual Feature Rollouts
 *
 * Features:
 * - Boolean and variant-based flags
 * - User/tenant targeting
 * - Percentage-based rollouts
 * - A/B testing support
 * - Remote flag updates
 * - Local overrides for development
 */

import { writable, derived, get } from 'svelte/store';

// ═══════════════════════════════════════════════════════════════════════════
// TYPES
// ═══════════════════════════════════════════════════════════════════════════

export type FlagValue = boolean | string | number | object;

export interface FeatureFlag {
  /** Unique flag identifier */
  key: string;
  /** Flag description */
  description?: string;
  /** Default value when no rules match */
  defaultValue: FlagValue;
  /** Whether the flag is enabled */
  enabled: boolean;
  /** Targeting rules */
  rules?: FlagRule[];
  /** Percentage rollout (0-100) */
  percentage?: number;
  /** Variants for A/B testing */
  variants?: FlagVariant[];
  /** Metadata */
  metadata?: Record<string, unknown>;
  /** Created timestamp */
  createdAt?: string;
  /** Updated timestamp */
  updatedAt?: string;
}

export type TargetingOperator = FlagRule['operator'];
export type TargetingRule = FlagRule;

export interface FlagRule {
  /** Rule identifier */
  id: string;
  /** Condition type */
  type: 'user' | 'tenant' | 'role' | 'attribute' | 'percentage';
  /** Attribute to check */
  attribute?: string;
  /** Operator for comparison */
  operator: 'eq' | 'neq' | 'in' | 'nin' | 'gt' | 'gte' | 'lt' | 'lte' | 'contains' | 'startsWith' | 'endsWith';
  /** Value to compare against */
  value: unknown;
  /** Value to return if rule matches */
  returnValue: FlagValue;
}

export type FeatureVariant = FlagVariant;

export interface FlagVariant {
  /** Variant identifier */
  key: string;
  /** Variant value */
  value: FlagValue;
  /** Percentage weight (all variants should sum to 100) */
  weight: number;
}

export interface FlagContext {
  /** User identifier */
  userId?: string;
  /** Tenant identifier */
  tenantId?: string;
  /** User roles */
  roles?: string[];
  /** User email */
  email?: string;
  /** Custom attributes */
  attributes?: Record<string, unknown>;
}

export interface FeatureFlagState {
  /** All feature flags */
  flags: Record<string, FeatureFlag>;
  /** Current evaluation context */
  context: FlagContext;
  /** Local overrides (for development) */
  overrides: Record<string, FlagValue>;
  /** Whether flags are loading */
  loading: boolean;
  /** Last fetch timestamp */
  lastFetched?: number;
  /** Error message */
  error?: string;
}

// ═══════════════════════════════════════════════════════════════════════════
// STORE
// ═══════════════════════════════════════════════════════════════════════════

const initialState: FeatureFlagState = {
  flags: {},
  context: {},
  overrides: {},
  loading: false,
};

const store = writable<FeatureFlagState>(initialState);

// ═══════════════════════════════════════════════════════════════════════════
// EVALUATION FUNCTIONS
// ═══════════════════════════════════════════════════════════════════════════

/**
 * Hash function for consistent percentage bucketing
 */
function hashString(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash = hash & hash; // Convert to 32-bit integer
  }
  return Math.abs(hash);
}

/**
 * Get percentage bucket for a user (0-100)
 */
function getPercentageBucket(key: string, userId: string): number {
  const hash = hashString(`${key}:${userId}`);
  return hash % 100;
}

/**
 * Evaluate a single rule against context
 */
function evaluateRule(rule: FlagRule, context: FlagContext): boolean {
  let targetValue: unknown;

  switch (rule.type) {
    case 'user':
      targetValue = context.userId;
      break;
    case 'tenant':
      targetValue = context.tenantId;
      break;
    case 'role':
      targetValue = context.roles;
      break;
    case 'attribute':
      targetValue = context.attributes?.[rule.attribute || ''];
      break;
    case 'percentage':
      if (!context.userId) return false;
      const bucket = getPercentageBucket(rule.attribute || 'default', context.userId);
      return bucket < (rule.value as number);
    default:
      return false;
  }

  // Handle array target value (e.g., roles)
  if (Array.isArray(targetValue)) {
    switch (rule.operator) {
      case 'in':
        return (rule.value as unknown[]).some((v) => targetValue.includes(v));
      case 'nin':
        return !(rule.value as unknown[]).some((v) => targetValue.includes(v));
      case 'contains':
        return targetValue.includes(rule.value);
      default:
        return false;
    }
  }

  // Handle scalar target value
  switch (rule.operator) {
    case 'eq':
      return targetValue === rule.value;
    case 'neq':
      return targetValue !== rule.value;
    case 'in':
      return (rule.value as unknown[]).includes(targetValue);
    case 'nin':
      return !(rule.value as unknown[]).includes(targetValue);
    case 'gt':
      return (targetValue as number) > (rule.value as number);
    case 'gte':
      return (targetValue as number) >= (rule.value as number);
    case 'lt':
      return (targetValue as number) < (rule.value as number);
    case 'lte':
      return (targetValue as number) <= (rule.value as number);
    case 'contains':
      return String(targetValue).includes(String(rule.value));
    case 'startsWith':
      return String(targetValue).startsWith(String(rule.value));
    case 'endsWith':
      return String(targetValue).endsWith(String(rule.value));
    default:
      return false;
  }
}

/**
 * Select variant based on user bucket
 */
function selectVariant(variants: FlagVariant[], key: string, userId: string): FlagValue {
  const bucket = getPercentageBucket(key, userId);
  let accumulated = 0;

  for (const variant of variants) {
    accumulated += variant.weight;
    if (bucket < accumulated) {
      return variant.value;
    }
  }

  // Fallback to last variant
  return variants[variants.length - 1]?.value ?? false;
}

/**
 * Evaluate a feature flag
 */
function evaluateFlag(flag: FeatureFlag, context: FlagContext): FlagValue {
  // Flag is disabled
  if (!flag.enabled) {
    return flag.defaultValue;
  }

  // Check percentage rollout
  if (flag.percentage !== undefined && context.userId) {
    const bucket = getPercentageBucket(flag.key, context.userId);
    if (bucket >= flag.percentage) {
      return flag.defaultValue;
    }
  }

  // Evaluate rules in order
  if (flag.rules && flag.rules.length > 0) {
    for (const rule of flag.rules) {
      if (evaluateRule(rule, context)) {
        return rule.returnValue;
      }
    }
  }

  // Handle variants (A/B testing)
  if (flag.variants && flag.variants.length > 0 && context.userId) {
    return selectVariant(flag.variants, flag.key, context.userId);
  }

  // Return default for boolean flags
  return flag.enabled;
}

// ═══════════════════════════════════════════════════════════════════════════
// PUBLIC API
// ═══════════════════════════════════════════════════════════════════════════

/**
 * Check if a feature is enabled
 */
export function isEnabled(key: string): boolean {
  const state = get(store);

  // Check local overrides first
  if (key in state.overrides) {
    return Boolean(state.overrides[key]);
  }

  const flag = state.flags[key];
  if (!flag) {
    return false;
  }

  const value = evaluateFlag(flag, state.context);
  return Boolean(value);
}

/**
 * Get feature flag value (for non-boolean flags)
 */
export function getFlagValue<T extends FlagValue = FlagValue>(key: string, defaultValue?: T): T {
  const state = get(store);

  // Check local overrides first
  if (key in state.overrides) {
    return state.overrides[key] as T;
  }

  const flag = state.flags[key];
  if (!flag) {
    return (defaultValue ?? false) as T;
  }

  return evaluateFlag(flag, state.context) as T;
}

/**
 * Reactive flag check
 */
export function flag(key: string) {
  return derived(store, ($store) => {
    // Check local overrides first
    if (key in $store.overrides) {
      return $store.overrides[key];
    }

    const flag = $store.flags[key];
    if (!flag) {
      return false;
    }

    return evaluateFlag(flag, $store.context);
  });
}

/**
 * Set evaluation context
 */
export function setContext(context: FlagContext): void {
  store.update((s) => ({ ...s, context }));
}

/**
 * Update evaluation context
 */
export function updateContext(context: Partial<FlagContext>): void {
  store.update((s) => ({
    ...s,
    context: { ...s.context, ...context },
  }));
}

/**
 * Set local override (for development/testing)
 */
export function setOverride(key: string, value: FlagValue): void {
  store.update((s) => ({
    ...s,
    overrides: { ...s.overrides, [key]: value },
  }));

  // Persist to localStorage
  if (typeof localStorage !== 'undefined') {
    const state = get(store);
    localStorage.setItem('feature_flag_overrides', JSON.stringify(state.overrides));
  }
}

/**
 * Clear local override
 */
export function clearOverride(key: string): void {
  store.update((s) => {
    const { [key]: _, ...rest } = s.overrides;
    return { ...s, overrides: rest };
  });

  // Persist to localStorage
  if (typeof localStorage !== 'undefined') {
    const state = get(store);
    localStorage.setItem('feature_flag_overrides', JSON.stringify(state.overrides));
  }
}

/**
 * Clear all local overrides
 */
export function clearAllOverrides(): void {
  store.update((s) => ({ ...s, overrides: {} }));

  if (typeof localStorage !== 'undefined') {
    localStorage.removeItem('feature_flag_overrides');
  }
}

/**
 * Set flags directly
 */
export function setFlags(flags: Record<string, FeatureFlag>): void {
  store.update((s) => ({
    ...s,
    flags,
    lastFetched: Date.now(),
    error: undefined,
  }));
}

/**
 * Update a single flag
 */
export function updateFlag(key: string, flag: FeatureFlag): void {
  store.update((s) => ({
    ...s,
    flags: { ...s.flags, [key]: flag },
  }));
}

/**
 * Fetch flags from remote source
 */
export async function fetchFlags(
  fetcher: () => Promise<Record<string, FeatureFlag>>
): Promise<void> {
  store.update((s) => ({ ...s, loading: true, error: undefined }));

  try {
    const flags = await fetcher();
    store.update((s) => ({
      ...s,
      flags,
      loading: false,
      lastFetched: Date.now(),
    }));
  } catch (error) {
    store.update((s) => ({
      ...s,
      loading: false,
      error: error instanceof Error ? error.message : 'Failed to fetch flags',
    }));
  }
}

/**
 * Initialize feature flags
 */
export function initFeatureFlags(options: {
  flags?: Record<string, FeatureFlag>;
  context?: FlagContext;
}): void {
  const { flags = {}, context = {} } = options;

  // Load overrides from localStorage
  let overrides: Record<string, FlagValue> = {};
  if (typeof localStorage !== 'undefined') {
    const saved = localStorage.getItem('feature_flag_overrides');
    if (saved) {
      try {
        overrides = JSON.parse(saved);
      } catch {
        // Ignore parse errors
      }
    }
  }

  store.set({
    flags,
    context,
    overrides,
    loading: false,
    lastFetched: Date.now(),
  });
}

/**
 * Get all flags
 */
export function getAllFlags(): Record<string, FeatureFlag> {
  return get(store).flags;
}

/**
 * Get all overrides
 */
export function getOverrides(): Record<string, FlagValue> {
  return get(store).overrides;
}

// ═══════════════════════════════════════════════════════════════════════════
// DERIVED STORES
// ═══════════════════════════════════════════════════════════════════════════

export const loading = derived(store, ($store) => $store.loading);
export const error = derived(store, ($store) => $store.error);
export const context = derived(store, ($store) => $store.context);
export const hasOverrides = derived(store, ($store) => Object.keys($store.overrides).length > 0);

// ═══════════════════════════════════════════════════════════════════════════
// EXPORT STORE
// ═══════════════════════════════════════════════════════════════════════════

/** @deprecated Use clearOverride */
export const removeOverride = clearOverride;
/** @deprecated Use clearAllOverrides */
export const clearOverrides = clearAllOverrides;
/** @deprecated Use setFlags */
export const registerFlags = setFlags;

export const featureFlagStore = {
  subscribe: store.subscribe,
  isEnabled,
  getFlagValue,
  flag,
  setContext,
  updateContext,
  setOverride,
  clearOverride,
  clearAllOverrides,
  setFlags,
  updateFlag,
  fetchFlags,
  initFeatureFlags,
  getAllFlags,
  getOverrides,
};
