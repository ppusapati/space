<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import {
    type TableColumn,
    type SortState,
    tableClasses,
    tableSizeClasses,
  } from './table.types';
  import { sortData, toggleSort, getNestedValue } from './table.logic';
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
  export let id: string = uid('table');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    sort: { column: string; direction: 'asc' | 'desc' | null };
    select: { keys: (string | number)[]; rows: T[] };
    rowClick: { row: T; index: number };
  }>();

  // Computed
  $: visibleColumns = columns.filter(c => c.visible !== false);
  $: sizeConfig = tableSizeClasses[size];
  $: sortedData = sortable && sortState.column ? sortData(data, sortState, columns) : data;

  $: wrapperStyle = maxHeight ? `max-height: ${maxHeight}; overflow: auto;` : '';

  $: isAllSelected = data.length > 0 && selectedKeys.length === data.length;
  $: isSomeSelected = selectedKeys.length > 0 && selectedKeys.length < data.length;

  function getRowKey(row: T, index: number): string | number {
    const key = getNestedValue(row, rowKey);
    return key !== undefined ? (key as string | number) : index;
  }

  function isRowSelected(row: T, index: number): boolean {
    return selectedKeys.includes(getRowKey(row, index));
  }

  function handleSort(column: TableColumn<T>) {
    if (!sortable || !column.sortable) return;

    sortState = toggleSort(sortState, column.key);
    dispatch('sort', {
      column: column.key,
      direction: sortState.direction,
    });
  }

  function handleSelectAll() {
    if (isAllSelected) {
      selectedKeys = [];
    } else {
      selectedKeys = data.map((row, index) => getRowKey(row, index));
    }
    dispatch('select', {
      keys: selectedKeys,
      rows: data.filter((row, index) => selectedKeys.includes(getRowKey(row, index))),
    });
  }

  function handleSelectRow(row: T, index: number) {
    const key = getRowKey(row, index);

    if (selectionMode === 'single') {
      selectedKeys = isRowSelected(row, index) ? [] : [key];
    } else {
      if (isRowSelected(row, index)) {
        selectedKeys = selectedKeys.filter(k => k !== key);
      } else {
        selectedKeys = [...selectedKeys, key];
      }
    }

    dispatch('select', {
      keys: selectedKeys,
      rows: data.filter((r, i) => selectedKeys.includes(getRowKey(r, i))),
    });
  }

  function handleRowClick(row: T, index: number) {
    dispatch('rowClick', { row, index });
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
      column.sortable && sortable && tableClasses.thSortable,
      bordered && tableClasses.tdBordered,
      column.align === 'center' && 'text-center',
      column.align === 'right' && 'text-right',
      column.headerClass
    );
  }
</script>

<div
  class={cn(tableClasses.wrapper, fullWidth && 'w-full', className)}
  style={wrapperStyle}
  {id}
  data-testid={testId || undefined}
>
  <table class={tableClasses.table} role="grid" aria-busy={loading}>
    <thead class={cn(tableClasses.thead, stickyHeader && tableClasses.theadSticky)}>
      <tr class={tableClasses.tr}>
        {#if selectable}
          <th class={cn(tableClasses.th, sizeConfig.th, 'w-10')} scope="col">
            {#if selectionMode === 'multiple'}
              <input
                type="checkbox"
                class={tableClasses.checkbox}
                checked={isAllSelected}
                indeterminate={isSomeSelected}
                on:change={handleSelectAll}
                aria-label="Select all rows"
              />
            {/if}
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
            {#if column.sortable && sortable}
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
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
                      </svg>
                    {:else}
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                      </svg>
                    {/if}
                  {:else}
                    <svg class="w-4 h-4 text-neutral-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4" />
                    </svg>
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
        {#each sortedData as row, index (getRowKey(row, index))}
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
    </tbody>
  </table>

  {#if loading}
    <div class={tableClasses.loading}>
      <svg class="w-8 h-8 text-brand-primary-500 animate-spin" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
    </div>
  {/if}
</div>
