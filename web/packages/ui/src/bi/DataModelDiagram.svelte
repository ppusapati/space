<script lang="ts">
  import type { DatasetNode, RelationshipEdge } from './DataModelDiagram.types';

  // ─── Props ──────────────────────────────────────────────────────────────────
  interface Props {
    datasets: DatasetNode[];
    relationships: RelationshipEdge[];
    editable?: boolean;
    selectedDatasetId?: string;
    class?: string;
    ondatasetSelect?: (e: CustomEvent<{ dataset: DatasetNode }>) => void;
    onrelationshipCreate?: (e: CustomEvent<{ sourceDatasetId: string; sourceFieldId: string; targetDatasetId: string; targetFieldId: string }>) => void;
    onrelationshipRemove?: (e: CustomEvent<{ relationship: RelationshipEdge }>) => void;
  }

  let {
    datasets = $bindable([]),
    relationships,
    editable = false,
    selectedDatasetId = $bindable(undefined),
    class: className = '',
    ondatasetSelect,
    onrelationshipCreate,
    onrelationshipRemove,
  }: Props = $props();

  // ─── Constants ──────────────────────────────────────────────────────────────
  const NODE_WIDTH = 200;
  const NODE_HEADER_HEIGHT = 36;
  const NODE_FIELD_HEIGHT = 24;
  const NODE_PADDING = 8;

  // ─── State ──────────────────────────────────────────────────────────────────
  let svgRef: SVGSVGElement | undefined = $state(undefined);
  let viewBox = $state({ x: 0, y: 0, width: 1200, height: 800 });
  let isPanning = $state(false);
  let panStart = $state({ x: 0, y: 0 });
  let dragNode = $state<{ id: string; offsetX: number; offsetY: number } | null>(null);
  let joinDrag = $state<{ sourceDatasetId: string; sourceFieldId: string; x: number; y: number } | null>(null);

  // ─── Derived ────────────────────────────────────────────────────────────────
  let nodeHeights = $derived.by(() => {
    const map = new Map<string, number>();
    for (const ds of datasets) {
      const fieldCount = ds.fields?.length || 0;
      const h = NODE_HEADER_HEIGHT + NODE_PADDING * 2 + Math.max(1, fieldCount) * NODE_FIELD_HEIGHT;
      map.set(ds.id, h);
    }
    return map;
  });

  // ─── Helpers ────────────────────────────────────────────────────────────────
  function getSvgPoint(e: MouseEvent): { x: number; y: number } {
    if (!svgRef) return { x: 0, y: 0 };
    const rect = svgRef.getBoundingClientRect();
    const scaleX = viewBox.width / rect.width;
    const scaleY = viewBox.height / rect.height;
    return {
      x: viewBox.x + (e.clientX - rect.left) * scaleX,
      y: viewBox.y + (e.clientY - rect.top) * scaleY,
    };
  }

  function getNodeCenter(ds: DatasetNode): { x: number; y: number } {
    const h = nodeHeights.get(ds.id) || 80;
    return { x: ds.x + NODE_WIDTH / 2, y: ds.y + h / 2 };
  }

  function getFieldY(ds: DatasetNode, fieldId: string): number {
    const fieldIdx = ds.fields?.findIndex(f => f.id === fieldId) ?? 0;
    return ds.y + NODE_HEADER_HEIGHT + NODE_PADDING + fieldIdx * NODE_FIELD_HEIGHT + NODE_FIELD_HEIGHT / 2;
  }

  function getEdgePath(rel: RelationshipEdge): string {
    const source = datasets.find(d => d.id === rel.sourceDatasetId);
    const target = datasets.find(d => d.id === rel.targetDatasetId);
    if (!source || !target) return '';

    const sy = getFieldY(source, rel.sourceFieldId);
    const ty = getFieldY(target, rel.targetFieldId);
    const sx = source.x + NODE_WIDTH;
    const tx = target.x;

    const midX = (sx + tx) / 2;
    return `M ${sx} ${sy} C ${midX} ${sy}, ${midX} ${ty}, ${tx} ${ty}`;
  }

  function getCardinalityLabel(cardinality: string): string {
    const labels: Record<string, string> = {
      '1:1': '1:1',
      '1:N': '1:N',
      'N:1': 'N:1',
      'M:N': 'M:N',
      'one-to-one': '1:1',
      'one-to-many': '1:N',
      'many-to-one': 'N:1',
      'many-to-many': 'M:N',
    };
    return labels[cardinality] || cardinality;
  }

  function getRoleColor(role: string): string {
    switch (role) {
      case 'dimension': return 'hsl(var(--primary))';
      case 'measure': return 'hsl(var(--chart-2, 220 70% 50%))';
      default: return 'hsl(var(--muted-foreground))';
    }
  }

  // ─── Pan & Zoom ─────────────────────────────────────────────────────────────
  function handleWheel(e: WheelEvent) {
    e.preventDefault();
    const factor = e.deltaY > 0 ? 1.1 : 0.9;
    const point = getSvgPoint(e);

    viewBox = {
      x: point.x - (point.x - viewBox.x) * factor,
      y: point.y - (point.y - viewBox.y) * factor,
      width: viewBox.width * factor,
      height: viewBox.height * factor,
    };
  }

  function handleBackgroundPointerDown(e: PointerEvent) {
    if ((e.target as Element).closest('.bi-dm__node')) return;
    isPanning = true;
    panStart = { x: e.clientX, y: e.clientY };
    selectedDatasetId = undefined;
  }

  function handlePointerMove(e: PointerEvent) {
    // Pan
    if (isPanning && svgRef) {
      const rect = svgRef.getBoundingClientRect();
      const scaleX = viewBox.width / rect.width;
      const scaleY = viewBox.height / rect.height;
      const dx = (e.clientX - panStart.x) * scaleX;
      const dy = (e.clientY - panStart.y) * scaleY;
      viewBox = { ...viewBox, x: viewBox.x - dx, y: viewBox.y - dy };
      panStart = { x: e.clientX, y: e.clientY };
      return;
    }

    // Node drag
    if (dragNode) {
      const pt = getSvgPoint(e);
      const idx = datasets.findIndex(d => d.id === dragNode!.id);
      if (idx >= 0) {
        datasets[idx] = { ...datasets[idx], x: pt.x - dragNode.offsetX, y: pt.y - dragNode.offsetY };
        datasets = [...datasets];
      }
      return;
    }

    // Join drag line
    if (joinDrag) {
      const pt = getSvgPoint(e);
      joinDrag = { ...joinDrag, x: pt.x, y: pt.y };
    }
  }

  function handlePointerUp() {
    isPanning = false;
    dragNode = null;

    if (joinDrag) {
      // Check if dropped on a field
      joinDrag = null;
    }
  }

  function handleNodePointerDown(e: PointerEvent, ds: DatasetNode) {
    e.stopPropagation();
    const pt = getSvgPoint(e);
    dragNode = { id: ds.id, offsetX: pt.x - ds.x, offsetY: pt.y - ds.y };
    selectedDatasetId = ds.id;
    ondatasetSelect?.(new CustomEvent('datasetSelect', { detail: { dataset: ds } }));
  }

  function handleNodeClick(ds: DatasetNode) {
    selectedDatasetId = ds.id;
    ondatasetSelect?.(new CustomEvent('datasetSelect', { detail: { dataset: ds } }));
  }

  // ─── Join Creation ──────────────────────────────────────────────────────────
  function handleFieldPointerDown(e: PointerEvent, dsId: string, fieldId: string) {
    if (!editable) return;
    e.stopPropagation();
    const pt = getSvgPoint(e);
    joinDrag = { sourceDatasetId: dsId, sourceFieldId: fieldId, x: pt.x, y: pt.y };
  }

  function handleFieldPointerUp(e: PointerEvent, dsId: string, fieldId: string) {
    if (!joinDrag) return;
    if (joinDrag.sourceDatasetId === dsId) {
      joinDrag = null;
      return;
    }

    onrelationshipCreate?.(new CustomEvent('relationshipCreate', {
      detail: {
        sourceDatasetId: joinDrag.sourceDatasetId,
        sourceFieldId: joinDrag.sourceFieldId,
        targetDatasetId: dsId,
        targetFieldId: fieldId,
      }
    }));
    joinDrag = null;
  }

  function handleRelationshipClick(rel: RelationshipEdge) {
    if (!editable) return;
    onrelationshipRemove?.(new CustomEvent('relationshipRemove', { detail: { relationship: rel } }));
  }
</script>

<div class="bi-data-model {className}">
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <svg
    bind:this={svgRef}
    class="bi-dm__svg"
    viewBox="{viewBox.x} {viewBox.y} {viewBox.width} {viewBox.height}"
    onwheel={handleWheel}
    onpointerdown={handleBackgroundPointerDown}
    onpointermove={handlePointerMove}
    onpointerup={handlePointerUp}
  >
    <!-- Relationship edges -->
    {#each relationships as rel (rel.id)}
      {@const path = getEdgePath(rel)}
      {#if path}
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
        <g
          class="bi-dm__edge"
          class:bi-dm__edge--editable={editable}
          onclick={() => handleRelationshipClick(rel)}
        >
          <path d={path} class="bi-dm__edge-hitbox" />
          <path d={path} class="bi-dm__edge-line" />
          <!-- Cardinality label at midpoint -->
          {#if true}
            {@const source = datasets.find(d => d.id === rel.sourceDatasetId)}
            {@const target = datasets.find(d => d.id === rel.targetDatasetId)}
            {#if source && target}
              {@const mx = (source.x + NODE_WIDTH + target.x) / 2}
              {@const my = (getFieldY(source, rel.sourceFieldId) + getFieldY(target, rel.targetFieldId)) / 2}
              <rect x={mx - 16} y={my - 10} width="32" height="20" rx="4" class="bi-dm__edge-label-bg" />
              <text x={mx} y={my + 4} class="bi-dm__edge-label" text-anchor="middle">
                {getCardinalityLabel(rel.cardinality)}
              </text>
            {/if}
          {/if}
        </g>
      {/if}
    {/each}

    <!-- Join drag line -->
    {#if joinDrag}
      {@const source = datasets.find(d => d.id === joinDrag.sourceDatasetId)}
      {#if source}
        <line
          x1={source.x + NODE_WIDTH}
          y1={getFieldY(source, joinDrag.sourceFieldId)}
          x2={joinDrag.x}
          y2={joinDrag.y}
          class="bi-dm__join-line"
        />
      {/if}
    {/if}

    <!-- Dataset nodes -->
    {#each datasets as ds (ds.id)}
      {@const nh = nodeHeights.get(ds.id) || 80}
      <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
      <g
        class="bi-dm__node"
        class:bi-dm__node--selected={selectedDatasetId === ds.id}
        transform="translate({ds.x}, {ds.y})"
        onpointerdown={(e: PointerEvent) => handleNodePointerDown(e, ds)}
        onclick={() => handleNodeClick(ds)}
      >
        <!-- Card background -->
        <rect
          width={NODE_WIDTH}
          height={nh}
          rx="8"
          class="bi-dm__node-bg"
        />

        <!-- Header -->
        <rect
          width={NODE_WIDTH}
          height={NODE_HEADER_HEIGHT}
          rx="8"
          class="bi-dm__node-header"
        />
        <rect
          x="0"
          y={NODE_HEADER_HEIGHT - 8}
          width={NODE_WIDTH}
          height="8"
          class="bi-dm__node-header"
        />

        <text x="12" y="23" class="bi-dm__node-name">{ds.name}</text>
        <text x={NODE_WIDTH - 12} y="23" class="bi-dm__node-module" text-anchor="end">{ds.module}</text>

        <!-- Fields -->
        {#if ds.fields && ds.fields.length > 0}
          {#each ds.fields as field, fi (field.id)}
            {@const fy = NODE_HEADER_HEIGHT + NODE_PADDING + fi * NODE_FIELD_HEIGHT}
            <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
            <g
              class="bi-dm__field"
              class:bi-dm__field--editable={editable}
              onpointerdown={(e: PointerEvent) => handleFieldPointerDown(e, ds.id, field.id)}
              onpointerup={(e: PointerEvent) => handleFieldPointerUp(e, ds.id, field.id)}
            >
              <rect
                x="4"
                y={fy}
                width={NODE_WIDTH - 8}
                height={NODE_FIELD_HEIGHT}
                rx="4"
                class="bi-dm__field-bg"
              />
              <circle
                cx="16"
                cy={fy + NODE_FIELD_HEIGHT / 2}
                r="3"
                fill={getRoleColor(field.role)}
              />
              <text x="26" y={fy + NODE_FIELD_HEIGHT / 2 + 4} class="bi-dm__field-label">
                {field.label}
              </text>
            </g>
          {/each}
        {:else}
          <text x={NODE_WIDTH / 2} y={NODE_HEADER_HEIGHT + NODE_PADDING + 14} class="bi-dm__field-empty" text-anchor="middle">
            {ds.fieldCount} fields
          </text>
        {/if}
      </g>
    {/each}
  </svg>
</div>

<style>
  .bi-data-model {
    width: 100%;
    height: 100%;
    background: hsl(var(--background));
    border: 1px solid hsl(var(--border));
    border-radius: var(--radius, 0.5rem);
    overflow: hidden;
  }

  .bi-dm__svg {
    width: 100%;
    height: 100%;
    cursor: grab;
  }

  .bi-dm__svg:active {
    cursor: grabbing;
  }

  /* Edges */
  .bi-dm__edge-hitbox {
    fill: none;
    stroke: transparent;
    stroke-width: 12;
    cursor: default;
  }

  .bi-dm__edge--editable .bi-dm__edge-hitbox {
    cursor: pointer;
  }

  .bi-dm__edge-line {
    fill: none;
    stroke: hsl(var(--border));
    stroke-width: 2;
    pointer-events: none;
  }

  .bi-dm__edge:hover .bi-dm__edge-line {
    stroke: hsl(var(--primary));
    stroke-width: 2.5;
  }

  .bi-dm__edge-label-bg {
    fill: hsl(var(--background));
    stroke: hsl(var(--border));
    stroke-width: 1;
  }

  .bi-dm__edge-label {
    fill: hsl(var(--muted-foreground));
    font-size: 10px;
    font-weight: 600;
    font-family: monospace;
  }

  .bi-dm__join-line {
    stroke: hsl(var(--primary));
    stroke-width: 2;
    stroke-dasharray: 6 3;
    pointer-events: none;
  }

  /* Nodes */
  .bi-dm__node {
    cursor: pointer;
  }

  .bi-dm__node-bg {
    fill: hsl(var(--card));
    stroke: hsl(var(--border));
    stroke-width: 1.5;
    transition: stroke 0.15s ease;
  }

  .bi-dm__node--selected .bi-dm__node-bg {
    stroke: hsl(var(--primary));
    stroke-width: 2.5;
  }

  .bi-dm__node-header {
    fill: hsl(var(--muted));
  }

  .bi-dm__node-name {
    fill: hsl(var(--foreground));
    font-size: 12px;
    font-weight: 700;
  }

  .bi-dm__node-module {
    fill: hsl(var(--muted-foreground));
    font-size: 9px;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  /* Fields */
  .bi-dm__field-bg {
    fill: transparent;
  }

  .bi-dm__field:hover .bi-dm__field-bg {
    fill: hsl(var(--accent) / 0.5);
  }

  .bi-dm__field--editable {
    cursor: crosshair;
  }

  .bi-dm__field-label {
    fill: hsl(var(--foreground));
    font-size: 11px;
  }

  .bi-dm__field-empty {
    fill: hsl(var(--muted-foreground));
    font-size: 11px;
    font-style: italic;
  }
</style>
