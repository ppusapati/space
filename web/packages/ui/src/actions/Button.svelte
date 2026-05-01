<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { ButtonProps } from './actions.types';
  import {
    buttonBaseClasses,
    buttonSizeClasses,
    buttonIconOnlySizeClasses,
    buttonVariantClasses,
    buttonLoadingSpinnerClasses,
  } from './actions.types';
  import { cn } from '../utils';

  type $$Props = ButtonProps;

  export let variant: $$Props['variant'] = 'primary';
  export let size: $$Props['size'] = 'md';
  export let type: $$Props['type'] = 'button';
  export let disabled: $$Props['disabled'] = false;
  export let loading: $$Props['loading'] = false;
  export let fullWidth: $$Props['fullWidth'] = false;
  export let iconOnly: $$Props['iconOnly'] = false;
  export let href: $$Props['href'] = undefined;
  export let target: $$Props['target'] = undefined;
  export let ariaLabel: $$Props['ariaLabel'] = undefined;
  export let id: $$Props['id'] = undefined;
  export let testId: $$Props['testId'] = undefined;
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{
    click: { event: MouseEvent };
  }>();

  $: isDisabled = disabled || loading;
  $: sizeClass = iconOnly
    ? buttonIconOnlySizeClasses[size || 'md']
    : buttonSizeClasses[size || 'md'];
  $: variantClass = buttonVariantClasses[variant || 'primary'];

  $: buttonClasses = cn(
    buttonBaseClasses,
    sizeClass,
    variantClass,
    fullWidth && 'w-full',
    className
  );

  function handleClick(event: MouseEvent) {
    if (!isDisabled) {
      dispatch('click', { event });
    }
  }
</script>

{#if href && !isDisabled}
  <a
    {href}
    {target}
    {id}
    class={buttonClasses}
    aria-label={ariaLabel}
    aria-disabled={isDisabled}
    data-testid={testId}
    rel={target === '_blank' ? 'noopener noreferrer' : undefined}
    on:click={handleClick}
  >
    {#if loading}
      <span class={buttonLoadingSpinnerClasses} aria-hidden="true"></span>
    {/if}
    <slot />
  </a>
{:else}
  <button
    {type}
    {id}
    class={buttonClasses}
    disabled={isDisabled}
    aria-label={ariaLabel}
    aria-busy={loading}
    data-testid={testId}
    on:click={handleClick}
  >
    {#if loading}
      <span class={buttonLoadingSpinnerClasses} aria-hidden="true"></span>
    {/if}
    <span class={loading ? 'opacity-0' : ''}>
      <slot />
    </span>
  </button>
{/if}
