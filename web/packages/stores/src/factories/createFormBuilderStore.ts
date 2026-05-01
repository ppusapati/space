/**
 * Form Builder Store Factory
 * Creates a store for managing the state of the visual form builder/designer component
 * (NOT for user-facing forms - use createFormStore for that)
 *
 * @packageDocumentation
 */

import { writable, derived, get, type Writable, type Readable } from 'svelte/store';
import type { FormFieldConfig, FormSchema } from '@samavāya/core';

// ============================================================================
// TYPES
// ============================================================================

export type FieldType = FormFieldConfig['type'];

export interface FormBuilderField {
  id: string;
  config: FormFieldConfig & { name: string; type: FieldType };
}

export interface FormBuilderSection {
  id: string;
  title?: string;
  fields: FormBuilderField[];
}

export interface FormBuilderLayout {
  type: 'vertical' | 'horizontal' | 'grid';
  columns?: number;
}

export interface FormBuilderState {
  sections: FormBuilderSection[];
  layout: FormBuilderLayout;
  selectedFieldId: string | null;
  selectedSectionId: string | null;
  isDragging: boolean;
  isPreviewMode: boolean;
}

export interface FormBuilderStoreState {
  state: FormBuilderState;
  errors: Record<string, string>;
  isDirty: boolean;
  lastSaved?: Date;
}

export interface FormBuilderStoreConfig {
  initialState?: FormBuilderState;
  schema?: FormSchema<Record<string, unknown>>;
}

export interface FormBuilderStoreReturn {
  // Subscribe to store
  subscribe: Writable<FormBuilderStoreState>['subscribe'];

  // Derived stores
  builderState: Readable<FormBuilderState>;
  selectedFieldId: Readable<string | null>;
  selectedSectionId: Readable<string | null>;
  errors: Readable<Record<string, string>>;
  isDirty: Readable<boolean>;
  isPreviewMode: Readable<boolean>;
  isDragging: Readable<boolean>;
  fieldCount: Readable<number>;
  sectionCount: Readable<number>;

  // Field operations
  addField: (sectionIndex: number, field: FormBuilderField) => void;
  removeField: (fieldId: string) => void;
  updateField: (fieldId: string, updates: Partial<FormBuilderField['config']>) => void;
  duplicateField: (fieldId: string, existingNames: string[]) => string | null;
  selectField: (fieldId: string | null) => void;
  deselectField: () => void;

  // Section operations
  addSection: (title?: string) => string;
  removeSection: (sectionId: string) => void;
  updateSection: (sectionId: string, updates: Partial<FormBuilderSection>) => void;
  selectSection: (sectionId: string | null) => void;
  deselectSection: () => void;

  // Layout operations
  updateLayout: (layout: Partial<FormBuilderLayout>) => void;

  // Validation and errors
  setErrors: (errors: Record<string, string>) => void;
  clearErrors: () => void;
  validate: () => Record<string, string>;

  // State management
  setState: (newState: FormBuilderState) => void;
  setDragging: (isDragging: boolean) => void;
  setPreviewMode: (isPreview: boolean) => void;
  markSaved: () => void;
  reset: () => void;

  // Getters
  getState: () => FormBuilderStoreState;
  getField: (fieldId: string) => FormBuilderField | null;
  getSection: (sectionId: string) => FormBuilderSection | null;
}

// ============================================================================
// FACTORY
// ============================================================================

const DEFAULT_INITIAL_STATE: FormBuilderState = {
  sections: [{ id: 'default', fields: [], title: 'Section 1' }],
  layout: { type: 'vertical', columns: 1 },
  selectedFieldId: null,
  selectedSectionId: null,
  isDragging: false,
  isPreviewMode: false,
};

export function createFormBuilderStore(config?: FormBuilderStoreConfig): FormBuilderStoreReturn {
  const initialState: FormBuilderStoreState = {
    state: config?.initialState || structuredClone(DEFAULT_INITIAL_STATE),
    errors: {},
    isDirty: false,
    lastSaved: undefined,
  };

  // ============================================================================
  // STORE
  // ============================================================================

  const store = writable<FormBuilderStoreState>(initialState);
  const { subscribe, set, update } = store;

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const builderState: Readable<FormBuilderState> = derived(store, ($s) => $s.state);
  const selectedFieldId: Readable<string | null> = derived(store, ($s) => $s.state.selectedFieldId);
  const selectedSectionId: Readable<string | null> = derived(store, ($s) => $s.state.selectedSectionId);
  const errors: Readable<Record<string, string>> = derived(store, ($s) => $s.errors);
  const isDirty: Readable<boolean> = derived(store, ($s) => $s.isDirty);
  const isPreviewMode: Readable<boolean> = derived(store, ($s) => $s.state.isPreviewMode);
  const isDragging: Readable<boolean> = derived(store, ($s) => $s.state.isDragging);

  const fieldCount: Readable<number> = derived(store, ($s) =>
    $s.state.sections.reduce((sum, section) => sum + section.fields.length, 0)
  );

  const sectionCount: Readable<number> = derived(store, ($s) => $s.state.sections.length);

  // ============================================================================
  // FIELD OPERATIONS
  // ============================================================================

  function addField(sectionIndex: number, field: FormBuilderField): void {
    update((current) => {
      const newSections = [...current.state.sections];
      if (newSections[sectionIndex]) {
        newSections[sectionIndex] = {
          ...newSections[sectionIndex],
          fields: [...newSections[sectionIndex].fields, field],
        };
      }
      return {
        ...current,
        state: { ...current.state, sections: newSections },
        isDirty: true,
      };
    });
  }

  function removeField(fieldId: string): void {
    update((current) => {
      const newSections = current.state.sections.map((section) => ({
        ...section,
        fields: section.fields.filter((f) => f.id !== fieldId),
      }));
      return {
        ...current,
        state: {
          ...current.state,
          sections: newSections,
          selectedFieldId: fieldId === current.state.selectedFieldId ? null : current.state.selectedFieldId,
        },
        isDirty: true,
      };
    });
  }

  function updateField(fieldId: string, updates: Partial<FormBuilderField['config']>): void {
    update((current) => {
      const newSections = current.state.sections.map((section) => ({
        ...section,
        fields: section.fields.map((f) => {
          if (f.id === fieldId) {
            return {
              ...f,
              config: { ...f.config, ...updates } as FormBuilderField['config'],
            } as FormBuilderField;
          }
          return f;
        }),
      })) as typeof current.state.sections;
      return {
        ...current,
        state: { ...current.state, sections: newSections },
        isDirty: true,
      };
    });
  }

  function duplicateField(fieldId: string, existingNames: string[]): string | null {
    let newFieldId: string | null = null;

    update((current) => {
      const newSections = current.state.sections.map((section) => {
        const fieldIndex = section.fields.findIndex((f) => f.id === fieldId);
        if (fieldIndex !== -1) {
          const originalField = section.fields[fieldIndex]!;
          newFieldId = `field_${originalField.config.type}_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`;

          // Generate unique field name
          let counter = 1;
          let newName = originalField.config.name;
          while (existingNames.includes(newName)) {
            newName = `${originalField.config.name}${counter}`;
            counter++;
          }

          const newField: FormBuilderField = {
            id: newFieldId,
            config: {
              ...originalField.config,
              name: newName,
              label: originalField.config.label ? `${originalField.config.label} (Copy)` : undefined,
            },
          };

          const newFields = [...section.fields];
          newFields.splice(fieldIndex + 1, 0, newField);

          return { ...section, fields: newFields };
        }
        return section;
      });

      return {
        ...current,
        state: { ...current.state, sections: newSections },
        isDirty: true,
      };
    });

    return newFieldId;
  }

  function selectField(fieldId: string | null): void {
    update((current) => ({
      ...current,
      state: {
        ...current.state,
        selectedFieldId: fieldId,
        selectedSectionId: null,
      },
    }));
  }

  function deselectField(): void {
    selectField(null);
  }

  // ============================================================================
  // SECTION OPERATIONS
  // ============================================================================

  function addSection(title?: string): string {
    let newSectionId = '';
    const timestamp = Date.now();
    const random = Math.random().toString(36).slice(2, 7);
    newSectionId = `section_${timestamp}_${random}`;

    update((current) => {
      const newSection: FormBuilderSection = {
        id: newSectionId,
        title: title || `Section ${current.state.sections.length + 1}`,
        fields: [],
      };

      return {
        ...current,
        state: {
          ...current.state,
          sections: [...current.state.sections, newSection],
        },
        isDirty: true,
      };
    });

    return newSectionId;
  }

  function removeSection(sectionId: string): void {
    update((current) => {
      const sectionToRemove = current.state.sections.find((s) => s.id === sectionId);
      let newSections = current.state.sections.filter((s) => s.id !== sectionId);

      // Move fields from removed section to first section if any exist
      if (sectionToRemove && sectionToRemove.fields.length > 0 && newSections.length > 0) {
        newSections[0] = {
          ...newSections[0]!,
          fields: [...newSections[0]!.fields, ...sectionToRemove.fields],
        };
      }

      return {
        ...current,
        state: {
          ...current.state,
          sections: newSections,
          selectedSectionId:
            sectionId === current.state.selectedSectionId ? null : current.state.selectedSectionId,
        },
        isDirty: true,
      };
    });
  }

  function updateSection(sectionId: string, updates: Partial<FormBuilderSection>): void {
    update((current) => {
      const newSections = current.state.sections.map((section) => {
        if (section.id === sectionId) {
          return { ...section, ...updates };
        }
        return section;
      });

      return {
        ...current,
        state: { ...current.state, sections: newSections },
        isDirty: true,
      };
    });
  }

  function selectSection(sectionId: string | null): void {
    update((current) => ({
      ...current,
      state: {
        ...current.state,
        selectedSectionId: sectionId,
        selectedFieldId: null,
      },
    }));
  }

  function deselectSection(): void {
    selectSection(null);
  }

  // ============================================================================
  // LAYOUT OPERATIONS
  // ============================================================================

  function updateLayout(layout: Partial<FormBuilderLayout>): void {
    update((current) => ({
      ...current,
      state: {
        ...current.state,
        layout: { ...current.state.layout, ...layout },
      },
      isDirty: true,
    }));
  }

  // ============================================================================
  // VALIDATION AND ERRORS
  // ============================================================================

  function setErrors(newErrors: Record<string, string>): void {
    update((current) => ({
      ...current,
      errors: newErrors,
    }));
  }

  function clearErrors(): void {
    update((current) => ({
      ...current,
      errors: {},
    }));
  }

  function validate(): Record<string, string> {
    const state = get(store);
    const errors: Record<string, string> = {};

    for (const section of state.state.sections) {
      for (const field of section.fields) {
        // Add validation logic as needed
        if (!field.config.name) {
          errors[field.id] = 'Field name is required';
        }
        if (!field.config.type) {
          errors[field.id] = 'Field type is required';
        }
      }
    }

    setErrors(errors);
    return errors;
  }

  // ============================================================================
  // STATE MANAGEMENT
  // ============================================================================

  function setState(newState: FormBuilderState): void {
    update((current) => ({
      ...current,
      state: newState,
      isDirty: true,
    }));
  }

  function setDragging(isDragging: boolean): void {
    update((current) => ({
      ...current,
      state: { ...current.state, isDragging },
    }));
  }

  function setPreviewMode(isPreview: boolean): void {
    update((current) => ({
      ...current,
      state: { ...current.state, isPreviewMode: isPreview },
    }));
  }

  function markSaved(): void {
    update((current) => ({
      ...current,
      isDirty: false,
      lastSaved: new Date(),
    }));
  }

  function reset(): void {
    set(initialState);
  }

  // ============================================================================
  // GETTERS
  // ============================================================================

  function getState(): FormBuilderStoreState {
    return get(store);
  }

  function getField(fieldId: string): FormBuilderField | null {
    const state = get(store);
    for (const section of state.state.sections) {
      const field = section.fields.find((f) => f.id === fieldId);
      if (field) return field;
    }
    return null;
  }

  function getSection(sectionId: string): FormBuilderSection | null {
    const state = get(store);
    return state.state.sections.find((s) => s.id === sectionId) || null;
  }

  // ============================================================================
  // RETURN
  // ============================================================================

  return {
    subscribe,
    // Derived stores
    builderState,
    selectedFieldId,
    selectedSectionId,
    errors,
    isDirty,
    isPreviewMode,
    isDragging,
    fieldCount,
    sectionCount,
    // Field operations
    addField,
    removeField,
    updateField,
    duplicateField,
    selectField,
    deselectField,
    // Section operations
    addSection,
    removeSection,
    updateSection,
    selectSection,
    deselectSection,
    // Layout operations
    updateLayout,
    // Validation and errors
    setErrors,
    clearErrors,
    validate,
    // State management
    setState,
    setDragging,
    setPreviewMode,
    markSaved,
    reset,
    // Getters
    getState,
    getField,
    getSection,
  };
}
