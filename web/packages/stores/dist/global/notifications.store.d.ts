/**
 * Notifications Store
 * Handles system notifications, toasts, and notification preferences
 */
import { type Readable } from 'svelte/store';
import type { ColorVariant, ExtendedPosition } from '@samavāya/core';
export interface Notification {
    id: string;
    type: 'info' | 'success' | 'warning' | 'error' | 'notification';
    title?: string;
    message: string;
    description?: string;
    icon?: string;
    timestamp: Date;
    read: boolean;
    persistent?: boolean;
    category?: string;
    metadata?: Record<string, unknown>;
    actions?: NotificationAction[];
    link?: {
        href: string;
        label?: string;
    };
}
export interface NotificationAction {
    id: string;
    label: string;
    variant?: 'primary' | 'secondary' | 'danger';
    handler: () => void | Promise<void>;
}
export interface Toast {
    id: string;
    type: ColorVariant;
    title?: string;
    message: string;
    duration?: number;
    dismissible?: boolean;
    position?: ExtendedPosition;
    action?: {
        label: string;
        onClick: () => void;
    };
    onDismiss?: () => void;
}
export interface NotificationPreferences {
    enabled: boolean;
    sound: boolean;
    desktop: boolean;
    email: boolean;
    categories: {
        [category: string]: {
            enabled: boolean;
            sound: boolean;
            desktop: boolean;
            email: boolean;
        };
    };
}
export interface NotificationState {
    items: Notification[];
    unreadCount: number;
    isLoading: boolean;
    preferences: NotificationPreferences;
    error: NotificationError | null;
}
export interface ToastState {
    items: Toast[];
    defaultPosition: ExtendedPosition;
    defaultDuration: number;
    maxToasts: number;
}
export interface NotificationError {
    code: string;
    message: string;
}
export declare const notificationStore: {
    subscribe: (this: void, run: import("svelte/store").Subscriber<NotificationState>, invalidate?: () => void) => import("svelte/store").Unsubscriber;
    items: Readable<Notification[]>;
    unreadCount: Readable<number>;
    unreadItems: Readable<Notification[]>;
    isLoading: Readable<boolean>;
    preferences: Readable<NotificationPreferences>;
    load: () => Promise<void>;
    add: (notification: Omit<Notification, "id" | "timestamp" | "read">) => string;
    markAsRead: (id: string) => void;
    markAllAsRead: () => void;
    remove: (id: string) => void;
    clear: () => void;
    clearRead: () => void;
    updatePreferences: (updates: Partial<NotificationPreferences>) => void;
    loadPreferences: () => void;
    requestDesktopPermission: () => Promise<boolean>;
};
export declare const toastStore: {
    subscribe: (this: void, run: import("svelte/store").Subscriber<ToastState>, invalidate?: () => void) => import("svelte/store").Unsubscriber;
    items: Readable<Toast[]>;
    show: (toast: Omit<Toast, "id">) => string;
    dismiss: (id: string) => void;
    dismissAll: () => void;
    setDefaultPosition: (position: ExtendedPosition) => void;
    setDefaultDuration: (duration: number) => void;
    setMaxToasts: (max: number) => void;
    success: (message: string, options?: Partial<Omit<Toast, "id" | "type" | "message">>) => string;
    error: (message: string, options?: Partial<Omit<Toast, "id" | "type" | "message">>) => string;
    warning: (message: string, options?: Partial<Omit<Toast, "id" | "type" | "message">>) => string;
    info: (message: string, options?: Partial<Omit<Toast, "id" | "type" | "message">>) => string;
};
//# sourceMappingURL=notifications.store.d.ts.map