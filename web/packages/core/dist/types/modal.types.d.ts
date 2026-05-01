import { Component, Snippet } from 'svelte';
import { ColorVariant } from './common.types.js';
import { FormValidationState } from './page.types.js';
/**
 * Generic Modal Type
 * @template TData - Input data for the modal
 * @template TResult - Result type when modal closes
 */
export interface Modal<TData = unknown, TResult = unknown> {
    isOpen: boolean;
    data: TData | null;
    config: ModalConfig;
    slots?: ModalSlots<TData>;
    open: (data?: TData) => void;
    close: () => void;
    submit: (result: TResult) => void;
    onOpen?: (data: TData) => void;
    onClose?: () => void;
    onSubmit?: (result: TResult) => void;
    onCancel?: () => void;
}
/** Modal configuration */
export interface ModalConfig {
    id: string;
    title?: string;
    description?: string;
    size: ModalSize;
    closable: boolean;
    closeOnBackdrop?: boolean;
    closeOnOverlay?: boolean;
    closeOnEscape: boolean;
    preventScroll?: boolean;
    preventClose?: boolean;
    centered?: boolean;
    persistent?: boolean;
    fullScreen?: boolean;
    animation?: 'fade' | 'slide' | 'scale' | 'none';
    showHeader?: boolean;
    showFooter?: boolean;
}
/** Modal size */
export type ModalSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl' | 'full';
/** Modal slots */
export interface ModalSlots<TData = unknown> {
    header?: Snippet<[TData]>;
    content?: Snippet<[TData]>;
    footer?: Snippet<[TData]>;
}
/**
 * Confirmation Modal
 * Simple confirm/cancel dialog
 */
export interface ConfirmationModal extends Modal<ConfirmationData, boolean> {
    variant: 'info' | 'warning' | 'danger' | 'success';
    confirmText: string;
    cancelText: string;
    loading: boolean;
}
/** Confirmation data */
export interface ConfirmationData {
    title: string;
    message: string;
    details?: string;
    confirmText?: string;
    cancelText?: string;
    variant?: 'info' | 'warning' | 'danger' | 'success';
    icon?: string;
}
/**
 * Alert Modal
 * Simple informational alert
 */
export interface AlertModal extends Modal<AlertData, void> {
    variant: ColorVariant;
    showIcon: boolean;
}
/** Alert data */
export interface AlertData {
    title: string;
    message: string;
    variant?: ColorVariant;
    icon?: string;
    confirmText?: string;
}
/**
 * Form Modal
 * Modal with embedded form
 */
export interface FormModal<TValues extends Record<string, unknown>> extends Modal<TValues | null, TValues> {
    mode: 'create' | 'edit';
    initialValues: TValues;
    validation: FormValidationState<TValues>;
    isDirty: boolean;
    isSubmitting: boolean;
    setFieldValue: <K extends keyof TValues>(field: K, value: TValues[K]) => void;
    setFieldError: <K extends keyof TValues>(field: K, error: string | null) => void;
    validate: () => Promise<boolean>;
    reset: () => void;
}
/**
 * Selection Modal
 * Modal for selecting item(s) from a list
 */
export interface SelectionModal<TItem, TSelected = TItem | TItem[]> extends Modal<SelectionModalData<TItem, TSelected>, TSelected> {
    items: TItem[];
    selected: TSelected;
    searchQuery: string;
    filteredItems: TItem[];
    multiSelect: boolean;
    searchable: boolean;
    maxSelection?: number;
    minSelection?: number;
    selectItem: (item: TItem) => void;
    deselectItem: (item: TItem) => void;
    toggleItem: (item: TItem) => void;
    selectAll: () => void;
    clearSelection: () => void;
    search: (query: string) => void;
    isSelected: (item: TItem) => boolean;
}
/** Selection modal data */
export interface SelectionModalData<TItem, TSelected> {
    items: TItem[];
    selected?: TSelected;
    title?: string;
    searchPlaceholder?: string;
}
/**
 * Preview Modal
 * Modal for previewing items (images, documents)
 */
export interface PreviewModal<TItem> extends Modal<TItem[], number> {
    items: TItem[];
    currentIndex: number;
    goTo: (index: number) => void;
    next: () => void;
    previous: () => void;
    hasNext: boolean;
    hasPrevious: boolean;
    download?: () => void;
    share?: () => void;
    delete?: () => void;
}
/**
 * Wizard Modal
 * Multi-step wizard in a modal
 */
export interface WizardModal<TSteps extends string, TData extends Record<TSteps, unknown>> extends Modal<Partial<TData>, TData> {
    steps: WizardStepConfig<TSteps>[];
    currentStep: TSteps;
    stepData: Partial<TData>;
    completedSteps: Set<TSteps>;
    isFirstStep: boolean;
    isLastStep: boolean;
    canProceed: boolean;
    canGoBack: boolean;
    progress: number;
    goToStep: (step: TSteps) => void;
    nextStep: () => void;
    prevStep: () => void;
    setStepData: <S extends TSteps>(step: S, data: TData[S]) => void;
    validateStep: (step: TSteps) => Promise<boolean>;
}
/** Wizard step config */
export interface WizardStepConfig<TId extends string = string> {
    id: TId;
    title: string;
    description?: string;
    icon?: string;
    optional?: boolean;
    validate?: () => boolean | Promise<boolean>;
}
/**
 * Command Palette Modal
 * Keyboard-driven command palette
 */
export interface CommandPaletteModal extends Modal<void, CommandAction | null> {
    query: string;
    commands: CommandGroup[];
    filteredCommands: CommandAction[];
    selectedIndex: number;
    recentCommands: CommandAction[];
    search: (query: string) => void;
    executeCommand: (command: CommandAction) => Promise<void>;
    selectNext: () => void;
    selectPrevious: () => void;
    selectFirst: () => void;
    selectLast: () => void;
}
/** Command group */
export interface CommandGroup {
    id: string;
    title: string;
    commands: CommandAction[];
    priority?: number;
}
/** Command action */
export interface CommandAction {
    id: string;
    label: string;
    description?: string;
    icon?: string;
    shortcut?: string;
    keywords?: string[];
    group?: string;
    disabled?: boolean;
    handler: () => void | Promise<void>;
}
/**
 * Drawer Modal
 * Side panel/drawer
 */
export interface DrawerModal<TData = unknown> extends Modal<TData, TData | null> {
    position: 'left' | 'right' | 'top' | 'bottom';
    drawerSize: string;
    overlay: boolean;
    push: boolean;
}
/**
 * Prompt Modal
 * Modal for user input
 */
export interface PromptModal extends Modal<PromptData, string | null> {
    value: string;
    placeholder: string;
    validation?: (value: string) => string | null;
    error: string | null;
}
/** Prompt data */
export interface PromptData {
    title: string;
    message?: string;
    placeholder?: string;
    defaultValue?: string;
    inputType?: 'text' | 'textarea' | 'number' | 'email';
    validation?: (value: string) => string | null;
    confirmText?: string;
    cancelText?: string;
}
/**
 * Image Crop Modal
 * Modal for cropping images
 */
export interface ImageCropModal extends Modal<ImageCropData, CroppedImage | null> {
    aspectRatio?: number;
    minWidth?: number;
    minHeight?: number;
    maxWidth?: number;
    maxHeight?: number;
    circular?: boolean;
}
/** Image crop data */
export interface ImageCropData {
    src: string;
    aspectRatio?: number;
    title?: string;
}
/** Cropped image result */
export interface CroppedImage {
    blob: Blob;
    dataUrl: string;
    width: number;
    height: number;
}
/** Modal instance */
export interface ModalInstance<TData = unknown, TResult = unknown> {
    id: string;
    component?: Component;
    props?: Record<string, unknown>;
    config: ModalConfig;
    data: TData | null;
    isOpen?: boolean;
    result?: TResult | null;
    resolve?: (value: TResult) => void;
    reject?: (error: unknown) => void;
}
/** Modal manager state */
export interface ModalManagerState {
    stack: ModalInstance[];
    isAnyOpen: boolean;
    topModal: ModalInstance | null;
}
/** Modal manager actions */
export interface ModalManagerActions {
    open: <TData, TResult>(component: Component, props?: Record<string, unknown>, config?: Partial<ModalConfig>) => Promise<TResult>;
    close: (id?: string) => void;
    closeAll: () => void;
    confirm: (data: ConfirmationData) => Promise<boolean>;
    alert: (data: AlertData) => Promise<void>;
    prompt: (data: PromptData) => Promise<string | null>;
}
//# sourceMappingURL=modal.types.d.ts.map