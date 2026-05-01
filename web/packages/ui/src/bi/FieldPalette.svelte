<script lang="ts">
  import type { DatasetFieldItem } from './FieldPalette.types';

  // ─── Props ──────────────────────────────────────────────────────────────────
  interface Props {
    fields: DatasetFieldItem[];
    searchable?: boolean;
    grouped?: boolean;
    class?: string;
    onfieldDragStart?: (e: CustomEvent<{ field: DatasetFieldItem }>) => void;
    onfieldAdd?: (e: CustomEvent<{ field: DatasetFieldItem }>) => void;
  }

  let {
    fields,
    searchable = true,
    grouped = true,
    class: className = '',
    onfieldDragStart,
    onfieldAdd,
  }: Props = $props();

  // ─── State ──────────────────────────────────────────────────────────────────
  let search = $state('');
  let collapsedGroups = $state<Set<string>>(new Set());

  // ─── Derived ────────────────────────────────────────────────────────────────
  let filteredFields = $derived(
    search.trim()
      ? fields.filter((f) =>
          f.label.toLowerCase().includes(search.trim().toLowerCase()) ||
          f.column_name.toLowerCase().includes(search.trim().toLowerCase())
        )
      : fields
  );

  let groupedFields = $derived.by(() => {
    if (!grouped) return new Map([['', filteredFields]]);
    const map = new Map<string, DatasetFieldItem[]>();
    for (const field of filteredFields) {
      const group = field.group_name || 'Other';
      if (!map.has(group)) map.set(group, []);
      map.get(group)!.push(field);
    }
    return map;
  });

  // ─── Helpers ────────────────────────────────────────────────────────────────
  function getRoleIcon(role: string): string {
    switch (role) {
      case 'dimension': return '#';
      case 'measure': return 'Σ';
      case 'attribute': return 'i';
      default: return '?';
    }
  }

  function getRoleSvg(role: string): string {
    switch (role) {
      case 'dimension':
        return '<path d="M6 2h8l4 4v12a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2z"/><path d="M9 13h6"/><path d="M9 9h3"/>';
      case 'measure':
        return '<path d="M4 20h16"/><path d="M4 20V10"/><path d="M10 20V4"/><path d="M16 20v-6"/>';
      case 'attribute':
        return '<circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/>';
      default:
        return '<circle cx="12" cy="12" r="10"/>';
    }
  }

  function getDataTypeBadge(dt: string): string {
    const map: Record<string, string> = {
      string: 'ABC',
      integer: '123',
      decimal: '1.2',
      date: 'CAL',
      datetime: 'DT',
      boolean: 'T/F',
      currency: '$',
      percentage: '%',
    };
    return map[dt] || dt.slice(0, 3).toUpperCase();
  }

  function toggleGroup(group: string) {
    const next = new Set(collapsedGroups);
    if (next.has(group)) next.delete(group);
    else next.add(group);
    collapsedGroups = next;
  }

  function handleDragStart(e: DragEvent, field: DatasetFieldItem) {
    if (!e.dataTransfer) return;
    e.dataTransfer.setData('application/x-bi-field', JSON.stringify(field));
    e.dataTransfer.effectAllowed = 'copy';
    onfieldDragStart?.(new CustomEvent('fieldDragStart', { detail: { field } }));
  }

  function handleDoubleClick(field: DatasetFieldItem) {
    onfieldAdd?.(new CustomEvent('fieldAdd', { detail: { field } }));
  }
</script>

<div class="bi-field-palette {className}" role="region" aria-label="Field palette">
  {#if searchable}
    <div class="bi-field-palette__search">
      <svg class="bi-field-palette__search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16">
        <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
      </svg>
      <input
        type="text"
        bind:value={search}
        placeholder="Search fields..."
        class="bi-field-palette__search-input"
        aria-label="Search fields"
      />
      {#if search}
        <button
          class="bi-field-palette__search-clear"
          onclick={() => search = ''}
          aria-label="Clear search"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
            <path d="M18 6 6 18"/><path d="m6 6 12 12"/>
          </svg>
        </button>
      {/if}
    </div>
  {/if}

  <div class="bi-field-palette__list" role="list">
    {#each [...groupedFields.entries()] as [group, groupFields]}
      {#if grouped && group}
        <button
          class="bi-field-palette__group-header"
          onclick={() => toggleGroup(group)}
          aria-expanded={!collapsedGroups.has(group)}
        >
          <svg
            class="bi-field-palette__chevron"
            class:bi-field-palette__chevron--collapsed={collapsedGroups.has(group)}
            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"
          >
            <path d="m6 9 6 6 6-6"/>
          </svg>
          <span class="bi-field-palette__group-name">{group}</span>
          <span class="bi-field-palette__group-count">{groupFields.length}</span>
        </button>
      {/if}

      {#if !grouped || !collapsedGroups.has(group)}
        {#each groupFields as field (field.id)}
          <div
            class="bi-field-palette__field"
            class:bi-field-palette__field--dimension={field.role === 'dimension'}
            class:bi-field-palette__field--measure={field.role === 'measure'}
            class:bi-field-palette__field--attribute={field.role === 'attribute'}
            draggable="true"
            role="listitem"
            tabindex="0"
            ondragstart={(e: DragEvent) => handleDragStart(e, field)}
            ondblclick={() => handleDoubleClick(field)}
            title="{field.label} ({field.data_type})"
          >
            <span class="bi-field-palette__field-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="14" height="14">
                {@html getRoleSvg(field.role)}
              </svg>
            </span>
            <span class="bi-field-palette__field-label">{field.label}</span>
            <span class="bi-field-palette__field-type">{getDataTypeBadge(field.data_type)}</span>
          </div>
        {/each}
      {/if}
    {/each}

    {#if filteredFields.length === 0}
      <div class="bi-field-palette__empty">
        {search ? 'No fields match your search' : 'No fields available'}
      </div>
    {/if}
  </div>
</div>

<style>
  .bi-field-palette {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: hsl(var(--background));
    border: 1px solid hsl(var(--border));
    border-radius: var(--radius, 0.5rem);
    overflow: hidden;
  }

  .bi-field-palette__search {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 0.75rem;
    border-bottom: 1px solid hsl(var(--border));
    background: hsl(var(--muted));
  }

  .bi-field-palette__search-icon {
    flex-shrink: 0;
    color: hsl(var(--muted-foreground));
  }

  .bi-field-palette__search-input {
    flex: 1;
    border: none;
    background: transparent;
    font-size: 0.8125rem;
    color: hsl(var(--foreground));
    outline: none;
  }

  .bi-field-palette__search-input::placeholder {
    color: hsl(var(--muted-foreground));
  }

  .bi-field-palette__search-clear {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.125rem;
    border: none;
    background: transparent;
    color: hsl(var(--muted-foreground));
    cursor: pointer;
    border-radius: var(--radius, 0.25rem);
  }

  .bi-field-palette__search-clear:hover {
    color: hsl(var(--foreground));
    background: hsl(var(--accent));
  }

  .bi-field-palette__list {
    flex: 1;
    overflow-y: auto;
    padding: 0.25rem 0;
  }

  .bi-field-palette__group-header {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    width: 100%;
    padding: 0.375rem 0.75rem;
    border: none;
    background: transparent;
    font-size: 0.6875rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: hsl(var(--muted-foreground));
    cursor: pointer;
  }

  .bi-field-palette__group-header:hover {
    background: hsl(var(--accent));
  }

  .bi-field-palette__chevron {
    flex-shrink: 0;
    transition: transform 0.15s ease;
  }

  .bi-field-palette__chevron--collapsed {
    transform: rotate(-90deg);
  }

  .bi-field-palette__group-name {
    flex: 1;
    text-align: left;
  }

  .bi-field-palette__group-count {
    font-size: 0.625rem;
    padding: 0 0.375rem;
    border-radius: 9999px;
    background: hsl(var(--muted));
    color: hsl(var(--muted-foreground));
  }

  .bi-field-palette__field {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.375rem 0.75rem;
    cursor: grab;
    user-select: none;
    font-size: 0.8125rem;
    color: hsl(var(--foreground));
    transition: background 0.1s ease;
  }

  .bi-field-palette__field:hover {
    background: hsl(var(--accent));
  }

  .bi-field-palette__field:active {
    cursor: grabbing;
    background: hsl(var(--accent));
  }

  .bi-field-palette__field:focus-visible {
    outline: 2px solid hsl(var(--ring));
    outline-offset: -2px;
    border-radius: var(--radius, 0.25rem);
  }

  .bi-field-palette__field-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 1.25rem;
    height: 1.25rem;
    flex-shrink: 0;
    border-radius: var(--radius, 0.25rem);
  }

  .bi-field-palette__field--dimension .bi-field-palette__field-icon {
    color: hsl(var(--primary));
  }

  .bi-field-palette__field--measure .bi-field-palette__field-icon {
    color: hsl(var(--chart-2, 220 70% 50%));
  }

  .bi-field-palette__field--attribute .bi-field-palette__field-icon {
    color: hsl(var(--muted-foreground));
  }

  .bi-field-palette__field-label {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .bi-field-palette__field-type {
    flex-shrink: 0;
    font-size: 0.625rem;
    font-weight: 500;
    padding: 0.0625rem 0.3125rem;
    border-radius: var(--radius, 0.25rem);
    background: hsl(var(--muted));
    color: hsl(var(--muted-foreground));
    font-family: monospace;
  }

  .bi-field-palette__empty {
    padding: 1.5rem 0.75rem;
    text-align: center;
    font-size: 0.8125rem;
    color: hsl(var(--muted-foreground));
  }
</style>
