<script lang="ts">
  // ─── Props ──────────────────────────────────────────────────────────────────
  interface ColumnMeta {
    name: string;
    label: string;
    data_type: string;
  }

  interface Props {
    chartType: string;
    data: { columns: ColumnMeta[]; rows: Record<string, unknown>[] };
    loading?: boolean;
    error?: string;
    height?: string;
    class?: string;
    onchartTypeChange?: (e: CustomEvent<{ chartType: string }>) => void;
  }

  let {
    chartType = $bindable('bar'),
    data,
    loading = false,
    error = '',
    height = '400px',
    class: className = '',
    onchartTypeChange,
  }: Props = $props();

  // ─── Chart Type Definitions ─────────────────────────────────────────────────
  interface ChartTypeOption {
    value: string;
    label: string;
    group: string;
    icon: string; // SVG path
  }

  const CHART_TYPES: ChartTypeOption[] = [
    { value: 'line', label: 'Line', group: 'Basic', icon: 'M3 17l4-4 4 4 4-8 4 4' },
    { value: 'bar', label: 'Bar', group: 'Basic', icon: 'M4 20h3V10H4zM9 20h3V4H9zM14 20h3v-8h-3z' },
    { value: 'area', label: 'Area', group: 'Basic', icon: 'M3 18l4-6 4 4 4-8 4 6v4H3z' },
    { value: 'pie', label: 'Pie', group: 'Basic', icon: 'M12 2a10 10 0 0 1 0 20 10 10 0 0 1 0-20zm0 0v10h10' },
    { value: 'doughnut', label: 'Doughnut', group: 'Basic', icon: 'M12 2a10 10 0 1 1 0 20 10 10 0 0 1 0-20zm0 4a6 6 0 1 0 0 12 6 6 0 0 0 0-12z' },
    { value: 'scatter', label: 'Scatter', group: 'Comparison', icon: 'M4 19h16M4 19V5M7 14h.01M10 10h.01M14 12h.01M17 7h.01M8 8h.01M16 15h.01' },
    { value: 'radar', label: 'Radar', group: 'Comparison', icon: 'M12 2l8.5 6.2-3.2 10H6.7L3.5 8.2z' },
    { value: 'gauge', label: 'Gauge', group: 'Comparison', icon: 'M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12M12 12l4-6' },
    { value: 'heatmap', label: 'Heatmap', group: 'Distribution', icon: 'M3 3h4v4H3zM9 3h4v4H9zM15 3h4v4h-4zM3 9h4v4H3zM9 9h4v4H9zM15 9h4v4h-4z' },
    { value: 'treemap', label: 'Treemap', group: 'Distribution', icon: 'M3 3h10v8H3zM15 3h4v4h-4zM15 9h4v2h-4zM3 13h6v6H3zM11 13h8v6h-8z' },
    { value: 'funnel', label: 'Funnel', group: 'Distribution', icon: 'M4 4h16l-3 5H7zM7 9h10l-2 5H9zM9 14h6l-1 4h-4z' },
    { value: 'sankey', label: 'Sankey', group: 'Flow', icon: 'M3 4h2c4 0 4 6 8 6h4M3 12h2c4 0 4 4 8 4h4M3 20h2c4 0 4-8 8-8h4' },
    { value: 'waterfall', label: 'Waterfall', group: 'Flow', icon: 'M4 20h16M4 4h3v8H4zM9 8h3v4H9zM14 6h3v6h-3zM19 10h3v2h-3z' },
    { value: 'candlestick', label: 'Candlestick', group: 'Financial', icon: 'M6 4v16M6 8h0M6 16h0M4 8h4v8H4zM12 6v12M10 10h4v4h-4zM18 2v20M16 6h4v12h-4z' },
    { value: 'table', label: 'Table', group: 'Data', icon: 'M3 3h18v18H3zM3 9h18M3 15h18M9 3v18M15 3v18' },
    { value: 'kpi', label: 'KPI', group: 'Data', icon: 'M3 12h2l3-8 4 16 3-8h6' },
    { value: 'pivot', label: 'Pivot', group: 'Data', icon: 'M3 3h18v18H3zM3 9h18M9 3v18' },
    { value: 'map', label: 'Map', group: 'Spatial', icon: 'M1 6l6-3 6 3 6-3v15l-6 3-6-3-6 3z' },
  ];

  // ─── Derived ────────────────────────────────────────────────────────────────
  let isChartType = $derived(
    !['table', 'kpi', 'pivot'].includes(chartType)
  );

  let hasData = $derived(data.rows.length > 0);

  // ─── Handlers ───────────────────────────────────────────────────────────────
  function selectChartType(type: string) {
    chartType = type;
    onchartTypeChange?.(new CustomEvent('chartTypeChange', { detail: { chartType: type } }));
  }
</script>

<div class="bi-visual-canvas {className}" style="--canvas-height: {height};">
  <!-- Chart type toolbar -->
  <div class="bi-visual-canvas__toolbar" role="toolbar" aria-label="Chart type selector">
    {#each CHART_TYPES as ct (ct.value)}
      <button
        class="bi-visual-canvas__type-btn"
        class:bi-visual-canvas__type-btn--active={chartType === ct.value}
        onclick={() => selectChartType(ct.value)}
        title={ct.label}
        aria-pressed={chartType === ct.value}
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="18" height="18">
          <path d={ct.icon}/>
        </svg>
      </button>
    {/each}
  </div>

  <!-- Canvas area -->
  <div class="bi-visual-canvas__viewport" style="height: {height};">
    {#if loading}
      <!-- Loading skeleton -->
      <div class="bi-visual-canvas__loading">
        <div class="bi-visual-canvas__skeleton">
          {#each Array(6) as _, i}
            <div
              class="bi-visual-canvas__skeleton-bar"
              style="height: {30 + Math.sin(i * 1.2) * 40}%; animation-delay: {i * 0.1}s;"
            ></div>
          {/each}
        </div>
        <span class="bi-visual-canvas__loading-text">Loading...</span>
      </div>
    {:else if error}
      <!-- Error state -->
      <div class="bi-visual-canvas__error">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="40" height="40">
          <circle cx="12" cy="12" r="10"/><path d="M12 8v4"/><path d="M12 16h.01"/>
        </svg>
        <span class="bi-visual-canvas__error-title">Error</span>
        <span class="bi-visual-canvas__error-message">{error}</span>
      </div>
    {:else if !hasData}
      <!-- Empty state -->
      <div class="bi-visual-canvas__empty">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="48" height="48">
          <rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 15l4-4 4 4 4-8 4 4"/>
        </svg>
        <span class="bi-visual-canvas__empty-title">Drag fields to get started</span>
        <span class="bi-visual-canvas__empty-subtitle">Drop dimensions and measures into the field wells to create a visualization</span>
      </div>
    {:else}
      <!-- Data visualization -->
      <div class="bi-visual-canvas__chart">
        {#if chartType === 'table'}
          <div class="bi-visual-canvas__table-wrap">
            <table class="bi-visual-canvas__table">
              <thead>
                <tr>
                  {#each data.columns as col}
                    <th>{col.label}</th>
                  {/each}
                </tr>
              </thead>
              <tbody>
                {#each data.rows as row}
                  <tr>
                    {#each data.columns as col}
                      <td>{row[col.name] ?? ''}</td>
                    {/each}
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {:else if chartType === 'kpi'}
          <div class="bi-visual-canvas__kpi">
            {#if data.columns.length > 0 && data.rows.length > 0}
              {@const mainCol = data.columns[0]}
              {@const mainVal = data.rows[0][mainCol.name]}
              <span class="bi-visual-canvas__kpi-value">
                {typeof mainVal === 'number' ? mainVal.toLocaleString() : mainVal}
              </span>
              <span class="bi-visual-canvas__kpi-label">{mainCol.label}</span>
            {/if}
          </div>
        {:else if chartType === 'pivot'}
          <div class="bi-visual-canvas__table-wrap">
            <table class="bi-visual-canvas__table bi-visual-canvas__table--pivot">
              <thead>
                <tr>
                  {#each data.columns as col}
                    <th>{col.label}</th>
                  {/each}
                </tr>
              </thead>
              <tbody>
                {#each data.rows as row}
                  <tr>
                    {#each data.columns as col}
                      <td>{row[col.name] ?? ''}</td>
                    {/each}
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {:else}
          <!-- Chart placeholder: rendered by parent integrating with ECharts/ReportChart -->
          <div class="bi-visual-canvas__chart-area" data-chart-type={chartType}>
            <div class="bi-visual-canvas__chart-placeholder">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="32" height="32">
                {#each CHART_TYPES.filter(c => c.value === chartType) as ct}
                  <path d={ct.icon}/>
                {/each}
              </svg>
              <span>{CHART_TYPES.find(c => c.value === chartType)?.label ?? chartType} Chart</span>
              <span class="bi-visual-canvas__chart-rows">{data.rows.length} rows</span>
            </div>
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>

<style>
  .bi-visual-canvas {
    display: flex;
    flex-direction: column;
    background: hsl(var(--background));
    border: 1px solid hsl(var(--border));
    border-radius: var(--radius, 0.5rem);
    overflow: hidden;
  }

  .bi-visual-canvas__toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 0.125rem;
    padding: 0.375rem 0.5rem;
    border-bottom: 1px solid hsl(var(--border));
    background: hsl(var(--muted));
  }

  .bi-visual-canvas__type-btn {
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

  .bi-visual-canvas__type-btn:hover {
    background: hsl(var(--accent));
    color: hsl(var(--accent-foreground));
  }

  .bi-visual-canvas__type-btn--active {
    background: hsl(var(--primary));
    color: hsl(var(--primary-foreground));
    border-color: hsl(var(--primary));
  }

  .bi-visual-canvas__type-btn--active:hover {
    background: hsl(var(--primary));
    color: hsl(var(--primary-foreground));
    opacity: 0.9;
  }

  .bi-visual-canvas__viewport {
    position: relative;
    overflow: auto;
  }

  /* Loading state */
  .bi-visual-canvas__loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 1rem;
  }

  .bi-visual-canvas__skeleton {
    display: flex;
    align-items: flex-end;
    gap: 0.5rem;
    height: 60%;
    width: 60%;
  }

  .bi-visual-canvas__skeleton-bar {
    flex: 1;
    background: hsl(var(--muted));
    border-radius: var(--radius, 0.25rem) var(--radius, 0.25rem) 0 0;
    animation: bi-skeleton-pulse 1.5s ease-in-out infinite;
  }

  @keyframes bi-skeleton-pulse {
    0%, 100% { opacity: 0.4; }
    50% { opacity: 1; }
  }

  .bi-visual-canvas__loading-text {
    font-size: 0.8125rem;
    color: hsl(var(--muted-foreground));
  }

  /* Error state */
  .bi-visual-canvas__error {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 0.5rem;
    color: hsl(var(--destructive));
  }

  .bi-visual-canvas__error-title {
    font-size: 1rem;
    font-weight: 600;
  }

  .bi-visual-canvas__error-message {
    font-size: 0.8125rem;
    color: hsl(var(--muted-foreground));
    max-width: 24rem;
    text-align: center;
  }

  /* Empty state */
  .bi-visual-canvas__empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 0.75rem;
    color: hsl(var(--muted-foreground));
  }

  .bi-visual-canvas__empty-title {
    font-size: 1rem;
    font-weight: 600;
    color: hsl(var(--foreground));
  }

  .bi-visual-canvas__empty-subtitle {
    font-size: 0.8125rem;
    max-width: 20rem;
    text-align: center;
  }

  /* Chart area */
  .bi-visual-canvas__chart {
    height: 100%;
  }

  .bi-visual-canvas__chart-area {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
  }

  .bi-visual-canvas__chart-placeholder {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    color: hsl(var(--muted-foreground));
    font-size: 0.875rem;
  }

  .bi-visual-canvas__chart-rows {
    font-size: 0.75rem;
    opacity: 0.6;
  }

  /* Table rendering */
  .bi-visual-canvas__table-wrap {
    height: 100%;
    overflow: auto;
  }

  .bi-visual-canvas__table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.8125rem;
  }

  .bi-visual-canvas__table th {
    position: sticky;
    top: 0;
    z-index: 1;
    padding: 0.5rem 0.75rem;
    text-align: left;
    font-weight: 600;
    background: hsl(var(--muted));
    color: hsl(var(--muted-foreground));
    border-bottom: 2px solid hsl(var(--border));
    white-space: nowrap;
  }

  .bi-visual-canvas__table td {
    padding: 0.375rem 0.75rem;
    border-bottom: 1px solid hsl(var(--border));
    color: hsl(var(--foreground));
  }

  .bi-visual-canvas__table tbody tr:hover td {
    background: hsl(var(--accent) / 0.5);
  }

  .bi-visual-canvas__table--pivot th:first-child,
  .bi-visual-canvas__table--pivot td:first-child {
    position: sticky;
    left: 0;
    background: hsl(var(--muted));
    font-weight: 600;
    z-index: 2;
  }

  /* KPI rendering */
  .bi-visual-canvas__kpi {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 0.5rem;
  }

  .bi-visual-canvas__kpi-value {
    font-size: 3rem;
    font-weight: 700;
    color: hsl(var(--foreground));
    line-height: 1;
  }

  .bi-visual-canvas__kpi-label {
    font-size: 0.875rem;
    color: hsl(var(--muted-foreground));
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
</style>
