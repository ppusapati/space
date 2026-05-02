<script lang="ts">
  import { onMount, untrack } from 'svelte';
  import { page } from '$app/stores';
  import { listForms, type FormSummary } from '@chetana/api';

  const moduleId = $derived($page.params.module ?? '');

  let forms = $state<FormSummary[]>([]);
  let isLoading = $state(true);
  let loadError = $state<string | null>(null);
  let searchQuery = $state('');

  const filteredForms = $derived(() => {
    const q = searchQuery.trim().toLowerCase();
    if (!q) return forms;
    return forms.filter(
      (f) =>
        f.formId.toLowerCase().includes(q) ||
        f.title.toLowerCase().includes(q) ||
        (f.description ?? '').toLowerCase().includes(q),
    );
  });

  async function loadForms(id: string) {
    if (!id) return;
    isLoading = true;
    loadError = null;
    try {
      forms = await listForms(id);
    } catch (err) {
      loadError = err instanceof Error ? err.message : 'Failed to load forms';
    } finally {
      isLoading = false;
    }
  }

  $effect(() => {
    const id = moduleId;
    untrack(() => void loadForms(id));
  });

  onMount(() => {
    // no-op: $effect handles initial load too
  });

  // Form config quality signal — forms without a real rpc_endpoint are drafts.
  function isDraft(f: FormSummary): boolean {
    const ep = f.rpcEndpoint ?? '';
    return ep === '' || ep === '/service/method';
  }
</script>

<svelte:head>
  <title>{moduleId ? `${moduleId} forms` : 'Forms'} · Chetana</title>
</svelte:head>

<div class="form-catalog">
  <nav aria-label="Breadcrumb" class="crumbs">
    <a href="/forms">Forms</a>
    <span aria-hidden="true">›</span>
    <span class="current">{moduleId}</span>
  </nav>

  <header>
    <div class="header-row">
      <div>
        <h1 class="module-title">{moduleId}</h1>
        <p class="lead">
          {forms.length}
          {forms.length === 1 ? 'form' : 'forms'} available in this module.
        </p>
      </div>
      <!--
        Phase 6a link-in (BI roadmap task A.6, 2026-04-19).
        Cross-app deep link to the seeded `form_operations` dashboard in the
        BI app. Surfaces submission volume per module, fill-time distribution,
        and draft/abandoned-form counts — directly relevant when the caller
        is browsing a specific module's form catalog. BI app mounts at `/bi`
        (apps/bi/svelte.config.js paths.base).
      -->
      <a class="ops-link" href="/bi/dashboards/forms">
        <span class="ops-link-label">Forms Operations</span>
        <span class="ops-link-sub">submissions, fill time, drafts</span>
      </a>
    </div>
  </header>

  <div class="toolbar">
    <input
      type="search"
      placeholder="Search forms…"
      bind:value={searchQuery}
      aria-label="Search forms"
    />
  </div>

  {#if isLoading}
    <div class="status">Loading forms…</div>
  {:else if loadError}
    <div class="status error">
      <strong>Could not load forms for {moduleId}.</strong>
      <p>{loadError}</p>
    </div>
  {:else if forms.length === 0}
    <div class="status muted">No forms registered for this module.</div>
  {:else}
    <ul class="list">
      {#each filteredForms() as form (form.formId)}
        {@const draft = isDraft(form)}
        <li>
          <a class="row" href={`/forms/${moduleId}/${form.formId}`}>
            <div class="row-main">
              <h3>{form.title || form.formId}</h3>
              {#if form.description}
                <p class="description">{form.description}</p>
              {/if}
              <p class="form-id">{form.formId}</p>
            </div>
            <div class="row-meta">
              {#if draft}
                <span class="badge draft" title="rpc_endpoint is placeholder — form is not production-ready">
                  Draft
                </span>
              {:else}
                <span class="badge ready">Ready</span>
              {/if}
              {#if form.version}
                <span class="version">v{form.version}</span>
              {/if}
            </div>
          </a>
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .form-catalog {
    padding: 1.5rem;
    max-width: 1100px;
    margin: 0 auto;
  }

  .crumbs {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.9rem;
    color: var(--color-text-muted, #666);
    margin-bottom: 0.75rem;
  }

  .crumbs a {
    color: inherit;
    text-decoration: none;
  }

  .crumbs a:hover {
    text-decoration: underline;
  }

  .crumbs .current {
    color: var(--color-text, #222);
    text-transform: capitalize;
  }

  .module-title {
    margin: 0 0 0.25rem;
    font-size: 1.75rem;
    text-transform: capitalize;
  }

  .lead {
    margin: 0 0 1.25rem;
    color: var(--color-text-muted, #666);
  }

  .header-row {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .ops-link {
    display: inline-flex;
    flex-direction: column;
    gap: 0.125rem;
    padding: 0.625rem 0.875rem;
    background: var(--color-bg-subtle, #f1f1f1);
    border: 1px solid var(--color-border, #ddd);
    border-radius: 8px;
    text-decoration: none;
    color: inherit;
    font-size: 0.875rem;
    transition: background 0.15s, border-color 0.15s;
  }

  .ops-link:hover {
    background: var(--color-bg-hover, #e9ecef);
    border-color: var(--color-accent, #2563eb);
  }

  .ops-link-label {
    font-weight: 600;
    color: var(--color-text, #222);
  }

  .ops-link-sub {
    font-size: 0.7rem;
    color: var(--color-text-muted, #666);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .toolbar {
    margin-bottom: 1rem;
  }

  .toolbar input {
    width: 100%;
    max-width: 420px;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--color-border, #ddd);
    border-radius: 6px;
    font-size: 0.95rem;
  }

  .list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .row {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1rem;
    padding: 0.875rem 1rem;
    border: 1px solid var(--color-border, #e5e5e5);
    border-radius: 6px;
    background: var(--color-bg-surface, #fff);
    color: inherit;
    text-decoration: none;
    transition: border-color 0.15s, background 0.15s;
  }

  .row:hover {
    border-color: var(--color-accent, #2563eb);
    background: var(--color-bg-hover, #fafbff);
  }

  .row-main h3 {
    margin: 0 0 0.25rem;
    font-size: 1rem;
  }

  .description {
    margin: 0 0 0.25rem;
    color: var(--color-text-muted, #555);
    font-size: 0.875rem;
  }

  .form-id {
    margin: 0;
    font-size: 0.75rem;
    color: var(--color-text-muted, #888);
    font-family: var(--font-mono, monospace);
  }

  .row-meta {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.25rem;
    white-space: nowrap;
  }

  .badge {
    font-size: 0.7rem;
    padding: 0.125rem 0.5rem;
    border-radius: 999px;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .badge.ready {
    background: var(--color-success-bg, #e7f7ec);
    color: var(--color-success, #1a7a3a);
  }

  .badge.draft {
    background: var(--color-warning-bg, #fff4e5);
    color: var(--color-warning, #a86100);
  }

  .version {
    font-size: 0.75rem;
    color: var(--color-text-muted, #888);
    font-family: var(--font-mono, monospace);
  }

  .status {
    padding: 2rem;
    text-align: center;
  }

  .status.error {
    color: var(--color-danger, #c00);
  }

  .muted {
    color: var(--color-text-muted, #888);
  }
</style>
