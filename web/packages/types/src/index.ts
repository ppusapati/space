/**
 * @chetana/types
 * Single source of truth for all shared TypeScript types across Chetana ERP.
 *
 * Re-exports all types from @chetana/core with convenient groupings.
 * Import from here instead of hunting across packages.
 *
 * @example
 * import type { FormSchema, Size, BaseEntity, MenuItem } from '@chetana/types';
 *
 * @packageDocumentation
 */

// ============================================================================
// ALL CORE TYPES (re-exported from @chetana/core)
// ============================================================================

export type {
  // ---- Common / Primitives ----
  Size,
  ColorVariant,
  ComponentState,
  ValidationState,
  Position,
  ExtendedPosition,
  Alignment,
  Justify,
  SortDirection,
  LoadingState,

  // ---- Base Component Props ----
  BaseProps,
  DisableableProps,
  LoadableProps,
  FormElementProps,

  // ---- Entity Types ----
  BaseEntity,
  SoftDeletableEntity,
  AuditableEntity,
  AuditEntry,
  FieldChange,

  // ---- Navigation Types ----
  MenuItem,
  BreadcrumbItem,
  TabItem,

  // ---- Data / Filter Types ----
  FilterOperator,
  FilterValue,
  SortConfig,
  PaginationState,
  SelectionState,

  // ---- Notification / Error Types ----
  NotificationOptions,
  BaseError,
  ApiError,
  ValidationError,

  // ---- Action Types ----
  Action,
  ConfirmationConfig,
  BulkAction,

  // ---- Date/Time Types ----
  DateRange,
  DateRangeWithPreset,
  DateRangePreset,

  // ---- Utility Types ----
  RequiredKeys,
  OptionalKeys,
  DeepPartial,
  KeysOfType,
  Nullable,
  Maybe,
} from '@chetana/core';

// ---- Form Types ----
export type {
  Form,
  FormStatus,
  FormErrors,
  FormTouched,
  FormDirty,
  FieldProps,
  FieldMeta,
  FieldState,
  FormValidation,
  ValidationSchema,
  ValidationRules,
  FieldValidation,
  ValidateFn,
  ValidationRule,
  FormField,
  TextField,
  NumberField,
  SelectField,
  SelectOption,
  SelectOptionGroup,
  DateField,
  DateRangeField,
  CheckboxField,
  CheckboxGroupField,
  RadioField,
  SwitchField,
  TextareaField,
  RichTextField,
  FileField,
  AutocompleteField,
  ColorField,
  SliderField,
  RatingField,
  ArrayField,
  ObjectField,
  HiddenField,
  CustomField,
  FormFieldConfig,
  FormSchema,
  FormSection,
  FormSubmission,
} from '@chetana/core';

// ---- Page Types ----
export type {
  Page,
  PageMeta,
  PageError,
  ListPage,
  DetailPage,
  FormPage,
  DashboardPage,
  ReportPage,
  PageConfig,
} from '@chetana/core';

// ---- Layout Types ----
export type {
  Layout,
  Breakpoint,
  LayoutConfig,
  AppShellLayout,
  AppShellConfig,
  FormLayout,
  FormLayoutConfig,
  ResponsiveValue,
} from '@chetana/core';

export { BREAKPOINTS } from '@chetana/core';

// ---- Modal Types ----
export type {
  Modal,
  ModalConfig,
  ModalSize,
  ConfirmationModal,
  AlertModal,
  FormModal,
  DrawerModal,
  ModalInstance,
  ModalManagerState,
} from '@chetana/core';

// ---- List / Table Types ----
export type {
  List,
  ListState,
  ListColumn,
  FilterType,
  FilterDefinition,
  DataGridList,
  SimpleList,
  TreeList,
  VirtualList,
} from '@chetana/core';

// ---- Detail Types ----
export type {
  Detail,
  DetailMode,
  DetailState,
  DetailSection,
  DetailField,
  ProfileDetail,
  DocumentDetail,
  TimelineDetail,
  AuditDetail,
} from '@chetana/core';

// ---- View Types ----
export type {
  View,
  ViewConfig,
  CardView,
  KanbanView,
  KanbanColumn,
  KanbanItem,
  CalendarView,
  CalendarEvent,
  TreeView,
  TreeNode,
  ActivityFeedView,
  ActivityItem,
  StatisticsView,
  MetricItem,
} from '@chetana/core';
