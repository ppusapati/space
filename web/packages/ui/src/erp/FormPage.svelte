<script lang="ts">
  import { get } from 'svelte/store';
  import type { FormSchema } from '@samavāya/core';
  import { authStore } from '@samavāya/stores';
  import CrudFormPage from './CrudFormPage.svelte';
  import { adaptFormDefinition, extractFormMeta, type ProtoFormDef } from '../forms/protoFormAdapter.js';

  // ============================================================================
  // PROPS
  // ============================================================================

  interface Props {
    /** The form ID to load and render */
    formId: string;
    /** Mode: create, edit, or view */
    mode?: 'create' | 'edit' | 'view';
    /** Optional pre-filled values (e.g. for edit mode) */
    initialValues?: Record<string, unknown>;
    /** Back-navigation URL */
    cancelHref?: string;
    /** Callback after successful submission */
    onSuccess?: (entityId: string) => void;
    /** Callback on submission error */
    onError?: (error: string) => void;
  }

  let {
    formId,
    mode = 'create',
    initialValues = {},
    cancelHref = '',
    onSuccess,
    onError,
  }: Props = $props();

  // ============================================================================
  // STATE
  // ============================================================================

  let schema = $state<FormSchema<Record<string, unknown>> | null>(null);
  let title = $state('');
  let subtitle = $state('');
  let rpcEndpoint = $state('');
  let isLoading = $state(true);
  let isSubmitting = $state(false);
  let error = $state<string | null>(null);
  let formErrors = $state<Record<string, string>>({});
  let values = $state<Record<string, unknown>>({});

  // ============================================================================
  // FORM SERVICE RPC (inline to avoid api↔ui circular dep; authStore already
  // a workspace dep so it's free)
  // ============================================================================
  //
  // The backend FormService handler reads tenant context from the JWT-injected
  // UserContext (server-side resolveTenantContext). So clients only need to
  // send a Bearer token; msg.Context can be empty — server is the source of
  // truth for tenant identity. Empty body context is intentional.

  const FORM_SERVICE_BASE = '/platform.formservice.api.v1.FormService';

  function getBaseUrl(): string {
    if (typeof import.meta !== 'undefined' && import.meta.env?.VITE_API_URL) {
      return import.meta.env.VITE_API_URL;
    }
    return 'http://localhost:9090';
  }

  function getAuthHeader(): Record<string, string> {
    if (typeof window === 'undefined') return {};
    const state = get(authStore);
    const token = state.tokens?.accessToken;
    return token ? { Authorization: `Bearer ${token}` } : {};
  }

  async function rpc<TReq, TRes>(method: string, request: TReq): Promise<TRes> {
    // 'omit' for Bearer-token auth — see ListPage.svelte rpc() comment
    // and DEPLOYMENT_READINESS.md item 45 round 3 for the CORS
    // wildcard-origin trap rationale.
    const url = `${getBaseUrl()}${FORM_SERVICE_BASE}/${method}`;
    const response = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...getAuthHeader() },
      credentials: 'omit',
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      const body = await response.text().catch(() => '');
      throw new Error(`FormService.${method} failed: ${response.status} ${response.statusText} ${body}`);
    }
    return response.json() as Promise<TRes>;
  }

  // ============================================================================
  // LOAD FORM SCHEMA
  // ============================================================================

  async function loadSchema(id: string): Promise<void> {
    isLoading = true;
    error = null;
    schema = null;

    try {
      const response = await rpc<
        { context: Record<string, unknown>; formId: string },
        { formDefinition?: ProtoFormDef; overrideCount?: number }
      >('GetFormSchema', { context: {}, formId: id });

      if (!response.formDefinition) {
        throw new Error(`No form definition returned for formId: ${id}`);
      }

      const def = response.formDefinition;
      const meta = extractFormMeta(def);
      const converted = adaptFormDefinition(def);

      schema = converted;
      title = meta.title;
      subtitle = meta.subtitle;
      rpcEndpoint = meta.rpcEndpoint;

      // Apply initial values
      values = { ...initialValues };
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load form schema';
      error = message;
      console.error('[FormPage] Schema load error:', err);
    } finally {
      isLoading = false;
    }
  }

  // ============================================================================
  // SUBMIT FORM
  // ============================================================================

  async function handleSubmit(formValues: Record<string, unknown>): Promise<void> {
    isSubmitting = true;
    error = null;
    formErrors = {};

    try {
      const response = await rpc<
        { context: Record<string, unknown>; formId: string; values: Record<string, unknown> },
        {
          entityId?: string;
          validationErrors?: Array<{ fieldId: string; message: string }>;
          responseStatus?: string;
          durationMs?: number;
        }
      >('SubmitForm', { context: {}, formId, values: formValues });

      // Handle validation errors from the server
      if (response.validationErrors && response.validationErrors.length > 0) {
        const errors: Record<string, string> = {};
        for (const ve of response.validationErrors) {
          errors[ve.fieldId] = ve.message;
        }
        formErrors = errors;
        return;
      }

      // Success
      if (onSuccess && response.entityId) {
        onSuccess(response.entityId);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Form submission failed';
      error = message;
      if (onError) {
        onError(message);
      }
      console.error('[FormPage] Submission error:', err);
    } finally {
      isSubmitting = false;
    }
  }

  // ============================================================================
  // REACTIVE LOAD
  // ============================================================================

  $effect(() => {
    if (formId) {
      loadSchema(formId);
    }
  });
</script>

{#if isLoading}
  <div class="form-page-loading">
    <div class="form-page-loading-spinner"></div>
    <p>Loading form...</p>
  </div>
{:else if error && !schema}
  <div class="form-page-error" role="alert">
    <h2>Failed to Load Form</h2>
    <p>{error}</p>
    <button type="button" class="form-page-retry-btn" onclick={() => loadSchema(formId)}>
      Retry
    </button>
  </div>
{:else if schema}
  <CrudFormPage
    {title}
    subtitle={subtitle}
    {mode}
    {schema}
    values={values}
    errors={formErrors}
    {isLoading}
    {isSubmitting}
    {error}
    {cancelHref}
    onSubmit={handleSubmit}
  />
{/if}

<style>
  .form-page-loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem 1.5rem;
    color: var(--color-text-secondary, #6b7280);
  }

  .form-page-loading-spinner {
    width: 2rem;
    height: 2rem;
    border: 3px solid var(--color-border, #e5e7eb);
    border-top-color: var(--color-primary, #4f46e5);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
    margin-bottom: 1rem;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .form-page-error {
    padding: 2rem;
    background: var(--color-destructive-muted, #fef2f2);
    border: 1px solid var(--color-destructive, #ef4444);
    border-radius: var(--radius-lg, 0.5rem);
    text-align: center;
  }

  .form-page-error h2 {
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--color-destructive, #ef4444);
    margin: 0 0 0.5rem 0;
  }

  .form-page-error p {
    font-size: 0.875rem;
    color: var(--color-text-secondary, #6b7280);
    margin: 0 0 1rem 0;
  }

  .form-page-retry-btn {
    padding: 0.5rem 1.25rem;
    background: var(--color-primary, #4f46e5);
    color: white;
    border: none;
    border-radius: var(--radius-md, 0.375rem);
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.15s;
  }

  .form-page-retry-btn:hover {
    background: var(--color-primary-hover, #4338ca);
  }
</style>
