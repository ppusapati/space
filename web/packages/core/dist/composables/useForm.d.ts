import { Writable, Readable } from 'svelte/store';
import { FormStatus, FormErrors, FormTouched, FormDirty, FieldProps, FieldMeta, FieldState, FormValidation } from '../types/index.js';
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
    values: Writable<TValues>;
    initialValues: Writable<TValues>;
    errors: Writable<FormErrors<TValues>>;
    touched: Writable<FormTouched<TValues>>;
    dirty: Writable<FormDirty<TValues>>;
    status: Writable<FormStatus>;
    isValid: Readable<boolean>;
    isDirty: Readable<boolean>;
    isSubmitting: Readable<boolean>;
    isValidating: Readable<boolean>;
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
    handleChange: (e: Event) => void;
    handleBlur: (e: Event) => void;
    handleSubmit: (e?: Event) => Promise<void>;
    handleReset: (e?: Event) => void;
    getFieldProps: <K extends keyof TValues>(field: K) => FieldProps<TValues[K]>;
    getFieldMeta: <K extends keyof TValues>(field: K) => FieldMeta;
    getFieldState: <K extends keyof TValues>(field: K) => FieldState;
    registerField: (name: string, defaultValue?: unknown) => void;
    unregisterField: (name: string) => void;
}
export declare function useForm<TValues extends Record<string, unknown>>(options: UseFormOptions<TValues>): UseFormReturn<TValues>;
//# sourceMappingURL=useForm.d.ts.map