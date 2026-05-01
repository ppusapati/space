/**
 * Common types shared across all components
 */

/** Component size variants */
export type Size = 'xs' | 'sm' | 'md' | 'lg' | 'xl';

/** Component color variants */
export type ColorVariant = 'primary' | 'secondary' | 'success' | 'warning' | 'error' | 'info' | 'neutral';

/** Component state */
export type ComponentState = 'default' | 'hover' | 'focus' | 'active' | 'disabled';

/** Form field validation state */
export type ValidationState = 'default' | 'valid' | 'invalid' | 'pending';

/** Position variants */
export type Position = 'top' | 'right' | 'bottom' | 'left';
export type Alignment = 'start' | 'center' | 'end';

/** Base component props that all components can extend */
export interface BaseProps {
  /** Additional CSS classes */
  class?: string;
  /** Component ID */
  id?: string;
  /** Test ID for testing */
  testId?: string;
}

/** Props for components that can be disabled */
export interface DisableableProps {
  disabled?: boolean;
}

/** Props for components with loading state */
export interface LoadableProps {
  loading?: boolean;
}

/** Props for form elements */
export interface FormElementProps extends BaseProps, DisableableProps {
  name?: string;
  required?: boolean;
  readonly?: boolean;
}

/** Generic event payload */
export interface EventPayload<T = unknown> {
  value: T;
  originalEvent?: Event;
}

/** Sort direction */
export type SortDirection = 'asc' | 'desc' | null;

/** Filter operator */
export type FilterOperator =
  | 'equals'
  | 'notEquals'
  | 'contains'
  | 'startsWith'
  | 'endsWith'
  | 'greaterThan'
  | 'lessThan'
  | 'greaterThanOrEqual'
  | 'lessThanOrEqual'
  | 'between'
  | 'isEmpty'
  | 'isNotEmpty';

/** Filter value */
export interface FilterValue {
  column: string;
  operator: FilterOperator;
  value: unknown;
  secondValue?: unknown; // For 'between' operator
}

/** Pagination state */
export interface PaginationState {
  page: number;
  pageSize: number;
  total: number;
}

/** Menu item structure */
export interface MenuItem {
  id: string;
  label: string;
  icon?: string;
  href?: string;
  disabled?: boolean;
  children?: MenuItem[];
  divider?: boolean;
}

/** Breadcrumb item */
export interface BreadcrumbItem {
  label: string;
  href?: string;
  icon?: string;
}

/** Tab item */
export interface TabItem {
  id: string;
  label: string;
  icon?: string;
  disabled?: boolean;
  badge?: string | number;
}

/** Notification/Toast options */
export interface NotificationOptions {
  id?: string;
  title?: string;
  message: string;
  type?: ColorVariant;
  duration?: number;
  dismissible?: boolean;
  action?: {
    label: string;
    onClick: () => void;
  };
}
