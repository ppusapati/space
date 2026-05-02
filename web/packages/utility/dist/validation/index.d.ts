export interface ValidationRule<T = any> {
    validate: (value: T) => boolean | Promise<boolean>;
    message: string | ((value: T) => string);
    priority?: number;
}
export interface ValidationResult {
    isValid: boolean;
    errors: string[];
    warnings?: string[];
}
export interface FormFieldValidation {
    rules: ValidationRule[];
    required?: boolean;
    validateOnBlur?: boolean;
    validateOnChange?: boolean;
}
export declare class FormValidator {
    private fields;
    addField(name: string, validation: FormFieldValidation): void;
    removeField(name: string): void;
    validateField(name: string, value: any): Promise<ValidationResult>;
    validateForm(values: Record<string, any>): Promise<Record<string, ValidationResult>>;
    private isEmpty;
}
export declare const ValidationRules: {
    required: (message?: string) => ValidationRule;
    email: (message?: string) => ValidationRule;
    minLength: (length: number, message?: string) => ValidationRule;
    maxLength: (length: number, message?: string) => ValidationRule;
    pattern: (regex: RegExp, message?: string) => ValidationRule;
    number: (message?: string) => ValidationRule;
    min: (minimum: number, message?: string) => ValidationRule;
    max: (maximum: number, message?: string) => ValidationRule;
    phone: (message?: string) => ValidationRule;
    url: (message?: string) => ValidationRule;
    password: (options?: {
        minLength?: number;
        requireUppercase?: boolean;
        requireLowercase?: boolean;
        requireNumbers?: boolean;
        requireSymbols?: boolean;
    }, message?: string) => ValidationRule;
    match: (fieldName: string, getValue: () => any, message?: string) => ValidationRule;
    fileSize: (maxSizeBytes: number, message?: string) => ValidationRule;
    fileType: (allowedTypes: string[], message?: string) => ValidationRule;
    async: (asyncValidator: (value: any) => Promise<boolean>, message?: string) => ValidationRule;
};
export declare function isEmpty(value: any): boolean;
export declare function formatBytes(bytes: number, decimals?: number): string;
export * from './schema-adapters';
//# sourceMappingURL=index.d.ts.map