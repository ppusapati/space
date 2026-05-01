/**
 * API Providers
 * Defines interfaces for external dependencies (stores) that the API package needs.
 * Consumers (e.g., @samavāya/stores) inject concrete implementations via configure().
 * This breaks the cyclic dependency: api ↔ stores.
 * @packageDocumentation
 */

// ============================================================================
// AUTH PROVIDER
// ============================================================================

export interface AuthTokens {
  accessToken: string;
  refreshToken?: string;
  expiresAt?: Date;
}

export interface AuthProvider {
  /** Returns true if user is authenticated */
  isAuthenticated(): boolean;

  /** Returns current tokens, or null if not authenticated */
  getTokens(): AuthTokens | null;

  /** Refreshes the access token */
  refreshTokens(): Promise<void>;

  /** Logs the user out */
  logout(): void;
}

// ============================================================================
// SESSION PROVIDER
// ============================================================================

export interface SessionContext {
  tenantId: string;
  companyId: string;
  branchId: string;
  userId: string;
}

export interface SessionProvider {
  /** Returns the current session ID, or null */
  getSessionId(): string | null;

  /** Returns the current organizational context, or null */
  getContext(): SessionContext | null;
}

// ============================================================================
// TOAST PROVIDER
// ============================================================================

export interface ToastProvider {
  /** Shows an error toast */
  error(message: string, options?: { title?: string; duration?: number }): void;
}

// ============================================================================
// PROVIDER REGISTRY
// ============================================================================

let authProvider: AuthProvider | null = null;
let sessionProvider: SessionProvider | null = null;
let toastProvider: ToastProvider | null = null;

/**
 * Configures the API providers. Must be called before using interceptors.
 * Typically called once during app initialization from @samavāya/stores.
 */
export function configureProviders(providers: {
  auth?: AuthProvider;
  session?: SessionProvider;
  toast?: ToastProvider;
}): void {
  if (providers.auth) authProvider = providers.auth;
  if (providers.session) sessionProvider = providers.session;
  if (providers.toast) toastProvider = providers.toast;
}

/** Gets the auth provider */
export function getAuthProvider(): AuthProvider {
  if (!authProvider) {
    throw new Error(
      'Auth provider not configured. Call configureProviders() during app initialization.'
    );
  }
  return authProvider;
}

/** Gets the session provider */
export function getSessionProvider(): SessionProvider {
  if (!sessionProvider) {
    throw new Error(
      'Session provider not configured. Call configureProviders() during app initialization.'
    );
  }
  return sessionProvider;
}

/** Gets the toast provider */
export function getToastProvider(): ToastProvider {
  if (!toastProvider) {
    throw new Error(
      'Toast provider not configured. Call configureProviders() during app initialization.'
    );
  }
  return toastProvider;
}
