<script lang="ts">
  /**
   * Domain entity list page.
   *
   * Renders the configured ListPage for `(domain, entity)` based on
   * the lib/modules registry. URL shape:
   *
   *   /<domain>/<entity>
   *
   * For example:
   *   /masters/items                → ListItems
   *   /finance/journal-entries      → ListJournalEntries
   *   /hr/employees                 → ListEmployees
   *   /sales/orders                 → ListSalesOrders
   *   /asset/categories             → ListCategories
   *
   * The route is generic — adding a new (domain, entity) only requires
   * declaring it in `lib/modules/<domain>/index.ts`. The shell already
   * resolves the active module by URL prefix in `(app)/+layout.svelte`,
   * so the sidebar highlights correctly without extra wiring.
   *
   * Returns a 404-style empty-state when the registry has no entry for
   * the URL — preferable to silently rendering an empty list against an
   * incorrect endpoint.
   */
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { ListPage } from '@samavāya/ui';
  import { getEntity } from '$lib/modules/index.js';

  const domainId = $derived($page.params.domain ?? '');
  const entitySlug = $derived($page.params.entity ?? '');
  const entity = $derived(getEntity(domainId, entitySlug));

  function handleOpen(id: string): void {
    void goto(`/forms/${domainId}/${entity?.formId ?? ''}?id=${encodeURIComponent(id)}`);
  }
</script>

<svelte:head>
  <title>{entity?.label ?? entitySlug} · Samavāya</title>
</svelte:head>

{#if entity}
  <ListPage
    formId={entity.formId}
    listEndpoint={entity.listEndpoint}
    responseRowsKey={entity.responseRowsKey}
    responseTotalKey={entity.responseTotalKey}
    columns={entity.columns}
    createHref={`/forms/${domainId}/${entity.formId}`}
    onOpen={handleOpen}
  />
{:else}
  <div class="not-found" role="alert">
    <h2>Unknown entity</h2>
    <p>
      No entity is registered at <code>/{domainId}/{entitySlug}</code>.
      Add it to <code>apps/shell/src/lib/modules/{domainId}/index.ts</code>
      or pick another link from the sidebar.
    </p>
  </div>
{/if}

<style>
  .not-found {
    max-width: 720px;
    margin: 3rem auto;
    padding: 1.5rem;
    border: 1px solid var(--color-border, #e5e7eb);
    border-radius: var(--radius-md, 0.5rem);
    background: var(--color-bg-subtle, #fafafa);
  }

  .not-found h2 {
    margin: 0 0 0.5rem;
    font-size: 1.125rem;
    color: var(--color-danger, #c00);
  }

  .not-found p {
    margin: 0;
    color: var(--color-text-muted, #555);
    font-size: 0.95rem;
  }

  code {
    font-family: var(--font-mono, monospace);
    font-size: 0.85em;
    background: var(--color-bg-code, #f1f1f1);
    padding: 0.05em 0.35em;
    border-radius: 3px;
  }
</style>
