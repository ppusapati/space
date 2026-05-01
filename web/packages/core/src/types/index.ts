/**
 * Core Types - Export all generic types
 * @packageDocumentation
 */

// ============================================================================
// COMMON TYPES
// ============================================================================
export type {
  // Base types
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

  // Base interfaces
  BaseProps,
  DisableableProps,
  LoadableProps,
  FormElementProps,

  // Entity types
  BaseEntity,
  SoftDeletableEntity,
  AuditableEntity,
  AuditEntry,
  FieldChange,

  // Navigation types
  MenuItem,
  BreadcrumbItem,
  TabItem,

  // Data types
  FilterOperator,
  FilterValue,
  SortConfig,
  PaginationState,
  SelectionState,

  // Notification types
  NotificationOptions,

  // Error types
  BaseError,
  ApiError,
  ValidationError,

  // Action types
  Action,
  ConfirmationConfig,
  BulkAction,

  // Date/Time types
  DateRange,
  DateRangeWithPreset,
  DateRangePreset,

  // Slot types
  SlotConfig,

  // Utility types
  RequiredKeys,
  OptionalKeys,
  DeepPartial,
  KeysOfType,
  Nullable,
  Maybe,
} from './common.types.js';

// ============================================================================
// PAGE TYPES
// ============================================================================
export type {
  // Core page types
  Page,
  PageMeta,
  PageError,

  // Page variants
  ListPage,
  DetailPage,
  FormPage,
  DashboardPage,
  ReportPage,

  // Supporting types
  ColumnConfig,
  FormValidationState,
  DateRangeFilter,
  WidgetConfig,
  ChartType,
  Aggregation,
  ReportSchedule,

  // Slots
  PageSlots,
  PageConfig,
} from './page.types.js';

// ============================================================================
// VIEW TYPES
// ============================================================================
export type {
  // Core view types
  View,
  ViewError,
  ViewConfig,
  ViewSlots,

  // View variants
  CardView,
  TimelineView,
  TimelineItem,
  KanbanView,
  KanbanColumn,
  KanbanItem,
  CalendarView,
  CalendarViewMode,
  CalendarEvent,
  RecurrenceRule,
  TreeView,
  TreeNode,
  GalleryView,
  GalleryItem,
  ActivityFeedView,
  ActivityItem,
  ActivityFilters,
  StatisticsView,
  MetricItem,

  // View actions
  ViewAction,
} from './view.types.js';

// ============================================================================
// LAYOUT TYPES
// ============================================================================
export type {
  // Core layout types
  Layout,
  Breakpoint,
  LayoutConfig,

  // Layout variants
  MasterDetailLayout,
  MasterDetailConfig,
  DashboardLayout,
  DashboardLayoutConfig,
  WidgetLayout,
  SplitLayout,
  SplitLayoutConfig,
  TabLayout,
  TabLayoutConfig,
  TabConfig,
  WizardLayout,
  WizardLayoutConfig,
  WizardStep,
  AppShellLayout,
  AppShellConfig,
  PageLayout,
  PageLayoutConfig,
  CardLayout,
  CardLayoutConfig,
  FormLayout,
  FormLayoutConfig,

  // Responsive utilities
  ResponsiveValue,
} from './layout.types.js';

export { BREAKPOINTS } from './layout.types.js';

// ============================================================================
// MODAL TYPES
// ============================================================================
export type {
  // Core modal types
  Modal,
  ModalConfig,
  ModalSize,
  ModalSlots,

  // Modal variants
  ConfirmationModal,
  ConfirmationData,
  AlertModal,
  AlertData,
  FormModal,
  SelectionModal,
  SelectionModalData,
  PreviewModal,
  WizardModal,
  WizardStepConfig,
  CommandPaletteModal,
  CommandGroup,
  CommandAction,
  DrawerModal,
  PromptModal,
  PromptData,
  ImageCropModal,
  ImageCropData,
  CroppedImage,

  // Modal manager
  ModalInstance,
  ModalManagerState,
  ModalManagerActions,
} from './modal.types.js';

// ============================================================================
// WIDGET TYPES
// ============================================================================
export type {
  // Core widget types
  Widget,
  WidgetType,
  WidgetState,
  WidgetConfig as BaseWidgetConfig,
  WidgetSlots,

  // Widget variants
  MetricWidget,
  MetricWidgetConfig,
  MetricData,
  ChartWidget,
  ChartWidgetConfig,
  ChartData,
  ChartDataset,
  TableWidget,
  TableWidgetConfig,
  WidgetTableColumn,
  ListWidget,
  ListWidgetConfig,
  CalendarWidget,
  CalendarWidgetConfig,
  TimelineWidget,
  TimelineWidgetConfig,
  MapWidget,
  MapWidgetConfig,
  MapMarker,
  ProgressWidget,
  ProgressWidgetConfig,
  ProgressData,
  ActivityWidget,
  ActivityWidgetConfig,
} from './widget.types.js';

// ============================================================================
// FORM TYPES
// ============================================================================
export type {
  // Core form types
  Form,
  FormStatus,
  FormErrors,
  FormTouched,
  FormDirty,
  FieldProps,
  FieldMeta,
  FieldState,

  // Validation types
  FormValidation,
  ValidationSchema,
  ValidationRules,
  FieldValidation,
  ValidateFn,
  ValidationRule,

  // Field types
  FormField,
  TextField,
  NumberField,
  SelectField,
  SelectOption,
  SelectOptionGroup,
  DateField,
  DateRangeField,
  DateRangePreset as FormDateRangePreset,
  CheckboxField,
  CheckboxGroupField,
  RadioField,
  SwitchField,
  TextareaField,
  RichTextField,
  RichTextToolbarItem,
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

  // Schema/Builder types
  FormSchema,
  FormLayout as FormSchemaLayout,
  FormSection,
  FormSubmission,

  // Slots
  FormSlots,
} from './form.types.js';

// ============================================================================
// LIST TYPES
// ============================================================================
export type {
  // Core list types
  List,
  ListState,
  ListError,
  ListConfig,

  // Column types
  ListColumn,
  ColumnGroup,

  // Filter types
  FilterType,
  FilterDefinition,
  FilterOption,
  ActiveFilterChip,

  // Export types
  ExportFormat,
  ExportOptions,

  // List variants
  DataGridList,
  SimpleList,
  GroupedList,
  ListGroup,
  TreeList,
  TreeListItem,
  VirtualList,

  // Slots and events
  ListSlots,
  ListEvents,
} from './list.types.js';

// ============================================================================
// DETAIL TYPES
// ============================================================================
export type {
  // Core detail types
  Detail,
  DetailMode,
  DetailState,
  DetailError,
  DetailConfig,
  DeleteConfirmation,

  // Section types
  DetailSection,
  DetailField,
  DetailFieldType,
  DetailInputConfig,
  FieldValidationRule,

  // Related data types
  RelatedDataState,
  RelatedDataConfig,

  // Detail variants
  ProfileDetail,
  ProfileStat,
  DocumentDetail,
  LineItemsConfig,
  DocumentTotal,
  TimelineDetail,
  ActivityConfig,
  AuditDetail,
  EntityDiff,

  // Slots and events
  DetailSlots,
  DetailEvents,

  // Action types
  StandardDetailAction,
  DetailActionGroup,
} from './detail.types.js';
