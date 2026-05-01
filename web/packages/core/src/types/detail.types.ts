/**
 * Detail Generic Types
 * Comprehensive detail view handling for entity display, editing, and actions
 */

import type { Component, Snippet } from 'svelte';
import type {
  BaseError,
  LoadingState,
  Action,
  ColorVariant,
  Size,
  AuditEntry,
} from './common.types.js';

// ============================================================================
// DETAIL GENERIC TYPE
// ============================================================================

/**
 * Generic Detail Type
 * @template TEntity - The entity type being displayed
 * @template TActions - Available action identifiers
 */
export interface Detail<TEntity, TActions extends string = string> {
  // Data
  entity: TEntity | null;
  originalEntity: TEntity | null;

  // State
  state: DetailState;
  mode: DetailMode;
  loadingState: LoadingState;

  // Configuration
  config: DetailConfig<TEntity>;

  // Sections
  sections: DetailSection<TEntity>[];
  activeSectionId?: string;

  // Related Data
  relatedData: Record<string, RelatedDataState<unknown>>;

  // Actions
  actions: Record<TActions, Action<TEntity>>;

  // Methods - Data loading
  load: (id: string | number) => Promise<void>;
  reload: () => Promise<void>;
  refresh: () => Promise<void>;

  // Methods - Mode switching
  setMode: (mode: DetailMode) => void;
  edit: () => void;
  view: () => void;
  create: (initialData?: Partial<TEntity>) => void;

  // Methods - Entity operations
  save: () => Promise<void>;
  delete: () => Promise<void>;
  duplicate: () => Promise<TEntity>;
  archive: () => Promise<void>;
  restore: () => Promise<void>;

  // Methods - Field operations
  setFieldValue: <K extends keyof TEntity>(field: K, value: TEntity[K]) => void;
  resetField: <K extends keyof TEntity>(field: K) => void;
  resetAll: () => void;
  isDirty: () => boolean;
  getChangedFields: () => Partial<TEntity>;

  // Methods - Validation
  validate: () => Promise<boolean>;
  validateField: <K extends keyof TEntity>(field: K) => Promise<string | null>;

  // Methods - Actions
  executeAction: (actionId: TActions) => Promise<void>;

  // Methods - Related data
  loadRelatedData: (key: string) => Promise<void>;
  refreshRelatedData: (key: string) => Promise<void>;

  // Methods - Sections
  setActiveSection: (sectionId: string) => void;
  toggleSection: (sectionId: string) => void;

  // Events
  onBeforeLoad?: (id: string | number) => void | Promise<void>;
  onAfterLoad?: (entity: TEntity) => void;
  onBeforeSave?: (entity: TEntity) => TEntity | Promise<TEntity>;
  onAfterSave?: (entity: TEntity) => void;
  onBeforeDelete?: (entity: TEntity) => boolean | Promise<boolean>;
  onAfterDelete?: () => void;
  onError?: (error: DetailError) => void;
  onModeChange?: (mode: DetailMode) => void;
  onFieldChange?: <K extends keyof TEntity>(field: K, value: TEntity[K], oldValue: TEntity[K]) => void;
}

/** Detail mode */
export type DetailMode = 'view' | 'edit' | 'create';

/** Detail state */
export interface DetailState {
  isLoading: boolean;
  isSaving: boolean;
  isDeleting: boolean;
  isDirty: boolean;
  isValid: boolean;
  hasError: boolean;
  error?: DetailError;
  lastSaved?: Date;
  lastModified?: Date;
}

/** Detail error */
export interface DetailError extends BaseError {
  field?: string;
  retryable?: boolean;
}

/** Detail configuration */
export interface DetailConfig<TEntity> {
  id: string;
  title?: string | ((entity: TEntity | null) => string);
  description?: string | ((entity: TEntity | null) => string);
  icon?: string;

  // Entity identification
  entityKey: keyof TEntity;
  entityLabel?: keyof TEntity | ((entity: TEntity) => string);

  // Features
  editable?: boolean;
  deletable?: boolean;
  duplicatable?: boolean;
  archivable?: boolean;
  printable?: boolean;

  // Auto-save
  autoSave?: boolean;
  autoSaveDelay?: number;

  // Confirmation dialogs
  confirmDelete?: boolean;
  confirmDiscard?: boolean;
  deleteConfirmation?: DeleteConfirmation;

  // Display
  layout?: 'default' | 'tabs' | 'accordion' | 'wizard';
  maxWidth?: Size | 'full';
  padding?: Size | 'none';

  // Back navigation
  showBackButton?: boolean;
  backUrl?: string;
  backLabel?: string;
}

/** Delete confirmation config */
export interface DeleteConfirmation {
  title: string;
  message: string | ((entity: unknown) => string);
  confirmText?: string;
  cancelText?: string;
  requireConfirmation?: boolean;
  confirmationText?: string; // Text user must type to confirm
}

// ============================================================================
// DETAIL SECTIONS
// ============================================================================

/** Detail section */
export interface DetailSection<TEntity> {
  id: string;
  title: string;
  description?: string;
  icon?: string;

  // Display
  collapsible?: boolean;
  defaultCollapsed?: boolean;
  visible?: boolean | ((entity: TEntity | null, mode: DetailMode) => boolean);
  disabled?: boolean | ((entity: TEntity | null, mode: DetailMode) => boolean);

  // Layout
  columns?: 1 | 2 | 3 | 4;
  gap?: Size;

  // Fields
  fields: DetailField<TEntity>[];

  // Custom content
  component?: Component;
  snippet?: Snippet<[TEntity | null, DetailMode]>;

  // Validation
  validate?: (entity: TEntity) => boolean | Promise<boolean>;

  // Badge/indicator
  badge?: string | number | ((entity: TEntity | null) => string | number | null);
  badgeVariant?: ColorVariant;
}

/** Detail field */
export interface DetailField<TEntity> {
  key: keyof TEntity | string;
  label: string;
  description?: string;

  // Display
  type: DetailFieldType;
  format?: (value: unknown, entity: TEntity) => string;
  component?: Component;
  snippet?: Snippet<[{ value: unknown; entity: TEntity; mode: DetailMode }]>;

  // Visibility
  visible?: boolean | ((entity: TEntity | null, mode: DetailMode) => boolean);
  visibleInModes?: DetailMode[];
  hidden?: boolean;

  // Editability
  editable?: boolean | ((entity: TEntity | null) => boolean);
  editableInModes?: DetailMode[];
  readonly?: boolean;

  // Validation
  required?: boolean | ((entity: TEntity | null) => boolean);
  validation?: FieldValidationRule<TEntity>[];

  // Layout
  span?: 1 | 2 | 3 | 4 | 'full';
  order?: number;

  // Styling
  labelClass?: string;
  valueClass?: string;
  containerClass?: string;

  // Input configuration (for edit mode)
  inputConfig?: DetailInputConfig;

  // Related data
  relatedDataKey?: string;
  relatedEntityDisplay?: (related: unknown) => string;
}

/** Detail field type */
export type DetailFieldType =
  | 'text'
  | 'number'
  | 'currency'
  | 'percent'
  | 'date'
  | 'datetime'
  | 'time'
  | 'boolean'
  | 'email'
  | 'phone'
  | 'url'
  | 'address'
  | 'user'
  | 'status'
  | 'badge'
  | 'tags'
  | 'rating'
  | 'progress'
  | 'avatar'
  | 'image'
  | 'file'
  | 'link'
  | 'json'
  | 'code'
  | 'html'
  | 'markdown'
  | 'color'
  | 'icon'
  | 'custom';

/** Detail input configuration */
export interface DetailInputConfig {
  type?: string;
  placeholder?: string;
  min?: number;
  max?: number;
  step?: number;
  minLength?: number;
  maxLength?: number;
  pattern?: string;
  options?: Array<{ label: string; value: unknown }>;
  multiple?: boolean;
  clearable?: boolean;
  searchable?: boolean;
  creatable?: boolean;
  loadOptions?: (query: string) => Promise<Array<{ label: string; value: unknown }>>;
}

/** Field validation rule */
export interface FieldValidationRule<TEntity> {
  validate: (value: unknown, entity: TEntity) => boolean | Promise<boolean>;
  message: string | ((value: unknown, entity: TEntity) => string);
  validateOn?: 'change' | 'blur' | 'submit';
}

// ============================================================================
// RELATED DATA
// ============================================================================

/** Related data state */
export interface RelatedDataState<TData> {
  data: TData | null;
  isLoading: boolean;
  hasError: boolean;
  error?: BaseError;
  lastLoaded?: Date;
}

/** Related data configuration */
export interface RelatedDataConfig<TEntity, TRelated = unknown> {
  key: string;
  title: string;
  description?: string;

  // Loading
  load: (entity: TEntity) => Promise<TRelated>;
  lazyLoad?: boolean;
  cacheTime?: number;

  // Display
  component?: Component;
  snippet?: Snippet<[TRelated, TEntity]>;
  emptyMessage?: string;

  // Features
  expandable?: boolean;
  defaultExpanded?: boolean;
  refreshable?: boolean;
  countBadge?: (data: TRelated) => number;
}

// ============================================================================
// DETAIL VARIANTS
// ============================================================================

/**
 * Profile Detail
 * For user/entity profile display
 */
export interface ProfileDetail<TEntity> extends Detail<TEntity> {
  // Profile specific
  avatarField: keyof TEntity;
  nameField: keyof TEntity;
  subtitleField?: keyof TEntity;
  coverImageField?: keyof TEntity;

  // Quick actions
  quickActions: Action<TEntity>[];

  // Stats
  stats?: ProfileStat<TEntity>[];
}

/** Profile stat */
export interface ProfileStat<TEntity> {
  label: string;
  value: keyof TEntity | ((entity: TEntity) => string | number);
  icon?: string;
  color?: ColorVariant;
}

/**
 * Document Detail
 * For document-like entities (invoices, orders, etc.)
 */
export interface DocumentDetail<TEntity, TLineItem = unknown> extends Detail<TEntity> {
  // Document header fields
  documentNumber: keyof TEntity;
  documentDate: keyof TEntity;
  documentStatus: keyof TEntity;

  // Line items
  lineItems: TLineItem[];
  lineItemsConfig: LineItemsConfig<TLineItem>;

  // Totals
  totals: DocumentTotal<TEntity>[];

  // Line item methods
  addLineItem: (item?: Partial<TLineItem>) => void;
  updateLineItem: (index: number, item: Partial<TLineItem>) => void;
  removeLineItem: (index: number) => void;
  reorderLineItems: (fromIndex: number, toIndex: number) => void;
  duplicateLineItem: (index: number) => void;
}

/** Line items configuration */
export interface LineItemsConfig<TLineItem> {
  columns: Array<{
    key: keyof TLineItem;
    header: string;
    width?: string;
    editable?: boolean;
    type?: DetailFieldType;
    format?: (value: unknown) => string;
  }>;
  minItems?: number;
  maxItems?: number;
  allowReorder?: boolean;
  allowDuplicate?: boolean;
  itemKey: keyof TLineItem;
  newItemDefaults?: Partial<TLineItem>;
}

/** Document total */
export interface DocumentTotal<TEntity> {
  label: string;
  field?: keyof TEntity;
  compute?: (entity: TEntity) => number;
  format?: (value: number) => string;
  highlight?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

/**
 * Timeline Detail
 * Detail view with activity timeline
 */
export interface TimelineDetail<TEntity, TActivity = unknown> extends Detail<TEntity> {
  // Timeline
  activities: TActivity[];
  activitiesLoading: boolean;

  // Timeline config
  activityConfig: ActivityConfig<TActivity>;

  // Timeline methods
  loadActivities: () => Promise<void>;
  addActivity: (activity: Partial<TActivity>) => Promise<void>;
}

/** Activity configuration */
export interface ActivityConfig<TActivity> {
  timestampField: keyof TActivity;
  typeField: keyof TActivity;
  actorField: keyof TActivity;
  descriptionField: keyof TActivity;
  iconMap?: Record<string, string>;
  colorMap?: Record<string, ColorVariant>;
}

/**
 * Audit Detail
 * Detail view with full audit history
 */
export interface AuditDetail<TEntity> extends Detail<TEntity> {
  // Audit trail
  auditEntries: AuditEntry[];
  auditLoading: boolean;

  // Audit methods
  loadAuditTrail: () => Promise<void>;
  viewVersion: (versionId: string) => Promise<TEntity>;
  compareVersions: (versionId1: string, versionId2: string) => Promise<EntityDiff<TEntity>>;
  restoreVersion: (versionId: string) => Promise<void>;
}

/** Entity diff */
export interface EntityDiff<TEntity> {
  changes: Array<{
    field: keyof TEntity;
    label: string;
    oldValue: unknown;
    newValue: unknown;
  }>;
}

// ============================================================================
// DETAIL SLOTS
// ============================================================================

/** Detail slots */
export interface DetailSlots<TEntity> {
  header?: Snippet<[TEntity | null, DetailMode]>;
  headerActions?: Snippet<[TEntity | null, DetailMode]>;
  beforeSections?: Snippet<[TEntity | null, DetailMode]>;
  afterSections?: Snippet<[TEntity | null, DetailMode]>;
  sidebar?: Snippet<[TEntity | null, DetailMode]>;
  footer?: Snippet<[TEntity | null, DetailMode]>;
  footerActions?: Snippet<[TEntity | null, DetailMode]>;
  empty?: Snippet;
  loading?: Snippet;
  error?: Snippet<[DetailError]>;
  sectionHeader?: Snippet<[DetailSection<TEntity>]>;
  sectionContent?: Snippet<[DetailSection<TEntity>, TEntity | null]>;
  fieldLabel?: Snippet<[DetailField<TEntity>]>;
  fieldValue?: Snippet<[DetailField<TEntity>, unknown, TEntity | null]>;
  fieldInput?: Snippet<[DetailField<TEntity>, unknown, TEntity | null]>;
}

// ============================================================================
// DETAIL EVENTS
// ============================================================================

/** Detail events */
export interface DetailEvents<TEntity> {
  onLoad?: (entity: TEntity) => void;
  onLoadError?: (error: DetailError) => void;
  onSave?: (entity: TEntity) => void;
  onSaveError?: (error: DetailError) => void;
  onDelete?: (entity: TEntity) => void;
  onDeleteError?: (error: DetailError) => void;
  onModeChange?: (mode: DetailMode) => void;
  onFieldChange?: (field: keyof TEntity, value: unknown, oldValue: unknown) => void;
  onDirtyChange?: (isDirty: boolean) => void;
  onValidationChange?: (isValid: boolean, errors: Record<string, string>) => void;
  onSectionToggle?: (sectionId: string, collapsed: boolean) => void;
  onRelatedDataLoad?: (key: string, data: unknown) => void;
}

// ============================================================================
// DETAIL ACTIONS
// ============================================================================

/** Standard detail action IDs */
export type StandardDetailAction =
  | 'edit'
  | 'save'
  | 'cancel'
  | 'delete'
  | 'duplicate'
  | 'archive'
  | 'restore'
  | 'print'
  | 'export'
  | 'share'
  | 'refresh';

/** Detail action group */
export interface DetailActionGroup<TEntity> {
  id: string;
  label?: string;
  actions: Action<TEntity>[];
  position?: 'header' | 'footer' | 'sidebar' | 'menu';
  variant?: 'primary' | 'secondary' | 'icon';
}
