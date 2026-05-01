/**
 * Select component types and logic
 */

import type { Size, ValidationState, FormElementProps } from '../types';

/** Option item for select */
export interface SelectOption<T = string> {
  value: T;
  label: string;
  disabled?: boolean;
  group?: string;
}

/** Option group */
export interface SelectOptionGroup<T = string> {
  label: string;
  options: SelectOption<T>[];
}

/** Select variant */
export type SelectVariant = 'default' | 'filled' | 'outlined';

/** Select props interface */
export interface SelectProps<T = string> extends FormElementProps {
  /** Current selected value(s) */
  value?: T | T[];
  /** Available options */
  options?: SelectOption<T>[];
  /** Grouped options */
  groups?: SelectOptionGroup<T>[];
  /** Placeholder text */
  placeholder?: string;
  /** Select size */
  size?: Size;
  /** Visual variant */
  variant?: SelectVariant;
  /** Validation state */
  state?: ValidationState;
  /** Label text */
  label?: string;
  /** Helper text */
  helperText?: string;
  /** Error message */
  errorText?: string;
  /** Allow multiple selection */
  multiple?: boolean;
  /** Searchable/filterable */
  searchable?: boolean;
  /** Clearable */
  clearable?: boolean;
  /** Full width */
  fullWidth?: boolean;
}

/** UnoCSS class mappings for select sizes */
export const selectSizeClasses: Record<Size, string> = {
  xs: 'h-7 px-2 text-xs',
  sm: 'h-8 px-2.5 text-sm',
  md: 'h-10 px-3 text-base',
  lg: 'h-12 px-4 text-lg',
  xl: 'h-14 px-5 text-xl',
};

/** UnoCSS class mappings for select variants */
export const selectVariantClasses: Record<SelectVariant, string> = {
  default: 'bg-neutral-white border-neutral-300',
  filled: 'bg-neutral-50 border-transparent focus:bg-neutral-white',
  outlined: 'bg-transparent border-neutral-300',
};

/** UnoCSS class mappings for validation states */
export const selectStateClasses: Record<ValidationState, string> = {
  default: 'border-neutral-300 focus:ring-brand-primary-500 focus:border-brand-primary-500',
  valid: 'border-semantic-success-500 focus:ring-semantic-success-500 focus:border-semantic-success-500',
  invalid: 'border-semantic-error-500 focus:ring-semantic-error-500 focus:border-semantic-error-500',
  pending: 'border-semantic-warning-500 focus:ring-semantic-warning-500 focus:border-semantic-warning-500',
};

/** Base select classes */
export const selectBaseClasses =
  'block w-full border rounded-md transition-all duration-200 appearance-none cursor-pointer ' +
  'focus:outline-none focus:ring-2 focus:ring-offset-1 ' +
  'disabled:opacity-50 disabled:cursor-not-allowed disabled:bg-neutral-100 ' +
  'pr-10';

/** Dropdown panel classes */
export const dropdownPanelClasses =
  'absolute z-dropdown mt-1 w-full bg-neutral-white border border-neutral-200 ' +
  'rounded-md shadow-lg max-h-60 overflow-auto focus:outline-none';

/** Option classes */
export const optionBaseClasses =
  'relative cursor-pointer select-none py-2 px-3 text-neutral-900';

export const optionHoverClasses = 'hover:bg-brand-primary-50';

export const optionSelectedClasses = 'bg-brand-primary-100 text-brand-primary-900';

export const optionDisabledClasses = 'opacity-50 cursor-not-allowed';

/** Group label classes */
export const groupLabelClasses =
  'px-3 py-2 text-xs font-semibold text-neutral-500 uppercase tracking-wide';

/** Chevron icon classes */
export const chevronClasses =
  'absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-neutral-400';

/** Search input classes */
export const searchInputClasses =
  'w-full px-3 py-2 text-sm border-b border-neutral-200 focus:outline-none focus:ring-0';

/** Helper text classes */
export const selectHelperClasses: Record<ValidationState, string> = {
  default: 'mt-1 text-sm text-neutral-500',
  valid: 'mt-1 text-sm text-semantic-success-600',
  invalid: 'mt-1 text-sm text-semantic-error-600',
  pending: 'mt-1 text-sm text-semantic-warning-600',
};

/** Filter options based on search query */
export function filterOptions<T>(
  options: SelectOption<T>[],
  query: string
): SelectOption<T>[] {
  if (!query) return options;
  const lowerQuery = query.toLowerCase();
  return options.filter((opt) =>
    opt.label.toLowerCase().includes(lowerQuery)
  );
}

/** Get label for selected value */
export function getSelectedLabel<T>(
  value: T | T[] | undefined,
  options: SelectOption<T>[],
  placeholder: string
): string {
  if (value === undefined || value === null) return placeholder;

  if (Array.isArray(value)) {
    if (value.length === 0) return placeholder;
    const labels = value
      .map((v) => options.find((o) => o.value === v)?.label)
      .filter(Boolean);
    return labels.join(', ') || placeholder;
  }

  const option = options.find((o) => o.value === value);
  return option?.label || placeholder;
}
