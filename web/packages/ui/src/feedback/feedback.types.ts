/**
 * Feedback component types
 */

import type { Size, ColorVariant, BaseProps, Position } from '../types';

/** Modal props */
export interface ModalProps extends BaseProps {
  /** Whether modal is open */
  open: boolean;
  /** Modal title */
  title?: string;
  /** Modal size */
  size?: Size | 'full';
  /** Close on backdrop click */
  closeOnBackdrop?: boolean;
  /** Close on escape key */
  closeOnEscape?: boolean;
  /** Show close button */
  showClose?: boolean;
  /** Center vertically */
  centered?: boolean;
  /** Prevent body scroll when open */
  preventScroll?: boolean;
}

/** Dialog props (extends Modal for confirmations) */
export interface DialogProps extends ModalProps {
  /** Dialog variant */
  variant?: 'info' | 'warning' | 'error' | 'success' | 'confirm';
  /** Confirm button text */
  confirmText?: string;
  /** Cancel button text */
  cancelText?: string;
  /** Loading state */
  loading?: boolean;
  /** Destructive action */
  destructive?: boolean;
}

/** Drawer props */
export interface DrawerProps extends BaseProps {
  /** Whether drawer is open */
  open: boolean;
  /** Drawer title */
  title?: string;
  /** Position */
  position?: 'left' | 'right' | 'top' | 'bottom';
  /** Drawer size */
  size?: Size | 'full';
  /** Close on backdrop click */
  closeOnBackdrop?: boolean;
  /** Close on escape key */
  closeOnEscape?: boolean;
  /** Show close button */
  showClose?: boolean;
  /** Show overlay/backdrop */
  overlay?: boolean;
}

/** Toast props */
export interface ToastProps extends BaseProps {
  /** Toast variant */
  variant?: ColorVariant;
  /** Toast title */
  title?: string;
  /** Toast message */
  message: string;
  /** Duration in ms (0 for persistent) */
  duration?: number;
  /** Show close button */
  dismissible?: boolean;
  /** Position */
  position?: 'top-left' | 'top-center' | 'top-right' | 'bottom-left' | 'bottom-center' | 'bottom-right';
  /** Action button */
  action?: {
    label: string;
    onClick: () => void;
  };
}

/** Alert props */
export interface AlertProps extends BaseProps {
  /** Alert variant */
  variant?: ColorVariant;
  /** Alert title */
  title?: string;
  /** Show icon */
  showIcon?: boolean;
  /** Dismissible */
  dismissible?: boolean;
  /** Size */
  size?: Size;
}

/** Notification props (similar to Toast but for system notifications) */
export interface NotificationProps extends BaseProps {
  /** Notification variant */
  variant?: ColorVariant;
  /** Title */
  title: string;
  /** Description */
  description?: string;
  /** Avatar/Icon */
  avatar?: string;
  /** Timestamp */
  timestamp?: string;
  /** Read state */
  read?: boolean;
  /** Dismissible */
  dismissible?: boolean;
}

/** Loader/Spinner props */
export interface SpinnerProps extends BaseProps {
  /** Size */
  size?: Size;
  /** Color variant */
  variant?: ColorVariant;
  /** Loading text */
  label?: string;
}

/** Skeleton props */
export interface SkeletonProps extends BaseProps {
  /** Skeleton variant */
  variant?: 'text' | 'circular' | 'rectangular' | 'rounded';
  /** Width */
  width?: string;
  /** Height */
  height?: string;
  /** Animation */
  animation?: 'pulse' | 'wave' | 'none';
  /** Number of lines (for text variant) */
  lines?: number;
}

/** Modal classes */
export const modalClasses = {
  overlay: 'fixed inset-0 z-modal bg-neutral-black/50 transition-opacity',
  container: 'fixed inset-0 z-modal overflow-y-auto',
  wrapper: 'flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0',
  wrapperCentered: 'items-center',
  panel: 'relative transform overflow-hidden rounded-lg bg-neutral-white text-left shadow-xl transition-all sm:my-8 w-full',
  header: 'flex items-center justify-between px-6 py-4 border-b border-neutral-200',
  title: 'text-lg font-semibold text-neutral-900',
  closeBtn: 'p-1 rounded-md text-neutral-400 hover:text-neutral-600 hover:bg-neutral-100 transition-colors',
  body: 'px-6 py-4',
  footer: 'flex items-center justify-end gap-3 px-6 py-4 border-t border-neutral-200',
};

export const modalSizeClasses: Record<Size | 'full', string> = {
  xs: 'sm:max-w-xs',
  sm: 'sm:max-w-sm',
  md: 'sm:max-w-md',
  lg: 'sm:max-w-lg',
  xl: 'sm:max-w-xl',
  full: 'sm:max-w-full sm:m-4',
};

/** Drawer classes */
export const drawerClasses = {
  overlay: 'fixed inset-0 z-drawer bg-neutral-black/50 transition-opacity',
  panel: 'fixed z-drawer bg-neutral-white shadow-xl transition-transform',
  left: 'inset-y-0 left-0',
  right: 'inset-y-0 right-0',
  top: 'inset-x-0 top-0',
  bottom: 'inset-x-0 bottom-0',
  header: 'flex items-center justify-between px-4 py-3 border-b border-neutral-200',
  title: 'text-lg font-semibold text-neutral-900',
  closeBtn: 'p-1 rounded-md text-neutral-400 hover:text-neutral-600 hover:bg-neutral-100',
  body: 'flex-1 overflow-y-auto p-4',
  footer: 'px-4 py-3 border-t border-neutral-200',
};

export const drawerSizeClasses = {
  left: { xs: 'w-64', sm: 'w-80', md: 'w-96', lg: 'w-[480px]', xl: 'w-[640px]', full: 'w-full' },
  right: { xs: 'w-64', sm: 'w-80', md: 'w-96', lg: 'w-[480px]', xl: 'w-[640px]', full: 'w-full' },
  top: { xs: 'h-32', sm: 'h-48', md: 'h-64', lg: 'h-96', xl: 'h-[480px]', full: 'h-full' },
  bottom: { xs: 'h-32', sm: 'h-48', md: 'h-64', lg: 'h-96', xl: 'h-[480px]', full: 'h-full' },
};

/** Toast classes */
export const toastClasses = {
  container: 'fixed z-toast pointer-events-none',
  toast: 'pointer-events-auto flex items-start gap-3 p-4 rounded-lg shadow-lg border transition-all',
  icon: 'w-5 h-5 shrink-0 mt-0.5',
  content: 'flex-1 min-w-0',
  title: 'text-sm font-semibold',
  message: 'text-sm',
  closeBtn: 'shrink-0 p-1 rounded hover:bg-neutral-black/10',
  action: 'text-sm font-medium underline hover:no-underline',
};

export const toastVariantClasses: Record<ColorVariant, { bg: string; border: string; icon: string; title: string; message: string }> = {
  primary: { bg: 'bg-brand-primary-50', border: 'border-brand-primary-200', icon: 'text-brand-primary-500', title: 'text-brand-primary-900', message: 'text-brand-primary-700' },
  secondary: { bg: 'bg-brand-secondary-50', border: 'border-brand-secondary-200', icon: 'text-brand-secondary-500', title: 'text-brand-secondary-900', message: 'text-brand-secondary-700' },
  success: { bg: 'bg-semantic-success-50', border: 'border-semantic-success-200', icon: 'text-semantic-success-500', title: 'text-semantic-success-900', message: 'text-semantic-success-700' },
  warning: { bg: 'bg-semantic-warning-50', border: 'border-semantic-warning-200', icon: 'text-semantic-warning-500', title: 'text-semantic-warning-900', message: 'text-semantic-warning-700' },
  error: { bg: 'bg-semantic-error-50', border: 'border-semantic-error-200', icon: 'text-semantic-error-500', title: 'text-semantic-error-900', message: 'text-semantic-error-700' },
  info: { bg: 'bg-semantic-info-50', border: 'border-semantic-info-200', icon: 'text-semantic-info-500', title: 'text-semantic-info-900', message: 'text-semantic-info-700' },
  neutral: { bg: 'bg-neutral-50', border: 'border-neutral-200', icon: 'text-neutral-500', title: 'text-neutral-900', message: 'text-neutral-700' },
};

export const toastPositionClasses = {
  'top-left': 'top-4 left-4',
  'top-center': 'top-4 left-1/2 -translate-x-1/2',
  'top-right': 'top-4 right-4',
  'bottom-left': 'bottom-4 left-4',
  'bottom-center': 'bottom-4 left-1/2 -translate-x-1/2',
  'bottom-right': 'bottom-4 right-4',
};

/** Alert classes */
export const alertClasses = {
  container: 'flex gap-3 rounded-lg border p-4',
  icon: 'w-5 h-5 shrink-0 mt-0.5',
  content: 'flex-1 min-w-0',
  title: 'font-semibold',
  message: 'mt-1',
  closeBtn: 'shrink-0 p-1 rounded hover:bg-neutral-black/10',
};

export const alertVariantClasses: Record<ColorVariant, { bg: string; border: string; icon: string; title: string; message: string }> = {
  primary: { bg: 'bg-brand-primary-50', border: 'border-brand-primary-200', icon: 'text-brand-primary-500', title: 'text-brand-primary-800', message: 'text-brand-primary-700' },
  secondary: { bg: 'bg-brand-secondary-50', border: 'border-brand-secondary-200', icon: 'text-brand-secondary-500', title: 'text-brand-secondary-800', message: 'text-brand-secondary-700' },
  success: { bg: 'bg-semantic-success-50', border: 'border-semantic-success-200', icon: 'text-semantic-success-500', title: 'text-semantic-success-800', message: 'text-semantic-success-700' },
  warning: { bg: 'bg-semantic-warning-50', border: 'border-semantic-warning-200', icon: 'text-semantic-warning-500', title: 'text-semantic-warning-800', message: 'text-semantic-warning-700' },
  error: { bg: 'bg-semantic-error-50', border: 'border-semantic-error-200', icon: 'text-semantic-error-500', title: 'text-semantic-error-800', message: 'text-semantic-error-700' },
  info: { bg: 'bg-semantic-info-50', border: 'border-semantic-info-200', icon: 'text-semantic-info-500', title: 'text-semantic-info-800', message: 'text-semantic-info-700' },
  neutral: { bg: 'bg-neutral-50', border: 'border-neutral-200', icon: 'text-neutral-500', title: 'text-neutral-800', message: 'text-neutral-700' },
};

export const alertSizeClasses: Record<Size, string> = {
  xs: 'text-xs p-2',
  sm: 'text-sm p-3',
  md: 'text-sm p-4',
  lg: 'text-base p-5',
  xl: 'text-lg p-6',
};

/** Spinner classes */
export const spinnerClasses = {
  container: 'inline-flex items-center gap-2',
  spinner: 'animate-spin',
  label: 'text-neutral-600',
};

export const spinnerSizeClasses: Record<Size, { spinner: string; label: string }> = {
  xs: { spinner: 'w-3 h-3', label: 'text-xs' },
  sm: { spinner: 'w-4 h-4', label: 'text-sm' },
  md: { spinner: 'w-6 h-6', label: 'text-sm' },
  lg: { spinner: 'w-8 h-8', label: 'text-base' },
  xl: { spinner: 'w-12 h-12', label: 'text-lg' },
};

export const spinnerVariantClasses: Record<ColorVariant, string> = {
  primary: 'text-brand-primary-500',
  secondary: 'text-brand-secondary-500',
  success: 'text-semantic-success-500',
  warning: 'text-semantic-warning-500',
  error: 'text-semantic-error-500',
  info: 'text-semantic-info-500',
  neutral: 'text-neutral-500',
};

/** Skeleton classes */
export const skeletonClasses = {
  base: 'bg-neutral-200',
  pulse: 'animate-pulse',
  wave: 'overflow-hidden relative before:absolute before:inset-0 before:-translate-x-full before:animate-shimmer before:bg-gradient-to-r before:from-transparent before:via-neutral-white/60 before:to-transparent',
  text: 'rounded h-4',
  circular: 'rounded-full',
  rectangular: '',
  rounded: 'rounded-md',
};
