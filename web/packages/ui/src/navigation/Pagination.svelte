<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import { paginationComponentClasses, paginationSizeClasses } from './navigation.types';
  import { getPageNumbers } from '../tables/table.logic';
  import type { Size } from '../types';

  // Props
  export let page: number = 1;
  export let totalPages: number = 1;
  export let pageSize: number = 10;
  export let totalItems: number = 0;
  export let showFirstLast: boolean = true;
  export let showPageSize: boolean = false;
  export let pageSizes: number[] = [10, 25, 50, 100];
  export let size: Size = 'md';
  export let variant: 'default' | 'simple' | 'minimal' = 'default';
  export let id: string = uid('pagination');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: { page: number; pageSize: number };
  }>();

  $: sizeConfig = paginationSizeClasses[size];
  $: pageNumbers = variant !== 'minimal' ? getPageNumbers(page, totalPages, 7) : [];
  $: startItem = totalItems > 0 ? (page - 1) * pageSize + 1 : 0;
  $: endItem = Math.min(page * pageSize, totalItems);

  function goToPage(newPage: number) {
    if (newPage < 1 || newPage > totalPages || newPage === page) return;
    page = newPage;
    dispatch('change', { page, pageSize });
  }

  function handlePageSizeChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    pageSize = parseInt(target.value);
    page = 1; // Reset to first page
    dispatch('change', { page, pageSize });
  }
</script>

<nav
  {id}
  class={cn(paginationComponentClasses.container, sizeConfig.text, className)}
  data-testid={testId || undefined}
  aria-label="Pagination"
>
  {#if showPageSize}
    <div class="flex items-center gap-2">
      <span class={cn(paginationComponentClasses.info, sizeConfig.text)}>Show</span>
      <select
        class={paginationComponentClasses.select}
        value={pageSize}
        on:change={handlePageSizeChange}
      >
        {#each pageSizes as ps}
          <option value={ps}>{ps}</option>
        {/each}
      </select>
    </div>
  {/if}

  {#if variant !== 'minimal' && totalItems > 0}
    <span class={paginationComponentClasses.info}>
      {startItem}-{endItem} of {totalItems}
    </span>
  {/if}

  <div class="flex items-center gap-1">
    {#if showFirstLast && variant === 'default'}
      <button
        type="button"
        class={cn(paginationComponentClasses.button, sizeConfig.button)}
        disabled={page === 1}
        on:click={() => goToPage(1)}
        aria-label="First page"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
        </svg>
      </button>
    {/if}

    <button
      type="button"
      class={cn(paginationComponentClasses.button, sizeConfig.button)}
      disabled={page === 1}
      on:click={() => goToPage(page - 1)}
      aria-label="Previous page"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
      </svg>
    </button>

    {#if variant === 'default'}
      {#each pageNumbers as num}
        {#if num === 'ellipsis'}
          <span class={paginationComponentClasses.ellipsis}>...</span>
        {:else}
          <button
            type="button"
            class={cn(
              paginationComponentClasses.pageButton,
              sizeConfig.button,
              page === num
                ? paginationComponentClasses.pageButtonActive
                : paginationComponentClasses.pageButtonInactive
            )}
            on:click={() => goToPage(num)}
            aria-label={`Page ${num}`}
            aria-current={page === num ? 'page' : undefined}
          >
            {num}
          </button>
        {/if}
      {/each}
    {:else if variant === 'simple'}
      <span class={cn(paginationComponentClasses.info, 'mx-2')}>
        Page {page} of {totalPages}
      </span>
    {:else}
      <span class={cn(paginationComponentClasses.info, 'mx-2')}>
        {page} / {totalPages}
      </span>
    {/if}

    <button
      type="button"
      class={cn(paginationComponentClasses.button, sizeConfig.button)}
      disabled={page >= totalPages}
      on:click={() => goToPage(page + 1)}
      aria-label="Next page"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
      </svg>
    </button>

    {#if showFirstLast && variant === 'default'}
      <button
        type="button"
        class={cn(paginationComponentClasses.button, sizeConfig.button)}
        disabled={page >= totalPages}
        on:click={() => goToPage(totalPages)}
        aria-label="Last page"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 5l7 7-7 7M5 5l7 7-7 7" />
        </svg>
      </button>
    {/if}
  </div>
</nav>
