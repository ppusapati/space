<script context="module" lang="ts">
  export type Placement = 'top' | 'bottom' | 'left' | 'right' |
                          'top-start' | 'top-end' | 'bottom-start' | 'bottom-end' |
                          'left-start' | 'left-end' | 'right-start' | 'right-end';
</script>

<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { cn } from '../utils/classnames';

  // Props
  export let open: boolean = false;
  export let placement: Placement = 'bottom';
  export let trigger: 'click' | 'hover' | 'manual' = 'click';
  export let closeOnClickOutside: boolean = true;
  export let closeOnEscape: boolean = true;
  export let offset: number = 8;
  export let showArrow: boolean = true;
  export let disabled: boolean = false;

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    open: void;
    close: void;
  }>();

  let triggerRef: HTMLDivElement;
  let contentRef: HTMLDivElement;
  let arrowRef: HTMLDivElement;

  let hoverTimeout: ReturnType<typeof setTimeout> | null = null;

  function updatePosition() {
    if (!triggerRef || !contentRef) return;

    const triggerRect = triggerRef.getBoundingClientRect();
    const contentRect = contentRef.getBoundingClientRect();

    let top = 0;
    let left = 0;
    let arrowTop = '';
    let arrowLeft = '';
    let arrowRotation = 0;

    const [mainPlacement, alignment] = placement.split('-') as [string, string | undefined];

    switch (mainPlacement) {
      case 'top':
        top = triggerRect.top - contentRect.height - offset;
        left = triggerRect.left + (triggerRect.width - contentRect.width) / 2;
        arrowTop = '100%';
        arrowLeft = '50%';
        arrowRotation = 180;
        break;
      case 'bottom':
        top = triggerRect.bottom + offset;
        left = triggerRect.left + (triggerRect.width - contentRect.width) / 2;
        arrowTop = '-6px';
        arrowLeft = '50%';
        arrowRotation = 0;
        break;
      case 'left':
        top = triggerRect.top + (triggerRect.height - contentRect.height) / 2;
        left = triggerRect.left - contentRect.width - offset;
        arrowTop = '50%';
        arrowLeft = '100%';
        arrowRotation = 90;
        break;
      case 'right':
        top = triggerRect.top + (triggerRect.height - contentRect.height) / 2;
        left = triggerRect.right + offset;
        arrowTop = '50%';
        arrowLeft = '-6px';
        arrowRotation = -90;
        break;
    }

    // Handle alignment
    if (alignment === 'start') {
      if (mainPlacement === 'top' || mainPlacement === 'bottom') {
        left = triggerRect.left;
        arrowLeft = `${Math.min(triggerRect.width / 2, 24)}px`;
      } else {
        top = triggerRect.top;
        arrowTop = `${Math.min(triggerRect.height / 2, 24)}px`;
      }
    } else if (alignment === 'end') {
      if (mainPlacement === 'top' || mainPlacement === 'bottom') {
        left = triggerRect.right - contentRect.width;
        arrowLeft = `calc(100% - ${Math.min(triggerRect.width / 2, 24)}px)`;
      } else {
        top = triggerRect.bottom - contentRect.height;
        arrowTop = `calc(100% - ${Math.min(triggerRect.height / 2, 24)}px)`;
      }
    }

    // Apply position
    contentRef.style.top = `${top + window.scrollY}px`;
    contentRef.style.left = `${left + window.scrollX}px`;

    // Apply arrow position
    if (arrowRef && showArrow) {
      arrowRef.style.top = arrowTop;
      arrowRef.style.left = arrowLeft;
      arrowRef.style.transform = `translate(-50%, -50%) rotate(${arrowRotation}deg)`;
    }
  }

  function openPopover() {
    if (disabled) return;
    open = true;
    dispatch('open');
    requestAnimationFrame(updatePosition);
  }

  function closePopover() {
    open = false;
    dispatch('close');
  }

  function togglePopover() {
    if (open) {
      closePopover();
    } else {
      openPopover();
    }
  }

  function handleTriggerClick() {
    if (trigger === 'click') {
      togglePopover();
    }
  }

  function handleMouseEnter() {
    if (trigger === 'hover') {
      if (hoverTimeout) clearTimeout(hoverTimeout);
      openPopover();
    }
  }

  function handleMouseLeave() {
    if (trigger === 'hover') {
      hoverTimeout = setTimeout(() => {
        closePopover();
      }, 100);
    }
  }

  function handleClickOutside(event: MouseEvent) {
    if (!closeOnClickOutside || !open) return;

    const target = event.target as HTMLElement;
    if (!triggerRef?.contains(target) && !contentRef?.contains(target)) {
      closePopover();
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (closeOnEscape && event.key === 'Escape' && open) {
      closePopover();
    }
  }

  onMount(() => {
    document.addEventListener('click', handleClickOutside);
    document.addEventListener('keydown', handleKeydown);
    window.addEventListener('scroll', updatePosition);
    window.addEventListener('resize', updatePosition);
  });

  onDestroy(() => {
    document.removeEventListener('click', handleClickOutside);
    document.removeEventListener('keydown', handleKeydown);
    window.removeEventListener('scroll', updatePosition);
    window.removeEventListener('resize', updatePosition);
    if (hoverTimeout) clearTimeout(hoverTimeout);
  });

  // Update position when open changes
  $: if (open) {
    requestAnimationFrame(updatePosition);
  }
</script>

<div
  bind:this={triggerRef}
  class="inline-block"
  on:click={handleTriggerClick}
  on:mouseenter={handleMouseEnter}
  on:mouseleave={handleMouseLeave}
>
  <slot name="trigger" />
</div>

{#if open}
  <div
    bind:this={contentRef}
    class={cn(
      'fixed z-50',
      'bg-[var(--color-surface-primary)]',
      'border border-[var(--color-border-primary)]',
      'rounded-lg shadow-lg',
      'animate-in fade-in zoom-in-95 duration-200',
      className
    )}
    on:mouseenter={handleMouseEnter}
    on:mouseleave={handleMouseLeave}
    role="dialog"
    aria-modal="false"
  >
    {#if showArrow}
      <div
        bind:this={arrowRef}
        class="absolute w-3 h-3 bg-[var(--color-surface-primary)] border-l border-t border-[var(--color-border-primary)] -z-10"
        style="clip-path: polygon(0 0, 100% 0, 0 100%);"
      />
    {/if}
    <slot />
  </div>
{/if}

<style>
  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes zoomIn {
    from { transform: scale(0.95); }
    to { transform: scale(1); }
  }

  .animate-in {
    animation: fadeIn 0.2s ease-out, zoomIn 0.2s ease-out;
  }
</style>
