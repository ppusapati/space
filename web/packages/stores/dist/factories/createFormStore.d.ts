/**
 * Form Store Factory
 * Creates a reusable store for form state management
 */
import { type Writable, type Readable } from 'svelte/store';
import type { FormStatus, FormValidation } from '@samavāya/core';
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
    subscribe: Writable<FormState<TValues>>['subscribe'];
    values: Readable<TValues>;
    errors: Readable<Partial<Record<keyof TValues, string>>>;
    touched: Readable<Partial<Record<keyof TValues, boolean>>>;
    dirty: Readable<Partial<Record<keyof TValues, boolean>>>;
    status: Readable<FormStatus>;
    isValid: Readable<boolean>;
    isDirty: Readable<boolean>;
    isSubmitting: Readable<boolean>;
    isValidating: Readable<boolean>;
    setFieldValue: <K extends keyof TValues>(field: K, value: TValues[K]) => void;
    setFieldError: <K extends keyof TValues>(field: K, error: string | null) => void;
    setFieldTouched: <K extends keyof TValues>(field: K, touched?: boolean) => void;
    setValues: (values: Partial<TValues>) => void;
    setErrors: (errors: Partial<Record<keyof TValues, string>>) => void;
    validateField: <K extends keyof TValues>(field: K) => Promise<string | null>;
    validateForm: () => Promise<boolean>;
    resetField: <K extends keyof TValues>(field: K) => void;
    resetForm: (nextValues?: TValues) => void;
    submitForm: () => Promise<void>;
    handleChange: (e: Event) => void;
    handleBlur: (e: Event) => void;
    handleSubmit: (e?: Event) => Promise<void>;
    handleReset: (e?: Event) => void;
    getFieldValue: <K extends keyof TValues>(field: K) => TValues[K];
    hasError: (field: keyof TValues) => boolean;
    isTouched: (field: keyof TValues) => boolean;
    isFieldDirty: (field: keyof TValues) => boolean;
}
export declare function createFormStore<TValues extends Record<string, unknown>>(config: FormStoreConfig<TValues>): FormStoreReturn<TValues>;
//# sourceMappingURL=createFormStore.d.ts.map