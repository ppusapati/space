<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import {
    type InputProps,
    type InputType,
    type InputVariant,
    inputBaseClasses,
    inputSizeClasses,
    inputVariantClasses,
    inputStateClasses,
    labelClasses,
    helperTextClasses,
    requiredClasses,
    iconContainerClasses,
    clearButtonClasses,
  } from './input.types';
  import type { Size, ValidationState } from '../types';

  // Props with defaults
  export let type: InputType = 'text';
  export let value: string | number = '';
  export let placeholder: string = '';
  export let size: Size = 'md';
  export let variant: InputVariant = 'default';
  export let state: ValidationState = 'default';
  export let label: string = '';
  export let helperText: string = '';
  export let errorText: string = '';
  export let iconLeft: boolean = false;
  export let iconRight: boolean = false;
  export let clearable: boolean = false;
  export let disabled: boolean = false;
  export let readonly: boolean = false;
  export let required: boolean = false;
  export let name: string = '';
  export let id: string = uid('input');
  export let testId: string = '';
  export let min: number | string | undefined = undefined;
  export let max: number | string | undefined = undefined;
  export let step: number | string | undefined = undefined;
  export let minlength: number | undefined = undefined;
  export let maxlength: number | undefined = undefined;
  export let pattern: string | undefined = undefined;
  export let autocomplete: import('svelte/elements').FullAutoFill = 'off';
  export let fullWidth: boolean = true;

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    input: { value: string | number; event: Event };
    change: { value: string | number; event: Event };
    focus: { event: FocusEvent };
    blur: { event: FocusEvent };
    clear: void;
    keydown: { event: KeyboardEvent };
  }>();

  // Computed classes
  $: inputClasses = cn(
    inputBaseClasses,
    inputSizeClasses[size],
    inputVariantClasses[variant],
    inputStateClasses[state === 'default' && errorText ? 'invalid' : state],
    iconLeft && 'pl-10',
    (iconRight || clearable) && 'pr-10',
    className
  );

  $: containerClasses = cn(
    'relative',
    fullWidth ? 'w-full' : 'inline-block'
  );

  $: displayedHelperText = state === 'invalid' || errorText ? errorText : helperText;
  $: helperClasses = helperTextClasses[errorText ? 'invalid' : state];

  // Handlers
  function handleInput(event: Event) {
    const target = event.target as HTMLInputElement;
    value = type === 'number' ? target.valueAsNumber : target.value;
    dispatch('input', { value, event });
  }

  function handleChange(event: Event) {
    const target = event.target as HTMLInputElement;
    value = type === 'number' ? target.valueAsNumber : target.value;
    dispatch('change', { value, event });
  }

  function handleFocus(event: FocusEvent) {
    dispatch('focus', { event });
  }

  function handleBlur(event: FocusEvent) {
    dispatch('blur', { event });
  }

  function handleKeydown(event: KeyboardEvent) {
    dispatch('keydown', { event });
  }

  function handleClear() {
    value = '';
    dispatch('clear');
    dispatch('input', { value: '', event: new Event('input') });
  }

  $: hasValue = value !== '' && value !== null && value !== undefined;
  $: showClearButton = clearable && hasValue && !disabled && !readonly;
</script>

<div class={containerClasses}>
  {#if label}
    <label for={id} class={labelClasses}>
      {label}
      {#if required}
        <span class={requiredClasses} aria-hidden="true">*</span>
      {/if}
    </label>
  {/if}

  <div class="relative">
    {#if iconLeft}
      <div class={iconContainerClasses.left}>
        <slot name="icon-left">
          <!-- Default left icon slot -->
        </slot>
      </div>
    {/if}

    <input
      {id}
      {type}
      {name}
      {placeholder}
      {disabled}
      {readonly}
      {required}
      {min}
      {max}
      {step}
      {minlength}
      {maxlength}
      {pattern}
      autocomplete={autocomplete}
      {value}
      class={inputClasses}
      data-testid={testId || undefined}
      aria-invalid={state === 'invalid' || !!errorText}
      aria-describedby={displayedHelperText ? `${id}-helper` : undefined}
      aria-required={required}
      on:input={handleInput}
      on:change={handleChange}
      on:focus={handleFocus}
      on:blur={handleBlur}
      on:keydown={handleKeydown}
    />

    {#if showClearButton}
      <button
        type="button"
        class={clearButtonClasses}
        on:click={handleClear}
        aria-label="Clear input"
        tabindex="-1"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    {:else if iconRight}
      <div class={iconContainerClasses.right}>
        <slot name="icon-right">
          <!-- Default right icon slot -->
        </slot>
      </div>
    {/if}
  </div>

  {#if displayedHelperText}
    <p id="{id}-helper" class={helperClasses}>
      {displayedHelperText}
    </p>
  {/if}
</div>
