/**
 * PWA Store
 *
 * Manages Progressive Web App functionality:
 * - Service worker registration and updates
 * - Install prompt handling
 * - Offline detection
 * - Background sync
 * - Push notification permissions
 * - Cache management
 */

import { writable, derived, get } from 'svelte/store';

// ============================================================================
// Types
// ============================================================================

export type InstallState = 'idle' | 'can-install' | 'installing' | 'installed';

export type ServiceWorkerState =
  | 'unsupported'
  | 'installing'
  | 'installed'
  | 'activating'
  | 'activated'
  | 'redundant';

export type NotificationPermission = 'default' | 'granted' | 'denied';

export interface PwaState {
  /** Whether the browser supports PWA features */
  supported: boolean;
  /** Current install state */
  installState: InstallState;
  /** Service worker registration */
  registration: ServiceWorkerRegistration | null;
  /** Service worker state */
  swState: ServiceWorkerState;
  /** Whether app is running as installed PWA */
  isInstalled: boolean;
  /** Whether device is online */
  isOnline: boolean;
  /** Push notification permission */
  notificationPermission: NotificationPermission;
  /** Whether an update is available */
  updateAvailable: boolean;
  /** Waiting service worker (for update) */
  waitingWorker: ServiceWorker | null;
  /** Last sync timestamp */
  lastSyncAt: Date | null;
  /** Pending offline operations count */
  pendingOperations: number;
}

export interface BeforeInstallPromptEvent extends Event {
  prompt(): Promise<void>;
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
}

// ============================================================================
// Initial State
// ============================================================================

const initialState: PwaState = {
  supported: typeof window !== 'undefined' && 'serviceWorker' in navigator,
  installState: 'idle',
  registration: null,
  swState: 'unsupported',
  isInstalled: false,
  isOnline: typeof navigator !== 'undefined' ? navigator.onLine : true,
  notificationPermission: 'default',
  updateAvailable: false,
  waitingWorker: null,
  lastSyncAt: null,
  pendingOperations: 0,
};

// ============================================================================
// Store
// ============================================================================

function createPwaStore() {
  const { subscribe, set, update } = writable<PwaState>(initialState);

  let deferredPrompt: BeforeInstallPromptEvent | null = null;
  let initialized = false;

  return {
    subscribe,

    /**
     * Initialize PWA functionality
     */
    async init() {
      if (initialized || typeof window === 'undefined') return;
      initialized = true;

      const state = get({ subscribe });

      // Check if already installed
      const isInstalled =
        window.matchMedia('(display-mode: standalone)').matches ||
        (window.navigator as any).standalone === true;

      update((s) => ({ ...s, isInstalled }));

      // Listen for install prompt
      window.addEventListener('beforeinstallprompt', (e) => {
        e.preventDefault();
        deferredPrompt = e as BeforeInstallPromptEvent;
        update((s) => ({ ...s, installState: 'can-install' }));
      });

      // Listen for app installed
      window.addEventListener('appinstalled', () => {
        deferredPrompt = null;
        update((s) => ({
          ...s,
          installState: 'installed',
          isInstalled: true,
        }));
      });

      // Listen for online/offline
      window.addEventListener('online', () => {
        update((s) => ({ ...s, isOnline: true }));
        this.triggerSync();
      });

      window.addEventListener('offline', () => {
        update((s) => ({ ...s, isOnline: false }));
      });

      // Check notification permission
      if ('Notification' in window) {
        update((s) => ({
          ...s,
          notificationPermission: Notification.permission as NotificationPermission,
        }));
      }

      // Register service worker
      if (state.supported) {
        await this.registerServiceWorker();
      }
    },

    /**
     * Register service worker
     */
    async registerServiceWorker() {
      if (!('serviceWorker' in navigator)) {
        update((s) => ({ ...s, swState: 'unsupported' }));
        return;
      }

      try {
        update((s) => ({ ...s, swState: 'installing' }));

        const registration = await navigator.serviceWorker.register('/sw.js', {
          scope: '/',
        });

        update((s) => ({
          ...s,
          registration,
          swState: 'installed',
        }));

        // Check for updates
        registration.addEventListener('updatefound', () => {
          const newWorker = registration.installing;

          if (newWorker) {
            newWorker.addEventListener('statechange', () => {
              if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
                // New update available
                update((s) => ({
                  ...s,
                  updateAvailable: true,
                  waitingWorker: newWorker,
                }));
              }
            });
          }
        });

        // Handle controller change (after update)
        navigator.serviceWorker.addEventListener('controllerchange', () => {
          window.location.reload();
        });

        // Listen for messages from service worker
        navigator.serviceWorker.addEventListener('message', (event) => {
          const { type, count } = event.data || {};

          if (type === 'SYNC_COMPLETE') {
            update((s) => ({
              ...s,
              lastSyncAt: new Date(),
              pendingOperations: Math.max(0, s.pendingOperations - (count || 0)),
            }));
          }
        });

        // Wait for activation
        if (registration.active) {
          update((s) => ({ ...s, swState: 'activated' }));
        } else if (registration.installing) {
          registration.installing.addEventListener('statechange', function () {
            if (this.state === 'activated') {
              update((s) => ({ ...s, swState: 'activated' }));
            }
          });
        }

        console.log('[PWA] Service worker registered');
      } catch (error) {
        console.error('[PWA] Service worker registration failed:', error);
        update((s) => ({ ...s, swState: 'redundant' }));
      }
    },

    /**
     * Prompt user to install the app
     */
    async promptInstall(): Promise<boolean> {
      if (!deferredPrompt) {
        console.log('[PWA] Install prompt not available');
        return false;
      }

      update((s) => ({ ...s, installState: 'installing' }));

      try {
        deferredPrompt.prompt();
        const { outcome } = await deferredPrompt.userChoice;

        if (outcome === 'accepted') {
          update((s) => ({ ...s, installState: 'installed', isInstalled: true }));
          return true;
        } else {
          update((s) => ({ ...s, installState: 'idle' }));
          return false;
        }
      } finally {
        deferredPrompt = null;
      }
    },

    /**
     * Apply pending update
     */
    applyUpdate() {
      const state = get({ subscribe });

      if (state.waitingWorker) {
        state.waitingWorker.postMessage({ type: 'SKIP_WAITING' });
      }
    },

    /**
     * Request notification permission
     */
    async requestNotificationPermission(): Promise<NotificationPermission> {
      if (!('Notification' in window)) {
        return 'denied';
      }

      const permission = await Notification.requestPermission();
      update((s) => ({ ...s, notificationPermission: permission as NotificationPermission }));
      return permission as NotificationPermission;
    },

    /**
     * Subscribe to push notifications
     */
    async subscribeToPush(vapidPublicKey: string): Promise<PushSubscription | null> {
      const state = get({ subscribe });

      if (!state.registration) {
        console.error('[PWA] No service worker registration');
        return null;
      }

      try {
        const subscription = await state.registration.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: urlBase64ToUint8Array(vapidPublicKey).buffer as ArrayBuffer,
        });

        return subscription;
      } catch (error) {
        console.error('[PWA] Push subscription failed:', error);
        return null;
      }
    },

    /**
     * Trigger background sync
     */
    async triggerSync() {
      const state = get({ subscribe });

      if (!state.registration) return;

      try {
        await (state.registration as any).sync?.register('sync-pending-operations');
        console.log('[PWA] Background sync registered');
      } catch (error) {
        console.log('[PWA] Background sync not supported');
      }
    },

    /**
     * Add operation to pending queue (for offline support)
     */
    async addPendingOperation(operation: {
      url: string;
      method: string;
      headers: Record<string, string>;
      body?: string;
    }) {
      try {
        const db = await openDatabase();
        const tx = db.transaction('pending-operations', 'readwrite');
        const store = tx.objectStore('pending-operations');
        await addToStore(store, operation);

        update((s) => ({ ...s, pendingOperations: s.pendingOperations + 1 }));

        // Try to sync immediately if online
        const state = get({ subscribe });
        if (state.isOnline) {
          this.triggerSync();
        }
      } catch (error) {
        console.error('[PWA] Failed to queue operation:', error);
      }
    },

    /**
     * Clear all caches
     */
    async clearCaches() {
      const state = get({ subscribe });

      if (state.registration?.active) {
        const messageChannel = new MessageChannel();

        return new Promise<void>((resolve) => {
          messageChannel.port1.onmessage = () => resolve();
          state.registration?.active?.postMessage(
            { type: 'CLEAR_CACHE' },
            [messageChannel.port2]
          );
        });
      }
    },

    /**
     * Get cache size
     */
    async getCacheSize(): Promise<number> {
      const state = get({ subscribe });

      if (!state.registration?.active) return 0;

      const messageChannel = new MessageChannel();

      return new Promise((resolve) => {
        messageChannel.port1.onmessage = (event) => {
          resolve(event.data.size || 0);
        };
        state.registration?.active?.postMessage(
          { type: 'GET_CACHE_SIZE' },
          [messageChannel.port2]
        );
      });
    },

    /**
     * Reset store
     */
    reset() {
      set(initialState);
      initialized = false;
    },
  };
}

export const pwaStore = createPwaStore();

// ============================================================================
// Derived Stores
// ============================================================================

/**
 * Whether install prompt can be shown
 */
export const canInstall = derived(
  pwaStore,
  ($state) => $state.installState === 'can-install'
);

/**
 * Whether app is running offline
 */
export const isOffline = derived(pwaStore, ($state) => !$state.isOnline);

/**
 * Whether an update is available
 */
export const hasUpdate = derived(pwaStore, ($state) => $state.updateAvailable);

/**
 * Whether push notifications are enabled
 */
export const notificationsEnabled = derived(
  pwaStore,
  ($state) => $state.notificationPermission === 'granted'
);

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Initialize PWA
 */
export function initPwa() {
  pwaStore.init();
}

/**
 * Convert VAPID key to Uint8Array
 */
function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
  const rawData = window.atob(base64);
  const outputArray = new Uint8Array(rawData.length);

  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i);
  }

  return outputArray;
}

/**
 * Open IndexedDB database
 */
function openDatabase(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('samavāya-erp-pwa', 1);

    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result);

    request.onupgradeneeded = (event) => {
      const db = (event.target as IDBOpenDBRequest).result;

      if (!db.objectStoreNames.contains('pending-operations')) {
        db.createObjectStore('pending-operations', {
          keyPath: 'id',
          autoIncrement: true,
        });
      }
    };
  });
}

/**
 * Add item to IndexedDB store
 */
function addToStore(store: IDBObjectStore, item: unknown): Promise<void> {
  return new Promise((resolve, reject) => {
    const request = store.add(item);
    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve();
  });
}

// ============================================================================
// Formatting Helpers
// ============================================================================

/**
 * Format bytes to human readable size
 */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}
