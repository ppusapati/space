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
    calculatePagination,
    getPageNumbers,
  } from './table.logic';
  import {
    type TreeConfig,
    type TreeRow,
    treeClasses,
    flattenTreeData,
    getDefaultExpandedIds,
    searchTreeData,
  } from './datatable.types';
  import type { Size, PaginationState } from '../types';

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

  // Tree specific props
  export let tree: TreeConfig<T> = { enabled: true };
  export let expandedRowIds: (string | number)[] = [];

  // Search props
  export let searchable: boolean = true;
  export let searchQuery: string = '';

  // Pagination props
  export let paginated: boolean = false;
  export let pagination: PaginationState = { page: 1, pageSize: 50, total: 0 };

  export let id: string = uid('treetable');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    sort: { column: string; direction: 'asc' | 'desc' | null };
    select: { keys: (string | number)[]; rows: T[] };
    rowClick: { row: T; index: number };
    expand: { rowId: string | number; expanded: boolean };
    search: { query: string };
    pageChange: { page: number; pageSize: number };
    loadChildren: { row: T; callback: (children: T[]) => void };
  }>();

  // Internal state
  let expandedIds: Set<string | number> = new Set(expandedRowIds);
  let loadingRows: Set<string | number> = new Set();

  // Configuration
  $: childrenKey = tree.childrenKey || 'children';
  $: indentWidth = tree.indentWidth || 24;
  $: sizeConfig = tableSizeClasses[size];

  // Initialize expanded state
  onMount(() => {
    if (expandedRowIds.length === 0 && tree.enabled) {
      const defaultExpanded = getDefaultExpandedIds(data, tree);
      expandedIds = defaultExpanded;
    }
  });

  // Filter data by search
  $: searchedData = searchQuery && searchable
    ? searchTreeData(data, searchQuery, columns, childrenKey)
    : data;

  // Flatten tree for display
  $: flatData = tree.enabled
    ? flattenTreeData(searchedData, tree, expandedIds)
    : searchedData.map((item, index) => ({
        id: (item as { id?: string | number }).id ?? index,
        data: item,
        depth: 0,
        hasChildren: false,
        expanded: false,
      } as TreeRow<T>));

  // Pagination
  $: paginatedData = paginated
    ? flatData.slice((pagination.page - 1) * pagination.pageSize, pagination.page * pagination.pageSize)
    : flatData;

  $: totalItems = flatData.length;
  $: paginationInfo = paginated
    ? calculatePagination(totalItems, pagination.page, pagination.pageSize)
    : null;
  $: pageNumbers = paginationInfo
    ? getPageNumbers(pagination.page, paginationInfo.totalPages)
    : [];

  $: visibleColumns = columns.filter((c) => c.visible !== false);
  $: wrapperStyle = maxHeight ? `max-height: ${maxHeight}; overflow: auto;` : '';

  $: isAllSelected = paginatedData.length > 0 && selectedKeys.length === paginatedData.length;
  $: isSomeSelected = selectedKeys.length > 0 && selectedKeys.length < paginatedData.length;

  // Helpers
  function getRowKey(row: TreeRow<T>): string | number {
    return row.id;
  }

  function isRowSelected(row: TreeRow<T>): boolean {
    return selectedKeys.includes(row.id);
  }

  // Event handlers
  function handleSort(column: TableColumn<T>) {
    if (!sortable || column.sortable === false) return;
    sortState = toggleSort(sortState, column.key);
    dispatch('sort', { column: column.key, direction: sortState.direction });
  }

  function handleSelectAll() {
    if (isAllSelected) {
      selectedKeys = [];
    } else {
      selectedKeys = paginatedData.map((row) => row.id);
    }
    dispatch('select', {
      keys: selectedKeys,
      rows: paginatedData.filter((row) => selectedKeys.includes(row.id)).map((r) => r.data),
    });
  }

  function handleSelectRow(row: TreeRow<T>) {
    if (selectionMode === 'single') {
      selectedKeys = isRowSelected(row) ? [] : [row.id];
    } else {
      if (isRowSelected(row)) {
        selectedKeys = selectedKeys.filter((k) => k !== row.id);
      } else {
        selectedKeys = [...selectedKeys, row.id];
      }
    }
    dispatch('select', {
      keys: selectedKeys,
      rows: paginatedData.filter((r) => selectedKeys.includes(r.id)).map((r) => r.data),
    });
  }

  function handleRowClick(row: TreeRow<T>, index: number) {
    dispatch('rowClick', { row: row.data, index });
  }

  function handleToggleExpand(row: TreeRow<T>) {
    const wasExpanded = expandedIds.has(row.id);

    if (wasExpanded) {
      expandedIds.delete(row.id);
    } else {
      // Check if we need to lazy load children
      const rowData = row.data;
      const children = (rowData as Record<string, unknown>)[childrenKey] as T[] | undefined;

      if (tree.loadChildren && (!children || children.length === 0) && row.hasChildren) {
        loadingRows.add(row.id);
        loadingRows = loadingRows;

        tree.loadChildren(rowData).then((loadedChildren) => {
          (rowData as Record<string, unknown>)[childrenKey] = loadedChildren;
          loadingRows.delete(row.id);
          loadingRows = loadingRows;
          expandedIds.add(row.id);
          expandedIds = expandedIds;
        });
        return;
      }

      expandedIds.add(row.id);
    }

    expandedIds = expandedIds;
    expandedRowIds = Array.from(expandedIds);
    dispatch('expand', { rowId: row.id, expanded: !wasExpanded });
  }

  function handleSearch(event: Event) {
    const target = event.target as HTMLInputElement;
    searchQuery = target.value;
    pagination = { ...pagination, page: 1 };
    dispatch('search', { query: searchQuery });
  }

  function handlePageChange(newPage: number) {
    pagination = { ...pagination, page: newPage };
    dispatch('pageChange', { page: newPage, pageSize: pagination.pageSize });
  }

  function getCellValue(row: TreeRow<T>, column: TableColumn<T>): string {
    const value = getNestedValue(row.data, column.key);
    if (column.format) {
      return column.format(value, row.data);
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

  // Expand/collapse all
  export function expandAll() {
    function collectIds(items: T[]): (string | number)[] {
      const ids: (string | number)[] = [];
      for (const item of items) {
        const id = (item as { id?: string | number }).id;
        if (id !== undefined) ids.push(id);
        const children = (item as Record<string, unknown>)[childrenKey] as T[] | undefined;
        if (children) ids.push(...collectIds(children));
      }
      return ids;
    }
    expandedIds = new Set(collectIds(data));
    expandedRowIds = Array.from(expandedIds);
  }

  export function collapseAll() {
    expandedIds = new Set();
    expandedRowIds = [];
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
      {#if tree.enabled}
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
    <table class={tableClasses.table} role="treegrid" aria-busy={loading}>
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

          {#each visibleColumns as column, colIndex (column.key)}
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
        {#if paginatedData.length === 0}
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
          {#each paginatedData as row, index (getRowKey(row))}
            {@const isSelected = isRowSelected(row)}
            {@const isLoading = loadingRows.has(row.id)}
            <tr
              class={cn(
                tableClasses.tr,
                treeClasses.row,
                hoverable && tableClasses.trHover,
                striped && tableClasses.trStriped,
                isSelected && tableClasses.trSelected
              )}
              role="row"
              aria-level={row.depth + 1}
              aria-expanded={row.hasChildren ? row.expanded : undefined}
              aria-selected={selectable ? isSelected : undefined}
              on:click={() => handleRowClick(row, index)}
            >
              {#if selectable}
                <td class={cn(tableClasses.td, sizeConfig.td, 'w-10')}>
                  <input
                    type={selectionMode === 'single' ? 'radio' : 'checkbox'}
                    class={tableClasses.checkbox}
                    checked={isSelected}
                    on:change={() => handleSelectRow(row)}
                    on:click|stopPropagation
                    aria-label={`Select row ${index + 1}`}
                  />
                </td>
              {/if}

              {#each visibleColumns as column, colIndex (column.key)}
                <td class={getCellClasses(column)}>
                  {#if colIndex === 0 && tree.enabled}
                    <!-- Tree indent and expand icon -->
                    <span
                      class={treeClasses.indent}
                      style="padding-left: {row.depth * indentWidth}px"
                    >
                      {#if row.hasChildren || isLoading}
                        <button
                          type="button"
                          class={cn(
                            treeClasses.expandIcon,
                            row.expanded && treeClasses.expandIconExpanded,
                            isLoading && treeClasses.expandIconLoading
                          )}
                          on:click|stopPropagation={() => handleToggleExpand(row)}
                          aria-label={row.expanded ? 'Collapse' : 'Expand'}
                        >
                          {#if isLoading}
                            <Icon name="loader" size="sm" class="animate-spin" />
                          {:else}
                            <Icon name="chevron-right" size="sm" />
                          {/if}
                        </button>
                      {:else}
                        <span class={treeClasses.leaf}></span>
                      {/if}
                    </span>
                  {/if}
                  <slot name="cell" {column} row={row.data} value={getNestedValue(row.data, column.key)}>
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
        <Icon name="loader" size="lg" class="text-brand-primary-500 animate-spin" />
      </div>
    {/if}
  </div>

  <!-- Pagination -->
  {#if paginated && paginationInfo}
    <div class={paginationClasses.container}>
      <div class={paginationClasses.info}>
        Showing {paginationInfo.startItem} to {paginationInfo.endItem} of {paginationInfo.total} entries
      </div>

      <div class={paginationClasses.controls}>
        <button
          type="button"
          class={paginationClasses.button}
          disabled={pagination.page === 1}
          on:click={() => handlePageChange(1)}
          aria-label="First page"
        >
          <Icon name="chevrons-left" size="sm" />
        </button>

        <button
          type="button"
          class={paginationClasses.button}
          disabled={pagination.page === 1}
          on:click={() => handlePageChange(pagination.page - 1)}
          aria-label="Previous page"
        >
          <Icon name="chevron-left" size="sm" />
        </button>

        <div class="flex items-center gap-1">
          {#each pageNumbers as pageNum}
            {#if pageNum === 'ellipsis'}
              <span class="px-2 text-neutral-400">...</span>
            {:else}
              <button
                type="button"
                class={cn(
                  paginationClasses.pageButton,
                  pagination.page === pageNum
                    ? paginationClasses.pageButtonActive
                    : paginationClasses.pageButtonInactive
                )}
                on:click={() => handlePageChange(pageNum)}
              >
                {pageNum}
              </button>
            {/if}
          {/each}
        </div>

        <button
          type="button"
          class={paginationClasses.button}
          disabled={pagination.page >= paginationInfo.totalPages}
          on:click={() => handlePageChange(pagination.page + 1)}
          aria-label="Next page"
        >
          <Icon name="chevron-right" size="sm" />
        </button>

        <button
          type="button"
          class={paginationClasses.button}
          disabled={pagination.page >= paginationInfo.totalPages}
          on:click={() => handlePageChange(paginationInfo.totalPages)}
          aria-label="Last page"
        >
          <Icon name="chevrons-right" size="sm" />
        </button>
      </div>
    </div>
  {/if}
</div>
