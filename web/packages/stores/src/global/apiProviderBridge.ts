// /**
//  * API Provider Bridge
//  * Bridges @samavāya/stores to @samavāya/api provider interfaces.
//  * Call initApiProviders() during app startup to wire up the stores
//  * as concrete implementations of the API provider interfaces.
//  * @packageDocumentation
//  */

// // import {
// //   configureProviders,
// //   type AuthProvider,
// //   type SessionProvider,
// //   type ToastProvider,
// // } from '@samavāya/api';
// import { get } from 'svelte/store';
// import { authStore } from './auth.store.js';
// import { sessionStore } from './session.store.js';
// import { toastStore } from './notifications.store.js';

// /**
//  * Creates an AuthProvider backed by the authStore
//  */
// function createAuthProviderFromStore(): AuthProvider {
//   return {
//     isAuthenticated() {
//       return get(authStore).isAuthenticated;
//     },
//     getTokens() {
//       const state = get(authStore);
//       if (!state.tokens) return null;
//       return {
//         accessToken: state.tokens.accessToken,
//         refreshToken: state.tokens.refreshToken,
//         expiresAt: state.tokens.expiresAt,
//       };
//     },
//     async refreshTokens() {
//       await authStore.refreshTokens();
//     },
//     logout() {
//       authStore.logout();
//     },
//   };
// }

// /**
//  * Creates a SessionProvider backed by the sessionStore
//  */
// function createSessionProviderFromStore(): SessionProvider {
//   return {
//     getSessionId() {
//       const state = get(sessionStore);
//       return state.session?.id ?? null;
//     },
//     getContext() {
//       const state = get(sessionStore);
//       const ctx = state.session?.context;
//       if (!ctx) return null;
//       return {
//         tenantId: ctx.tenantId,
//         companyId: ctx.companyId,
//         branchId: ctx.branchId,
//         userId: ctx.userId,
//       };
//     },
//   };
// }

// /**
//  * Creates a ToastProvider backed by the toastStore
//  */
// function createToastProviderFromStore(): ToastProvider {
//   return {
//     error(message: string, options?: { title?: string; duration?: number }) {
//       toastStore.error(message, options);
//     },
//   };
// }

// /**
//  * Initializes API providers with concrete store implementations.
//  * Must be called once during app startup (e.g., in root +layout.ts or shell app init).
//  */
// export function initApiProviders(): void {
//   configureProviders({
//     auth: createAuthProviderFromStore(),
//     session: createSessionProviderFromStore(),
//     toast: createToastProviderFromStore(),
//   });
// }
