/**
 * FormService & FormBuilder TypeScript types
 * Generated from formservice.proto and formbuilder.proto
 *
 * These hand-written types mirror the proto definitions and can be replaced
 * by codegen output (protoc-gen-es) once the formservice proto is added to
 * the buf pipeline.
 */

// ============================================================================
// FormBuilder types (from formbuilder.proto)
// ============================================================================

/** Proto FieldType enum — mirrors workflow.formbuilder.api.v1.FieldType */
export enum ProtoFieldType {
  TEXT = 0,
  NUMBER = 1,
  EMAIL = 2,
  DROPDOWN = 3,
  RADIO = 4,
  CHECKBOX = 5,
  DATE = 6,
  DATETIME = 7,
  FILE = 8,
  TEXTAREA = 9,
  MULTI_SELECT = 10,
  CURRENCY = 11,
  PHONE = 12,
  URL = 13,
  JSON = 14,
  ARRAY = 15,
  NESTED_FORM = 16,
  PASSWORD = 17,
  RICHTEXT = 18,
  PERCENTAGE = 19,
  TIME = 20,
  DATERANGE = 21,
  MONTHPICKER = 22,
  CHECKBOXGROUP = 23,
  SWITCH = 24,
  LOOKUP = 25,
  MULTILOOKUP = 26,
  TREE = 27,
  CASCADE = 28,
  TABLE = 29,
  OBJECT = 30,
  KEYVALUE = 31,
  IMAGE = 32,
  FORMULA = 33,
  BARCODE = 34,
  COLOR = 35,
  RATING = 36,
  SLIDER = 37,
  CRON = 38,
}

/** Proto ConditionOperator enum */
export enum ProtoConditionOperator {
  EQUALS = 0,
  NOT_EQUALS = 1,
  GREATER_THAN = 2,
  LESS_THAN = 3,
  GREATER_THAN_OR_EQUAL = 4,
  LESS_THAN_OR_EQUAL = 5,
  CONTAINS = 6,
  NOT_CONTAINS = 7,
  IN = 8,
  NOT_IN = 9,
  REGEX_MATCH = 10,
}

/** Proto ActionType enum */
export enum ProtoActionType {
  SHOW = 0,
  HIDE = 1,
  ENABLE = 2,
  DISABLE = 3,
  REQUIRE = 4,
  SET_VALUE = 5,
  CLEAR = 6,
  VALIDATE = 7,
}

export interface ProtoValidation {
  min?: number;
  max?: number;
  minLength?: number;
  maxLength?: number;
  pattern?: string;
  allowedValues?: string[];
  customValidator?: string;
  rules?: ProtoValidationRule[];
  unique?: boolean;
  encrypted?: boolean;
  businessRule?: string;
  readonlyWhen?: string;
  hiddenWhen?: string;
  required?: boolean;
  readonly?: boolean;
  hidden?: boolean;
  minItems?: number;
  maxItems?: number;
}

export interface ProtoValidationRule {
  type: string;
  message: string;
  params?: Record<string, string>;
}

export interface ProtoFieldDependency {
  field: string;
  param: string;
  apiEndpoint: string;
}

export interface ProtoApiConfig {
  method?: string;
  headers?: Record<string, string>;
  authType?: string;
  timeoutSeconds?: number;
  retryCount?: number;
  dependentFields?: string[];
  requestTransform?: string;
  responseTransform?: string;
}

export interface ProtoCacheConfig {
  enabled: boolean;
  ttlSeconds?: number;
  realTime?: boolean;
  cacheKeyPattern?: string;
}

export interface ProtoTableConfig {
  columns?: ProtoFormField[];
  allowAddRows?: boolean;
  allowDeleteRows?: boolean;
  allowReorder?: boolean;
  minRows?: number;
  maxRows?: number;
  allowAdd?: boolean;
  allowDelete?: boolean;
  allowEdit?: boolean;
  paginated?: boolean;
  pageSize?: number;
  selectable?: boolean;
  sortable?: boolean;
}

export interface ProtoFormulaConfig {
  expression: string;
  dependentFields?: string[];
  resultType?: string;
  autoCalculate?: boolean;
  decimalPlaces?: number;
}

export interface ProtoLookupConfig {
  entityType: string;
  searchEndpoint: string;
  displayTemplate?: string;
  allowMultiple?: boolean;
  searchable?: boolean;
  minSearchLength?: number;
  clearable?: boolean;
  creatable?: boolean;
  multiSelect?: boolean;
  minSearchChars?: number;
  searchDelay?: number;
  pageSize?: number;
  cacheResults?: boolean;
  cacheTtlSeconds?: number;
  columns?: ProtoLookupColumn[];
  filters?: ProtoFilterCondition[];
  dependentFields?: string[];
  allowCustomValue?: boolean;
  caseSensitive?: boolean;
}

export interface ProtoLookupColumn {
  columnId: string;
  displayName: string;
  width?: string;
  sortable?: boolean;
}

export interface ProtoFilterCondition {
  field: string;
  operator: string;
  value: string;
}

export interface ProtoCascadeConfig {
  parentField?: string;
  levels?: ProtoCascadeLevel[];
  clearable?: boolean;
  cacheResults?: boolean;
  cacheTtlSeconds?: number;
  searchDelay?: number;
}

export interface ProtoCascadeLevel {
  fieldId: string;
  parentField?: string;
  endpoint?: string;
  level?: number;
  name?: string;
  entityType?: string;
  searchEndpoint?: string;
  displayTemplate?: string;
  parentFieldId?: string;
  parentEntityField?: string;
}

export interface ProtoFileConfig {
  maxFileSize?: number;
  allowedMimeTypes?: string[];
  allowMultiple?: boolean;
}

export interface ProtoFieldRoles {
  read?: string[];
  write?: string[];
  hide?: string[];
}

export interface ProtoCondition {
  field: string;
  operator: ProtoConditionOperator;
  value?: unknown;
}

export interface ProtoDependencyAction {
  field: string;
  action: ProtoActionType;
  value?: unknown;
}

export interface ProtoConditionalRule {
  name: string;
  conditions: ProtoCondition[];
  actions: ProtoDependencyAction[];
}

export interface ProtoDependency {
  if: ProtoCondition[];
  then: ProtoDependencyAction;
}

export interface ProtoCrossFieldValidation {
  fields: string[];
  rule: string;
  message: string;
  validatorFunction?: string;
}

/** Proto FormField — mirrors workflow.formbuilder.api.v1.FormField */
export interface ProtoFormField {
  id: string;
  type: ProtoFieldType;
  label: string;
  hint?: string;
  required?: boolean;
  roles?: ProtoFieldRoles;
  validation?: ProtoValidation;
  optionsSource?: string;
  dependsOn?: ProtoFieldDependency;
  defaultValue?: unknown;
  attributes?: Record<string, string>;
  conditionalRules?: ProtoConditionalRule[];
  apiConfig?: ProtoApiConfig;
  isCoreField?: boolean;
  cacheConfig?: ProtoCacheConfig;
  tableConfig?: ProtoTableConfig;
  formulaConfig?: ProtoFormulaConfig;
  lookupConfig?: ProtoLookupConfig;
  cascadeConfig?: ProtoCascadeConfig;
  maxFileSize?: number;
  allowedMimeTypes?: string[];
  multipleFiles?: boolean;
  fieldId?: string;
  displayName?: string;
  fieldType?: ProtoFieldType;
  orderIndex?: number;
  order?: number;
  readonly?: boolean;
  placeholder?: string;
  hidden?: boolean;
  sortable?: boolean;
  filterable?: boolean;
  conditionallyRequired?: string;
  visibleWhen?: string;
  readonlyWhen?: string;
  hiddenWhen?: string;
  fileConfig?: ProtoFileConfig;
}

/** Proto FormStep — mirrors workflow.formbuilder.api.v1.FormStep */
export interface ProtoFormStep {
  id: string;
  label: string;
  fields: ProtoFormField[];
  order: number;
  metadata?: Record<string, string>;
  stepId?: string;
  description?: string;
  title?: string;
  fieldIds?: string[];
}

/** Proto FormMetadata — mirrors workflow.formbuilder.api.v1.FormMetadata */
export interface ProtoFormMetadata {
  formId: string;
  title: string;
  description?: string;
  version?: string;
  createdBy?: string;
  allowedRoles?: string[];
  audit?: boolean;
  tableName?: string;
  module?: string;
  schemaVersion?: number;
  coreFields?: string[];
  friendlyEndpoint: string;
  rpcEndpoint: string;
  createdAt?: string;
  updatedAt?: string;
  service?: string;
  entityType?: string;
}

/** Proto FormDefinition — mirrors workflow.formbuilder.api.v1.FormDefinition */
export interface ProtoFormDefinition {
  metadata: ProtoFormMetadata;
  steps: ProtoFormStep[];
  dependencies?: ProtoDependency[];
  crossFieldValidations?: ProtoCrossFieldValidation[];
  allFields?: ProtoFormField[];
  version?: string;
}

// ============================================================================
// FormService types (from formservice.proto)
// ============================================================================

/** ModuleSummary — mirrors platform.formservice.api.v1.ModuleSummary */
export interface ModuleSummary {
  moduleId: string;
  label: string;
  formCount: number;
}

/** FormSummary — mirrors platform.formservice.api.v1.FormSummary */
export interface FormSummary {
  formId: string;
  title: string;
  description: string;
  friendlyEndpoint: string;
  rpcEndpoint: string;
  moduleId: string;
  version: string;
}

/** ListModulesResponse */
export interface ListModulesResponse {
  modules: ModuleSummary[];
  totalModules: number;
}

/** ListFormsResponse */
export interface ListFormsResponse {
  forms: FormSummary[];
  totalForms: number;
}

/** GetFormSchemaResponse */
export interface GetFormSchemaResponse {
  formDefinition: ProtoFormDefinition;
  overrideCount: number;
}

/** ValidationError from submission */
export interface SubmitValidationError {
  fieldId: string;
  message: string;
}

/** SubmitFormResponse */
export interface SubmitFormResponse {
  entityId: string;
  validationErrors: SubmitValidationError[];
  responseStatus: string;
  durationMs: number;
}

// ============================================================================
// FormRegistry types (from formbuilder.proto)
// ============================================================================

/** FormRegistryEntry */
export interface FormRegistryEntry {
  formId: string;
  module: string;
  service: string;
  title: string;
  description: string;
  friendlyEndpoint: string;
  rpcEndpoint: string;
  allowedRoles: string[];
  version: string;
}

/** FormRegistryModule */
export interface FormRegistryModule {
  moduleId: string;
  label: string;
  forms: FormRegistryEntry[];
}

/** FormRegistry */
export interface FormRegistry {
  modules: FormRegistryModule[];
  totalForms: number;
  generatedAt: string;
  generatorVersion: string;
}
