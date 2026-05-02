/**
 * @chetana/stores
 * State management stores for Chetana ERP
 * @packageDocumentation
 */

// ============================================================================
// GLOBAL STORES
// ============================================================================
export * from './global/index.js';

// ============================================================================
// API PROVIDER BRIDGE (wires stores → @chetana/api providers)
// ============================================================================
// Call initApiProviders() ONCE from apps/shell/src/hooks.client.ts BEFORE
// initializeApi(). The interceptors registered by initializeApi start
// calling getAuthProvider()/getSessionProvider() on the first request.
export { initApiProviders } from './apiProviderBridge.js';

// ============================================================================
// MODULE STORE (API-driven module/form discovery)
// ============================================================================
export * from './modules/index.js';

// ============================================================================
// STORE FACTORIES
// ============================================================================
export * from './factories/index.js';

// Theme stores are already exported from ./global/index.js
// The ./theme/ module contains advanced theme configuration which can be imported directly
// import { ... } from '@chetana/stores/theme'
