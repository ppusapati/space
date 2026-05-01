<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import Icon from '../display/Icon.svelte';
  import type { Size, ValidationState } from '../types';

  interface PercentageInputProps {
    value?: number;
    label?: string;
    helperText?: string;
    errorText?: string;
    disabled?: boolean;
    readonly?: boolean;
    required?: boolean;
    size?: Size;
    state?: ValidationState;
    name?: string;
    id?: string;
    min?: number;
    max?: number;
    step?: number;
    precision?: number;
    showButtons?: boolean;
  }

  // Props
  export let value: number = 0;
  export let label: string = '';
  export let helperText: string = '';
  export let errorText: string = '';
  export let disabled: boolean = false;
  export let readonly: boolean = false;
  export let required: boolean = false;
  export let size: Size = 'md';
  export let state: ValidationState = 'default';
  export let name: string = '';
  export let id: string = uid('percentage');
  export let min: number = 0;
  export let max: number = 100;
  export let step: number = 1;
  export let precision: number = 2;
  export let showButtons: boolean = true;

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: number;
    blur: void;
    focus: void;
  }>();

  const stateClasses = {
    default: 'border-neutral-300 focus:border-primary-500 focus:ring-primary-500',
    success: 'border-green-500 focus:border-green-600 focus:ring-green-500',
    error: 'border-red-500 focus:border-red-600 focus:ring-red-500',
    warning: 'border-yellow-500 focus:border-yellow-600 focus:ring-yellow-500',
  };

  const sizeClasses = {
    sm: 'px-3 py-1.5 text-sm',
    md: 'px-4 py-2 text-base',
    lg: 'px-4 py-3 text-lg',
  };

  function handleInput(e: Event) {
    const target = e.target as HTMLInputElement;
    let newValue = parseFloat(target.value) || 0;

    // Clamp value between min and max
    newValue = Math.max(min, Math.min(max, newValue));
    value = parseFloat(newValue.toFixed(precision));

    dispatch('change', value);
  }

  function increment() {
    let newValue = value + step;
    newValue = Math.max(min, Math.min(max, newValue));
    value = parseFloat(newValue.toFixed(precision));
    dispatch('change', value);
  }

  function decrement() {
    let newValue = value - step;
    newValue = Math.max(min, Math.min(max, newValue));
    value = parseFloat(newValue.toFixed(precision));
    dispatch('change', value);
  }

  function handleBlur() {
    // Ensure value is within bounds on blur
    if (value < min) value = min;
    if (value > max) value = max;
    dispatch('blur');
  }

  function handleFocus() {
    dispatch('focus');
  }
</script>

<div class={cn('w-full', className)}>
  {#if label}
    <label for={id} class="block text-sm font-medium text-neutral-700 mb-1">
      {label}
      {#if required}
        <span class="text-red-500 ml-1">*</span>
      {/if}
    </label>
  {/if}

  <div class="relative">
    <div
      class={cn(
        'flex items-center border rounded-md transition-colors',
        'bg-white',
        sizeClasses[size as keyof typeof sizeClasses] ?? sizeClasses.md,
        stateClasses[state as keyof typeof stateClasses] ?? stateClasses.default,
        disabled && 'bg-neutral-50 cursor-not-allowed opacity-50'
      )}
    >
      {#if showButtons}
        <button
          type="button"
          on:click={decrement}
          disabled={disabled || readonly || value <= min}
          class="text-neutral-400 hover:text-neutral-600 disabled:opacity-50 p-1 flex-shrink-0"
          aria-label="Decrease percentage"
        >
          <Icon name="minus" size="sm" />
        </button>
      {/if}

      <input
        {id}
        {name}
        {disabled}
        {readonly}
        type="number"
        {min}
        {max}
        {step}
        value={value.toFixed(precision)}
        on:input={handleInput}
        on:blur={handleBlur}
        on:focus={handleFocus}
        class={cn(
          'flex-1 bg-transparent border-0 outline-none text-center',
          'placeholder-neutral-400 text-neutral-900',
          disabled && 'cursor-not-allowed'
        )}
      />

      <span class="text-neutral-500 font-medium ml-2 flex-shrink-0">%</span>

      {#if showButtons}
        <button
          type="button"
          on:click={increment}
          disabled={disabled || readonly || value >= max}
          class="text-neutral-400 hover:text-neutral-600 disabled:opacity-50 p-1 flex-shrink-0"
          aria-label="Increase percentage"
        >
          <Icon name="plus" size="sm" />
        </button>
      {/if}
    </div>
  </div>

  {#if errorText}
    <p class="mt-1 text-sm text-red-500">{errorText}</p>
  {:else if helperText}
    <p class="mt-1 text-sm text-neutral-500">{helperText}</p>
  {/if}
</div>

<style>
  input[type='number']::-webkit-outer-spin-button,
  input[type='number']::-webkit-inner-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }

  input[type='number'] {
    -moz-appearance: textfield;
  }
</style>
