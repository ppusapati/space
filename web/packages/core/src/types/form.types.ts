/**
 * Form Generic Types
 * Comprehensive form handling with validation, state management, and field types
 */

import type { Component, Snippet } from 'svelte';
import type { Size, ColorVariant, ValidationState } from './common.types.js';

// ============================================================================
// FORM GENERIC TYPE
// ============================================================================

/**
 * Generic Form Type
 * @template TValues - Form values object type
 * @template TValidation - Validation schema type
 */
export interface Form<
  TValues extends Record<string, unknown> = Record<string, unknown>,
  TValidation extends FormValidation<TValues> = FormValidation<TValues>
> {
  // State
  values: TValues;
  initialValues: TValues;
  errors: FormErrors<TValues>;
  touched: FormTouched<TValues>;
  dirty: FormDirty<TValues>;

  // Status
  status: FormStatus;
  isValid: boolean;
  isDirty: boolean;
  isSubmitting: boolean;
  isValidating: boolean;
  submitCount: number;

  // Validation
  validation: TValidation;

  // Methods
  setFieldValue: <K extends keyof TValues>(field: K, value: TValues[K]) => void;
  setFieldError: <K extends keyof TValues>(field: K, error: string | null) => void;
  setFieldTouched: <K extends keyof TValues>(field: K, touched?: boolean) => void;
  setValues: (values: Partial<TValues>) => void;
  setErrors: (errors: Partial<FormErrors<TValues>>) => void;

  validateField: <K extends keyof TValues>(field: K) => Promise<string | null>;
  validateForm: () => Promise<boolean>;

  resetField: <K extends keyof TValues>(field: K) => void;
  resetForm: (nextValues?: TValues) => void;

  submitForm: () => Promise<void>;

  // Handlers
  handleChange: (e: Event) => void;
  handleBlur: (e: Event) => void;
  handleSubmit: (e?: Event) => Promise<void>;
  handleReset: (e?: Event) => void;

  // Field helpers
  getFieldProps: <K extends keyof TValues>(field: K) => FieldProps<TValues[K]>;
  getFieldMeta: <K extends keyof TValues>(field: K) => FieldMeta;
  getFieldState: <K extends keyof TValues>(field: K) => FieldState;

  // Events
  onSubmit?: (values: TValues) => void | Promise<void>;
  onReset?: () => void;
  onChange?: (values: TValues) => void;
  onError?: (errors: FormErrors<TValues>) => void;
}

/** Form status */
export type FormStatus = 'idle' | 'submitting' | 'success' | 'error';

/** Form errors */
export type FormErrors<TValues> = {
  [K in keyof TValues]?: string;
};

/** Form touched state */
export type FormTouched<TValues> = {
  [K in keyof TValues]?: boolean;
};

/** Form dirty state */
export type FormDirty<TValues> = {
  [K in keyof TValues]?: boolean;
};

/** Field props */
export interface FieldProps<TValue> {
  name: string;
  value: TValue;
  onChange: (value: TValue) => void;
  onBlur: () => void;
}

/** Field meta */
export interface FieldMeta {
  touched: boolean;
  dirty: boolean;
  error: string | null;
  valid: boolean;
  validating: boolean;
}

/** Field state */
export interface FieldState {
  value: unknown;
  touched: boolean;
  dirty: boolean;
  error: string | null;
  validationState: ValidationState;
}

// ============================================================================
// FORM VALIDATION
// ============================================================================

/** Form validation config */
export interface FormValidation<TValues> {
  schema?: ValidationSchema<TValues>;
  rules: ValidationRules<TValues>;
  mode: 'onChange' | 'onBlur' | 'onSubmit' | 'all';
  revalidateMode: 'onChange' | 'onBlur' | 'onSubmit';
}

/** Validation schema */
export type ValidationSchema<TValues> = {
  [K in keyof TValues]?: FieldValidation<TValues[K]>;
};

/** Validation rules */
export type ValidationRules<TValues> = {
  [K in keyof TValues]?: ValidationRule<TValues[K]>[];
};

/** Field validation */
export interface FieldValidation<TValue> {
  required?: boolean | string;
  min?: number | { value: number; message: string };
  max?: number | { value: number; message: string };
  minLength?: number | { value: number; message: string };
  maxLength?: number | { value: number; message: string };
  pattern?: RegExp | { value: RegExp; message: string };
  email?: boolean | string;
  url?: boolean | string;
  validate?: ValidateFn<TValue>;
  deps?: string[]; // Dependent fields
}

/** Validate function */
export type ValidateFn<TValue> = (
  value: TValue,
  formValues?: Record<string, unknown>
) => boolean | string | Promise<boolean | string>;

/** Validation rule */
export interface ValidationRule<TValue = unknown> {
  validate: (value: TValue, formValues?: Record<string, unknown>) => boolean | Promise<boolean>;
  message: string | ((value: TValue) => string);
  priority?: number;
}

// ============================================================================
// FORM FIELD TYPES
// ============================================================================

/** Base field interface */
export interface FormField<TValue = unknown> {
  name: string;
  label?: string;
  placeholder?: string;
  helperText?: string;
  required?: boolean;
  disabled?: boolean;
  readonly?: boolean;
  hidden?: boolean;
  defaultValue?: TValue;
  validation?: FieldValidation<TValue>;
  condition?: (values: Record<string, unknown>) => boolean;
  transform?: {
    input?: (value: unknown) => TValue;
    output?: (value: TValue) => unknown;
  };
}

/** Text field */
export interface TextField extends FormField<string> {
  type: 'text' | 'email' | 'password' | 'tel' | 'url' | 'search';
  minLength?: number;
  maxLength?: number;
  pattern?: string;
  autocomplete?: string;
  inputMode?: 'text' | 'numeric' | 'decimal' | 'tel' | 'email' | 'url' | 'search';
  clearable?: boolean;
  showPassword?: boolean; // For password type
  prefix?: string;
  suffix?: string;
}

/** Number field */
export interface NumberField extends FormField<number | null> {
  type: 'number';
  min?: number;
  max?: number;
  step?: number;
  precision?: number;
  format?: 'integer' | 'decimal' | 'currency' | 'percent';
  currency?: string;
  locale?: string;
  showButtons?: boolean;
  prefix?: string;
  suffix?: string;
}

/** Select field */
export interface SelectField<TOption = unknown> extends FormField<TOption | TOption[] | null> {
  type: 'select';
  options: SelectOption<TOption>[];
  groups?: SelectOptionGroup<TOption>[];
  multiple?: boolean;
  searchable?: boolean;
  clearable?: boolean;
  creatable?: boolean;
  loading?: boolean;
  loadOptions?: (query: string) => Promise<SelectOption<TOption>[]>;
  optionLabel?: keyof TOption | ((option: TOption) => string);
  optionValue?: keyof TOption | ((option: TOption) => unknown);
  optionDisabled?: keyof TOption | ((option: TOption) => boolean);
  groupBy?: keyof TOption | ((option: TOption) => string);
  maxSelections?: number;
}

/** Select option */
export interface SelectOption<TValue = unknown> {
  label: string;
  value: TValue;
  disabled?: boolean;
  icon?: string;
  description?: string;
  group?: string;
}

/** Select option group */
export interface SelectOptionGroup<TValue = unknown> {
  label: string;
  options: SelectOption<TValue>[];
}

/** Date field */
export interface DateField extends FormField<Date | null> {
  type: 'date' | 'datetime' | 'time' | 'month' | 'year';
  min?: Date;
  max?: Date;
  format?: string;
  locale?: string;
  clearable?: boolean;
  showTime?: boolean;
  timeFormat?: '12' | '24';
  disabledDates?: Date[] | ((date: Date) => boolean);
  firstDayOfWeek?: 0 | 1 | 2 | 3 | 4 | 5 | 6;
}

/** Date range field */
export interface DateRangeField extends FormField<{ start: Date | null; end: Date | null }> {
  type: 'daterange';
  minDate?: Date;
  maxDate?: Date;
  minDays?: number;
  maxDays?: number;
  format?: string;
  locale?: string;
  presets?: DateRangePreset[];
  clearable?: boolean;
}

/** Date range preset */
export interface DateRangePreset {
  label: string;
  range: { start: Date; end: Date };
}

/** Checkbox field */
export interface CheckboxField extends FormField<boolean> {
  type: 'checkbox';
  indeterminate?: boolean;
  labelPosition?: 'left' | 'right';
}

/** Checkbox group field */
export interface CheckboxGroupField<TValue = unknown> extends FormField<TValue[]> {
  type: 'checkbox-group';
  options: Array<{ label: string; value: TValue; disabled?: boolean; description?: string }>;
  orientation?: 'horizontal' | 'vertical';
  columns?: number;
  minSelections?: number;
  maxSelections?: number;
}

/** Radio field */
export interface RadioField<TValue = unknown> extends FormField<TValue> {
  type: 'radio';
  options: Array<{ label: string; value: TValue; disabled?: boolean; description?: string }>;
  orientation?: 'horizontal' | 'vertical';
  columns?: number;
}

/** Switch/Toggle field */
export interface SwitchField extends FormField<boolean> {
  type: 'switch';
  onLabel?: string;
  offLabel?: string;
  size?: Size;
}

/** Textarea field */
export interface TextareaField extends FormField<string> {
  type: 'textarea';
  rows?: number;
  minRows?: number;
  maxRows?: number;
  autoResize?: boolean;
  minLength?: number;
  maxLength?: number;
  showCount?: boolean;
  resize?: 'none' | 'vertical' | 'horizontal' | 'both';
}

/** Rich text editor field */
export interface RichTextField extends FormField<string> {
  type: 'richtext';
  toolbar?: RichTextToolbarItem[];
  minHeight?: string;
  maxHeight?: string;
  modules?: Record<string, unknown>;
  formats?: string[];
}

/** Rich text toolbar item */
export type RichTextToolbarItem =
  | 'bold'
  | 'italic'
  | 'underline'
  | 'strike'
  | 'heading'
  | 'blockquote'
  | 'code'
  | 'link'
  | 'image'
  | 'video'
  | 'list'
  | 'bullet'
  | 'indent'
  | 'align'
  | 'color'
  | 'background'
  | 'clean';

/** File upload field */
export interface FileField extends FormField<File | File[] | null> {
  type: 'file';
  accept?: string;
  multiple?: boolean;
  maxSize?: number; // in bytes
  maxFiles?: number;
  showPreview?: boolean;
  dragDrop?: boolean;
  uploadUrl?: string;
  uploadHandler?: (file: File) => Promise<string>;
}

/** Autocomplete field */
export interface AutocompleteField<TOption = unknown>
  extends FormField<TOption | TOption[] | null> {
  type: 'autocomplete';
  options?: TOption[];
  loadOptions?: (query: string) => Promise<TOption[]>;
  multiple?: boolean;
  minChars?: number;
  debounceMs?: number;
  optionLabel?: keyof TOption | ((option: TOption) => string);
  optionValue?: keyof TOption | ((option: TOption) => unknown);
  freeSolo?: boolean;
  clearable?: boolean;
  loading?: boolean;
}

/** Color picker field */
export interface ColorField extends FormField<string> {
  type: 'color';
  format?: 'hex' | 'rgb' | 'hsl';
  presets?: string[];
  showInput?: boolean;
  showAlpha?: boolean;
}

/** Slider field */
export interface SliderField extends FormField<number | [number, number]> {
  type: 'slider';
  min: number;
  max: number;
  step?: number;
  range?: boolean;
  marks?: Array<{ value: number; label?: string }>;
  showTooltip?: boolean | 'always';
  showInput?: boolean;
}

/** Rating field */
export interface RatingField extends FormField<number> {
  type: 'rating';
  max?: number;
  allowHalf?: boolean;
  allowClear?: boolean;
  icon?: string;
  size?: Size;
}

/** Array/Repeater field */
export interface ArrayField<TItem = Record<string, unknown>> extends FormField<TItem[]> {
  type: 'array';
  itemFields: FormFieldConfig[];
  minItems?: number;
  maxItems?: number;
  addLabel?: string;
  removeLabel?: string;
  sortable?: boolean;
  collapsible?: boolean;
  defaultItem?: TItem;
}

/** Object/Group field */
export interface ObjectField<TValue = Record<string, unknown>> extends FormField<TValue> {
  type: 'object';
  fields: FormFieldConfig[];
  columns?: number;
  collapsible?: boolean;
  defaultCollapsed?: boolean;
}

/** Hidden field */
export interface HiddenField<TValue = unknown> extends FormField<TValue> {
  type: 'hidden';
}

/** Custom field */
export interface CustomField<TValue = unknown> extends FormField<TValue> {
  type: 'custom';
  component: Component;
  props?: Record<string, unknown>;
}

/** Union type for all field configs */
export type FormFieldConfig =
  | TextField
  | NumberField
  | SelectField
  | DateField
  | DateRangeField
  | CheckboxField
  | CheckboxGroupField
  | RadioField
  | SwitchField
  | TextareaField
  | RichTextField
  | FileField
  | AutocompleteField
  | ColorField
  | SliderField
  | RatingField
  | ArrayField
  | ObjectField
  | HiddenField
  | CustomField;

// ============================================================================
// FORM SCHEMA/BUILDER
// ============================================================================

/** Form schema */
export interface FormSchema<TValues extends Record<string, unknown>> {
  fields: FormFieldConfig[];
  layout?: FormLayout;
  validation?: FormValidation<TValues>;
  submission?: FormSubmission<TValues>;
}

/** Form layout */
export interface FormLayout {
  type: 'vertical' | 'horizontal' | 'inline' | 'grid';
  columns?: number;
  gap?: Size;
  labelWidth?: string;
  labelPosition?: 'top' | 'left' | 'right';
  sections?: FormSection[];
  responsive?: {
    sm?: Partial<FormLayout>;
    md?: Partial<FormLayout>;
    lg?: Partial<FormLayout>;
  };
}

/** Form section */
export interface FormSection {
  id: string;
  title?: string;
  description?: string;
  icon?: string;
  fields: string[]; // Field names
  collapsible?: boolean;
  defaultCollapsed?: boolean;
  columns?: number;
  condition?: (values: Record<string, unknown>) => boolean;
}

/** Form submission config */
export interface FormSubmission<TValues> {
  endpoint?: string;
  method?: 'POST' | 'PUT' | 'PATCH';
  headers?: Record<string, string>;
  transform?: (values: TValues) => unknown;
  onSuccess?: (response: unknown) => void;
  onError?: (error: unknown) => void;
  resetOnSuccess?: boolean;
  redirectOnSuccess?: string;
}

// ============================================================================
// FORM SLOTS
// ============================================================================

/** Form slots */
export interface FormSlots<TValues = unknown> {
  header?: Snippet<[TValues]>;
  footer?: Snippet<[TValues]>;
  actions?: Snippet<[TValues, FormStatus]>;
  field?: Snippet<[FormFieldConfig, FieldState]>;
}
