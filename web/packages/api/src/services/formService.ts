/**
 * FormService API Client
 * Provides typed access to the FormService RPCs for module discovery,
 * form schema retrieval, and form submission.
 *
 * Uses the generated ConnectRPC client from @samavāya/proto. The service
 * descriptor lives at core/platform/formservice/proto/formservice_pb.ts
 * and is regenerated via packages/proto/generate-formservice.sh.
 */

import type { Client } from '@connectrpc/connect';
import { getApiClient } from '../client/client.js';
import {
  FormService,
  type ListModulesResponse as PbListModulesResponse,
  type ListFormsResponse as PbListFormsResponse,
  type GetFormSchemaResponse as PbGetFormSchemaResponse,
  type SubmitFormResponse as PbSubmitFormResponse,
  type ModuleSummary as PbModuleSummary,
  type FormSummary as PbFormSummary,
} from '@samavāya/proto/gen/core/platform/formservice/proto/formservice_pb.js';
import type {
  ModuleSummary,
  FormSummary,
  ProtoFormDefinition,
  SubmitFormResponse,
  ListModulesResponse,
  ListFormsResponse,
  GetFormSchemaResponse,
} from '../types/formservice.types.js';

export type {
  ModuleSummary,
  FormSummary,
  ProtoFormDefinition,
  SubmitFormResponse,
  ListModulesResponse,
  ListFormsResponse,
  GetFormSchemaResponse,
};
export type { SubmitValidationError } from '../types/formservice.types.js';

export { FormService };

// ============================================================================
// ERROR TYPE
// ============================================================================

export class FormServiceError extends Error {
  constructor(
    message: string,
    public readonly statusCode: number,
    public readonly body: string
  ) {
    super(message);
    this.name = 'FormServiceError';
  }
}

// ============================================================================
// TYPED CLIENT
// ============================================================================

/** Typed ConnectRPC client for FormService. */
export function getFormServiceClient(): Client<typeof FormService> {
  return getApiClient().getService(FormService);
}

// ============================================================================
// PUBLIC API
// ============================================================================

/**
 * List all modules that have registered forms.
 * Returns permission-filtered module summaries with form counts.
 */
export async function listModules(): Promise<ModuleSummary[]> {
  const client = getFormServiceClient();
  const response = (await client.listModules({})) as PbListModulesResponse;
  return response.modules.map(toModuleSummary);
}

/**
 * List all forms for a given module.
 * Returns form summaries (id, title, endpoints) for sidebar navigation.
 */
export async function listForms(moduleId: string): Promise<FormSummary[]> {
  if (!moduleId) {
    throw new FormServiceError('moduleId is required', 400, '');
  }
  const client = getFormServiceClient();
  const response = (await client.listForms({ moduleId })) as PbListFormsResponse;
  return response.forms.map(toFormSummary);
}

/**
 * Get the full form schema (FormDefinition) for a specific form.
 * Includes tenant overrides merged into the base definition.
 */
export async function getFormSchema(formId: string): Promise<ProtoFormDefinition> {
  if (!formId) {
    throw new FormServiceError('formId is required', 400, '');
  }
  const client = getFormServiceClient();
  const response = (await client.getFormSchema({ formId })) as PbGetFormSchemaResponse;
  if (!response.formDefinition) {
    throw new FormServiceError(`No form definition returned for formId: ${formId}`, 404, '');
  }
  return response.formDefinition as unknown as ProtoFormDefinition;
}

/**
 * Submit form values through the FormService proxy.
 * The server validates, routes to the target RPC, and creates an audit trail.
 */
export async function submitForm(
  formId: string,
  values: Record<string, unknown>
): Promise<SubmitFormResponse> {
  if (!formId) {
    throw new FormServiceError('formId is required', 400, '');
  }
  const client = getFormServiceClient();
  const response = (await client.submitForm({
    formId,
    values: { fields: valuesToStructFields(values) },
  } as never)) as PbSubmitFormResponse;
  return {
    entityId: response.entityId,
    validationErrors: response.validationErrors.map((e) => ({
      fieldId: e.fieldId,
      message: e.message,
    })),
    responseStatus: response.responseStatus,
    durationMs: Number(response.durationMs),
  };
}

// ============================================================================
// INTERNAL MAPPERS
// ============================================================================

function toModuleSummary(m: PbModuleSummary): ModuleSummary {
  return {
    moduleId: m.moduleId,
    label: m.label,
    formCount: m.formCount,
  };
}

function toFormSummary(f: PbFormSummary): FormSummary {
  return {
    formId: f.formId,
    title: f.title,
    description: f.description,
    friendlyEndpoint: f.friendlyEndpoint,
    rpcEndpoint: f.rpcEndpoint,
    moduleId: f.moduleId,
    version: f.version,
  };
}

/**
 * google.protobuf.Struct stores values as `fields: Record<string, Value>`.
 * We map a plain JS object into the Value wire shape. Nested objects/arrays
 * collapse into JSON strings for now — FormService validates against the
 * proto shape, not deep-typed struct values, so this is sufficient.
 */
function valuesToStructFields(values: Record<string, unknown>) {
  const fields: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(values)) {
    fields[key] = toStructValue(value);
  }
  return fields;
}

function toStructValue(v: unknown): Record<string, unknown> {
  if (v === null || v === undefined) {
    return { kind: { case: 'nullValue', value: 0 } };
  }
  if (typeof v === 'boolean') {
    return { kind: { case: 'boolValue', value: v } };
  }
  if (typeof v === 'number') {
    return { kind: { case: 'numberValue', value: v } };
  }
  if (typeof v === 'string') {
    return { kind: { case: 'stringValue', value: v } };
  }
  if (Array.isArray(v)) {
    return {
      kind: {
        case: 'listValue',
        value: { values: v.map(toStructValue) },
      },
    };
  }
  if (typeof v === 'object') {
    const inner: Record<string, unknown> = {};
    for (const [k, val] of Object.entries(v as Record<string, unknown>)) {
      inner[k] = toStructValue(val);
    }
    return { kind: { case: 'structValue', value: { fields: inner } } };
  }
  return { kind: { case: 'stringValue', value: String(v) } };
}
