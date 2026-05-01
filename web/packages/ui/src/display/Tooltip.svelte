<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import type { TooltipProps } from './display.types';
  import { tooltipClasses } from './display.types';
  import { cn } from '../utils';

  type $$Props = TooltipProps;

  export let content: $$Props['content'];
  export let position: $$Props['position'] = 'top';
  export let trigger: $$Props['trigger'] = 'hover';
  export let delay: $$Props['delay'] = 200;
  export let disabled: $$Props['disabled'] = false;
  export let maxWidth: $$Props['maxWidth'] = '250px';
  let className: $$Props['class'] = undefined;
  export { className as class };

  let visible = false;
  let showTimeout: ReturnType<typeof setTimeout> | null = null;
  let hideTimeout: ReturnType<typeof setTimeout> | null = null;
  let triggerRef: HTMLDivElement;
  let tooltipRef: HTMLDivElement;

  function updatePosition() {
    if (!triggerRef || !tooltipRef) return;

    const triggerRect = triggerRef.getBoundingClientRect();
    const tooltipRect = tooltipRef.getBoundingClientRect();
    const offset = 8;

    let top = 0;
    let left = 0;

    switch (position) {
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

  function toggle() {
    if (disabled || !content) return;
    if (visible) {
      hide();
    } else {
      show();
    }
  }

  function handleMouseEnter() {
    if (trigger === 'hover') {
      show();
    }
  }

  function handleMouseLeave() {
    if (trigger === 'hover') {
      hide();
    }
  }

  function handleClick() {
    if (trigger === 'click') {
      toggle();
    }
  }

  function handleFocus() {
    if (trigger === 'focus' || trigger === 'hover') {
      show();
    }
  }

  function handleBlur() {
    if (trigger === 'focus' || trigger === 'hover') {
      hide();
    }
  }

  function handleClickOutside(e: MouseEvent) {
    if (trigger === 'click' && visible && triggerRef && !triggerRef.contains(e.target as Node)) {
      hide();
    }
  }

  function handleEscape(e: KeyboardEvent) {
    if (e.key === 'Escape' && visible) {
      hide();
    }
  }

  onMount(() => {
    if (typeof window !== 'undefined') {
      window.addEventListener('scroll', updatePosition);
      window.addEventListener('resize', updatePosition);
      document.addEventListener('click', handleClickOutside);
      document.addEventListener('keydown', handleEscape);
    }
  });

  onDestroy(() => {
    if (showTimeout) clearTimeout(showTimeout);
    if (hideTimeout) clearTimeout(hideTimeout);
    if (typeof window !== 'undefined') {
      window.removeEventListener('scroll', updatePosition);
      window.removeEventListener('resize', updatePosition);
      document.removeEventListener('click', handleClickOutside);
      document.removeEventListener('keydown', handleEscape);
    }
  });

  $: if (visible) {
    requestAnimationFrame(updatePosition);
  }

  // Arrow position classes
  const arrowPositionClasses = {
    top: 'bottom-[-4px] left-1/2 -translate-x-1/2',
    bottom: 'top-[-4px] left-1/2 -translate-x-1/2',
    left: 'right-[-4px] top-1/2 -translate-y-1/2',
    right: 'left-[-4px] top-1/2 -translate-y-1/2',
  };
</script>

<div
  bind:this={triggerRef}
  class={cn(tooltipClasses.trigger, className)}
  on:mouseenter={handleMouseEnter}
  on:mouseleave={handleMouseLeave}
  on:click={handleClick}
  on:focus={handleFocus}
  on:blur={handleBlur}
  role="button"
  tabindex="0"
>
  <slot />
</div>

{#if visible && content}
  <div
    bind:this={tooltipRef}
    class={cn(
      'fixed z-tooltip px-3 py-1.5 text-sm',
      'bg-neutral-800 text-neutral-white',
      'rounded-md shadow-lg',
      'pointer-events-none',
      'animate-tooltip'
    )}
    style="max-width: {maxWidth};"
    role="tooltip"
  >
    {content}
    <div
      class={cn(
        'absolute w-2 h-2 bg-neutral-800 rotate-45',
        arrowPositionClasses[position || 'top']
      )}
    ></div>
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
