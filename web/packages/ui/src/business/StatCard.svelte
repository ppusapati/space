<script context="module" lang="ts">
  export type TrendDirection = 'up' | 'down' | 'neutral';
</script>

<script lang="ts">
  import { cn } from '../utils/classnames';

  // Props
  export let title: string;
  export let value: string | number;
  export let previousValue: string | number | null = null;
  export let trend: TrendDirection | null = null;
  export let trendValue: string = '';
  export let trendLabel: string = '';
  export let icon: string = '';
  export let iconColor: string = 'var(--color-interactive-primary)';
  export let loading: boolean = false;
  export let href: string = '';

  let className: string = '';
  export { className as class };

  $: computedTrend = trend ?? (previousValue !== null ? calculateTrend() : null);

  function calculateTrend(): TrendDirection {
    if (previousValue === null) return 'neutral';
    const current = typeof value === 'string' ? parseFloat(value.replace(/[^0-9.-]/g, '')) : value;
    const previous = typeof previousValue === 'string' ? parseFloat(previousValue.replace(/[^0-9.-]/g, '')) : previousValue;

    if (current > previous) return 'up';
    if (current < previous) return 'down';
    return 'neutral';
  }

  const trendConfig = {
    up: {
      color: 'text-[var(--color-success)]',
      bgColor: 'bg-[var(--color-success)]/10',
      icon: `<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"/></svg>`,
    },
    down: {
      color: 'text-[var(--color-error)]',
      bgColor: 'bg-[var(--color-error)]/10',
      icon: `<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 17h8m0 0v-8m0 8l-8-8-4 4-6-6"/></svg>`,
    },
    neutral: {
      color: 'text-[var(--color-text-tertiary)]',
      bgColor: 'bg-[var(--color-neutral-200)]',
      icon: `<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14"/></svg>`,
    },
  };
</script>

<svelte:element
  this={href ? 'a' : 'div'}
  href={href || undefined}
  class={cn(
    'block p-5 rounded-xl',
    'bg-[var(--color-surface-primary)]',
    'border border-[var(--color-border-primary)]',
    href && 'hover:border-[var(--color-interactive-primary)] hover:shadow-md transition-all',
    className
  )}
>
  <div class="flex items-start justify-between">
    <div class="flex-1">
      <!-- Title -->
      <p class="text-sm font-medium text-[var(--color-text-secondary)]">
        {title}
      </p>

      <!-- Value -->
      {#if loading}
        <div class="mt-2 h-8 w-24 bg-[var(--color-surface-tertiary)] rounded animate-pulse" />
      {:else}
        <p class="mt-2 text-2xl font-bold text-[var(--color-text-primary)] tabular-nums">
          {value}
        </p>
      {/if}

      <!-- Trend -->
      {#if computedTrend && (trendValue || trendLabel)}
        <div class="flex items-center gap-2 mt-2">
          <span
            class={cn(
              'inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium',
              trendConfig[computedTrend].bgColor,
              trendConfig[computedTrend].color
            )}
          >
            {@html trendConfig[computedTrend].icon}
            {trendValue}
          </span>
          {#if trendLabel}
            <span class="text-xs text-[var(--color-text-tertiary)]">
              {trendLabel}
            </span>
          {/if}
        </div>
      {/if}
    </div>

    <!-- Icon -->
    {#if icon}
      <div
        class="p-3 rounded-lg"
        style="background-color: {iconColor}20;"
      >
        <span style="color: {iconColor};">
          {@html icon}
        </span>
      </div>
    {/if}
  </div>

  <!-- Slot for additional content -->
  <slot />
</svelte:element>
