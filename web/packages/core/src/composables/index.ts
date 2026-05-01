/**
 * Composables - Export all composable functions
 * @packageDocumentation
 */

// List composable
export { useList } from './useList.js';
export type { UseListOptions, UseListReturn, FetchParams, FetchResult } from './useList.js';

// Detail composable
export { useDetail } from './useDetail.js';
export type { UseDetailOptions, UseDetailReturn } from './useDetail.js';

// Form composable
export { useForm } from './useForm.js';
export type { UseFormOptions, UseFormReturn } from './useForm.js';

// Modal composable
export {
  useModal,
  useModalManager,
  useConfirmation,
  useAlert,
  usePrompt,
} from './useModal.js';
export type {
  UseModalOptions,
  UseModalReturn,
  UseModalManagerReturn,
  ConfirmOptions,
  AlertOptions,
  PromptOptions,
} from './useModal.js';

// Pagination composable
export { usePagination } from './usePagination.js';
export type { UsePaginationOptions, UsePaginationReturn } from './usePagination.js';

// Selection composable
export {
  useSelection,
  useRowSelection,
  useCheckboxGroup,
} from './useSelection.js';
export type {
  UseSelectionOptions,
  UseSelectionReturn,
  UseRowSelectionOptions,
  UseCheckboxGroupOptions,
} from './useSelection.js';
