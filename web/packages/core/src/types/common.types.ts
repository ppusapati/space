/**
 * Common Types - Shared across all generic types
 */

import type { Component, Snippet } from 'svelte';

// ============================================================================
// BASE TYPES
// ============================================================================

/** Size variants used across the system */
export type Size = 'xs' | 'sm' | 'md' | 'lg' | 'xl';

/** Color variants for semantic styling */
export type ColorVariant =
  | 'primary'
  | 'secondary'
  | 'success'
  | 'warning'
  | 'error'
  | 'info'
  | 'neutral';

/** Component interaction states */
export type ComponentState = 'default' | 'hover' | 'focus' | 'active' | 'disabled';

/** Validation states for form elements */
export type ValidationState = 'default' | 'valid' | 'invalid' | 'pending';

/** Position for tooltips, popovers, etc. */
export type Position = 'top' | 'right' | 'bottom' | 'left';

/** Extended position with corners */
export type ExtendedPosition =
  | Position
  | 'top-left'
  | 'top-right'
  | 'bottom-left'
  | 'bottom-right';

/** Alignment options */
export type Alignment = 'start' | 'center' | 'end';

/** Justify content options */
export type Justify = 'start' | 'end' | 'center' | 'between' | 'around' | 'evenly';

/** Sort direction */
export type SortDirection = 'asc' | 'desc' | null;

/** Loading state */
export type LoadingState = 'idle' | 'loading' | 'refreshing' | 'error' | 'success';

// ============================================================================
// BASE INTERFACES
// ============================================================================

/** Base props for all components */
export interface BaseProps {
  class?: string;
  id?: string;
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

// ============================================================================
// ENTITY TYPES
// ============================================================================

/** Base entity with common fields */
export interface BaseEntity {
  id: string;
  createdAt: Date;
  updatedAt: Date;
  createdBy?: string;
  updatedBy?: string;
}

/** Soft-deletable entity */
export interface SoftDeletableEntity extends BaseEntity {
  deletedAt?: Date | null;
  deletedBy?: string | null;
  isDeleted: boolean;
}

/** Auditable entity with version tracking */
export interface AuditableEntity extends BaseEntity {
  version: number;
  auditTrail?: AuditEntry[];
}

/** Audit trail entry */
export interface AuditEntry {
  id: string;
  action: 'create' | 'update' | 'delete' | 'restore';
  timestamp: Date;
  userId: string;
  userName?: string;
  changes?: FieldChange[];
  metadata?: Record<string, unknown>;
}

/** Field change record */
export interface FieldChange {
  field: string;
  label?: string;
  oldValue: unknown;
  newValue: unknown;
}

// ============================================================================
// NAVIGATION TYPES
// ============================================================================

/** Menu item */
export interface MenuItem {
  id: string;
  label: string;
  icon?: string;
  href?: string;
  disabled?: boolean;
  hidden?: boolean;
  children?: MenuItem[];
  divider?: boolean;
  badge?: string | number;
  badgeVariant?: ColorVariant;
  permissions?: string[];
  onClick?: () => void;
}

/** Breadcrumb item */
export interface BreadcrumbItem {
  label: string;
  href?: string;
  icon?: string;
  disabled?: boolean;
}

/** Tab item */
export interface TabItem {
  id: string;
  label: string;
  icon?: string;
  disabled?: boolean;
  badge?: string | number;
  closable?: boolean;
}

// ============================================================================
// DATA TYPES
// ============================================================================

/** Filter operator */
export type FilterOperator =
  | 'equals'
  | 'notEquals'
  | 'contains'
  | 'notContains'
  | 'startsWith'
  | 'endsWith'
  | 'gt'
  | 'gte'
  | 'lt'
  | 'lte'
  | 'between'
  | 'in'
  | 'notIn'
  | 'isEmpty'
  | 'isNotEmpty'
  | 'isNull'
  | 'isNotNull';

/** Filter value */
export interface FilterValue {
  field: string;
  operator: FilterOperator;
  value: unknown;
  secondValue?: unknown; // For 'between' operator
}

/** Sort configuration */
export interface SortConfig {
  column: string | null;
  direction: SortDirection;
  multiSort?: Array<{ column: string; direction: 'asc' | 'desc' }>;
}

/** Pagination state */
export interface PaginationState {
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
  hasNext: boolean;
  hasPrevious: boolean;
}

/** Selection state */
export interface SelectionState<T> {
  selectedItems: T[];
  selectedKeys: Set<string | number>;
  isAllSelected: boolean;
  isIndeterminate: boolean;
}

// ============================================================================
// NOTIFICATION TYPES
// ============================================================================

/** Notification options */
export interface NotificationOptions {
  id?: string;
  title?: string;
  message: string;
  type?: ColorVariant;
  duration?: number;
  dismissible?: boolean;
  position?: ExtendedPosition;
  action?: {
    label: string;
    onClick: () => void;
  };
}

// ============================================================================
// ERROR TYPES
// ============================================================================

/** Base error interface */
export interface BaseError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
  timestamp?: Date;
  retryable?: boolean;
}

/** API error */
export interface ApiError extends BaseError {
  status?: number;
  endpoint?: string;
  requestId?: string;
}

/** Validation error */
export interface ValidationError extends BaseError {
  field?: string;
  constraint?: string;
}

// ============================================================================
// ACTION TYPES
// ============================================================================

/** Action definition */
export interface Action<TData = unknown> {
  id: string;
  label: string;
  icon?: string;
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
  disabled?: boolean | ((data: TData) => boolean);
  visible?: boolean | ((data: TData) => boolean);
  loading?: boolean;
  shortcut?: string;
  confirmation?: ConfirmationConfig;
  handler: (data: TData) => void | Promise<void>;
}

/** Confirmation dialog config */
export interface ConfirmationConfig {
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  destructive?: boolean;
  variant?: ColorVariant;
}

/** Bulk action for multiple items */
export interface BulkAction<TItem = unknown> extends Omit<Action<TItem[]>, 'handler'> {
  minSelection?: number;
  maxSelection?: number;
  handler: (items: TItem[]) => void | Promise<void>;
}

// ============================================================================
// DATE/TIME TYPES
// ============================================================================

/** Date range */
export interface DateRange {
  start: Date | null;
  end: Date | null;
}

/** Date range with preset */
export interface DateRangeWithPreset extends DateRange {
  preset?: DateRangePreset;
}

/** Date range preset */
export type DateRangePreset =
  | 'today'
  | 'yesterday'
  | 'last7days'
  | 'last30days'
  | 'thisWeek'
  | 'lastWeek'
  | 'thisMonth'
  | 'lastMonth'
  | 'thisQuarter'
  | 'lastQuarter'
  | 'thisYear'
  | 'lastYear'
  | 'custom';

// ============================================================================
// SLOT TYPES
// ============================================================================

/** Generic slot configuration */
export interface SlotConfig<TData = unknown> {
  snippet?: Snippet<[TData]>;
  component?: Component;
  props?: Record<string, unknown>;
}

// ============================================================================
// UTILITY TYPES
// ============================================================================

/** Make specific keys required */
export type RequiredKeys<T, K extends keyof T> = T & Required<Pick<T, K>>;

/** Make specific keys optional */
export type OptionalKeys<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>;

/** Deep partial */
export type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P];
};

/** Extract keys of type */
export type KeysOfType<T, V> = {
  [K in keyof T]: T[K] extends V ? K : never;
}[keyof T];

/** Nullable */
export type Nullable<T> = T | null;

/** Maybe */
export type Maybe<T> = T | null | undefined;
