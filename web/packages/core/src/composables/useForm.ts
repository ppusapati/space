/**
 * useForm Composable
 * Creates a reactive form state with validation, field handling, and submission
 */

import { writable, derived, get, type Writable, type Readable } from 'svelte/store';
import type {
  Form,
  FormStatus,
  FormErrors,
  FormTouched,
  FormDirty,
  FieldProps,
  FieldMeta,
  FieldState,
  FormValidation,
  ValidationRule,
  FieldValidation,
  ValidationState,
} from '../types/index.js';

// ============================================================================
// TYPES
// ============================================================================

export interface UseFormOptions<TValues extends Record<string, unknown>> {
  initialValues: TValues;
  validation?: FormValidation<TValues>;
  validateOnChange?: boolean;
  validateOnBlur?: boolean;
  validateOnMount?: boolean;
  onSubmit?: (values: TValues) => void | Promise<void>;
  onReset?: () => void;
  onChange?: (values: TValues) => void;
  onError?: (errors: FormErrors<TValues>) => void;
}

export interface UseFormReturn<TValues extends Record<string, unknown>> {
  // Stores
  values: Writable<TValues>;
  initialValues: Writable<TValues>;
  errors: Writable<FormErrors<TValues>>;
  touched: Writable<FormTouched<TValues>>;
  dirty: Writable<FormDirty<TValues>>;
  status: Writable<FormStatus>;

  // Derived
  isValid: Readable<boolean>;
  isDirty: Readable<boolean>;
  isSubmitting: Readable<boolean>;
  isValidating: Readable<boolean>;

  // Field methods
  setFieldValue: <K extends keyof TValues>(field: K, value: TValues[K]) => void;
  setFieldError: <K extends keyof TValues>(field: K, error: string | null) => void;
  setFieldTouched: <K extends keyof TValues>(field: K, touched?: boolean) => void;
  setValues: (values: Partial<TValues>) => void;
  setErrors: (errors: Partial<FormErrors<TValues>>) => void;

  // Validation methods
  validateField: <K extends keyof TValues>(field: K) => Promise<string | null>;
  validateForm: () => Promise<boolean>;

  // Reset methods
  resetField: <K extends keyof TValues>(field: K) => void;
  resetForm: (nextValues?: TValues) => void;

  // Submit methods
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

  // Register field (for dynamic forms)
  registerField: (name: string, defaultValue?: unknown) => void;
  unregisterField: (name: string) => void;
}

// ============================================================================
// IMPLEMENTATION
// ============================================================================

export function useForm<TValues extends Record<string, unknown>>(
  options: UseFormOptions<TValues>
): UseFormReturn<TValues> {
  const {
    initialValues: initialVals,
    validation,
    validateOnChange = false,
    validateOnBlur = true,
    validateOnMount = false,
    onSubmit,
    onReset,
    onChange,
    onError,
  } = options;

  // ============================================================================
  // STORES
  // ============================================================================

  const values = writable<TValues>(structuredClone(initialVals));
  const initialValues = writable<TValues>(structuredClone(initialVals));
  const errors = writable<FormErrors<TValues>>({});
  const touched = writable<FormTouched<TValues>>({});
  const dirty = writable<FormDirty<TValues>>({});
  const status = writable<FormStatus>('idle');

  let isValidatingField = false;
  let submitCount = 0;

  // ============================================================================
  // DERIVED STORES
  // ============================================================================

  const isValid = derived(errors, ($errors) => Object.keys($errors).length === 0);

  const isDirty = derived(dirty, ($dirty) => Object.values($dirty).some(Boolean));

  const isSubmitting = derived(status, ($status) => $status === 'submitting');

  const isValidating = writable<boolean>(false);

  // ============================================================================
  // FIELD METHODS
  // ============================================================================

  function setFieldValue<K extends keyof TValues>(field: K, value: TValues[K]): void {
    values.update(($v) => ({ ...$v, [field]: value }));

    // Mark as dirty
    const $initial = get(initialValues);
    dirty.update(($d) => ({
      ...$d,
      [field]: JSON.stringify(value) !== JSON.stringify($initial[field]),
    }));

    // Validate on change if enabled
    if (validateOnChange && !isValidatingField) {
      validateField(field);
    }

    // Trigger onChange callback
    onChange?.(get(values));
  }

  function setFieldError<K extends keyof TValues>(field: K, error: string | null): void {
    errors.update(($e) => {
      if (error === null) {
        const newErrors = { ...$e };
        delete newErrors[field];
        return newErrors;
      }
      return { ...$e, [field]: error };
    });
  }

  function setFieldTouched<K extends keyof TValues>(field: K, isTouched = true): void {
    touched.update(($t) => ({ ...$t, [field]: isTouched }));
  }

  function setValues(newValues: Partial<TValues>): void {
    values.update(($v) => ({ ...$v, ...newValues }));

    // Update dirty state
    const $initial = get(initialValues);
    dirty.update(($d) => {
      const newDirty = { ...$d };
      for (const key of Object.keys(newValues) as (keyof TValues)[]) {
        newDirty[key] =
          JSON.stringify(newValues[key]) !== JSON.stringify($initial[key]);
      }
      return newDirty;
    });

    onChange?.(get(values));
  }

  function setErrors(newErrors: Partial<FormErrors<TValues>>): void {
    errors.update(($e) => ({ ...$e, ...newErrors }));
  }

  // ============================================================================
  // VALIDATION METHODS
  // ============================================================================

  async function validateField<K extends keyof TValues>(field: K): Promise<string | null> {
    if (!validation) return null;

    isValidatingField = true;
    const $values = get(values);
    const value = $values[field];

    // Check schema validation
    if (validation.schema?.[field]) {
      const fieldValidation = validation.schema[field] as FieldValidation<TValues[K]>;
      const error = await runFieldValidation(field, value, fieldValidation, $values);
      if (error) {
        setFieldError(field, error);
        isValidatingField = false;
        return error;
      }
    }

    // Check rules
    if (validation.rules?.[field]) {
      const rules = validation.rules[field] as ValidationRule<TValues[K]>[];
      for (const rule of rules) {
        try {
          const isValid = await rule.validate(value, $values as Record<string, unknown>);
          if (!isValid) {
            const error = typeof rule.message === 'function' ? rule.message(value) : rule.message;
            setFieldError(field, error);
            isValidatingField = false;
            return error;
          }
        } catch {
          const error = 'Validation error';
          setFieldError(field, error);
          isValidatingField = false;
          return error;
        }
      }
    }

    setFieldError(field, null);
    isValidatingField = false;
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

    // Skip other validations if empty and not required
    if (value == null || value === '') return null;

    // Min check (number)
    if (fieldValidation.min !== undefined && typeof value === 'number') {
      const minVal = typeof fieldValidation.min === 'number' ? fieldValidation.min : fieldValidation.min.value;
      const message = typeof fieldValidation.min === 'number'
        ? `Must be at least ${minVal}`
        : fieldValidation.min.message;
      if (value < minVal) return message;
    }

    // Max check (number)
    if (fieldValidation.max !== undefined && typeof value === 'number') {
      const maxVal = typeof fieldValidation.max === 'number' ? fieldValidation.max : fieldValidation.max.value;
      const message = typeof fieldValidation.max === 'number'
        ? `Must be at most ${maxVal}`
        : fieldValidation.max.message;
      if (value > maxVal) return message;
    }

    // MinLength check (string)
    if (fieldValidation.minLength !== undefined && typeof value === 'string') {
      const minLen = typeof fieldValidation.minLength === 'number'
        ? fieldValidation.minLength
        : fieldValidation.minLength.value;
      const message = typeof fieldValidation.minLength === 'number'
        ? `Must be at least ${minLen} characters`
        : fieldValidation.minLength.message;
      if (value.length < minLen) return message;
    }

    // MaxLength check (string)
    if (fieldValidation.maxLength !== undefined && typeof value === 'string') {
      const maxLen = typeof fieldValidation.maxLength === 'number'
        ? fieldValidation.maxLength
        : fieldValidation.maxLength.value;
      const message = typeof fieldValidation.maxLength === 'number'
        ? `Must be at most ${maxLen} characters`
        : fieldValidation.maxLength.message;
      if (value.length > maxLen) return message;
    }

    // Pattern check
    if (fieldValidation.pattern && typeof value === 'string') {
      const pattern = fieldValidation.pattern instanceof RegExp
        ? fieldValidation.pattern
        : fieldValidation.pattern.value;
      const message = fieldValidation.pattern instanceof RegExp
        ? 'Invalid format'
        : fieldValidation.pattern.message;
      if (!pattern.test(value)) return message;
    }

    // Email check
    if (fieldValidation.email && typeof value === 'string') {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      const message = typeof fieldValidation.email === 'string'
        ? fieldValidation.email
        : 'Invalid email address';
      if (!emailRegex.test(value)) return message;
    }

    // URL check
    if (fieldValidation.url && typeof value === 'string') {
      try {
        new URL(value);
      } catch {
        const message = typeof fieldValidation.url === 'string'
          ? fieldValidation.url
          : 'Invalid URL';
        return message;
      }
    }

    // Custom validate function
    if (fieldValidation.validate) {
      const result = await fieldValidation.validate(value, allValues as Record<string, unknown>);
      if (typeof result === 'string') return result;
      if (result === false) return `${String(field)} is invalid`;
    }

    return null;
  }

  async function validateForm(): Promise<boolean> {
    isValidating.set(true);
    const $values = get(values);
    const newErrors: FormErrors<TValues> = {};

    if (validation) {
      // Validate schema
      if (validation.schema) {
        for (const field of Object.keys(validation.schema) as (keyof TValues)[]) {
          const fieldValidation = validation.schema[field];
          if (fieldValidation) {
            const error = await runFieldValidation(
              field,
              $values[field],
              fieldValidation as FieldValidation<TValues[typeof field]>,
              $values
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
          if (newErrors[field]) continue; // Skip if already has error

          const rules = validation.rules[field];
          if (rules) {
            for (const rule of rules as ValidationRule<TValues[typeof field]>[]) {
              try {
                const isValid = await rule.validate(
                  $values[field],
                  $values as Record<string, unknown>
                );
                if (!isValid) {
                  newErrors[field] =
                    typeof rule.message === 'function'
                      ? rule.message($values[field])
                      : rule.message;
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

    errors.set(newErrors);
    isValidating.set(false);

    const isFormValid = Object.keys(newErrors).length === 0;
    if (!isFormValid) {
      onError?.(newErrors);
    }

    return isFormValid;
  }

  // ============================================================================
  // RESET METHODS
  // ============================================================================

  function resetField<K extends keyof TValues>(field: K): void {
    const $initial = get(initialValues);
    values.update(($v) => ({ ...$v, [field]: $initial[field] }));
    errors.update(($e) => {
      const newErrors = { ...$e };
      delete newErrors[field];
      return newErrors;
    });
    touched.update(($t) => {
      const newTouched = { ...$t };
      delete newTouched[field];
      return newTouched;
    });
    dirty.update(($d) => {
      const newDirty = { ...$d };
      delete newDirty[field];
      return newDirty;
    });
  }

  function resetForm(nextValues?: TValues): void {
    const resetTo = nextValues ?? get(initialValues);
    values.set(structuredClone(resetTo));
    if (nextValues) {
      initialValues.set(structuredClone(nextValues));
    }
    errors.set({});
    touched.set({});
    dirty.set({});
    status.set('idle');
    submitCount = 0;
    onReset?.();
  }

  // ============================================================================
  // SUBMIT METHODS
  // ============================================================================

  async function submitForm(): Promise<void> {
    submitCount++;

    // Mark all fields as touched
    const $values = get(values);
    const allTouched: FormTouched<TValues> = {};
    for (const key of Object.keys($values) as (keyof TValues)[]) {
      allTouched[key] = true;
    }
    touched.set(allTouched);

    // Validate
    const isFormValid = await validateForm();
    if (!isFormValid) {
      status.set('error');
      return;
    }

    // Submit
    status.set('submitting');

    try {
      await onSubmit?.($values);
      status.set('success');
    } catch (error) {
      status.set('error');
      throw error;
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
    } else if (type === 'file') {
      value = (target as HTMLInputElement).files;
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
  // FIELD HELPERS
  // ============================================================================

  function getFieldProps<K extends keyof TValues>(field: K): FieldProps<TValues[K]> {
    return {
      name: field as string,
      value: get(values)[field],
      onChange: (value: TValues[K]) => setFieldValue(field, value),
      onBlur: () => {
        setFieldTouched(field, true);
        if (validateOnBlur) validateField(field);
      },
    };
  }

  function getFieldMeta<K extends keyof TValues>(field: K): FieldMeta {
    const $errors = get(errors);
    const $touched = get(touched);
    const $dirty = get(dirty);
    const $isValidating = get(isValidating);

    return {
      touched: $touched[field] ?? false,
      dirty: $dirty[field] ?? false,
      error: $errors[field] ?? null,
      valid: !$errors[field],
      validating: $isValidating,
    };
  }

  function getFieldState<K extends keyof TValues>(field: K): FieldState {
    const $values = get(values);
    const $errors = get(errors);
    const $touched = get(touched);
    const $dirty = get(dirty);

    const error = $errors[field] ?? null;
    const isTouched = $touched[field] ?? false;

    let validationState: ValidationState = 'default';
    if (isTouched) {
      if (error) {
        validationState = 'invalid';
      } else if ($dirty[field]) {
        validationState = 'valid';
      }
    }

    return {
      value: $values[field],
      touched: isTouched,
      dirty: $dirty[field] ?? false,
      error,
      validationState,
    };
  }

  // ============================================================================
  // DYNAMIC FIELD REGISTRATION
  // ============================================================================

  function registerField(name: string, defaultValue?: unknown): void {
    values.update(($v) => {
      if (name in $v) return $v;
      return { ...$v, [name]: defaultValue };
    });
    initialValues.update(($iv) => {
      if (name in $iv) return $iv;
      return { ...$iv, [name]: defaultValue };
    });
  }

  function unregisterField(name: string): void {
    values.update(($v) => {
      const newValues = { ...$v };
      delete newValues[name];
      return newValues;
    });
    errors.update(($e) => {
      const newErrors = { ...$e };
      delete newErrors[name as keyof TValues];
      return newErrors;
    });
    touched.update(($t) => {
      const newTouched = { ...$t };
      delete newTouched[name as keyof TValues];
      return newTouched;
    });
    dirty.update(($d) => {
      const newDirty = { ...$d };
      delete newDirty[name as keyof TValues];
      return newDirty;
    });
  }

  // ============================================================================
  // INITIALIZE
  // ============================================================================

  if (validateOnMount) {
    validateForm();
  }

  // ============================================================================
  // RETURN
  // ============================================================================

  return {
    // Stores
    values,
    initialValues,
    errors,
    touched,
    dirty,
    status,

    // Derived
    isValid,
    isDirty,
    isSubmitting,
    isValidating: { subscribe: isValidating.subscribe } as Readable<boolean>,

    // Field methods
    setFieldValue,
    setFieldError,
    setFieldTouched,
    setValues,
    setErrors,

    // Validation methods
    validateField,
    validateForm,

    // Reset methods
    resetField,
    resetForm,

    // Submit methods
    submitForm,

    // Handlers
    handleChange,
    handleBlur,
    handleSubmit,
    handleReset,

    // Field helpers
    getFieldProps,
    getFieldMeta,
    getFieldState,

    // Dynamic fields
    registerField,
    unregisterField,
  };
}
