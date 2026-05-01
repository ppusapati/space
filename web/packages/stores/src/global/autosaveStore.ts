/**
 * Autosave Store
 *
 * Automatic saving functionality for forms and editors:
 * - Debounced saves to prevent excessive API calls
 * - Local storage backup for crash recovery
 * - Save status indicators
 * - Conflict detection
 * - Retry on failure
 */

import { writable, derived, get } from 'svelte/store';

// ============================================================================
// Types
// ============================================================================

export type AutosaveStatus = 'idle' | 'pending' | 'saving' | 'saved' | 'error' | 'conflict';

export interface AutosaveConfig<T> {
  /** Unique key for local storage backup */
  key: string;
  /** Debounce delay in ms (default: 2000) */
  debounceMs: number;
  /** Maximum retries on failure (default: 3) */
  maxRetries: number;
  /** Retry delay in ms (default: 1000) */
  retryDelayMs: number;
  /** Save function */
  onSave: (data: T) => Promise<T | void>;
  /** Optional conflict check (returns server version if conflict) */
  onConflictCheck?: (data: T) => Promise<T | null>;
  /** Optional callback on save success */
  onSuccess?: (data: T) => void;
  /** Optional callback on save error */
  onError?: (error: Error) => void;
  /** Enable local storage backup (default: true) */
  enableLocalBackup: boolean;
  /** Version field for conflict detection */
  versionField?: keyof T;
}

export interface AutosaveState<T> {
  data: T;
  status: AutosaveStatus;
  lastSavedAt: Date | null;
  lastSavedData: T | null;
  error: string | null;
  retryCount: number;
  isDirty: boolean;
  hasLocalBackup: boolean;
  serverVersion: T | null;
}

// ============================================================================
// Factory
// ============================================================================

const DEFAULT_CONFIG = {
  debounceMs: 2000,
  maxRetries: 3,
  retryDelayMs: 1000,
  enableLocalBackup: true,
};

/**
 * Create an autosave-enabled store
 */
export function createAutosaveStore<T extends Record<string, unknown>>(
  initialData: T,
  config: Omit<AutosaveConfig<T>, keyof typeof DEFAULT_CONFIG> & Partial<typeof DEFAULT_CONFIG>
) {
  const cfg: AutosaveConfig<T> = { ...DEFAULT_CONFIG, ...config };
  const storageKey = `autosave_${cfg.key}`;

  const { subscribe, set, update } = writable<AutosaveState<T>>({
    data: initialData,
    status: 'idle',
    lastSavedAt: null,
    lastSavedData: null,
    error: null,
    retryCount: 0,
    isDirty: false,
    hasLocalBackup: false,
    serverVersion: null,
  });

  let debounceTimer: ReturnType<typeof setTimeout> | null = null;
  let retryTimer: ReturnType<typeof setTimeout> | null = null;

  // Check for local backup on initialization
  function checkLocalBackup(): T | null {
    if (!cfg.enableLocalBackup || typeof localStorage === 'undefined') return null;
    try {
      const stored = localStorage.getItem(storageKey);
      if (stored) {
        const parsed = JSON.parse(stored);
        update(s => ({ ...s, hasLocalBackup: true }));
        return parsed.data as T;
      }
    } catch { /* ignore */ }
    return null;
  }

  function saveToLocalStorage(data: T) {
    if (!cfg.enableLocalBackup || typeof localStorage === 'undefined') return;
    try {
      localStorage.setItem(storageKey, JSON.stringify({ data, timestamp: Date.now() }));
      update(s => ({ ...s, hasLocalBackup: true }));
    } catch { /* ignore storage errors */ }
  }

  function clearLocalBackup() {
    if (typeof localStorage === 'undefined') return;
    try {
      localStorage.removeItem(storageKey);
      update(s => ({ ...s, hasLocalBackup: false }));
    } catch { /* ignore */ }
  }

  async function performSave(data: T, isRetry = false): Promise<boolean> {
    const state = get({ subscribe });
    if (state.status === 'saving' && !isRetry) return false;

    update(s => ({ ...s, status: 'saving', error: null }));

    try {
      // Check for conflicts if configured
      if (cfg.onConflictCheck) {
        const serverData = await cfg.onConflictCheck(data);
        if (serverData !== null) {
          update(s => ({ ...s, status: 'conflict', serverVersion: serverData }));
          return false;
        }
      }

      const result = await cfg.onSave(data);
      const savedData = (result || data) as T;

      update(s => ({
        ...s,
        status: 'saved',
        lastSavedAt: new Date(),
        lastSavedData: savedData,
        data: savedData,
        isDirty: false,
        retryCount: 0,
        serverVersion: null,
      }));

      clearLocalBackup();
      cfg.onSuccess?.(savedData);
      return true;
    } catch (error) {
      const err = error as Error;
      const state = get({ subscribe });

      if (state.retryCount < cfg.maxRetries) {
        update(s => ({ ...s, retryCount: s.retryCount + 1, status: 'pending' }));
        retryTimer = setTimeout(() => performSave(data, true), cfg.retryDelayMs);
        return false;
      }

      update(s => ({ ...s, status: 'error', error: err.message }));
      cfg.onError?.(err);
      return false;
    }
  }

  function scheduleSave() {
    if (debounceTimer) clearTimeout(debounceTimer);
    update(s => ({ ...s, status: 'pending' }));

    debounceTimer = setTimeout(() => {
      const state = get({ subscribe });
      performSave(state.data);
    }, cfg.debounceMs);
  }

  return {
    subscribe,

    /** Initialize and check for local backup */
    init(): T | null {
      return checkLocalBackup();
    },

    /** Update data (triggers autosave) */
    update(newData: Partial<T>) {
      update(s => {
        const data = { ...s.data, ...newData };
        saveToLocalStorage(data);
        return { ...s, data, isDirty: true };
      });
      scheduleSave();
    },

    /** Set entire data object */
    setData(data: T) {
      update(s => ({ ...s, data, isDirty: true }));
      saveToLocalStorage(data);
      scheduleSave();
    },

    /** Force immediate save */
    async saveNow(): Promise<boolean> {
      if (debounceTimer) clearTimeout(debounceTimer);
      if (retryTimer) clearTimeout(retryTimer);
      const state = get({ subscribe });
      return performSave(state.data);
    },

    /** Restore from local backup */
    restoreBackup(): T | null {
      const backup = checkLocalBackup();
      if (backup) {
        update(s => ({ ...s, data: backup, isDirty: true }));
        return backup;
      }
      return null;
    },

    /** Discard local backup */
    discardBackup() {
      clearLocalBackup();
    },

    /** Resolve conflict by keeping local version */
    resolveConflictWithLocal() {
      update(s => ({ ...s, status: 'pending', serverVersion: null }));
      scheduleSave();
    },

    /** Resolve conflict by keeping server version */
    resolveConflictWithServer() {
      const state = get({ subscribe });
      if (state.serverVersion) {
        update(s => ({ ...s, data: s.serverVersion!, status: 'idle', isDirty: false, serverVersion: null }));
        clearLocalBackup();
      }
    },

    /** Reset to initial state */
    reset() {
      if (debounceTimer) clearTimeout(debounceTimer);
      if (retryTimer) clearTimeout(retryTimer);
      clearLocalBackup();
      set({
        data: initialData,
        status: 'idle',
        lastSavedAt: null,
        lastSavedData: null,
        error: null,
        retryCount: 0,
        isDirty: false,
        hasLocalBackup: false,
        serverVersion: null,
      });
    },

    /** Destroy and cleanup */
    destroy() {
      if (debounceTimer) clearTimeout(debounceTimer);
      if (retryTimer) clearTimeout(retryTimer);
    },
  };
}

// ============================================================================
// Derived Helpers
// ============================================================================

export function deriveAutosaveState<T extends Record<string, unknown>>(
  store: ReturnType<typeof createAutosaveStore<T>>
) {
  return {
    data: derived(store, $s => $s.data),
    status: derived(store, $s => $s.status),
    isDirty: derived(store, $s => $s.isDirty),
    isSaving: derived(store, $s => $s.status === 'saving'),
    hasError: derived(store, $s => $s.status === 'error'),
    hasConflict: derived(store, $s => $s.status === 'conflict'),
    lastSavedAt: derived(store, $s => $s.lastSavedAt),
  };
}

// ============================================================================
// useAutosave Composable (Svelte 5)
// ============================================================================

/**
 * Svelte 5 composable for autosave functionality
 * Usage:
 * ```svelte
 * <script>
 *   const form = useAutosave(
 *     { name: '', email: '' },
 *     {
 *       key: 'user-form',
 *       onSave: async (data) => await api.saveUser(data),
 *     }
 *   );
 *
 *   // Restore backup if exists
 *   onMount(() => form.init());
 * </script>
 *
 * <input bind:value={$form.data.name} oninput={() => form.update({ name: $form.data.name })} />
 * {#if $form.status === 'saving'}Saving...{/if}
 * {#if $form.status === 'saved'}Saved{/if}
 * ```
 */
export function useAutosave<T extends Record<string, unknown>>(
  initialData: T,
  config: Omit<AutosaveConfig<T>, keyof typeof DEFAULT_CONFIG> & Partial<typeof DEFAULT_CONFIG>
) {
  return createAutosaveStore(initialData, config);
}
