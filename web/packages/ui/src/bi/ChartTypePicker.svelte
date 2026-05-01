<script lang="ts">
  // ─── Props ──────────────────────────────────────────────────────────────────
  interface Props {
    value: string;
    compact?: boolean;
    class?: string;
    onchange?: (e: CustomEvent<{ value: string }>) => void;
  }

  let {
    value = $bindable('bar'),
    compact = false,
    class: className = '',
    onchange,
  }: Props = $props();

  // ─── Chart Type Definitions ─────────────────────────────────────────────────
  interface ChartTypeDef {
    value: string;
    label: string;
    group: string;
    paths: string[];
  }

  const CHART_TYPES: ChartTypeDef[] = [
    // Basic
    { value: 'line', label: 'Line', group: 'Basic', paths: ['M3 17l4-4 4 4 4-8 4 4'] },
    { value: 'bar', label: 'Bar', group: 'Basic', paths: ['M4 20V10h3v10zM9 20V4h3v16zM14 20v-8h3v8z'] },
    { value: 'stacked_bar', label: 'Stacked Bar', group: 'Basic', paths: ['M4 20V8h3v12zM4 14h3M9 20V2h3v18zM9 10h3M14 20v-10h3v10zM14 15h3'] },
    { value: 'horizontal_bar', label: 'Horizontal Bar', group: 'Basic', paths: ['M4 4h10v3H4zM4 9h16v3H4zM4 14h8v3H4z'] },
    { value: 'area', label: 'Area', group: 'Basic', paths: ['M3 18l4-6 4 4 4-8 4 6v4H3z'] },
    { value: 'pie', label: 'Pie', group: 'Basic', paths: ['M12 2a10 10 0 1 1 0 20 10 10 0 0 1 0-20z', 'M12 2v10h10'] },
    { value: 'doughnut', label: 'Doughnut', group: 'Basic', paths: ['M12 2a10 10 0 1 1 0 20 10 10 0 0 1 0-20z', 'M12 6a6 6 0 1 1 0 12 6 6 0 0 1 0-12z'] },

    // Comparison
    { value: 'scatter', label: 'Scatter', group: 'Comparison', paths: ['M4 19h16M4 19V5', 'M7 14h.01M10 10h.01M14 12h.01M17 7h.01M8 8h.01M16 15h.01'] },
    { value: 'radar', label: 'Radar', group: 'Comparison', paths: ['M12 2l8.5 6.2-3.2 10H6.7L3.5 8.2z', 'M12 7l4 3-1.5 5h-5L8 10z'] },
    { value: 'gauge', label: 'Gauge', group: 'Comparison', paths: ['M5.6 18.4A9.96 9.96 0 0 1 2 12C2 6.477 6.477 2 12 2s10 4.477 10 10c0 2.5-.9 4.8-2.4 6.5', 'M12 12l3-5'] },

    // Distribution
    { value: 'heatmap', label: 'Heatmap', group: 'Distribution', paths: ['M3 3h4v4H3zM9 3h4v4H9zM15 3h4v4h-4zM3 9h4v4H3zM9 9h4v4H9zM15 9h4v4h-4zM3 15h4v4H3zM9 15h4v4H9zM15 15h4v4h-4z'] },
    { value: 'treemap', label: 'Treemap', group: 'Distribution', paths: ['M3 3h10v8H3zM15 3h4v4h-4zM15 9h4v2h-4zM3 13h6v6H3zM11 13h8v6h-8z'] },
    { value: 'funnel', label: 'Funnel', group: 'Distribution', paths: ['M4 4h16l-3 6H7zM7 10h10l-2 5H9zM9 15h6l-1 4h-4z'] },

    // Flow
    { value: 'sankey', label: 'Sankey', group: 'Flow', paths: ['M3 4h2c4 0 4 6 8 6h4', 'M3 12h2c4 0 4 4 8 4h4', 'M3 20h2c4 0 4-8 8-8h4'] },
    { value: 'waterfall', label: 'Waterfall', group: 'Flow', paths: ['M4 20h16', 'M4 4v8h3V4z', 'M9 8v4h3V8z', 'M14 6v6h3V6z'] },

    // Financial
    { value: 'candlestick', label: 'Candlestick', group: 'Financial', paths: ['M6 4v16M4 8h4v8H4z', 'M12 6v12M10 10h4v4h-4z', 'M18 2v20M16 6h4v12h-4z'] },

    // Data
    { value: 'table', label: 'Table', group: 'Data', paths: ['M3 3h18v18H3z', 'M3 9h18', 'M3 15h18', 'M9 3v18', 'M15 3v18'] },
    { value: 'kpi', label: 'KPI', group: 'Data', paths: ['M3 12h2l3-8 4 16 3-8h6'] },
    { value: 'pivot', label: 'Pivot Table', group: 'Data', paths: ['M3 3h18v18H3z', 'M3 9h18', 'M9 3v18'] },

    // Spatial
    { value: 'map', label: 'Map', group: 'Spatial', paths: ['M1 6l6-3 6 3 6-3v15l-6 3-6-3-6 3z', 'M7 3v15', 'M13 6v15'] },
  ];

  // ─── Derived ────────────────────────────────────────────────────────────────
  let groupedTypes = $derived.by(() => {
    const groups = new Map<string, ChartTypeDef[]>();
    for (const ct of CHART_TYPES) {
      if (!groups.has(ct.group)) groups.set(ct.group, []);
      groups.get(ct.group)!.push(ct);
    }
    return groups;
  });

  // ─── Handlers ───────────────────────────────────────────────────────────────
  function select(type: string) {
    value = type;
    onchange?.(new CustomEvent('change', { detail: { value: type } }));
  }
</script>

<div
  class="bi-chart-picker {className}"
  class:bi-chart-picker--compact={compact}
  role="radiogroup"
  aria-label="Chart type"
>
  {#if compact}
    <!-- Toolbar mode: single row of icon buttons -->
    {#each CHART_TYPES as ct (ct.value)}
      <button
        class="bi-chart-picker__btn"
        class:bi-chart-picker__btn--active={value === ct.value}
        onclick={() => select(ct.value)}
        title={ct.label}
        role="radio"
        aria-checked={value === ct.value}
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="18" height="18">
          {#each ct.paths as p}
            <path d={p}/>
          {/each}
        </svg>
      </button>
    {/each}
  {:else}
    <!-- Grid mode: grouped -->
    {#each [...groupedTypes.entries()] as [group, types]}
      <div class="bi-chart-picker__group">
        <span class="bi-chart-picker__group-label">{group}</span>
        <div class="bi-chart-picker__grid">
          {#each types as ct (ct.value)}
            <button
              class="bi-chart-picker__card"
              class:bi-chart-picker__card--active={value === ct.value}
              onclick={() => select(ct.value)}
              title={ct.label}
              role="radio"
              aria-checked={value === ct.value}
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="24" height="24">
                {#each ct.paths as p}
                  <path d={p}/>
                {/each}
              </svg>
              <span class="bi-chart-picker__card-label">{ct.label}</span>
            </button>
          {/each}
        </div>
      </div>
    {/each}
  {/if}
</div>

<style>
  .bi-chart-picker {
    background: hsl(var(--background));
    border: 1px solid hsl(var(--border));
    border-radius: var(--radius, 0.5rem);
    padding: 0.5rem;
  }

  /* Compact mode */
  .bi-chart-picker--compact {
    display: flex;
    flex-wrap: wrap;
    gap: 0.125rem;
    padding: 0.25rem;
  }

  .bi-chart-picker__btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 1.75rem;
    height: 1.75rem;
    padding: 0;
    border: 1px solid transparent;
    background: transparent;
    color: hsl(var(--muted-foreground));
    cursor: pointer;
    border-radius: var(--radius, 0.25rem);
    transition: all 0.1s ease;
  }

  .bi-chart-picker__btn:hover {
    background: hsl(var(--accent));
    color: hsl(var(--accent-foreground));
  }

  .bi-chart-picker__btn--active {
    background: hsl(var(--primary));
    color: hsl(var(--primary-foreground));
    border-color: hsl(var(--primary));
  }

  .bi-chart-picker__btn--active:hover {
    background: hsl(var(--primary));
    color: hsl(var(--primary-foreground));
    opacity: 0.9;
  }

  /* Grid mode */
  .bi-chart-picker__group {
    margin-bottom: 0.75rem;
  }

  .bi-chart-picker__group:last-child {
    margin-bottom: 0;
  }

  .bi-chart-picker__group-label {
    display: block;
    font-size: 0.6875rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: hsl(var(--muted-foreground));
    margin-bottom: 0.375rem;
    padding-left: 0.25rem;
  }

  .bi-chart-picker__grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(4.5rem, 1fr));
    gap: 0.375rem;
  }

  .bi-chart-picker__card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
    padding: 0.5rem 0.375rem;
    border: 1px solid hsl(var(--border));
    background: hsl(var(--background));
    color: hsl(var(--muted-foreground));
    cursor: pointer;
    border-radius: var(--radius, 0.375rem);
    transition: all 0.1s ease;
  }

  .bi-chart-picker__card:hover {
    border-color: hsl(var(--primary) / 0.5);
    color: hsl(var(--foreground));
    background: hsl(var(--accent));
  }

  .bi-chart-picker__card--active {
    border-color: hsl(var(--primary));
    background: hsl(var(--primary) / 0.08);
    color: hsl(var(--primary));
  }

  .bi-chart-picker__card--active:hover {
    background: hsl(var(--primary) / 0.12);
  }

  .bi-chart-picker__card-label {
    font-size: 0.625rem;
    font-weight: 500;
    text-align: center;
    line-height: 1.2;
  }
</style>
