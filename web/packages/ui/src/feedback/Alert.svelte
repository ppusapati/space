<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { AlertProps } from './feedback.types';
  import { alertClasses, alertVariantClasses, alertSizeClasses } from './feedback.types';
  import { cn } from '../utils';

  type $$Props = AlertProps;

  export let variant: $$Props['variant'] = 'info';
  export let title: $$Props['title'] = undefined;
  export let showIcon: $$Props['showIcon'] = true;
  export let dismissible: $$Props['dismissible'] = false;
  export let size: $$Props['size'] = 'md';
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{ dismiss: void }>();

  $: variantStyles = alertVariantClasses[variant || 'info'];
  $: sizeClass = alertSizeClasses[size || 'md'];

  let visible = true;

  function handleDismiss() {
    visible = false;
    dispatch('dismiss');
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
  <div
    class={cn(
      alertClasses.container,
      variantStyles.bg,
      variantStyles.border,
      sizeClass,
      className
    )}
    role="alert"
  >
    {#if showIcon}
      <svg
        class={cn(alertClasses.icon, variantStyles.icon)}
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        stroke-width="2"
        aria-hidden="true"
      >
        <path stroke-linecap="round" stroke-linejoin="round" d={icons[variant || 'info']} />
      </svg>
    {/if}

    <div class={alertClasses.content}>
      {#if title}
        <p class={cn(alertClasses.title, variantStyles.title)}>{title}</p>
      {/if}
      <div class={cn(alertClasses.message, variantStyles.message)}>
        <slot />
      </div>
    </div>

    {#if dismissible}
      <button
        type="button"
        class={alertClasses.closeBtn}
        on:click={handleDismiss}
        aria-label="Dismiss alert"
      >
        <svg class="w-4 h-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    {/if}
  </div>
{/if}
