/**
 * Notifications Store
 * Handles system notifications, toasts, and notification preferences
 */
import { writable, derived, get } from 'svelte/store';
// ============================================================================
// INITIAL STATE
// ============================================================================
const defaultPreferences = {
    enabled: true,
    sound: true,
    desktop: false,
    email: true,
    categories: {},
};
const initialNotificationState = {
    items: [],
    unreadCount: 0,
    isLoading: false,
    preferences: defaultPreferences,
    error: null,
};
const initialToastState = {
    items: [],
    defaultPosition: 'top-right',
    defaultDuration: 5000,
    maxToasts: 5,
};
// ============================================================================
// NOTIFICATION STORE
// ============================================================================
function createNotificationStore() {
    const store = writable(initialNotificationState);
    const { subscribe, update } = store;
    // ============================================================================
    // DERIVED STORES
    // ============================================================================
    const items = derived(store, ($s) => $s.items);
    const unreadCount = derived(store, ($s) => $s.unreadCount);
    const unreadItems = derived(store, ($s) => $s.items.filter((n) => !n.read));
    const isLoading = derived(store, ($s) => $s.isLoading);
    const preferences = derived(store, ($s) => $s.preferences);
    // ============================================================================
    // ACTIONS
    // ============================================================================
    async function load() {
        update((s) => ({ ...s, isLoading: true, error: null }));
        try {
            // API call would go here
            // const notifications = await notificationApi.getNotifications();
            const mockNotifications = [];
            update((s) => ({
                ...s,
                items: mockNotifications,
                unreadCount: mockNotifications.filter((n) => !n.read).length,
                isLoading: false,
            }));
        }
        catch (error) {
            update((s) => ({
                ...s,
                isLoading: false,
                error: {
                    code: 'LOAD_FAILED',
                    message: error instanceof Error ? error.message : 'Failed to load notifications',
                },
            }));
        }
    }
    function add(notification) {
        const id = `notif-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
        const newNotification = {
            ...notification,
            id,
            timestamp: new Date(),
            read: false,
        };
        update((s) => ({
            ...s,
            items: [newNotification, ...s.items],
            unreadCount: s.unreadCount + 1,
        }));
        // Play sound if enabled
        const state = get(store);
        if (state.preferences.enabled && state.preferences.sound) {
            playNotificationSound();
        }
        // Show desktop notification if enabled
        if (state.preferences.enabled && state.preferences.desktop) {
            showDesktopNotification(newNotification);
        }
        return id;
    }
    function markAsRead(id) {
        update((s) => {
            const item = s.items.find((n) => n.id === id);
            if (!item || item.read)
                return s;
            return {
                ...s,
                items: s.items.map((n) => (n.id === id ? { ...n, read: true } : n)),
                unreadCount: Math.max(0, s.unreadCount - 1),
            };
        });
    }
    function markAllAsRead() {
        update((s) => ({
            ...s,
            items: s.items.map((n) => ({ ...n, read: true })),
            unreadCount: 0,
        }));
    }
    function remove(id) {
        update((s) => {
            const item = s.items.find((n) => n.id === id);
            const wasUnread = item && !item.read;
            return {
                ...s,
                items: s.items.filter((n) => n.id !== id),
                unreadCount: wasUnread ? Math.max(0, s.unreadCount - 1) : s.unreadCount,
            };
        });
    }
    function clear() {
        update((s) => ({
            ...s,
            items: [],
            unreadCount: 0,
        }));
    }
    function clearRead() {
        update((s) => ({
            ...s,
            items: s.items.filter((n) => !n.read),
        }));
    }
    function updatePreferences(updates) {
        update((s) => ({
            ...s,
            preferences: { ...s.preferences, ...updates },
        }));
        // Persist preferences
        const state = get(store);
        localStorage.setItem('notification_preferences', JSON.stringify(state.preferences));
    }
    function loadPreferences() {
        const stored = localStorage.getItem('notification_preferences');
        if (stored) {
            try {
                const prefs = JSON.parse(stored);
                update((s) => ({
                    ...s,
                    preferences: { ...defaultPreferences, ...prefs },
                }));
            }
            catch {
                // Use defaults
            }
        }
    }
    // ============================================================================
    // HELPERS
    // ============================================================================
    function playNotificationSound() {
        try {
            // Create a simple notification sound
            const audioContext = new (window.AudioContext || window.webkitAudioContext)();
            const oscillator = audioContext.createOscillator();
            const gainNode = audioContext.createGain();
            oscillator.connect(gainNode);
            gainNode.connect(audioContext.destination);
            oscillator.frequency.value = 800;
            oscillator.type = 'sine';
            gainNode.gain.value = 0.1;
            oscillator.start();
            oscillator.stop(audioContext.currentTime + 0.1);
        }
        catch {
            // Audio not supported
        }
    }
    function showDesktopNotification(notification) {
        if (!('Notification' in window))
            return;
        if (Notification.permission === 'granted') {
            new Notification(notification.title || 'Notification', {
                body: notification.message,
                icon: notification.icon,
            });
        }
        else if (Notification.permission !== 'denied') {
            Notification.requestPermission().then((permission) => {
                if (permission === 'granted') {
                    new Notification(notification.title || 'Notification', {
                        body: notification.message,
                        icon: notification.icon,
                    });
                }
            });
        }
    }
    async function requestDesktopPermission() {
        if (!('Notification' in window))
            return false;
        const permission = await Notification.requestPermission();
        return permission === 'granted';
    }
    // ============================================================================
    // RETURN
    // ============================================================================
    return {
        subscribe,
        // Derived stores
        items,
        unreadCount,
        unreadItems,
        isLoading,
        preferences,
        // Actions
        load,
        add,
        markAsRead,
        markAllAsRead,
        remove,
        clear,
        clearRead,
        updatePreferences,
        loadPreferences,
        requestDesktopPermission,
    };
}
// ============================================================================
// TOAST STORE
// ============================================================================
function createToastStore() {
    const store = writable(initialToastState);
    const { subscribe, update } = store;
    // Timeout map for auto-dismiss
    const timeouts = new Map();
    // ============================================================================
    // DERIVED STORES
    // ============================================================================
    const items = derived(store, ($s) => $s.items);
    // ============================================================================
    // ACTIONS
    // ============================================================================
    function show(toast) {
        const id = `toast-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
        const state = get(store);
        const newToast = {
            ...toast,
            id,
            duration: toast.duration ?? state.defaultDuration,
            position: toast.position ?? state.defaultPosition,
            dismissible: toast.dismissible ?? true,
        };
        update((s) => {
            let items = [newToast, ...s.items];
            // Limit max toasts
            if (items.length > s.maxToasts) {
                const removed = items.slice(s.maxToasts);
                removed.forEach((t) => {
                    const timeout = timeouts.get(t.id);
                    if (timeout) {
                        clearTimeout(timeout);
                        timeouts.delete(t.id);
                    }
                });
                items = items.slice(0, s.maxToasts);
            }
            return { ...s, items };
        });
        // Auto-dismiss
        if (newToast.duration && newToast.duration > 0) {
            const timeout = setTimeout(() => {
                dismiss(id);
            }, newToast.duration);
            timeouts.set(id, timeout);
        }
        return id;
    }
    function dismiss(id) {
        const state = get(store);
        const toast = state.items.find((t) => t.id === id);
        // Clear timeout
        const timeout = timeouts.get(id);
        if (timeout) {
            clearTimeout(timeout);
            timeouts.delete(id);
        }
        // Call onDismiss callback
        toast?.onDismiss?.();
        update((s) => ({
            ...s,
            items: s.items.filter((t) => t.id !== id),
        }));
    }
    function dismissAll() {
        // Clear all timeouts
        timeouts.forEach((timeout) => clearTimeout(timeout));
        timeouts.clear();
        update((s) => ({ ...s, items: [] }));
    }
    function setDefaultPosition(position) {
        update((s) => ({ ...s, defaultPosition: position }));
    }
    function setDefaultDuration(duration) {
        update((s) => ({ ...s, defaultDuration: duration }));
    }
    function setMaxToasts(max) {
        update((s) => ({ ...s, maxToasts: max }));
    }
    // Convenience methods
    function success(message, options) {
        return show({ type: 'success', message, ...options });
    }
    function error(message, options) {
        return show({ type: 'error', message, duration: 0, ...options }); // Errors don't auto-dismiss
    }
    function warning(message, options) {
        return show({ type: 'warning', message, ...options });
    }
    function info(message, options) {
        return show({ type: 'info', message, ...options });
    }
    // ============================================================================
    // RETURN
    // ============================================================================
    return {
        subscribe,
        // Derived stores
        items,
        // Actions
        show,
        dismiss,
        dismissAll,
        setDefaultPosition,
        setDefaultDuration,
        setMaxToasts,
        // Convenience methods
        success,
        error,
        warning,
        info,
    };
}
// ============================================================================
// EXPORT
// ============================================================================
export const notificationStore = createNotificationStore();
export const toastStore = createToastStore();
