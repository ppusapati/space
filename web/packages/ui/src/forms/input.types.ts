/**
 * Input component types and logic
 */

import type { Size, ValidationState, FormElementProps } from '../types';

/** Input type attribute values */
export type InputType =
  | 'text'
  | 'password'
  | 'email'
  | 'number'
  | 'tel'
  | 'url'
  | 'search'
  | 'date'
  | 'time'
  | 'datetime-local';

/** Input variant for styling */
export type InputVariant = 'default' | 'filled' | 'outlined';

/** Input props interface */
export interface InputProps extends FormElementProps {
  /** Input type */
  type?: InputType;
  /** Current value */
  value?: string | number;
  /** Placeholder text */
  placeholder?: string;
  /** Input size */
  size?: Size;
  /** Visual variant */
  variant?: InputVariant;
  /** Validation state */
  state?: ValidationState;
  /** Label text */
  label?: string;
  /** Helper text shown below input */
  helperText?: string;
  /** Error message (shown when state is invalid) */
  errorText?: string;
  /** Left icon slot content */
  iconLeft?: boolean;
  /** Right icon slot content */
  iconRight?: boolean;
  /** Show clear button when input has value */
  clearable?: boolean;
  /** Minimum value (for number inputs) */
  min?: number | string;
  /** Maximum value (for number inputs) */
  max?: number | string;
  /** Step value (for number inputs) */
  step?: number | string;
  /** Minimum length */
  minlength?: number;
  /** Maximum length */
  maxlength?: number;
  /** Input pattern for validation */
  pattern?: string;
  /** Autocomplete attribute */
  autocomplete?: string;
  /** Full width */
  fullWidth?: boolean;
}

/** UnoCSS class mappings for input sizes */
export const inputSizeClasses: Record<Size, string> = {
  xs: 'h-7 px-2 text-xs',
  sm: 'h-8 px-2.5 text-sm',
  md: 'h-10 px-3 text-base',
  lg: 'h-12 px-4 text-lg',
  xl: 'h-14 px-5 text-xl',
};

/** UnoCSS class mappings for input variants */
export const inputVariantClasses: Record<InputVariant, string> = {
  default: 'bg-neutral-white border-neutral-300',
  filled: 'bg-neutral-50 border-transparent focus:bg-neutral-white',
  outlined: 'bg-transparent border-neutral-300',
};

/** UnoCSS class mappings for validation states */
export const inputStateClasses: Record<ValidationState, string> = {
  default: 'border-neutral-300 focus:ring-brand-primary-500 focus:border-brand-primary-500',
  valid: 'border-semantic-success-500 focus:ring-semantic-success-500 focus:border-semantic-success-500',
  invalid: 'border-semantic-error-500 focus:ring-semantic-error-500 focus:border-semantic-error-500',
  pending: 'border-semantic-warning-500 focus:ring-semantic-warning-500 focus:border-semantic-warning-500',
};

/** Base input classes using UnoCSS tokens */
export const inputBaseClasses =
  'block w-full border rounded-md transition-all duration-200 ' +
  'focus:outline-none focus:ring-2 focus:ring-offset-1 ' +
  'disabled:opacity-50 disabled:cursor-not-allowed disabled:bg-neutral-100 ' +
  'placeholder:text-neutral-400';

/** Label classes */
export const labelClasses = 'block text-sm font-medium text-neutral-700 mb-1';

/** Helper text classes */
export const helperTextClasses: Record<ValidationState, string> = {
  default: 'mt-1 text-sm text-neutral-500',
  valid: 'mt-1 text-sm text-semantic-success-600',
  invalid: 'mt-1 text-sm text-semantic-error-600',
  pending: 'mt-1 text-sm text-semantic-warning-600',
};

/** Required asterisk classes */
export const requiredClasses = 'text-semantic-error-500 ml-0.5';

/** Icon container classes */
export const iconContainerClasses = {
  left: 'absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-neutral-400',
  right: 'absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-neutral-400',
};

/** Clear button classes */
export const clearButtonClasses =
  'absolute inset-y-0 right-0 flex items-center pr-3 ' +
  'text-neutral-400 hover:text-neutral-600 ' +
  'focus:outline-none focus:ring-2 focus:ring-brand-primary-500 rounded cursor-pointer';
