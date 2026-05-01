/**
 * Form Store Factory
 * Creates a reusable store for form state management
 */

import { writable, derived, get, type Writable, type Readable } from 'svelte/store';
import type { FormStatus, FormValidation, ValidationRule, FieldValidation } from '@samavāya/core';

// ============================================================================
// TYPES
// ============================================================================

export interface FormStoreConfig<TValues extends Record<string, unknown>> {
  name: string;
  initialValues: TValues;
  validation?: FormValidation<TValues>;
  validateOnChange?: boolean;
  validateOnBlur?: boolean;
  onSubmit?: (values: TValues) => void | Promise<void>;
  onReset?: () => void;
  onChange?: (values: TValues) => void;
}

export interface FormState<TValues extends Record<string, unknown>> {
  values: TValues;
  initialValues: TValues;
  errors: Partial<Record<keyof TValues, string>>;
  touched: Partial<Record<keyof TValues, boolean>>;
  dirty: Partial<Record<keyof TValues, boolean>>;
  status: FormStatus;
  submitCount: number;
  isValidating: boolean;
}

export interface FormStoreReturn<TValues extends Record<string, unknown>> {
  // State
  subscribe: Writable<FormState<TValues>>['subscribe'];

  // Derived stores
  values: Readable<TValues>;
  errors: Readable<Partial<Record<keyof TValues, string>>>;
  touched: Readable<Partial<Record<keyof TValues, boolean>>>;
  dirty: Readable<Partial<Record<keyof TValues, boolean>>>;
  status: Readable<FormStatus>;
  isValid: Readable<boolean>;
  isDirty: Readable<boolean>;
  isSubmitting: Readable<boolean>;
  isValidating: Readable<boolean>;

  // Field methods
  setFieldValue: <K extends keyof TValues>(field: K, value: TValues[K]) => void;
  setFieldError: <K extends keyof TValues>(field: K, error: string | null) => void;
  setFieldTouched: <K extends keyof TValues>(field: K, touched?: boolean) => void;
  setValues: (values: Partial<TValues>) => void;
  setErrors: (errors: Partial<Record<keyof TValues, string>>) => void;

  // Validation
  validateField: <K extends keyof TValues>(field: K) => Promise<string | null>;
  validateForm: () => Promise<boolean>;

  // Form actions
  resetField: <K extends keyof TValues>(field: K) => void;
  resetForm: (nextValues?: TValues) => void;
  submitForm: () => Promise<void>;

  // Handlers
  handleChange: (e: Event) => void;
  handleBlur: (e: Event) => void;
  handleSubmit: (e?: Event) => Promise<void>;
  handleReset: (e?: Event) => void;

  // Utility
  getFieldValue: <K extends keyof TValues>(field: K) => TValues[K];
  hasError: (field: keyof TValues) => boolean;
  isTouched: (field: keyof TValues) => boolean;
  isFieldDirty: (field: keyof TValues) => boolean;
}

// ============================================================================
// FACTORY
// ============================================================================

export function createFormStore<TValues extends Record<string, unknown>>(
  config: FormStoreConfig<TValues>
): FormStoreReturn<TValues> {
  const {
    name,
    initialValues,
    validation,
    validateOnChange = false,
    validateOnBlur = true,
    onSubmit,
    onReset,
    onChange,
  } = config;

  // ============================================================================
  // INITIAL STATE
  // ============================================================================

  const initialState: FormState<TValues> = {
    values: structuredClone(initialValues),
    initialValues: structuredClone(initialValues),
    errors: {},
    touched: {},
    dirty: {},
    status: 'idle',
    submitCount: 0,
    isValidating: false,
  };

  // ============================================================================
  // STORE
  // ============================================================================

  const store = writable<FormState<TValues>>(initialState);
  const { subscribe, set, update } = store;

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const values: Readable<TValues> = derived(store, ($s) => $s.values);
  const errors: Readable<Partial<Record<keyof TValues, string>>> = derived(store, ($s) => $s.errors);
  const touched: Readable<Partial<Record<keyof TValues, boolean>>> = derived(store, ($s) => $s.touched);
  const dirty: Readable<Partial<Record<keyof TValues, boolean>>> = derived(store, ($s) => $s.dirty);
  const status: Readable<FormStatus> = derived(store, ($s) => $s.status);
  const isValid: Readable<boolean> = derived(store, ($s) => Object.keys($s.errors).length === 0);
  const isDirty: Readable<boolean> = derived(store, ($s) => Object.values($s.dirty).some(Boolean));
  const isSubmitting: Readable<boolean> = derived(store, ($s) => $s.status === 'submitting');
  const isValidating: Readable<boolean> = derived(store, ($s) => $s.isValidating);

  // ============================================================================
  // FIELD METHODS
  // ============================================================================

  function setFieldValue<K extends keyof TValues>(field: K, value: TValues[K]): void {
    update((s) => {
      const newValues = { ...s.values, [field]: value };
      const isDirty = JSON.stringify(value) !== JSON.stringify(s.initialValues[field]);

      onChange?.(newValues);

      return {
        ...s,
        values: newValues,
        dirty: { ...s.dirty, [field]: isDirty },
      };
    });

    if (validateOnChange) {
      validateField(field);
    }
  }

  function setFieldError<K extends keyof TValues>(field: K, error: string | null): void {
    update((s) => {
      if (error === null) {
        const newErrors = { ...s.errors };
        delete newErrors[field];
        return { ...s, errors: newErrors };
      }
      return { ...s, errors: { ...s.errors, [field]: error } };
    });
  }

  function setFieldTouched<K extends keyof TValues>(field: K, isTouched = true): void {
    update((s) => ({
      ...s,
      touched: { ...s.touched, [field]: isTouched },
    }));
  }

  function setValues(newValues: Partial<TValues>): void {
    update((s) => {
      const values = { ...s.values, ...newValues };
      const dirty = { ...s.dirty };

      for (const key of Object.keys(newValues) as (keyof TValues)[]) {
        dirty[key] = JSON.stringify(newValues[key]) !== JSON.stringify(s.initialValues[key]);
      }

      onChange?.(values);
      return { ...s, values, dirty };
    });
  }

  function setErrors(newErrors: Partial<Record<keyof TValues, string>>): void {
    update((s) => ({
      ...s,
      errors: { ...s.errors, ...newErrors },
    }));
  }

  // ============================================================================
  // VALIDATION
  // ============================================================================

  async function validateField<K extends keyof TValues>(field: K): Promise<string | null> {
    if (!validation) return null;

    const state = get(store);
    const value = state.values[field];

    // Check schema validation
    if (validation.schema?.[field]) {
      const fieldValidation = validation.schema[field] as FieldValidation<TValues[K]>;
      const error = await runFieldValidation(field, value, fieldValidation, state.values);
      if (error) {
        setFieldError(field, error);
        return error;
      }
    }

    // Check rules
    if (validation.rules?.[field]) {
      const rules = validation.rules[field] as ValidationRule<TValues[K]>[];
      for (const rule of rules) {
        try {
          const isValid = await rule.validate(value, state.values as Record<string, unknown>);
          if (!isValid) {
            const error = typeof rule.message === 'function' ? rule.message(value) : rule.message;
            setFieldError(field, error);
            return error;
          }
        } catch {
          const error = 'Validation error';
          setFieldError(field, error);
          return error;
        }
      }
    }

    setFieldError(field, null);
    return null;
  }

  async function runFieldValidation<K extends keyof TValues>(
    field: K,
    value: TValues[K],
    fieldValidation: FieldValidation<TValues[K]>,
    allValues: TValues
  ): Promise<string | null> {
    // Required check
    if (fieldValidation.required) {
      const isEmpty = value == null || value === '' || (Array.isArray(value) && value.length === 0);
      if (isEmpty) {
        return typeof fieldValidation.required === 'string'
          ? fieldValidation.required
          : `${String(field)} is required`;
      }
    }

    if (value == null || value === '') return null;

    // Min/max for numbers
    if (typeof value === 'number') {
      if (fieldValidation.min !== undefined) {
        const minVal = typeof fieldValidation.min === 'number' ? fieldValidation.min : fieldValidation.min.value;
        const message = typeof fieldValidation.min === 'number' ? `Must be at least ${minVal}` : fieldValidation.min.message;
        if (value < minVal) return message;
      }
      if (fieldValidation.max !== undefined) {
        const maxVal = typeof fieldValidation.max === 'number' ? fieldValidation.max : fieldValidation.max.value;
        const message = typeof fieldValidation.max === 'number' ? `Must be at most ${maxVal}` : fieldValidation.max.message;
        if (value > maxVal) return message;
      }
    }

    // MinLength/maxLength for strings
    if (typeof value === 'string') {
      if (fieldValidation.minLength !== undefined) {
        const minLen = typeof fieldValidation.minLength === 'number' ? fieldValidation.minLength : fieldValidation.minLength.value;
        const message = typeof fieldValidation.minLength === 'number' ? `Must be at least ${minLen} characters` : fieldValidation.minLength.message;
        if (value.length < minLen) return message;
      }
      if (fieldValidation.maxLength !== undefined) {
        const maxLen = typeof fieldValidation.maxLength === 'number' ? fieldValidation.maxLength : fieldValidation.maxLength.value;
        const message = typeof fieldValidation.maxLength === 'number' ? `Must be at most ${maxLen} characters` : fieldValidation.maxLength.message;
        if (value.length > maxLen) return message;
      }
    }

    // Pattern check
    if (fieldValidation.pattern && typeof value === 'string') {
      const pattern = fieldValidation.pattern instanceof RegExp ? fieldValidation.pattern : fieldValidation.pattern.value;
      const message = fieldValidation.pattern instanceof RegExp ? 'Invalid format' : fieldValidation.pattern.message;
      if (!pattern.test(value)) return message;
    }

    // Email check
    if (fieldValidation.email && typeof value === 'string') {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      const message = typeof fieldValidation.email === 'string' ? fieldValidation.email : 'Invalid email address';
      if (!emailRegex.test(value)) return message;
    }

    // Custom validate
    if (fieldValidation.validate) {
      const result = await fieldValidation.validate(value, allValues as Record<string, unknown>);
      if (typeof result === 'string') return result;
      if (result === false) return `${String(field)} is invalid`;
    }

    return null;
  }

  async function validateForm(): Promise<boolean> {
    update((s) => ({ ...s, isValidating: true }));

    const state = get(store);
    const newErrors: Partial<Record<keyof TValues, string>> = {};

    if (validation) {
      // Validate schema
      if (validation.schema) {
        for (const field of Object.keys(validation.schema) as (keyof TValues)[]) {
          const fieldValidation = validation.schema[field];
          if (fieldValidation) {
            const error = await runFieldValidation(
              field,
              state.values[field],
              fieldValidation as FieldValidation<TValues[typeof field]>,
              state.values
            );
            if (error) {
              newErrors[field] = error;
            }
          }
        }
      }

      // Validate rules
      if (validation.rules) {
        for (const field of Object.keys(validation.rules) as (keyof TValues)[]) {
          if (newErrors[field]) continue;

          const rules = validation.rules[field];
          if (rules) {
            for (const rule of rules as ValidationRule<TValues[typeof field]>[]) {
              try {
                const isValid = await rule.validate(state.values[field], state.values as Record<string, unknown>);
                if (!isValid) {
                  newErrors[field] = typeof rule.message === 'function' ? rule.message(state.values[field]) : rule.message;
                  break;
                }
              } catch {
                newErrors[field] = 'Validation error';
                break;
              }
            }
          }
        }
      }
    }

    update((s) => ({ ...s, errors: newErrors, isValidating: false }));
    return Object.keys(newErrors).length === 0;
  }

  // ============================================================================
  // FORM ACTIONS
  // ============================================================================

  function resetField<K extends keyof TValues>(field: K): void {
    const state = get(store);
    update((s) => ({
      ...s,
      values: { ...s.values, [field]: s.initialValues[field] },
      errors: (() => {
        const newErrors = { ...s.errors };
        delete newErrors[field];
        return newErrors;
      })(),
      touched: (() => {
        const newTouched = { ...s.touched };
        delete newTouched[field];
        return newTouched;
      })(),
      dirty: (() => {
        const newDirty = { ...s.dirty };
        delete newDirty[field];
        return newDirty;
      })(),
    }));
  }

  function resetForm(nextValues?: TValues): void {
    const resetTo = nextValues ?? get(store).initialValues;
    set({
      ...initialState,
      values: structuredClone(resetTo),
      initialValues: nextValues ? structuredClone(nextValues) : get(store).initialValues,
    });
    onReset?.();
  }

  async function submitForm(): Promise<void> {
    update((s) => ({
      ...s,
      submitCount: s.submitCount + 1,
      touched: Object.keys(s.values).reduce(
        (acc, key) => ({ ...acc, [key]: true }),
        {} as Record<keyof TValues, boolean>
      ),
    }));

    const isValid = await validateForm();
    if (!isValid) {
      update((s) => ({ ...s, status: 'error' }));
      return;
    }

    update((s) => ({ ...s, status: 'submitting' }));

    try {
      await onSubmit?.(get(store).values);
      update((s) => ({ ...s, status: 'success' }));
    } catch {
      update((s) => ({ ...s, status: 'error' }));
    }
  }

  // ============================================================================
  // EVENT HANDLERS
  // ============================================================================

  function handleChange(e: Event): void {
    const target = e.target as HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement;
    const { name, type } = target;

    let value: unknown;
    if (type === 'checkbox') {
      value = (target as HTMLInputElement).checked;
    } else if (type === 'number' || type === 'range') {
      value = target.value === '' ? null : Number(target.value);
    } else {
      value = target.value;
    }

    setFieldValue(name as keyof TValues, value as TValues[keyof TValues]);
  }

  function handleBlur(e: Event): void {
    const target = e.target as HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement;
    const { name } = target;

    setFieldTouched(name as keyof TValues, true);

    if (validateOnBlur) {
      validateField(name as keyof TValues);
    }
  }

  async function handleSubmit(e?: Event): Promise<void> {
    e?.preventDefault();
    await submitForm();
  }

  function handleReset(e?: Event): void {
    e?.preventDefault();
    resetForm();
  }

  // ============================================================================
  // UTILITY
  // ============================================================================

  function getFieldValue<K extends keyof TValues>(field: K): TValues[K] {
    return get(store).values[field];
  }

  function hasError(field: keyof TValues): boolean {
    return field in get(store).errors;
  }

  function isTouched(field: keyof TValues): boolean {
    return get(store).touched[field] ?? false;
  }

  function isFieldDirty(field: keyof TValues): boolean {
    return get(store).dirty[field] ?? false;
  }

  // ============================================================================
  // RETURN
  // ============================================================================

  return {
    subscribe,
    // Derived stores
    values,
    errors,
    touched,
    dirty,
    status,
    isValid,
    isDirty,
    isSubmitting,
    isValidating,
    // Field methods
    setFieldValue,
    setFieldError,
    setFieldTouched,
    setValues,
    setErrors,
    // Validation
    validateField,
    validateForm,
    // Form actions
    resetField,
    resetForm,
    submitForm,
    // Handlers
    handleChange,
    handleBlur,
    handleSubmit,
    handleReset,
    // Utility
    getFieldValue,
    hasError,
    isTouched,
    isFieldDirty,
  };
}
