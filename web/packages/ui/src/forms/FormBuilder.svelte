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

  // State
  let state: FormBuilderState = {
    sections: [{ id: 'default', fields: [] }],
    layout: { type: 'vertical', columns: 1 },
    selectedFieldId: null,
    selectedSectionId: null,
    isDragging: false,
    isPreviewMode: false,
  };

  let paletteCollapsed: Set<string> = new Set();
  let errors: Map<string, string> = new Map();

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

  // Initialize from schema
  onMount(() => {
    if (schema) {
      state = schemaToBuilderState(schema);
    }
  });

  // Notify changes
  function notifyChange() {
    const outputSchema = builderStateToSchema(state);
    dispatch('change', { schema: outputSchema, state });
  }

  // Validate all fields
  function validateFields() {
    errors.clear();

    const duplicates = findDuplicateNames(state.sections);
    for (const section of state.sections) {
      for (const field of section.fields) {
        const error = validateFieldConfig(field.config);
        if (error) {
          errors.set(field.id, error);
        } else if (duplicates.includes(field.config.name)) {
          errors.set(field.id, `Duplicate field name: ${field.config.name}`);
        }
      }
    }

    errors = errors;

    if (errors.size > 0) {
      dispatch('error', { errors: Array.from(errors.values()) });
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
    state.selectedFieldId = fieldId;
    state.selectedSectionId = null;
    state = state;
    dispatch('select', { fieldId, sectionId: null });
  }

  // Select section
  function selectSection(sectionId: string | null) {
    state.selectedSectionId = sectionId;
    state.selectedFieldId = null;
    state = state;
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

    state.sections[sectionIndex]!.fields = [...state.sections[sectionIndex]!.fields, newField];
    state = state;

    selectField(newField.id);
    validateFields();
    notifyChange();
  }

  // Remove field
  function removeField(fieldId: string) {
    if (readonly) return;

    for (const section of state.sections) {
      const index = section.fields.findIndex((f) => f.id === fieldId);
      if (index !== -1) {
        section.fields = section.fields.filter((f) => f.id !== fieldId);
        break;
      }
    }

    if (state.selectedFieldId === fieldId) {
      state.selectedFieldId = null;
    }

    errors.delete(fieldId);
    errors = errors;
    state = state;
    validateFields();
    notifyChange();
  }

  // Duplicate field
  function duplicateField(fieldId: string) {
    if (!canAddFields || readonly) return;

    for (const section of state.sections) {
      const field = section.fields.find((f) => f.id === fieldId);
      if (field) {
        const newField: FormBuilderField = {
          id: generateFieldId(field.config.type),
          config: {
            ...field.config,
            name: generateFieldName(field.config.type, allFieldNames),
            label: field.config.label ? `${field.config.label} (Copy)` : undefined,
          },
        };
        const index = section.fields.findIndex((f) => f.id === fieldId);
        section.fields = [...section.fields.slice(0, index + 1), newField, ...section.fields.slice(index + 1)];
        selectField(newField.id);
        break;
      }
    }

    state = state;
    validateFields();
    notifyChange();
  }

  // Update field property
  function updateFieldProperty(fieldId: string, key: string, value: unknown) {
    if (readonly) return;

    for (const section of state.sections) {
      const field = section.fields.find((f) => f.id === fieldId);
      if (field) {
        (field.config as unknown as Record<string, unknown>)[key] = value;
        break;
      }
    }

    state = state;
    validateFields();
    notifyChange();
  }

  // Add section
  function addSection() {
    if (!allowSections || readonly) return;

    const newSection: FormBuilderSection = {
      id: uid('section'),
      title: `Section ${state.sections.length + 1}`,
      fields: [],
    };

    state.sections = [...state.sections, newSection];
    selectSection(newSection.id);
    notifyChange();
  }

  // Remove section
  function removeSection(sectionId: string) {
    if (readonly || state.sections.length <= 1) return;

    // Move fields to first section
    const section = state.sections.find((s) => s.id === sectionId);
    if (section && section.fields.length > 0) {
      const targetSection = state.sections.find((s) => s.id !== sectionId);
      if (targetSection) {
        targetSection.fields = [...targetSection.fields, ...section.fields];
      }
    }

    state.sections = state.sections.filter((s) => s.id !== sectionId);

    if (state.selectedSectionId === sectionId) {
      state.selectedSectionId = null;
    }

    state = state;
    notifyChange();
  }

  // Update section property
  function updateSectionProperty(sectionId: string, key: string, value: unknown) {
    if (readonly) return;

    const section = state.sections.find((s) => s.id === sectionId);
    if (section) {
      (section as unknown as Record<string, unknown>)[key] = value;
    }

    state = state;
    notifyChange();
  }

  // Handle field reorder within section
  function handleFieldSort(sectionId: string, result: { sourceIndex: number; destinationIndex: number }) {
    if (readonly) return;

    const section = state.sections.find((s) => s.id === sectionId);
    if (section) {
      section.fields = reorderItems(section.fields, result.sourceIndex, result.destinationIndex);
      state = state;
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
    state.isPreviewMode = !state.isPreviewMode;
    state = state;
  }

  // Export schema
  export function getSchema(): FormSchema<Record<string, unknown>> {
    return builderStateToSchema(state);
  }

  // Import schema
  export function setSchema(newSchema: FormSchema<Record<string, unknown>>) {
    state = schemaToBuilderState(newSchema);
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
</script>

<div class={cn(formBuilderClasses.container, className)}>
  <!-- Field Palette -->
  {#if showPalette && !state.isPreviewMode}
    <div class={formBuilderClasses.palette}>
      <div class={formBuilderClasses.paletteHeader}>
        <span>Fields</span>
        {#if maxFields}
          <span class="text-sm text-neutral-500">{totalFields}/{maxFields}</span>
        {/if}
      </div>

      {#each effectiveCategories as category}
        <div class={formBuilderClasses.paletteCategory}>
          <button
            type="button"
            class={formBuilderClasses.paletteCategoryHeader}
            on:click={() => toggleCategory(category.id)}
          >
            <span class="flex items-center gap-2">
              {#if category.icon}
                <Icon name={category.icon} size="sm" />
              {/if}
              {category.label}
            </span>
            <Icon name={paletteCollapsed.has(category.id) ? 'chevron-right' : 'chevron-down'} size="sm" />
          </button>

          {#if !paletteCollapsed.has(category.id)}
            <div class={formBuilderClasses.paletteCategoryContent}>
              {#each category.fields as item}
                <div
                  class={formBuilderClasses.paletteItem}
                  draggable="true"
                  role="button"
                  tabindex="0"
                  on:dragstart={(e) => {
                    e.dataTransfer?.setData('application/form-builder-field', JSON.stringify(item));
                    state.isDragging = true;
                  }}
                  on:dragend={() => {
                    state.isDragging = false;
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
                  <Icon name={item.icon} size="sm" class={formBuilderClasses.paletteItemIcon} />
                  <span class={formBuilderClasses.paletteItemLabel}>{item.label}</span>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}

  <!-- Main Canvas -->
  <div class="flex-1 flex flex-col overflow-hidden">
    <!-- Toolbar -->
    <div class={formBuilderClasses.toolbar}>
      {#if allowSections && !readonly && !state.isPreviewMode}
        <button
          type="button"
          class={formBuilderClasses.toolbarButton}
          on:click={addSection}
        >
          <Icon name="plus" size="sm" class="inline mr-1" />
          Add Section
        </button>
      {/if}

      <div class="flex-1"></div>

      {#if showPreview}
        <button
          type="button"
          class={cn(
            formBuilderClasses.toolbarButton,
            state.isPreviewMode && formBuilderClasses.toolbarButtonActive
          )}
          on:click={togglePreview}
        >
          <Icon name={state.isPreviewMode ? 'edit' : 'eye'} size="sm" class="inline mr-1" />
          {state.isPreviewMode ? 'Edit' : 'Preview'}
        </button>
      {/if}
    </div>

    <!-- Canvas Area -->
    <div class={formBuilderClasses.canvas}>
      {#if state.sections.every((s) => s.fields.length === 0)}
        <!-- Empty State -->
        <div class={cn(formBuilderClasses.canvasDropZone, formBuilderClasses.canvasEmpty)}>
          <Icon name="layout" class={formBuilderClasses.canvasEmptyIcon} />
          <p class="text-lg font-medium mb-2">Start building your form</p>
          <p class="text-sm">Drag fields from the palette or click to add</p>
        </div>
      {:else}
        <!-- Sections -->
        {#each state.sections as section, sectionIndex (section.id)}
          <div
            class={cn(
              formBuilderClasses.section,
              state.selectedSectionId === section.id && formBuilderClasses.sectionSelected
            )}
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
              e.currentTarget.classList.add(formBuilderClasses.canvasDropZoneActive.split(' ')[0]!);
            }}
            on:dragleave={(e) => {
              e.currentTarget.classList.remove(formBuilderClasses.canvasDropZoneActive.split(' ')[0]!);
            }}
            on:drop={(e) => {
              e.preventDefault();
              e.currentTarget.classList.remove(formBuilderClasses.canvasDropZoneActive.split(' ')[0]!);
              const data = e.dataTransfer?.getData('application/form-builder-field');
              if (data) {
                const item = JSON.parse(data) as FieldPaletteItem;
                handlePaletteDrop(item, section.id);
              }
            }}
          >
            {#if section.title || allowSections}
              <div class={formBuilderClasses.sectionHeader}>
                <div class={formBuilderClasses.sectionTitle}>
                  {#if !state.isPreviewMode && state.selectedSectionId === section.id}
                    <input
                      type="text"
                      value={section.title || ''}
                      class="bg-transparent border-b border-neutral-300 focus:border-brand-primary-500 focus:outline-none px-1"
                      placeholder="Section title"
                      on:input={(e) => updateSectionProperty(section.id, 'title', e.currentTarget.value)}
                    />
                  {:else}
                    {section.title || `Section ${sectionIndex + 1}`}
                  {/if}
                </div>

                {#if !state.isPreviewMode && !readonly}
                  <div class="flex items-center gap-1">
                    {#if state.sections.length > 1}
                      <button
                        type="button"
                        class={formBuilderClasses.fieldActionButton}
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
              class={cn(formBuilderClasses.sectionContent, section.fields.length === 0 && 'min-h-[100px]')}
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
                  class={cn(
                    formBuilderClasses.field,
                    state.selectedFieldId === field.id && formBuilderClasses.fieldSelected,
                    errors.has(field.id) && formBuilderClasses.fieldError
                  )}
                  data-field-id={field.id}
                  draggable={!state.isPreviewMode && !readonly}
                  role="button"
                  tabindex="0"
                  on:click|stopPropagation={() => selectField(field.id)}
                  on:keydown={(e) => {
                    if (e.key === 'Enter') {
                      selectField(field.id);
                    }
                  }}
                >
                  {#if !state.isPreviewMode && !readonly}
                    <div class={formBuilderClasses.fieldHandle}>
                      <Icon name="grip-vertical" size="sm" class="text-neutral-400" />
                    </div>
                  {/if}

                  <div class={formBuilderClasses.fieldContent}>
                    <div class={formBuilderClasses.fieldLabel}>
                      {getFieldLabel(field)}
                      {#if field.config.required}
                        <span class="text-semantic-error-500 ml-0.5">*</span>
                      {/if}
                    </div>
                    <div class={formBuilderClasses.fieldType}>
                      {getFieldTypeLabel(field.config.type)}
                      {#if errors.has(field.id)}
                        <span class="text-semantic-error-500 ml-2">{errors.get(field.id)}</span>
                      {/if}
                    </div>
                  </div>

                  {#if !state.isPreviewMode && !readonly}
                    <div class={formBuilderClasses.fieldActions}>
                      <button
                        type="button"
                        class={formBuilderClasses.fieldActionButton}
                        on:click|stopPropagation={() => duplicateField(field.id)}
                        title="Duplicate"
                      >
                        <Icon name="copy" size="sm" />
                      </button>
                      <button
                        type="button"
                        class={formBuilderClasses.fieldActionButton}
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
                <div class="text-center text-neutral-400 py-8">
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
    <div class={formBuilderClasses.properties}>
      <div class={formBuilderClasses.propertiesHeader}>
        <span class={formBuilderClasses.propertiesTitle}>Properties</span>
      </div>

      {#if selectedField}
        <div class={formBuilderClasses.propertiesContent}>
          {#each Array.from(propertyGroups.entries()) as [groupKey, groupProps]}
            <div class={formBuilderClasses.propertiesGroup}>
              <h4 class={formBuilderClasses.propertiesGroupTitle}>
                {propertyGroupLabels[groupKey] || groupKey}
              </h4>

              {#each groupProps as prop}
                {#if !prop.condition || prop.condition(selectedField.config)}
                  <div class="space-y-1">
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
                      <!-- Options editor for select/radio/checkbox-group -->
                      <div class="space-y-2">
                        <label class="block text-sm font-medium text-neutral-700">{prop.label}</label>
                        {#if prop.helperText}
                          <p class="text-xs text-neutral-500">{prop.helperText}</p>
                        {/if}
                        <!-- Simple options editor - would need a more complex component for full functionality -->
                        <p class="text-xs text-neutral-400 italic">Edit options in code view</p>
                      </div>
                    {/if}
                  </div>
                {/if}
              {/each}
            </div>
          {/each}
        </div>
      {:else if selectedSection}
        <div class={formBuilderClasses.propertiesContent}>
          <div class={formBuilderClasses.propertiesGroup}>
            <h4 class={formBuilderClasses.propertiesGroupTitle}>Section Settings</h4>

            <Input
              label="Title"
              size="sm"
              value={selectedSection.title || ''}
              disabled={readonly}
              on:input={(e) => updateSectionProperty(selectedSection.id, 'title', e.detail.value)}
            />

            <TextArea
              label="Description"
              size="sm"
              value={selectedSection.description || ''}
              disabled={readonly}
              rows={2}
              on:input={(e) => updateSectionProperty(selectedSection.id, 'description', e.detail.value)}
            />

            <Input
              type="number"
              label="Columns"
              size="sm"
              value={String(selectedSection.columns || 1)}
              disabled={readonly}
              on:input={(e) => updateSectionProperty(selectedSection.id, 'columns', e.detail.value ? Number(e.detail.value) : 1)}
            />

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
        <div class={formBuilderClasses.propertiesEmpty}>
          Select a field or section to edit its properties
        </div>
      {/if}
    </div>
  {/if}
</div>
