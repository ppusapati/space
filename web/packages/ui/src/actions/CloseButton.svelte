<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { CloseButtonProps } from './actions.types';
  import { closeButtonClasses, closeButtonSizeClasses } from './actions.types';
  import { cn } from '../utils';

  type $$Props = CloseButtonProps;

  export let size: $$Props['size'] = 'md';
  export let disabled: $$Props['disabled'] = false;
  export let ariaLabel: $$Props['ariaLabel'] = 'Close';
  export let id: $$Props['id'] = undefined;
  export let testId: $$Props['testId'] = undefined;
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{
    click: { event: MouseEvent };
  }>();

  $: sizeConfig = closeButtonSizeClasses[size || 'md'];

  $: buttonClasses = cn(
    closeButtonClasses.base,
    closeButtonClasses.hover,
    sizeConfig.button,
    className
  );

  function handleClick(event: MouseEvent) {
    if (!disabled) {
      dispatch('click', { event });
    }
  }
</script>

<button
  type="button"
  {id}
  class={buttonClasses}
  {disabled}
  aria-label={ariaLabel}
  data-testid={testId}
  on:click={handleClick}
>
  <svg
    class={sizeConfig.icon}
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
    stroke-width="2"
    aria-hidden="true"
  >
    <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
  </svg>
</button>
