<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { TagProps } from './display.types';
  import { tagClasses, tagSizeClasses, tagVariantClasses } from './display.types';
  import { cn } from '../utils';

  type $$Props = TagProps;

  export let variant: $$Props['variant'] = 'neutral';
  export let size: $$Props['size'] = 'md';
  export let removable: $$Props['removable'] = false;
  export let clickable: $$Props['clickable'] = false;
  export let disabled: $$Props['disabled'] = false;
  export let icon: $$Props['icon'] = undefined;
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{ remove: void; click: void }>();

  $: sizeStyles = tagSizeClasses[size || 'md'];
  $: variantStyles = tagVariantClasses[variant || 'neutral'];

  function handleClick() {
    if (!disabled && clickable) {
      dispatch('click');
    }
  }

  function handleRemove(e: MouseEvent) {
    e.stopPropagation();
    if (!disabled) {
      dispatch('remove');
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleClick();
    }
    if (e.key === 'Backspace' || e.key === 'Delete') {
      if (removable) {
        dispatch('remove');
      }
    }
  }
</script>

<span
  class={cn(
    tagClasses.base,
    sizeStyles.container,
    variantStyles.bg,
    clickable && !disabled && variantStyles.hover,
    clickable && !disabled && tagClasses.clickable,
    disabled && tagClasses.disabled,
    className
  )}
  role={clickable ? 'button' : undefined}
  tabindex={clickable && !disabled ? 0 : undefined}
  on:click={handleClick}
  on:keydown={handleKeydown}
>
  {#if icon}
    <svg
      class={sizeStyles.icon}
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      stroke-width="2"
      aria-hidden="true"
    >
      <path stroke-linecap="round" stroke-linejoin="round" d={icon} />
    </svg>
  {/if}

  <slot />

  {#if removable && !disabled}
    <button
      type="button"
      class={tagClasses.removeBtn}
      on:click={handleRemove}
      aria-label="Remove tag"
    >
      <svg
        class={sizeStyles.removeIcon}
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        stroke-width="2"
      >
        <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
      </svg>
    </button>
  {/if}
</span>
