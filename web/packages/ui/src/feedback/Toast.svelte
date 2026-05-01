<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import type { ToastProps } from './feedback.types';
  import { toastClasses, toastVariantClasses, toastPositionClasses } from './feedback.types';
  import { cn } from '../utils';

  type $$Props = ToastProps;

  export let variant: $$Props['variant'] = 'neutral';
  export let title: $$Props['title'] = undefined;
  export let message: $$Props['message'];
  export let duration: $$Props['duration'] = 5000;
  export let dismissible: $$Props['dismissible'] = true;
  export let position: $$Props['position'] = 'top-right';
  export let action: $$Props['action'] = undefined;
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{ dismiss: void; action: void }>();

  $: variantStyles = toastVariantClasses[variant || 'neutral'];
  $: positionClass = toastPositionClasses[position || 'top-right'];

  let visible = true;
  let timeoutId: ReturnType<typeof setTimeout> | undefined;

  onMount(() => {
    if (duration && duration > 0) {
      timeoutId = setTimeout(() => {
        handleDismiss();
      }, duration);
    }
  });

  onDestroy(() => {
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
  });

  function handleDismiss() {
    visible = false;
    dispatch('dismiss');
  }

  function handleAction() {
    if (action?.onClick) {
      action.onClick();
    }
    dispatch('action');
  }

  const icons: Record<string, string> = {
    primary: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
    secondary: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
    success: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
    warning: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z',
    error: 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z',
    info: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
    neutral: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
  };
</script>

{#if visible}
  <div class={cn(toastClasses.container, positionClass)}>
    <div
      class={cn(
        toastClasses.toast,
        variantStyles.bg,
        variantStyles.border,
        className
      )}
      role="alert"
      aria-live="polite"
    >
      <svg
        class={cn(toastClasses.icon, variantStyles.icon)}
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        stroke-width="2"
        aria-hidden="true"
      >
        <path stroke-linecap="round" stroke-linejoin="round" d={icons[variant || 'neutral']} />
      </svg>

      <div class={toastClasses.content}>
        {#if title}
          <p class={cn(toastClasses.title, variantStyles.title)}>{title}</p>
        {/if}
        <p class={cn(toastClasses.message, variantStyles.message)}>{message}</p>
        {#if action}
          <button
            type="button"
            class={toastClasses.action}
            on:click={handleAction}
          >
            {action.label}
          </button>
        {/if}
      </div>

      {#if dismissible}
        <button
          type="button"
          class={toastClasses.closeBtn}
          on:click={handleDismiss}
          aria-label="Dismiss notification"
        >
          <svg class="w-4 h-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      {/if}
    </div>
  </div>
{/if}
