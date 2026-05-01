<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import Icon from '../display/Icon.svelte';
  import {
    type TableColumn,
    type SortState,
    tableClasses,
    tableSizeClasses,
    toolbarClasses,
    paginationClasses,
  } from './table.types';
  import {
    toggleSort,
    getNestedValue,
    sortData,
    searchData,
  } from './table.logic';
  import {
    type GroupConfig,
    type RowGroup,
    groupClasses,
    groupData,
    flattenGroups,
  } from './datatable.types';
  import type { Size } from '../types';

  type T = $$Generic<Record<string, unknown>>;

  // Props
  export let columns: TableColumn<T>[] = [];
  export let data: T[] = [];
  export let rowKey: string = 'id';
  export let size: Size = 'md';
  export let striped: boolean = false;
  export let hoverable: boolean = true;
  export let bordered: boolean = false;
  export let selectable: boolean = false;
  export let selectionMode: 'single' | 'multiple' = 'multiple';
  export let selectedKeys: (string | number)[] = [];
  export let sortable: boolean = true;
  export let sortState: SortState = { column: null, direction: null };
  export let loading: boolean = false;
  export let emptyMessage: string = 'No data available';
  export let stickyHeader: boolean = false;
  export let maxHeight: string = '';
  export let fullWidth: boolean = true;

  // Grouping specific props
  export let grouping: GroupConfig<T> = { enabled: true, groupBy: '' };
  export let expandedGroupKeys: string[] = [];

  // Search props
  export let searchable: boolean = true;
  export let searchQuery: string = '';

  export let id: string = uid('grouptable');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    sort: { column: string; direction: 'asc' | 'desc' | null };
    select: { keys: (string | number)[]; rows: T[] };
    rowClick: { row: T; index: number };
    groupToggle: { groupKey: string; expanded: boolean };
    search: { query: string };
  }>();

  // Internal state
  let expandedGroups: Set<string> = new Set(expandedGroupKeys);

  // Configuration
  $: sizeConfig = tableSizeClasses[size];
  $: visibleColumns = columns.filter((c) => c.visible !== false);

  // Process data
  $: searchedData = searchQuery && searchable
    ? searchData(data, searchQuery, columns)
    : data;

  $: sortedData = sortState.column
    ? sortData(searchedData, sortState, columns)
    : searchedData;

  // Group data
  $: groups = grouping.enabled && grouping.groupBy
    ? groupData(sortedData, grouping, expandedGroups)
    : [];

  $: flattenedData = grouping.enabled && grouping.groupBy
    ? flattenGroups(groups, expandedGroups)
    : sortedData.map((row) => ({ type: 'row' as const, row, groupKey: '' }));

  // Initialize expanded state
  onMount(() => {
    if (expandedGroupKeys.length === 0 && grouping.defaultExpanded) {
      expandedGroups = new Set(groups.map((g) => g.key));
    }
  });

  $: wrapperStyle = maxHeight ? `max-height: ${maxHeight}; overflow: auto;` : '';

  // Helpers
  function getRowKeyValue(row: T, index: number): string | number {
    const key = getNestedValue(row, rowKey);
    return key !== undefined ? (key as string | number) : index;
  }

  function isRowSelected(row: T, index: number): boolean {
    return selectedKeys.includes(getRowKeyValue(row, index));
  }

  // Event handlers
  function handleSort(column: TableColumn<T>) {
    if (!sortable || column.sortable === false) return;
    sortState = toggleSort(sortState, column.key);
    dispatch('sort', { column: column.key, direction: sortState.direction });
  }

  function handleSelectRow(row: T, index: number) {
    const key = getRowKeyValue(row, index);
    if (selectionMode === 'single') {
      selectedKeys = isRowSelected(row, index) ? [] : [key];
    } else {
      if (isRowSelected(row, index)) {
        selectedKeys = selectedKeys.filter((k) => k !== key);
      } else {
        selectedKeys = [...selectedKeys, key];
      }
    }
    dispatch('select', {
      keys: selectedKeys,
      rows: data.filter((r, i) => selectedKeys.includes(getRowKeyValue(r, i))),
    });
  }

  function handleSelectGroup(group: RowGroup<T>) {
    const groupRowKeys = group.rows.map((row, index) => getRowKeyValue(row, index));
    const allSelected = groupRowKeys.every((key) => selectedKeys.includes(key));

    if (allSelected) {
      selectedKeys = selectedKeys.filter((k) => !groupRowKeys.includes(k));
    } else {
      selectedKeys = [...new Set([...selectedKeys, ...groupRowKeys])];
    }

    dispatch('select', {
      keys: selectedKeys,
      rows: data.filter((r, i) => selectedKeys.includes(getRowKeyValue(r, i))),
    });
  }

  function handleRowClick(row: T, index: number) {
    dispatch('rowClick', { row, index });
  }

  function handleToggleGroup(groupKey: string) {
    const wasExpanded = expandedGroups.has(groupKey);

    if (wasExpanded) {
      expandedGroups.delete(groupKey);
    } else {
      expandedGroups.add(groupKey);
    }

    expandedGroups = expandedGroups;
    expandedGroupKeys = Array.from(expandedGroups);
    dispatch('groupToggle', { groupKey, expanded: !wasExpanded });
  }

  function handleSearch(event: Event) {
    const target = event.target as HTMLInputElement;
    searchQuery = target.value;
    dispatch('search', { query: searchQuery });
  }

  function getCellValue(row: T, column: TableColumn<T>): string {
    const value = getNestedValue(row, column.key);
    if (column.format) {
      return column.format(value, row);
    }
    return String(value ?? '');
  }

  function getCellClasses(column: TableColumn<T>): string {
    return cn(
      tableClasses.td,
      sizeConfig.td,
      bordered && tableClasses.tdBordered,
      column.align === 'center' && 'text-center',
      column.align === 'right' && 'text-right',
      column.cellClass
    );
  }

  function getHeaderClasses(column: TableColumn<T>): string {
    return cn(
      tableClasses.th,
      sizeConfig.th,
      column.sortable !== false && sortable && tableClasses.thSortable,
      bordered && tableClasses.tdBordered,
      column.align === 'center' && 'text-center',
      column.align === 'right' && 'text-right',
      column.headerClass
    );
  }

  function isGroupAllSelected(group: RowGroup<T>): boolean {
    const groupRowKeys = group.rows.map((row, index) => getRowKeyValue(row, index));
    return groupRowKeys.every((key) => selectedKeys.includes(key));
  }

  function isGroupSomeSelected(group: RowGroup<T>): boolean {
    const groupRowKeys = group.rows.map((row, index) => getRowKeyValue(row, index));
    return groupRowKeys.some((key) => selectedKeys.includes(key)) && !isGroupAllSelected(group);
  }

  function formatAggregateValue(value: unknown, column: TableColumn<T>): string {
    if (value === null || value === undefined) return '-';
    if (column.format) {
      return column.format(value, {} as T);
    }
    if (typeof value === 'number') {
      return value.toLocaleString(undefined, { maximumFractionDigits: 2 });
    }
    return String(value);
  }

  // Expand/collapse all groups
  export function expandAll() {
    expandedGroups = new Set(groups.map((g) => g.key));
    expandedGroupKeys = Array.from(expandedGroups);
  }

  export function collapseAll() {
    expandedGroups = new Set();
    expandedGroupKeys = [];
  }
</script>

<div class={cn('bg-neutral-white rounded-lg border border-neutral-200', fullWidth && 'w-full', className)} {id} data-testid={testId || undefined}>
  <!-- Toolbar -->
  <div class={toolbarClasses.container}>
    <div class={toolbarClasses.left}>
      {#if searchable}
        <div class={toolbarClasses.search}>
          <svg class={toolbarClasses.searchIcon} fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            type="text"
            class={toolbarClasses.searchInput}
            placeholder="Search..."
            value={searchQuery}
            on:input={handleSearch}
          />
        </div>
      {/if}
    </div>

    <div class={toolbarClasses.right}>
      {#if grouping.enabled && grouping.collapsible !== false}
        <button
          type="button"
          class={cn(toolbarClasses.button, toolbarClasses.buttonSecondary)}
          on:click={expandAll}
        >
          Expand All
        </button>
        <button
          type="button"
          class={cn(toolbarClasses.button, toolbarClasses.buttonSecondary)}
          on:click={collapseAll}
        >
          Collapse All
        </button>
      {/if}
      <slot name="toolbar-actions" />
    </div>
  </div>

  <!-- Table -->
  <div class={tableClasses.wrapper} style={wrapperStyle}>
    <table class={tableClasses.table} role="grid" aria-busy={loading}>
      <thead class={cn(tableClasses.thead, stickyHeader && tableClasses.theadSticky)}>
        <tr class={tableClasses.tr}>
          {#if selectable}
            <th class={cn(tableClasses.th, sizeConfig.th, 'w-10')} scope="col">
              <!-- Group select header is handled per-group -->
            </th>
          {/if}

          {#each visibleColumns as column (column.key)}
            <th
              class={getHeaderClasses(column)}
              scope="col"
              style={column.width ? `width: ${column.width}` : ''}
              aria-sort={sortState.column === column.key
                ? sortState.direction === 'asc'
                  ? 'ascending'
                  : 'descending'
                : 'none'}
            >
              {#if column.sortable !== false && sortable}
                <button
                  type="button"
                  class="inline-flex items-center gap-1 w-full"
                  on:click={() => handleSort(column)}
                >
                  <slot name="header" {column}>
                    {column.header}
                  </slot>
                  <span class={tableClasses.sortIcon}>
                    {#if sortState.column === column.key}
                      {#if sortState.direction === 'asc'}
                        <Icon name="chevron-up" size="sm" />
                      {:else}
                        <Icon name="chevron-down" size="sm" />
                      {/if}
                    {/if}
                  </span>
                </button>
              {:else}
                <slot name="header" {column}>
                  {column.header}
                </slot>
              {/if}
            </th>
          {/each}
        </tr>
      </thead>

      <tbody>
        {#if !grouping.enabled || !grouping.groupBy || flattenedData.length === 0}
          <!-- No grouping - render flat rows -->
          {#if sortedData.length === 0}
            <tr>
              <td
                colspan={visibleColumns.length + (selectable ? 1 : 0)}
                class={tableClasses.empty}
              >
                <slot name="empty">
                  {emptyMessage}
                </slot>
              </td>
            </tr>
          {:else}
            {#each sortedData as row, index (getRowKeyValue(row, index))}
              {@const isSelected = isRowSelected(row, index)}
              <tr
                class={cn(
                  tableClasses.tr,
                  hoverable && tableClasses.trHover,
                  striped && tableClasses.trStriped,
                  isSelected && tableClasses.trSelected
                )}
                role="row"
                aria-selected={selectable ? isSelected : undefined}
                on:click={() => handleRowClick(row, index)}
              >
                {#if selectable}
                  <td class={cn(tableClasses.td, sizeConfig.td, 'w-10')}>
                    <input
                      type={selectionMode === 'single' ? 'radio' : 'checkbox'}
                      class={tableClasses.checkbox}
                      checked={isSelected}
                      on:change={() => handleSelectRow(row, index)}
                      on:click|stopPropagation
                      aria-label={`Select row ${index + 1}`}
                    />
                  </td>
                {/if}

                {#each visibleColumns as column (column.key)}
                  <td class={getCellClasses(column)}>
                    <slot name="cell" {column} {row} value={getNestedValue(row, column.key)}>
                      {getCellValue(row, column)}
                    </slot>
                  </td>
                {/each}
              </tr>
            {/each}
          {/if}
        {:else}
          <!-- Grouped data -->
          {#each flattenedData as item, flatIndex}
            {#if item.type === 'group'}
              {@const group = item.group}
              {@const isExpanded = expandedGroups.has(group.key)}
              {@const groupAllSelected = isGroupAllSelected(group)}
              {@const groupSomeSelected = isGroupSomeSelected(group)}

              <tr
                class={cn(groupClasses.row)}
                role="row"
                aria-expanded={isExpanded}
                style="padding-left: {group.depth * 16}px"
              >
                {#if selectable && selectionMode === 'multiple'}
                  <td class={cn(groupClasses.cell, 'w-10')}>
                    <input
                      type="checkbox"
                      class={tableClasses.checkbox}
                      checked={groupAllSelected}
                      indeterminate={groupSomeSelected}
                      on:change={() => handleSelectGroup(group)}
                      on:click|stopPropagation
                      aria-label={`Select group ${group.label}`}
                    />
                  </td>
                {/if}

                <td
                  class={groupClasses.cell}
                  colspan={selectable ? visibleColumns.length : visibleColumns.length}
                >
                  <button
                    type="button"
                    class="flex items-center w-full text-left"
                    on:click={() => handleToggleGroup(group.key)}
                  >
                    {#if grouping.collapsible !== false}
                      <span class={cn(groupClasses.expandIcon, isExpanded && 'rotate-90')}>
                        <Icon name="chevron-right" size="sm" />
                      </span>
                    {/if}

                    <span class={groupClasses.label}>
                      <slot name="group-label" {group}>
                        {group.label}
                      </slot>
                    </span>

                    {#if grouping.showCount !== false}
                      <span class={groupClasses.count}>
                        ({group.count})
                      </span>
                    {/if}

                    <!-- Aggregates -->
                    {#if group.aggregates && grouping.aggregates}
                      <span class="flex-1"></span>
                      <span class={groupClasses.aggregate}>
                        {#each Object.entries(grouping.aggregates) as [columnKey, aggType]}
                          {@const column = columns.find((c) => c.key === columnKey)}
                          {#if column && group.aggregates[columnKey] !== undefined}
                            <span class="ml-4">
                              {column.header}: {formatAggregateValue(group.aggregates[columnKey], column)}
                            </span>
                          {/if}
                        {/each}
                      </span>
                    {/if}
                  </button>
                </td>
              </tr>
            {:else if item.type === 'row'}
              {@const row = item.row}
              {@const rowIndex = sortedData.indexOf(row)}
              {@const isSelected = isRowSelected(row, rowIndex)}

              <tr
                class={cn(
                  tableClasses.tr,
                  hoverable && tableClasses.trHover,
                  striped && tableClasses.trStriped,
                  isSelected && tableClasses.trSelected
                )}
                role="row"
                aria-selected={selectable ? isSelected : undefined}
                on:click={() => handleRowClick(row, rowIndex)}
              >
                {#if selectable}
                  <td class={cn(tableClasses.td, sizeConfig.td, 'w-10')}>
                    <input
                      type={selectionMode === 'single' ? 'radio' : 'checkbox'}
                      class={tableClasses.checkbox}
                      checked={isSelected}
                      on:change={() => handleSelectRow(row, rowIndex)}
                      on:click|stopPropagation
                      aria-label={`Select row ${rowIndex + 1}`}
                    />
                  </td>
                {/if}

                {#each visibleColumns as column (column.key)}
                  <td class={getCellClasses(column)}>
                    <slot name="cell" {column} {row} value={getNestedValue(row, column.key)}>
                      {getCellValue(row, column)}
                    </slot>
                  </td>
                {/each}
              </tr>
            {/if}
          {/each}
        {/if}
      </tbody>
    </table>

    {#if loading}
      <div class={tableClasses.loading}>
        <Icon name="loader" size="lg" class="text-brand-primary-500 animate-spin" />
      </div>
    {/if}
  </div>
</div>
