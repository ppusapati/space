/**
 * Audit Logging Store
 *
 * Comprehensive audit trail for compliance and tracking:
 * - Track user actions (CRUD, auth, workflow)
 * - Record data changes with before/after values
 * - Batch logging for performance
 * - Backend integration
 */

import { writable, derived, get } from 'svelte/store';

// ============================================================================
// Types
// ============================================================================

export type AuditAction =
  | 'create' | 'read' | 'update' | 'delete'
  | 'login' | 'logout' | 'export' | 'import'
  | 'print' | 'approve' | 'reject' | 'submit'
  | 'cancel' | 'archive' | 'restore' | 'custom';

export type AuditLevel = 'info' | 'warning' | 'critical';
export type AuditCategory = 'authentication' | 'authorization' | 'data' | 'configuration' | 'system' | 'security' | 'workflow';

export interface AuditUser {
  id: string;
  email: string;
  name: string;
  role?: string;
}

export interface AuditContext {
  tenantId?: string;
  companyId?: string;
  branchId?: string;
  module?: string;
  ipAddress?: string;
  userAgent?: string;
  sessionId?: string;
}

export interface DataChange {
  field: string;
  oldValue: unknown;
  newValue: unknown;
  displayName?: string;
}

export interface AuditEntry {
  id: string;
  timestamp: Date;
  action: AuditAction;
  category: AuditCategory;
  level: AuditLevel;
  user: AuditUser;
  context: AuditContext;
  entityType: string;
  entityId: string;
  entityName?: string;
  description: string;
  changes?: DataChange[];
  metadata?: Record<string, unknown>;
  success: boolean;
  errorMessage?: string;
}

export interface AuditConfig {
  enabled: boolean;
  minLevel: AuditLevel;
  categories: AuditCategory[];
  batchSize: number;
  flushInterval: number;
  maxEntries: number;
  endpoint?: string;
  logReads: boolean;
}

export interface AuditState {
  entries: AuditEntry[];
  pendingEntries: AuditEntry[];
  config: AuditConfig;
  isFlushing: boolean;
  lastFlushAt: Date | null;
}

// ============================================================================
// Store Implementation
// ============================================================================

const DEFAULT_CONFIG: AuditConfig = {
  enabled: true,
  minLevel: 'info',
  categories: ['authentication', 'data', 'security', 'workflow'],
  batchSize: 50,
  flushInterval: 30000,
  maxEntries: 1000,
  logReads: false,
};

const LEVEL_PRIORITY: Record<AuditLevel, number> = { info: 0, warning: 1, critical: 2 };

function generateId(): string {
  return `audit_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

function computeChanges<T extends Record<string, unknown>>(
  oldData: T | null,
  newData: T | null,
  fieldLabels?: Record<string, string>
): DataChange[] {
  const changes: DataChange[] = [];
  if (!oldData && !newData) return changes;

  const allKeys = new Set([...Object.keys(oldData || {}), ...Object.keys(newData || {})]);

  for (const key of allKeys) {
    const oldValue = oldData?.[key];
    const newValue = newData?.[key];
    if (JSON.stringify(oldValue) !== JSON.stringify(newValue)) {
      changes.push({ field: key, oldValue, newValue, displayName: fieldLabels?.[key] || key });
    }
  }
  return changes;
}

function createAuditStore() {
  const { subscribe, set, update } = writable<AuditState>({
    entries: [],
    pendingEntries: [],
    config: DEFAULT_CONFIG,
    isFlushing: false,
    lastFlushAt: null,
  });

  let flushTimer: ReturnType<typeof setInterval> | null = null;
  let currentUser: AuditUser | null = null;
  let currentContext: AuditContext = {};

  async function sendToBackend(entries: AuditEntry[]): Promise<boolean> {
    const state = get({ subscribe });
    if (!state.config.endpoint) return true;

    try {
      const response = await fetch(state.config.endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ entries }),
      });
      return response.ok;
    } catch {
      return false;
    }
  }

  return {
    subscribe,

    init(user: AuditUser, context?: AuditContext, config?: Partial<AuditConfig>) {
      currentUser = user;
      currentContext = context || {};
      update((s) => ({ ...s, config: { ...s.config, ...config } }));

      if (flushTimer) clearInterval(flushTimer);
      const state = get({ subscribe });
      flushTimer = setInterval(() => this.flush(), state.config.flushInterval);
    },

    setUser(user: AuditUser | null) { currentUser = user; },
    setContext(context: Partial<AuditContext>) { currentContext = { ...currentContext, ...context }; },

    log(params: {
      action: AuditAction;
      category: AuditCategory;
      level?: AuditLevel;
      entityType: string;
      entityId: string;
      entityName?: string;
      description: string;
      changes?: DataChange[];
      metadata?: Record<string, unknown>;
      success?: boolean;
      errorMessage?: string;
    }): string | null {
      const state = get({ subscribe });
      const level = params.level || 'info';

      if (!state.config.enabled) return null;
      if (!state.config.logReads && params.action === 'read') return null;
      if (!state.config.categories.includes(params.category)) return null;
      if (LEVEL_PRIORITY[level] < LEVEL_PRIORITY[state.config.minLevel]) return null;
      if (!currentUser) return null;

      const entry: AuditEntry = {
        id: generateId(),
        timestamp: new Date(),
        action: params.action,
        category: params.category,
        level,
        user: currentUser,
        context: currentContext,
        entityType: params.entityType,
        entityId: params.entityId,
        entityName: params.entityName,
        description: params.description,
        changes: params.changes,
        metadata: params.metadata,
        success: params.success !== false,
        errorMessage: params.errorMessage,
      };

      update((s) => {
        const pendingEntries = [...s.pendingEntries, entry];
        let entries = [...s.entries, entry];
        if (entries.length > s.config.maxEntries) entries = entries.slice(-s.config.maxEntries);
        if (pendingEntries.length >= s.config.batchSize) setTimeout(() => this.flush(), 0);
        return { ...s, entries, pendingEntries };
      });

      return entry.id;
    },

    logCreate<T extends Record<string, unknown>>(entityType: string, entityId: string, data: T, entityName?: string, fieldLabels?: Record<string, string>) {
      return this.log({ action: 'create', category: 'data', entityType, entityId, entityName, description: `Created ${entityType} ${entityName || entityId}`, changes: computeChanges(null, data, fieldLabels) });
    },

    logUpdate<T extends Record<string, unknown>>(entityType: string, entityId: string, oldData: T, newData: T, entityName?: string, fieldLabels?: Record<string, string>) {
      const changes = computeChanges(oldData, newData, fieldLabels);
      if (changes.length === 0) return null;
      return this.log({ action: 'update', category: 'data', entityType, entityId, entityName, description: `Updated ${entityType} ${entityName || entityId}`, changes });
    },

    logDelete(entityType: string, entityId: string, entityName?: string) {
      return this.log({ action: 'delete', category: 'data', level: 'warning', entityType, entityId, entityName, description: `Deleted ${entityType} ${entityName || entityId}` });
    },

    logAuth(action: 'login' | 'logout', success: boolean, errorMessage?: string) {
      return this.log({ action, category: 'authentication', level: success ? 'info' : 'warning', entityType: 'session', entityId: currentContext.sessionId || 'unknown', description: action === 'login' ? (success ? 'User logged in' : 'Login failed') : 'User logged out', success, errorMessage });
    },

    async flush(): Promise<void> {
      const state = get({ subscribe });
      if (state.pendingEntries.length === 0 || state.isFlushing) return;

      update((s) => ({ ...s, isFlushing: true }));
      const success = await sendToBackend([...state.pendingEntries]);
      update((s) => ({ ...s, isFlushing: false, pendingEntries: success ? [] : s.pendingEntries, lastFlushAt: success ? new Date() : s.lastFlushAt }));
    },

    getEntries(filter?: { startDate?: Date; endDate?: Date; actions?: AuditAction[]; entityType?: string; userId?: string }): AuditEntry[] {
      let entries = [...get({ subscribe }).entries];
      if (filter?.startDate) entries = entries.filter((e) => e.timestamp >= filter.startDate!);
      if (filter?.endDate) entries = entries.filter((e) => e.timestamp <= filter.endDate!);
      if (filter?.actions?.length) entries = entries.filter((e) => filter.actions!.includes(e.action));
      if (filter?.entityType) entries = entries.filter((e) => e.entityType === filter.entityType);
      if (filter?.userId) entries = entries.filter((e) => e.user.id === filter.userId);
      return entries;
    },

    clear() { update((s) => ({ ...s, entries: [], pendingEntries: [] })); },

    destroy() {
      if (flushTimer) clearInterval(flushTimer);
      this.flush();
    },
  };
}

export const auditStore = createAuditStore();

export const recentAuditEntries = derived(auditStore, ($s) => $s.entries.slice(-100).reverse());
export const pendingAuditCount = derived(auditStore, ($s) => $s.pendingEntries.length);

export function initAudit(user: AuditUser, context?: AuditContext, config?: Partial<AuditConfig>) {
  auditStore.init(user, context, config);
}

export function audit(params: Parameters<typeof auditStore.log>[0]) {
  return auditStore.log(params);
}

export { computeChanges };
