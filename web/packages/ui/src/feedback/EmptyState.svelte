<script lang="ts">
  import { cn } from '../utils/classnames';
  import type { Size } from '../types';

  // Props
  export let title: string = 'No data';
  export let description: string = '';
  export let icon: 'empty' | 'search' | 'error' | 'folder' | 'inbox' | 'custom' = 'empty';
  export let size: Size = 'md';
  export let actionLabel: string = '';
  export let secondaryActionLabel: string = '';

  let className: string = '';
  export { className as class };

  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher<{
    action: void;
    secondaryAction: void;
  }>();

  const sizeConfig = {
    sm: { icon: 'w-12 h-12', title: 'text-base', desc: 'text-sm', gap: 'gap-2' },
    md: { icon: 'w-16 h-16', title: 'text-lg', desc: 'text-base', gap: 'gap-3' },
    lg: { icon: 'w-24 h-24', title: 'text-xl', desc: 'text-lg', gap: 'gap-4' },
  };

  $: config = sizeConfig[size as keyof typeof sizeConfig] ?? sizeConfig.md;

  const icons = {
    empty: `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
    </svg>`,
    search: `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
    </svg>`,
    error: `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
    </svg>`,
    folder: `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
    </svg>`,
    inbox: `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
    </svg>`,
  };
</script>

<div class={cn('flex flex-col items-center justify-center text-center p-8', config.gap, className)}>
  <!-- Icon -->
  <div class={cn('text-[var(--color-text-tertiary)]', config.icon)}>
    {#if icon === 'custom'}
      <slot name="icon" />
    {:else}
      {@html icons[icon]}
    {/if}
  </div>

  <!-- Title -->
  <h3 class={cn('font-semibold text-[var(--color-text-primary)]', config.title)}>
    {title}
  </h3>

  <!-- Description -->
  {#if description}
    <p class={cn('text-[var(--color-text-secondary)] max-w-md', config.desc)}>
      {description}
    </p>
  {/if}

  <!-- Custom content slot -->
  <slot />

  <!-- Actions -->
  {#if actionLabel || secondaryActionLabel}
    <div class="flex items-center gap-3 mt-4">
      {#if actionLabel}
        <button
          type="button"
          class={cn(
            'px-4 py-2 rounded-lg font-medium',
            'bg-[var(--color-interactive-primary)] text-white',
            'hover:bg-[var(--color-interactive-primary-hover)]',
            'focus:outline-none focus:ring-2 focus:ring-[var(--color-interactive-primary)] focus:ring-offset-2'
          )}
          on:click={() => dispatch('action')}
        >
          {actionLabel}
        </button>
      {/if}

      {#if secondaryActionLabel}
        <button
          type="button"
          class={cn(
            'px-4 py-2 rounded-lg font-medium',
            'border border-[var(--color-border-primary)]',
            'text-[var(--color-text-primary)]',
            'hover:bg-[var(--color-surface-secondary)]',
            'focus:outline-none focus:ring-2 focus:ring-[var(--color-interactive-primary)] focus:ring-offset-2'
          )}
          on:click={() => dispatch('secondaryAction')}
        >
          {secondaryActionLabel}
        </button>
      {/if}
    </div>
  {/if}
</div>
