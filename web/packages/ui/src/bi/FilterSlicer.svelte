<script lang="ts">
  import type { SlicerFilter } from './FilterSlicer.types';

  // ─── Props ──────────────────────────────────────────────────────────────────
  interface FilterableField {
    id: string;
    label: string;
    data_type: string;
    values?: string[];
  }

  interface Props {
    filters: SlicerFilter[];
    fields: FilterableField[];
    orientation?: 'horizontal' | 'vertical';
    class?: string;
    onfilterChange?: (e: CustomEvent<{ filters: SlicerFilter[] }>) => void;
  }

  let {
    filters = $bindable([]),
    fields,
    orientation = 'horizontal',
    class: className = '',
    onfilterChange,
  }: Props = $props();

  // ─── State ──────────────────────────────────────────────────────────────────
  let showAddPicker = $state(false);
  let expandedFilter = $state<string | null>(null);

  // ─── Derived ────────────────────────────────────────────────────────────────
  let availableFields = $derived(
    fields.filter(f => !filters.some(fl => fl.field_id === f.id))
  );

  // ─── Helpers ────────────────────────────────────────────────────────────────
  function addFilter(field: FilterableField) {
    const newFilter: SlicerFilter = {
      field_id: field.id,
      label: field.label,
      data_type: field.data_type,
      operator: getDefaultOperator(field.data_type),
      value: null,
      values: [],
    };
    filters = [...filters, newFilter];
    expandedFilter = field.id;
    showAddPicker = false;
    emitChange();
  }

  function getDefaultOperator(dataType: string): string {
    switch (dataType) {
      case 'string': return 'in';
      case 'integer':
      case 'decimal':
      case 'currency':
      case 'percentage': return 'between';
      case 'date':
      case 'datetime': return 'between';
      case 'boolean': return 'eq';
      default: return 'eq';
    }
  }

  function removeFilter(fieldId: string) {
    filters = filters.filter(f => f.field_id !== fieldId);
    if (expandedFilter === fieldId) expandedFilter = null;
    emitChange();
  }

  function clearAll() {
    filters = [];
    expandedFilter = null;
    emitChange();
  }

  function toggleExpanded(fieldId: string) {
    expandedFilter = expandedFilter === fieldId ? null : fieldId;
  }

  function emitChange() {
    onfilterChange?.(new CustomEvent('filterChange', { detail: { filters: [...filters] } }));
  }

  // ─── String/Enum Filter ─────────────────────────────────────────────────────
  function toggleValue(filter: SlicerFilter, val: string) {
    const idx = filters.findIndex(f => f.field_id === filter.field_id);
    if (idx < 0) return;

    const current = (filter.values || []) as string[];
    const valueSet = new Set(current);
    if (valueSet.has(val)) valueSet.delete(val);
    else valueSet.add(val);

    filters[idx] = { ...filter, values: [...valueSet], operator: 'in' };
    filters = [...filters];
    emitChange();
  }

  function isValueSelected(filter: SlicerFilter, val: string): boolean {
    return ((filter.values || []) as string[]).includes(val);
  }

  // ─── Number Range Filter ───────────────────────────────────────────────────
  function updateRangeMin(filter: SlicerFilter, val: string) {
    const idx = filters.findIndex(f => f.field_id === filter.field_id);
    if (idx < 0) return;
    const numVal = val === '' ? null : Number(val);
    filters[idx] = { ...filter, value: numVal, operator: 'between' };
    filters = [...filters];
    emitChange();
  }

  function updateRangeMax(filter: SlicerFilter, val: string) {
    const idx = filters.findIndex(f => f.field_id === filter.field_id);
    if (idx < 0) return;
    const numVal = val === '' ? null : Number(val);
    filters[idx] = { ...filter, second_value: numVal, operator: 'between' };
    filters = [...filters];
    emitChange();
  }

  // ─── Date Range Filter ─────────────────────────────────────────────────────
  function updateDateStart(filter: SlicerFilter, val: string) {
    const idx = filters.findIndex(f => f.field_id === filter.field_id);
    if (idx < 0) return;
    filters[idx] = { ...filter, value: val || null, operator: 'between' };
    filters = [...filters];
    emitChange();
  }

  function updateDateEnd(filter: SlicerFilter, val: string) {
    const idx = filters.findIndex(f => f.field_id === filter.field_id);
    if (idx < 0) return;
    filters[idx] = { ...filter, second_value: val || null, operator: 'between' };
    filters = [...filters];
    emitChange();
  }

  // ─── Boolean Filter ────────────────────────────────────────────────────────
  function toggleBoolean(filter: SlicerFilter) {
    const idx = filters.findIndex(f => f.field_id === filter.field_id);
    if (idx < 0) return;
    filters[idx] = { ...filter, value: !filter.value, operator: 'eq' };
    filters = [...filters];
    emitChange();
  }

  function getFieldValues(fieldId: string): string[] {
    return fields.find(f => f.id === fieldId)?.values || [];
  }

  function getFilterSummary(filter: SlicerFilter): string {
    if (filter.data_type === 'boolean') {
      return filter.value ? 'Yes' : 'No';
    }
    if (filter.operator === 'in' && filter.values) {
      const vals = filter.values as string[];
      if (vals.length === 0) return 'All';
      if (vals.length <= 2) return vals.join(', ');
      return `${vals.length} selected`;
    }
    if (filter.operator === 'between') {
      const from = filter.value != null ? String(filter.value) : '';
      const to = filter.second_value != null ? String(filter.second_value) : '';
      if (from && to) return `${from} - ${to}`;
      if (from) return `>= ${from}`;
      if (to) return `<= ${to}`;
      return 'Any';
    }
    return filter.value != null ? String(filter.value) : 'Any';
  }
</script>

<div
  class="bi-filter-slicer {className}"
  class:bi-filter-slicer--vertical={orientation === 'vertical'}
  role="region"
  aria-label="Filters"
>
  <div class="bi-filter-slicer__bar">
    <svg class="bi-filter-slicer__icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
      <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"/>
    </svg>

    {#each filters as filter (filter.field_id)}
      <div class="bi-filter-slicer__chip" class:bi-filter-slicer__chip--expanded={expandedFilter === filter.field_id}>
        <button
          class="bi-filter-slicer__chip-btn"
          onclick={() => toggleExpanded(filter.field_id)}
        >
          <span class="bi-filter-slicer__chip-label">{filter.label}</span>
          <span class="bi-filter-slicer__chip-value">{getFilterSummary(filter)}</span>
        </button>
        <button
          class="bi-filter-slicer__chip-remove"
          onclick={() => removeFilter(filter.field_id)}
          aria-label="Remove {filter.label} filter"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="12" height="12">
            <path d="M18 6 6 18"/><path d="m6 6 12 12"/>
          </svg>
        </button>

        <!-- Expanded filter panel -->
        {#if expandedFilter === filter.field_id}
          <div class="bi-filter-slicer__panel">
            {#if filter.data_type === 'string'}
              <!-- Multi-select checkboxes -->
              {@const fieldValues = getFieldValues(filter.field_id)}
              {#if fieldValues.length > 0}
                <div class="bi-filter-slicer__check-list">
                  {#each fieldValues as val}
                    <label class="bi-filter-slicer__check-item">
                      <input
                        type="checkbox"
                        checked={isValueSelected(filter, val)}
                        onchange={() => toggleValue(filter, val)}
                      />
                      <span>{val}</span>
                    </label>
                  {/each}
                </div>
              {:else}
                <span class="bi-filter-slicer__empty">No values available</span>
              {/if}

            {:else if filter.data_type === 'integer' || filter.data_type === 'decimal' || filter.data_type === 'currency' || filter.data_type === 'percentage'}
              <!-- Number range -->
              <div class="bi-filter-slicer__range">
                <label class="bi-filter-slicer__range-field">
                  <span>Min</span>
                  <input
                    type="number"
                    value={filter.value ?? ''}
                    oninput={(e: Event) => updateRangeMin(filter, (e.target as HTMLInputElement).value)}
                    placeholder="Min"
                  />
                </label>
                <span class="bi-filter-slicer__range-sep">-</span>
                <label class="bi-filter-slicer__range-field">
                  <span>Max</span>
                  <input
                    type="number"
                    value={filter.second_value ?? ''}
                    oninput={(e: Event) => updateRangeMax(filter, (e.target as HTMLInputElement).value)}
                    placeholder="Max"
                  />
                </label>
              </div>

            {:else if filter.data_type === 'date' || filter.data_type === 'datetime'}
              <!-- Date range -->
              <div class="bi-filter-slicer__range">
                <label class="bi-filter-slicer__range-field">
                  <span>From</span>
                  <input
                    type="date"
                    value={filter.value ?? ''}
                    oninput={(e: Event) => updateDateStart(filter, (e.target as HTMLInputElement).value)}
                  />
                </label>
                <span class="bi-filter-slicer__range-sep">-</span>
                <label class="bi-filter-slicer__range-field">
                  <span>To</span>
                  <input
                    type="date"
                    value={filter.second_value ?? ''}
                    oninput={(e: Event) => updateDateEnd(filter, (e.target as HTMLInputElement).value)}
                  />
                </label>
              </div>

            {:else if filter.data_type === 'boolean'}
              <!-- Boolean toggle -->
              <div class="bi-filter-slicer__toggle">
                <button
                  class="bi-filter-slicer__toggle-btn"
                  class:bi-filter-slicer__toggle-btn--on={!!filter.value}
                  onclick={() => toggleBoolean(filter)}
                  role="switch"
                  aria-checked={!!filter.value}
                >
                  <span class="bi-filter-slicer__toggle-track">
                    <span class="bi-filter-slicer__toggle-thumb"></span>
                  </span>
                  <span>{filter.value ? 'Yes' : 'No'}</span>
                </button>
              </div>
            {/if}
          </div>
        {/if}
      </div>
    {/each}

    <!-- Add filter button -->
    <div class="bi-filter-slicer__add-wrap">
      <button
        class="bi-filter-slicer__add-btn"
        onclick={() => showAddPicker = !showAddPicker}
        disabled={availableFields.length === 0}
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
          <path d="M12 5v14"/><path d="M5 12h14"/>
        </svg>
        Add Filter
      </button>

      {#if showAddPicker}
        <div class="bi-filter-slicer__add-picker">
          {#each availableFields as field (field.id)}
            <button
              class="bi-filter-slicer__add-item"
              onclick={() => addFilter(field)}
            >
              <span>{field.label}</span>
              <span class="bi-filter-slicer__add-type">{field.data_type}</span>
            </button>
          {/each}
        </div>
      {/if}
    </div>

    {#if filters.length > 0}
      <button
        class="bi-filter-slicer__clear"
        onclick={clearAll}
      >
        Clear All
      </button>
    {/if}
  </div>
</div>

<style>
  .bi-filter-slicer {
    background: hsl(var(--background));
    border: 1px solid hsl(var(--border));
    border-radius: var(--radius, 0.5rem);
  }

  .bi-filter-slicer__bar {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.375rem;
    padding: 0.5rem 0.75rem;
  }

  .bi-filter-slicer--vertical .bi-filter-slicer__bar {
    flex-direction: column;
    align-items: stretch;
  }

  .bi-filter-slicer__icon {
    flex-shrink: 0;
    color: hsl(var(--muted-foreground));
  }

  /* Chips */
  .bi-filter-slicer__chip {
    position: relative;
    display: flex;
    align-items: center;
    gap: 0.125rem;
    background: hsl(var(--secondary));
    border-radius: 9999px;
    overflow: visible;
  }

  .bi-filter-slicer__chip--expanded {
    z-index: 10;
  }

  .bi-filter-slicer__chip-btn {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem 0.5rem;
    border: none;
    background: transparent;
    cursor: pointer;
    font-size: 0.75rem;
    color: hsl(var(--secondary-foreground));
    border-radius: 9999px 0 0 9999px;
  }

  .bi-filter-slicer__chip-btn:hover {
    background: hsl(var(--accent) / 0.5);
  }

  .bi-filter-slicer__chip-label {
    font-weight: 600;
  }

  .bi-filter-slicer__chip-value {
    color: hsl(var(--muted-foreground));
    max-width: 8rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .bi-filter-slicer__chip-remove {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 1.375rem;
    height: 1.375rem;
    padding: 0;
    border: none;
    background: transparent;
    color: hsl(var(--muted-foreground));
    cursor: pointer;
    border-radius: 9999px;
  }

  .bi-filter-slicer__chip-remove:hover {
    background: hsl(var(--destructive) / 0.15);
    color: hsl(var(--destructive));
  }

  /* Filter panel dropdown */
  .bi-filter-slicer__panel {
    position: absolute;
    top: 100%;
    left: 0;
    z-index: 50;
    min-width: 14rem;
    margin-top: 0.25rem;
    padding: 0.5rem;
    background: hsl(var(--popover));
    border: 1px solid hsl(var(--border));
    border-radius: var(--radius, 0.5rem);
    box-shadow: 0 4px 16px hsl(var(--foreground) / 0.1);
  }

  /* Checkbox list */
  .bi-filter-slicer__check-list {
    display: flex;
    flex-direction: column;
    max-height: 12rem;
    overflow-y: auto;
  }

  .bi-filter-slicer__check-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.25rem 0.375rem;
    font-size: 0.8125rem;
    color: hsl(var(--popover-foreground));
    cursor: pointer;
    border-radius: var(--radius, 0.25rem);
  }

  .bi-filter-slicer__check-item:hover {
    background: hsl(var(--accent));
  }

  .bi-filter-slicer__check-item input[type="checkbox"] {
    accent-color: hsl(var(--primary));
  }

  .bi-filter-slicer__empty {
    font-size: 0.75rem;
    color: hsl(var(--muted-foreground));
    padding: 0.5rem;
    text-align: center;
  }

  /* Range inputs */
  .bi-filter-slicer__range {
    display: flex;
    align-items: flex-end;
    gap: 0.375rem;
  }

  .bi-filter-slicer__range-field {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
    flex: 1;
  }

  .bi-filter-slicer__range-field span {
    font-size: 0.6875rem;
    font-weight: 500;
    color: hsl(var(--muted-foreground));
  }

  .bi-filter-slicer__range-field input {
    width: 100%;
    padding: 0.375rem 0.5rem;
    border: 1px solid hsl(var(--border));
    border-radius: var(--radius, 0.25rem);
    background: hsl(var(--background));
    color: hsl(var(--foreground));
    font-size: 0.8125rem;
    outline: none;
  }

  .bi-filter-slicer__range-field input:focus {
    border-color: hsl(var(--ring));
    box-shadow: 0 0 0 2px hsl(var(--ring) / 0.2);
  }

  .bi-filter-slicer__range-sep {
    color: hsl(var(--muted-foreground));
    padding-bottom: 0.375rem;
  }

  /* Boolean toggle */
  .bi-filter-slicer__toggle {
    display: flex;
    justify-content: center;
    padding: 0.25rem;
  }

  .bi-filter-slicer__toggle-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.25rem;
    border: none;
    background: transparent;
    cursor: pointer;
    font-size: 0.8125rem;
    color: hsl(var(--popover-foreground));
  }

  .bi-filter-slicer__toggle-track {
    position: relative;
    width: 2.25rem;
    height: 1.25rem;
    background: hsl(var(--muted));
    border-radius: 9999px;
    transition: background 0.2s ease;
  }

  .bi-filter-slicer__toggle-btn--on .bi-filter-slicer__toggle-track {
    background: hsl(var(--primary));
  }

  .bi-filter-slicer__toggle-thumb {
    position: absolute;
    top: 0.125rem;
    left: 0.125rem;
    width: 1rem;
    height: 1rem;
    background: white;
    border-radius: 9999px;
    transition: transform 0.2s ease;
  }

  .bi-filter-slicer__toggle-btn--on .bi-filter-slicer__toggle-thumb {
    transform: translateX(1rem);
  }

  /* Add filter */
  .bi-filter-slicer__add-wrap {
    position: relative;
  }

  .bi-filter-slicer__add-btn {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem 0.5rem;
    border: 1px dashed hsl(var(--border));
    background: transparent;
    color: hsl(var(--muted-foreground));
    font-size: 0.75rem;
    cursor: pointer;
    border-radius: 9999px;
    white-space: nowrap;
  }

  .bi-filter-slicer__add-btn:hover:not(:disabled) {
    border-color: hsl(var(--primary));
    color: hsl(var(--primary));
  }

  .bi-filter-slicer__add-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .bi-filter-slicer__add-picker {
    position: absolute;
    top: 100%;
    left: 0;
    z-index: 50;
    min-width: 12rem;
    max-height: 16rem;
    overflow-y: auto;
    margin-top: 0.25rem;
    padding: 0.25rem;
    background: hsl(var(--popover));
    border: 1px solid hsl(var(--border));
    border-radius: var(--radius, 0.5rem);
    box-shadow: 0 4px 12px hsl(var(--foreground) / 0.1);
  }

  .bi-filter-slicer__add-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 0.375rem 0.5rem;
    border: none;
    background: transparent;
    cursor: pointer;
    font-size: 0.8125rem;
    color: hsl(var(--popover-foreground));
    border-radius: var(--radius, 0.25rem);
    text-align: left;
  }

  .bi-filter-slicer__add-item:hover {
    background: hsl(var(--accent));
  }

  .bi-filter-slicer__add-type {
    font-size: 0.625rem;
    color: hsl(var(--muted-foreground));
    font-family: monospace;
  }

  /* Clear all */
  .bi-filter-slicer__clear {
    margin-left: auto;
    padding: 0.25rem 0.5rem;
    border: none;
    background: transparent;
    color: hsl(var(--muted-foreground));
    font-size: 0.75rem;
    cursor: pointer;
    border-radius: var(--radius, 0.25rem);
  }

  .bi-filter-slicer__clear:hover {
    color: hsl(var(--destructive));
    background: hsl(var(--destructive) / 0.1);
  }
</style>
