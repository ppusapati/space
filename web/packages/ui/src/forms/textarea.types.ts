/**
 * TextArea component types and logic
 */

import type { Size, ValidationState, FormElementProps } from '../types';

/** TextArea resize options */
export type TextAreaResize = 'none' | 'vertical' | 'horizontal' | 'both';

/** TextArea variant */
export type TextAreaVariant = 'default' | 'filled' | 'outlined';

/** TextArea props interface */
export interface TextAreaProps extends FormElementProps {
  /** Current value */
  value?: string;
  /** Placeholder text */
  placeholder?: string;
  /** TextArea size */
  size?: Size;
  /** Visual variant */
  variant?: TextAreaVariant;
  /** Validation state */
  state?: ValidationState;
  /** Label text */
  label?: string;
  /** Helper text */
  helperText?: string;
  /** Error message */
  errorText?: string;
  /** Number of visible rows */
  rows?: number;
  /** Minimum rows (for auto-resize) */
  minRows?: number;
  /** Maximum rows (for auto-resize) */
  maxRows?: number;
  /** Resize behavior */
  resize?: TextAreaResize;
  /** Maximum length */
  maxlength?: number;
  /** Show character count */
  showCount?: boolean;
  /** Auto-resize based on content */
  autoResize?: boolean;
  /** Full width */
  fullWidth?: boolean;
}

/** UnoCSS class mappings for textarea sizes */
export const textareaSizeClasses: Record<Size, string> = {
  xs: 'px-2 py-1.5 text-xs',
  sm: 'px-2.5 py-2 text-sm',
  md: 'px-3 py-2.5 text-base',
  lg: 'px-4 py-3 text-lg',
  xl: 'px-5 py-4 text-xl',
};

/** UnoCSS class mappings for textarea variants */
export const textareaVariantClasses: Record<TextAreaVariant, string> = {
  default: 'bg-neutral-white border-neutral-300',
  filled: 'bg-neutral-50 border-transparent focus:bg-neutral-white',
  outlined: 'bg-transparent border-neutral-300',
};

/** UnoCSS class mappings for validation states */
export const textareaStateClasses: Record<ValidationState, string> = {
  default: 'border-neutral-300 focus:ring-brand-primary-500 focus:border-brand-primary-500',
  valid: 'border-semantic-success-500 focus:ring-semantic-success-500 focus:border-semantic-success-500',
  invalid: 'border-semantic-error-500 focus:ring-semantic-error-500 focus:border-semantic-error-500',
  pending: 'border-semantic-warning-500 focus:ring-semantic-warning-500 focus:border-semantic-warning-500',
};

/** UnoCSS class mappings for resize behavior */
export const textareaResizeClasses: Record<TextAreaResize, string> = {
  none: 'resize-none',
  vertical: 'resize-y',
  horizontal: 'resize-x',
  both: 'resize',
};

/** Base textarea classes */
export const textareaBaseClasses =
  'block w-full border rounded-md transition-all duration-200 ' +
  'focus:outline-none focus:ring-2 focus:ring-offset-1 ' +
  'disabled:opacity-50 disabled:cursor-not-allowed disabled:bg-neutral-100 ' +
  'placeholder:text-neutral-400';

/** Character count classes */
export const charCountClasses = 'text-xs text-neutral-500 mt-1 text-right';

/** Helper text classes */
export const textareaHelperClasses: Record<ValidationState, string> = {
  default: 'mt-1 text-sm text-neutral-500',
  valid: 'mt-1 text-sm text-semantic-success-600',
  invalid: 'mt-1 text-sm text-semantic-error-600',
  pending: 'mt-1 text-sm text-semantic-warning-600',
};

/** Calculate height for auto-resize */
export function calculateAutoHeight(
  element: HTMLTextAreaElement,
  minRows: number,
  maxRows: number
): void {
  // Reset height to recalculate
  element.style.height = 'auto';

  const lineHeight = parseInt(getComputedStyle(element).lineHeight) || 20;
  const paddingTop = parseInt(getComputedStyle(element).paddingTop) || 0;
  const paddingBottom = parseInt(getComputedStyle(element).paddingBottom) || 0;

  const minHeight = lineHeight * minRows + paddingTop + paddingBottom;
  const maxHeight = lineHeight * maxRows + paddingTop + paddingBottom;

  const newHeight = Math.min(Math.max(element.scrollHeight, minHeight), maxHeight);
  element.style.height = `${newHeight}px`;
}
