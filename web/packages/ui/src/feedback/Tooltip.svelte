<script context="module" lang="ts">
  export type TooltipPlacement = 'top' | 'bottom' | 'left' | 'right';
</script>

<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { cn } from '../utils/classnames';

  // Props
  export let content: string = '';
  export let placement: TooltipPlacement = 'top';
  export let delay: number = 200;
  export let disabled: boolean = false;
  export let maxWidth: string = '250px';

  let className: string = '';
  export { className as class };

  let visible = false;
  let triggerRef: HTMLDivElement;
  let tooltipRef: HTMLDivElement;
  let showTimeout: ReturnType<typeof setTimeout> | null = null;
  let hideTimeout: ReturnType<typeof setTimeout> | null = null;

  function updatePosition() {
    if (!triggerRef || !tooltipRef) return;

    const triggerRect = triggerRef.getBoundingClientRect();
    const tooltipRect = tooltipRef.getBoundingClientRect();
    const offset = 8;

    let top = 0;
    let left = 0;

    switch (placement) {
      case 'top':
        top = triggerRect.top - tooltipRect.height - offset;
        left = triggerRect.left + (triggerRect.width - tooltipRect.width) / 2;
        break;
      case 'bottom':
        top = triggerRect.bottom + offset;
        left = triggerRect.left + (triggerRect.width - tooltipRect.width) / 2;
        break;
      case 'left':
        top = triggerRect.top + (triggerRect.height - tooltipRect.height) / 2;
        left = triggerRect.left - tooltipRect.width - offset;
        break;
      case 'right':
        top = triggerRect.top + (triggerRect.height - tooltipRect.height) / 2;
        left = triggerRect.right + offset;
        break;
    }

    // Keep within viewport
    const padding = 8;
    left = Math.max(padding, Math.min(left, window.innerWidth - tooltipRect.width - padding));
    top = Math.max(padding, Math.min(top, window.innerHeight - tooltipRect.height - padding));

    tooltipRef.style.top = `${top + window.scrollY}px`;
    tooltipRef.style.left = `${left + window.scrollX}px`;
  }

  function show() {
    if (disabled || !content) return;

    if (hideTimeout) {
      clearTimeout(hideTimeout);
      hideTimeout = null;
    }

    showTimeout = setTimeout(() => {
      visible = true;
      requestAnimationFrame(updatePosition);
    }, delay);
  }

  function hide() {
    if (showTimeout) {
      clearTimeout(showTimeout);
      showTimeout = null;
    }

    hideTimeout = setTimeout(() => {
      visible = false;
    }, 100);
  }

  onMount(() => {
    window.addEventListener('scroll', updatePosition);
    window.addEventListener('resize', updatePosition);
  });

  onDestroy(() => {
    window.removeEventListener('scroll', updatePosition);
    window.removeEventListener('resize', updatePosition);
    if (showTimeout) clearTimeout(showTimeout);
    if (hideTimeout) clearTimeout(hideTimeout);
  });

  $: if (visible) {
    requestAnimationFrame(updatePosition);
  }
</script>

<div
  bind:this={triggerRef}
  class="inline-block"
  on:mouseenter={show}
  on:mouseleave={hide}
  on:focus={show}
  on:blur={hide}
>
  <slot />
</div>

{#if visible}
  <div
    bind:this={tooltipRef}
    class={cn(
      'fixed z-[9999] px-3 py-1.5 text-sm',
      'bg-[var(--color-neutral-900)] text-white',
      'rounded-md shadow-lg',
      'pointer-events-none',
      'animate-tooltip',
      className
    )}
    style="max-width: {maxWidth};"
    role="tooltip"
  >
    {content}
    <!-- Arrow -->
    <div
      class={cn(
        'absolute w-2 h-2 bg-[var(--color-neutral-900)] rotate-45',
        placement === 'top' && 'bottom-[-4px] left-1/2 -translate-x-1/2',
        placement === 'bottom' && 'top-[-4px] left-1/2 -translate-x-1/2',
        placement === 'left' && 'right-[-4px] top-1/2 -translate-y-1/2',
        placement === 'right' && 'left-[-4px] top-1/2 -translate-y-1/2'
      )}
    />
  </div>
{/if}

<style>
  @keyframes tooltipIn {
    from {
      opacity: 0;
      transform: scale(0.95);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }

  .animate-tooltip {
    animation: tooltipIn 0.15s ease-out;
  }
</style>
