/**
 * Checkbox component types and logic
 */

import type { Size, ColorVariant, FormElementProps } from '../types';

/** Checkbox props interface */
export interface CheckboxProps extends FormElementProps {
  /** Whether checked */
  checked?: boolean;
  /** Indeterminate state (for "select all" patterns) */
  indeterminate?: boolean;
  /** Value when checked (for form submission) */
  value?: string;
  /** Label text */
  label?: string;
  /** Description text */
  description?: string;
  /** Size */
  size?: Size;
  /** Color variant */
  color?: ColorVariant;
  /** Label position */
  labelPosition?: 'left' | 'right';
}

/** UnoCSS class mappings for checkbox sizes */
export const checkboxSizeClasses: Record<Size, { box: string; label: string; icon: string }> = {
  xs: { box: 'w-3 h-3', label: 'text-xs', icon: 'w-2 h-2' },
  sm: { box: 'w-4 h-4', label: 'text-sm', icon: 'w-2.5 h-2.5' },
  md: { box: 'w-5 h-5', label: 'text-base', icon: 'w-3 h-3' },
  lg: { box: 'w-6 h-6', label: 'text-lg', icon: 'w-4 h-4' },
  xl: { box: 'w-7 h-7', label: 'text-xl', icon: 'w-5 h-5' },
};

/** UnoCSS class mappings for color variants */
export const checkboxColorClasses: Record<ColorVariant, { checked: string; focus: string }> = {
  primary: {
    checked: 'bg-brand-primary-500 border-brand-primary-500',
    focus: 'focus:ring-brand-primary-500',
  },
  secondary: {
    checked: 'bg-brand-secondary-500 border-brand-secondary-500',
    focus: 'focus:ring-brand-secondary-500',
  },
  success: {
    checked: 'bg-semantic-success-500 border-semantic-success-500',
    focus: 'focus:ring-semantic-success-500',
  },
  warning: {
    checked: 'bg-semantic-warning-500 border-semantic-warning-500',
    focus: 'focus:ring-semantic-warning-500',
  },
  error: {
    checked: 'bg-semantic-error-500 border-semantic-error-500',
    focus: 'focus:ring-semantic-error-500',
  },
  info: {
    checked: 'bg-semantic-info-500 border-semantic-info-500',
    focus: 'focus:ring-semantic-info-500',
  },
  neutral: {
    checked: 'bg-neutral-600 border-neutral-600',
    focus: 'focus:ring-neutral-500',
  },
};

/** Base checkbox box classes */
export const checkboxBoxBaseClasses =
  'shrink-0 border-2 border-neutral-300 rounded ' +
  'transition-all duration-200 ' +
  'focus:outline-none focus:ring-2 focus:ring-offset-2 ' +
  'disabled:opacity-50 disabled:cursor-not-allowed ' +
  'cursor-pointer';

/** Unchecked box classes */
export const checkboxUncheckedClasses = 'bg-neutral-white';

/** Checkbox container classes */
export const checkboxContainerClasses = 'inline-flex items-start gap-2';

/** Label classes */
export const checkboxLabelClasses = 'text-neutral-700 cursor-pointer select-none';

/** Disabled label classes */
export const checkboxLabelDisabledClasses = 'text-neutral-400 cursor-not-allowed';

/** Description classes */
export const checkboxDescriptionClasses = 'text-sm text-neutral-500 mt-0.5';
