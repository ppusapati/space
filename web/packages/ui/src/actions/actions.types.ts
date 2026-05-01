/**
 * Action component types
 */

import type { Size, BaseProps, DisableableProps, LoadableProps } from '../types';

/** Button variant types */
export type ButtonVariant = 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger' | 'success' | 'warning';

/** Button props */
export interface ButtonProps extends BaseProps, DisableableProps, LoadableProps {
  /** Button variant */
  variant?: ButtonVariant;
  /** Size */
  size?: Size;
  /** Button type */
  type?: 'button' | 'submit' | 'reset';
  /** Full width */
  fullWidth?: boolean;
  /** Icon only (square button) */
  iconOnly?: boolean;
  /** Href for link buttons */
  href?: string;
  /** Target for link buttons */
  target?: '_blank' | '_self' | '_parent' | '_top';
  /** Aria label for accessibility */
  ariaLabel?: string;
}

/** ButtonGroup props */
export interface ButtonGroupProps extends BaseProps {
  /** Size (applies to all buttons) */
  size?: Size;
  /** Variant (applies to all buttons) */
  variant?: ButtonVariant;
  /** Vertical orientation */
  vertical?: boolean;
  /** Attached buttons (no gaps) */
  attached?: boolean;
}

/** CloseButton props */
export interface CloseButtonProps extends BaseProps, DisableableProps {
  /** Size */
  size?: Size;
  /** Aria label */
  ariaLabel?: string;
}

/** Collapse props */
export interface CollapseProps extends BaseProps {
  /** Whether the content is visible */
  open?: boolean;
  /** Animation duration in ms */
  duration?: number;
}

/** AccordionItemData data structure (for data-driven accordion) */
export interface AccordionItemData {
  id: string;
  title: string;
  content?: string;
  disabled?: boolean;
}

/** AccordionItem props (for slot-based accordion) */
export interface AccordionItemProps extends BaseProps, DisableableProps {
  /** Item ID */
  id?: string;
  /** Item title */
  title: string;
}

/** Accordion props */
export interface AccordionProps extends BaseProps {
  /** Accordion items (for data-driven mode) */
  items?: AccordionItemData[];
  /** Size */
  size?: Size;
  /** Allow multiple items open */
  multiple?: boolean;
  /** Default open item IDs */
  defaultOpen?: string[];
  /** Flush style (no outer border/radius) */
  flush?: boolean;
}

/** ScrollspyItem data structure */
export interface ScrollspyItem {
  id: string;
  label: string;
  href: string;
}

/** Scrollspy props */
export interface ScrollspyProps extends BaseProps {
  /** Items to track */
  items: ScrollspyItem[];
  /** Offset from top in pixels */
  offset?: number;
  /** Scroll container selector (defaults to window) */
  container?: string;
}

// ============================================================================
// Button Classes
// ============================================================================

export const buttonBaseClasses =
  'inline-flex items-center justify-center gap-2 focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed relative transition-all duration-200 ease-out font-medium border rounded-md';

export const buttonSizeClasses: Record<Size, string> = {
  xs: 'px-2 py-1 text-xs h-6',
  sm: 'px-3 py-1.5 text-sm h-8',
  md: 'px-4 py-2 text-sm h-10',
  lg: 'px-6 py-3 text-base h-12',
  xl: 'px-8 py-4 text-lg h-14',
};

export const buttonIconOnlySizeClasses: Record<Size, string> = {
  xs: 'p-1 h-6 w-6',
  sm: 'p-1.5 h-8 w-8',
  md: 'p-2 h-10 w-10',
  lg: 'p-3 h-12 w-12',
  xl: 'p-4 h-14 w-14',
};

export const buttonVariantClasses: Record<ButtonVariant, string> = {
  primary:
    'bg-brand-primary-500 text-neutral-white border-brand-primary-500 shadow-sm hover:bg-brand-primary-600 hover:border-brand-primary-600 hover:-translate-y-0.5 active:bg-brand-primary-700 active:border-brand-primary-700 active:translate-y-0 focus:ring-brand-primary-500',
  secondary:
    'bg-neutral-white text-brand-primary-500 border-neutral-300 shadow-sm hover:bg-neutral-50 hover:border-brand-primary-300 hover:-translate-y-0.5 focus:ring-brand-primary-500',
  outline:
    'bg-transparent text-brand-primary-500 border-2 border-brand-primary-500 hover:bg-brand-primary-50 hover:border-brand-primary-600 active:bg-brand-primary-100 active:border-brand-primary-700 focus:ring-brand-primary-500',
  ghost:
    'bg-transparent text-brand-primary-500 border-transparent hover:bg-neutral-100 focus:ring-brand-primary-500',
  danger:
    'bg-semantic-error-500 text-neutral-white border-semantic-error-500 shadow-sm hover:bg-semantic-error-600 hover:border-semantic-error-600 active:bg-semantic-error-700 active:border-semantic-error-700 focus:ring-semantic-error-500',
  success:
    'bg-semantic-success-500 text-neutral-white border-semantic-success-500 shadow-sm hover:bg-semantic-success-600 hover:border-semantic-success-600 active:bg-semantic-success-700 active:border-semantic-success-700 focus:ring-semantic-success-500',
  warning:
    'bg-semantic-warning-500 text-neutral-900 border-semantic-warning-500 shadow-sm hover:bg-semantic-warning-600 hover:border-semantic-warning-600 active:bg-semantic-warning-700 active:border-semantic-warning-700 focus:ring-semantic-warning-500',
};

export const buttonLoadingSpinnerClasses = 'animate-spin h-4 w-4 border-2 border-current border-t-transparent rounded-full';

// ============================================================================
// ButtonGroup Classes
// ============================================================================

export const buttonGroupClasses = {
  base: 'inline-flex',
  horizontal: 'flex-row',
  vertical: 'flex-col',
  gap: 'gap-1',
  attached: '[&>*:not(:first-child):not(:last-child)]:rounded-none',
  attachedHorizontal: '[&>*:first-child]:rounded-r-none [&>*:last-child]:rounded-l-none',
  attachedVertical: '[&>*:first-child]:rounded-b-none [&>*:last-child]:rounded-t-none',
};

// ============================================================================
// CloseButton Classes
// ============================================================================

export const closeButtonClasses = {
  base: 'inline-flex items-center justify-center rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-brand-primary-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed',
  hover: 'hover:bg-neutral-100 text-neutral-500 hover:text-neutral-700',
};

export const closeButtonSizeClasses: Record<Size, { button: string; icon: string }> = {
  xs: { button: 'h-5 w-5', icon: 'h-3 w-3' },
  sm: { button: 'h-6 w-6', icon: 'h-3.5 w-3.5' },
  md: { button: 'h-8 w-8', icon: 'h-4 w-4' },
  lg: { button: 'h-10 w-10', icon: 'h-5 w-5' },
  xl: { button: 'h-12 w-12', icon: 'h-6 w-6' },
};

// ============================================================================
// Collapse Classes
// ============================================================================

export const collapseClasses = {
  container: 'overflow-hidden transition-all duration-200 ease-in-out',
  content: 'w-full',
};

// ============================================================================
// Accordion Classes
// ============================================================================

export const accordionClasses = {
  container: 'border border-neutral-200 rounded-md divide-y divide-neutral-200 overflow-hidden',
  containerFlush: 'divide-y divide-neutral-200',
};

export const accordionItemClasses = {
  header:
    'w-full text-left flex items-center justify-between transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-brand-primary-500 focus:ring-inset font-medium',
  headerClosed: 'bg-neutral-white text-neutral-900 hover:bg-neutral-50',
  headerOpen: 'bg-brand-primary-50 text-brand-primary-900 hover:bg-brand-primary-100',
  headerDisabled: 'cursor-not-allowed opacity-50 bg-neutral-50 text-neutral-400',
  content: 'overflow-hidden transition-all duration-200 ease-in-out text-neutral-600',
  chevron: 'w-5 h-5 transition-transform duration-200 transform shrink-0',
  chevronOpen: 'rotate-180',
};

export const accordionSizeClasses: Record<Size, { header: string; content: string }> = {
  xs: { header: 'px-3 py-2 text-xs', content: 'px-3 pb-2 text-xs' },
  sm: { header: 'px-4 py-3 text-sm', content: 'px-4 pb-3 text-sm' },
  md: { header: 'px-5 py-4 text-base', content: 'px-5 pb-4 text-base' },
  lg: { header: 'px-6 py-5 text-lg', content: 'px-6 pb-5 text-lg' },
  xl: { header: 'px-8 py-6 text-xl', content: 'px-8 pb-6 text-xl' },
};

// ============================================================================
// Scrollspy Classes
// ============================================================================

export const scrollspyClasses = {
  nav: 'flex flex-col gap-1',
  item: 'px-3 py-2 text-sm rounded-md transition-colors text-neutral-600 hover:text-neutral-900 hover:bg-neutral-100',
  itemActive: 'bg-brand-primary-50 text-brand-primary-700 font-medium',
};
