<script context="module" lang="ts">
  export interface TimelineItem {
    id: string;
    title: string;
    description?: string;
    date: string | Date;
    icon?: string;
    color?: 'primary' | 'success' | 'warning' | 'error' | 'info' | 'neutral';
    status?: 'completed' | 'current' | 'upcoming';
    metadata?: Record<string, string>;
  }
</script>

<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import type { Size } from '../types';

  // Props
  export let items: TimelineItem[] = [];
  export let size: Size = 'md';
  export let orientation: 'vertical' | 'horizontal' = 'vertical';
  export let alternating: boolean = false;
  export let showConnector: boolean = true;
  export let dateFormat: Intl.DateTimeFormatOptions = {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  };

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    click: { item: TimelineItem };
  }>();

  // Size configurations
  const sizeConfig = {
    sm: { dot: 'w-3 h-3', line: 'w-0.5', title: 'text-sm', desc: 'text-xs', gap: 'gap-4' },
    md: { dot: 'w-4 h-4', line: 'w-0.5', title: 'text-base', desc: 'text-sm', gap: 'gap-6' },
    lg: { dot: 'w-5 h-5', line: 'w-1', title: 'text-lg', desc: 'text-base', gap: 'gap-8' },
  };

  const colorMap = {
    primary: 'bg-[var(--color-interactive-primary)]',
    success: 'bg-[var(--color-success)]',
    warning: 'bg-[var(--color-warning)]',
    error: 'bg-[var(--color-error)]',
    info: 'bg-[var(--color-info)]',
    neutral: 'bg-[var(--color-neutral-400)]',
  };

  $: config = sizeConfig[size as keyof typeof sizeConfig] ?? sizeConfig.md;

  function formatDate(date: string | Date): string {
    const d = typeof date === 'string' ? new Date(date) : date;
    return d.toLocaleDateString('en-US', dateFormat);
  }

  function handleItemClick(item: TimelineItem) {
    dispatch('click', { item });
  }

  function getItemPosition(index: number): 'left' | 'right' | 'center' {
    if (!alternating) return 'right';
    return index % 2 === 0 ? 'left' : 'right';
  }
</script>

<div
  class={cn(
    'timeline',
    orientation === 'horizontal' ? 'flex overflow-x-auto' : 'flex flex-col',
    config.gap,
    className
  )}
>
  {#each items as item, index (item.id)}
    {@const position = getItemPosition(index)}
    {@const isLast = index === items.length - 1}

    <div
      class={cn(
        'timeline-item relative',
        orientation === 'horizontal' ? 'flex flex-col items-center flex-shrink-0 min-w-[200px]' : 'flex',
        alternating && orientation === 'vertical' && 'flex-row',
        position === 'left' && alternating && 'flex-row-reverse'
      )}
    >
      {#if alternating && orientation === 'vertical'}
        <!-- Left content (for alternating) -->
        <div class={cn('flex-1', position === 'right' ? 'text-right pr-6' : 'text-left pl-6')}>
          {#if position === 'left'}
            <button
              type="button"
              class="w-full text-left focus:outline-none focus:ring-2 focus:ring-[var(--color-interactive-primary)] rounded"
              on:click={() => handleItemClick(item)}
            >
              <span class={cn('font-medium text-[var(--color-text-secondary)]', 'text-sm')}>
                {formatDate(item.date)}
              </span>
            </button>
          {:else}
            <button
              type="button"
              class="w-full text-left focus:outline-none focus:ring-2 focus:ring-[var(--color-interactive-primary)] rounded p-2 -m-2 hover:bg-[var(--color-surface-secondary)] transition-colors"
              on:click={() => handleItemClick(item)}
            >
              <h4 class={cn('font-semibold text-[var(--color-text-primary)]', config.title)}>
                {item.title}
              </h4>
              {#if item.description}
                <p class={cn('text-[var(--color-text-secondary)] mt-1', config.desc)}>
                  {item.description}
                </p>
              {/if}
              {#if item.metadata}
                <div class="flex flex-wrap gap-2 mt-2">
                  {#each Object.entries(item.metadata) as [key, value]}
                    <span class="text-xs px-2 py-0.5 rounded bg-[var(--color-surface-tertiary)] text-[var(--color-text-tertiary)]">
                      {key}: {value}
                    </span>
                  {/each}
                </div>
              {/if}
            </button>
          {/if}
        </div>
      {/if}

      <!-- Timeline dot and connector -->
      <div
        class={cn(
          'timeline-marker flex flex-col items-center',
          orientation === 'horizontal' && 'flex-row'
        )}
      >
        <div
          class={cn(
            'timeline-dot rounded-full flex items-center justify-center z-10',
            config.dot,
            colorMap[item.color || 'primary'],
            item.status === 'current' && 'ring-4 ring-[var(--color-interactive-primary)]/30',
            item.status === 'upcoming' && 'opacity-50'
          )}
        >
          {#if item.icon}
            <span class="text-white text-xs">{item.icon}</span>
          {/if}
        </div>

        {#if showConnector && !isLast}
          <div
            class={cn(
              'timeline-connector',
              orientation === 'vertical' ? cn('h-full min-h-[40px]', config.line) : cn('w-full min-w-[40px] h-0.5'),
              'bg-[var(--color-border-primary)]'
            )}
          />
        {/if}
      </div>

      <!-- Content -->
      {#if alternating && orientation === 'vertical'}
        <!-- Right content (for alternating) -->
        <div class={cn('flex-1', position === 'left' ? 'text-right pr-6' : 'text-left pl-6')}>
          {#if position === 'right'}
            <span class={cn('font-medium text-[var(--color-text-secondary)]', 'text-sm')}>
              {formatDate(item.date)}
            </span>
          {:else}
            <button
              type="button"
              class="w-full text-right focus:outline-none focus:ring-2 focus:ring-[var(--color-interactive-primary)] rounded p-2 -m-2 hover:bg-[var(--color-surface-secondary)] transition-colors"
              on:click={() => handleItemClick(item)}
            >
              <h4 class={cn('font-semibold text-[var(--color-text-primary)]', config.title)}>
                {item.title}
              </h4>
              {#if item.description}
                <p class={cn('text-[var(--color-text-secondary)] mt-1', config.desc)}>
                  {item.description}
                </p>
              {/if}
            </button>
          {/if}
        </div>
      {:else}
        <!-- Standard layout content -->
        <div class={cn('flex-1', orientation === 'vertical' ? 'pl-4 pb-4' : 'pt-4 text-center')}>
          <button
            type="button"
            class="w-full text-left focus:outline-none focus:ring-2 focus:ring-[var(--color-interactive-primary)] rounded p-2 -m-2 hover:bg-[var(--color-surface-secondary)] transition-colors"
            on:click={() => handleItemClick(item)}
          >
            <span class={cn('font-medium text-[var(--color-text-secondary)]', 'text-sm')}>
              {formatDate(item.date)}
            </span>
            <h4 class={cn('font-semibold text-[var(--color-text-primary)] mt-1', config.title)}>
              {item.title}
            </h4>
            {#if item.description}
              <p class={cn('text-[var(--color-text-secondary)] mt-1', config.desc)}>
                {item.description}
              </p>
            {/if}
            {#if item.metadata}
              <div class="flex flex-wrap gap-2 mt-2">
                {#each Object.entries(item.metadata) as [key, value]}
                  <span class="text-xs px-2 py-0.5 rounded bg-[var(--color-surface-tertiary)] text-[var(--color-text-tertiary)]">
                    {key}: {value}
                  </span>
                {/each}
              </div>
            {/if}
          </button>
        </div>
      {/if}
    </div>
  {/each}
</div>
