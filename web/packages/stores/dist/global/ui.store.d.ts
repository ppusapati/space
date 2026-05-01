/**
 * UI Store
 * Handles UI state like modals, drawers, loading, and command palette
 */
import { type Readable } from 'svelte/store';
import type { Component, Snippet } from 'svelte';
export interface ModalItem {
    id: string;
    component?: Component;
    snippet?: Snippet;
    props?: Record<string, unknown>;
    config: {
        title?: string;
        size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
        closable?: boolean;
        closeOnEscape?: boolean;
        closeOnOverlay?: boolean;
        preventClose?: boolean;
    };
    resolve?: (value: unknown) => void;
    reject?: (reason?: unknown) => void;
}
export interface DrawerItem {
    id: string;
    component?: Component;
    snippet?: Snippet;
    props?: Record<string, unknown>;
    config: {
        title?: string;
        position?: 'left' | 'right' | 'top' | 'bottom';
        size?: 'sm' | 'md' | 'lg' | 'xl';
        closable?: boolean;
        closeOnEscape?: boolean;
        closeOnOverlay?: boolean;
        overlay?: boolean;
    };
}
export interface LoadingItem {
    id: string;
    message?: string;
    progress?: number;
    cancelable?: boolean;
    onCancel?: () => void;
}
export interface CommandItem {
    id: string;
    label: string;
    description?: string;
    icon?: string;
    shortcut?: string;
    group?: string;
    keywords?: string[];
    disabled?: boolean;
    action: () => void | Promise<void>;
}
export interface CommandGroup {
    id: string;
    label: string;
    priority?: number;
    items: CommandItem[];
}
export interface ModalState {
    stack: ModalItem[];
}
export interface DrawerState {
    current: DrawerItem | null;
    stack: DrawerItem[];
}
export interface LoadingState {
    global: boolean;
    globalMessage?: string;
    items: Map<string, LoadingItem>;
    progress?: {
        value: number;
        message?: string;
    };
}
export interface CommandPaletteState {
    isOpen: boolean;
    query: string;
    groups: CommandGroup[];
    selectedIndex: number;
    mode: 'search' | 'command';
}
export declare const modalStore: {
    subscribe: (this: void, run: import("svelte/store").Subscriber<ModalState>, invalidate?: () => void) => import("svelte/store").Unsubscriber;
    stack: Readable<ModalItem[]>;
    current: Readable<ModalItem | null>;
    isAnyOpen: Readable<boolean>;
    count: Readable<number>;
    open: <T = unknown>(modal: Omit<ModalItem, "id" | "resolve" | "reject">) => Promise<T>;
    close: (id?: string, result?: unknown) => void;
    closeAll: () => void;
    reject: (id?: string, reason?: unknown) => void;
    updateProps: (id: string, props: Record<string, unknown>) => void;
};
export declare const drawerStore: {
    subscribe: (this: void, run: import("svelte/store").Subscriber<DrawerState>, invalidate?: () => void) => import("svelte/store").Unsubscriber;
    current: Readable<DrawerItem | null>;
    isOpen: Readable<boolean>;
    stack: Readable<DrawerItem[]>;
    open: (drawer: Omit<DrawerItem, "id">) => string;
    close: (id?: string) => void;
    closeAll: () => void;
    updateProps: (id: string, props: Record<string, unknown>) => void;
};
export declare const loadingStore: {
    subscribe: (this: void, run: import("svelte/store").Subscriber<LoadingState>, invalidate?: () => void) => import("svelte/store").Unsubscriber;
    isGlobalLoading: Readable<boolean>;
    globalMessage: Readable<string | undefined>;
    activeCount: Readable<number>;
    isAnyLoading: Readable<boolean>;
    progress: Readable<{
        value: number;
        message?: string;
    } | undefined>;
    setGlobal: (loading: boolean, message?: string) => void;
    start: (id: string, options?: Omit<LoadingItem, "id">) => void;
    stop: (id: string) => void;
    stopAll: () => void;
    updateProgress: (id: string, progress: number, message?: string) => void;
    setProgress: (value: number, message?: string) => void;
    clearProgress: () => void;
    isLoading: (id: string) => boolean;
};
export declare const commandPaletteStore: {
    subscribe: (this: void, run: import("svelte/store").Subscriber<CommandPaletteState>, invalidate?: () => void) => import("svelte/store").Unsubscriber;
    isOpen: Readable<boolean>;
    query: Readable<string>;
    groups: Readable<CommandGroup[]>;
    selectedIndex: Readable<number>;
    filteredGroups: Readable<CommandGroup[]>;
    flatItems: Readable<CommandItem[]>;
    open: (mode?: "search" | "command") => void;
    close: () => void;
    toggle: () => void;
    setQuery: (query: string) => void;
    registerGroup: (group: CommandGroup) => void;
    unregisterGroup: (groupId: string) => void;
    registerCommand: (groupId: string, command: CommandItem) => void;
    unregisterCommand: (commandId: string) => void;
    selectNext: () => void;
    selectPrev: () => void;
    selectIndex: (index: number) => void;
    executeSelected: () => Promise<void>;
    execute: (commandId: string) => Promise<void>;
};
//# sourceMappingURL=ui.store.d.ts.map