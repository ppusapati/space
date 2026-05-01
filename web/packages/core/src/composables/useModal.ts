/**
 * useModal Composable
 * Creates a reactive modal state with open/close handling and data management
 */

import { writable, derived, get, type Writable, type Readable } from 'svelte/store';
import type { ModalConfig, ModalInstance, ModalManagerState, ModalManagerActions } from '../types/index.js';

// ============================================================================
// TYPES
// ============================================================================

export interface UseModalOptions<TData = unknown, TResult = unknown> {
  id?: string;
  config?: Partial<ModalConfig>;
  onOpen?: (data?: TData) => void;
  onClose?: () => void;
  onSubmit?: (result: TResult) => void;
  onCancel?: () => void;
}

export interface UseModalReturn<TData = unknown, TResult = unknown> {
  // State
  isOpen: Writable<boolean>;
  data: Writable<TData | null>;
  config: Writable<ModalConfig>;
  result: Writable<TResult | null>;

  // Methods
  open: (data?: TData) => void;
  close: () => void;
  submit: (result: TResult) => void;
  cancel: () => void;
  toggle: () => void;
  updateConfig: (config: Partial<ModalConfig>) => void;
  setData: (data: TData) => void;
}

export interface UseModalManagerReturn {
  // State
  modals: Readable<Map<string, ModalInstance<unknown, unknown>>>;
  activeModal: Readable<ModalInstance<unknown, unknown> | null>;
  isAnyOpen: Readable<boolean>;
  modalStack: Readable<string[]>;

  // Methods
  open: <TData = unknown, TResult = unknown>(
    id: string,
    config: ModalConfig,
    data?: TData
  ) => Promise<TResult | null>;
  close: (id: string) => void;
  closeAll: () => void;
  closeTop: () => void;
  isOpen: (id: string) => boolean;
  getModal: <TData = unknown, TResult = unknown>(id: string) => ModalInstance<TData, TResult> | undefined;
  updateData: <TData>(id: string, data: TData) => void;
}

// ============================================================================
// SINGLE MODAL
// ============================================================================

let modalIdCounter = 0;

export function useModal<TData = unknown, TResult = unknown>(
  options: UseModalOptions<TData, TResult> = {}
): UseModalReturn<TData, TResult> {
  const { id = `modal-${++modalIdCounter}`, config: initialConfig, onOpen, onClose, onSubmit, onCancel } = options;

  // ============================================================================
  // STORES
  // ============================================================================

  const isOpen = writable<boolean>(false);
  const data = writable<TData | null>(null);
  const result = writable<TResult | null>(null);

  const config = writable<ModalConfig>({
    id,
    title: '',
    size: 'md',
    closable: true,
    closeOnEscape: true,
    closeOnOverlay: true,
    preventClose: false,
    showHeader: true,
    showFooter: true,
    ...initialConfig,
  });

  // ============================================================================
  // METHODS
  // ============================================================================

  function open(openData?: TData): void {
    data.set(openData ?? null);
    result.set(null);
    isOpen.set(true);
    onOpen?.(openData);

    // Handle escape key
    const $config = get(config);
    if ($config.closeOnEscape) {
      const handleEscape = (e: KeyboardEvent) => {
        if (e.key === 'Escape' && get(isOpen) && !$config.preventClose) {
          cancel();
        }
      };
      document.addEventListener('keydown', handleEscape);

      // Clean up on close
      const unsubscribe = isOpen.subscribe(($isOpen) => {
        if (!$isOpen) {
          document.removeEventListener('keydown', handleEscape);
          unsubscribe();
        }
      });
    }
  }

  function close(): void {
    const $config = get(config);
    if ($config.preventClose) return;

    isOpen.set(false);
    onClose?.();
  }

  function submit(submitResult: TResult): void {
    result.set(submitResult);
    onSubmit?.(submitResult);
    close();
  }

  function cancel(): void {
    result.set(null);
    onCancel?.();
    close();
  }

  function toggle(): void {
    if (get(isOpen)) {
      cancel();
    } else {
      open();
    }
  }

  function updateConfig(newConfig: Partial<ModalConfig>): void {
    config.update(($c) => ({ ...$c, ...newConfig }));
  }

  function setData(newData: TData): void {
    data.set(newData);
  }

  // ============================================================================
  // RETURN
  // ============================================================================

  return {
    isOpen,
    data,
    config,
    result,
    open,
    close,
    submit,
    cancel,
    toggle,
    updateConfig,
    setData,
  };
}

// ============================================================================
// MODAL MANAGER
// ============================================================================

export function useModalManager(): UseModalManagerReturn {
  // ============================================================================
  // STORES
  // ============================================================================

  const modals = writable<Map<string, ModalInstance<unknown, unknown>>>(new Map());
  const modalStack = writable<string[]>([]);

  // Resolvers for promises
  const resolvers = new Map<string, (result: unknown) => void>();

  // ============================================================================
  // DERIVED
  // ============================================================================

  const activeModal = derived([modals, modalStack], ([$modals, $stack]) => {
    if ($stack.length === 0) return null;
    const topId = $stack[$stack.length - 1];
    if (!topId) return null;
    return $modals.get(topId) ?? null;
  });

  const isAnyOpen = derived(modalStack, ($stack) => $stack.length > 0);

  // ============================================================================
  // METHODS
  // ============================================================================

  function open<TData = unknown, TResult = unknown>(
    id: string,
    config: ModalConfig,
    data?: TData
  ): Promise<TResult | null> {
    return new Promise((resolve) => {
      const instance: ModalInstance<TData, TResult> = {
        id,
        config,
        data: data ?? null,
        isOpen: true,
        result: null,
      };

      modals.update(($m) => {
        const newMap = new Map($m);
        newMap.set(id, instance as ModalInstance<unknown, unknown>);
        return newMap;
      });

      modalStack.update(($s) => [...$s, id]);
      resolvers.set(id, resolve as (result: unknown) => void);

      // Handle escape key for stack
      const handleEscape = (e: KeyboardEvent) => {
        if (e.key === 'Escape' && !config.preventClose) {
          const $stack = get(modalStack);
          if ($stack[$stack.length - 1] === id) {
            closeModal(id, null);
          }
        }
      };

      if (config.closeOnEscape) {
        document.addEventListener('keydown', handleEscape);

        // Store cleanup function
        const cleanup = () => {
          document.removeEventListener('keydown', handleEscape);
        };

        // Check if modal is still open
        const checkOpen = () => {
          const $modals = get(modals);
          if (!$modals.has(id)) {
            cleanup();
          }
        };

        modals.subscribe(checkOpen);
      }
    });
  }

  function closeModal(id: string, result: unknown): void {
    const resolver = resolvers.get(id);

    modals.update(($m) => {
      const newMap = new Map($m);
      newMap.delete(id);
      return newMap;
    });

    modalStack.update(($s) => $s.filter((modalId) => modalId !== id));
    resolvers.delete(id);

    resolver?.(result);
  }

  function close(id: string): void {
    closeModal(id, null);
  }

  function closeAll(): void {
    const $stack = get(modalStack);
    for (const id of $stack) {
      closeModal(id, null);
    }
  }

  function closeTop(): void {
    const $stack = get(modalStack);
    const topId = $stack[$stack.length - 1];
    if (topId) {
      closeModal(topId, null);
    }
  }

  function isModalOpen(id: string): boolean {
    return get(modals).has(id);
  }

  function getModal<TData = unknown, TResult = unknown>(
    id: string
  ): ModalInstance<TData, TResult> | undefined {
    return get(modals).get(id) as ModalInstance<TData, TResult> | undefined;
  }

  function updateData<TData>(id: string, data: TData): void {
    modals.update(($m) => {
      const modal = $m.get(id);
      if (modal) {
        const newMap = new Map($m);
        newMap.set(id, { ...modal, data });
        return newMap;
      }
      return $m;
    });
  }

  // ============================================================================
  // RETURN
  // ============================================================================

  return {
    modals: { subscribe: modals.subscribe },
    activeModal,
    isAnyOpen,
    modalStack: { subscribe: modalStack.subscribe },
    open,
    close,
    closeAll,
    closeTop,
    isOpen: isModalOpen,
    getModal,
    updateData,
  };
}

// ============================================================================
// CONFIRMATION MODAL HELPER
// ============================================================================

export interface ConfirmOptions {
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  destructive?: boolean;
}

export function useConfirmation(modalManager: UseModalManagerReturn) {
  return async (options: ConfirmOptions): Promise<boolean> => {
    const result = await modalManager.open<ConfirmOptions, boolean>('confirmation', {
      id: 'confirmation',
      title: options.title,
      size: 'sm',
      closable: true,
      closeOnEscape: true,
      closeOnOverlay: false,
      showHeader: true,
      showFooter: true,
    }, options);

    return result ?? false;
  };
}

// ============================================================================
// ALERT MODAL HELPER
// ============================================================================

export interface AlertOptions {
  title: string;
  message: string;
  type?: 'info' | 'success' | 'warning' | 'error';
  confirmText?: string;
}

export function useAlert(modalManager: UseModalManagerReturn) {
  return async (options: AlertOptions): Promise<void> => {
    await modalManager.open<AlertOptions, void>('alert', {
      id: 'alert',
      title: options.title,
      size: 'sm',
      closable: true,
      closeOnEscape: true,
      closeOnOverlay: true,
      showHeader: true,
      showFooter: true,
    }, options);
  };
}

// ============================================================================
// PROMPT MODAL HELPER
// ============================================================================

export interface PromptOptions {
  title: string;
  message?: string;
  placeholder?: string;
  defaultValue?: string;
  confirmText?: string;
  cancelText?: string;
  validation?: (value: string) => boolean | string;
}

export function usePrompt(modalManager: UseModalManagerReturn) {
  return async (options: PromptOptions): Promise<string | null> => {
    const result = await modalManager.open<PromptOptions, string | null>('prompt', {
      id: 'prompt',
      title: options.title,
      size: 'sm',
      closable: true,
      closeOnEscape: true,
      closeOnOverlay: false,
      showHeader: true,
      showFooter: true,
    }, options);

    return result;
  };
}
