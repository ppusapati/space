<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { fade, scale } from 'svelte/transition';
  import type { ModalProps } from './feedback.types';
  import { modalClasses, modalSizeClasses } from './feedback.types';
  import { cn } from '../utils';

  type $$Props = ModalProps;

  export let open: $$Props['open'] = false;
  export let title: $$Props['title'] = undefined;
  export let size: $$Props['size'] = 'md';
  export let closeOnBackdrop: $$Props['closeOnBackdrop'] = true;
  export let closeOnEscape: $$Props['closeOnEscape'] = true;
  export let showClose: $$Props['showClose'] = true;
  export let centered: $$Props['centered'] = true;
  export let preventScroll: $$Props['preventScroll'] = true;
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{ close: void }>();

  let modalElement: HTMLDivElement;
  let previousActiveElement: Element | null = null;

  $: sizeClass = modalSizeClasses[size || 'md'];

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
    // Trap focus within modal
    if (e.key === 'Tab' && modalElement) {
      const focusableElements = modalElement.querySelectorAll<HTMLElement>(
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

  onMount(() => {
    if (open) {
      previousActiveElement = document.activeElement;
      if (preventScroll) {
        document.body.style.overflow = 'hidden';
      }
    }
  });

  onDestroy(() => {
    if (preventScroll) {
      document.body.style.overflow = '';
    }
  });

  $: if (open) {
    previousActiveElement = document.activeElement;
    if (preventScroll && typeof document !== 'undefined') {
      document.body.style.overflow = 'hidden';
    }
  } else {
    if (preventScroll && typeof document !== 'undefined') {
      document.body.style.overflow = '';
    }
    if (previousActiveElement instanceof HTMLElement) {
      previousActiveElement.focus();
    }
  }
</script>

<svelte:window on:keydown={open ? handleKeydown : undefined} />

{#if open}
  <div class={modalClasses.overlay} transition:fade={{ duration: 150 }} aria-hidden="true" />

  <div
    class={modalClasses.container}
    role="dialog"
    aria-modal="true"
    aria-labelledby={title ? 'modal-title' : undefined}
    on:click={handleBackdropClick}
  >
    <div class={cn(modalClasses.wrapper, centered && modalClasses.wrapperCentered)}>
      <div
        bind:this={modalElement}
        class={cn(modalClasses.panel, sizeClass, className)}
        transition:scale={{ duration: 150, start: 0.95 }}
      >
        {#if title || showClose}
          <div class={modalClasses.header}>
            {#if title}
              <h2 id="modal-title" class={modalClasses.title}>{title}</h2>
            {:else}
              <div />
            {/if}
            {#if showClose}
              <button
                type="button"
                class={modalClasses.closeBtn}
                on:click={handleClose}
                aria-label="Close modal"
              >
                <svg class="w-5 h-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            {/if}
          </div>
        {/if}

        <div class={modalClasses.body}>
          <slot />
        </div>

        {#if $$slots.footer}
          <div class={modalClasses.footer}>
            <slot name="footer" />
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}
