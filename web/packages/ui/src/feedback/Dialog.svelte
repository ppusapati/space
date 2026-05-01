<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { fade, scale } from 'svelte/transition';
  import type { DialogProps } from './feedback.types';
  import { modalClasses, modalSizeClasses } from './feedback.types';
  import { cn } from '../utils';

  type $$Props = DialogProps;

  export let open: $$Props['open'] = false;
  export let title: $$Props['title'] = undefined;
  export let size: $$Props['size'] = 'sm';
  export let closeOnBackdrop: $$Props['closeOnBackdrop'] = false;
  export let closeOnEscape: $$Props['closeOnEscape'] = true;
  export let showClose: $$Props['showClose'] = false;
  export let centered: $$Props['centered'] = true;
  export let preventScroll: $$Props['preventScroll'] = true;
  export let variant: $$Props['variant'] = 'confirm';
  export let confirmText: $$Props['confirmText'] = 'Confirm';
  export let cancelText: $$Props['cancelText'] = 'Cancel';
  export let loading: $$Props['loading'] = false;
  export let destructive: $$Props['destructive'] = false;
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{ confirm: void; cancel: void; close: void }>();

  let dialogElement: HTMLDivElement;

  $: sizeClass = modalSizeClasses[size || 'sm'];

  function handleConfirm() {
    if (!loading) {
      dispatch('confirm');
    }
  }

  function handleCancel() {
    if (!loading) {
      dispatch('cancel');
      dispatch('close');
    }
  }

  function handleBackdropClick(e: MouseEvent) {
    if (closeOnBackdrop && e.target === e.currentTarget && !loading) {
      handleCancel();
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (closeOnEscape && e.key === 'Escape' && !loading) {
      handleCancel();
    }
  }

  const variantIcons: Record<string, { path: string; color: string; bg: string }> = {
    info: {
      path: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
      color: 'text-semantic-info-500',
      bg: 'bg-semantic-info-100',
    },
    warning: {
      path: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z',
      color: 'text-semantic-warning-500',
      bg: 'bg-semantic-warning-100',
    },
    error: {
      path: 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z',
      color: 'text-semantic-error-500',
      bg: 'bg-semantic-error-100',
    },
    success: {
      path: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
      color: 'text-semantic-success-500',
      bg: 'bg-semantic-success-100',
    },
    confirm: {
      path: 'M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
      color: 'text-brand-primary-500',
      bg: 'bg-brand-primary-100',
    },
  };

  $: iconStyles = variantIcons[variant || 'confirm']!;

  $: if (open && typeof document !== 'undefined') {
    document.body.style.overflow = preventScroll ? 'hidden' : '';
  } else if (typeof document !== 'undefined') {
    document.body.style.overflow = '';
  }
</script>

<svelte:window on:keydown={open ? handleKeydown : undefined} />

{#if open}
  <div class={modalClasses.overlay} transition:fade={{ duration: 150 }} aria-hidden="true" />

  <div
    class={modalClasses.container}
    role="alertdialog"
    aria-modal="true"
    aria-labelledby={title ? 'dialog-title' : undefined}
    on:click={handleBackdropClick}
  >
    <div class={cn(modalClasses.wrapper, centered && modalClasses.wrapperCentered)}>
      <div
        bind:this={dialogElement}
        class={cn(modalClasses.panel, sizeClass, className)}
        transition:scale={{ duration: 150, start: 0.95 }}
      >
        <div class="p-6 text-center">
          <div class={cn('mx-auto w-12 h-12 rounded-full flex items-center justify-center mb-4', iconStyles.bg)}>
            <svg
              class={cn('w-6 h-6', iconStyles.color)}
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
              aria-hidden="true"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d={iconStyles.path} />
            </svg>
          </div>

          {#if title}
            <h3 id="dialog-title" class="text-lg font-semibold text-neutral-900 mb-2">
              {title}
            </h3>
          {/if}

          <div class="text-sm text-neutral-600">
            <slot />
          </div>
        </div>

        <div class="flex items-center justify-center gap-3 px-6 pb-6">
          <button
            type="button"
            class={cn(
              'px-4 py-2 text-sm font-medium rounded-md transition-colors',
              'border border-neutral-300 bg-neutral-white text-neutral-700',
              'hover:bg-neutral-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-neutral-500',
              loading && 'opacity-50 cursor-not-allowed'
            )}
            on:click={handleCancel}
            disabled={loading}
          >
            {cancelText}
          </button>
          <button
            type="button"
            class={cn(
              'px-4 py-2 text-sm font-medium rounded-md transition-colors',
              'focus:outline-none focus:ring-2 focus:ring-offset-2',
              destructive
                ? 'bg-semantic-error-600 text-neutral-white hover:bg-semantic-error-700 focus:ring-semantic-error-500'
                : 'bg-brand-primary-600 text-neutral-white hover:bg-brand-primary-700 focus:ring-brand-primary-500',
              loading && 'opacity-50 cursor-not-allowed'
            )}
            on:click={handleConfirm}
            disabled={loading}
          >
            {#if loading}
              <svg class="animate-spin -ml-1 mr-2 h-4 w-4 inline" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
            {/if}
            {confirmText}
          </button>
        </div>

        {#if showClose}
          <button
            type="button"
            class={cn(modalClasses.closeBtn, 'absolute top-4 right-4')}
            on:click={handleCancel}
            aria-label="Close dialog"
            disabled={loading}
          >
            <svg class="w-5 h-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        {/if}
      </div>
    </div>
  </div>
{/if}
