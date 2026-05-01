<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import type { ReportExportFormat } from './report.export';

  // ─── Props ──────────────────────────────────────────────────────────────────
  export let formats: ReportExportFormat[] = ['csv', 'xlsx', 'pdf', 'print'];
  export let disabled: boolean = false;
  export let loading: boolean = false;
  export let size: 'sm' | 'md' = 'md';
  export let label: string = 'Export';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    export: { format: ReportExportFormat };
  }>();

  let isOpen = false;

  const formatMeta: Record<ReportExportFormat, { label: string; icon: string }> = {
    csv: { label: 'CSV', icon: '📄' },
    xlsx: { label: 'Excel (XLSX)', icon: '📊' },
    pdf: { label: 'PDF', icon: '📕' },
    print: { label: 'Print', icon: '🖨️' },
  };

  function handleSelect(format: ReportExportFormat) {
    isOpen = false;
    dispatch('export', { format });
  }

  function handleClickOutside(e: MouseEvent) {
    if (isOpen) {
      const target = e.target as HTMLElement;
      if (!target.closest('.export-menu')) {
        isOpen = false;
      }
    }
  }
</script>

<svelte:window on:click={handleClickOutside} />

<div class={cn('export-menu', size === 'sm' ? 'export-menu--sm' : '', className)}>
  <button
    class="export-trigger"
    {disabled}
    on:click|stopPropagation={() => (isOpen = !isOpen)}
    aria-haspopup="true"
    aria-expanded={isOpen}
  >
    {#if loading}
      <span class="export-spinner"></span>
    {:else}
      <svg class="export-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
        <polyline points="7 10 12 15 17 10"></polyline>
        <line x1="12" y1="15" x2="12" y2="3"></line>
      </svg>
    {/if}
    <span>{label}</span>
    <svg class="export-caret" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <polyline points="6 9 12 15 18 9"></polyline>
    </svg>
  </button>

  {#if isOpen}
    <div class="export-dropdown" role="menu">
      {#each formats as fmt}
        <button
          class="export-option"
          role="menuitem"
          on:click|stopPropagation={() => handleSelect(fmt)}
        >
          <span class="export-option-icon">{formatMeta[fmt].icon}</span>
          <span class="export-option-label">{formatMeta[fmt].label}</span>
        </button>
      {/each}
    </div>
  {/if}
</div>

<style lang="postcss">
  .export-menu {
    @apply relative inline-flex;
  }

  .export-trigger {
    @apply inline-flex items-center gap-1.5 rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 shadow-sm transition-colors;
    @apply hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1;
    @apply disabled:opacity-50 disabled:cursor-not-allowed;
  }

  .export-menu--sm .export-trigger {
    @apply px-2 py-1 text-xs;
  }

  .export-icon {
    @apply h-4 w-4;
  }

  .export-caret {
    @apply h-3.5 w-3.5 text-gray-400;
  }

  .export-spinner {
    @apply inline-block h-3.5 w-3.5 border-2 border-gray-300 border-t-blue-500 rounded-full animate-spin;
  }

  .export-dropdown {
    @apply absolute right-0 top-full z-50 mt-1 min-w-[160px] rounded-md border border-gray-200 bg-white py-1 shadow-lg;
  }

  .export-option {
    @apply flex w-full items-center gap-2 px-3 py-2 text-sm text-gray-700 transition-colors;
    @apply hover:bg-gray-100;
  }

  .export-option-icon {
    @apply text-base;
  }

  .export-option-label {
    @apply font-medium;
  }
</style>
