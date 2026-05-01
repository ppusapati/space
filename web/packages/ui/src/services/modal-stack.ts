/**
 * Modal Stack Management Service
 * Manages multiple modals, dialogs, and drawers in a stack-based system
 */

import { writable, derived, get } from 'svelte/store';
import type { Writable, Readable } from 'svelte/store';
import type { Size, ColorVariant } from '../types';

/** Base modal item interface */
export interface ModalStackItem {
  /** Unique identifier */
  id: string;
  /** Modal type */
  type: 'modal' | 'dialog' | 'drawer';
  /** Z-index level (auto-calculated) */
  zIndex: number;
  /** Timestamp when opened */
  openedAt: number;
  /** Custom props passed to the modal */
  props: Record<string, unknown>;
  /** Resolve function for promise-based modals */
  resolve?: (value: unknown) => void;
  /** Reject function for promise-based modals */
  reject?: (reason?: unknown) => void;
}

/** Modal configuration */
export interface ModalConfig {
  /** Modal title */
  title?: string;
  /** Modal size */
  size?: Size | 'full';
  /** Close on backdrop click */
  closeOnBackdrop?: boolean;
  /** Close on escape key */
  closeOnEscape?: boolean;
  /** Show close button */
  showClose?: boolean;
  /** Center vertically */
  centered?: boolean;
  /** Prevent body scroll */
  preventScroll?: boolean;
  /** Additional props */
  [key: string]: unknown;
}

/** Dialog configuration */
export interface DialogConfig extends ModalConfig {
  /** Dialog variant */
  variant?: 'info' | 'warning' | 'error' | 'success' | 'confirm';
  /** Confirm button text */
  confirmText?: string;
  /** Cancel button text */
  cancelText?: string;
  /** Destructive action styling */
  destructive?: boolean;
  /** Message content */
  message?: string;
}

/** Drawer configuration */
export interface DrawerConfig {
  /** Drawer title */
  title?: string;
  /** Position */
  position?: 'left' | 'right' | 'top' | 'bottom';
  /** Drawer size */
  size?: Size | 'full';
  /** Close on backdrop click */
  closeOnBackdrop?: boolean;
  /** Close on escape key */
  closeOnEscape?: boolean;
  /** Show close button */
  showClose?: boolean;
  /** Show overlay/backdrop */
  overlay?: boolean;
  /** Additional props */
  [key: string]: unknown;
}

/** Result from closing a modal */
export interface ModalResult<T = unknown> {
  /** Whether the modal was confirmed (vs cancelled) */
  confirmed: boolean;
  /** Result data */
  data?: T;
}

/** Modal stack state */
interface ModalStackState {
  /** Stack of open modals */
  stack: ModalStackItem[];
  /** Base z-index for modal stacking */
  baseZIndex: number;
  /** Z-index increment per modal */
  zIndexIncrement: number;
}

const INITIAL_STATE: ModalStackState = {
  stack: [],
  baseZIndex: 1000,
  zIndexIncrement: 10,
};

/** Generate unique ID */
function generateId(): string {
  return `modal-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
}

/**
 * Create a modal stack store
 */
function createModalStack() {
  const store: Writable<ModalStackState> = writable(INITIAL_STATE);
  const { subscribe, update, set } = store;

  /** Get current z-index for new modal */
  function getNextZIndex(state: ModalStackState): number {
    return state.baseZIndex + state.stack.length * state.zIndexIncrement;
  }

  /** Open a modal */
  function openModal<T = unknown>(config: ModalConfig = {}): Promise<ModalResult<T>> {
    return new Promise((resolve, reject) => {
      update((state) => {
        const id = generateId();
        const item: ModalStackItem = {
          id,
          type: 'modal',
          zIndex: getNextZIndex(state),
          openedAt: Date.now(),
          props: { ...config },
          resolve: resolve as (value: unknown) => void,
          reject,
        };
        return { ...state, stack: [...state.stack, item] };
      });
    });
  }

  /** Open a dialog */
  function openDialog<T = unknown>(config: DialogConfig = {}): Promise<ModalResult<T>> {
    return new Promise((resolve, reject) => {
      update((state) => {
        const id = generateId();
        const item: ModalStackItem = {
          id,
          type: 'dialog',
          zIndex: getNextZIndex(state),
          openedAt: Date.now(),
          props: {
            variant: 'confirm',
            confirmText: 'Confirm',
            cancelText: 'Cancel',
            closeOnBackdrop: false,
            ...config,
          },
          resolve: resolve as (value: unknown) => void,
          reject,
        };
        return { ...state, stack: [...state.stack, item] };
      });
    });
  }

  /** Open a drawer */
  function openDrawer<T = unknown>(config: DrawerConfig = {}): Promise<ModalResult<T>> {
    return new Promise((resolve, reject) => {
      update((state) => {
        const id = generateId();
        const item: ModalStackItem = {
          id,
          type: 'drawer',
          zIndex: getNextZIndex(state),
          openedAt: Date.now(),
          props: {
            position: 'right',
            closeOnBackdrop: true,
            closeOnEscape: true,
            showClose: true,
            overlay: true,
            ...config,
          },
          resolve: resolve as (value: unknown) => void,
          reject,
        };
        return { ...state, stack: [...state.stack, item] };
      });
    });
  }

  /** Close a specific modal by ID */
  function close<T = unknown>(id: string, result?: ModalResult<T>): void {
    update((state) => {
      const item = state.stack.find((m) => m.id === id);
      if (item?.resolve) {
        item.resolve(result ?? { confirmed: false });
      }
      return {
        ...state,
        stack: state.stack.filter((m) => m.id !== id),
      };
    });
  }

  /** Close the topmost modal */
  function closeTop<T = unknown>(result?: ModalResult<T>): void {
    const state = get(store);
    const topModal = state.stack[state.stack.length - 1];
    if (topModal) {
      close(topModal.id, result);
    }
  }

  /** Confirm and close the topmost modal */
  function confirm<T = unknown>(data?: T): void {
    closeTop({ confirmed: true, data });
  }

  /** Cancel and close the topmost modal */
  function cancel(): void {
    closeTop({ confirmed: false });
  }

  /** Close all modals */
  function closeAll(): void {
    update((state) => {
      state.stack.forEach((item) => {
        if (item.resolve) {
          item.resolve({ confirmed: false });
        }
      });
      return { ...state, stack: [] };
    });
  }

  /** Check if a modal is open by ID */
  function isOpen(id: string): boolean {
    return get(store).stack.some((m) => m.id === id);
  }

  /** Get modal by ID */
  function getModal(id: string): ModalStackItem | undefined {
    return get(store).stack.find((m) => m.id === id);
  }

  /** Derived store for current stack */
  const stack: Readable<ModalStackItem[]> = derived(store, ($state) => $state.stack);

  /** Derived store for whether any modal is open */
  const hasOpenModals: Readable<boolean> = derived(store, ($state) => $state.stack.length > 0);

  /** Derived store for topmost modal */
  const topModal: Readable<ModalStackItem | null> = derived(
    store,
    ($state) => $state.stack[$state.stack.length - 1] ?? null
  );

  /** Derived store for modal count */
  const count: Readable<number> = derived(store, ($state) => $state.stack.length);

  return {
    subscribe,
    // Opening methods
    openModal,
    openDialog,
    openDrawer,
    // Alias methods
    open: openModal,
    alert: (config: DialogConfig) => openDialog({ ...config, variant: 'info', cancelText: '' }),
    confirm: (config: DialogConfig) => openDialog({ ...config, variant: 'confirm' }),
    warning: (config: DialogConfig) => openDialog({ ...config, variant: 'warning' }),
    error: (config: DialogConfig) => openDialog({ ...config, variant: 'error' }),
    success: (config: DialogConfig) => openDialog({ ...config, variant: 'success' }),
    // Closing methods
    close,
    closeTop,
    closeAll,
    // Result methods
    confirmTop: confirm,
    cancelTop: cancel,
    // Query methods
    isOpen,
    getModal,
    // Derived stores
    stack,
    hasOpenModals,
    topModal,
    count,
  };
}

/** Singleton modal stack instance */
export const modalStack = createModalStack();

/** Convenience exports */
export const {
  openModal,
  openDialog,
  openDrawer,
  close: closeModal,
  closeTop: closeTopModal,
  closeAll: closeAllModals,
} = modalStack;

/** Modal stack classes for ModalStackRenderer */
export const modalStackClasses = {
  container: 'modal-stack-container fixed inset-0 pointer-events-none',
  item: 'modal-stack-item pointer-events-auto',
};
