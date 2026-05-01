<script lang="ts">
  import { createEventDispatcher, getContext } from 'svelte';
  import type { ListItemProps } from './display.types';
  import { listItemClasses, listItemSizeClasses } from './display.types';
  import { cn } from '../utils';
  import type { Size } from '../types';

  type $$Props = ListItemProps;

  export let active: $$Props['active'] = false;
  export let disabled: $$Props['disabled'] = false;
  export let clickable: $$Props['clickable'] = false;
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{ click: MouseEvent }>();

  // Try to get size from parent List context, default to 'md'
  const size: Size = 'md';
  $: sizeClass = listItemSizeClasses[size];

  function handleClick(e: MouseEvent) {
    if (!disabled && clickable) {
      dispatch('click', e);
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (!disabled && clickable && (e.key === 'Enter' || e.key === ' ')) {
      e.preventDefault();
      dispatch('click', new MouseEvent('click'));
    }
  }
</script>

<li
  class={cn(
    listItemClasses.base,
    sizeClass,
    clickable && !disabled && listItemClasses.clickable,
    active && listItemClasses.active,
    disabled && listItemClasses.disabled,
    className
  )}
  role={clickable ? 'button' : 'listitem'}
  tabindex={clickable && !disabled ? 0 : undefined}
  aria-disabled={disabled || undefined}
  aria-current={active ? 'true' : undefined}
  on:click={handleClick}
  on:keydown={handleKeydown}
>
  {#if $$slots.leading}
    <div class="shrink-0">
      <slot name="leading" />
    </div>
  {/if}

  <div class="flex-1 min-w-0">
    <slot />
  </div>

  {#if $$slots.trailing}
    <div class="shrink-0">
      <slot name="trailing" />
    </div>
  {/if}
</li>
