import { Writable, Readable } from 'svelte/store';
import { ModalConfig, ModalInstance } from '../types/index.js';
export interface UseModalOptions<TData = unknown, TResult = unknown> {
    id?: string;
    config?: Partial<ModalConfig>;
    onOpen?: (data?: TData) => void;
    onClose?: () => void;
    onSubmit?: (result: TResult) => void;
    onCancel?: () => void;
}
export interface UseModalReturn<TData = unknown, TResult = unknown> {
    isOpen: Writable<boolean>;
    data: Writable<TData | null>;
    config: Writable<ModalConfig>;
    result: Writable<TResult | null>;
    open: (data?: TData) => void;
    close: () => void;
    submit: (result: TResult) => void;
    cancel: () => void;
    toggle: () => void;
    updateConfig: (config: Partial<ModalConfig>) => void;
    setData: (data: TData) => void;
}
export interface UseModalManagerReturn {
    modals: Readable<Map<string, ModalInstance<unknown, unknown>>>;
    activeModal: Readable<ModalInstance<unknown, unknown> | null>;
    isAnyOpen: Readable<boolean>;
    modalStack: Readable<string[]>;
    open: <TData = unknown, TResult = unknown>(id: string, config: ModalConfig, data?: TData) => Promise<TResult | null>;
    close: (id: string) => void;
    closeAll: () => void;
    closeTop: () => void;
    isOpen: (id: string) => boolean;
    getModal: <TData = unknown, TResult = unknown>(id: string) => ModalInstance<TData, TResult> | undefined;
    updateData: <TData>(id: string, data: TData) => void;
}
export declare function useModal<TData = unknown, TResult = unknown>(options?: UseModalOptions<TData, TResult>): UseModalReturn<TData, TResult>;
export declare function useModalManager(): UseModalManagerReturn;
export interface ConfirmOptions {
    title: string;
    message: string;
    confirmText?: string;
    cancelText?: string;
    destructive?: boolean;
}
export declare function useConfirmation(modalManager: UseModalManagerReturn): (options: ConfirmOptions) => Promise<boolean>;
export interface AlertOptions {
    title: string;
    message: string;
    type?: 'info' | 'success' | 'warning' | 'error';
    confirmText?: string;
}
export declare function useAlert(modalManager: UseModalManagerReturn): (options: AlertOptions) => Promise<void>;
export interface PromptOptions {
    title: string;
    message?: string;
    placeholder?: string;
    defaultValue?: string;
    confirmText?: string;
    cancelText?: string;
    validation?: (value: string) => boolean | string;
}
export declare function usePrompt(modalManager: UseModalManagerReturn): (options: PromptOptions) => Promise<string | null>;
//# sourceMappingURL=useModal.d.ts.map