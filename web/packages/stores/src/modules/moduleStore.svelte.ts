/**
 * Module Store (API-driven)
 *
 * Svelte 5 runes-based store that fetches modules and forms from the
 * FormService API. Falls back to the static MODULE_REGISTRY when the
 * API is unreachable.
 *
 * Usage:
 *   import { moduleStore } from '@chetana/stores/modules';
 *   moduleStore.loadModules();
 *   moduleStore.selectModule('finance');
 */

// Types live in moduleStore.types.ts. Svelte's compileModule rejects
// top-level `interface` declarations directly in this rune file (the
// dev-server optimizer would call compileModule WITHOUT vitePreprocess,
// causing "Unexpected token" errors at any `interface`/`: T`/`import type`
// it sees). The fix is two-pronged:
//   1. types extracted to the sibling .types.ts file (this import).
//   2. vite.config.ts excludes `@chetana/stores` from optimizeDeps so
//      SvelteKit's own pipeline (which DOES run vitePreprocess) handles
//      this rune module instead of the bare optimizer.
import type {
  ApiModuleSummary,
  ApiFormSummary,
  ModuleStoreState,
} from './moduleStore.types.js';

// ============================================================================
// FORM SERVICE RPC CLIENT (inline to avoid circular deps)
// ============================================================================

const FORM_SERVICE_BASE = '/platform.formservice.api.v1.FormService';

function getBaseUrl(): string {
  if (typeof import.meta !== 'undefined' && import.meta.env?.VITE_API_URL) {
    return import.meta.env.VITE_API_URL;
  }
  return 'http://localhost:8130';
}

/**
 * Read the current Bearer token from localStorage / sessionStorage where
 * authStore.login persists it. Inline rather than importing authStore to
 * keep this file free of the circular dep that prompted the original
 * "inline RPC client" comment above. Returns null when no token is
 * present (pre-login or post-logout) — the FormService call will then be
 * rejected by the backend with 401, surfaced as a Failed-to-fetch on the
 * `state.formError` channel without spinning the effect loop.
 */
function readAccessToken(): string | null {
  if (typeof window === 'undefined') return null;
  try {
    const raw = localStorage.getItem('auth_tokens') ?? sessionStorage.getItem('auth_tokens');
    if (!raw) return null;
    const parsed = JSON.parse(raw) as { accessToken?: string };
    return typeof parsed.accessToken === 'string' && parsed.accessToken.length > 0
      ? parsed.accessToken
      : null;
  } catch {
    return null;
  }
}

async function rpcCall<TReq, TRes>(method: string, request: TReq): Promise<TRes> {
  const url = `${getBaseUrl()}${FORM_SERVICE_BASE}/${method}`;
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  // Attach Bearer token under DEPLOY_AUTH_MODE=jwt. Backend's jwtAuthMiddleware
  // requires it on every non-skip-listed RPC; without it the response is 401
  // and the FormService call surfaces as "Failed to load forms".
  const tok = readAccessToken();
  if (tok) headers['Authorization'] = `Bearer ${tok}`;

  const response = await fetch(url, {
    method: 'POST',
    headers,
    // 'omit' for Bearer-token auth — the same CORS wildcard-origin trap
    // documented in packages/api/src/client/transport.ts applies here.
    // Backend CORS uses Access-Control-Allow-Origin: '*'; the browser
    // rejects credentialed requests against wildcard origins as
    // "Failed to fetch" with no HTTP status reaching JS.
    credentials: 'omit',
    body: JSON.stringify(request),
  });
  if (!response.ok) {
    throw new Error(`FormService.${method}: ${response.status} ${response.statusText}`);
  }
  return response.json() as Promise<TRes>;
}

// ============================================================================
// STORE IMPLEMENTATION
// ============================================================================

function createModuleStore() {
  let state = $state<ModuleStoreState>({
    modules: [],
    selectedModuleId: null,
    forms: [],
    isLoadingModules: false,
    isLoadingForms: false,
    moduleError: null,
    formError: null,
    isApiDriven: false,
  });

  // Same loop-protection as selectModule: callers in $effect blocks
  // would otherwise re-fire on every state mutation inside loadModules.
  let modulesAttempted = false;
  let modulesInFlight = false;

  /**
   * Load modules from FormService API.
   * Falls back to empty list on failure (the UI should use static registry as fallback).
   * Idempotent — only the first call hits the network; refresh() resets.
   */
  async function loadModules(): Promise<void> {
    if (modulesInFlight) return;
    if (modulesAttempted) return;
    modulesInFlight = true;
    state.isLoadingModules = true;
    state.moduleError = null;

    try {
      const response = await rpcCall<
        { context: Record<string, unknown> },
        { modules?: ApiModuleSummary[]; totalModules?: number }
      >('ListModules', { context: {} });

      state.modules = response.modules ?? [];
      state.isApiDriven = state.modules.length > 0;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load modules';
      state.moduleError = message;
      state.isApiDriven = false;
      console.warn('[moduleStore] API unavailable, UI should fall back to static registry:', message);
    } finally {
      state.isLoadingModules = false;
      modulesInFlight = false;
      modulesAttempted = true;
    }
  }

  /**
   * In-flight request tracker. Without this guard, an `$effect` that
   * calls selectModule will re-fire on every state mutation (forms,
   * formError, isLoadingForms all reactive) — selectModule mutates state
   * INSIDE itself, so the effect re-runs mid-call → call again → loop.
   * The outer "selectedModuleId === moduleId && forms.length > 0" guard
   * only catches success cases; failure leaves forms.length === 0 so the
   * loop continues. Failed-module set short-circuits subsequent calls
   * until refresh() is called.
   */
  const inFlight = new Set<string>();
  const failedModules = new Set<string>();

  /**
   * Select a module and load its forms from the API.
   */
  async function selectModule(moduleId: string): Promise<void> {
    if (state.selectedModuleId === moduleId && state.forms.length > 0) {
      return; // Already loaded
    }
    if (inFlight.has(moduleId)) {
      return; // Coalesce concurrent calls for the same module
    }
    if (failedModules.has(moduleId)) {
      // Permanently failed in this session; refresh() clears the set.
      // Avoids the $effect → mutation → $effect re-fire infinite loop
      // when the backend is unreachable or returns non-2xx.
      return;
    }

    state.selectedModuleId = moduleId;
    state.forms = [];
    state.isLoadingForms = true;
    state.formError = null;
    inFlight.add(moduleId);

    try {
      const response = await rpcCall<
        { context: Record<string, unknown>; moduleId: string },
        { forms?: ApiFormSummary[]; totalForms?: number }
      >('ListForms', { context: {}, moduleId });

      state.forms = response.forms ?? [];
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load forms';
      state.formError = message;
      failedModules.add(moduleId);
      console.warn(`[moduleStore] Failed to load forms for ${moduleId}:`, message);
    } finally {
      state.isLoadingForms = false;
      inFlight.delete(moduleId);
    }
  }

  /**
   * Clear the selected module and its forms.
   */
  function clearSelection(): void {
    state.selectedModuleId = null;
    state.forms = [];
    state.formError = null;
  }

  /**
   * Force refresh modules from the API. Clears the per-module failure
   * set so a transient failure can be retried after a network/backend
   * blip without requiring a full page reload.
   */
  async function refresh(): Promise<void> {
    failedModules.clear();
    modulesAttempted = false;
    await loadModules();
    if (state.selectedModuleId) {
      const current = state.selectedModuleId;
      state.selectedModuleId = null; // Force reload
      await selectModule(current);
    }
  }

  return {
    /** Read the full state (reactive via $state) */
    get state() { return state; },

    /** Reactive getters for individual state properties */
    get modules() { return state.modules; },
    get selectedModuleId() { return state.selectedModuleId; },
    get forms() { return state.forms; },
    get isLoadingModules() { return state.isLoadingModules; },
    get isLoadingForms() { return state.isLoadingForms; },
    get moduleError() { return state.moduleError; },
    get formError() { return state.formError; },
    get isApiDriven() { return state.isApiDriven; },
    get isLoading() { return state.isLoadingModules || state.isLoadingForms; },

    /** Actions */
    loadModules,
    selectModule,
    clearSelection,
    refresh,
  };
}

export const moduleStore = createModuleStore();
