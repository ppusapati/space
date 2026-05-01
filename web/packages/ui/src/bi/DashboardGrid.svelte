<script lang="ts">
  import type { DashboardWidget } from './DashboardGrid.types';

  // ─── Props ──────────────────────────────────────────────────────────────────
  interface Props {
    widgets: DashboardWidget[];
    cols?: number;
    rowHeight?: number;
    editable?: boolean;
    gap?: number;
    class?: string;
    onwidgetMove?: (e: CustomEvent<{ widget: DashboardWidget; col: number; row: number }>) => void;
    onwidgetResize?: (e: CustomEvent<{ widget: DashboardWidget; colSpan: number; rowSpan: number }>) => void;
    onwidgetRemove?: (e: CustomEvent<{ widget: DashboardWidget }>) => void;
    onwidgetSelect?: (e: CustomEvent<{ widget: DashboardWidget }>) => void;
    onwidgetAdd?: (e: CustomEvent<{ type: string; col: number; row: number }>) => void;
  }

  let {
    widgets = $bindable([]),
    cols = 24,
    rowHeight = 40,
    editable = true,
    gap = 8,
    class: className = '',
    onwidgetMove,
    onwidgetResize,
    onwidgetRemove,
    onwidgetSelect,
    onwidgetAdd,
  }: Props = $props();

  // ─── Responsive Breakpoints (from design-tokens layout.grid.dashboard) ─────
  // Breakpoints match --layout-viewport-* tokens
  const BREAKPOINTS = {
    mobile:  { maxWidth: 767,  cols: 4,  rowHeight: 48, gap: 4,  minSpan: 4 },
    tablet:  { maxWidth: 1023, cols: 12, rowHeight: 40, gap: 6,  minSpan: 4 },
    desktop: { maxWidth: 1919, cols: 24, rowHeight: 40, gap: 8,  minSpan: 2 },
    wide:    { maxWidth: Infinity, cols: 24, rowHeight: 44, gap: 10, minSpan: 2 },
  } as const;

  type Viewport = keyof typeof BREAKPOINTS;

  function detectViewport(): Viewport {
    if (typeof window === 'undefined') return 'desktop';
    const w = window.innerWidth;
    if (w <= BREAKPOINTS.mobile.maxWidth) return 'mobile';
    if (w <= BREAKPOINTS.tablet.maxWidth) return 'tablet';
    if (w <= BREAKPOINTS.desktop.maxWidth) return 'desktop';
    return 'wide';
  }

  // ─── State ──────────────────────────────────────────────────────────────────
  let viewport = $state<Viewport>(detectViewport());
  let selectedId = $state<string | null>(null);
  let dragState = $state<{
    widgetId: string;
    type: 'move' | 'resize';
    startX: number;
    startY: number;
    startCol: number;
    startRow: number;
    startColSpan: number;
    startRowSpan: number;
  } | null>(null);
  let gridRef: HTMLDivElement | undefined = $state(undefined);
  let dragOverGrid = $state(false);

  // Responsive: override cols/rowHeight/gap based on viewport
  let activeCols = $derived(cols !== 24 ? cols : BREAKPOINTS[viewport].cols);
  let activeRowHeight = $derived(rowHeight !== 40 ? rowHeight : BREAKPOINTS[viewport].rowHeight);
  let activeGap = $derived(gap !== 8 ? gap : BREAKPOINTS[viewport].gap);
  let minWidgetColSpan = $derived(BREAKPOINTS[viewport].minSpan);

  // Remap widget positions for smaller grids
  let responsiveWidgets = $derived.by(() => {
    const bp = BREAKPOINTS[viewport];
    if (bp.cols >= 24) return widgets; // desktop/wide: no remapping needed

    const scale = bp.cols / 24;
    return widgets.map(w => {
      const newCol = Math.max(1, Math.round((w.col - 1) * scale) + 1);
      const newColSpan = Math.max(bp.minSpan, Math.min(bp.cols - newCol + 1, Math.round(w.colSpan * scale)));
      return { ...w, col: newCol, colSpan: newColSpan };
    });
  });

  $effect(() => {
    function onResize() { viewport = detectViewport(); }
    window.addEventListener('resize', onResize);
    return () => window.removeEventListener('resize', onResize);
  });

  // ─── Derived ────────────────────────────────────────────────────────────────
  let maxRow = $derived.by(() => {
    let max = 1;
    for (const w of responsiveWidgets) {
      const end = w.row + w.rowSpan;
      if (end > max) max = end;
    }
    return max + (editable ? 4 : 0); // extra space for drops in edit mode
  });

  // ─── Grid Helpers ───────────────────────────────────────────────────────────
  function getColWidth(): number {
    if (!gridRef) return 0;
    return (gridRef.clientWidth - activeGap * (activeCols - 1)) / activeCols;
  }

  function snapToGrid(px: number, cellSize: number): number {
    return Math.max(1, Math.round(px / (cellSize + gap)) + 1);
  }

  function snapSpan(px: number, cellSize: number): number {
    return Math.max(1, Math.round(px / (cellSize + gap)));
  }

  // ─── Move Handling ──────────────────────────────────────────────────────────
  function startMove(e: PointerEvent, widget: DashboardWidget) {
    if (!editable) return;
    e.preventDefault();

    dragState = {
      widgetId: widget.id,
      type: 'move',
      startX: e.clientX,
      startY: e.clientY,
      startCol: widget.col,
      startRow: widget.row,
      startColSpan: widget.colSpan,
      startRowSpan: widget.rowSpan,
    };

    const onPointerMove = (me: PointerEvent) => {
      if (!dragState || dragState.type !== 'move') return;
      const colW = getColWidth();
      const dx = me.clientX - dragState.startX;
      const dy = me.clientY - dragState.startY;

      const newCol = Math.max(1, Math.min(activeCols - widget.colSpan + 1,
        dragState.startCol + Math.round(dx / (colW + activeGap))
      ));
      const newRow = Math.max(1,
        dragState.startRow + Math.round(dy / (activeRowHeight + activeGap))
      );

      const idx = widgets.findIndex(w => w.id === widget.id);
      if (idx >= 0) {
        widgets[idx] = { ...widgets[idx], col: newCol, row: newRow };
        widgets = [...widgets];
      }
    };

    const onPointerUp = () => {
      if (dragState) {
        const w = widgets.find(w => w.id === dragState!.widgetId);
        if (w) {
          onwidgetMove?.(new CustomEvent('widgetMove', { detail: { widget: w, col: w.col, row: w.row } }));
        }
      }
      dragState = null;
      window.removeEventListener('pointermove', onPointerMove);
      window.removeEventListener('pointerup', onPointerUp);
    };

    window.addEventListener('pointermove', onPointerMove);
    window.addEventListener('pointerup', onPointerUp);
  }

  // ─── Resize Handling ────────────────────────────────────────────────────────
  function startResize(e: PointerEvent, widget: DashboardWidget) {
    if (!editable) return;
    e.preventDefault();
    e.stopPropagation();

    dragState = {
      widgetId: widget.id,
      type: 'resize',
      startX: e.clientX,
      startY: e.clientY,
      startCol: widget.col,
      startRow: widget.row,
      startColSpan: widget.colSpan,
      startRowSpan: widget.rowSpan,
    };

    const onPointerMove = (me: PointerEvent) => {
      if (!dragState || dragState.type !== 'resize') return;
      const colW = getColWidth();
      const dx = me.clientX - dragState.startX;
      const dy = me.clientY - dragState.startY;

      const newColSpan = Math.max(minWidgetColSpan, Math.min(activeCols - widget.col + 1,
        dragState.startColSpan + Math.round(dx / (colW + activeGap))
      ));
      const newRowSpan = Math.max(2,
        dragState.startRowSpan + Math.round(dy / (activeRowHeight + activeGap))
      );

      const idx = widgets.findIndex(w => w.id === widget.id);
      if (idx >= 0) {
        widgets[idx] = { ...widgets[idx], colSpan: newColSpan, rowSpan: newRowSpan };
        widgets = [...widgets];
      }
    };

    const onPointerUp = () => {
      if (dragState) {
        const w = widgets.find(w => w.id === dragState!.widgetId);
        if (w) {
          onwidgetResize?.(new CustomEvent('widgetResize', { detail: { widget: w, colSpan: w.colSpan, rowSpan: w.rowSpan } }));
        }
      }
      dragState = null;
      window.removeEventListener('pointermove', onPointerMove);
      window.removeEventListener('pointerup', onPointerUp);
    };

    window.addEventListener('pointermove', onPointerMove);
    window.addEventListener('pointerup', onPointerUp);
  }

  // ─── Widget Actions ─────────────────────────────────────────────────────────
  function selectWidget(widget: DashboardWidget) {
    selectedId = widget.id;
    onwidgetSelect?.(new CustomEvent('widgetSelect', { detail: { widget } }));
  }

  function removeWidget(widget: DashboardWidget) {
    widgets = widgets.filter(w => w.id !== widget.id);
    if (selectedId === widget.id) selectedId = null;
    onwidgetRemove?.(new CustomEvent('widgetRemove', { detail: { widget } }));
  }

  // ─── External Drop ──────────────────────────────────────────────────────────
  function handleGridDragOver(e: DragEvent) {
    if (!editable) return;
    if (!e.dataTransfer?.types.includes('application/x-bi-widget')) return;
    e.preventDefault();
    e.dataTransfer.dropEffect = 'copy';
    dragOverGrid = true;
  }

  function handleGridDragLeave(e: DragEvent) {
    const target = e.currentTarget as HTMLElement;
    const related = e.relatedTarget as HTMLElement | null;
    if (related && target.contains(related)) return;
    dragOverGrid = false;
  }

  function handleGridDrop(e: DragEvent) {
    e.preventDefault();
    dragOverGrid = false;
    if (!editable || !e.dataTransfer) return;

    const typeData = e.dataTransfer.getData('application/x-bi-widget');
    if (!typeData || !gridRef) return;

    const rect = gridRef.getBoundingClientRect();
    const colW = getColWidth();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;

    const col = Math.min(snapToGrid(x, colW), activeCols);
    const row = snapToGrid(y, activeRowHeight);

    onwidgetAdd?.(new CustomEvent('widgetAdd', { detail: { type: typeData, col, row } }));
  }

  // ─── Widget Type Icons ──────────────────────────────────────────────────────
  function getWidgetIcon(type: string): string {
    const icons: Record<string, string> = {
      chart: 'M4 20h3V10H4zM9 20h3V4H9zM14 20h3v-8h-3z',
      table: 'M3 3h18v18H3zM3 9h18M3 15h18M9 3v18',
      kpi: 'M3 12h2l3-8 4 16 3-8h6',
      map: 'M1 6l6-3 6 3 6-3v15l-6 3-6-3-6 3z',
      gauge: 'M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12M12 12l4-6',
      pivot: 'M3 3h18v18H3zM3 9h18M9 3v18',
      summary: 'M4 6h16M4 12h10M4 18h14',
    };
    return icons[type] || icons.chart;
  }
</script>

<div
  class="bi-dashboard-grid {className}"
  class:bi-dashboard-grid--editable={editable}
  class:bi-dashboard-grid--dragover={dragOverGrid}
  class:bi-dashboard-grid--mobile={viewport === 'mobile'}
  class:bi-dashboard-grid--tablet={viewport === 'tablet'}
  bind:this={gridRef}
  style="
    --grid-cols: {activeCols};
    --grid-row-height: {activeRowHeight}px;
    --grid-gap: {activeGap}px;
    --grid-rows: {maxRow};
  "
  ondragover={handleGridDragOver}
  ondragleave={handleGridDragLeave}
  ondrop={handleGridDrop}
  role="region"
  aria-label="Dashboard grid"
>
  {#each responsiveWidgets as widget (widget.id)}
    <div
      class="bi-dashboard-grid__widget"
      class:bi-dashboard-grid__widget--selected={selectedId === widget.id}
      class:bi-dashboard-grid__widget--dragging={dragState?.widgetId === widget.id}
      style="
        grid-column: {widget.col} / span {widget.colSpan};
        grid-row: {widget.row} / span {widget.rowSpan};
      "
      onclick={() => selectWidget(widget)}
      onkeydown={(e) => { if (e.key === 'Enter') selectWidget(widget); }}
      role="article"
      tabindex="0"
      aria-label="{widget.title} widget"
    >
      {#if editable}
        <div
          class="bi-dashboard-grid__drag-handle"
          onpointerdown={(e: PointerEvent) => startMove(e, widget)}
        >
          <svg viewBox="0 0 24 24" fill="currentColor" width="14" height="14">
            <circle cx="8" cy="6" r="1.5"/><circle cx="16" cy="6" r="1.5"/>
            <circle cx="8" cy="12" r="1.5"/><circle cx="16" cy="12" r="1.5"/>
            <circle cx="8" cy="18" r="1.5"/><circle cx="16" cy="18" r="1.5"/>
          </svg>
          <span class="bi-dashboard-grid__widget-title">{widget.title}</span>
          <button
            class="bi-dashboard-grid__close-btn"
            onclick={(e: MouseEvent) => { e.stopPropagation(); removeWidget(widget); }}
            aria-label="Remove {widget.title}"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
              <path d="M18 6 6 18"/><path d="m6 6 12 12"/>
            </svg>
          </button>
        </div>
      {:else}
        <div class="bi-dashboard-grid__widget-header">
          <span class="bi-dashboard-grid__widget-title">{widget.title}</span>
        </div>
      {/if}

      <div class="bi-dashboard-grid__widget-content">
        <!-- Widget content area: the parent application renders actual content here based on widget.config -->
        <div class="bi-dashboard-grid__widget-placeholder">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="24" height="24">
            <path d={getWidgetIcon(widget.type)}/>
          </svg>
          <span>{widget.type}</span>
        </div>
      </div>

      {#if editable}
        <div
          class="bi-dashboard-grid__resize-handle"
          onpointerdown={(e: PointerEvent) => startResize(e, widget)}
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="12" height="12">
            <path d="M15 3h6v6"/><path d="M9 21H3v-6"/><path d="M21 3l-7 7"/><path d="M3 21l7-7"/>
          </svg>
        </div>
      {/if}
    </div>
  {/each}
</div>

<style>
  .bi-dashboard-grid {
    display: grid;
    grid-template-columns: repeat(var(--grid-cols), 1fr);
    grid-auto-rows: var(--grid-row-height);
    gap: var(--grid-gap);
    padding: var(--grid-gap);
    min-height: calc(var(--grid-rows) * (var(--grid-row-height) + var(--grid-gap)));
    background: hsl(var(--background));
    border: 1px solid hsl(var(--border));
    border-radius: var(--radius, 0.5rem);
    position: relative;
  }

  .bi-dashboard-grid--editable {
    background-image:
      linear-gradient(hsl(var(--border) / 0.3) 1px, transparent 1px),
      linear-gradient(90deg, hsl(var(--border) / 0.3) 1px, transparent 1px);
    background-size: calc(100% / var(--grid-cols)) var(--grid-row-height);
  }

  .bi-dashboard-grid--dragover {
    outline: 2px dashed hsl(var(--primary));
    outline-offset: -2px;
  }

  .bi-dashboard-grid__widget {
    display: flex;
    flex-direction: column;
    background: hsl(var(--card));
    border: 1px solid hsl(var(--border));
    border-radius: var(--radius, 0.5rem);
    overflow: hidden;
    transition: box-shadow 0.15s ease;
    position: relative;
  }

  .bi-dashboard-grid__widget:focus-visible {
    outline: 2px solid hsl(var(--ring));
    outline-offset: 2px;
  }

  .bi-dashboard-grid__widget--selected {
    box-shadow: 0 0 0 2px hsl(var(--primary));
  }

  .bi-dashboard-grid__widget--dragging {
    opacity: 0.7;
    z-index: 10;
  }

  .bi-dashboard-grid__drag-handle {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0.375rem 0.5rem;
    cursor: grab;
    background: hsl(var(--muted));
    border-bottom: 1px solid hsl(var(--border));
    color: hsl(var(--muted-foreground));
    user-select: none;
  }

  .bi-dashboard-grid__drag-handle:active {
    cursor: grabbing;
  }

  .bi-dashboard-grid__widget-header {
    padding: 0.375rem 0.5rem;
    border-bottom: 1px solid hsl(var(--border));
    background: hsl(var(--muted));
  }

  .bi-dashboard-grid__widget-title {
    flex: 1;
    font-size: 0.75rem;
    font-weight: 600;
    color: hsl(var(--foreground));
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .bi-dashboard-grid__close-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 1.25rem;
    height: 1.25rem;
    padding: 0;
    margin-left: auto;
    border: none;
    background: transparent;
    color: hsl(var(--muted-foreground));
    cursor: pointer;
    border-radius: var(--radius, 0.25rem);
    flex-shrink: 0;
  }

  .bi-dashboard-grid__close-btn:hover {
    background: hsl(var(--destructive) / 0.15);
    color: hsl(var(--destructive));
  }

  .bi-dashboard-grid__widget-content {
    flex: 1;
    overflow: auto;
  }

  .bi-dashboard-grid__widget-placeholder {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    min-height: 4rem;
    gap: 0.25rem;
    color: hsl(var(--muted-foreground));
    font-size: 0.75rem;
    text-transform: capitalize;
  }

  .bi-dashboard-grid__resize-handle {
    position: absolute;
    bottom: 0;
    right: 0;
    width: 1.25rem;
    height: 1.25rem;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: nwse-resize;
    color: hsl(var(--muted-foreground));
    opacity: 0;
    transition: opacity 0.15s ease;
  }

  .bi-dashboard-grid__widget:hover .bi-dashboard-grid__resize-handle {
    opacity: 1;
  }

  /* ─── Mobile: stack widgets, disable drag, bigger touch targets ──────── */
  .bi-dashboard-grid--mobile {
    grid-template-columns: 1fr !important;
    grid-auto-rows: auto;
  }

  .bi-dashboard-grid--mobile .bi-dashboard-grid__widget {
    grid-column: 1 / -1 !important;
    grid-row: auto !important;
  }

  .bi-dashboard-grid--mobile .bi-dashboard-grid__drag-handle {
    padding: 0.5rem 0.75rem;
    cursor: default;
  }

  .bi-dashboard-grid--mobile .bi-dashboard-grid__resize-handle {
    display: none;
  }

  .bi-dashboard-grid--mobile .bi-dashboard-grid__widget-content {
    min-height: 12rem;
  }

  /* ─── Tablet: medium density, keep drag but enlarge handles ──────── */
  .bi-dashboard-grid--tablet .bi-dashboard-grid__drag-handle {
    padding: 0.5rem 0.625rem;
  }

  .bi-dashboard-grid--tablet .bi-dashboard-grid__resize-handle {
    width: 1.5rem;
    height: 1.5rem;
  }
</style>
