<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { NotificationProps } from './feedback.types';
  import { cn } from '../utils';

  type $$Props = NotificationProps;

  export let variant: $$Props['variant'] = 'neutral';
  export let title: $$Props['title'];
  export let description: $$Props['description'] = undefined;
  export let avatar: $$Props['avatar'] = undefined;
  export let timestamp: $$Props['timestamp'] = undefined;
  export let read: $$Props['read'] = false;
  export let dismissible: $$Props['dismissible'] = false;
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{ dismiss: void; click: void }>();

  let visible = true;

  function handleDismiss(e: MouseEvent) {
    e.stopPropagation();
    visible = false;
    dispatch('dismiss');
  }

  function handleClick() {
    dispatch('click');
  }

  const variantColors: Record<string, { dot: string; bg: string }> = {
    primary: { dot: 'bg-brand-primary-500', bg: 'hover:bg-brand-primary-50' },
    secondary: { dot: 'bg-brand-secondary-500', bg: 'hover:bg-brand-secondary-50' },
    success: { dot: 'bg-semantic-success-500', bg: 'hover:bg-semantic-success-50' },
    warning: { dot: 'bg-semantic-warning-500', bg: 'hover:bg-semantic-warning-50' },
    error: { dot: 'bg-semantic-error-500', bg: 'hover:bg-semantic-error-50' },
    info: { dot: 'bg-semantic-info-500', bg: 'hover:bg-semantic-info-50' },
    neutral: { dot: 'bg-neutral-500', bg: 'hover:bg-neutral-50' },
  };

  $: colors = variantColors[variant || 'neutral']!;
</script>

{#if visible}
  <div
    class={cn(
      'flex items-start gap-3 p-4 rounded-lg border border-neutral-200 bg-neutral-white transition-colors cursor-pointer',
      colors.bg,
      !read && 'bg-neutral-50',
      className
    )}
    role="article"
    tabindex="0"
    on:click={handleClick}
    on:keydown={(e) => e.key === 'Enter' && handleClick()}
  >
    {#if avatar}
      <img
        src={avatar}
        alt=""
        class="w-10 h-10 rounded-full shrink-0 object-cover"
      />
    {:else}
      <div class={cn('w-10 h-10 rounded-full shrink-0 flex items-center justify-center bg-neutral-100')}>
        <svg class="w-5 h-5 text-neutral-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
        </svg>
      </div>
    {/if}

    <div class="flex-1 min-w-0">
      <div class="flex items-start justify-between gap-2">
        <p class="text-sm font-semibold text-neutral-900 truncate">{title}</p>
        {#if !read}
          <span class={cn('w-2 h-2 rounded-full shrink-0 mt-1.5', colors.dot)} aria-label="Unread" />
        {/if}
      </div>
      {#if description}
        <p class="text-sm text-neutral-600 mt-0.5 line-clamp-2">{description}</p>
      {/if}
      {#if timestamp}
        <p class="text-xs text-neutral-400 mt-1">{timestamp}</p>
      {/if}
    </div>

    {#if dismissible}
      <button
        type="button"
        class="shrink-0 p-1 rounded hover:bg-neutral-200 transition-colors"
        on:click={handleDismiss}
        aria-label="Dismiss notification"
      >
        <svg class="w-4 h-4 text-neutral-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    {/if}
  </div>
{/if}
