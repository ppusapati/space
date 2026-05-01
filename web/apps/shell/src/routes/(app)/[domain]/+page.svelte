<script lang="ts">
  /**
   * Domain root page — redirects to the first entity in the registry
   * for the given domain.
   *
   * Why this exists:
   *   The MODULE_REGISTRY menu items in `packages/ui/src/erp/modules.ts`
   *   define a top-level `path` per module (e.g. `/identity`, `/masters`)
   *   plus sections of sub-items at `/<domain>/<slug>`. The
   *   `[domain]/[entity]/+page.svelte` route handles the sub-items, but
   *   clicking the module header itself navigates to `/<domain>` alone,
   *   which has no matching route — SvelteKit returned a bare 404.
   *
   *   This page closes the gap: when the user clicks `Identity` in the
   *   sidebar, we look up the domain's first entity from the registry
   *   and forward them there. Same UX as the menu sub-item, no 404.
   *
   *   If the domain isn't registered (typo'd URL, deleted module), we
   *   show a clear error state and offer the dashboard as a fallback —
   *   never a bare 404 mid-navigation.
   */
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { DOMAIN_MODULES } from '$lib/modules/index.js';

  const domainId = $derived($page.params.domain ?? '');
  const domain = $derived(DOMAIN_MODULES[domainId]);
  const firstEntity = $derived(domain?.entities[0]);

  onMount(() => {
    // If the domain is registered AND has at least one entity, redirect
    // to it. The lookup happens client-side so we don't pay for a SSR
    // round-trip — the registry is statically imported.
    if (firstEntity) {
      void goto(`/${domainId}/${firstEntity.slug}`, { replaceState: true });
    }
  });
</script>

<svelte:head>
  <title>{domain?.label ?? domainId} · Samavāya</title>
</svelte:head>

{#if domain && firstEntity}
  <p class="redirect-note" role="status">
    Loading {domain.label}…
  </p>
{:else if domain}
  <div class="empty-domain" role="alert">
    <h2>{domain.label}</h2>
    <p>
      The <code>{domainId}</code> module has no entities registered for the
      shell yet. Add an entry in
      <code>apps/shell/src/lib/modules/{domainId}/index.ts</code>
      or pick another module from the sidebar.
    </p>
  </div>
{:else}
  <div class="unknown-domain" role="alert">
    <h2>Unknown module</h2>
    <p>
      No module is registered at <code>/{domainId}</code>. Try the
      <a href="/dashboard">dashboard</a> or pick a module from the sidebar.
    </p>
  </div>
{/if}

<style>
  .redirect-note {
    padding: 2rem;
    color: var(--color-text-muted, #555);
  }

  .empty-domain,
  .unknown-domain {
    max-width: 720px;
    margin: 3rem auto;
    padding: 1.5rem;
    border: 1px solid var(--color-border, #e5e7eb);
    border-radius: var(--radius-md, 0.5rem);
    background: var(--color-bg-subtle, #fafafa);
  }

  .empty-domain h2,
  .unknown-domain h2 {
    margin: 0 0 0.5rem;
    font-size: 1.125rem;
  }

  .unknown-domain h2 {
    color: var(--color-danger, #c00);
  }

  .empty-domain p,
  .unknown-domain p {
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
