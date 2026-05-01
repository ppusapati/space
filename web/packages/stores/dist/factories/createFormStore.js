/**
 * Form Store Factory
 * Creates a reusable store for form state management
 */
import { writable, derived, get } from 'svelte/store';
// ============================================================================
// FACTORY
// ============================================================================
export function createFormStore(config) {
    const { name, initialValues, validation, validateOnChange = false, validateOnBlur = true, onSubmit, onReset, onChange, } = config;
    // ============================================================================
    // INITIAL STATE
    // ============================================================================
    const initialState = {
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
    const store = writable(initialState);
    const { subscribe, set, update } = store;
    // ============================================================================
    // DERIVED STORES
    // ============================================================================
    const values = derived(store, ($s) => $s.values);
    const errors = derived(store, ($s) => $s.errors);
    const touched = derived(store, ($s) => $s.touched);
    const dirty = derived(store, ($s) => $s.dirty);
    const status = derived(store, ($s) => $s.status);
    const isValid = derived(store, ($s) => Object.keys($s.errors).length === 0);
    const isDirty = derived(store, ($s) => Object.values($s.dirty).some(Boolean));
    const isSubmitting = derived(store, ($s) => $s.status === 'submitting');
    const isValidating = derived(store, ($s) => $s.isValidating);
    // ============================================================================
    // FIELD METHODS
    // ============================================================================
    function setFieldValue(field, value) {
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
    function setFieldError(field, error) {
        update((s) => {
            if (error === null) {
                const newErrors = { ...s.errors };
                delete newErrors[field];
                return { ...s, errors: newErrors };
            }
            return { ...s, errors: { ...s.errors, [field]: error } };
        });
    }
    function setFieldTouched(field, isTouched = true) {
        update((s) => ({
            ...s,
            touched: { ...s.touched, [field]: isTouched },
        }));
    }
    function setValues(newValues) {
        update((s) => {
            const values = { ...s.values, ...newValues };
            const dirty = { ...s.dirty };
            for (const key of Object.keys(newValues)) {
                dirty[key] = JSON.stringify(newValues[key]) !== JSON.stringify(s.initialValues[key]);
            }
            onChange?.(values);
            return { ...s, values, dirty };
        });
    }
    function setErrors(newErrors) {
        update((s) => ({
            ...s,
            errors: { ...s.errors, ...newErrors },
        }));
    }
    // ============================================================================
    // VALIDATION
    // ============================================================================
    async function validateField(field) {
        if (!validation)
            return null;
        const state = get(store);
        const value = state.values[field];
        // Check schema validation
        if (validation.schema?.[field]) {
            const fieldValidation = validation.schema[field];
            const error = await runFieldValidation(field, value, fieldValidation, state.values);
            if (error) {
                setFieldError(field, error);
                return error;
            }
        }
        // Check rules
        if (validation.rules?.[field]) {
            const rules = validation.rules[field];
            for (const rule of rules) {
                try {
                    const isValid = await rule.validate(value, state.values);
                    if (!isValid) {
                        const error = typeof rule.message === 'function' ? rule.message(value) : rule.message;
                        setFieldError(field, error);
                        return error;
                    }
                }
                catch {
                    const error = 'Validation error';
                    setFieldError(field, error);
                    return error;
                }
            }
        }
        setFieldError(field, null);
        return null;
    }
    async function runFieldValidation(field, value, fieldValidation, allValues) {
        // Required check
        if (fieldValidation.required) {
            const isEmpty = value == null || value === '' || (Array.isArray(value) && value.length === 0);
            if (isEmpty) {
                return typeof fieldValidation.required === 'string'
                    ? fieldValidation.required
                    : `${String(field)} is required`;
            }
        }
        if (value == null || value === '')
            return null;
        // Min/max for numbers
        if (typeof value === 'number') {
            if (fieldValidation.min !== undefined) {
                const minVal = typeof fieldValidation.min === 'number' ? fieldValidation.min : fieldValidation.min.value;
                const message = typeof fieldValidation.min === 'number' ? `Must be at least ${minVal}` : fieldValidation.min.message;
                if (value < minVal)
                    return message;
            }
            if (fieldValidation.max !== undefined) {
                const maxVal = typeof fieldValidation.max === 'number' ? fieldValidation.max : fieldValidation.max.value;
                const message = typeof fieldValidation.max === 'number' ? `Must be at most ${maxVal}` : fieldValidation.max.message;
                if (value > maxVal)
                    return message;
            }
        }
        // MinLength/maxLength for strings
        if (typeof value === 'string') {
            if (fieldValidation.minLength !== undefined) {
                const minLen = typeof fieldValidation.minLength === 'number' ? fieldValidation.minLength : fieldValidation.minLength.value;
                const message = typeof fieldValidation.minLength === 'number' ? `Must be at least ${minLen} characters` : fieldValidation.minLength.message;
                if (value.length < minLen)
                    return message;
            }
            if (fieldValidation.maxLength !== undefined) {
                const maxLen = typeof fieldValidation.maxLength === 'number' ? fieldValidation.maxLength : fieldValidation.maxLength.value;
                const message = typeof fieldValidation.maxLength === 'number' ? `Must be at most ${maxLen} characters` : fieldValidation.maxLength.message;
                if (value.length > maxLen)
                    return message;
            }
        }
        // Pattern check
        if (fieldValidation.pattern && typeof value === 'string') {
            const pattern = fieldValidation.pattern instanceof RegExp ? fieldValidation.pattern : fieldValidation.pattern.value;
            const message = fieldValidation.pattern instanceof RegExp ? 'Invalid format' : fieldValidation.pattern.message;
            if (!pattern.test(value))
                return message;
        }
        // Email check
        if (fieldValidation.email && typeof value === 'string') {
            const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            const message = typeof fieldValidation.email === 'string' ? fieldValidation.email : 'Invalid email address';
            if (!emailRegex.test(value))
                return message;
        }
        // Custom validate
        if (fieldValidation.validate) {
            const result = await fieldValidation.validate(value, allValues);
            if (typeof result === 'string')
                return result;
            if (result === false)
                return `${String(field)} is invalid`;
        }
        return null;
    }
    async function validateForm() {
        update((s) => ({ ...s, isValidating: true }));
        const state = get(store);
        const newErrors = {};
        if (validation) {
            // Validate schema
            if (validation.schema) {
                for (const field of Object.keys(validation.schema)) {
                    const fieldValidation = validation.schema[field];
                    if (fieldValidation) {
                        const error = await runFieldValidation(field, state.values[field], fieldValidation, state.values);
                        if (error) {
                            newErrors[field] = error;
                        }
                    }
                }
            }
            // Validate rules
            if (validation.rules) {
                for (const field of Object.keys(validation.rules)) {
                    if (newErrors[field])
                        continue;
                    const rules = validation.rules[field];
                    if (rules) {
                        for (const rule of rules) {
                            try {
                                const isValid = await rule.validate(state.values[field], state.values);
                                if (!isValid) {
                                    newErrors[field] = typeof rule.message === 'function' ? rule.message(state.values[field]) : rule.message;
                                    break;
                                }
                            }
                            catch {
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
    function resetField(field) {
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
    function resetForm(nextValues) {
        const resetTo = nextValues ?? get(store).initialValues;
        set({
            ...initialState,
            values: structuredClone(resetTo),
            initialValues: nextValues ? structuredClone(nextValues) : get(store).initialValues,
        });
        onReset?.();
    }
    async function submitForm() {
        update((s) => ({
            ...s,
            submitCount: s.submitCount + 1,
            touched: Object.keys(s.values).reduce((acc, key) => ({ ...acc, [key]: true }), {}),
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
        }
        catch {
            update((s) => ({ ...s, status: 'error' }));
        }
    }
    // ============================================================================
    // EVENT HANDLERS
    // ============================================================================
    function handleChange(e) {
        const target = e.target;
        const { name, type } = target;
        let value;
        if (type === 'checkbox') {
            value = target.checked;
        }
        else if (type === 'number' || type === 'range') {
            value = target.value === '' ? null : Number(target.value);
        }
        else {
            value = target.value;
        }
        setFieldValue(name, value);
    }
    function handleBlur(e) {
        const target = e.target;
        const { name } = target;
        setFieldTouched(name, true);
        if (validateOnBlur) {
            validateField(name);
        }
    }
    async function handleSubmit(e) {
        e?.preventDefault();
        await submitForm();
    }
    function handleReset(e) {
        e?.preventDefault();
        resetForm();
    }
    // ============================================================================
    // UTILITY
    // ============================================================================
    function getFieldValue(field) {
        return get(store).values[field];
    }
    function hasError(field) {
        return field in get(store).errors;
    }
    function isTouched(field) {
        return get(store).touched[field] ?? false;
    }
    function isFieldDirty(field) {
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
