/**
 * useDetail Composable
 * Creates a reactive detail state for entity display and editing
 */

import { writable, derived, get, type Writable, type Readable } from 'svelte/store';
import type {
  Detail,
  DetailState,
  DetailMode,
  DetailConfig,
  DetailError,
  DetailSection,
  RelatedDataState,
  Action,
} from '../types/index.js';

// ============================================================================
// TYPES
// ============================================================================

export interface UseDetailOptions<TEntity, TActions extends string = string> {
  config: DetailConfig<TEntity>;
  sections?: DetailSection<TEntity>[];
  actions?: Record<TActions, Action<TEntity>>;
  fetchEntity?: (id: string | number) => Promise<TEntity>;
  saveEntity?: (entity: TEntity) => Promise<TEntity>;
  deleteEntity?: (entity: TEntity) => Promise<void>;
  onError?: (error: DetailError) => void;
  onSave?: (entity: TEntity) => void;
  onDelete?: (entity: TEntity) => void;
}

export interface UseDetailReturn<TEntity, TActions extends string = string> {
  // Stores
  entity: Writable<TEntity | null>;
  originalEntity: Writable<TEntity | null>;
  state: Writable<DetailState>;
  mode: Writable<DetailMode>;
  sections: Writable<DetailSection<TEntity>[]>;
  activeSectionId: Writable<string | undefined>;
  relatedData: Writable<Record<string, RelatedDataState<unknown>>>;
  validationErrors: Writable<Record<string, string>>;

  // Derived
  isDirty: Readable<boolean>;
  isValid: Readable<boolean>;
  changedFields: Readable<Partial<TEntity>>;

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
  duplicate: () => Promise<TEntity | null>;

  // Methods - Field operations
  setFieldValue: <K extends keyof TEntity>(field: K, value: TEntity[K]) => void;
  resetField: <K extends keyof TEntity>(field: K) => void;
  resetAll: () => void;
  getChangedFields: () => Partial<TEntity>;

  // Methods - Validation
  validate: () => Promise<boolean>;
  validateField: <K extends keyof TEntity>(field: K) => Promise<string | null>;
  setFieldError: (field: string, error: string | null) => void;
  clearErrors: () => void;

  // Methods - Actions
  executeAction: (actionId: TActions) => Promise<void>;

  // Methods - Related data
  loadRelatedData: (key: string, loader: () => Promise<unknown>) => Promise<void>;
  refreshRelatedData: (key: string, loader: () => Promise<unknown>) => Promise<void>;

  // Methods - Sections
  setActiveSection: (sectionId: string) => void;
  toggleSection: (sectionId: string) => void;
}

// ============================================================================
// IMPLEMENTATION
// ============================================================================

export function useDetail<TEntity, TActions extends string = string>(
  options: UseDetailOptions<TEntity, TActions>
): UseDetailReturn<TEntity, TActions> {
  const {
    config,
    sections: initialSections = [],
    actions = {} as Record<TActions, Action<TEntity>>,
    fetchEntity,
    saveEntity,
    deleteEntity,
    onError,
    onSave,
    onDelete,
  } = options;

  let currentEntityId: string | number | null = null;

  // ============================================================================
  // STORES
  // ============================================================================

  const entity = writable<TEntity | null>(null);
  const originalEntity = writable<TEntity | null>(null);

  const state = writable<DetailState>({
    isLoading: false,
    isSaving: false,
    isDeleting: false,
    isDirty: false,
    isValid: true,
    hasError: false,
  });

  const mode = writable<DetailMode>('view');
  const sections = writable<DetailSection<TEntity>[]>(initialSections);
  const activeSectionId = writable<string | undefined>(initialSections[0]?.id);
  const relatedData = writable<Record<string, RelatedDataState<unknown>>>({});
  const validationErrors = writable<Record<string, string>>({});

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const isDirty = derived([entity, originalEntity], ([$entity, $originalEntity]) => {
    if (!$entity || !$originalEntity) return false;
    return JSON.stringify($entity) !== JSON.stringify($originalEntity);
  });

  const isValid = derived(validationErrors, ($errors) => Object.keys($errors).length === 0);

  const changedFields = derived([entity, originalEntity], ([$entity, $originalEntity]) => {
    if (!$entity || !$originalEntity) return {};

    const changes: Partial<TEntity> = {};
    for (const key of Object.keys($entity) as (keyof TEntity)[]) {
      if (JSON.stringify($entity[key]) !== JSON.stringify($originalEntity[key])) {
        changes[key] = $entity[key];
      }
    }
    return changes;
  });

  // ============================================================================
  // SUBSCRIPTIONS
  // ============================================================================

  // Update state.isDirty when isDirty changes
  isDirty.subscribe(($isDirty) => {
    state.update(($s) => ({ ...$s, isDirty: $isDirty }));
  });

  // Update state.isValid when isValid changes
  isValid.subscribe(($isValid) => {
    state.update(($s) => ({ ...$s, isValid: $isValid }));
  });

  // ============================================================================
  // DATA LOADING
  // ============================================================================

  async function load(id: string | number): Promise<void> {
    if (!fetchEntity) {
      console.warn('fetchEntity not provided');
      return;
    }

    currentEntityId = id;
    state.update(($s) => ({ ...$s, isLoading: true, hasError: false, error: undefined }));

    try {
      const data = await fetchEntity(id);
      entity.set(data);
      originalEntity.set(structuredClone(data));
      validationErrors.set({});

      state.update(($s) => ({
        ...$s,
        isLoading: false,
        lastModified: new Date(),
      }));
    } catch (error) {
      const detailError: DetailError = {
        code: 'LOAD_ERROR',
        message: error instanceof Error ? error.message : 'Failed to load entity',
        retryable: true,
      };

      state.update(($s) => ({ ...$s, isLoading: false, hasError: true, error: detailError }));
      onError?.(detailError);
    }
  }

  async function reload(): Promise<void> {
    if (currentEntityId != null) {
      await load(currentEntityId);
    }
  }

  async function refresh(): Promise<void> {
    await reload();
  }

  // ============================================================================
  // MODE SWITCHING
  // ============================================================================

  function setMode(newMode: DetailMode): void {
    mode.set(newMode);
  }

  function edit(): void {
    if (!config.editable) return;
    mode.set('edit');
  }

  function view(): void {
    // Reset to original if dirty
    const $isDirty = get(isDirty);
    if ($isDirty) {
      const $original = get(originalEntity);
      if ($original) {
        entity.set(structuredClone($original));
      }
    }
    validationErrors.set({});
    mode.set('view');
  }

  function create(initialData?: Partial<TEntity>): void {
    currentEntityId = null;
    entity.set((initialData ?? {}) as TEntity);
    originalEntity.set((initialData ?? {}) as TEntity);
    validationErrors.set({});
    mode.set('create');
  }

  // ============================================================================
  // ENTITY OPERATIONS
  // ============================================================================

  async function save(): Promise<void> {
    if (!saveEntity) {
      console.warn('saveEntity not provided');
      return;
    }

    const $entity = get(entity);
    if (!$entity) return;

    // Validate before saving
    const isValidResult = await validate();
    if (!isValidResult) return;

    state.update(($s) => ({ ...$s, isSaving: true }));

    try {
      const savedEntity = await saveEntity($entity);
      entity.set(savedEntity);
      originalEntity.set(structuredClone(savedEntity));

      // Update ID if this was a create
      if (get(mode) === 'create') {
        currentEntityId = savedEntity[config.entityKey] as string | number;
      }

      state.update(($s) => ({
        ...$s,
        isSaving: false,
        lastSaved: new Date(),
      }));

      onSave?.(savedEntity);
      mode.set('view');
    } catch (error) {
      const detailError: DetailError = {
        code: 'SAVE_ERROR',
        message: error instanceof Error ? error.message : 'Failed to save entity',
        retryable: true,
      };

      state.update(($s) => ({ ...$s, isSaving: false, hasError: true, error: detailError }));
      onError?.(detailError);
    }
  }

  async function deleteEntityFn(): Promise<void> {
    if (!deleteEntity) {
      console.warn('deleteEntity not provided');
      return;
    }

    const $entity = get(entity);
    if (!$entity) return;

    state.update(($s) => ({ ...$s, isDeleting: true }));

    try {
      await deleteEntity($entity);

      state.update(($s) => ({ ...$s, isDeleting: false }));
      onDelete?.($entity);

      // Clear entity after delete
      entity.set(null);
      originalEntity.set(null);
      currentEntityId = null;
    } catch (error) {
      const detailError: DetailError = {
        code: 'DELETE_ERROR',
        message: error instanceof Error ? error.message : 'Failed to delete entity',
        retryable: true,
      };

      state.update(($s) => ({ ...$s, isDeleting: false, hasError: true, error: detailError }));
      onError?.(detailError);
    }
  }

  async function duplicate(): Promise<TEntity | null> {
    const $entity = get(entity);
    if (!$entity) return null;

    // Create a copy without the ID
    const duplicated = structuredClone($entity);
    delete (duplicated as Record<string, unknown>)[config.entityKey as string];

    create(duplicated);
    return duplicated;
  }

  // ============================================================================
  // FIELD OPERATIONS
  // ============================================================================

  function setFieldValue<K extends keyof TEntity>(field: K, value: TEntity[K]): void {
    entity.update(($e) => {
      if (!$e) return $e;
      return { ...$e, [field]: value };
    });
  }

  function resetField<K extends keyof TEntity>(field: K): void {
    const $original = get(originalEntity);
    if (!$original) return;

    entity.update(($e) => {
      if (!$e) return $e;
      return { ...$e, [field]: $original[field] };
    });

    validationErrors.update(($errors) => {
      const newErrors = { ...$errors };
      delete newErrors[field as string];
      return newErrors;
    });
  }

  function resetAll(): void {
    const $original = get(originalEntity);
    if ($original) {
      entity.set(structuredClone($original));
    }
    validationErrors.set({});
  }

  function getChangedFields(): Partial<TEntity> {
    return get(changedFields);
  }

  // ============================================================================
  // VALIDATION
  // ============================================================================

  async function validate(): Promise<boolean> {
    const $entity = get(entity);
    const $sections = get(sections);

    if (!$entity) return false;

    const errors: Record<string, string> = {};

    // Validate all fields in all sections
    for (const section of $sections) {
      for (const field of section.fields) {
        if (field.validation) {
          const value = ($entity as Record<string, unknown>)[field.key as string];

          for (const rule of field.validation) {
            try {
              const isValid = await rule.validate(value, $entity);
              if (!isValid) {
                errors[field.key as string] =
                  typeof rule.message === 'function' ? rule.message(value, $entity) : rule.message;
                break;
              }
            } catch {
              errors[field.key as string] = 'Validation failed';
              break;
            }
          }
        }

        // Check required
        if (field.required) {
          const isRequired =
            typeof field.required === 'function' ? field.required($entity) : field.required;

          if (isRequired) {
            const value = ($entity as Record<string, unknown>)[field.key as string];
            if (value == null || value === '') {
              errors[field.key as string] = `${field.label} is required`;
            }
          }
        }
      }
    }

    validationErrors.set(errors);
    return Object.keys(errors).length === 0;
  }

  async function validateField<K extends keyof TEntity>(field: K): Promise<string | null> {
    const $entity = get(entity);
    const $sections = get(sections);

    if (!$entity) return null;

    // Find field definition
    let fieldDef = null;
    for (const section of $sections) {
      fieldDef = section.fields.find((f) => f.key === field);
      if (fieldDef) break;
    }

    if (!fieldDef) return null;

    const value = $entity[field];

    // Check required
    if (fieldDef.required) {
      const isRequired =
        typeof fieldDef.required === 'function' ? fieldDef.required($entity) : fieldDef.required;

      if (isRequired && (value == null || value === '')) {
        const error = `${fieldDef.label} is required`;
        setFieldError(field as string, error);
        return error;
      }
    }

    // Run validation rules
    if (fieldDef.validation) {
      for (const rule of fieldDef.validation) {
        try {
          const isValid = await rule.validate(value, $entity);
          if (!isValid) {
            const error =
              typeof rule.message === 'function' ? rule.message(value, $entity) : rule.message;
            setFieldError(field as string, error);
            return error;
          }
        } catch {
          const error = 'Validation failed';
          setFieldError(field as string, error);
          return error;
        }
      }
    }

    setFieldError(field as string, null);
    return null;
  }

  function setFieldError(field: string, error: string | null): void {
    validationErrors.update(($errors) => {
      if (error === null) {
        const newErrors = { ...$errors };
        delete newErrors[field];
        return newErrors;
      }
      return { ...$errors, [field]: error };
    });
  }

  function clearErrors(): void {
    validationErrors.set({});
  }

  // ============================================================================
  // ACTIONS
  // ============================================================================

  async function executeAction(actionId: TActions): Promise<void> {
    const action = actions[actionId];
    if (!action) {
      console.warn(`Action ${actionId} not found`);
      return;
    }

    const $entity = get(entity);
    if (!$entity) return;

    // Check if disabled
    if (typeof action.disabled === 'function') {
      if (action.disabled($entity)) return;
    } else if (action.disabled) {
      return;
    }

    await action.handler($entity);
  }

  // ============================================================================
  // RELATED DATA
  // ============================================================================

  async function loadRelatedData(key: string, loader: () => Promise<unknown>): Promise<void> {
    relatedData.update(($rd) => ({
      ...$rd,
      [key]: { data: null, isLoading: true, hasError: false },
    }));

    try {
      const data = await loader();
      relatedData.update(($rd) => ({
        ...$rd,
        [key]: { data, isLoading: false, hasError: false, lastLoaded: new Date() },
      }));
    } catch (error) {
      relatedData.update(($rd) => ({
        ...$rd,
        [key]: {
          data: null,
          isLoading: false,
          hasError: true,
          error: {
            code: 'LOAD_ERROR',
            message: error instanceof Error ? error.message : 'Failed to load related data',
          },
        },
      }));
    }
  }

  async function refreshRelatedData(key: string, loader: () => Promise<unknown>): Promise<void> {
    await loadRelatedData(key, loader);
  }

  // ============================================================================
  // SECTIONS
  // ============================================================================

  function setActiveSection(sectionId: string): void {
    activeSectionId.set(sectionId);
  }

  function toggleSection(sectionId: string): void {
    sections.update(($sections) =>
      $sections.map((s) =>
        s.id === sectionId ? { ...s, defaultCollapsed: !s.defaultCollapsed } : s
      )
    );
  }

  // ============================================================================
  // RETURN
  // ============================================================================

  return {
    // Stores
    entity,
    originalEntity,
    state,
    mode,
    sections,
    activeSectionId,
    relatedData,
    validationErrors,

    // Derived
    isDirty,
    isValid,
    changedFields,

    // Methods - Data loading
    load,
    reload,
    refresh,

    // Methods - Mode switching
    setMode,
    edit,
    view,
    create,

    // Methods - Entity operations
    save,
    delete: deleteEntityFn,
    duplicate,

    // Methods - Field operations
    setFieldValue,
    resetField,
    resetAll,
    getChangedFields,

    // Methods - Validation
    validate,
    validateField,
    setFieldError,
    clearErrors,

    // Methods - Actions
    executeAction,

    // Methods - Related data
    loadRelatedData,
    refreshRelatedData,

    // Methods - Sections
    setActiveSection,
    toggleSection,
  };
}
