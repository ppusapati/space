<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import {
    checkboxSizeClasses,
    checkboxColorClasses,
    checkboxBoxBaseClasses,
    checkboxUncheckedClasses,
    checkboxContainerClasses,
    checkboxLabelClasses,
    checkboxLabelDisabledClasses,
    checkboxDescriptionClasses,
  } from './checkbox.types';
  import type { Size, ColorVariant } from '../types';

  // Props
  export let checked: boolean = false;
  export let indeterminate: boolean = false;
  export let value: string = '';
  export let label: string = '';
  export let description: string = '';
  export let size: Size = 'md';
  export let color: ColorVariant = 'primary';
  export let labelPosition: 'left' | 'right' = 'right';
  export let disabled: boolean = false;
  export let required: boolean = false;
  export let name: string = '';
  export let id: string = uid('checkbox');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  let inputRef: HTMLInputElement;

  const dispatch = createEventDispatcher<{
    change: { checked: boolean; value: string; event: Event };
  }>();

  // Sync indeterminate property (can't be set via attribute)
  $: if (inputRef) {
    inputRef.indeterminate = indeterminate;
  }

  // Computed classes
  $: sizeConfig = checkboxSizeClasses[size];
  $: colorConfig = checkboxColorClasses[color];

  $: boxClasses = cn(
    checkboxBoxBaseClasses,
    sizeConfig.box,
    colorConfig.focus,
    checked || indeterminate ? colorConfig.checked : checkboxUncheckedClasses,
    'flex items-center justify-center'
  );

  $: containerClasses = cn(
    checkboxContainerClasses,
    labelPosition === 'left' && 'flex-row-reverse',
    className
  );

  $: labelClasses = cn(
    checkboxLabelClasses,
    sizeConfig.label,
    disabled && checkboxLabelDisabledClasses
  );

  // Handlers
  function handleChange(event: Event) {
    const target = event.target as HTMLInputElement;
    checked = target.checked;
    if (indeterminate) {
      indeterminate = false;
    }
    dispatch('change', { checked, value, event });
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === ' ' || event.key === 'Enter') {
      event.preventDefault();
      if (!disabled) {
        checked = !checked;
        dispatch('change', { checked, value, event });
      }
    }
  }
</script>

<div class={containerClasses}>
  <div class="relative">
    <input
      bind:this={inputRef}
      type="checkbox"
      {id}
      {name}
      {value}
      {disabled}
      {required}
      {checked}
      class="sr-only peer"
      data-testid={testId || undefined}
      aria-checked={indeterminate ? 'mixed' : checked}
      aria-describedby={description ? `${id}-description` : undefined}
      on:change={handleChange}
    />
    <div
      class={boxClasses}
      role="checkbox"
      tabindex={disabled ? -1 : 0}
      aria-checked={indeterminate ? 'mixed' : checked}
      aria-disabled={disabled}
      on:click={() => !disabled && (inputRef.click())}
      on:keydown={handleKeydown}
    >
      {#if checked && !indeterminate}
        <!-- Checkmark icon -->
        <svg
          class={cn(sizeConfig.icon, 'text-neutral-white')}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="3"
            d="M5 13l4 4L19 7"
          />
        </svg>
      {:else if indeterminate}
        <!-- Indeterminate icon (minus) -->
        <svg
          class={cn(sizeConfig.icon, 'text-neutral-white')}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="3"
            d="M5 12h14"
          />
        </svg>
      {/if}
    </div>
  </div>

  {#if label || description}
    <div class="flex flex-col">
      {#if label}
        <label for={id} class={labelClasses}>
          {label}
          {#if required}
            <span class="text-semantic-error-500 ml-0.5" aria-hidden="true">*</span>
          {/if}
        </label>
      {/if}
      {#if description}
        <span id="{id}-description" class={checkboxDescriptionClasses}>
          {description}
        </span>
      {/if}
    </div>
  {/if}
</div>
