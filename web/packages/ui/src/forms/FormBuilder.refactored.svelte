<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import Icon from '../display/Icon.svelte';
  import Input from './Input.svelte';
  import Select from './Select.svelte';
  import Checkbox from './Checkbox.svelte';
  import TextArea from './TextArea.svelte';
  import { sortable, reorderItems, draggable, droppable } from '../actions';
  import type { FormFieldConfig, FormSchema } from '@samavāya/core';
  import {
    type FormBuilderProps,
    type FormBuilderState,
    type FormBuilderField,
    type FormBuilderSection,
    type FieldPaletteCategory,
    type FieldPaletteItem,
    defaultFieldCategories,
    formBuilderClasses,
    generateFieldId,
    generateFieldName,
    builderStateToSchema,
    schemaToBuilderState,
    validateFieldConfig,
    findDuplicateNames,
    getFieldProperties,
  } from './formbuilder.types';
  import type { Size } from '../types';

  // Design tokens (local fallback - these mirror the design system's color/spacing scale)
  // Defined as const object literals so TypeScript knows each key's value (avoiding noUncheckedIndexedAccess issues)
  const tokens = {
    colors: {
      neutral: { 0: '#ffffff', 50: '#f9fafb', 100: '#f3f4f6', 200: '#e5e7eb', 300: '#d1d5db', 400: '#9ca3af', 500: '#6b7280', 700: '#374151', 900: '#111827' },
      primary: { 50: '#eef2ff', 100: '#e0e7ff', 300: '#a5b4fc', 500: '#6366f1', 700: '#4338ca' },
      error: { 50: '#fef2f2', 500: '#ef4444' },
    },
    spacing: { 1: '0.25rem', 1.5: '0.375rem', 2: '0.5rem', 3: '0.75rem', 4: '1rem', 6: '1.5rem' },
  };

  // Form state management (use createFormBuilderStore from @samavāya/stores)
  import { createFormBuilderStore } from '@samavāya/stores';
  const _builderStore = createFormBuilderStore();
  const formBuilder = _builderStore;
  const setFormBuilderState = (s: any) => _builderStore.setState(s);
  const addFormBuilderField = (idx: number, f: any) => _builderStore.addField(idx, f);
  const removeFormBuilderField = (id: string) => _builderStore.removeField(id);
  const updateFormBuilderField = (id: string, key: string, val: unknown) => _builderStore.updateField(id, { [key]: val } as any);
  const duplicateFormBuilderField = (id: string, names: string[]) => _builderStore.duplicateField(id, names);
  const addFormBuilderSection = (title?: string) => _builderStore.addSection(title);
  const removeFormBuilderSection = (id: string) => _builderStore.removeSection(id);
  const updateFormBuilderSection = (id: string, key: string, val: unknown) => _builderStore.updateSection(id, { [key]: val } as any);
  const selectFormBuilderField = (id: string | null) => _builderStore.selectField(id);
  const selectFormBuilderSection = (id: string | null) => _builderStore.selectSection(id);
  const setFormBuilderErrors = (errs: Record<string, string>) => _builderStore.setErrors(errs);
  const clearFormBuilderErrors = () => _builderStore.clearErrors();

  // Props
  export let schema: FormSchema<Record<string, unknown>> | undefined = undefined;
  export let availableFields: FormFieldConfig['type'][] | undefined = undefined;
  export let allowSections: boolean = true;
  export let maxFields: number | undefined = undefined;
  export let size: Size = 'md';
  export let readonly: boolean = false;
  export let showPalette: boolean = true;
  export let showProperties: boolean = true;
  export let showPreview: boolean = true;
  export let fieldCategories: FieldPaletteCategory[] | undefined = undefined;

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: { schema: FormSchema<Record<string, unknown>>; state: FormBuilderState };
    select: { fieldId: string | null; sectionId: string | null };
    error: { errors: string[] };
  }>();

  // Local state for UI interactions
  let paletteCollapsed: Set<string> = new Set();

  // Subscribe to form store
  let state: FormBuilderState = {
    sections: [{ id: 'default', fields: [] }],
    layout: { type: 'vertical', columns: 1 },
    selectedFieldId: null,
    selectedSectionId: null,
    isDragging: false,
    isPreviewMode: false,
  };

  let formErrors: Map<string, string> = new Map();

  // Subscribe to store changes
  const unsubscribe = formBuilder.subscribe((value: any) => {
    state = value.state;
    formErrors = new Map(Object.entries(value.errors || {}));
  });

  // Cleanup on destroy
  onMount(() => {
    if (schema) {
      setFormBuilderState(schemaToBuilderState(schema));
    }

    return () => {
      unsubscribe();
    };
  });

  // Get effective field categories
  $: effectiveCategories = filterCategories(fieldCategories || defaultFieldCategories, availableFields);

  // Get all field names for duplicate checking
  $: allFieldNames = state.sections.flatMap((s) => s.fields.map((f) => f.config.name));

  // Selected field
  $: selectedField = state.selectedFieldId
    ? state.sections.flatMap((s) => s.fields).find((f) => f.id === state.selectedFieldId)
    : null;

  // Selected section
  $: selectedSection = state.selectedSectionId
    ? state.sections.find((s) => s.id === state.selectedSectionId)
    : null;

  // Total field count
  $: totalFields = state.sections.reduce((sum, s) => sum + s.fields.length, 0);

  // Check max fields limit
  $: canAddFields = maxFields === undefined || totalFields < maxFields;

  // Field properties for selected field
  $: fieldProperties = selectedField ? getFieldProperties(selectedField.config.type) : [];

  // Group properties by category
  $: propertyGroups = groupProperties(fieldProperties);

  function filterCategories(
    categories: FieldPaletteCategory[],
    allowed?: FormFieldConfig['type'][]
  ): FieldPaletteCategory[] {
    if (!allowed) return categories;

    return categories
      .map((cat) => ({
        ...cat,
        fields: cat.fields.filter((f) => allowed.includes(f.type)),
      }))
      .filter((cat) => cat.fields.length > 0);
  }

  function groupProperties(properties: typeof fieldProperties): Map<string, typeof fieldProperties> {
    const groups = new Map<string, typeof fieldProperties>();

    for (const prop of properties) {
      const group = prop.group || 'basic';
      if (!groups.has(group)) {
        groups.set(group, []);
      }
      groups.get(group)!.push(prop);
    }

    return groups;
  }

  // Notify changes
  function notifyChange() {
    const outputSchema = builderStateToSchema(state);
    dispatch('change', { schema: outputSchema, state });
  }

  // Validate all fields
  function validateFields() {
    clearFormBuilderErrors();

    const errors: Record<string, string> = {};
    const duplicates = findDuplicateNames(state.sections);

    for (const section of state.sections) {
      for (const field of section.fields) {
        const error = validateFieldConfig(field.config);
        if (error) {
          errors[field.id] = error;
        } else if (duplicates.includes(field.config.name)) {
          errors[field.id] = `Duplicate field name: ${field.config.name}`;
        }
      }
    }

    if (Object.keys(errors).length > 0) {
      setFormBuilderErrors(errors);
      dispatch('error', { errors: Object.values(errors) });
    }
  }

  // Toggle palette category
  function toggleCategory(categoryId: string) {
    if (paletteCollapsed.has(categoryId)) {
      paletteCollapsed.delete(categoryId);
    } else {
      paletteCollapsed.add(categoryId);
    }
    paletteCollapsed = paletteCollapsed;
  }

  // Select field
  function selectField(fieldId: string | null) {
    selectFormBuilderField(fieldId);
    dispatch('select', { fieldId, sectionId: null });
  }

  // Select section
  function selectSection(sectionId: string | null) {
    selectFormBuilderSection(sectionId);
    dispatch('select', { fieldId: null, sectionId });
  }

  // Add field from palette
  function addField(item: FieldPaletteItem, sectionIndex: number = 0) {
    if (!canAddFields || readonly) return;

    const newField: FormBuilderField = {
      id: generateFieldId(item.type),
      config: {
        ...item.defaultConfig,
        name: generateFieldName(item.type, allFieldNames),
      } as FormFieldConfig,
    };

    addFormBuilderField(sectionIndex, newField);
    selectField(newField.id);
    validateFields();
    notifyChange();
  }

  // Remove field
  function removeField(fieldId: string) {
    if (readonly) return;

    removeFormBuilderField(fieldId);

    if (state.selectedFieldId === fieldId) {
      selectField(null);
    }

    validateFields();
    notifyChange();
  }

  // Duplicate field
  function duplicateField(fieldId: string) {
    if (!canAddFields || readonly) return;

    const newFieldId = duplicateFormBuilderField(fieldId, allFieldNames);
    if (newFieldId) {
      selectField(newFieldId);
      validateFields();
      notifyChange();
    }
  }

  // Update field property
  function updateFieldProperty(fieldId: string, key: string, value: unknown) {
    if (readonly) return;

    updateFormBuilderField(fieldId, key, value);
    validateFields();
    notifyChange();
  }

  // Add section
  function addSection() {
    if (!allowSections || readonly) return;

    const newSectionId = addFormBuilderSection();
    if (newSectionId) {
      selectSection(newSectionId);
      notifyChange();
    }
  }

  // Remove section
  function removeSection(sectionId: string) {
    if (readonly || state.sections.length <= 1) return;

    removeFormBuilderSection(sectionId);

    if (state.selectedSectionId === sectionId) {
      selectSection(null);
    }

    notifyChange();
  }

  // Update section property
  function updateSectionProperty(sectionId: string, key: string, value: unknown) {
    if (readonly) return;

    updateFormBuilderSection(sectionId, key, value);
    notifyChange();
  }

  // Handle field reorder within section
  function handleFieldSort(sectionId: string, result: { sourceIndex: number; destinationIndex: number }) {
    if (readonly) return;

    const section = state.sections.find((s) => s.id === sectionId);
    if (section) {
      section.fields = reorderItems(section.fields, result.sourceIndex, result.destinationIndex);
      setFormBuilderState(state);
      notifyChange();
    }
  }

  // Handle field drop from palette
  function handlePaletteDrop(item: FieldPaletteItem, sectionId: string) {
    if (!canAddFields || readonly) return;

    const sectionIndex = state.sections.findIndex((s) => s.id === sectionId);
    if (sectionIndex !== -1) {
      addField(item, sectionIndex);
    }
  }

  // Toggle preview mode
  function togglePreview() {
    const newState = { ...state, isPreviewMode: !state.isPreviewMode };
    setFormBuilderState(newState);
  }

  // Export schema
  export function getSchema(): FormSchema<Record<string, unknown>> {
    return builderStateToSchema(state);
  }

  // Import schema
  export function setSchema(newSchema: FormSchema<Record<string, unknown>>) {
    setFormBuilderState(schemaToBuilderState(newSchema));
    validateFields();
    notifyChange();
  }

  // Get field label display
  function getFieldLabel(field: FormBuilderField): string {
    return field.config.label || field.config.name || 'Unnamed Field';
  }

  // Get field type display
  function getFieldTypeLabel(type: FormFieldConfig['type']): string {
    const labels: Record<string, string> = {
      text: 'Text',
      email: 'Email',
      password: 'Password',
      tel: 'Phone',
      url: 'URL',
      search: 'Search',
      number: 'Number',
      textarea: 'Textarea',
      select: 'Select',
      checkbox: 'Checkbox',
      'checkbox-group': 'Checkbox Group',
      radio: 'Radio',
      switch: 'Switch',
      date: 'Date',
      datetime: 'Date & Time',
      time: 'Time',
      month: 'Month',
      year: 'Year',
      daterange: 'Date Range',
      file: 'File Upload',
      richtext: 'Rich Text',
      autocomplete: 'Autocomplete',
      color: 'Color',
      slider: 'Slider',
      rating: 'Rating',
      array: 'Repeater',
      object: 'Field Group',
      hidden: 'Hidden',
      custom: 'Custom',
    };
    return labels[type] || type;
  }

  // Property group labels
  const propertyGroupLabels: Record<string, string> = {
    basic: 'Basic',
    validation: 'Validation',
    display: 'Display',
    behavior: 'Behavior',
    options: 'Options',
    state: 'State',
    advanced: 'Advanced',
  };

  // Build design token styles
  const containerStyle = `
    display: flex;
    height: 100%;
    background-color: ${tokens.colors.neutral[0]};
    border: 1px solid ${tokens.colors.neutral[200]};
    border-radius: 0.5rem;
    overflow: hidden;
  `;

  const paletteStyle = `
    width: 16rem;
    border-right: 1px solid ${tokens.colors.neutral[200]};
    background-color: ${tokens.colors.neutral[50]};
    overflow-y: auto;
    flex-shrink: 0;
  `;

  const canvasStyle = `
    flex: 1;
    overflow-y: auto;
    padding: ${tokens.spacing[6]};
    background-color: ${tokens.colors.neutral[100]};
  `;

  const propertiesStyle = `
    width: 20rem;
    border-left: 1px solid ${tokens.colors.neutral[200]};
    background-color: ${tokens.colors.neutral[0]};
    overflow-y: auto;
    flex-shrink: 0;
  `;
</script>

<div class={cn('form-builder', className)} style={containerStyle}>
  <!-- Field Palette -->
  {#if showPalette && !state.isPreviewMode}
    <div style={paletteStyle}>
      <div style={`
        padding: ${tokens.spacing[3]} ${tokens.spacing[4]};
        border-bottom: 1px solid ${tokens.colors.neutral[200]};
        background-color: ${tokens.colors.neutral[0]};
        font-weight: 500;
        color: ${tokens.colors.neutral[900]};
        display: flex;
        justify-content: space-between;
        align-items: center;
      `}>
        <span>Fields</span>
        {#if maxFields}
          <span style={`font-size: 0.875rem; color: ${tokens.colors.neutral[500]}`}>
            {totalFields}/{maxFields}
          </span>
        {/if}
      </div>

      {#each effectiveCategories as category}
        <div style={`border-bottom: 1px solid ${tokens.colors.neutral[100]}`}>
          <button
            type="button"
            style={`
              width: 100%;
              display: flex;
              align-items: center;
              justify-content: space-between;
              padding: ${tokens.spacing[2]} ${tokens.spacing[4]};
              background: none;
              border: none;
              cursor: pointer;
              font-size: 0.875rem;
              font-weight: 500;
              color: ${tokens.colors.neutral[700]};
              transition: background-color 0.15s;
            `}
            on:mouseover={(e) => {
              e.currentTarget.style.backgroundColor = tokens.colors.neutral[100];
            }}
            on:mouseout={(e) => {
              e.currentTarget.style.backgroundColor = 'transparent';
            }}
            on:click={() => toggleCategory(category.id)}
          >
            <span style="display: flex; align-items: center; gap: 0.5rem">
              {#if category.icon}
                <Icon name={category.icon} size="sm" />
              {/if}
              {category.label}
            </span>
            <Icon
              name={paletteCollapsed.has(category.id) ? 'chevron-right' : 'chevron-down'}
              size="sm"
            />
          </button>

          {#if !paletteCollapsed.has(category.id)}
            <div style={`padding: ${tokens.spacing[2]}`}>
              {#each category.fields as item}
                <div
                  style={`
                    display: flex;
                    align-items: center;
                    gap: ${tokens.spacing[2]};
                    padding: ${tokens.spacing[2]} ${tokens.spacing[3]};
                    border-radius: 0.375rem;
                    background-color: ${tokens.colors.neutral[0]};
                    border: 1px solid ${tokens.colors.neutral[200]};
                    cursor: grab;
                    font-size: 0.875rem;
                    transition: all 0.15s;
                  `}
                  draggable="true"
                  role="button"
                  tabindex="0"
                  on:mouseover={(e) => {
                    e.currentTarget.style.borderColor = tokens.colors.primary[300];
                    e.currentTarget.style.backgroundColor = tokens.colors.primary[50];
                  }}
                  on:mouseout={(e) => {
                    e.currentTarget.style.borderColor = tokens.colors.neutral[200];
                    e.currentTarget.style.backgroundColor = tokens.colors.neutral[0];
                  }}
                  on:dragstart={(e) => {
                    e.dataTransfer?.setData('application/form-builder-field', JSON.stringify(item));
                    const newState = { ...state, isDragging: true };
                    setFormBuilderState(newState);
                  }}
                  on:dragend={() => {
                    const newState = { ...state, isDragging: false };
                    setFormBuilderState(newState);
                  }}
                  on:click={() => addField(item)}
                  on:keydown={(e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                      e.preventDefault();
                      addField(item);
                    }
                  }}
                  title={item.description || item.label}
                >
                  <span style={`color: ${tokens.colors.neutral[500]}`}><Icon name={item.icon} size="sm" /></span>
                  <span style={`color: ${tokens.colors.neutral[700]}`}>{item.label}</span>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}

  <!-- Main Canvas -->
  <div style="display: flex; flex-direction: column; overflow: hidden; flex: 1">
    <!-- Toolbar -->
    <div style={`
      display: flex;
      align-items: center;
      gap: ${tokens.spacing[2]};
      padding: ${tokens.spacing[2]} ${tokens.spacing[4]};
      border-bottom: 1px solid ${tokens.colors.neutral[200]};
      background-color: ${tokens.colors.neutral[0]};
    `}>
      {#if allowSections && !readonly && !state.isPreviewMode}
        <button
          type="button"
          style={`
            padding: ${tokens.spacing[1.5]} ${tokens.spacing[3]};
            font-size: 0.875rem;
            border-radius: 0.375rem;
            background: none;
            border: none;
            cursor: pointer;
            color: ${tokens.colors.neutral[700]};
            transition: background-color 0.15s;
          `}
          on:mouseover={(e) => {
            e.currentTarget.style.backgroundColor = tokens.colors.neutral[100];
          }}
          on:mouseout={(e) => {
            e.currentTarget.style.backgroundColor = 'transparent';
          }}
          on:click={addSection}
        >
          <span style="display: inline; margin-right: 0.25rem"><Icon name="plus" size="sm" /></span>
          Add Section
        </button>
      {/if}

      <div style="flex: 1"></div>

      {#if showPreview}
        <button
          type="button"
          style={`
            padding: ${tokens.spacing[1.5]} ${tokens.spacing[3]};
            font-size: 0.875rem;
            border-radius: 0.375rem;
            background: none;
            border: none;
            cursor: pointer;
            transition: all 0.15s;
            color: ${state.isPreviewMode ? tokens.colors.primary[700] : tokens.colors.neutral[700]};
            background-color: ${state.isPreviewMode ? tokens.colors.primary[100] : 'transparent'};
          `}
          on:mouseover={(e) => {
            if (!state.isPreviewMode) {
              e.currentTarget.style.backgroundColor = tokens.colors.neutral[100];
            }
          }}
          on:mouseout={(e) => {
            if (!state.isPreviewMode) {
              e.currentTarget.style.backgroundColor = 'transparent';
            }
          }}
          on:click={togglePreview}
        >
          <span style="display: inline; margin-right: 0.25rem"><Icon
            name={state.isPreviewMode ? 'edit' : 'eye'}
            size="sm"
          /></span>
          {state.isPreviewMode ? 'Edit' : 'Preview'}
        </button>
      {/if}
    </div>

    <!-- Canvas Area -->
    <div style={canvasStyle}>
      {#if state.sections.every((s) => s.fields.length === 0)}
        <!-- Empty State -->
        <div style={`
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          height: 100%;
          text-align: center;
          color: ${tokens.colors.neutral[500]};
          min-height: 12.5rem;
          background-color: ${tokens.colors.neutral[0]};
          border-radius: 0.5rem;
          border: 2px dashed ${tokens.colors.neutral[300]};
        `}>
          <span style={`
            width: 4rem;
            height: 4rem;
            margin-bottom: 1rem;
            color: ${tokens.colors.neutral[300]};
            display: inline-block;
          `}><Icon name="layout" /></span>
          <p style={`font-size: 1.125rem; font-weight: 500; margin-bottom: 0.5rem`}>Start building your form</p>
          <p style={`font-size: 0.875rem`}>Drag fields from the palette or click to add</p>
        </div>
      {:else}
        <!-- Sections -->
        {#each state.sections as section, sectionIndex (section.id)}
          <div
            style={`
              background-color: ${tokens.colors.neutral[0]};
              border-radius: 0.5rem;
              border: 1px solid ${tokens.colors.neutral[200]};
              margin-bottom: ${tokens.spacing[4]};
              min-height: 12.5rem;
              transition: all 0.15s;
              ${state.selectedSectionId === section.id ? `ring: 2px solid ${tokens.colors.primary[500]};` : ''}
            `}
            role="region"
            aria-label={section.title || `Section ${sectionIndex + 1}`}
            on:click|self={() => selectSection(section.id)}
            on:keydown={(e) => {
              if (e.key === 'Enter' && e.target === e.currentTarget) {
                selectSection(section.id);
              }
            }}
            on:dragover={(e) => {
              e.preventDefault();
              e.currentTarget.style.backgroundColor = tokens.colors.primary[50];
            }}
            on:dragleave={(e) => {
              e.currentTarget.style.backgroundColor = tokens.colors.neutral[0];
            }}
            on:drop={(e) => {
              e.preventDefault();
              e.currentTarget.style.backgroundColor = tokens.colors.neutral[0];
              const data = e.dataTransfer?.getData('application/form-builder-field');
              if (data) {
                const item = JSON.parse(data) as FieldPaletteItem;
                handlePaletteDrop(item, section.id);
              }
            }}
          >
            {#if section.title || allowSections}
              <div style={`
                padding: ${tokens.spacing[3]} ${tokens.spacing[4]};
                border-bottom: 1px solid ${tokens.colors.neutral[200]};
                display: flex;
                align-items: center;
                justify-content: space-between;
              `}>
                <div style={`font-weight: 500; color: ${tokens.colors.neutral[900]}`}>
                  {#if !state.isPreviewMode && state.selectedSectionId === section.id}
                    <input
                      type="text"
                      value={section.title || ''}
                      style={`
                        background: transparent;
                        border: none;
                        border-bottom: 1px solid ${tokens.colors.neutral[300]};
                        padding: 0 ${tokens.spacing[1]};
                        font-weight: 500;
                        color: ${tokens.colors.neutral[900]};
                      `}
                      placeholder="Section title"
                      on:input={(e) => updateSectionProperty(section.id, 'title', e.currentTarget.value)}
                    />
                  {:else}
                    {section.title || `Section ${sectionIndex + 1}`}
                  {/if}
                </div>

                {#if !state.isPreviewMode && !readonly}
                  <div style="display: flex; align-items: center; gap: 0.25rem">
                    {#if state.sections.length > 1}
                      <button
                        type="button"
                        style={`
                          padding: ${tokens.spacing[1]};
                          border-radius: 0.375rem;
                          background: none;
                          border: none;
                          cursor: pointer;
                          color: ${tokens.colors.neutral[500]};
                          transition: all 0.15s;
                        `}
                        on:mouseover={(e) => {
                          e.currentTarget.style.backgroundColor = tokens.colors.neutral[100];
                          e.currentTarget.style.color = tokens.colors.neutral[700];
                        }}
                        on:mouseout={(e) => {
                          e.currentTarget.style.backgroundColor = 'transparent';
                          e.currentTarget.style.color = tokens.colors.neutral[500];
                        }}
                        on:click|stopPropagation={() => removeSection(section.id)}
                        title="Remove section"
                      >
                        <Icon name="trash" size="sm" />
                      </button>
                    {/if}
                  </div>
                {/if}
              </div>
            {/if}

            <div
              style={`
                padding: ${tokens.spacing[4]};
                ${section.fields.length === 0 ? `min-height: 6.25rem;` : ''}
              `}
              use:sortable={{
                id: section.id,
                type: 'form-builder-field',
                itemSelector: '[data-field-id]',
                itemIdAttribute: 'data-field-id',
                enabled: !state.isPreviewMode && !readonly,
                onSortEnd: (result) => handleFieldSort(section.id, result),
              }}
            >
              {#each section.fields as field (field.id)}
                <div
                  style={`
                    position: relative;
                    padding: ${tokens.spacing[3]};
                    border-radius: 0.375rem;
                    border: 1px solid ${tokens.colors.neutral[200]};
                    background-color: ${tokens.colors.neutral[0]};
                    transition: all 0.15s;
                    group;
                    margin-bottom: ${tokens.spacing[2]};
                    cursor: pointer;
                    ${state.selectedFieldId === field.id ? `ring: 2px solid ${tokens.colors.primary[500]}; border-color: ${tokens.colors.primary[500]};` : ''}
                    ${formErrors.has(field.id) ? `border-color: ${tokens.colors.error[500]}; background-color: ${tokens.colors.error[50]};` : ''}
                  `}
                  data-field-id={field.id}
                  draggable={!state.isPreviewMode && !readonly}
                  role="button"
                  tabindex="0"
                  on:mouseover={(e) => {
                    if (!formErrors.has(field.id)) {
                      e.currentTarget.style.borderColor = tokens.colors.neutral[300];
                    }
                  }}
                  on:mouseout={(e) => {
                    if (!formErrors.has(field.id)) {
                      e.currentTarget.style.borderColor = tokens.colors.neutral[200];
                    }
                  }}
                  on:click|stopPropagation={() => selectField(field.id)}
                  on:keydown={(e) => {
                    if (e.key === 'Enter') {
                      selectField(field.id);
                    }
                  }}
                >
                  {#if !state.isPreviewMode && !readonly}
                    <div style={`
                      position: absolute;
                      left: 0;
                      top: 0;
                      bottom: 0;
                      width: 1.5rem;
                      display: flex;
                      align-items: center;
                      justify-content: center;
                      cursor: grab;
                      opacity: 0;
                      transition: opacity 0.15s;
                    `} class="group-hover:opacity-100">
                      <span style={`color: ${tokens.colors.neutral[400]}`}><Icon name="grip-vertical" size="sm" /></span>
                    </div>
                  {/if}

                  <div style={`margin-left: ${tokens.spacing[4]}`}>
                    <div style={`
                      font-size: 0.875rem;
                      font-weight: 500;
                      color: ${tokens.colors.neutral[700]};
                    `}>
                      {getFieldLabel(field)}
                      {#if field.config.required}
                        <span style={`color: ${tokens.colors.error[500]}; margin-left: 0.125rem`}>*</span>
                      {/if}
                    </div>
                    <div style={`
                      font-size: 0.75rem;
                      color: ${tokens.colors.neutral[500]};
                      margin-top: 0.375rem;
                      ${formErrors.has(field.id) ? `color: ${tokens.colors.error[500]};` : ''}
                    `}>
                      {getFieldTypeLabel(field.config.type)}
                      {#if formErrors.has(field.id)}
                        <span style={`margin-left: 0.5rem`}>{formErrors.get(field.id)}</span>
                      {/if}
                    </div>
                  </div>

                  {#if !state.isPreviewMode && !readonly}
                    <div style={`
                      position: absolute;
                      right: ${tokens.spacing[2]};
                      top: ${tokens.spacing[2]};
                      opacity: 0;
                      transition: opacity 0.15s;
                      display: flex;
                      gap: 0.25rem;
                    `} class="group-hover:opacity-100">
                      <button
                        type="button"
                        style={`
                          padding: ${tokens.spacing[1]};
                          border-radius: 0.375rem;
                          background: none;
                          border: none;
                          cursor: pointer;
                          color: ${tokens.colors.neutral[500]};
                          transition: all 0.15s;
                        `}
                        on:mouseover={(e) => {
                          e.currentTarget.style.backgroundColor = tokens.colors.neutral[100];
                          e.currentTarget.style.color = tokens.colors.neutral[700];
                        }}
                        on:mouseout={(e) => {
                          e.currentTarget.style.backgroundColor = 'transparent';
                          e.currentTarget.style.color = tokens.colors.neutral[500];
                        }}
                        on:click|stopPropagation={() => duplicateField(field.id)}
                        title="Duplicate"
                      >
                        <Icon name="copy" size="sm" />
                      </button>
                      <button
                        type="button"
                        style={`
                          padding: ${tokens.spacing[1]};
                          border-radius: 0.375rem;
                          background: none;
                          border: none;
                          cursor: pointer;
                          color: ${tokens.colors.neutral[500]};
                          transition: all 0.15s;
                        `}
                        on:mouseover={(e) => {
                          e.currentTarget.style.backgroundColor = tokens.colors.neutral[100];
                          e.currentTarget.style.color = tokens.colors.neutral[700];
                        }}
                        on:mouseout={(e) => {
                          e.currentTarget.style.backgroundColor = 'transparent';
                          e.currentTarget.style.color = tokens.colors.neutral[500];
                        }}
                        on:click|stopPropagation={() => removeField(field.id)}
                        title="Remove"
                      >
                        <Icon name="trash" size="sm" />
                      </button>
                    </div>
                  {/if}
                </div>
              {/each}

              {#if section.fields.length === 0}
                <div style={`
                  text-align: center;
                  color: ${tokens.colors.neutral[400]};
                  padding: 2rem 0;
                `}>
                  Drop fields here
                </div>
              {/if}
            </div>
          </div>
        {/each}
      {/if}
    </div>
  </div>

  <!-- Properties Panel -->
  {#if showProperties && !state.isPreviewMode}
    <div style={propertiesStyle}>
      <div style={`
        padding: ${tokens.spacing[3]} ${tokens.spacing[4]};
        border-bottom: 1px solid ${tokens.colors.neutral[200]};
        display: flex;
        align-items: center;
        justify-content: space-between;
      `}>
        <span style={`
          font-weight: 500;
          color: ${tokens.colors.neutral[900]};
        `}>Properties</span>
      </div>

      {#if selectedField}
        <div style={`padding: ${tokens.spacing[4]}`}>
          {#each Array.from(propertyGroups.entries()) as [groupKey, groupProps]}
            <div style={`margin-bottom: ${tokens.spacing[4]}`}>
              <h4 style={`
                font-size: 0.75rem;
                font-weight: 600;
                color: ${tokens.colors.neutral[500]};
                text-transform: uppercase;
                letter-spacing: 0.05em;
                margin-bottom: ${tokens.spacing[2]};
              `}>
                {propertyGroupLabels[groupKey] || groupKey}
              </h4>

              {#each groupProps as prop}
                {#if !prop.condition || prop.condition(selectedField.config)}
                  <div style={`margin-bottom: ${tokens.spacing[3]}`}>
                    {#if prop.type === 'text'}
                      <Input
                        label={prop.label}
                        size="sm"
                        placeholder={prop.placeholder}
                        helperText={prop.helperText}
                        value={(selectedField.config as unknown as Record<string, unknown>)[prop.key] as string || ''}
                        required={prop.required}
                        disabled={readonly}
                        on:input={(e) => updateFieldProperty(selectedField.id, prop.key, e.detail.value)}
                      />
                    {:else if prop.type === 'number'}
                      <Input
                        type="number"
                        label={prop.label}
                        size="sm"
                        placeholder={prop.placeholder}
                        helperText={prop.helperText}
                        value={String((selectedField.config as unknown as Record<string, unknown>)[prop.key] || '')}
                        required={prop.required}
                        disabled={readonly}
                        on:input={(e) => updateFieldProperty(selectedField.id, prop.key, e.detail.value ? Number(e.detail.value) : undefined)}
                      />
                    {:else if prop.type === 'boolean'}
                      <Checkbox
                        label={prop.label}
                        size="sm"
                        checked={(selectedField.config as unknown as Record<string, unknown>)[prop.key] as boolean || false}
                        disabled={readonly}
                        on:change={(e) => updateFieldProperty(selectedField.id, prop.key, e.detail.checked)}
                      />
                    {:else if prop.type === 'select' && prop.options}
                      <Select
                        label={prop.label}
                        size="sm"
                        options={prop.options.map((o) => ({ value: String(o.value), label: o.label }))}
                        value={String((selectedField.config as unknown as Record<string, unknown>)[prop.key] || '')}
                        disabled={readonly}
                        on:change={(e) => {
                          const opt = prop.options?.find((o) => String(o.value) === e.detail.value);
                          if (opt) {
                            updateFieldProperty(selectedField.id, prop.key, opt.value);
                          }
                        }}
                      />
                    {:else if prop.type === 'textarea'}
                      <TextArea
                        label={prop.label}
                        size="sm"
                        placeholder={prop.placeholder}
                        helperText={prop.helperText}
                        value={(selectedField.config as unknown as Record<string, unknown>)[prop.key] as string || ''}
                        required={prop.required}
                        disabled={readonly}
                        rows={3}
                        on:input={(e) => updateFieldProperty(selectedField.id, prop.key, e.detail.value)}
                      />
                    {:else if prop.type === 'array'}
                      <div style={`margin-bottom: ${tokens.spacing[2]}`}>
                        <label style={`
                          display: block;
                          font-size: 0.875rem;
                          font-weight: 500;
                          color: ${tokens.colors.neutral[700]};
                        `}>{prop.label}</label>
                        {#if prop.helperText}
                          <p style={`
                            font-size: 0.75rem;
                            color: ${tokens.colors.neutral[500]};
                            margin-top: ${tokens.spacing[1]};
                          `}>{prop.helperText}</p>
                        {/if}
                        <p style={`
                          font-size: 0.75rem;
                          color: ${tokens.colors.neutral[400]};
                          font-style: italic;
                          margin-top: ${tokens.spacing[1]};
                        `}>Edit options in code view</p>
                      </div>
                    {/if}
                  </div>
                {/if}
              {/each}
            </div>
          {/each}
        </div>
      {:else if selectedSection}
        <div style={`padding: ${tokens.spacing[4]}`}>
          <div style={`margin-bottom: ${tokens.spacing[4]}`}>
            <h4 style={`
              font-size: 0.75rem;
              font-weight: 600;
              color: ${tokens.colors.neutral[500]};
              text-transform: uppercase;
              letter-spacing: 0.05em;
              margin-bottom: ${tokens.spacing[2]};
            `}>Section Settings</h4>

            <div style={`margin-bottom: ${tokens.spacing[3]}`}>
              <Input
                label="Title"
                size="sm"
                value={selectedSection.title || ''}
                disabled={readonly}
                on:input={(e) => updateSectionProperty(selectedSection.id, 'title', e.detail.value)}
              />
            </div>

            <div style={`margin-bottom: ${tokens.spacing[3]}`}>
              <TextArea
                label="Description"
                size="sm"
                value={selectedSection.description || ''}
                disabled={readonly}
                rows={2}
                on:input={(e) => updateSectionProperty(selectedSection.id, 'description', e.detail.value)}
              />
            </div>

            <div style={`margin-bottom: ${tokens.spacing[3]}`}>
              <Input
                type="number"
                label="Columns"
                size="sm"
                value={String(selectedSection.columns || 1)}
                disabled={readonly}
                on:input={(e) => updateSectionProperty(selectedSection.id, 'columns', e.detail.value ? Number(e.detail.value) : 1)}
              />
            </div>

            <Checkbox
              label="Collapsible"
              size="sm"
              checked={selectedSection.collapsible || false}
              disabled={readonly}
              on:change={(e) => updateSectionProperty(selectedSection.id, 'collapsible', e.detail.checked)}
            />
          </div>
        </div>
      {:else}
        <div style={`
          padding: ${tokens.spacing[4]};
          text-align: center;
          color: ${tokens.colors.neutral[500]};
          font-size: 0.875rem;
        `}>
          Select a field or section to edit its properties
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  :global(.group:hover) :global(.group-hover\:opacity-100) {
    opacity: 1 !important;
  }
</style>
