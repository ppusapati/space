/**
 * FormField and FormSection component types
 */

import type { ValidationState, BaseProps } from '../types';

/** FormField props interface */
export interface FormFieldProps extends BaseProps {
  /** Field label */
  label?: string;
  /** Label position */
  labelPosition?: 'top' | 'left' | 'right';
  /** Label width (for side-positioned labels) */
  labelWidth?: string;
  /** Helper text */
  helperText?: string;
  /** Error message */
  errorText?: string;
  /** Validation state */
  state?: ValidationState;
  /** Whether field is required */
  required?: boolean;
  /** Whether to show optional badge */
  showOptional?: boolean;
  /** Full width */
  fullWidth?: boolean;
}

/** FormSection props interface */
export interface FormSectionProps extends BaseProps {
  /** Section title */
  title?: string;
  /** Section description */
  description?: string;
  /** Collapsible */
  collapsible?: boolean;
  /** Initial collapsed state */
  collapsed?: boolean;
  /** Divider style */
  divider?: 'none' | 'top' | 'bottom' | 'both';
  /** Columns layout */
  columns?: 1 | 2 | 3 | 4;
  /** Gap between fields */
  gap?: 'none' | 'sm' | 'md' | 'lg';
}

/** FormField classes */
export const formFieldClasses = {
  container: 'w-full',
  labelTop: 'flex flex-col gap-1',
  labelLeft: 'flex items-start gap-4',
  labelRight: 'flex items-start gap-4 flex-row-reverse',
  label: 'block text-sm font-medium text-neutral-700',
  labelRequired: 'text-semantic-error-500 ml-0.5',
  labelOptional: 'text-xs font-normal text-neutral-400 ml-1',
  content: 'flex-1',
  helper: {
    default: 'mt-1 text-sm text-neutral-500',
    valid: 'mt-1 text-sm text-semantic-success-600',
    invalid: 'mt-1 text-sm text-semantic-error-600',
    pending: 'mt-1 text-sm text-semantic-warning-600',
  },
};

/** FormSection classes */
export const formSectionClasses = {
  container: 'w-full',
  header: 'mb-4',
  headerCollapsible: 'flex items-center justify-between cursor-pointer',
  title: 'text-lg font-semibold text-neutral-900',
  description: 'mt-1 text-sm text-neutral-500',
  content: 'space-y-4',
  dividerTop: 'pt-6 border-t border-neutral-200',
  dividerBottom: 'pb-6 border-b border-neutral-200',
  chevron: 'w-5 h-5 text-neutral-500 transition-transform duration-200',
  chevronCollapsed: 'rotate-180',
  columns: {
    1: 'grid grid-cols-1',
    2: 'grid grid-cols-1 md:grid-cols-2',
    3: 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3',
    4: 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4',
  },
  gap: {
    none: 'gap-0',
    sm: 'gap-2',
    md: 'gap-4',
    lg: 'gap-6',
  },
};
