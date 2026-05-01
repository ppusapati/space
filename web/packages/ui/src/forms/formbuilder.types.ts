/**
 * Form Builder Types
 * Types for the visual drag-and-drop form builder
 */

import type { FormFieldConfig, FormSchemaLayout as FormLayout, FormSection, FormSchema } from '@samavāya/core';
import type { Size } from '../types';

// ============================================================================
// Form Builder Types
// ============================================================================

/** Field palette category */
export interface FieldPaletteCategory {
  id: string;
  label: string;
  icon?: string;
  fields: FieldPaletteItem[];
  collapsed?: boolean;
}

/** Field palette item */
export interface FieldPaletteItem {
  type: FormFieldConfig['type'];
  label: string;
  icon: string;
  description?: string;
  defaultConfig: Partial<FormFieldConfig>;
}

/** Form builder field (field with builder metadata) */
export interface FormBuilderField {
  id: string;
  config: FormFieldConfig;
  selected?: boolean;
  error?: string;
}

/** Form builder section */
export interface FormBuilderSection {
  id: string;
  title?: string;
  description?: string;
  icon?: string;
  fields: FormBuilderField[];
  columns?: number;
  collapsible?: boolean;
  collapsed?: boolean;
}

/** Form builder state */
export interface FormBuilderState {
  sections: FormBuilderSection[];
  layout: FormLayout;
  selectedFieldId: string | null;
  selectedSectionId: string | null;
  isDragging: boolean;
  isPreviewMode: boolean;
}

/** Form builder props */
export interface FormBuilderProps {
  /** Initial schema to load */
  schema?: FormSchema<Record<string, unknown>>;
  /** Available field types (defaults to all) */
  availableFields?: FormFieldConfig['type'][];
  /** Allow sections */
  allowSections?: boolean;
  /** Max fields limit */
  maxFields?: number;
  /** Size variant */
  size?: Size;
  /** Read-only mode */
  readonly?: boolean;
  /** Show field palette */
  showPalette?: boolean;
  /** Show properties panel */
  showProperties?: boolean;
  /** Show preview toggle */
  showPreview?: boolean;
  /** Custom field palette categories */
  fieldCategories?: FieldPaletteCategory[];
  /** Custom field validators */
  fieldValidators?: Record<string, (config: FormFieldConfig) => string | null>;
}

/** Field property editor config */
export interface FieldPropertyConfig {
  key: string;
  label: string;
  type: 'text' | 'number' | 'boolean' | 'select' | 'textarea' | 'array' | 'object' | 'code';
  options?: { label: string; value: unknown }[];
  placeholder?: string;
  helperText?: string;
  required?: boolean;
  condition?: (field: FormFieldConfig) => boolean;
  group?: string;
}

// ============================================================================
// Default Field Palette
// ============================================================================

export const defaultFieldCategories: FieldPaletteCategory[] = [
  {
    id: 'basic',
    label: 'Basic Fields',
    icon: 'edit',
    fields: [
      {
        type: 'text',
        label: 'Text Input',
        icon: 'type',
        description: 'Single-line text input',
        defaultConfig: { type: 'text', name: '', label: 'Text Field' },
      },
      {
        type: 'textarea',
        label: 'Textarea',
        icon: 'align-left',
        description: 'Multi-line text area',
        defaultConfig: { type: 'textarea', name: '', label: 'Textarea', rows: 3 },
      },
      {
        type: 'number',
        label: 'Number',
        icon: 'hash',
        description: 'Numeric input',
        defaultConfig: { type: 'number', name: '', label: 'Number' },
      },
      {
        type: 'email',
        label: 'Email',
        icon: 'mail',
        description: 'Email address input',
        defaultConfig: { type: 'email', name: '', label: 'Email' } as Partial<FormFieldConfig>,
      },
      {
        type: 'password',
        label: 'Password',
        icon: 'lock',
        description: 'Password input',
        defaultConfig: { type: 'password', name: '', label: 'Password' } as Partial<FormFieldConfig>,
      },
    ],
  },
  {
    id: 'selection',
    label: 'Selection Fields',
    icon: 'list',
    fields: [
      {
        type: 'select',
        label: 'Select',
        icon: 'chevron-down',
        description: 'Dropdown select',
        defaultConfig: { type: 'select', name: '', label: 'Select', options: [] },
      },
      {
        type: 'checkbox',
        label: 'Checkbox',
        icon: 'check-square',
        description: 'Single checkbox',
        defaultConfig: { type: 'checkbox', name: '', label: 'Checkbox' },
      },
      {
        type: 'checkbox-group',
        label: 'Checkbox Group',
        icon: 'check-square',
        description: 'Multiple checkboxes',
        defaultConfig: { type: 'checkbox-group', name: '', label: 'Checkbox Group', options: [] },
      },
      {
        type: 'radio',
        label: 'Radio',
        icon: 'circle',
        description: 'Radio button group',
        defaultConfig: { type: 'radio', name: '', label: 'Radio', options: [] },
      },
      {
        type: 'switch',
        label: 'Switch',
        icon: 'toggle-left',
        description: 'Toggle switch',
        defaultConfig: { type: 'switch', name: '', label: 'Switch' },
      },
    ],
  },
  {
    id: 'date',
    label: 'Date & Time',
    icon: 'calendar',
    fields: [
      {
        type: 'date',
        label: 'Date',
        icon: 'calendar',
        description: 'Date picker',
        defaultConfig: { type: 'date', name: '', label: 'Date' },
      },
      {
        type: 'datetime',
        label: 'Date & Time',
        icon: 'clock',
        description: 'Date and time picker',
        defaultConfig: { type: 'datetime', name: '', label: 'Date & Time' } as Partial<FormFieldConfig>,
      },
      {
        type: 'time',
        label: 'Time',
        icon: 'clock',
        description: 'Time picker',
        defaultConfig: { type: 'time', name: '', label: 'Time' } as Partial<FormFieldConfig>,
      },
      {
        type: 'daterange',
        label: 'Date Range',
        icon: 'calendar',
        description: 'Date range picker',
        defaultConfig: { type: 'daterange', name: '', label: 'Date Range' },
      },
    ],
  },
  {
    id: 'advanced',
    label: 'Advanced Fields',
    icon: 'sliders',
    fields: [
      {
        type: 'file',
        label: 'File Upload',
        icon: 'upload',
        description: 'File upload field',
        defaultConfig: { type: 'file', name: '', label: 'File Upload' },
      },
      {
        type: 'richtext',
        label: 'Rich Text',
        icon: 'file-text',
        description: 'Rich text editor',
        defaultConfig: { type: 'richtext', name: '', label: 'Rich Text' },
      },
      {
        type: 'color',
        label: 'Color',
        icon: 'palette',
        description: 'Color picker',
        defaultConfig: { type: 'color', name: '', label: 'Color' },
      },
      {
        type: 'slider',
        label: 'Slider',
        icon: 'sliders',
        description: 'Range slider',
        defaultConfig: { type: 'slider', name: '', label: 'Slider', min: 0, max: 100 },
      },
      {
        type: 'rating',
        label: 'Rating',
        icon: 'star',
        description: 'Star rating',
        defaultConfig: { type: 'rating', name: '', label: 'Rating', max: 5 },
      },
    ],
  },
  {
    id: 'structure',
    label: 'Structure',
    icon: 'layout',
    fields: [
      {
        type: 'array',
        label: 'Repeater',
        icon: 'repeat',
        description: 'Repeatable field group',
        defaultConfig: { type: 'array', name: '', label: 'Repeater', itemFields: [] },
      },
      {
        type: 'object',
        label: 'Field Group',
        icon: 'folder',
        description: 'Grouped fields',
        defaultConfig: { type: 'object', name: '', label: 'Field Group', fields: [] },
      },
      {
        type: 'hidden',
        label: 'Hidden',
        icon: 'eye-off',
        description: 'Hidden field',
        defaultConfig: { type: 'hidden', name: '' },
      },
    ],
  },
];

// ============================================================================
// Field Property Definitions
// ============================================================================

export const commonFieldProperties: FieldPropertyConfig[] = [
  { key: 'name', label: 'Field Name', type: 'text', required: true, group: 'basic', helperText: 'Unique identifier for this field' },
  { key: 'label', label: 'Label', type: 'text', group: 'basic' },
  { key: 'placeholder', label: 'Placeholder', type: 'text', group: 'basic' },
  { key: 'helperText', label: 'Helper Text', type: 'text', group: 'basic' },
  { key: 'required', label: 'Required', type: 'boolean', group: 'validation' },
  { key: 'disabled', label: 'Disabled', type: 'boolean', group: 'state' },
  { key: 'readonly', label: 'Read-only', type: 'boolean', group: 'state' },
  { key: 'hidden', label: 'Hidden', type: 'boolean', group: 'state' },
];

export const textFieldProperties: FieldPropertyConfig[] = [
  { key: 'minLength', label: 'Min Length', type: 'number', group: 'validation' },
  { key: 'maxLength', label: 'Max Length', type: 'number', group: 'validation' },
  { key: 'pattern', label: 'Pattern (Regex)', type: 'text', group: 'validation' },
  { key: 'autocomplete', label: 'Autocomplete', type: 'text', group: 'advanced' },
  { key: 'prefix', label: 'Prefix', type: 'text', group: 'display' },
  { key: 'suffix', label: 'Suffix', type: 'text', group: 'display' },
  { key: 'clearable', label: 'Clearable', type: 'boolean', group: 'display' },
];

export const numberFieldProperties: FieldPropertyConfig[] = [
  { key: 'min', label: 'Minimum', type: 'number', group: 'validation' },
  { key: 'max', label: 'Maximum', type: 'number', group: 'validation' },
  { key: 'step', label: 'Step', type: 'number', group: 'validation' },
  { key: 'precision', label: 'Decimal Precision', type: 'number', group: 'display' },
  {
    key: 'format',
    label: 'Format',
    type: 'select',
    group: 'display',
    options: [
      { label: 'Integer', value: 'integer' },
      { label: 'Decimal', value: 'decimal' },
      { label: 'Currency', value: 'currency' },
      { label: 'Percent', value: 'percent' },
    ],
  },
  { key: 'showButtons', label: 'Show Buttons', type: 'boolean', group: 'display' },
];

export const selectFieldProperties: FieldPropertyConfig[] = [
  { key: 'options', label: 'Options', type: 'array', group: 'options', helperText: 'Add options for the select' },
  { key: 'multiple', label: 'Multiple Selection', type: 'boolean', group: 'behavior' },
  { key: 'searchable', label: 'Searchable', type: 'boolean', group: 'behavior' },
  { key: 'clearable', label: 'Clearable', type: 'boolean', group: 'behavior' },
  { key: 'creatable', label: 'Allow Create', type: 'boolean', group: 'behavior' },
  { key: 'maxSelections', label: 'Max Selections', type: 'number', group: 'validation', condition: (f) => (f as { multiple?: boolean }).multiple === true },
];

export const textareaFieldProperties: FieldPropertyConfig[] = [
  { key: 'rows', label: 'Rows', type: 'number', group: 'display' },
  { key: 'minRows', label: 'Min Rows', type: 'number', group: 'display' },
  { key: 'maxRows', label: 'Max Rows', type: 'number', group: 'display' },
  { key: 'autoResize', label: 'Auto Resize', type: 'boolean', group: 'behavior' },
  { key: 'minLength', label: 'Min Length', type: 'number', group: 'validation' },
  { key: 'maxLength', label: 'Max Length', type: 'number', group: 'validation' },
  { key: 'showCount', label: 'Show Count', type: 'boolean', group: 'display' },
  {
    key: 'resize',
    label: 'Resize',
    type: 'select',
    group: 'display',
    options: [
      { label: 'None', value: 'none' },
      { label: 'Vertical', value: 'vertical' },
      { label: 'Horizontal', value: 'horizontal' },
      { label: 'Both', value: 'both' },
    ],
  },
];

export const dateFieldProperties: FieldPropertyConfig[] = [
  { key: 'format', label: 'Date Format', type: 'text', group: 'display', placeholder: 'YYYY-MM-DD' },
  { key: 'clearable', label: 'Clearable', type: 'boolean', group: 'behavior' },
  {
    key: 'firstDayOfWeek',
    label: 'First Day',
    type: 'select',
    group: 'display',
    options: [
      { label: 'Sunday', value: 0 },
      { label: 'Monday', value: 1 },
      { label: 'Saturday', value: 6 },
    ],
  },
];

/** Get properties for a field type */
export function getFieldProperties(type: FormFieldConfig['type']): FieldPropertyConfig[] {
  const properties = [...commonFieldProperties];

  switch (type) {
    case 'text':
    case 'email':
    case 'password':
    case 'tel':
    case 'url':
    case 'search':
      properties.push(...textFieldProperties);
      break;
    case 'number':
      properties.push(...numberFieldProperties);
      break;
    case 'select':
      properties.push(...selectFieldProperties);
      break;
    case 'textarea':
      properties.push(...textareaFieldProperties);
      break;
    case 'date':
    case 'datetime':
    case 'time':
    case 'month':
    case 'year':
      properties.push(...dateFieldProperties);
      break;
    case 'checkbox':
      properties.push({ key: 'indeterminate', label: 'Indeterminate', type: 'boolean', group: 'state' });
      properties.push({
        key: 'labelPosition',
        label: 'Label Position',
        type: 'select',
        group: 'display',
        options: [
          { label: 'Left', value: 'left' },
          { label: 'Right', value: 'right' },
        ],
      });
      break;
    case 'checkbox-group':
    case 'radio':
      properties.push({ key: 'options', label: 'Options', type: 'array', group: 'options' });
      properties.push({
        key: 'orientation',
        label: 'Orientation',
        type: 'select',
        group: 'display',
        options: [
          { label: 'Horizontal', value: 'horizontal' },
          { label: 'Vertical', value: 'vertical' },
        ],
      });
      properties.push({ key: 'columns', label: 'Columns', type: 'number', group: 'display' });
      break;
    case 'switch':
      properties.push({ key: 'onLabel', label: 'On Label', type: 'text', group: 'display' });
      properties.push({ key: 'offLabel', label: 'Off Label', type: 'text', group: 'display' });
      break;
    case 'slider':
      properties.push({ key: 'min', label: 'Minimum', type: 'number', required: true, group: 'validation' });
      properties.push({ key: 'max', label: 'Maximum', type: 'number', required: true, group: 'validation' });
      properties.push({ key: 'step', label: 'Step', type: 'number', group: 'validation' });
      properties.push({ key: 'range', label: 'Range Mode', type: 'boolean', group: 'behavior' });
      properties.push({ key: 'showTooltip', label: 'Show Tooltip', type: 'boolean', group: 'display' });
      properties.push({ key: 'showInput', label: 'Show Input', type: 'boolean', group: 'display' });
      break;
    case 'rating':
      properties.push({ key: 'max', label: 'Max Stars', type: 'number', group: 'display' });
      properties.push({ key: 'allowHalf', label: 'Allow Half', type: 'boolean', group: 'behavior' });
      properties.push({ key: 'allowClear', label: 'Allow Clear', type: 'boolean', group: 'behavior' });
      break;
    case 'color':
      properties.push({
        key: 'format',
        label: 'Format',
        type: 'select',
        group: 'display',
        options: [
          { label: 'HEX', value: 'hex' },
          { label: 'RGB', value: 'rgb' },
          { label: 'HSL', value: 'hsl' },
        ],
      });
      properties.push({ key: 'showInput', label: 'Show Input', type: 'boolean', group: 'display' });
      properties.push({ key: 'showAlpha', label: 'Show Alpha', type: 'boolean', group: 'display' });
      break;
    case 'file':
      properties.push({ key: 'accept', label: 'Accept Types', type: 'text', group: 'validation', placeholder: '.jpg,.png,.pdf' });
      properties.push({ key: 'multiple', label: 'Multiple Files', type: 'boolean', group: 'behavior' });
      properties.push({ key: 'maxSize', label: 'Max Size (bytes)', type: 'number', group: 'validation' });
      properties.push({ key: 'maxFiles', label: 'Max Files', type: 'number', group: 'validation' });
      properties.push({ key: 'showPreview', label: 'Show Preview', type: 'boolean', group: 'display' });
      properties.push({ key: 'dragDrop', label: 'Drag & Drop', type: 'boolean', group: 'behavior' });
      break;
    case 'richtext':
      properties.push({ key: 'minHeight', label: 'Min Height', type: 'text', group: 'display', placeholder: '150px' });
      properties.push({ key: 'maxHeight', label: 'Max Height', type: 'text', group: 'display', placeholder: '400px' });
      break;
    case 'array':
      properties.push({ key: 'minItems', label: 'Min Items', type: 'number', group: 'validation' });
      properties.push({ key: 'maxItems', label: 'Max Items', type: 'number', group: 'validation' });
      properties.push({ key: 'addLabel', label: 'Add Button Label', type: 'text', group: 'display' });
      properties.push({ key: 'sortable', label: 'Sortable', type: 'boolean', group: 'behavior' });
      properties.push({ key: 'collapsible', label: 'Collapsible', type: 'boolean', group: 'display' });
      break;
    case 'object':
      properties.push({ key: 'columns', label: 'Columns', type: 'number', group: 'display' });
      properties.push({ key: 'collapsible', label: 'Collapsible', type: 'boolean', group: 'display' });
      properties.push({ key: 'defaultCollapsed', label: 'Collapsed by Default', type: 'boolean', group: 'display' });
      break;
  }

  return properties;
}

// ============================================================================
// CSS Classes
// ============================================================================

export const formBuilderClasses = {
  container: 'flex h-full bg-white border border-neutral-200 rounded-lg overflow-hidden',

  // Palette
  palette: 'w-64 border-r border-neutral-200 bg-neutral-50 overflow-y-auto flex-shrink-0',
  paletteHeader: 'px-4 py-3 border-b border-neutral-200 bg-white font-medium text-neutral-900',
  paletteCategory: 'border-b border-neutral-100',
  paletteCategoryHeader: 'flex items-center justify-between px-4 py-2 hover:bg-neutral-100 cursor-pointer text-sm font-medium text-neutral-700',
  paletteCategoryContent: 'p-2 space-y-1',
  paletteItem: 'flex items-center gap-2 px-3 py-2 rounded-md bg-white border border-neutral-200 hover:border-brand-primary-300 hover:bg-brand-primary-50 cursor-grab text-sm transition-colors',
  paletteItemIcon: 'text-neutral-500',
  paletteItemLabel: 'text-neutral-700',
  paletteItemDragging: 'opacity-50 shadow-lg ring-2 ring-brand-primary-500',

  // Canvas
  canvas: 'flex-1 overflow-y-auto p-6 bg-neutral-100',
  canvasEmpty: 'flex flex-col items-center justify-center h-full text-center text-neutral-500',
  canvasEmptyIcon: 'w-16 h-16 mb-4 text-neutral-300',
  canvasDropZone: 'min-h-[200px] bg-white rounded-lg border-2 border-dashed border-neutral-300 transition-colors',
  canvasDropZoneActive: 'border-brand-primary-500 bg-brand-primary-50',

  // Form preview
  formArea: 'bg-white rounded-lg border border-neutral-200 shadow-sm',
  formHeader: 'px-4 py-3 border-b border-neutral-200 flex items-center justify-between',
  formContent: 'p-4 space-y-4',

  // Section
  section: 'bg-white rounded-lg border border-neutral-200 mb-4',
  sectionHeader: 'px-4 py-3 border-b border-neutral-200 flex items-center justify-between',
  sectionTitle: 'font-medium text-neutral-900',
  sectionContent: 'p-4',
  sectionSelected: 'ring-2 ring-brand-primary-500',

  // Field
  field: 'relative p-3 rounded-md border border-neutral-200 bg-white hover:border-neutral-300 transition-colors group',
  fieldSelected: 'ring-2 ring-brand-primary-500 border-brand-primary-500',
  fieldDragging: 'opacity-50',
  fieldHandle: 'absolute left-0 top-0 bottom-0 w-6 flex items-center justify-center cursor-grab opacity-0 group-hover:opacity-100 transition-opacity',
  fieldContent: 'ml-4',
  fieldLabel: 'text-sm font-medium text-neutral-700',
  fieldType: 'text-xs text-neutral-500 mt-0.5',
  fieldActions: 'absolute right-2 top-2 opacity-0 group-hover:opacity-100 transition-opacity flex gap-1',
  fieldActionButton: 'p-1 rounded hover:bg-neutral-100 text-neutral-500 hover:text-neutral-700',
  fieldError: 'border-semantic-error-500 bg-semantic-error-50',

  // Properties panel
  properties: 'w-80 border-l border-neutral-200 bg-white overflow-y-auto flex-shrink-0',
  propertiesHeader: 'px-4 py-3 border-b border-neutral-200 flex items-center justify-between',
  propertiesTitle: 'font-medium text-neutral-900',
  propertiesEmpty: 'p-4 text-center text-neutral-500 text-sm',
  propertiesContent: 'p-4 space-y-4',
  propertiesGroup: 'space-y-3',
  propertiesGroupTitle: 'text-xs font-medium text-neutral-500 uppercase tracking-wider',

  // Toolbar
  toolbar: 'flex items-center gap-2 px-4 py-2 border-b border-neutral-200 bg-white',
  toolbarButton: 'px-3 py-1.5 text-sm rounded-md hover:bg-neutral-100 text-neutral-700 transition-colors',
  toolbarButtonActive: 'bg-brand-primary-100 text-brand-primary-700 hover:bg-brand-primary-200',
  toolbarDivider: 'w-px h-6 bg-neutral-200 mx-2',

  // Drop indicator
  dropIndicator: 'absolute left-0 right-0 h-0.5 bg-brand-primary-500 z-10',
};

// ============================================================================
// Utilities
// ============================================================================

/** Generate unique field ID */
export function generateFieldId(type: string): string {
  return `field_${type}_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`;
}

/** Generate unique field name */
export function generateFieldName(type: string, existingNames: string[]): string {
  let counter = 1;
  let name = type;
  while (existingNames.includes(name)) {
    name = `${type}${counter}`;
    counter++;
  }
  return name;
}

/** Convert builder state to form schema */
export function builderStateToSchema(state: FormBuilderState): FormSchema<Record<string, unknown>> {
  const fields: FormFieldConfig[] = [];
  const sections: FormSection[] = [];

  for (const section of state.sections) {
    const sectionFieldNames: string[] = [];

    for (const field of section.fields) {
      fields.push(field.config);
      sectionFieldNames.push(field.config.name);
    }

    if (section.title || section.description) {
      sections.push({
        id: section.id,
        title: section.title,
        description: section.description,
        icon: section.icon,
        fields: sectionFieldNames,
        collapsible: section.collapsible,
        columns: section.columns,
      });
    }
  }

  return {
    fields,
    layout: {
      ...state.layout,
      sections: sections.length > 0 ? sections : undefined,
    },
  };
}

/** Convert form schema to builder state */
export function schemaToBuilderState(schema: FormSchema<Record<string, unknown>>): FormBuilderState {
  const sections: FormBuilderSection[] = [];
  const fieldMap = new Map(schema.fields.map((f) => [f.name, f]));

  if (schema.layout?.sections && schema.layout.sections.length > 0) {
    for (const section of schema.layout.sections) {
      const sectionFields: FormBuilderField[] = [];

      for (const fieldName of section.fields) {
        const config = fieldMap.get(fieldName);
        if (config) {
          sectionFields.push({
            id: generateFieldId(config.type),
            config,
          });
          fieldMap.delete(fieldName);
        }
      }

      sections.push({
        id: section.id,
        title: section.title,
        description: section.description,
        icon: section.icon,
        fields: sectionFields,
        columns: section.columns,
        collapsible: section.collapsible,
      });
    }
  }

  // Add remaining fields to a default section
  const remainingFields: FormBuilderField[] = [];
  for (const config of fieldMap.values()) {
    remainingFields.push({
      id: generateFieldId(config.type),
      config,
    });
  }

  if (remainingFields.length > 0) {
    sections.push({
      id: 'default',
      fields: remainingFields,
    });
  }

  // If no sections exist, create a default one
  if (sections.length === 0) {
    sections.push({
      id: 'default',
      fields: [],
    });
  }

  return {
    sections,
    layout: schema.layout || { type: 'vertical', columns: 1 },
    selectedFieldId: null,
    selectedSectionId: null,
    isDragging: false,
    isPreviewMode: false,
  };
}

/** Validate field configuration */
export function validateFieldConfig(config: FormFieldConfig): string | null {
  if (!config.name || config.name.trim() === '') {
    return 'Field name is required';
  }

  if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(config.name)) {
    return 'Field name must start with a letter or underscore and contain only letters, numbers, and underscores';
  }

  return null;
}

/** Check for duplicate field names */
export function findDuplicateNames(sections: FormBuilderSection[]): string[] {
  const names = new Map<string, number>();

  for (const section of sections) {
    for (const field of section.fields) {
      const name = field.config.name;
      names.set(name, (names.get(name) || 0) + 1);
    }
  }

  return Array.from(names.entries())
    .filter(([_, count]) => count > 1)
    .map(([name]) => name);
}
