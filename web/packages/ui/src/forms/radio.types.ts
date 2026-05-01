/**
 * Radio component types and logic
 */

import type { Size, ColorVariant, FormElementProps } from '../types';

/** Radio option */
export interface RadioOption<T = string> {
  value: T;
  label: string;
  description?: string;
  disabled?: boolean;
}

/** Radio props interface */
export interface RadioProps extends FormElementProps {
  /** Current selected value */
  value?: string;
  /** Radio options */
  options?: RadioOption[];
  /** Group name */
  name: string;
  /** Size */
  size?: Size;
  /** Color variant */
  color?: ColorVariant;
  /** Layout direction */
  orientation?: 'horizontal' | 'vertical';
  /** Label position relative to radio */
  labelPosition?: 'left' | 'right';
}

/** Single Radio Button props */
export interface RadioButtonProps extends FormElementProps {
  /** Value for this radio */
  value: string;
  /** Currently selected value in the group */
  groupValue?: string;
  /** Label text */
  label?: string;
  /** Description text */
  description?: string;
  /** Group name */
  name: string;
  /** Size */
  size?: Size;
  /** Color variant */
  color?: ColorVariant;
  /** Label position */
  labelPosition?: 'left' | 'right';
}

/** UnoCSS class mappings for radio sizes */
export const radioSizeClasses: Record<Size, { radio: string; dot: string; label: string }> = {
  xs: { radio: 'w-3 h-3', dot: 'w-1.5 h-1.5', label: 'text-xs' },
  sm: { radio: 'w-4 h-4', dot: 'w-2 h-2', label: 'text-sm' },
  md: { radio: 'w-5 h-5', dot: 'w-2.5 h-2.5', label: 'text-base' },
  lg: { radio: 'w-6 h-6', dot: 'w-3 h-3', label: 'text-lg' },
  xl: { radio: 'w-7 h-7', dot: 'w-3.5 h-3.5', label: 'text-xl' },
};

/** UnoCSS class mappings for color variants */
export const radioColorClasses: Record<ColorVariant, { selected: string; dot: string; focus: string }> = {
  primary: {
    selected: 'border-brand-primary-500',
    dot: 'bg-brand-primary-500',
    focus: 'focus:ring-brand-primary-500',
  },
  secondary: {
    selected: 'border-brand-secondary-500',
    dot: 'bg-brand-secondary-500',
    focus: 'focus:ring-brand-secondary-500',
  },
  success: {
    selected: 'border-semantic-success-500',
    dot: 'bg-semantic-success-500',
    focus: 'focus:ring-semantic-success-500',
  },
  warning: {
    selected: 'border-semantic-warning-500',
    dot: 'bg-semantic-warning-500',
    focus: 'focus:ring-semantic-warning-500',
  },
  error: {
    selected: 'border-semantic-error-500',
    dot: 'bg-semantic-error-500',
    focus: 'focus:ring-semantic-error-500',
  },
  info: {
    selected: 'border-semantic-info-500',
    dot: 'bg-semantic-info-500',
    focus: 'focus:ring-semantic-info-500',
  },
  neutral: {
    selected: 'border-neutral-600',
    dot: 'bg-neutral-600',
    focus: 'focus:ring-neutral-500',
  },
};

/** Base radio classes */
export const radioBaseClasses =
  'shrink-0 rounded-full border-2 border-neutral-300 bg-neutral-white ' +
  'transition-all duration-200 cursor-pointer ' +
  'focus:outline-none focus:ring-2 focus:ring-offset-2 ' +
  'disabled:opacity-50 disabled:cursor-not-allowed ' +
  'flex items-center justify-center';

/** Radio group container classes */
export const radioGroupClasses = {
  horizontal: 'flex flex-wrap gap-4',
  vertical: 'flex flex-col gap-2',
};

/** Radio item container classes */
export const radioItemClasses = 'inline-flex items-start gap-2';

/** Label classes */
export const radioLabelClasses = 'text-neutral-700 cursor-pointer select-none';

/** Disabled label classes */
export const radioLabelDisabledClasses = 'text-neutral-400 cursor-not-allowed';

/** Description classes */
export const radioDescriptionClasses = 'text-sm text-neutral-500 mt-0.5';
