/**
 * Data Display component types
 */

import type { Size, ColorVariant, BaseProps } from '../types';

/** Card props */
export interface CardProps extends BaseProps {
  /** Card variant */
  variant?: 'elevated' | 'outlined' | 'filled';
  /** Padding size */
  padding?: Size | 'none';
  /** Hoverable state */
  hoverable?: boolean;
  /** Clickable (adds cursor pointer) */
  clickable?: boolean;
}

/** List props */
export interface ListProps extends BaseProps {
  /** List variant */
  variant?: 'simple' | 'divided' | 'bordered';
  /** Size */
  size?: Size;
  /** Ordered list */
  ordered?: boolean;
}

/** ListItem props */
export interface ListItemProps extends BaseProps {
  /** Active/selected state */
  active?: boolean;
  /** Disabled state */
  disabled?: boolean;
  /** Clickable */
  clickable?: boolean;
  /** Leading icon/content slot */
  hasLeading?: boolean;
  /** Trailing icon/content slot */
  hasTrailing?: boolean;
}

/** Tree item data structure */
export interface TreeNode {
  id: string;
  label: string;
  children?: TreeNode[];
  expanded?: boolean;
  selected?: boolean;
  disabled?: boolean;
  icon?: string;
  data?: Record<string, unknown>;
}

/** Tree props */
export interface TreeProps extends BaseProps {
  /** Tree nodes */
  nodes: TreeNode[];
  /** Allow multi-select */
  multiSelect?: boolean;
  /** Checkbox selection */
  checkable?: boolean;
  /** Default expanded keys */
  expandedKeys?: string[];
  /** Default selected keys */
  selectedKeys?: string[];
  /** Show connecting lines */
  showLines?: boolean;
  /** Size */
  size?: Size;
}

/** Badge props */
export interface BadgeProps extends BaseProps {
  /** Badge variant */
  variant?: ColorVariant;
  /** Size */
  size?: Size;
  /** Pill shape */
  pill?: boolean;
  /** Dot only (no content) */
  dot?: boolean;
  /** Badge content */
  content?: string | number;
  /** Max number (shows 99+ if exceeded) */
  max?: number;
  /** Show zero */
  showZero?: boolean;
}

/** Tag props */
export interface TagProps extends BaseProps {
  /** Tag variant */
  variant?: ColorVariant;
  /** Size */
  size?: Size;
  /** Removable */
  removable?: boolean;
  /** Clickable */
  clickable?: boolean;
  /** Disabled */
  disabled?: boolean;
  /** Icon (left side) */
  icon?: string;
}

/** Tooltip props */
export interface TooltipProps extends BaseProps {
  /** Tooltip content */
  content: string;
  /** Position */
  position?: 'top' | 'right' | 'bottom' | 'left';
  /** Trigger */
  trigger?: 'hover' | 'click' | 'focus';
  /** Delay in ms */
  delay?: number;
  /** Disabled */
  disabled?: boolean;
  /** Max width */
  maxWidth?: string;
}

/** Avatar props */
export interface AvatarProps extends BaseProps {
  /** Image source */
  src?: string;
  /** Alt text */
  alt?: string;
  /** Fallback text (name for initials) */
  fallback?: string;
  /** Size */
  size?: Size;
  /** Shape */
  shape?: 'circle' | 'square' | 'rounded';
  /** Status indicator */
  status?: 'online' | 'offline' | 'away' | 'busy';
  /** Show ring around avatar */
  ring?: boolean;
  /** Show badge */
  showBadge?: boolean;
  /** Badge content (number or string) */
  badgeContent?: string | number;
  /** Use dynamic background color based on fallback text (default: true) */
  dynamicColor?: boolean;
}

/** Avatar data for AvatarGroup */
export interface AvatarData {
  /** Image source */
  src?: string;
  /** Name for fallback initials */
  name: string;
  /** Alt text */
  alt?: string;
}

/** AvatarGroup props */
export interface AvatarGroupProps extends BaseProps {
  /** Max avatars to show */
  max?: number;
  /** Size */
  size?: Size;
  /** Stacking direction */
  direction?: 'left' | 'right';
  /** Avatar data array (alternative to using slots) */
  avatars?: AvatarData[];
  /** Shape of avatars */
  shape?: 'circle' | 'square' | 'rounded';
}

/** Card classes */
export const cardClasses = {
  base: 'bg-neutral-white rounded-lg transition-shadow',
  elevated: 'shadow-md',
  outlined: 'border border-neutral-200',
  filled: 'bg-neutral-50',
  hoverable: 'hover:shadow-lg',
  clickable: 'cursor-pointer',
};

export const cardPaddingClasses: Record<Size | 'none', string> = {
  none: '',
  xs: 'p-2',
  sm: 'p-3',
  md: 'p-4',
  lg: 'p-6',
  xl: 'p-8',
};

/** List classes */
export const listClasses = {
  base: 'w-full',
  simple: '',
  divided: 'divide-y divide-neutral-200',
  bordered: 'border border-neutral-200 rounded-lg divide-y divide-neutral-200',
};

export const listSizeClasses: Record<Size, string> = {
  xs: 'text-xs',
  sm: 'text-sm',
  md: 'text-sm',
  lg: 'text-base',
  xl: 'text-lg',
};

/** ListItem classes */
export const listItemClasses = {
  base: 'flex items-center gap-3',
  clickable: 'cursor-pointer hover:bg-neutral-50 transition-colors',
  active: 'bg-brand-primary-50 text-brand-primary-700',
  disabled: 'opacity-50 cursor-not-allowed',
};

export const listItemSizeClasses: Record<Size, string> = {
  xs: 'px-2 py-1',
  sm: 'px-3 py-1.5',
  md: 'px-4 py-2',
  lg: 'px-4 py-3',
  xl: 'px-6 py-4',
};

/** Tree classes */
export const treeClasses = {
  container: 'w-full',
  node: 'select-none',
  nodeContent: 'flex items-center gap-1 py-1 px-2 rounded hover:bg-neutral-100 transition-colors',
  nodeSelected: 'bg-brand-primary-50 text-brand-primary-700 hover:bg-brand-primary-100',
  nodeDisabled: 'opacity-50 cursor-not-allowed hover:bg-transparent',
  expandIcon: 'w-4 h-4 transition-transform shrink-0',
  expandIconExpanded: 'rotate-90',
  checkbox: 'mr-1',
  icon: 'w-4 h-4 shrink-0',
  label: 'truncate',
  children: 'ml-4',
  lines: 'border-l border-neutral-200 ml-2',
};

/** Badge classes */
export const badgeClasses = {
  base: 'inline-flex items-center justify-center font-medium',
  dot: 'w-2 h-2 min-w-0 p-0',
};

export const badgeSizeClasses: Record<Size, string> = {
  xs: 'text-xs px-1.5 py-0.5 min-w-4',
  sm: 'text-xs px-2 py-0.5 min-w-5',
  md: 'text-sm px-2.5 py-0.5 min-w-6',
  lg: 'text-sm px-3 py-1 min-w-7',
  xl: 'text-base px-3.5 py-1 min-w-8',
};

export const badgeVariantClasses: Record<ColorVariant, string> = {
  primary: 'bg-brand-primary-100 text-brand-primary-700',
  secondary: 'bg-brand-secondary-100 text-brand-secondary-700',
  success: 'bg-semantic-success-100 text-semantic-success-700',
  warning: 'bg-semantic-warning-100 text-semantic-warning-700',
  error: 'bg-semantic-error-100 text-semantic-error-700',
  info: 'bg-semantic-info-100 text-semantic-info-700',
  neutral: 'bg-neutral-100 text-neutral-700',
};

export const badgeDotVariantClasses: Record<ColorVariant, string> = {
  primary: 'bg-brand-primary-500',
  secondary: 'bg-brand-secondary-500',
  success: 'bg-semantic-success-500',
  warning: 'bg-semantic-warning-500',
  error: 'bg-semantic-error-500',
  info: 'bg-semantic-info-500',
  neutral: 'bg-neutral-500',
};

/** Tag classes */
export const tagClasses = {
  base: 'inline-flex items-center gap-1 font-medium rounded transition-colors',
  clickable: 'cursor-pointer',
  disabled: 'opacity-50 cursor-not-allowed',
  removeBtn: 'ml-0.5 hover:bg-neutral-black/10 rounded p-0.5',
};

export const tagSizeClasses: Record<Size, { container: string; icon: string; removeIcon: string }> = {
  xs: { container: 'text-xs px-1.5 py-0.5', icon: 'w-3 h-3', removeIcon: 'w-3 h-3' },
  sm: { container: 'text-xs px-2 py-0.5', icon: 'w-3.5 h-3.5', removeIcon: 'w-3 h-3' },
  md: { container: 'text-sm px-2.5 py-1', icon: 'w-4 h-4', removeIcon: 'w-3.5 h-3.5' },
  lg: { container: 'text-sm px-3 py-1.5', icon: 'w-4 h-4', removeIcon: 'w-4 h-4' },
  xl: { container: 'text-base px-4 py-2', icon: 'w-5 h-5', removeIcon: 'w-4 h-4' },
};

export const tagVariantClasses: Record<ColorVariant, { bg: string; hover: string }> = {
  primary: { bg: 'bg-brand-primary-100 text-brand-primary-700', hover: 'hover:bg-brand-primary-200' },
  secondary: { bg: 'bg-brand-secondary-100 text-brand-secondary-700', hover: 'hover:bg-brand-secondary-200' },
  success: { bg: 'bg-semantic-success-100 text-semantic-success-700', hover: 'hover:bg-semantic-success-200' },
  warning: { bg: 'bg-semantic-warning-100 text-semantic-warning-700', hover: 'hover:bg-semantic-warning-200' },
  error: { bg: 'bg-semantic-error-100 text-semantic-error-700', hover: 'hover:bg-semantic-error-200' },
  info: { bg: 'bg-semantic-info-100 text-semantic-info-700', hover: 'hover:bg-semantic-info-200' },
  neutral: { bg: 'bg-neutral-100 text-neutral-700', hover: 'hover:bg-neutral-200' },
};

/** Tooltip classes */
export const tooltipClasses = {
  trigger: 'inline-block',
  content: 'absolute z-tooltip px-2 py-1 text-xs font-medium text-neutral-white bg-neutral-800 rounded shadow-lg whitespace-nowrap',
  arrow: 'absolute w-2 h-2 bg-neutral-800 rotate-45',
};

export const tooltipPositionClasses = {
  top: 'bottom-full left-1/2 -translate-x-1/2 mb-2',
  right: 'left-full top-1/2 -translate-y-1/2 ml-2',
  bottom: 'top-full left-1/2 -translate-x-1/2 mt-2',
  left: 'right-full top-1/2 -translate-y-1/2 mr-2',
};

export const tooltipArrowPositionClasses = {
  top: '-bottom-1 left-1/2 -translate-x-1/2',
  right: '-left-1 top-1/2 -translate-y-1/2',
  bottom: '-top-1 left-1/2 -translate-x-1/2',
  left: '-right-1 top-1/2 -translate-y-1/2',
};

/** Avatar classes */
export const avatarClasses = {
  container: 'relative inline-flex shrink-0',
  image: 'w-full h-full object-cover',
  fallback: 'flex items-center justify-center bg-neutral-200 text-neutral-600 font-medium uppercase',
  fallbackDynamic: 'flex items-center justify-center font-medium uppercase text-neutral-white',
  circle: 'rounded-full',
  square: 'rounded-none',
  rounded: 'rounded-md',
  ring: 'ring-2 ring-neutral-white',
  status: 'absolute bottom-0 right-0 rounded-full border-2 border-neutral-white',
  badge: 'absolute -top-1 -right-1 flex items-center justify-center rounded-full bg-semantic-error-500 text-neutral-white font-medium',
};

export const avatarSizeClasses: Record<Size, { container: string; text: string; status: string; badge: string }> = {
  xs: { container: 'w-6 h-6', text: 'text-xs', status: 'w-2 h-2', badge: 'w-4 h-4 text-[8px]' },
  sm: { container: 'w-8 h-8', text: 'text-xs', status: 'w-2.5 h-2.5', badge: 'w-4 h-4 text-[8px]' },
  md: { container: 'w-10 h-10', text: 'text-sm', status: 'w-3 h-3', badge: 'w-5 h-5 text-[10px]' },
  lg: { container: 'w-12 h-12', text: 'text-base', status: 'w-3.5 h-3.5', badge: 'w-6 h-6 text-xs' },
  xl: { container: 'w-16 h-16', text: 'text-lg', status: 'w-4 h-4', badge: 'w-7 h-7 text-sm' },
};

export const avatarStatusClasses = {
  online: 'bg-semantic-success-500',
  offline: 'bg-neutral-400',
  away: 'bg-semantic-warning-500',
  busy: 'bg-semantic-error-500',
};

/** Avatar dynamic colors for fallback backgrounds */
export const avatarDynamicColors = [
  'bg-brand-primary-500',
  'bg-semantic-success-500',
  'bg-semantic-warning-500',
  'bg-semantic-error-500',
  'bg-semantic-info-500',
  'bg-brand-secondary-500',
  'bg-neutral-500',
];

/** AvatarGroup classes */
export const avatarGroupClasses = {
  container: 'flex items-center',
  left: '-space-x-2',
  right: '-space-x-2 flex-row-reverse',
  overflow: 'flex items-center justify-center bg-neutral-200 text-neutral-600 font-medium rounded-full',
};

/** Picture props */
export interface PictureProps extends BaseProps {
  /** Base path to the image (without extension for multi-format, or with extension for single format) */
  src: string;
  /** Alt text for accessibility */
  alt: string;
  /** Optional width */
  width?: number;
  /** Optional height */
  height?: number;
  /** Loading strategy - lazy by default for performance */
  loading?: 'lazy' | 'eager';
  /** Decoding hint - async by default for non-blocking */
  decoding?: 'async' | 'auto' | 'sync';
  /** Fallback extension if original format not available */
  fallback?: 'jpg' | 'png' | 'webp';
  /** Responsive sizes for srcset */
  sizes?: string;
  /** Widths for responsive images (generates -320w, -640w, etc. variants) */
  widths?: number[];
  /** Object fit style */
  fit?: 'contain' | 'cover' | 'fill' | 'none' | 'scale-down';
  /** Border radius preset */
  rounded?: 'none' | 'sm' | 'md' | 'lg' | 'xl' | 'full';
}

/** Picture classes */
export const pictureClasses = {
  fit: {
    contain: 'object-contain',
    cover: 'object-cover',
    fill: 'object-fill',
    none: 'object-none',
    'scale-down': 'object-scale-down',
  },
  rounded: {
    none: '',
    sm: 'rounded-sm',
    md: 'rounded-md',
    lg: 'rounded-lg',
    xl: 'rounded-xl',
    full: 'rounded-full',
  },
};
