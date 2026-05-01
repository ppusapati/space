<script lang="ts">
  /**
   * ListPage — page-level wrapper that loads a form schema (for column
   * derivation) and a page of rows from a list RPC. Mirrors FormPage's
   * role for the create/edit side.
   *
   * Why split from CrudListPage:
   *
   *   CrudListPage is a "dumb" presentation component (table + paginator
   *   + empty state). ListPage adds the IO: schema fetch via FormService
   *   + list RPC call via direct fetch with auth header. Same separation
   *   of concerns as CrudFormPage / FormPage.
   *
   * Usage (in a route):
   *
   *   <ListPage
   *     formId="form_masters_item_master"
   *     listEndpoint="/masters.item.api.v1.ItemService/ListItems"
   *     responseRowsKey="items"
   *     responseTotalKey="totalCount"
   *     createHref="/forms/masters/form_masters_item_master"
   *     onOpen={(id) => goto(`/forms/masters/form_masters_item_master?id=${id}`)}
   *   />
   *
   * The route is responsible for picking the listEndpoint + the keys
   * the list response uses for the rows array and total count. We do
   * NOT auto-derive these from rpcEndpoint or `service` — endpoint
   * naming is service-specific (ListItems vs ListCategories vs
   * GetEmployees) and silent derivation hid bugs in earlier sprints.
   */
  import { get } from 'svelte/store';
  import type { FormSchema } from '@samavāya/core';
  import { authStore } from '@samavāya/stores';
  import { adaptFormDefinition, extractFormMeta, type ProtoFormDef } from '../forms/protoFormAdapter.js';
  import CrudListPage from './CrudListPage.svelte';

  interface Props {
    /** The form ID to load — schema drives column metadata + display title */
    formId: string;
    /**
     * The full list-RPC endpoint, e.g.
     *   "/masters.item.api.v1.ItemService/ListItems"
     * Set explicitly per-route; no derivation from rpcEndpoint/service.
     */
    listEndpoint: string;
    /**
     * Key in the list response that holds the rows array, e.g. "items"
     * or "categories". Defaults to "items".
     */
    responseRowsKey?: string;
    /**
     * Key in the list response that holds the total row count.
     * Supports dotted paths for nested fields, e.g. "pagination.totalCount"
     * (used by services that embed totals in a Pagination message) or
     * "totalCount" (used by services that surface it at the top level).
     * Defaults to "totalCount". Some services don't return a count;
     * in that case the loader uses rows.length.
     */
    responseTotalKey?: string;
    /**
     * Page size (default 25).
     */
    pageSize?: number;
    /**
     * Optional column-name override; defaults to schema's coreFields.
     */
    columns?: string[];
    /**
     * URL for the "+ New" button.
     */
    createHref?: string;
    /**
     * Callback when a row is clicked. Receives the row's id (id column).
     */
    onOpen?: (id: string) => void;
    /**
     * Identifier column name (default 'id').
     */
    idColumn?: string;
  }

  let {
    formId,
    listEndpoint,
    responseRowsKey = 'items',
    responseTotalKey = 'totalCount',
    pageSize = 25,
    columns,
    createHref,
    onOpen,
    idColumn = 'id',
  }: Props = $props();

  let schema = $state<FormSchema<Record<string, unknown>> | null>(null);
  let title = $state('');
  let subtitle = $state('');
  let derivedColumns = $state<string[] | undefined>(undefined);
  let schemaError = $state<string | null>(null);

  // ============================================================================
  // RPC plumbing — mirrors FormPage's pattern (uses authStore JWT, default URL)
  // ============================================================================

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

  async function rpc<TReq, TRes>(url: string, request: TReq): Promise<TRes> {
    // 'omit' for Bearer-token auth — backend CORS uses Allow-Origin:'*'
    // and the browser silently rejects credentialed wildcard responses
    // as "Failed to fetch". See packages/api/src/client/transport.ts for
    // the full rationale (DEPLOYMENT_READINESS.md item 45 round 3).
    const response = await fetch(`${getBaseUrl()}${url}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...getAuthHeader() },
      credentials: 'omit',
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      const body = await response.text().catch(() => '');
      throw new Error(`${url} failed: ${response.status} ${response.statusText} ${body}`);
    }
    return response.json() as Promise<TRes>;
  }

  // ============================================================================
  // Schema load
  // ============================================================================

  async function loadSchema(id: string): Promise<void> {
    schemaError = null;
    schema = null;

    try {
      const response = await rpc<
        { formId: string },
        { formDefinition?: ProtoFormDef; overrideCount?: number }
      >('/platform.formservice.api.v1.FormService/GetFormSchema', { formId: id });

      if (!response.formDefinition) {
        throw new Error(`No form definition returned for formId: ${id}`);
      }

      const def = response.formDefinition;
      const meta = extractFormMeta(def);
      schema = adaptFormDefinition(def);
      title = meta.title;
      subtitle = meta.subtitle;
      derivedColumns = columns && columns.length > 0 ? columns : meta.coreFields;
    } catch (err) {
      schemaError = err instanceof Error ? err.message : 'Failed to load schema';
      console.error('[ListPage] schema load:', err);
    }
  }

  // ============================================================================
  // List loader
  // ============================================================================

  // Loader closure for CrudListPage. Each call reads the latest pageSize/offset
  // and posts to the configured listEndpoint with the standard {pagination:{}}
  // shape. Returns rows + totalCount extracted by the configured keys.
  async function loader(params: { pageSize: number; pageOffset: number }): Promise<{
    rows: Record<string, unknown>[];
    totalCount: number;
  }> {
    const response = await rpc<
      { pagination: { pageSize: number; pageOffset: number } },
      Record<string, unknown>
    >(listEndpoint, {
      pagination: { pageSize: params.pageSize, pageOffset: params.pageOffset },
    });
    const rows = (lookupPath(response, responseRowsKey) as Record<string, unknown>[] | undefined) ?? [];
    const totalRaw = lookupPath(response, responseTotalKey);
    const totalCount = typeof totalRaw === 'number' ? totalRaw : rows.length;
    return { rows, totalCount };
  }

  /**
   * Resolve a dotted path against a JSON-decoded RPC response. Used for
   * the responseRowsKey / responseTotalKey props because some services
   * embed the total in `pagination.totalCount` (HR EmployeeService,
   * Finance JournalEntryService) while others surface `totalCount` at
   * the top level (Masters ItemService). Returns undefined for any
   * missing segment instead of throwing — the caller falls back to
   * rows.length when the count is unavailable.
   */
  function lookupPath(obj: Record<string, unknown>, path: string): unknown {
    if (!path) return undefined;
    let cursor: unknown = obj;
    for (const segment of path.split('.')) {
      if (cursor === null || cursor === undefined || typeof cursor !== 'object') {
        return undefined;
      }
      cursor = (cursor as Record<string, unknown>)[segment];
    }
    return cursor;
  }

  // Reactive load on formId change.
  $effect(() => {
    if (formId) {
      void loadSchema(formId);
    }
  });
</script>

{#if schemaError}
  <div class="list-page-error" role="alert">
    <h2>Failed to load list</h2>
    <p>{schemaError}</p>
    <button type="button" onclick={() => loadSchema(formId)}>Retry</button>
  </div>
{:else if !schema}
  <div class="list-page-loading" role="status">Loading…</div>
{:else}
  <CrudListPage
    title={title || 'List'}
    {subtitle}
    {schema}
    columns={derivedColumns}
    {loader}
    {pageSize}
    {createHref}
    {onOpen}
    {idColumn}
  />
{/if}

<style>
  .list-page-loading,
  .list-page-error {
    padding: 2rem;
    text-align: center;
    color: var(--color-text-secondary, #6b7280);
  }

  .list-page-error {
    color: var(--color-destructive, #ef4444);
  }

  .list-page-error h2 {
    margin: 0 0 0.5rem;
    font-size: 1.125rem;
  }

  .list-page-error button {
    margin-top: 1rem;
    padding: 0.5rem 1rem;
    background: var(--color-primary, #4f46e5);
    color: white;
    border: none;
    border-radius: var(--radius-sm, 0.25rem);
    cursor: pointer;
  }
</style>
