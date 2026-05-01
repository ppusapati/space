/**
 * Navigation component types
 */

import type { Size, BaseProps, MenuItem, BreadcrumbItem, TabItem } from '../types';

/** Sidebar props */
export interface SidebarProps extends BaseProps {
  /** Whether sidebar is collapsed */
  collapsed?: boolean;
  /** Sidebar width when expanded */
  width?: string;
  /** Sidebar width when collapsed */
  collapsedWidth?: string;
  /** Position */
  position?: 'left' | 'right';
  /** Fixed or static positioning */
  fixed?: boolean;
  /** Enable collapse toggle */
  collapsible?: boolean;
  /** Overlay mode for mobile */
  overlay?: boolean;
}

/** NavBar props */
export interface NavBarProps extends BaseProps {
  /** Whether navbar is fixed */
  fixed?: boolean;
  /** Navbar height */
  height?: string;
  /** Show shadow */
  shadow?: boolean;
  /** Transparent background */
  transparent?: boolean;
}

/** Breadcrumbs props */
export interface BreadcrumbsProps extends BaseProps {
  /** Breadcrumb items */
  items: BreadcrumbItem[];
  /** Separator character/element */
  separator?: string;
  /** Maximum items to show (rest collapsed) */
  maxItems?: number;
  /** Size */
  size?: Size;
}

/** Tabs props */
export interface TabsProps extends BaseProps {
  /** Tab items */
  items: TabItem[];
  /** Currently active tab */
  activeId?: string;
  /** Tab variant */
  variant?: 'line' | 'enclosed' | 'pills' | 'soft';
  /** Size */
  size?: Size;
  /** Orientation */
  orientation?: 'horizontal' | 'vertical';
  /** Full width tabs */
  fullWidth?: boolean;
}

/** Stepper props */
export interface StepperProps extends BaseProps {
  /** Steps */
  steps: StepItem[];
  /** Current step (0-indexed) */
  currentStep: number;
  /** Orientation */
  orientation?: 'horizontal' | 'vertical';
  /** Size */
  size?: Size;
  /** Allow clicking on completed steps */
  clickable?: boolean;
}

export interface StepItem {
  id: string;
  title: string;
  description?: string;
  icon?: string;
  optional?: boolean;
  error?: boolean;
}

/** Menu props */
export interface MenuProps extends BaseProps {
  /** Menu items */
  items: MenuItem[];
  /** Whether menu is open */
  open?: boolean;
  /** Trigger element ref */
  trigger?: HTMLElement;
  /** Position relative to trigger */
  position?: 'bottom-start' | 'bottom-end' | 'top-start' | 'top-end';
  /** Size */
  size?: Size;
}

/** Pagination props */
export interface PaginationProps extends BaseProps {
  /** Current page */
  page: number;
  /** Total pages */
  totalPages: number;
  /** Items per page */
  pageSize?: number;
  /** Total items */
  totalItems?: number;
  /** Show first/last buttons */
  showFirstLast?: boolean;
  /** Show page size selector */
  showPageSize?: boolean;
  /** Available page sizes */
  pageSizes?: number[];
  /** Size */
  size?: Size;
  /** Variant */
  variant?: 'default' | 'simple' | 'minimal';
}

/** Navigation classes */
export const sidebarClasses = {
  container: 'flex flex-col h-full bg-neutral-white border-r border-neutral-200 transition-all duration-300',
  containerRight: 'border-l border-r-0',
  fixed: 'fixed top-0 bottom-0 z-drawer',
  overlay: 'fixed inset-0 z-drawer',
  overlayBackdrop: 'absolute inset-0 bg-neutral-black/50',
  header: 'flex items-center justify-between p-4 border-b border-neutral-200',
  content: 'flex-1 overflow-y-auto p-2',
  footer: 'p-4 border-t border-neutral-200',
  collapseBtn: 'p-2 rounded-md hover:bg-neutral-100 text-neutral-500',
  nav: 'space-y-1',
  navItem: 'flex items-center gap-3 px-3 py-2 rounded-md text-neutral-700 hover:bg-neutral-100 transition-colors',
  navItemActive: 'bg-brand-primary-50 text-brand-primary-700 font-medium',
  navItemIcon: 'w-5 h-5 shrink-0',
  navItemLabel: 'truncate',
  navGroup: 'mt-4',
  navGroupLabel: 'px-3 py-2 text-xs font-semibold text-neutral-400 uppercase tracking-wide',
};

export const navbarClasses = {
  container: 'w-full bg-neutral-white border-b border-neutral-200',
  fixed: 'fixed top-0 left-0 right-0 z-header',
  shadow: 'shadow-sm',
  transparent: 'bg-transparent border-transparent',
  inner: 'flex items-center justify-between px-4',
  brand: 'flex items-center gap-2',
  nav: 'flex items-center gap-1',
  navItem: 'px-3 py-2 text-sm font-medium text-neutral-600 hover:text-neutral-900 rounded-md hover:bg-neutral-100 transition-colors',
  navItemActive: 'text-brand-primary-600 bg-brand-primary-50',
  actions: 'flex items-center gap-2',
  mobile: 'lg:hidden',
  desktop: 'hidden lg:flex',
};

export const breadcrumbsClasses = {
  container: 'flex items-center flex-wrap gap-1',
  item: 'flex items-center gap-1',
  link: 'text-neutral-600 hover:text-brand-primary-600 transition-colors',
  current: 'text-neutral-900 font-medium',
  separator: 'text-neutral-400 mx-1',
  icon: 'w-4 h-4',
};

export const breadcrumbsSizeClasses: Record<Size, string> = {
  xs: 'text-xs',
  sm: 'text-sm',
  md: 'text-sm',
  lg: 'text-base',
  xl: 'text-lg',
};

export const tabsClasses = {
  container: 'w-full',
  list: 'flex',
  listVertical: 'flex-col',
  tab: 'relative px-4 py-2 text-sm font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-brand-primary-500 focus:ring-offset-2',
  tabActive: 'text-brand-primary-600',
  tabInactive: 'text-neutral-600 hover:text-neutral-900',
  panel: 'mt-4',
  // Variants
  line: {
    list: 'border-b border-neutral-200',
    tab: 'border-b-2 border-transparent -mb-px',
    tabActive: 'border-brand-primary-500',
  },
  enclosed: {
    list: 'border-b border-neutral-200',
    tab: 'border border-transparent rounded-t-md -mb-px',
    tabActive: 'border-neutral-200 border-b-neutral-white bg-neutral-white',
  },
  pills: {
    list: 'gap-2',
    tab: 'rounded-full',
    tabActive: 'bg-brand-primary-500 text-neutral-white',
  },
  soft: {
    list: 'gap-1 p-1 bg-neutral-100 rounded-lg',
    tab: 'rounded-md',
    tabActive: 'bg-neutral-white shadow-sm',
  },
};

export const tabsSizeClasses: Record<Size, string> = {
  xs: 'px-2 py-1 text-xs',
  sm: 'px-3 py-1.5 text-sm',
  md: 'px-4 py-2 text-sm',
  lg: 'px-5 py-2.5 text-base',
  xl: 'px-6 py-3 text-lg',
};

export const stepperClasses = {
  container: 'flex',
  containerVertical: 'flex-col',
  step: 'flex items-center',
  stepVertical: 'flex-col items-start',
  indicator: 'flex items-center justify-center w-8 h-8 rounded-full border-2 font-medium text-sm transition-colors',
  indicatorPending: 'border-neutral-300 text-neutral-500 bg-neutral-white',
  indicatorActive: 'border-brand-primary-500 text-brand-primary-600 bg-brand-primary-50',
  indicatorCompleted: 'border-brand-primary-500 bg-brand-primary-500 text-neutral-white',
  indicatorError: 'border-semantic-error-500 bg-semantic-error-50 text-semantic-error-600',
  content: 'ml-3',
  title: 'text-sm font-medium text-neutral-900',
  description: 'text-xs text-neutral-500',
  connector: 'flex-1 h-0.5 bg-neutral-200 mx-4',
  connectorActive: 'bg-brand-primary-500',
  connectorVertical: 'w-0.5 h-8 ml-4 mt-1 mb-1',
};

export const menuClasses = {
  container: 'absolute z-dropdown bg-neutral-white border border-neutral-200 rounded-md shadow-lg py-1 min-w-40',
  item: 'flex items-center gap-2 w-full px-3 py-2 text-sm text-neutral-700 hover:bg-neutral-100 cursor-pointer transition-colors',
  itemDisabled: 'opacity-50 cursor-not-allowed hover:bg-transparent',
  itemIcon: 'w-4 h-4 text-neutral-500',
  divider: 'my-1 border-t border-neutral-200',
  submenu: 'relative',
  submenuIndicator: 'ml-auto',
};

export const paginationComponentClasses = {
  container: 'flex items-center gap-2',
  button: 'p-2 rounded-md border border-neutral-300 text-neutral-600 hover:bg-neutral-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors',
  pageButton: 'min-w-8 h-8 px-2 rounded-md text-sm font-medium transition-colors',
  pageButtonActive: 'bg-brand-primary-500 text-neutral-white',
  pageButtonInactive: 'text-neutral-700 hover:bg-neutral-100',
  ellipsis: 'px-2 text-neutral-400',
  info: 'text-sm text-neutral-600',
  select: 'px-2 py-1 text-sm border border-neutral-300 rounded-md',
};

export const paginationSizeClasses: Record<Size, { button: string; text: string }> = {
  xs: { button: 'min-w-6 h-6 p-1', text: 'text-xs' },
  sm: { button: 'min-w-7 h-7 p-1.5', text: 'text-sm' },
  md: { button: 'min-w-8 h-8 p-2', text: 'text-sm' },
  lg: { button: 'min-w-10 h-10 p-2.5', text: 'text-base' },
  xl: { button: 'min-w-12 h-12 p-3', text: 'text-lg' },
};
