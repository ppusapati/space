export interface ValidationRule<T = any> {
  validate: (value: T) => boolean | Promise<boolean>
  message: string | ((value: T) => string)
  priority?: number
}

export interface ValidationResult {
  isValid: boolean
  errors: string[]
  warnings?: string[]
}

export interface FormFieldValidation {
  rules: ValidationRule[]
  required?: boolean
  validateOnBlur?: boolean
  validateOnChange?: boolean
}

export class FormValidator {
  private fields: Map<string, FormFieldValidation> = new Map()

  addField(name: string, validation: FormFieldValidation): void {
    this.fields.set(name, validation)
  }

  removeField(name: string): void {
    this.fields.delete(name)
  }

  async validateField(name: string, value: any): Promise<ValidationResult> {
    const fieldValidation = this.fields.get(name)
    
    if (!fieldValidation) {
      return { isValid: true, errors: [] }
    }

    const errors: string[] = []

    // Check required validation
    if (fieldValidation.required && this.isEmpty(value)) {
      errors.push('This field is required')
    }

    // If value is empty and not required, skip other validations
    if (this.isEmpty(value) && !fieldValidation.required) {
      return { isValid: true, errors: [] }
    }

    // Run validation rules
    for (const rule of fieldValidation.rules) {
      try {
        const isValid = await rule.validate(value)
        if (!isValid) {
          const message = typeof rule.message === 'function' 
            ? rule.message(value) 
            : rule.message
          errors.push(message)
        }
      } catch (error) {
        errors.push('Validation error occurred')
      }
    }

    return {
      isValid: errors.length === 0,
      errors
    }
  }

  async validateForm(values: Record<string, any>): Promise<Record<string, ValidationResult>> {
    const results: Record<string, ValidationResult> = {}
    
    for (const [name, fieldValidation] of this.fields) {
      const value = values[name]
      results[name] = await this.validateField(name, value)
    }

    return results
  }

  private isEmpty(value: any): boolean {
    return value === null || 
           value === undefined || 
           value === '' || 
           (Array.isArray(value) && value.length === 0)
  }
}

// Common validation rules
export const ValidationRules = {
  required: (message = 'This field is required'): ValidationRule => ({
    validate: (value) => !isEmpty(value),
    message
  }),

  email: (message = 'Please enter a valid email address'): ValidationRule => ({
    validate: (value: string) => {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
      return emailRegex.test(value)
    },
    message
  }),

  minLength: (length: number, message?: string): ValidationRule => ({
    validate: (value: string) => value.length >= length,
    message: message || `Must be at least ${length} characters long`
  }),

  maxLength: (length: number, message?: string): ValidationRule => ({
    validate: (value: string) => value.length <= length,
    message: message || `Must be no more than ${length} characters long`
  }),

  pattern: (regex: RegExp, message = 'Invalid format'): ValidationRule => ({
    validate: (value: string) => regex.test(value),
    message
  }),

  number: (message = 'Must be a valid number'): ValidationRule => ({
    validate: (value: string | number) => !isNaN(Number(value)),
    message
  }),

  min: (minimum: number, message?: string): ValidationRule => ({
    validate: (value: string | number) => Number(value) >= minimum,
    message: message || `Must be at least ${minimum}`
  }),

  max: (maximum: number, message?: string): ValidationRule => ({
    validate: (value: string | number) => Number(value) <= maximum,
    message: message || `Must be no more than ${maximum}`
  }),

  phone: (message = 'Please enter a valid phone number'): ValidationRule => ({
    validate: (value: string) => {
      const phoneRegex = /^[\+]?[1-9][\d]{0,15}$/
      return phoneRegex.test(value.replace(/[\s\-\(\)]/g, ''))
    },
    message
  }),

  url: (message = 'Please enter a valid URL'): ValidationRule => ({
    validate: (value: string) => {
      try {
        new URL(value)
        return true
      } catch {
        return false
      }
    },
    message
  }),

  password: (options: {
    minLength?: number
    requireUppercase?: boolean
    requireLowercase?: boolean
    requireNumbers?: boolean
    requireSymbols?: boolean
  } = {}, message?: string): ValidationRule => ({
    validate: (value: string) => {
      const {
        minLength = 8,
        requireUppercase = true,
        requireLowercase = true,
        requireNumbers = true,
        requireSymbols = false
      } = options

      if (value.length < minLength) return false
      if (requireUppercase && !/[A-Z]/.test(value)) return false
      if (requireLowercase && !/[a-z]/.test(value)) return false
      if (requireNumbers && !/\d/.test(value)) return false
      if (requireSymbols && !/[^A-Za-z0-9]/.test(value)) return false
      
      return true
    },
    message: message || 'Password does not meet requirements'
  }),

  match: (fieldName: string, getValue: () => any, message?: string): ValidationRule => ({
    validate: (value) => value === getValue(),
    message: message || `Must match ${fieldName}`
  }),

  fileSize: (maxSizeBytes: number, message?: string): ValidationRule => ({
    validate: (file: File) => file.size <= maxSizeBytes,
    message: message || `File size must be less than ${formatBytes(maxSizeBytes)}`
  }),

  fileType: (allowedTypes: string[], message?: string): ValidationRule => ({
    validate: (file: File) => allowedTypes.includes(file.type),
    message: message || `File type must be one of: ${allowedTypes.join(', ')}`
  }),

  async: (asyncValidator: (value: any) => Promise<boolean>, message = 'Validation failed'): ValidationRule => ({
    validate: asyncValidator,
    message
  })
}

// Utility functions
export function isEmpty(value: any): boolean {
  return value === null || 
         value === undefined || 
         value === '' || 
         (Array.isArray(value) && value.length === 0)
}

export function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return '0 Bytes'

  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']

  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i]
}

// Re-export schema adapters for Zod/Yup integration
export * from './schema-adapters'