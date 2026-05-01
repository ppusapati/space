<script lang="ts">
  import { fade } from 'svelte/transition';
  import type { SpinnerProps } from './feedback.types';
  import { spinnerClasses, spinnerSizeClasses, spinnerVariantClasses } from './feedback.types';
  import { cn } from '../utils';

  type $$Props = SpinnerProps & {
    fullscreen?: boolean;
    backdrop?: boolean;
    backdropOpacity?: 'light' | 'medium' | 'dark';
  };

  export let size: $$Props['size'] = 'lg';
  export let variant: $$Props['variant'] = 'primary';
  export let label: $$Props['label'] = 'Loading...';
  export let fullscreen: $$Props['fullscreen'] = false;
  export let backdrop: $$Props['backdrop'] = true;
  export let backdropOpacity: $$Props['backdropOpacity'] = 'medium';
  let className: $$Props['class'] = undefined;
  export { className as class };

  $: sizeClass = spinnerSizeClasses[size || 'lg'];
  $: variantClass = spinnerVariantClasses[variant || 'primary'];

  const opacityClasses = {
    light: 'bg-neutral-white/50',
    medium: 'bg-neutral-white/75',
    dark: 'bg-neutral-white/90',
  };

  $: backdropClass = opacityClasses[backdropOpacity || 'medium'];
</script>

{#if fullscreen}
  <div
    class={cn(
      'fixed inset-0 z-loader flex items-center justify-center',
      backdrop && backdropClass,
      className
    )}
    role="status"
    aria-live="polite"
    aria-busy="true"
    transition:fade={{ duration: 150 }}
  >
    <div class="flex flex-col items-center gap-3">
      <svg
        class={cn(spinnerClasses.spinner, sizeClass.spinner, variantClass)}
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        aria-hidden="true"
      >
        <circle
          class="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          stroke-width="4"
        />
        <path
          class="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        />
      </svg>
      {#if label}
        <span class={cn('text-neutral-600 font-medium', sizeClass.label)}>{label}</span>
      {/if}
    </div>
  </div>
{:else}
  <div
    class={cn('inline-flex flex-col items-center gap-2', className)}
    role="status"
    aria-live="polite"
    aria-busy="true"
  >
    <svg
      class={cn(spinnerClasses.spinner, sizeClass.spinner, variantClass)}
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <circle
        class="opacity-25"
        cx="12"
        cy="12"
        r="10"
        stroke="currentColor"
        stroke-width="4"
      />
      <path
        class="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
      />
    </svg>
    {#if label}
      <span class={cn('text-neutral-600', sizeClass.label)}>{label}</span>
    {/if}
    <span class="sr-only">{label || 'Loading...'}</span>
  </div>
{/if}
