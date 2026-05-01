<script lang="ts">
  import type { FieldWellItem } from './FieldWell.types';
  import type { DatasetFieldItem } from './FieldPalette.types';

  // ─── Props ──────────────────────────────────────────────────────────────────
  interface Props {
    label: string;
    acceptRoles: string[];
    fields: FieldWellItem[];
    maxFields?: number;
    allowAggregateChange?: boolean;
    allowGranularityChange?: boolean;
    class?: string;
    onfieldDrop?: (e: CustomEvent<{ field: DatasetFieldItem; wellLabel: string }>) => void;
    onfieldRemove?: (e: CustomEvent<{ item: FieldWellItem; index: number }>) => void;
    onfieldReorder?: (e: CustomEvent<{ fields: FieldWellItem[] }>) => void;
    onaggregateChange?: (e: CustomEvent<{ item: FieldWellItem; aggregate: string }>) => void;
    ongranularityChange?: (e: CustomEvent<{ item: FieldWellItem; granularity: string }>) => void;
  }

  let {
    label,
    acceptRoles,
    fields = $bindable([]),
    maxFields = 0,
    allowAggregateChange = true,
    allowGranularityChange = true,
    class: className = '',
    onfieldDrop,
    onfieldRemove,
    onfieldReorder,
    onaggregateChange,
    ongranularityChange,
  }: Props = $props();

  // ─── State ──────────────────────────────────────────────────────────────────
  let dragOver = $state(false);
  let dragOverIndex = $state<number | null>(null);
  let internalDragIndex = $state<number | null>(null);
  let activeAggPicker = $state<string | null>(null);
  let activeGranPicker = $state<string | null>(null);

  // ─── Constants ──────────────────────────────────────────────────────────────
  const AGGREGATES = ['SUM', 'AVG', 'COUNT', 'MIN', 'MAX', 'COUNT_DISTINCT', 'MEDIAN'];
  const GRANULARITIES = ['Day', 'Week', 'Month', 'Quarter', 'Year'];

  // ─── Derived ────────────────────────────────────────────────────────────────
  let isFull = $derived(maxFields > 0 && fields.length >= maxFields);

  // ─── Drag & Drop ────────────────────────────────────────────────────────────
  function handleDragOver(e: DragEvent) {
    if (isFull) return;
    if (!e.dataTransfer) return;

    const types = e.dataTransfer.types;
    if (types.includes('application/x-bi-field') || types.includes('application/x-bi-well-item')) {
      e.preventDefault();
      e.dataTransfer.dropEffect = 'copy';
      dragOver = true;
    }
  }

  function handleDragLeave(e: DragEvent) {
    const target = e.currentTarget as HTMLElement;
    const related = e.relatedTarget as HTMLElement | null;
    if (related && target.contains(related)) return;
    dragOver = false;
    dragOverIndex = null;
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    dragOver = false;
    dragOverIndex = null;

    if (!e.dataTransfer) return;

    // Handle internal reorder
    const wellData = e.dataTransfer.getData('application/x-bi-well-item');
    if (wellData && internalDragIndex !== null) {
      // Internal reorder handled in item drop
      internalDragIndex = null;
      return;
    }

    // Handle external field drop
    const fieldData = e.dataTransfer.getData('application/x-bi-field');
    if (!fieldData) return;

    try {
      const field: DatasetFieldItem = JSON.parse(fieldData);
      if (!acceptRoles.includes(field.role)) return;
      if (isFull) return;

      onfieldDrop?.(new CustomEvent('fieldDrop', { detail: { field, wellLabel: label } }));
    } catch {
      // Invalid JSON
    }
  }

  function handleItemDragStart(e: DragEvent, index: number) {
    if (!e.dataTransfer) return;
    internalDragIndex = index;
    e.dataTransfer.setData('application/x-bi-well-item', JSON.stringify(fields[index]));
    e.dataTransfer.effectAllowed = 'move';
  }

  function handleItemDragOver(e: DragEvent, index: number) {
    if (internalDragIndex === null) return;
    e.preventDefault();
    dragOverIndex = index;
  }

  function handleItemDrop(e: DragEvent, targetIndex: number) {
    e.preventDefault();
    e.stopPropagation();

    if (internalDragIndex === null || internalDragIndex === targetIndex) {
      internalDragIndex = null;
      dragOverIndex = null;
      return;
    }

    const updated = [...fields];
    const [moved] = updated.splice(internalDragIndex, 1);
    updated.splice(targetIndex, 0, moved);
    fields = updated;

    onfieldReorder?.(new CustomEvent('fieldReorder', { detail: { fields: updated } }));
    internalDragIndex = null;
    dragOverIndex = null;
  }

  function handleRemove(index: number) {
    const item = fields[index];
    const updated = fields.filter((_, i) => i !== index);
    fields = updated;
    onfieldRemove?.(new CustomEvent('fieldRemove', { detail: { item, index } }));
  }

  function handleAggregateChange(item: FieldWellItem, agg: string) {
    activeAggPicker = null;
    onaggregateChange?.(new CustomEvent('aggregateChange', { detail: { item, aggregate: agg } }));
  }

  function handleGranularityChange(item: FieldWellItem, gran: string) {
    activeGranPicker = null;
    ongranularityChange?.(new CustomEvent('granularityChange', { detail: { item, granularity: gran } }));
  }

  function toggleAggPicker(id: string) {
    activeAggPicker = activeAggPicker === id ? null : id;
    activeGranPicker = null;
  }

  function toggleGranPicker(id: string) {
    activeGranPicker = activeGranPicker === id ? null : id;
    activeAggPicker = null;
  }
</script>

<div
  class="bi-field-well {className}"
  class:bi-field-well--dragover={dragOver}
  class:bi-field-well--full={isFull}
  role="region"
  aria-label="{label} field well"
  ondragover={handleDragOver}
  ondragleave={handleDragLeave}
  ondrop={handleDrop}
>
  <div class="bi-field-well__header">
    <span class="bi-field-well__label">{label}</span>
    {#if maxFields > 0}
      <span class="bi-field-well__count">{fields.length}/{maxFields}</span>
    {/if}
  </div>

  <div class="bi-field-well__content">
    {#if fields.length === 0}
      <div class="bi-field-well__placeholder">
        Drop {acceptRoles.join(' or ')} fields here
      </div>
    {:else}
      {#each fields as item, index (item.id)}
        <div
          class="bi-field-well__pill"
          class:bi-field-well__pill--drag-target={dragOverIndex === index}
          draggable="true"
          ondragstart={(e: DragEvent) => handleItemDragStart(e, index)}
          ondragover={(e: DragEvent) => handleItemDragOver(e, index)}
          ondrop={(e: DragEvent) => handleItemDrop(e, index)}
        >
          <span class="bi-field-well__pill-label" title={item.alias || item.label}>
            {item.alias || item.label}
          </span>

          {#if item.role === 'measure' && item.aggregate && allowAggregateChange}
            <button
              class="bi-field-well__pill-badge"
              onclick={() => toggleAggPicker(item.id)}
              title="Change aggregation"
            >
              {item.aggregate}
            </button>
            {#if activeAggPicker === item.id}
              <div class="bi-field-well__picker">
                {#each AGGREGATES as agg}
                  <button
                    class="bi-field-well__picker-item"
                    class:bi-field-well__picker-item--active={item.aggregate === agg}
                    onclick={() => handleAggregateChange(item, agg)}
                  >
                    {agg}
                  </button>
                {/each}
              </div>
            {/if}
          {/if}

          {#if (item.data_type === 'date' || item.data_type === 'datetime') && item.granularity && allowGranularityChange}
            <button
              class="bi-field-well__pill-badge bi-field-well__pill-badge--gran"
              onclick={() => toggleGranPicker(item.id)}
              title="Change granularity"
            >
              {item.granularity}
            </button>
            {#if activeGranPicker === item.id}
              <div class="bi-field-well__picker">
                {#each GRANULARITIES as gran}
                  <button
                    class="bi-field-well__picker-item"
                    class:bi-field-well__picker-item--active={item.granularity === gran}
                    onclick={() => handleGranularityChange(item, gran)}
                  >
                    {gran}
                  </button>
                {/each}
              </div>
            {/if}
          {/if}

          <button
            class="bi-field-well__pill-remove"
            onclick={() => handleRemove(index)}
            aria-label="Remove {item.label}"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="12" height="12">
              <path d="M18 6 6 18"/><path d="m6 6 12 12"/>
            </svg>
          </button>
        </div>
      {/each}
    {/if}
  </div>
</div>

<style>
  .bi-field-well {
    display: flex;
    flex-direction: column;
    border: 2px dashed hsl(var(--border));
    border-radius: var(--radius, 0.5rem);
    background: hsl(var(--background));
    transition: border-color 0.15s ease, background 0.15s ease;
    min-height: 3.5rem;
  }

  .bi-field-well--dragover {
    border-color: hsl(var(--primary));
    background: hsl(var(--primary) / 0.05);
  }

  .bi-field-well--full {
    border-style: solid;
    opacity: 0.7;
  }

  .bi-field-well__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.375rem 0.625rem;
    border-bottom: 1px solid hsl(var(--border));
  }

  .bi-field-well__label {
    font-size: 0.6875rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: hsl(var(--muted-foreground));
  }

  .bi-field-well__count {
    font-size: 0.625rem;
    color: hsl(var(--muted-foreground));
    font-family: monospace;
  }

  .bi-field-well__content {
    display: flex;
    flex-wrap: wrap;
    gap: 0.375rem;
    padding: 0.5rem;
    min-height: 2rem;
  }

  .bi-field-well__placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    font-size: 0.75rem;
    color: hsl(var(--muted-foreground));
    padding: 0.25rem 0;
  }

  .bi-field-well__pill {
    position: relative;
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.1875rem 0.25rem 0.1875rem 0.5rem;
    background: hsl(var(--secondary));
    color: hsl(var(--secondary-foreground));
    border-radius: 9999px;
    font-size: 0.75rem;
    cursor: grab;
    user-select: none;
    transition: background 0.1s ease;
  }

  .bi-field-well__pill:hover {
    background: hsl(var(--accent));
  }

  .bi-field-well__pill:active {
    cursor: grabbing;
  }

  .bi-field-well__pill--drag-target {
    box-shadow: inset 0 0 0 2px hsl(var(--primary));
  }

  .bi-field-well__pill-label {
    max-width: 8rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .bi-field-well__pill-badge {
    display: inline-flex;
    align-items: center;
    padding: 0 0.25rem;
    border: none;
    background: hsl(var(--primary) / 0.15);
    color: hsl(var(--primary));
    font-size: 0.5625rem;
    font-weight: 600;
    border-radius: var(--radius, 0.25rem);
    cursor: pointer;
    line-height: 1.5;
  }

  .bi-field-well__pill-badge--gran {
    background: hsl(var(--chart-2, 220 70% 50%) / 0.15);
    color: hsl(var(--chart-2, 220 70% 50%));
  }

  .bi-field-well__pill-badge:hover {
    opacity: 0.8;
  }

  .bi-field-well__picker {
    position: absolute;
    top: 100%;
    left: 0;
    z-index: 50;
    display: flex;
    flex-direction: column;
    min-width: 8rem;
    margin-top: 0.25rem;
    padding: 0.25rem;
    background: hsl(var(--popover));
    border: 1px solid hsl(var(--border));
    border-radius: var(--radius, 0.5rem);
    box-shadow: 0 4px 12px hsl(var(--foreground) / 0.1);
  }

  .bi-field-well__picker-item {
    display: block;
    width: 100%;
    text-align: left;
    padding: 0.375rem 0.5rem;
    border: none;
    background: transparent;
    font-size: 0.75rem;
    color: hsl(var(--popover-foreground));
    cursor: pointer;
    border-radius: var(--radius, 0.25rem);
  }

  .bi-field-well__picker-item:hover {
    background: hsl(var(--accent));
  }

  .bi-field-well__picker-item--active {
    background: hsl(var(--primary) / 0.1);
    color: hsl(var(--primary));
    font-weight: 600;
  }

  .bi-field-well__pill-remove {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 1.125rem;
    height: 1.125rem;
    padding: 0;
    border: none;
    background: transparent;
    color: hsl(var(--muted-foreground));
    cursor: pointer;
    border-radius: 9999px;
  }

  .bi-field-well__pill-remove:hover {
    background: hsl(var(--destructive) / 0.15);
    color: hsl(var(--destructive));
  }
</style>
