<script lang="ts">
  import { createEventDispatcher, onDestroy } from 'svelte';
  import { fade, fly } from 'svelte/transition';
  import type { DrawerProps } from './feedback.types';
  import { drawerClasses, drawerSizeClasses } from './feedback.types';
  import { cn } from '../utils';

  type $$Props = DrawerProps;

  export let open: $$Props['open'] = false;
  export let title: $$Props['title'] = undefined;
  export let position: $$Props['position'] = 'right';
  export let size: $$Props['size'] = 'md';
  export let closeOnBackdrop: $$Props['closeOnBackdrop'] = true;
  export let closeOnEscape: $$Props['closeOnEscape'] = true;
  export let showClose: $$Props['showClose'] = true;
  export let overlay: $$Props['overlay'] = true;
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{ close: void }>();

  let drawerElement: HTMLDivElement;
  let previousActiveElement: Element | null = null;

  $: positionClass = drawerClasses[position || 'right'];
  $: sizeClass = drawerSizeClasses[position || 'right'][size || 'md'];

  $: flyParams = {
    left: { x: -300, duration: 200 },
    right: { x: 300, duration: 200 },
    top: { y: -300, duration: 200 },
    bottom: { y: 300, duration: 200 },
  }[position || 'right'];

  function handleClose() {
    dispatch('close');
  }

  function handleBackdropClick(e: MouseEvent) {
    if (closeOnBackdrop && e.target === e.currentTarget) {
      handleClose();
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (closeOnEscape && e.key === 'Escape') {
      handleClose();
    }
    // Trap focus within drawer
    if (e.key === 'Tab' && drawerElement) {
      const focusableElements = drawerElement.querySelectorAll<HTMLElement>(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
      );
      const firstElement = focusableElements[0];
      const lastElement = focusableElements[focusableElements.length - 1];

      if (e.shiftKey && document.activeElement === firstElement) {
        e.preventDefault();
        lastElement?.focus();
      } else if (!e.shiftKey && document.activeElement === lastElement) {
        e.preventDefault();
        firstElement?.focus();
      }
    }
  }

  onDestroy(() => {
    if (typeof document !== 'undefined') {
      document.body.style.overflow = '';
    }
  });

  $: if (open) {
    previousActiveElement = typeof document !== 'undefined' ? document.activeElement : null;
    if (typeof document !== 'undefined') {
      document.body.style.overflow = 'hidden';
    }
  } else {
    if (typeof document !== 'undefined') {
      document.body.style.overflow = '';
    }
    if (previousActiveElement instanceof HTMLElement) {
      previousActiveElement.focus();
    }
  }
</script>

<svelte:window on:keydown={open ? handleKeydown : undefined} />

{#if open}
  {#if overlay}
    <div
      class={drawerClasses.overlay}
      transition:fade={{ duration: 150 }}
      on:click={handleBackdropClick}
      aria-hidden="true"
    />
  {/if}

  <div
    bind:this={drawerElement}
    class={cn(
      drawerClasses.panel,
      positionClass,
      sizeClass,
      'flex flex-col',
      className
    )}
    role="dialog"
    aria-modal="true"
    aria-labelledby={title ? 'drawer-title' : undefined}
    transition:fly={flyParams}
  >
    {#if title || showClose}
      <div class={drawerClasses.header}>
        {#if title}
          <h2 id="drawer-title" class={drawerClasses.title}>{title}</h2>
        {:else}
          <div />
        {/if}
        {#if showClose}
          <button
            type="button"
            class={drawerClasses.closeBtn}
            on:click={handleClose}
            aria-label="Close drawer"
          >
            <svg class="w-5 h-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        {/if}
      </div>
    {/if}

    <div class={drawerClasses.body}>
      <slot />
    </div>

    {#if $$slots.footer}
      <div class={drawerClasses.footer}>
        <slot name="footer" />
      </div>
    {/if}
  </div>
{/if}
