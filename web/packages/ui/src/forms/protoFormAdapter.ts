/**
 * Proto-to-FormSchema Adapter
 *
 * Converts backend proto FormDefinition into the UI's FormSchema type
 * consumed by DynamicFormRenderer.svelte and CrudFormPage.svelte.
 *
 * This is the critical bridge between the API layer and the UI layer.
 */

import type {
  FormSchema,
  FormFieldConfig,
  FormSection,
  // FormSchemaLayout is the form-pipeline FormLayout type (aliased in
  // @chetana/core's barrel export so it doesn't collide with the
  // layout-system FormLayout used by DashboardLayout, MasterDetailLayout,
  // etc.). DynamicFormRenderer.svelte and CrudFormPage.svelte both
  // consume schema.layout as { type, columns, gap, sections } — that's
  // FormSchemaLayout, not the layout-system FormLayout.
  FormSchemaLayout,
  TextField,
  NumberField,
  SelectField,
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
  SelectOption,
} from '@chetana/core';

// ============================================================================
// PROTO TYPES (inline to avoid cross-package dependency from UI -> API)
// These mirror the types in @chetana/api/types/formservice.types
// ============================================================================

/** Proto FieldType enum values */
const ProtoFieldType = {
  TEXT: 0,
  NUMBER: 1,
  EMAIL: 2,
  DROPDOWN: 3,
  RADIO: 4,
  CHECKBOX: 5,
  DATE: 6,
  DATETIME: 7,
  FILE: 8,
  TEXTAREA: 9,
  MULTI_SELECT: 10,
  CURRENCY: 11,
  PHONE: 12,
  URL: 13,
  JSON: 14,
  ARRAY: 15,
  NESTED_FORM: 16,
  PASSWORD: 17,
  RICHTEXT: 18,
  PERCENTAGE: 19,
  TIME: 20,
  DATERANGE: 21,
  MONTHPICKER: 22,
  CHECKBOXGROUP: 23,
  SWITCH: 24,
  LOOKUP: 25,
  MULTILOOKUP: 26,
  TREE: 27,
  CASCADE: 28,
  TABLE: 29,
  OBJECT: 30,
  KEYVALUE: 31,
  IMAGE: 32,
  FORMULA: 33,
  BARCODE: 34,
  COLOR: 35,
  RATING: 36,
  SLIDER: 37,
  CRON: 38,
} as const;

type ProtoFieldTypeValue = (typeof ProtoFieldType)[keyof typeof ProtoFieldType];

/** Minimal proto field shape for the adapter */
export interface ProtoField {
  id: string;
  type: number;
  label: string;
  hint?: string;
  required?: boolean;
  validation?: ProtoValidation;
  optionsSource?: string;
  defaultValue?: unknown;
  attributes?: Record<string, string>;
  tableConfig?: ProtoTableConfig;
  lookupConfig?: ProtoLookupConfig;
  cascadeConfig?: ProtoCascadeConfig;
  maxFileSize?: number;
  allowedMimeTypes?: string[];
  multipleFiles?: boolean;
  fieldId?: string;
  displayName?: string;
  fieldType?: number;
  orderIndex?: number;
  order?: number;
  readonly?: boolean;
  placeholder?: string;
  hidden?: boolean;
  visibleWhen?: string;
  readonlyWhen?: string;
  hiddenWhen?: string;
  formulaConfig?: ProtoFormulaConfig;
  fileConfig?: ProtoFileConfig;
  apiConfig?: ProtoApiConfig;
}

interface ProtoValidation {
  min?: number;
  max?: number;
  minLength?: number;
  maxLength?: number;
  pattern?: string;
  allowedValues?: string[];
  required?: boolean;
  readonly?: boolean;
  hidden?: boolean;
  minItems?: number;
  maxItems?: number;
}

interface ProtoTableConfig {
  columns?: ProtoField[];
  allowAddRows?: boolean;
  allowDeleteRows?: boolean;
  allowReorder?: boolean;
  minRows?: number;
  maxRows?: number;
  allowAdd?: boolean;
  allowDelete?: boolean;
  allowEdit?: boolean;
}

interface ProtoLookupConfig {
  entityType: string;
  searchEndpoint: string;
  displayTemplate?: string;
  allowMultiple?: boolean;
  searchable?: boolean;
  minSearchLength?: number;
  clearable?: boolean;
  creatable?: boolean;
  pageSize?: number;
}

interface ProtoCascadeConfig {
  parentField?: string;
  levels?: ProtoCascadeLevel[];
  clearable?: boolean;
}

interface ProtoCascadeLevel {
  fieldId: string;
  parentField?: string;
  endpoint?: string;
  name?: string;
  entityType?: string;
  searchEndpoint?: string;
  displayTemplate?: string;
}

interface ProtoFormulaConfig {
  expression: string;
  dependentFields?: string[];
  resultType?: string;
  autoCalculate?: boolean;
  decimalPlaces?: number;
}

interface ProtoFileConfig {
  maxFileSize?: number;
  allowedMimeTypes?: string[];
  allowMultiple?: boolean;
}

interface ProtoApiConfig {
  method?: string;
  dependentFields?: string[];
  responseTransform?: string;
}

/** Minimal proto step shape */
export interface ProtoStep {
  id: string;
  label: string;
  fields: ProtoField[];
  order: number;
  stepId?: string;
  description?: string;
  title?: string;
  fieldIds?: string[];
}

/** Minimal proto metadata shape */
export interface ProtoMetadata {
  formId: string;
  title: string;
  description?: string;
  version?: string;
  module?: string;
  friendlyEndpoint: string;
  rpcEndpoint: string;
  service?: string;
  entityType?: string;
  /**
   * Field names that should appear as default columns in the entity's
   * list view. The proto FormDefinition declares this on `metadata.coreFields`
   * (snake_case in JSON, camelCase in TS via protojson). CrudListPage
   * uses these to derive the default `columns` prop when the caller
   * doesn't override.
   */
  coreFields?: string[];
}

/** Minimal proto FormDefinition shape */
export interface ProtoFormDef {
  metadata: ProtoMetadata;
  steps: ProtoStep[];
  allFields?: ProtoField[];
  version?: string;
}

// ============================================================================
// FIELD TYPE MAPPING
// ============================================================================

/**
 * Maps proto FieldType enum to the UI's FormFieldConfig type string.
 * This covers all 39 proto field types and maps them to the closest
 * UI component registered in DynamicFormRenderer's componentMap.
 */
function mapFieldType(protoType: number): FormFieldConfig['type'] {
  // Routes every proto FieldType to the richest component DynamicFormRenderer
  // has registered. Strings that aren't in FormFieldConfig['type'] (e.g.
  // 'currency', 'lookup', 'table') are accepted at runtime by the renderer's
  // componentMap which is typed `Partial<Record<string, any>>` — the cast is
  // just to satisfy TS callers. See docs/form-field-types.md for the matrix.
  const mapping: Record<number, string> = {
    [ProtoFieldType.TEXT]: 'text',
    [ProtoFieldType.NUMBER]: 'number',
    [ProtoFieldType.EMAIL]: 'email',
    [ProtoFieldType.DROPDOWN]: 'select',
    [ProtoFieldType.RADIO]: 'radio',
    [ProtoFieldType.CHECKBOX]: 'checkbox',
    [ProtoFieldType.DATE]: 'date',
    [ProtoFieldType.DATETIME]: 'datetime',
    [ProtoFieldType.FILE]: 'file',
    [ProtoFieldType.TEXTAREA]: 'textarea',
    [ProtoFieldType.MULTI_SELECT]: 'select',
    [ProtoFieldType.CURRENCY]: 'currency',
    [ProtoFieldType.PHONE]: 'phone',
    [ProtoFieldType.URL]: 'url',
    [ProtoFieldType.JSON]: 'json',
    [ProtoFieldType.ARRAY]: 'array',
    [ProtoFieldType.NESTED_FORM]: 'object',
    [ProtoFieldType.PASSWORD]: 'password',
    [ProtoFieldType.RICHTEXT]: 'richtext',
    [ProtoFieldType.PERCENTAGE]: 'percent',
    [ProtoFieldType.TIME]: 'time',
    [ProtoFieldType.DATERANGE]: 'daterange',
    [ProtoFieldType.MONTHPICKER]: 'month',
    [ProtoFieldType.CHECKBOXGROUP]: 'checkbox-group',
    [ProtoFieldType.SWITCH]: 'switch',
    [ProtoFieldType.LOOKUP]: 'lookup',
    [ProtoFieldType.MULTILOOKUP]: 'multi-lookup',
    [ProtoFieldType.TREE]: 'tree',
    [ProtoFieldType.CASCADE]: 'cascade',
    [ProtoFieldType.TABLE]: 'table',
    [ProtoFieldType.OBJECT]: 'object',
    // KEYVALUE renders as a real key/value pair editor, NOT a free-form
    // textarea. Routing it to 'textarea' (the prior behavior) made the
    // user type/parse JSON-ish text by hand and silently lost malformed
    // entries. See KeyValueEditor.svelte.
    [ProtoFieldType.KEYVALUE]: 'keyvalue',
    [ProtoFieldType.IMAGE]: 'image',
    // FORMULA renders as a read-only computed value, NOT an editable
    // NumberInput. Routing it to 'number' (the prior behavior) let the
    // user type values that the server-side expression engine would
    // silently overwrite or ignore. See FormulaField.svelte.
    [ProtoFieldType.FORMULA]: 'formula',
    [ProtoFieldType.BARCODE]: 'barcode',
    [ProtoFieldType.COLOR]: 'color',
    [ProtoFieldType.RATING]: 'rating',
    [ProtoFieldType.SLIDER]: 'slider',
    [ProtoFieldType.CRON]: 'cron',
  };

  return (mapping[protoType] ?? 'text') as FormFieldConfig['type'];
}

// ============================================================================
// FIELD CONVERSION
// ============================================================================

/**
 * Resolves the effective field type from a proto field.
 * The proto has both `type` and `fieldType`; prefer `fieldType` when present
 * (used by generators), fall back to `type`.
 */
function resolveFieldType(field: ProtoField): number {
  if (field.fieldType !== undefined && field.fieldType !== null && field.fieldType !== 0) {
    return field.fieldType;
  }
  return field.type;
}

/**
 * Convert allowed_values from proto validation into SelectOption[].
 */
function toSelectOptions(values: string[]): SelectOption[] {
  return values.map((v) => ({
    label: v.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase()),
    value: v,
  }));
}

/**
 * Build a visibility condition function from a `visibleWhen` expression string.
 * Supports simple `field == value` / `field != value` expressions.
 * For complex expressions the function always returns true (visible).
 */
function buildCondition(expr: string): ((values: Record<string, unknown>) => boolean) | undefined {
  if (!expr) return undefined;

  const eqMatch = expr.match(/^(\w+)\s*==\s*['"]?(.+?)['"]?$/);
  if (eqMatch && eqMatch[1] !== undefined && eqMatch[2] !== undefined) {
    const fieldName = eqMatch[1];
    const expected = eqMatch[2];
    return (values) => String(values[fieldName] ?? '') === expected;
  }

  const neqMatch = expr.match(/^(\w+)\s*!=\s*['"]?(.+?)['"]?$/);
  if (neqMatch && neqMatch[1] !== undefined && neqMatch[2] !== undefined) {
    const fieldName = neqMatch[1];
    const expected = neqMatch[2];
    return (values) => String(values[fieldName] ?? '') !== expected;
  }

  // Can't parse complex expressions — always visible
  return undefined;
}

/**
 * Converts a single ProtoField into the UI's FormFieldConfig.
 */
function convertField(field: ProtoField): FormFieldConfig {
  const effectiveType = resolveFieldType(field);
  const uiType = mapFieldType(effectiveType);
  const name = field.fieldId || field.id;
  const label = field.displayName || field.label;
  const required = field.required ?? field.validation?.required ?? false;
  const isReadonly = field.readonly ?? field.validation?.readonly ?? false;
  const isHidden = field.hidden ?? field.validation?.hidden ?? false;
  const placeholder = field.placeholder ?? field.hint ?? '';
  const helperText = field.hint ?? '';

  // Build the condition function if visibleWhen is present
  const condition = field.visibleWhen ? buildCondition(field.visibleWhen) : undefined;

  // Base properties shared across all field types
  const base = {
    name,
    label,
    required,
    readonly: isReadonly,
    hidden: isHidden,
    placeholder,
    helperText,
    disabled: false,
    condition,
    defaultValue: field.defaultValue as any,
  };

  // Type-specific conversion
  switch (uiType) {
    case 'text':
    case 'email':
    case 'password':
    case 'tel':
    case 'url':
    case 'search': {
      const config: TextField = {
        ...base,
        type: uiType,
        minLength: field.validation?.minLength,
        maxLength: field.validation?.maxLength,
        pattern: field.validation?.pattern,
      };
      return config;
    }

    case 'number': {
      const isCurrency = effectiveType === ProtoFieldType.CURRENCY;
      const isPercent = effectiveType === ProtoFieldType.PERCENTAGE;
      const isFormula = effectiveType === ProtoFieldType.FORMULA;

      const config: NumberField = {
        ...base,
        type: 'number',
        min: field.validation?.min,
        max: field.validation?.max,
        precision: field.formulaConfig?.decimalPlaces,
        format: isCurrency ? 'currency' : isPercent ? 'percent' : isFormula ? 'decimal' : undefined,
      };
      if (isReadonly || isFormula) {
        config.disabled = true;
      }
      return config;
    }

    case 'select': {
      const isMulti = effectiveType === ProtoFieldType.MULTI_SELECT;
      const options = field.validation?.allowedValues
        ? toSelectOptions(field.validation.allowedValues)
        : [];

      const config: SelectField = {
        ...base,
        type: 'select',
        options,
        multiple: isMulti,
        searchable: effectiveType === ProtoFieldType.TREE || effectiveType === ProtoFieldType.CASCADE,
        clearable: true,
      };
      return config;
    }

    case 'radio': {
      const options = (field.validation?.allowedValues ?? []).map((v) => ({
        label: v.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase()),
        value: v,
      }));
      return {
        ...base,
        type: 'radio',
        options,
      } as RadioField;
    }

    case 'checkbox': {
      return {
        ...base,
        type: 'checkbox',
      } as CheckboxField;
    }

    case 'checkbox-group': {
      const options = (field.validation?.allowedValues ?? []).map((v) => ({
        label: v.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase()),
        value: v,
      }));
      return {
        ...base,
        type: 'checkbox-group',
        options,
        minSelections: field.validation?.minItems,
        maxSelections: field.validation?.maxItems,
      } as CheckboxGroupField;
    }

    case 'switch': {
      return {
        ...base,
        type: 'switch',
      } as SwitchField;
    }

    case 'date':
    case 'datetime':
    case 'time':
    case 'month':
    case 'year': {
      return {
        ...base,
        type: uiType,
        clearable: true,
      } as DateField;
    }

    case 'daterange': {
      return {
        ...base,
        type: 'daterange',
        clearable: true,
      } as DateRangeField;
    }

    case 'textarea': {
      const isJson = effectiveType === ProtoFieldType.JSON;
      const isKeyValue = effectiveType === ProtoFieldType.KEYVALUE;
      return {
        ...base,
        type: 'textarea',
        rows: isJson || isKeyValue ? 8 : 4,
        minLength: field.validation?.minLength,
        maxLength: field.validation?.maxLength,
        showCount: true,
      } as TextareaField;
    }

    case 'richtext': {
      return {
        ...base,
        type: 'richtext',
      } as RichTextField;
    }

    case 'file': {
      const isImage = effectiveType === ProtoFieldType.IMAGE;
      const mimes = field.allowedMimeTypes ?? field.fileConfig?.allowedMimeTypes ?? [];
      const maxSize = field.maxFileSize ?? field.fileConfig?.maxFileSize;
      const multiple = field.multipleFiles ?? field.fileConfig?.allowMultiple ?? false;

      return {
        ...base,
        type: 'file',
        accept: mimes.length > 0 ? mimes.join(',') : isImage ? 'image/*' : undefined,
        multiple,
        maxSize: maxSize ?? undefined,
        showPreview: isImage,
        dragDrop: true,
      } as FileField;
    }

    case 'autocomplete': {
      const lc = field.lookupConfig;
      const isMulti = effectiveType === ProtoFieldType.MULTILOOKUP || lc?.allowMultiple;

      return {
        ...base,
        type: 'autocomplete',
        multiple: isMulti,
        minChars: lc?.minSearchLength ?? 2,
        debounceMs: 300,
        clearable: lc?.clearable ?? true,
        freeSolo: false,
        loadOptions: lc?.searchEndpoint
          ? createLookupLoader(lc.searchEndpoint, lc.displayTemplate)
          : undefined,
      } as AutocompleteField;
    }

    case 'color': {
      return {
        ...base,
        type: 'color',
        format: 'hex',
        showInput: true,
      } as ColorField;
    }

    case 'rating': {
      return {
        ...base,
        type: 'rating',
        max: field.validation?.max ?? 5,
        allowHalf: false,
        allowClear: true,
      } as RatingField;
    }

    case 'slider': {
      return {
        ...base,
        type: 'slider',
        min: field.validation?.min ?? 0,
        max: field.validation?.max ?? 100,
        step: 1,
        showTooltip: true,
      } as SliderField;
    }

    case 'array': {
      if (effectiveType === ProtoFieldType.TABLE && field.tableConfig?.columns) {
        const itemFields = field.tableConfig.columns.map(convertField);
        return {
          ...base,
          type: 'array',
          itemFields,
          minItems: field.tableConfig.minRows ?? 0,
          maxItems: field.tableConfig.maxRows ?? 100,
          addLabel: 'Add Row',
          removeLabel: 'Remove',
          sortable: field.tableConfig.allowReorder ?? false,
        } as ArrayField;
      }
      return {
        ...base,
        type: 'array',
        itemFields: [],
        minItems: field.validation?.minItems ?? 0,
        maxItems: field.validation?.maxItems ?? 100,
        addLabel: 'Add Item',
        removeLabel: 'Remove',
      } as ArrayField;
    }

    case 'object': {
      return {
        ...base,
        type: 'object',
        fields: [],
        collapsible: true,
      } as ObjectField;
    }

    case 'hidden': {
      return {
        ...base,
        type: 'hidden',
      } as HiddenField;
    }

    default: {
      return {
        ...base,
        type: 'text',
      } as TextField;
    }
  }
}

/**
 * Creates a loadOptions function for lookup/autocomplete fields.
 * Fetches options from the search endpoint with a query parameter.
 */
function createLookupLoader(
  searchEndpoint: string,
  displayTemplate?: string
): (query: string) => Promise<Array<{ label: string; value: unknown }>> {
  return async (query: string) => {
    try {
      const url = new URL(searchEndpoint, window.location.origin);
      url.searchParams.set('q', query);
      // 'omit' for Bearer-token auth — see DEPLOYMENT_READINESS.md item
      // 45 round 3 for the CORS wildcard-origin trap rationale.
      const response = await fetch(url.toString(), {
        credentials: 'omit',
        headers: { 'Content-Type': 'application/json' },
      });

      if (!response.ok) return [];
      const data = await response.json();
      const items = Array.isArray(data) ? data : data.items ?? data.results ?? [];

      return items.map((item: Record<string, unknown>) => ({
        label: displayTemplate
          ? renderTemplate(displayTemplate, item)
          : String(item.name ?? item.label ?? item.title ?? item.id ?? ''),
        value: item.id ?? item.value ?? item,
      }));
    } catch {
      return [];
    }
  };
}

/**
 * Renders a simple mustache-style template: "{{name}} ({{id}})"
 */
function renderTemplate(template: string, data: Record<string, unknown>): string {
  return template.replace(/\{\{(\w+)\}\}/g, (_, key) => String(data[key] ?? ''));
}

// ============================================================================
// MAIN ADAPTER
// ============================================================================

/**
 * Converts a proto FormDefinition into the UI's FormSchema.
 *
 * Steps are converted to FormSections, and fields within each step
 * become form fields in the schema. The flat `allFields` list is used
 * when steps have `fieldIds` references instead of embedded fields.
 */
export function adaptFormDefinition(def: ProtoFormDef): FormSchema<Record<string, unknown>> {
  // Build a map of all fields by ID for quick lookup
  const fieldMap = new Map<string, ProtoField>();
  if (def.allFields) {
    for (const f of def.allFields) {
      fieldMap.set(f.fieldId || f.id, f);
    }
  }
  for (const step of def.steps) {
    for (const f of step.fields ?? []) {
      fieldMap.set(f.fieldId || f.id, f);
    }
  }

  // Collect all converted fields (deduped by name)
  const allFields: FormFieldConfig[] = [];
  const seenNames = new Set<string>();

  // Build sections from steps
  const sections: FormSection[] = [];

  // Sort steps by order
  const sortedSteps = [...def.steps].sort((a, b) => (a.order ?? 0) - (b.order ?? 0));

  for (const step of sortedSteps) {
    const stepFieldNames: string[] = [];

    // Resolve fields: use embedded fields if present, otherwise use fieldIds + lookup from allFields
    let stepFields: ProtoField[] = step.fields ?? [];
    if (stepFields.length === 0 && step.fieldIds && step.fieldIds.length > 0) {
      stepFields = step.fieldIds
        .map((id) => fieldMap.get(id))
        .filter((f): f is ProtoField => f !== undefined);
    }

    // Sort fields by order
    const sortedFields = [...stepFields].sort(
      (a, b) => (a.orderIndex ?? a.order ?? 0) - (b.orderIndex ?? b.order ?? 0)
    );

    for (const protoField of sortedFields) {
      const converted = convertField(protoField);
      if (!seenNames.has(converted.name)) {
        allFields.push(converted);
        seenNames.add(converted.name);
      }
      stepFieldNames.push(converted.name);
    }

    sections.push({
      id: step.stepId || step.id,
      title: step.title || step.label,
      description: step.description,
      fields: stepFieldNames,
    });
  }

  // If there are fields in allFields not assigned to any step, add them to a default section
  if (def.allFields) {
    const unassigned: string[] = [];
    for (const f of def.allFields) {
      const name = f.fieldId || f.id;
      if (!seenNames.has(name)) {
        const converted = convertField(f);
        allFields.push(converted);
        seenNames.add(name);
        unassigned.push(name);
      }
    }
    if (unassigned.length > 0) {
      sections.push({
        id: 'additional',
        title: 'Additional Fields',
        fields: unassigned,
      });
    }
  }

  // Determine layout
  const layout: FormSchemaLayout = {
    type: 'vertical',
    columns: 1,
    gap: 'md',
    sections: sections.length > 0 ? sections : undefined,
  };

  return {
    fields: allFields,
    layout,
  };
}

/**
 * Extracts form metadata for display purposes (title, subtitle, etc.)
 * plus list-view inputs (coreFields → CrudListPage column defaults,
 * service → caller can derive list-RPC route by convention).
 */
export function extractFormMeta(def: ProtoFormDef): {
  title: string;
  subtitle: string;
  formId: string;
  module: string;
  service: string;
  rpcEndpoint: string;
  friendlyEndpoint: string;
  coreFields: string[];
} {
  const meta = def.metadata;
  return {
    title: meta.title,
    subtitle: meta.description ?? '',
    formId: meta.formId,
    module: meta.module ?? '',
    service: meta.service ?? '',
    rpcEndpoint: meta.rpcEndpoint,
    friendlyEndpoint: meta.friendlyEndpoint,
    coreFields: meta.coreFields ?? [],
  };
}
