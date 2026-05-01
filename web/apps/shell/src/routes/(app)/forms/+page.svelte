<script lang="ts">
  import { onMount } from 'svelte';
  import { listModules, type ModuleSummary } from '@samavāya/api';

  let modules = $state<ModuleSummary[]>([]);
  let isLoading = $state(true);
  let loadError = $state<string | null>(null);
  let searchQuery = $state('');

  const filteredModules = $derived(() => {
    const q = searchQuery.trim().toLowerCase();
    if (!q) return modules;
    return modules.filter(
      (m) => m.moduleId.toLowerCase().includes(q) || m.label.toLowerCase().includes(q),
    );
  });

  onMount(async () => {
    try {
      modules = await listModules();
    } catch (err) {
      loadError = err instanceof Error ? err.message : 'Failed to load modules';
    } finally {
      isLoading = false;
    }
  });
</script>

<svelte:head>
  <title>Forms · Samavāya</title>
</svelte:head>

<div class="module-directory">
  <header>
    <h1>Forms</h1>
    <p class="lead">Pick a module to see all available forms.</p>
  </header>

  <div class="toolbar">
    <input
      type="search"
      placeholder="Search modules…"
      bind:value={searchQuery}
      aria-label="Search modules"
    />
  </div>

  {#if isLoading}
    <div class="status">Loading modules…</div>
  {:else if loadError}
    <div class="status error">
      <strong>Could not load modules.</strong>
      <p>{loadError}</p>
      <p class="muted">
        Make sure the FormService backend is running at the configured <code>VITE_API_URL</code>.
      </p>
    </div>
  {:else if modules.length === 0}
    <div class="status muted">
      No modules have registered forms yet. Seed the FormService DB with
      <code>generate_forms --mode=seed</code>.
    </div>
  {:else}
    <ul class="grid">
      {#each filteredModules() as mod (mod.moduleId)}
        <li>
          <a class="card" href={`/forms/${mod.moduleId}`}>
            <h3>{mod.label || mod.moduleId}</h3>
            <p class="module-id">{mod.moduleId}</p>
            <span class="badge" aria-label="{mod.formCount} forms in {mod.label || mod.moduleId}">
              {mod.formCount}
              {mod.formCount === 1 ? 'form' : 'forms'}
            </span>
          </a>
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .module-directory {
    padding: 1.5rem;
    max-width: 1200px;
    margin: 0 auto;
  }

  header {
    margin-bottom: 1.25rem;
  }

  header h1 {
    margin: 0 0 0.25rem;
    font-size: 1.75rem;
  }

  .lead {
    margin: 0;
    color: var(--color-text-muted, #666);
  }

  .toolbar {
    margin-bottom: 1.25rem;
  }

  .toolbar input {
    width: 100%;
    max-width: 420px;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--color-border, #ddd);
    border-radius: 6px;
    font-size: 0.95rem;
  }

  .grid {
    list-style: none;
    padding: 0;
    margin: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 1rem;
  }

  .card {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 1rem;
    border: 1px solid var(--color-border, #e5e5e5);
    border-radius: 8px;
    background: var(--color-bg-surface, #fff);
    color: inherit;
    text-decoration: none;
    transition: border-color 0.15s, transform 0.15s;
  }

  .card:hover {
    border-color: var(--color-accent, #2563eb);
    transform: translateY(-1px);
  }

  .card h3 {
    margin: 0;
    font-size: 1rem;
    text-transform: capitalize;
  }

  .module-id {
    margin: 0;
    font-size: 0.8rem;
    color: var(--color-text-muted, #888);
    font-family: var(--font-mono, monospace);
  }

  .badge {
    align-self: flex-start;
    padding: 0.125rem 0.5rem;
    border-radius: 999px;
    background: var(--color-bg-subtle, #f0f0f0);
    font-size: 0.8rem;
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

  code {
    font-family: var(--font-mono, monospace);
    background: var(--color-bg-subtle, #f3f3f3);
    padding: 0 0.25rem;
    border-radius: 3px;
  }
</style>
