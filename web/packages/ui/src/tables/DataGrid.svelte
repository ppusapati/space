<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import {
    type TableColumn,
    type SortState,
    tableClasses,
    tableSizeClasses,
    toolbarClasses,
    filterPanelClasses,
    paginationClasses,
    exportDropdownClasses,
  } from './table.types';
  import {
    processTableData,
    toggleSort,
    getNestedValue,
    calculatePagination,
    getPageNumbers,
    getFilterOperators,
  } from './table.logic';
  import {
    exportData,
    getExportFormatLabel,
    type ExportFormat,
  } from './table.export';
  import type { Size, FilterValue, PaginationState, FilterOperator } from '../types';

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

  // DataGrid specific props
  export let filterable: boolean = true;
  export let filters: FilterValue[] = [];
  export let searchable: boolean = true;
  export let searchQuery: string = '';
  export let paginated: boolean = true;
  export let pagination: PaginationState = { page: 1, pageSize: 10, total: 0 };
  export let pageSizes: number[] = [10, 25, 50, 100];
  export let exportable: boolean = true;
  export let exportFormats: ExportFormat[] = ['csv', 'xlsx', 'pdf'];
  export let exportFilename: string = 'export';
  export let columnToggle: boolean = false;
  export let toolbarPosition: 'top' | 'bottom' | 'both' = 'top';
  export let id: string = uid('datagrid');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    sort: { column: string; direction: 'asc' | 'desc' | null };
    select: { keys: (string | number)[]; rows: T[] };
    rowClick: { row: T; index: number };
    filter: { filters: FilterValue[] };
    search: { query: string };
    pageChange: { page: number; pageSize: number };
    export: { format: ExportFormat };
  }>();

  // Internal state
  let showFilterPanel = false;
  let showExportDropdown = false;
  let showColumnToggle = false;
  let exportDropdownRef: HTMLDivElement;

  // Initialize column visibility
  $: visibleColumns = columns.filter(c => c.visible !== false);
  $: filterableColumns = columns.filter(c => c.filterable !== false);
  $: sizeConfig = tableSizeClasses[size];

  // Process data through all transformations
  // Note: Only pass page/pageSize to avoid circular dependency with pagination.total
  $: processedResult = processTableData(data, columns, {
    searchQuery,
    filters,
    sortState,
    pagination: paginated ? { page: pagination.page, pageSize: pagination.pageSize, total: 0 } : undefined,
  });

  $: displayData = processedResult.data;
  $: totalItems = processedResult.total;

  $: paginationInfo = paginated
    ? calculatePagination(totalItems, pagination.page, pagination.pageSize)
    : null;

  $: pageNumbers = paginationInfo
    ? getPageNumbers(pagination.page, paginationInfo.totalPages)
    : [];

  $: wrapperStyle = maxHeight ? `max-height: ${maxHeight}; overflow: auto;` : '';

  $: isAllSelected = displayData.length > 0 && selectedKeys.length === displayData.length;
  $: isSomeSelected = selectedKeys.length > 0 && selectedKeys.length < displayData.length;

  // Helpers
  function getRowKey(row: T, index: number): string | number {
    const key = getNestedValue(row, rowKey);
    return key !== undefined ? (key as string | number) : index;
  }

  function isRowSelected(row: T, index: number): boolean {
    return selectedKeys.includes(getRowKey(row, index));
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
      selectedKeys = displayData.map((row, index) => getRowKey(row, index));
    }
    dispatch('select', {
      keys: selectedKeys,
      rows: displayData.filter((row, index) => selectedKeys.includes(getRowKey(row, index))),
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
      rows: displayData.filter((r, i) => selectedKeys.includes(getRowKey(r, i))),
    });
  }

  function handleRowClick(row: T, index: number) {
    dispatch('rowClick', { row, index });
  }

  function handleSearch(event: Event) {
    const target = event.target as HTMLInputElement;
    searchQuery = target.value;
    pagination = { ...pagination, page: 1 }; // Reset to first page
    dispatch('search', { query: searchQuery });
  }

  function handlePageChange(newPage: number) {
    pagination = { ...pagination, page: newPage };
    dispatch('pageChange', { page: newPage, pageSize: pagination.pageSize });
  }

  function handlePageSizeChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    const newPageSize = parseInt(target.value);
    pagination = { ...pagination, pageSize: newPageSize, page: 1 };
    dispatch('pageChange', { page: 1, pageSize: newPageSize });
  }

  // Filter handlers
  function addFilter() {
    const defaultColumn = filterableColumns[0];
    if (!defaultColumn) return;

    filters = [
      ...filters,
      {
        column: defaultColumn.key,
        operator: 'contains' as FilterOperator,
        value: '',
      },
    ];
  }

  function removeFilter(index: number) {
    filters = filters.filter((_, i) => i !== index);
    dispatch('filter', { filters });
  }

  function updateFilter(index: number, updates: Partial<FilterValue>) {
    filters = filters.map((f, i) => (i === index ? { ...f, ...updates } : f));
    pagination = { ...pagination, page: 1 };
    dispatch('filter', { filters });
  }

  function clearFilters() {
    filters = [];
    dispatch('filter', { filters });
  }

  // Export handler
  async function handleExport(format: ExportFormat) {
    showExportDropdown = false;
    try {
      // Export all filtered/searched data (not just current page)
      const { data: exportableData } = processTableData(data, columns, {
        searchQuery,
        filters,
        sortState,
        // No pagination - export all data
      });

      await exportData({
        data: exportableData,
        columns,
        filename: exportFilename,
        format,
        title: exportFilename,
      });
      dispatch('export', { format });
    } catch (error) {
      console.error('Export failed:', error);
    }
  }

  // Column visibility toggle
  function toggleColumnVisibility(columnKey: string) {
    columns = columns.map(c =>
      c.key === columnKey ? { ...c, visible: c.visible === false ? true : !c.visible } : c
    );
  }

  // Click outside handlers
  function handleClickOutside(event: MouseEvent) {
    if (exportDropdownRef && !exportDropdownRef.contains(event.target as Node)) {
      showExportDropdown = false;
    }
  }

  onMount(() => {
    document.addEventListener('click', handleClickOutside);
  });

  onDestroy(() => {
    document.removeEventListener('click', handleClickOutside);
  });

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
</script>

<div class={cn('bg-neutral-white rounded-lg border border-neutral-200', fullWidth && 'w-full', className)} {id} data-testid={testId || undefined}>
  <!-- Top Toolbar -->
  {#if toolbarPosition === 'top' || toolbarPosition === 'both'}
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

        {#if filterable}
          <button
            type="button"
            class={cn(toolbarClasses.button, toolbarClasses.buttonSecondary)}
            on:click={() => (showFilterPanel = !showFilterPanel)}
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
            </svg>
            Filter
            {#if filters.length > 0}
              <span class={toolbarClasses.filterBadge}>{filters.length}</span>
            {/if}
          </button>
        {/if}
      </div>

      <div class={toolbarClasses.right}>
        {#if columnToggle}
          <div class="relative">
            <button
              type="button"
              class={cn(toolbarClasses.button, toolbarClasses.buttonSecondary)}
              on:click={() => (showColumnToggle = !showColumnToggle)}
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2" />
              </svg>
              Columns
            </button>
            {#if showColumnToggle}
              <div class={exportDropdownClasses.dropdown}>
                {#each columns as column}
                  <label class="flex items-center gap-2 px-4 py-2 hover:bg-neutral-50 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={column.visible !== false}
                      on:change={() => toggleColumnVisibility(column.key)}
                    />
                    <span class="text-sm">{column.header}</span>
                  </label>
                {/each}
              </div>
            {/if}
          </div>
        {/if}

        {#if exportable}
          <div class={exportDropdownClasses.container} bind:this={exportDropdownRef}>
            <button
              type="button"
              class={exportDropdownClasses.button}
              on:click|stopPropagation={() => (showExportDropdown = !showExportDropdown)}
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              Export
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            {#if showExportDropdown}
              <div class={exportDropdownClasses.dropdown}>
                {#each exportFormats as format}
                  <button
                    type="button"
                    class={exportDropdownClasses.item}
                    on:click={() => handleExport(format)}
                  >
                    {getExportFormatLabel(format)}
                  </button>
                {/each}
              </div>
            {/if}
          </div>
        {/if}

        <slot name="toolbar-actions" />
      </div>
    </div>
  {/if}

  <!-- Filter Panel -->
  {#if filterable && showFilterPanel}
    <div class={filterPanelClasses.container}>
      {#each filters as filter, index}
        {@const filterColumn = columns.find(c => c.key === filter.column)}
        <div class={filterPanelClasses.row}>
          <select
            class={filterPanelClasses.select}
            value={filter.column}
            on:change={(e) => updateFilter(index, { column: e.currentTarget.value })}
          >
            {#each filterableColumns as column}
              <option value={column.key}>{column.header}</option>
            {/each}
          </select>

          <select
            class={filterPanelClasses.select}
            value={filter.operator}
            on:change={(e) => updateFilter(index, { operator: e.currentTarget.value as FilterOperator })}
          >
            {#each getFilterOperators(filterColumn?.filterType || 'text') as op}
              <option value={op.value}>{op.label}</option>
            {/each}
          </select>

          {#if filter.operator !== 'isEmpty' && filter.operator !== 'isNotEmpty'}
            <input
              type="text"
              class={filterPanelClasses.input}
              placeholder="Value"
              value={filter.value}
              on:input={(e) => updateFilter(index, { value: e.currentTarget.value })}
            />

            {#if filter.operator === 'between'}
              <input
                type="text"
                class={filterPanelClasses.input}
                placeholder="Second value"
                value={filter.secondValue || ''}
                on:input={(e) => updateFilter(index, { secondValue: e.currentTarget.value })}
              />
            {/if}
          {/if}

          <button
            type="button"
            class={filterPanelClasses.removeBtn}
            on:click={() => removeFilter(index)}
            aria-label="Remove filter"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      {/each}

      <div class="flex items-center gap-4">
        <button type="button" class={filterPanelClasses.addBtn} on:click={addFilter}>
          + Add filter
        </button>
        {#if filters.length > 0}
          <button
            type="button"
            class="text-sm text-neutral-500 hover:text-neutral-700"
            on:click={clearFilters}
          >
            Clear all
          </button>
        {/if}
      </div>
    </div>
  {/if}

  <!-- Table -->
  <div class={tableClasses.wrapper} style={wrapperStyle}>
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
        {#if displayData.length === 0}
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
          {#each displayData as row, index (getRowKey(row, index))}
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

  <!-- Pagination -->
  {#if paginated && paginationInfo}
    <div class={paginationClasses.container}>
      <div class={paginationClasses.info}>
        Showing {paginationInfo.startItem} to {paginationInfo.endItem} of {paginationInfo.total} entries
      </div>

      <div class={paginationClasses.controls}>
        <select
          class={paginationClasses.pageSizeSelect}
          value={pagination.pageSize}
          on:change={handlePageSizeChange}
        >
          {#each pageSizes as pageSize}
            <option value={pageSize}>{pageSize} / page</option>
          {/each}
        </select>

        <button
          type="button"
          class={paginationClasses.button}
          disabled={pagination.page === 1}
          on:click={() => handlePageChange(1)}
          aria-label="First page"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
          </svg>
        </button>

        <button
          type="button"
          class={paginationClasses.button}
          disabled={pagination.page === 1}
          on:click={() => handlePageChange(pagination.page - 1)}
          aria-label="Previous page"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
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
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </button>

        <button
          type="button"
          class={paginationClasses.button}
          disabled={pagination.page >= paginationInfo.totalPages}
          on:click={() => handlePageChange(paginationInfo.totalPages)}
          aria-label="Last page"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 5l7 7-7 7M5 5l7 7-7 7" />
          </svg>
        </button>
      </div>
    </div>
  {/if}
</div>
