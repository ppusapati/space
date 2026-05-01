/**
 * Layout component types
 */

import type { Size, BaseProps } from '../types';

/** Container props */
export interface ContainerProps extends BaseProps {
  /** Max width */
  maxWidth?: Size | 'full' | 'prose';
  /** Center content */
  centered?: boolean;
  /** Padding */
  padding?: Size | 'none';
}

/** Grid props */
export interface GridProps extends BaseProps {
  /** Number of columns */
  cols?: 1 | 2 | 3 | 4 | 5 | 6 | 12;
  /** Gap between items */
  gap?: Size | 'none';
  /** Responsive columns */
  responsive?: boolean;
}

/** Flex props */
export interface FlexProps extends BaseProps {
  /** Direction */
  direction?: 'row' | 'row-reverse' | 'col' | 'col-reverse';
  /** Justify content */
  justify?: 'start' | 'end' | 'center' | 'between' | 'around' | 'evenly';
  /** Align items */
  align?: 'start' | 'end' | 'center' | 'baseline' | 'stretch';
  /** Wrap */
  wrap?: 'nowrap' | 'wrap' | 'wrap-reverse';
  /** Gap */
  gap?: Size | 'none';
  /** Inline flex */
  inline?: boolean;
}

/** Stack props */
export interface StackProps extends BaseProps {
  /** Direction */
  direction?: 'horizontal' | 'vertical';
  /** Gap */
  gap?: Size | 'none';
  /** Align items */
  align?: 'start' | 'end' | 'center' | 'baseline' | 'stretch';
  /** Justify content */
  justify?: 'start' | 'end' | 'center' | 'between' | 'around';
  /** Wrap items */
  wrap?: boolean;
}

/** Spacer props */
export interface SpacerProps extends BaseProps {
  /** Size */
  size?: Size;
  /** Axis */
  axis?: 'horizontal' | 'vertical' | 'both';
}

/** Divider props */
export interface DividerProps extends BaseProps {
  /** Orientation */
  orientation?: 'horizontal' | 'vertical';
  /** Variant */
  variant?: 'solid' | 'dashed' | 'dotted';
  /** Color */
  color?: 'light' | 'medium' | 'dark';
  /** With text label */
  label?: string;
  /** Label position */
  labelPosition?: 'start' | 'center' | 'end';
}

/** AppShell props */
export interface AppShellProps extends BaseProps {
  /** Fixed header */
  fixedHeader?: boolean;
  /** Fixed sidebar */
  fixedSidebar?: boolean;
  /** Sidebar collapsed */
  sidebarCollapsed?: boolean;
  /** Sidebar position */
  sidebarPosition?: 'left' | 'right';
  /** Header height */
  headerHeight?: string;
  /** Sidebar width */
  sidebarWidth?: string;
  /** Collapsed sidebar width */
  collapsedWidth?: string;
}

/** PageLayout props */
export interface PageLayoutProps extends BaseProps {
  /** Layout type */
  layout?: 'default' | 'centered' | 'sidebar' | 'split';
  /** Max width */
  maxWidth?: Size | 'full';
  /** Padding */
  padding?: Size | 'none';
}

/** Container classes */
export const containerClasses = {
  base: 'w-full',
  centered: 'mx-auto',
};

export const containerMaxWidthClasses: Record<Size | 'full' | 'prose', string> = {
  xs: 'max-w-xs',
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
  xl: 'max-w-xl',
  full: 'max-w-full',
  prose: 'max-w-prose',
};

export const containerPaddingClasses: Record<Size | 'none', string> = {
  none: '',
  xs: 'px-2',
  sm: 'px-4',
  md: 'px-6',
  lg: 'px-8',
  xl: 'px-12',
};

/** Grid classes */
export const gridClasses = {
  base: 'grid',
};

export const gridColsClasses: Record<number, string> = {
  1: 'grid-cols-1',
  2: 'grid-cols-2',
  3: 'grid-cols-3',
  4: 'grid-cols-4',
  5: 'grid-cols-5',
  6: 'grid-cols-6',
  12: 'grid-cols-12',
};

export const gridResponsiveClasses: Record<number, string> = {
  1: 'grid-cols-1',
  2: 'grid-cols-1 sm:grid-cols-2',
  3: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3',
  4: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-4',
  5: 'grid-cols-2 sm:grid-cols-3 lg:grid-cols-5',
  6: 'grid-cols-2 sm:grid-cols-3 lg:grid-cols-6',
  12: 'grid-cols-4 sm:grid-cols-6 lg:grid-cols-12',
};

export const gapClasses: Record<Size | 'none', string> = {
  none: 'gap-0',
  xs: 'gap-1',
  sm: 'gap-2',
  md: 'gap-4',
  lg: 'gap-6',
  xl: 'gap-8',
};

/** Flex classes */
export const flexClasses = {
  base: 'flex',
  inline: 'inline-flex',
};

export const flexDirectionClasses = {
  row: 'flex-row',
  'row-reverse': 'flex-row-reverse',
  col: 'flex-col',
  'col-reverse': 'flex-col-reverse',
};

export const flexJustifyClasses = {
  start: 'justify-start',
  end: 'justify-end',
  center: 'justify-center',
  between: 'justify-between',
  around: 'justify-around',
  evenly: 'justify-evenly',
};

export const flexAlignClasses = {
  start: 'items-start',
  end: 'items-end',
  center: 'items-center',
  baseline: 'items-baseline',
  stretch: 'items-stretch',
};

export const flexWrapClasses = {
  nowrap: 'flex-nowrap',
  wrap: 'flex-wrap',
  'wrap-reverse': 'flex-wrap-reverse',
};

/** Stack classes */
export const stackClasses = {
  base: 'flex',
  horizontal: 'flex-row',
  vertical: 'flex-col',
};

/** Spacer classes */
export const spacerClasses: Record<Size, { x: string; y: string }> = {
  xs: { x: 'w-1', y: 'h-1' },
  sm: { x: 'w-2', y: 'h-2' },
  md: { x: 'w-4', y: 'h-4' },
  lg: { x: 'w-6', y: 'h-6' },
  xl: { x: 'w-8', y: 'h-8' },
};

/** Divider classes */
export const dividerClasses = {
  base: 'shrink-0',
  horizontal: 'w-full h-px',
  vertical: 'h-full w-px',
  solid: 'border-solid',
  dashed: 'border-dashed',
  dotted: 'border-dotted',
  light: 'bg-neutral-100',
  medium: 'bg-neutral-200',
  dark: 'bg-neutral-300',
  withLabel: 'flex items-center',
  label: 'px-3 text-sm text-neutral-500 whitespace-nowrap',
  line: 'flex-1',
};

/** AppShell classes */
export const appShellClasses = {
  container: 'min-h-screen bg-neutral-50',
  header: 'bg-neutral-white border-b border-neutral-200 z-header',
  headerFixed: 'fixed top-0 left-0 right-0',
  sidebar: 'bg-neutral-white border-r border-neutral-200 transition-width duration-200',
  sidebarFixed: 'fixed top-0 bottom-0 z-sidebar',
  sidebarLeft: 'left-0',
  sidebarRight: 'right-0 border-r-0 border-l',
  main: 'flex-1',
  mainWithHeader: 'pt-header',
  mainWithSidebar: 'transition-margin duration-200',
  footer: 'bg-neutral-white border-t border-neutral-200',
};

/** PageLayout classes */
export const pageLayoutClasses = {
  base: 'w-full',
  default: '',
  centered: 'flex flex-col items-center',
  sidebar: 'flex flex-row',
  split: 'grid grid-cols-2',
};

export const pageLayoutPaddingClasses: Record<Size | 'none', string> = {
  none: '',
  xs: 'p-2',
  sm: 'p-4',
  md: 'p-6',
  lg: 'p-8',
  xl: 'p-12',
};
