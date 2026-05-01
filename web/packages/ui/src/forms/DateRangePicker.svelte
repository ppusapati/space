<script context="module" lang="ts">
  export interface DateRange {
    start: string;
    end: string;
  }
</script>

<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import type { Size } from '../types';

  // Props
  export let value: DateRange = { start: '', end: '' };
  export let disabled: boolean = false;
  export let size: Size = 'md';
  export let label: string = '';
  export let startPlaceholder: string = 'Start date';
  export let endPlaceholder: string = 'End date';
  export let minDate: string = '';
  export let maxDate: string = '';
  export let name: string = '';
  export let id: string = uid('daterange');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: { value: DateRange };
  }>();

  // Size configurations
  const sizeConfig = {
    sm: { input: 'h-8 text-sm px-2' },
    md: { input: 'h-10 text-base px-3' },
    lg: { input: 'h-12 text-lg px-4' },
  };

  $: config = sizeConfig[size as keyof typeof sizeConfig] ?? sizeConfig.md;

  // Preset ranges
  const presets = [
    { label: 'Today', getValue: () => {
      const today = new Date().toISOString().split('T')[0]!;
      return { start: today, end: today };
    }},
    { label: 'Yesterday', getValue: () => {
      const yesterday = new Date(Date.now() - 86400000).toISOString().split('T')[0]!;
      return { start: yesterday, end: yesterday };
    }},
    { label: 'Last 7 days', getValue: () => {
      const end = new Date().toISOString().split('T')[0]!;
      const start = new Date(Date.now() - 7 * 86400000).toISOString().split('T')[0]!;
      return { start, end };
    }},
    { label: 'Last 30 days', getValue: () => {
      const end = new Date().toISOString().split('T')[0]!;
      const start = new Date(Date.now() - 30 * 86400000).toISOString().split('T')[0]!;
      return { start, end };
    }},
    { label: 'This month', getValue: () => {
      const now = new Date();
      const start = new Date(now.getFullYear(), now.getMonth(), 1).toISOString().split('T')[0]!;
      const end = new Date(now.getFullYear(), now.getMonth() + 1, 0).toISOString().split('T')[0]!;
      return { start, end };
    }},
    { label: 'Last month', getValue: () => {
      const now = new Date();
      const start = new Date(now.getFullYear(), now.getMonth() - 1, 1).toISOString().split('T')[0]!;
      const end = new Date(now.getFullYear(), now.getMonth(), 0).toISOString().split('T')[0]!;
      return { start, end };
    }},
  ];

  let showPresets = false;

  function handleStartChange(event: Event) {
    const target = event.target as HTMLInputElement;
    value = { ...value, start: target.value };

    // Ensure end is after start
    if (value.end && value.start > value.end) {
      value = { ...value, end: value.start };
    }

    dispatch('change', { value });
  }

  function handleEndChange(event: Event) {
    const target = event.target as HTMLInputElement;
    value = { ...value, end: target.value };
    dispatch('change', { value });
  }

  function applyPreset(preset: typeof presets[0]) {
    value = preset.getValue();
    showPresets = false;
    dispatch('change', { value });
  }

  function clearRange() {
    value = { start: '', end: '' };
    dispatch('change', { value });
  }

  function formatDateDisplay(dateStr: string): string {
    if (!dateStr) return '';
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  }

  $: displayValue = value.start && value.end
    ? `${formatDateDisplay(value.start)} - ${formatDateDisplay(value.end)}`
    : '';
</script>

<div class={cn('w-full', className)}>
  {#if label}
    <label
      for={id}
      class="block mb-2 font-medium text-[var(--color-text-primary)]"
    >
      {label}
    </label>
  {/if}

  <div class="relative">
    <!-- Presets dropdown -->
    <div class="mb-2">
      <button
        type="button"
        class="text-sm text-[var(--color-interactive-primary)] hover:underline"
        on:click={() => showPresets = !showPresets}
        {disabled}
      >
        Quick select
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="inline-block h-4 w-4 ml-1 transition-transform"
          class:rotate-180={showPresets}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {#if showPresets}
        <div class="flex flex-wrap gap-2 mt-2">
          {#each presets as preset}
            <button
              type="button"
              class={cn(
                'px-3 py-1 text-sm rounded-full',
                'border border-[var(--color-border-primary)]',
                'hover:bg-[var(--color-surface-secondary)]',
                'focus:outline-none focus:ring-2 focus:ring-[var(--color-interactive-primary)]'
              )}
              on:click={() => applyPreset(preset)}
              {disabled}
            >
              {preset.label}
            </button>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Date inputs -->
    <div class="flex items-center gap-2">
      <div class="flex-1">
        <input
          type="date"
          id="{id}-start"
          name="{name}-start"
          value={value.start}
          min={minDate}
          max={value.end || maxDate}
          {disabled}
          placeholder={startPlaceholder}
          class={cn(
            'w-full rounded-lg border border-[var(--color-border-primary)]',
            'bg-[var(--color-surface-primary)] text-[var(--color-text-primary)]',
            'focus:outline-none focus:ring-2 focus:ring-[var(--color-interactive-primary)]',
            config.input,
            disabled && 'opacity-50 cursor-not-allowed'
          )}
          data-testid={testId ? `${testId}-start` : undefined}
          on:change={handleStartChange}
        />
      </div>

      <span class="text-[var(--color-text-tertiary)]">to</span>

      <div class="flex-1">
        <input
          type="date"
          id="{id}-end"
          name="{name}-end"
          value={value.end}
          min={value.start || minDate}
          max={maxDate}
          {disabled}
          placeholder={endPlaceholder}
          class={cn(
            'w-full rounded-lg border border-[var(--color-border-primary)]',
            'bg-[var(--color-surface-primary)] text-[var(--color-text-primary)]',
            'focus:outline-none focus:ring-2 focus:ring-[var(--color-interactive-primary)]',
            config.input,
            disabled && 'opacity-50 cursor-not-allowed'
          )}
          data-testid={testId ? `${testId}-end` : undefined}
          on:change={handleEndChange}
        />
      </div>

      {#if value.start || value.end}
        <button
          type="button"
          class="p-2 text-[var(--color-text-tertiary)] hover:text-[var(--color-text-primary)]"
          on:click={clearRange}
          {disabled}
          aria-label="Clear date range"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      {/if}
    </div>

    {#if displayValue}
      <div class="mt-2 text-sm text-[var(--color-text-secondary)]">
        Selected: {displayValue}
      </div>
    {/if}
  </div>
</div>
