<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { CardProps } from './display.types';
  import { cardClasses, cardPaddingClasses } from './display.types';
  import { cn } from '../utils';

  type $$Props = CardProps;

  export let variant: $$Props['variant'] = 'elevated';
  export let padding: $$Props['padding'] = 'md';
  export let hoverable: $$Props['hoverable'] = false;
  export let clickable: $$Props['clickable'] = false;
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{ click: MouseEvent }>();

  $: variantClass = cardClasses[variant || 'elevated'];
  $: paddingClass = cardPaddingClasses[padding || 'md'];

  function handleClick(e: MouseEvent) {
    if (clickable) {
      dispatch('click', e);
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (clickable && (e.key === 'Enter' || e.key === ' ')) {
      e.preventDefault();
      dispatch('click', new MouseEvent('click'));
    }
  }
</script>

<div
  class={cn(
    cardClasses.base,
    variantClass,
    paddingClass,
    hoverable && cardClasses.hoverable,
    clickable && cardClasses.clickable,
    className
  )}
  role={clickable ? 'button' : undefined}
  tabindex={clickable ? 0 : undefined}
  on:click={handleClick}
  on:keydown={handleKeydown}
>
  {#if $$slots.header}
    <div class="mb-4">
      <slot name="header" />
    </div>
  {/if}

  <slot />

  {#if $$slots.footer}
    <div class="mt-4">
      <slot name="footer" />
    </div>
  {/if}
</div>
