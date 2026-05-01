<script lang="ts">
  /**
   * CrudListPage — page-level wrapper for an entity list view.
   *
   * Mirrors CrudFormPage's role for create/edit: takes a form schema +
   * the list-RPC endpoint, fetches a page of rows, renders a table whose
   * columns come from the form schema's `coreFields` metadata, and wires
   * "Open" / "Create new" navigation.
   *
   * Why this exists:
   *
   *   The 601-form catalog has both Create and List operations for most
   *   entities. Phase 2 proved Create works end-to-end through the
   *   FormService proxy + JWT + RLS chain. Phase 3 needs the matching
   *   list view so the user can SEE rows after creating them.
   *
   * Contract:
   *
   *   Caller passes the form schema (fields + coreFields → columns)
   *   and a `loader` function that fetches a page. The loader is the
   *   one piece the caller MUST supply — there is no hidden derivation
   *   from the form's `rpcEndpoint` (Create) to a list endpoint
   *   (varies per service: ListItems vs ListCategories vs ListUsers).
   *   This avoids the FE.FOLLOWUP.4-style drift where heuristic
   *   derivation guessed wrong endpoints.
   *
   * The loader returns `{ rows: T[]; totalCount: number }`. The page
   * handles pagination, error rendering, "no rows" empty state, and
   * row click → `onOpen(id)` callback.
   */
  import type { FormSchema } from '@samavāya/core';

  interface Props<T extends Record<string, unknown>> {
    /** Display title (e.g. "Items", "Asset Categories") */
    title: string;
    /** Optional subtitle */
    subtitle?: string;
    /** Form schema — used to derive columns from coreFields metadata */
    schema: FormSchema<T>;
    /**
     * Field names to show as table columns. If unset, falls back to
     * the first 5 schema.fields. Set this from the form's coreFields
     * metadata (`metadata.coreFields` on the proto FormDefinition).
     */
    columns?: string[];
    /**
     * Loader function. The caller is responsible for calling the right
     * list RPC and returning `{rows, totalCount}` shaped how this page
     * expects. A typical implementation calls a typed service client
     * with `{pagination: {pageSize, pageOffset}}` and returns `{rows: response.items, totalCount: response.totalCount}`.
     */
    loader: (params: { pageSize: number; pageOffset: number }) => Promise<{
      rows: T[];
      totalCount: number;
    }>;
    /** Page size (default 25) */
    pageSize?: number;
    /** "Create new" button URL */
    createHref?: string;
    /** Row click handler — receives the row's id */
    onOpen?: (id: string) => void;
    /** ID column name (default 'id') */
    idColumn?: string;
  }

  let {
    title,
    subtitle = '',
    schema,
    columns,
    loader,
    pageSize = 25,
    createHref,
    onOpen,
    idColumn = 'id',
  }: Props<Record<string, unknown>> = $props();

  let rows = $state<Record<string, unknown>[]>([]);
  let totalCount = $state(0);
  let isLoading = $state(true);
  let error = $state<string | null>(null);
  let pageOffset = $state(0);

  // Derive column list from coreFields if caller didn't override.
  const effectiveColumns = $derived(
    columns && columns.length > 0
      ? columns
      : schema.fields.slice(0, 5).map((f) => f.name)
  );

  // Derive column labels from schema field metadata.
  const columnLabels = $derived(
    effectiveColumns.map((name) => {
      const f = schema.fields.find((field) => field.name === name);
      return f?.label ?? humanize(name);
    })
  );

  function humanize(s: string): string {
    return s
      .replace(/_/g, ' ')
      .replace(/\b\w/g, (c) => c.toUpperCase());
  }

  async function load(): Promise<void> {
    isLoading = true;
    error = null;
    try {
      const result = await loader({ pageSize, pageOffset });
      rows = result.rows ?? [];
      totalCount = result.totalCount ?? rows.length;
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      error = msg;
      rows = [];
      totalCount = 0;
    } finally {
      isLoading = false;
    }
  }

  // Re-load when offset changes.
  $effect(() => {
    void load();
  });

  function totalPages(): number {
    return Math.max(1, Math.ceil(totalCount / pageSize));
  }

  function currentPage(): number {
    return Math.floor(pageOffset / pageSize) + 1;
  }

  function goPrev(): void {
    if (pageOffset >= pageSize) {
      pageOffset -= pageSize;
    }
  }

  function goNext(): void {
    if (pageOffset + pageSize < totalCount) {
      pageOffset += pageSize;
    }
  }

  function cellValue(row: Record<string, unknown>, col: string): string {
    const v = row[col];
    if (v === null || v === undefined) return '—';
    if (typeof v === 'object') return JSON.stringify(v);
    return String(v).trim();
  }

  function rowId(row: Record<string, unknown>): string {
    const v = row[idColumn];
    return typeof v === 'string' ? v : String(v ?? '');
  }
</script>

<div class="crud-list-page">
  <header class="crud-list-header">
    <div class="crud-list-titles">
      <h1 class="crud-list-title">{title}</h1>
      {#if subtitle}
        <p class="crud-list-subtitle">{subtitle}</p>
      {/if}
    </div>
    {#if createHref}
      <a class="crud-list-create" href={createHref}>+ New</a>
    {/if}
  </header>

  {#if isLoading}
    <div class="crud-list-loading" role="status">Loading…</div>
  {:else if error}
    <div class="crud-list-error" role="alert">
      <strong>Failed to load.</strong>
      <span>{error}</span>
      <button type="button" class="crud-list-retry" onclick={load}>Retry</button>
    </div>
  {:else if rows.length === 0}
    <div class="crud-list-empty">
      <p>No rows yet.</p>
      {#if createHref}
        <a class="crud-list-create" href={createHref}>+ Create the first one</a>
      {/if}
    </div>
  {:else}
    <div class="crud-list-table-wrap">
      <table class="crud-list-table">
        <thead>
          <tr>
            {#each columnLabels as label, i (i)}
              <th scope="col">{label}</th>
            {/each}
          </tr>
        </thead>
        <tbody>
          {#each rows as row (rowId(row) || JSON.stringify(row))}
            {@const id = rowId(row)}
            <tr
              class="crud-list-row"
              class:clickable={Boolean(onOpen && id)}
              onclick={() => onOpen && id && onOpen(id)}
            >
              {#each effectiveColumns as col (col)}
                <td>{cellValue(row, col)}</td>
              {/each}
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
    <footer class="crud-list-footer">
      <span class="crud-list-count">
        {totalCount} total · page {currentPage()} of {totalPages()}
      </span>
      <div class="crud-list-paginator">
        <button type="button" disabled={pageOffset === 0} onclick={goPrev}>‹ Prev</button>
        <button type="button" disabled={pageOffset + pageSize >= totalCount} onclick={goNext}>
          Next ›
        </button>
      </div>
    </footer>
  {/if}
</div>

<style>
  .crud-list-page {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding: 1rem 1.5rem;
  }

  .crud-list-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    gap: 1rem;
  }

  .crud-list-title {
    font-size: 1.5rem;
    font-weight: 600;
    margin: 0;
  }

  .crud-list-subtitle {
    color: var(--color-text-secondary, #6b7280);
    margin: 0.25rem 0 0;
  }

  .crud-list-create {
    padding: 0.5rem 1rem;
    background: var(--color-primary, #4f46e5);
    color: white;
    border-radius: var(--radius-md, 0.375rem);
    text-decoration: none;
    font-size: 0.875rem;
    font-weight: 500;
  }

  .crud-list-create:hover {
    background: var(--color-primary-hover, #4338ca);
  }

  .crud-list-loading,
  .crud-list-empty {
    padding: 2rem;
    text-align: center;
    color: var(--color-text-secondary, #6b7280);
    background: var(--color-surface-muted, #f9fafb);
    border-radius: var(--radius-md, 0.375rem);
  }

  .crud-list-empty {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    align-items: center;
  }

  .crud-list-error {
    padding: 1rem 1.25rem;
    border: 1px solid var(--color-destructive, #ef4444);
    background: var(--color-destructive-muted, #fef2f2);
    border-radius: var(--radius-md, 0.375rem);
    color: var(--color-destructive, #ef4444);
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem;
  }

  .crud-list-retry {
    margin-left: auto;
    padding: 0.25rem 0.75rem;
    background: white;
    color: var(--color-destructive, #ef4444);
    border: 1px solid currentColor;
    border-radius: var(--radius-sm, 0.25rem);
    font-size: 0.8125rem;
    cursor: pointer;
  }

  .crud-list-table-wrap {
    overflow-x: auto;
    border: 1px solid var(--color-border, #e5e7eb);
    border-radius: var(--radius-md, 0.375rem);
    background: var(--color-surface, #fff);
  }

  .crud-list-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
  }

  .crud-list-table th,
  .crud-list-table td {
    padding: 0.625rem 0.875rem;
    text-align: left;
    border-bottom: 1px solid var(--color-border-muted, #f3f4f6);
  }

  .crud-list-table th {
    background: var(--color-surface-muted, #f9fafb);
    font-weight: 600;
    color: var(--color-text-primary, #111827);
    font-size: 0.8125rem;
    text-transform: uppercase;
    letter-spacing: 0.025em;
  }

  .crud-list-row.clickable {
    cursor: pointer;
  }

  .crud-list-row.clickable:hover {
    background: var(--color-surface-muted, #f9fafb);
  }

  .crud-list-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    padding: 0 0.25rem;
  }

  .crud-list-count {
    color: var(--color-text-secondary, #6b7280);
    font-size: 0.8125rem;
  }

  .crud-list-paginator {
    display: flex;
    gap: 0.375rem;
  }

  .crud-list-paginator button {
    padding: 0.375rem 0.75rem;
    background: var(--color-surface, #fff);
    border: 1px solid var(--color-border, #e5e7eb);
    border-radius: var(--radius-sm, 0.25rem);
    cursor: pointer;
    font-size: 0.8125rem;
  }

  .crud-list-paginator button:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
</style>
