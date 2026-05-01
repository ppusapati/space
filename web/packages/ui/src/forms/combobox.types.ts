/**
 * Advanced Combobox/Autocomplete component types
 */

import type { Size, ValidationState, FormElementProps } from '../types';

/** Combobox option item */
export interface ComboboxOption<T = unknown> {
  /** Unique value */
  value: T;
  /** Display label */
  label: string;
  /** Optional description text */
  description?: string;
  /** Whether option is disabled */
  disabled?: boolean;
  /** Optional group name */
  group?: string;
  /** Optional icon or image URL */
  icon?: string;
  /** Additional custom data */
  data?: Record<string, unknown>;
}

/** Combobox option group */
export interface ComboboxOptionGroup<T = unknown> {
  /** Group label */
  label: string;
  /** Options in this group */
  options: ComboboxOption<T>[];
}

/** Async load function type */
export type ComboboxLoadFunction<T = unknown> = (
  query: string,
  signal?: AbortSignal
) => Promise<ComboboxOption<T>[]>;

/** Match type for filtering */
export type ComboboxMatchType = 'contains' | 'startsWith' | 'fuzzy';

/** Combobox props */
export interface ComboboxProps<T = unknown> extends FormElementProps {
  /** Current selected value(s) */
  value?: T | T[];
  /** Static options */
  options?: ComboboxOption<T>[];
  /** Grouped options */
  groups?: ComboboxOptionGroup<T>[];
  /** Async load function for remote data */
  loadOptions?: ComboboxLoadFunction<T>;
  /** Debounce delay for async search (ms) */
  debounceMs?: number;
  /** Minimum characters before triggering search */
  minChars?: number;
  /** Placeholder text */
  placeholder?: string;
  /** Size variant */
  size?: Size;
  /** Validation state */
  state?: ValidationState;
  /** Label text */
  label?: string;
  /** Helper text */
  helperText?: string;
  /** Error text */
  errorText?: string;
  /** Allow multiple selection */
  multiple?: boolean;
  /** Allow clearing value */
  clearable?: boolean;
  /** Allow creating new options */
  creatable?: boolean;
  /** Text for create option */
  createText?: string;
  /** Full width */
  fullWidth?: boolean;
  /** Highlight matching text */
  highlightMatches?: boolean;
  /** Match type for filtering */
  matchType?: ComboboxMatchType;
  /** Fields to search in option data */
  searchFields?: string[];
  /** No results text */
  noResultsText?: string;
  /** Loading text */
  loadingText?: string;
  /** Maximum visible items before scrolling */
  maxVisibleItems?: number;
  /** Enable virtual scrolling for large lists */
  virtualScroll?: boolean;
  /** Cache async results */
  cacheResults?: boolean;
}

/** Highlighted text segment */
export interface HighlightSegment {
  text: string;
  isMatch: boolean;
}

/** Combobox base classes */
export const comboboxClasses = {
  container: 'relative w-full',
  inputWrapper: 'relative',
  input: 'block w-full border rounded-lg transition-all duration-200 pr-10 focus:outline-none focus:ring-2 focus:ring-offset-1 disabled:opacity-50 disabled:cursor-not-allowed disabled:bg-neutral-100',
  clearButton: 'absolute inset-y-0 right-8 flex items-center px-1 text-neutral-400 hover:text-neutral-600',
  chevron: 'absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-neutral-400',
  chevronOpen: 'transform rotate-180',
  dropdown: 'absolute z-dropdown mt-1 w-full bg-neutral-white border border-neutral-200 rounded-lg shadow-lg max-h-60 overflow-auto focus:outline-none',
  loadingWrapper: 'flex items-center justify-center py-4',
  noResults: 'px-4 py-3 text-sm text-neutral-500 text-center',
  createOption: 'px-4 py-2 text-sm text-brand-primary-600 hover:bg-brand-primary-50 cursor-pointer flex items-center gap-2',
  optionList: 'py-1',
  groupLabel: 'px-4 py-2 text-xs font-semibold text-neutral-500 uppercase tracking-wide bg-neutral-50',
};

/** Combobox option classes */
export const comboboxOptionClasses = {
  base: 'relative cursor-pointer select-none py-2 px-4 text-neutral-900 flex items-center gap-3',
  hover: 'hover:bg-neutral-50',
  active: 'bg-brand-primary-50',
  selected: 'bg-brand-primary-100 text-brand-primary-900',
  disabled: 'opacity-50 cursor-not-allowed',
  icon: 'w-5 h-5 flex-shrink-0',
  content: 'flex-1 min-w-0',
  label: 'block truncate',
  description: 'block text-xs text-neutral-500 truncate',
  checkmark: 'w-4 h-4 text-brand-primary-600 flex-shrink-0',
  highlight: 'bg-semantic-warning-200 text-semantic-warning-900 rounded px-0.5',
};

/** Combobox size classes */
export const comboboxSizeClasses: Record<Size, string> = {
  xs: 'h-7 px-2 text-xs',
  sm: 'h-8 px-2.5 text-sm',
  md: 'h-10 px-3 text-base',
  lg: 'h-12 px-4 text-lg',
  xl: 'h-14 px-5 text-xl',
};

/** Combobox state classes */
export const comboboxStateClasses: Record<ValidationState, string> = {
  default: 'border-neutral-300 focus:ring-brand-primary-500 focus:border-brand-primary-500',
  valid: 'border-semantic-success-500 focus:ring-semantic-success-500 focus:border-semantic-success-500',
  invalid: 'border-semantic-error-500 focus:ring-semantic-error-500 focus:border-semantic-error-500',
  pending: 'border-semantic-warning-500 focus:ring-semantic-warning-500 focus:border-semantic-warning-500',
};

/** Helper text classes */
export const comboboxHelperClasses: Record<ValidationState, string> = {
  default: 'mt-1 text-sm text-neutral-500',
  valid: 'mt-1 text-sm text-semantic-success-600',
  invalid: 'mt-1 text-sm text-semantic-error-600',
  pending: 'mt-1 text-sm text-semantic-warning-600',
};

/**
 * Debounce function
 */
export function debounce<T extends (...args: unknown[]) => unknown>(
  fn: T,
  delay: number
): (...args: Parameters<T>) => void {
  let timeoutId: ReturnType<typeof setTimeout> | null = null;
  return (...args: Parameters<T>) => {
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
    timeoutId = setTimeout(() => {
      fn(...args);
      timeoutId = null;
    }, delay);
  };
}

/**
 * Highlight matching text in a string
 */
export function highlightMatches(text: string, query: string): HighlightSegment[] {
  if (!query || !text) {
    return [{ text, isMatch: false }];
  }

  const lowerText = text.toLowerCase();
  const lowerQuery = query.toLowerCase();
  const segments: HighlightSegment[] = [];
  let lastIndex = 0;

  let index = lowerText.indexOf(lowerQuery);
  while (index !== -1) {
    if (index > lastIndex) {
      segments.push({ text: text.slice(lastIndex, index), isMatch: false });
    }
    segments.push({ text: text.slice(index, index + query.length), isMatch: true });
    lastIndex = index + query.length;
    index = lowerText.indexOf(lowerQuery, lastIndex);
  }

  if (lastIndex < text.length) {
    segments.push({ text: text.slice(lastIndex), isMatch: false });
  }

  return segments.length > 0 ? segments : [{ text, isMatch: false }];
}

/**
 * Fuzzy match algorithm
 */
export function fuzzyMatch(query: string, text: string): boolean {
  const lowerQuery = query.toLowerCase();
  const lowerText = text.toLowerCase();
  let queryIndex = 0;

  for (let i = 0; i < lowerText.length && queryIndex < lowerQuery.length; i++) {
    if (lowerText[i] === lowerQuery[queryIndex]) {
      queryIndex++;
    }
  }

  return queryIndex === lowerQuery.length;
}

/**
 * Filter options based on query and match type
 */
export function filterComboboxOptions<T>(
  options: ComboboxOption<T>[],
  query: string,
  matchType: ComboboxMatchType = 'contains',
  searchFields?: string[]
): ComboboxOption<T>[] {
  if (!query) return options;

  const lowerQuery = query.toLowerCase();

  return options.filter((option) => {
    // Always search in label
    const labelMatch = matchByType(option.label, lowerQuery, matchType);
    if (labelMatch) return true;

    // Search in description if exists
    if (option.description) {
      const descMatch = matchByType(option.description, lowerQuery, matchType);
      if (descMatch) return true;
    }

    // Search in additional fields if specified
    if (searchFields && option.data) {
      for (const field of searchFields) {
        const value = option.data[field];
        if (typeof value === 'string') {
          const fieldMatch = matchByType(value, lowerQuery, matchType);
          if (fieldMatch) return true;
        }
      }
    }

    return false;
  });
}

function matchByType(text: string, query: string, matchType: ComboboxMatchType): boolean {
  const lowerText = text.toLowerCase();
  switch (matchType) {
    case 'startsWith':
      return lowerText.startsWith(query);
    case 'fuzzy':
      return fuzzyMatch(query, text);
    case 'contains':
    default:
      return lowerText.includes(query);
  }
}

/**
 * Get display label for selected value(s)
 */
export function getComboboxDisplayValue<T>(
  value: T | T[] | undefined,
  options: ComboboxOption<T>[],
  placeholder: string
): string {
  if (value === undefined || value === null) return '';

  if (Array.isArray(value)) {
    if (value.length === 0) return '';
    const labels = value
      .map((v) => options.find((o) => o.value === v)?.label)
      .filter(Boolean);
    return labels.join(', ') || '';
  }

  const option = options.find((o) => o.value === value);
  return option?.label || '';
}
