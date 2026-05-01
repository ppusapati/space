<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import Icon from '../display/Icon.svelte';
  import type { Size, ValidationState } from '../types';

  interface UrlInputProps {
    value?: string;
    placeholder?: string;
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
    clearable?: boolean;
    protocol?: 'http' | 'https' | 'both';
  }

  // Props
  export let value: string = '';
  export let placeholder: string = 'https://example.com';
  export let label: string = '';
  export let helperText: string = '';
  export let errorText: string = '';
  export let disabled: boolean = false;
  export let readonly: boolean = false;
  export let required: boolean = false;
  export let size: Size = 'md';
  export let state: ValidationState = 'default';
  export let name: string = '';
  export let id: string = uid('url');
  export let clearable: boolean = false;
  export let protocol: 'http' | 'https' | 'both' = 'both';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: string;
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

  function isValidUrl(url: string): boolean {
    if (!url) return true; // Empty is valid, let required handle it

    try {
      const urlObj = new URL(url);
      if (protocol === 'https' && urlObj.protocol !== 'https:') return false;
      if (protocol === 'http' && !['http:', 'https:'].includes(urlObj.protocol)) return false;
      return true;
    } catch {
      return false;
    }
  }

  function handleInput(e: Event) {
    const target = e.target as HTMLInputElement;
    value = target.value;
    dispatch('change', value);
  }

  function handleClear() {
    value = '';
    dispatch('change', value);
  }

  function handleBlur() {
    dispatch('blur');
  }

  function handleFocus() {
    dispatch('focus');
  }

  $: isValid = isValidUrl(value);
  $: displayState = !isValid && value ? 'error' : state;
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
        stateClasses[displayState as keyof typeof stateClasses] ?? stateClasses.default,
        disabled && 'bg-neutral-50 cursor-not-allowed opacity-50'
      )}
    >
      <Icon name="link" class="text-neutral-400 flex-shrink-0" />

      <input
        {id}
        {name}
        {placeholder}
        {disabled}
        {readonly}
        type="url"
        value={value}
        on:input={handleInput}
        on:blur={handleBlur}
        on:focus={handleFocus}
        class={cn(
          'flex-1 bg-transparent border-0 outline-none px-2',
          'placeholder-neutral-400 text-neutral-900',
          disabled && 'cursor-not-allowed'
        )}
      />

      {#if clearable && value}
        <button
          type="button"
          on:click={handleClear}
          disabled={disabled || readonly}
          class="text-neutral-400 hover:text-neutral-600 flex-shrink-0 p-1"
          aria-label="Clear URL"
        >
          <Icon name="x" size="sm" />
        </button>
      {/if}
    </div>
  </div>

  {#if !isValid && value}
    <p class="mt-1 text-sm text-red-500">Please enter a valid URL</p>
  {:else if errorText}
    <p class="mt-1 text-sm text-red-500">{errorText}</p>
  {:else if helperText}
    <p class="mt-1 text-sm text-neutral-500">{helperText}</p>
  {/if}
</div>

<style>
  input::placeholder {
    @apply text-neutral-400;
  }
</style>
