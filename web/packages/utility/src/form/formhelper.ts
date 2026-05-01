import { validateMessage } from '@protovalidate/core';

export type FormField = {
  name: string;
  label: string;
  type: 'text' | 'number' | 'boolean' | 'object';
  widget: string;
  min?: number;
  max?: number;
  repeated: boolean;
  enum?: string[];
  nested?: FormField[];
  placeholder?: string;
  help?: string;
  error?: string;
};

/**
 * Map proto field to FormField with metadata
 */
function mapField(f: any): FormField {
  const baseField: FormField = {
    name: f.name,
    repeated: f.repeated,
    label: f.options?.label || f.name,
    widget: f.options?.widget || 'input',
    min: f.options?.min,
    max: f.options?.max,
    enum: f.enum ? f.enum.values : undefined,
    type: 'text',
    placeholder: f.options?.placeholder,
    help: f.options?.help,
    error: ''
  };

  switch (f.kind) {
    case 9: // string
      baseField.type = 'text';
      break;
    case 5: // int32
    case 1: // double / float
      baseField.type = 'number';
      break;
    case 8: // bool
      baseField.type = 'boolean';
      break;
    case 11: // message
      baseField.type = 'object';
      baseField.nested = createFormConfig(f.message);
      break;
  }

  return baseField;
}

/**
 * Convert messageDesc to dynamic form config
 */
export function createFormConfig(messageDesc: any): FormField[] {
  return messageDesc.fields.map(mapField);
}

/**
 * Initialize empty form data
 */
export function initFormData(formConfig: FormField[]): any {
  const data: any = {};
  formConfig.forEach(f => {
    if (f.repeated) data[f.name] = [];
    else if (f.type === 'object') data[f.name] = initFormData(f.nested!);
    else if (f.type === 'boolean') data[f.name] = false;
    else data[f.name] = '';
  });
  return data;
}

/**
 * Add a repeated field item
 */
export function addRepeatedField(formData: any, field: FormField) {
  if (!field.repeated) return;
  if (field.type === 'object') {
    formData[field.name].push(initFormData(field.nested!));
  } else {
    formData[field.name].push('');
  }
}

/**
 * Remove a repeated field item
 */
export function removeRepeatedField(formData: any, fieldName: string, index: number) {
  formData[fieldName].splice(index, 1);
}

/**
 * Validate form data using protovalidate
 */
export function validateForm(messageDesc: any, formFields: FormField[], data: any) {
  const result = validateMessage(messageDesc, data);

  // reset errors
  formFields.forEach(f => f.error = '');

  if (!result.valid) {
    result.errors.forEach(err => {
      const field = formFields.find(f => f.name === err.field);
      if (field) field.error = err.message;
    });
  }

  return result.valid;
}
