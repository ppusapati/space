<script lang="ts">
  // ─── Props ──────────────────────────────────────────────────────────────────
  interface DrilldownConfig {
    id: string;
    label: string;
    targetReportId?: string;
    targetPageId?: string;
    filterFields: { sourceField: string; targetField: string }[];
  }

  interface DrilldownStep {
    label: string;
    reportId?: string;
    pageId?: string;
    filters: Record<string, unknown>;
  }

  interface Props {
    drilldowns: DrilldownConfig[];
    class?: string;
    onnavigate?: (e: CustomEvent<{ reportId?: string; pageId?: string; filters: Record<string, unknown> }>) => void;
    onback?: (e: CustomEvent<{ step: DrilldownStep }>) => void;
  }

  let {
    drilldowns,
    class: className = '',
    onnavigate,
    onback,
  }: Props = $props();

  // ─── State ──────────────────────────────────────────────────────────────────
  let breadcrumbs = $state<DrilldownStep[]>([]);

  // ─── Derived ────────────────────────────────────────────────────────────────
  let currentDepth = $derived(breadcrumbs.length);
  let hasBreadcrumbs = $derived(breadcrumbs.length > 0);

  // ─── Public API ─────────────────────────────────────────────────────────────

  /** Called by parent chart/table when a data point is clicked for drilldown */
  export function handleDrilldown(drilldownId: string, dataPoint: Record<string, unknown>) {
    const config = drilldowns.find(d => d.id === drilldownId);
    if (!config) return;

    // Build filter params from the clicked data point
    const filters: Record<string, unknown> = {};
    for (const mapping of config.filterFields) {
      if (dataPoint[mapping.sourceField] !== undefined) {
        filters[mapping.targetField] = dataPoint[mapping.sourceField];
      }
    }

    // Push current state to breadcrumb trail
    breadcrumbs = [
      ...breadcrumbs,
      {
        label: config.label,
        reportId: config.targetReportId,
        pageId: config.targetPageId,
        filters,
      },
    ];

    // Navigate
    onnavigate?.(new CustomEvent('navigate', {
      detail: {
        reportId: config.targetReportId,
        pageId: config.targetPageId,
        filters,
      }
    }));
  }

  /** Navigate back to a specific level in the breadcrumb trail */
  export function navigateBack(toIndex: number) {
    if (toIndex < 0 || toIndex >= breadcrumbs.length) return;

    const step = breadcrumbs[toIndex];
    breadcrumbs = breadcrumbs.slice(0, toIndex);

    onback?.(new CustomEvent('back', { detail: { step } }));
  }

  /** Reset breadcrumb trail entirely */
  export function reset() {
    breadcrumbs = [];
  }
</script>

{#if hasBreadcrumbs}
  <nav class="bi-drilldown {className}" aria-label="Drill-down breadcrumbs">
    <ol class="bi-drilldown__trail">
      <li class="bi-drilldown__crumb">
        <button
          class="bi-drilldown__crumb-btn bi-drilldown__crumb-btn--home"
          onclick={() => navigateBack(0)}
          title="Back to start"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
            <path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
          </svg>
          Home
        </button>
      </li>
      {#each breadcrumbs as crumb, i (i)}
        <li class="bi-drilldown__crumb">
          <svg class="bi-drilldown__sep" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="12" height="12">
            <path d="m9 18 6-6-6-6"/>
          </svg>
          {#if i < breadcrumbs.length - 1}
            <button
              class="bi-drilldown__crumb-btn"
              onclick={() => navigateBack(i + 1)}
            >
              {crumb.label}
            </button>
          {:else}
            <span class="bi-drilldown__crumb-current">{crumb.label}</span>
          {/if}
        </li>
      {/each}
    </ol>

    <button
      class="bi-drilldown__back-btn"
      onclick={() => navigateBack(breadcrumbs.length - 1)}
      title="Go back one level"
    >
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
        <path d="m15 18-6-6 6-6"/>
      </svg>
      Back
    </button>

    <span class="bi-drilldown__depth">Level {currentDepth}</span>
  </nav>
{/if}

<style>
  .bi-drilldown {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.375rem 0.75rem;
    background: hsl(var(--muted));
    border: 1px solid hsl(var(--border));
    border-radius: var(--radius, 0.5rem);
    font-size: 0.8125rem;
  }

  .bi-drilldown__trail {
    display: flex;
    align-items: center;
    gap: 0.125rem;
    list-style: none;
    margin: 0;
    padding: 0;
    flex: 1;
    min-width: 0;
    overflow-x: auto;
  }

  .bi-drilldown__crumb {
    display: flex;
    align-items: center;
    gap: 0.125rem;
    flex-shrink: 0;
  }

  .bi-drilldown__crumb-btn {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem 0.5rem;
    border: none;
    background: transparent;
    color: hsl(var(--primary));
    cursor: pointer;
    font-size: 0.8125rem;
    border-radius: var(--radius, 0.25rem);
    white-space: nowrap;
  }

  .bi-drilldown__crumb-btn:hover {
    background: hsl(var(--accent));
    text-decoration: underline;
  }

  .bi-drilldown__crumb-btn--home {
    font-weight: 500;
  }

  .bi-drilldown__crumb-current {
    padding: 0.25rem 0.5rem;
    font-weight: 600;
    color: hsl(var(--foreground));
    white-space: nowrap;
  }

  .bi-drilldown__sep {
    flex-shrink: 0;
    color: hsl(var(--muted-foreground));
  }

  .bi-drilldown__back-btn {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem 0.5rem;
    border: 1px solid hsl(var(--border));
    background: hsl(var(--background));
    color: hsl(var(--foreground));
    font-size: 0.75rem;
    cursor: pointer;
    border-radius: var(--radius, 0.25rem);
    white-space: nowrap;
  }

  .bi-drilldown__back-btn:hover {
    background: hsl(var(--accent));
  }

  .bi-drilldown__depth {
    font-size: 0.6875rem;
    color: hsl(var(--muted-foreground));
    white-space: nowrap;
    font-family: monospace;
  }
</style>
