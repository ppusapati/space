<script lang="ts">
  import type { SpinnerProps } from './feedback.types';
  import { spinnerClasses, spinnerSizeClasses, spinnerVariantClasses } from './feedback.types';
  import { cn } from '../utils';

  type $$Props = SpinnerProps;

  export let size: $$Props['size'] = 'md';
  export let variant: $$Props['variant'] = 'primary';
  export let label: $$Props['label'] = undefined;
  let className: $$Props['class'] = undefined;
  export { className as class };

  $: sizeClass = spinnerSizeClasses[size || 'md'];
  $: variantClass = spinnerVariantClasses[variant || 'primary'];
</script>

<div
  class={cn(spinnerClasses.container, className)}
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
    <span class={cn(spinnerClasses.label, sizeClass.label)}>{label}</span>
  {/if}
  <span class="sr-only">{label || 'Loading...'}</span>
</div>
