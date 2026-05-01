/**
 * Schema adapters for Zod and Yup validation libraries
 * These adapters convert schema validation to our ValidationRule format
 */

import type { ValidationRule, ValidationResult } from './index';

// Type definitions for Zod (optional peer dependency)
interface ZodSchema<T = unknown> {
  safeParse(data: unknown): { success: true; data: T } | { success: false; error: ZodError };
  safeParseAsync(data: unknown): Promise<{ success: true; data: T } | { success: false; error: ZodError }>;
}

interface ZodError {
  errors: Array<{ message: string; path: (string | number)[] }>;
  format(): Record<string, { _errors: string[] }>;
}

// Type definitions for Yup (optional peer dependency)
interface YupSchema<T = unknown> {
  validateSync(value: unknown, options?: { abortEarly?: boolean }): T;
  validate(value: unknown, options?: { abortEarly?: boolean }): Promise<T>;
  isValidSync(value: unknown): boolean;
  isValid(value: unknown): Promise<boolean>;
}

interface YupError {
  errors: string[];
  inner: Array<{ path: string; message: string; errors: string[] }>;
}

/**
 * Result type for schema validation
 */
export interface SchemaValidationResult<T = unknown> extends ValidationResult {
  data?: T;
  fieldErrors: Record<string, string[]>;
}

/**
 * Convert a Zod schema to ValidationRule array for a single field
 */
export function zodToRule<T>(
  schema: ZodSchema<T>,
  message?: string
): ValidationRule<T> {
  return {
    validate: (value: T) => {
      const result = schema.safeParse(value);
      return result.success;
    },
    message: message || 'Validation failed'
  };
}

/**
 * Convert a Zod schema to async ValidationRule
 */
export function zodToAsyncRule<T>(
  schema: ZodSchema<T>,
  message?: string
): ValidationRule<T> {
  return {
    validate: async (value: T) => {
      const result = await schema.safeParseAsync(value);
      return result.success;
    },
    message: message || 'Validation failed'
  };
}

/**
 * Validate entire form data against a Zod schema
 */
export function validateWithZod<T>(
  schema: ZodSchema<T>,
  data: unknown
): SchemaValidationResult<T> {
  const result = schema.safeParse(data);

  if (result.success) {
    return {
      isValid: true,
      errors: [],
      fieldErrors: {},
      data: result.data
    };
  }

  const errors: string[] = [];
  const fieldErrors: Record<string, string[]> = {};

  for (const issue of result.error.errors) {
    const path = issue.path.join('.');
    errors.push(issue.message);

    if (path) {
      if (!fieldErrors[path]) {
        fieldErrors[path] = [];
      }
      fieldErrors[path].push(issue.message);
    }
  }

  return {
    isValid: false,
    errors,
    fieldErrors
  };
}

/**
 * Async validation with Zod schema
 */
export async function validateWithZodAsync<T>(
  schema: ZodSchema<T>,
  data: unknown
): Promise<SchemaValidationResult<T>> {
  const result = await schema.safeParseAsync(data);

  if (result.success) {
    return {
      isValid: true,
      errors: [],
      fieldErrors: {},
      data: result.data
    };
  }

  const errors: string[] = [];
  const fieldErrors: Record<string, string[]> = {};

  for (const issue of result.error.errors) {
    const path = issue.path.join('.');
    errors.push(issue.message);

    if (path) {
      if (!fieldErrors[path]) {
        fieldErrors[path] = [];
      }
      fieldErrors[path].push(issue.message);
    }
  }

  return {
    isValid: false,
    errors,
    fieldErrors
  };
}

/**
 * Convert a Yup schema to ValidationRule for a single field
 */
export function yupToRule<T>(
  schema: YupSchema<T>,
  message?: string
): ValidationRule<T> {
  return {
    validate: (value: T) => {
      try {
        schema.validateSync(value);
        return true;
      } catch {
        return false;
      }
    },
    message: message || 'Validation failed'
  };
}

/**
 * Convert a Yup schema to async ValidationRule
 */
export function yupToAsyncRule<T>(
  schema: YupSchema<T>,
  message?: string
): ValidationRule<T> {
  return {
    validate: async (value: T) => {
      return schema.isValid(value);
    },
    message: message || 'Validation failed'
  };
}

/**
 * Validate entire form data against a Yup schema
 */
export function validateWithYup<T>(
  schema: YupSchema<T>,
  data: unknown
): SchemaValidationResult<T> {
  try {
    const validData = schema.validateSync(data, { abortEarly: false });
    return {
      isValid: true,
      errors: [],
      fieldErrors: {},
      data: validData
    };
  } catch (err) {
    const yupError = err as YupError;
    const errors: string[] = yupError.errors || [];
    const fieldErrors: Record<string, string[]> = {};

    if (yupError.inner) {
      for (const inner of yupError.inner) {
        if (inner.path) {
          if (!fieldErrors[inner.path]) {
            fieldErrors[inner.path] = [];
          }
          fieldErrors[inner.path].push(...inner.errors);
        }
      }
    }

    return {
      isValid: false,
      errors,
      fieldErrors
    };
  }
}

/**
 * Async validation with Yup schema
 */
export async function validateWithYupAsync<T>(
  schema: YupSchema<T>,
  data: unknown
): Promise<SchemaValidationResult<T>> {
  try {
    const validData = await schema.validate(data, { abortEarly: false });
    return {
      isValid: true,
      errors: [],
      fieldErrors: {},
      data: validData
    };
  } catch (err) {
    const yupError = err as YupError;
    const errors: string[] = yupError.errors || [];
    const fieldErrors: Record<string, string[]> = {};

    if (yupError.inner) {
      for (const inner of yupError.inner) {
        if (inner.path) {
          if (!fieldErrors[inner.path]) {
            fieldErrors[inner.path] = [];
          }
          fieldErrors[inner.path].push(...inner.errors);
        }
      }
    }

    return {
      isValid: false,
      errors,
      fieldErrors
    };
  }
}

/**
 * Schema validator class that works with both Zod and Yup
 */
export class SchemaValidator<T = unknown> {
  private schema: ZodSchema<T> | YupSchema<T>;
  private type: 'zod' | 'yup';

  constructor(schema: ZodSchema<T> | YupSchema<T>, type: 'zod' | 'yup') {
    this.schema = schema;
    this.type = type;
  }

  /**
   * Create validator from Zod schema
   */
  static fromZod<T>(schema: ZodSchema<T>): SchemaValidator<T> {
    return new SchemaValidator(schema, 'zod');
  }

  /**
   * Create validator from Yup schema
   */
  static fromYup<T>(schema: YupSchema<T>): SchemaValidator<T> {
    return new SchemaValidator(schema, 'yup');
  }

  /**
   * Validate data synchronously
   */
  validate(data: unknown): SchemaValidationResult<T> {
    if (this.type === 'zod') {
      return validateWithZod(this.schema as ZodSchema<T>, data);
    }
    return validateWithYup(this.schema as YupSchema<T>, data);
  }

  /**
   * Validate data asynchronously
   */
  async validateAsync(data: unknown): Promise<SchemaValidationResult<T>> {
    if (this.type === 'zod') {
      return validateWithZodAsync(this.schema as ZodSchema<T>, data);
    }
    return validateWithYupAsync(this.schema as YupSchema<T>, data);
  }

  /**
   * Get a ValidationRule for a specific field path
   */
  getFieldRule(fieldPath: string, message?: string): ValidationRule {
    return {
      validate: (value) => {
        const result = this.validate({ [fieldPath]: value });
        return !result.fieldErrors[fieldPath]?.length;
      },
      message: message || `${fieldPath} is invalid`
    };
  }

  /**
   * Convert schema to ValidationRule array for use with FormValidator
   */
  toValidationRules(): ValidationRule[] {
    return [{
      validate: (data) => this.validate(data).isValid,
      message: 'Form validation failed'
    }];
  }
}

/**
 * Helper to create form validation from Zod schema
 * Returns field errors mapped by field name
 */
export function createZodFormValidator<T extends Record<string, unknown>>(
  schema: ZodSchema<T>
) {
  return {
    validate: (values: T) => validateWithZod(schema, values),
    validateAsync: (values: T) => validateWithZodAsync(schema, values),
    validateField: (field: keyof T, value: unknown, allValues?: Partial<T>) => {
      const testData = allValues ? { ...allValues, [field]: value } : { [field]: value };
      const result = validateWithZod(schema, testData);
      return {
        isValid: !result.fieldErrors[field as string]?.length,
        errors: result.fieldErrors[field as string] || []
      };
    }
  };
}

/**
 * Helper to create form validation from Yup schema
 * Returns field errors mapped by field name
 */
export function createYupFormValidator<T extends Record<string, unknown>>(
  schema: YupSchema<T>
) {
  return {
    validate: (values: T) => validateWithYup(schema, values),
    validateAsync: (values: T) => validateWithYupAsync(schema, values),
    validateField: (field: keyof T, value: unknown, allValues?: Partial<T>) => {
      const testData = allValues ? { ...allValues, [field]: value } : { [field]: value };
      const result = validateWithYup(schema, testData);
      return {
        isValid: !result.fieldErrors[field as string]?.length,
        errors: result.fieldErrors[field as string] || []
      };
    }
  };
}

/**
 * Utility to extract error messages for a specific field from validation result
 */
export function getFieldErrors(
  result: SchemaValidationResult,
  fieldPath: string
): string[] {
  return result.fieldErrors[fieldPath] || [];
}

/**
 * Utility to check if a specific field has errors
 */
export function hasFieldError(
  result: SchemaValidationResult,
  fieldPath: string
): boolean {
  return (result.fieldErrors[fieldPath]?.length || 0) > 0;
}

/**
 * Utility to get first error message for a field
 */
export function getFirstFieldError(
  result: SchemaValidationResult,
  fieldPath: string
): string | undefined {
  return result.fieldErrors[fieldPath]?.[0];
}

/**
 * Map validation result field errors to form-compatible format
 */
export function mapToFormErrors(
  result: SchemaValidationResult
): Record<string, string> {
  const formErrors: Record<string, string> = {};

  for (const [field, errors] of Object.entries(result.fieldErrors)) {
    if (errors.length > 0) {
      formErrors[field] = errors[0];
    }
  }

  return formErrors;
}
