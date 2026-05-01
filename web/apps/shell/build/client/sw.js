/**
 * samavāya ERP - Service Worker
 *
 * Provides:
 * - Offline-first caching strategy
 * - Background sync for offline operations
 * - Push notifications
 * - Periodic sync for data freshness
 */

const CACHE_VERSION = 'v1';
const CACHE_NAME = `samavāya-erp-${CACHE_VERSION}`;
const OFFLINE_URL = '/offline.html';

// Assets to cache on install (App Shell)
const PRECACHE_ASSETS = [
  '/',
  '/offline.html',
  '/manifest.json',
  '/favicon.png',
  '/icons.svg',
];

// Cache strategies
const CACHE_STRATEGIES = {
  // Static assets - Cache First
  static: [
    /\.(css|js|woff2?|ttf|eot|svg|png|jpg|jpeg|gif|ico|webp)$/,
    /\/_app\//,
  ],
  // API calls - Network First with cache fallback
  api: [
    /\/api\//,
  ],
  // HTML pages - Network First
  pages: [
    /^\/(?!api)/,
  ],
};

// ============================================================================
// Install Event
// ============================================================================

self.addEventListener('install', (event) => {
  console.log('[SW] Installing service worker...');

  event.waitUntil(
    (async () => {
      const cache = await caches.open(CACHE_NAME);

      // Cache app shell
      console.log('[SW] Caching app shell...');
      await cache.addAll(PRECACHE_ASSETS);

      // Skip waiting to activate immediately
      await self.skipWaiting();

      console.log('[SW] Service worker installed');
    })()
  );
});

// ============================================================================
// Activate Event
// ============================================================================

self.addEventListener('activate', (event) => {
  console.log('[SW] Activating service worker...');

  event.waitUntil(
    (async () => {
      // Clean up old caches
      const cacheNames = await caches.keys();
      await Promise.all(
        cacheNames
          .filter((name) => name.startsWith('samavāya-erp-') && name !== CACHE_NAME)
          .map((name) => {
            console.log('[SW] Deleting old cache:', name);
            return caches.delete(name);
          })
      );

      // Take control of all clients
      await self.clients.claim();

      console.log('[SW] Service worker activated');
    })()
  );
});

// ============================================================================
// Fetch Event
// ============================================================================

self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);

  // Skip non-GET requests
  if (request.method !== 'GET') {
    return;
  }

  // Skip cross-origin requests
  if (url.origin !== self.location.origin) {
    return;
  }

  // Determine caching strategy
  event.respondWith(handleFetch(request));
});

async function handleFetch(request) {
  const url = new URL(request.url);
  const pathname = url.pathname;

  // Check if this is an API request
  if (CACHE_STRATEGIES.api.some((pattern) => pattern.test(pathname))) {
    return networkFirstStrategy(request);
  }

  // Check if this is a static asset
  if (CACHE_STRATEGIES.static.some((pattern) => pattern.test(pathname))) {
    return cacheFirstStrategy(request);
  }

  // Default: Network first for pages
  return networkFirstStrategy(request);
}

// Cache First Strategy (for static assets)
async function cacheFirstStrategy(request) {
  const cachedResponse = await caches.match(request);

  if (cachedResponse) {
    // Return cached response and update cache in background
    updateCache(request);
    return cachedResponse;
  }

  try {
    const networkResponse = await fetch(request);
    const cache = await caches.open(CACHE_NAME);
    cache.put(request, networkResponse.clone());
    return networkResponse;
  } catch (error) {
    console.error('[SW] Cache first fetch failed:', error);
    return new Response('Offline', { status: 503 });
  }
}

// Network First Strategy (for dynamic content)
async function networkFirstStrategy(request) {
  try {
    const networkResponse = await fetch(request);

    // Cache successful responses
    if (networkResponse.ok) {
      const cache = await caches.open(CACHE_NAME);
      cache.put(request, networkResponse.clone());
    }

    return networkResponse;
  } catch (error) {
    console.log('[SW] Network first failed, trying cache:', request.url);

    const cachedResponse = await caches.match(request);

    if (cachedResponse) {
      return cachedResponse;
    }

    // Return offline page for navigation requests
    if (request.mode === 'navigate') {
      const offlineResponse = await caches.match(OFFLINE_URL);
      if (offlineResponse) {
        return offlineResponse;
      }
    }

    return new Response('Offline', { status: 503 });
  }
}

// Background cache update
async function updateCache(request) {
  try {
    const networkResponse = await fetch(request);
    if (networkResponse.ok) {
      const cache = await caches.open(CACHE_NAME);
      cache.put(request, networkResponse);
    }
  } catch (error) {
    // Silently fail - we have cached version
  }
}

// ============================================================================
// Background Sync
// ============================================================================

self.addEventListener('sync', (event) => {
  console.log('[SW] Background sync triggered:', event.tag);

  if (event.tag === 'sync-pending-operations') {
    event.waitUntil(syncPendingOperations());
  }
});

async function syncPendingOperations() {
  try {
    // Get pending operations from IndexedDB
    const db = await openDatabase();
    const tx = db.transaction('pending-operations', 'readonly');
    const store = tx.objectStore('pending-operations');
    const operations = await getAllFromStore(store);

    console.log('[SW] Syncing', operations.length, 'pending operations');

    for (const operation of operations) {
      try {
        const response = await fetch(operation.url, {
          method: operation.method,
          headers: operation.headers,
          body: operation.body,
        });

        if (response.ok) {
          // Remove successful operation from pending
          await removeFromPending(operation.id);
        }
      } catch (error) {
        console.error('[SW] Failed to sync operation:', operation.id, error);
      }
    }

    // Notify clients of sync completion
    const clients = await self.clients.matchAll();
    clients.forEach((client) => {
      client.postMessage({
        type: 'SYNC_COMPLETE',
        count: operations.length,
      });
    });
  } catch (error) {
    console.error('[SW] Background sync failed:', error);
  }
}

// ============================================================================
// Push Notifications
// ============================================================================

self.addEventListener('push', (event) => {
  console.log('[SW] Push notification received');

  if (!event.data) {
    return;
  }

  const data = event.data.json();

  const options = {
    body: data.body || 'New notification',
    icon: '/icons/icon-192x192.png',
    badge: '/icons/badge-72x72.png',
    vibrate: [100, 50, 100],
    data: data.data || {},
    actions: data.actions || [],
    tag: data.tag || 'default',
    renotify: data.renotify || false,
    requireInteraction: data.requireInteraction || false,
    silent: data.silent || false,
  };

  event.waitUntil(
    self.registration.showNotification(data.title || 'samavāya ERP', options)
  );
});

self.addEventListener('notificationclick', (event) => {
  console.log('[SW] Notification clicked:', event.notification.tag);

  event.notification.close();

  const urlToOpen = event.notification.data?.url || '/';

  event.waitUntil(
    (async () => {
      const allClients = await self.clients.matchAll({
        type: 'window',
        includeUncontrolled: true,
      });

      // Try to focus existing client
      for (const client of allClients) {
        if (client.url.includes(urlToOpen) && 'focus' in client) {
          return client.focus();
        }
      }

      // Open new window
      if (self.clients.openWindow) {
        return self.clients.openWindow(urlToOpen);
      }
    })()
  );
});

// ============================================================================
// Periodic Sync (for data freshness)
// ============================================================================

self.addEventListener('periodicsync', (event) => {
  console.log('[SW] Periodic sync triggered:', event.tag);

  if (event.tag === 'refresh-data') {
    event.waitUntil(refreshCriticalData());
  }
});

async function refreshCriticalData() {
  try {
    // Refresh critical data endpoints
    const criticalEndpoints = [
      '/api/user/profile',
      '/api/notifications',
      '/api/dashboard/summary',
    ];

    await Promise.all(
      criticalEndpoints.map(async (endpoint) => {
        try {
          const response = await fetch(endpoint);
          if (response.ok) {
            const cache = await caches.open(CACHE_NAME);
            cache.put(endpoint, response);
          }
        } catch (error) {
          // Silently fail
        }
      })
    );

    console.log('[SW] Critical data refreshed');
  } catch (error) {
    console.error('[SW] Failed to refresh critical data:', error);
  }
}

// ============================================================================
// Message Handler (communication with main thread)
// ============================================================================

self.addEventListener('message', (event) => {
  const { type, payload } = event.data || {};

  switch (type) {
    case 'SKIP_WAITING':
      self.skipWaiting();
      break;

    case 'CACHE_URLS':
      event.waitUntil(cacheUrls(payload.urls));
      break;

    case 'CLEAR_CACHE':
      event.waitUntil(clearCache());
      break;

    case 'GET_CACHE_SIZE':
      event.waitUntil(getCacheSize().then((size) => {
        event.ports[0].postMessage({ size });
      }));
      break;

    default:
      console.log('[SW] Unknown message type:', type);
  }
});

async function cacheUrls(urls) {
  const cache = await caches.open(CACHE_NAME);
  await cache.addAll(urls);
  console.log('[SW] Cached', urls.length, 'URLs');
}

async function clearCache() {
  const cacheNames = await caches.keys();
  await Promise.all(cacheNames.map((name) => caches.delete(name)));
  console.log('[SW] All caches cleared');
}

async function getCacheSize() {
  const cache = await caches.open(CACHE_NAME);
  const keys = await cache.keys();

  let totalSize = 0;
  for (const request of keys) {
    const response = await cache.match(request);
    if (response) {
      const blob = await response.blob();
      totalSize += blob.size;
    }
  }

  return totalSize;
}

// ============================================================================
// IndexedDB Helpers
// ============================================================================

function openDatabase() {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('samavāya-erp-sw', 1);

    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result);

    request.onupgradeneeded = (event) => {
      const db = event.target.result;

      if (!db.objectStoreNames.contains('pending-operations')) {
        db.createObjectStore('pending-operations', { keyPath: 'id', autoIncrement: true });
      }
    };
  });
}

function getAllFromStore(store) {
  return new Promise((resolve, reject) => {
    const request = store.getAll();
    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result);
  });
}

async function removeFromPending(id) {
  const db = await openDatabase();
  const tx = db.transaction('pending-operations', 'readwrite');
  const store = tx.objectStore('pending-operations');
  store.delete(id);
}
