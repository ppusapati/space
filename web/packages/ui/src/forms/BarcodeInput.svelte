<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import Icon from '../display/Icon.svelte';
  import type { Size, ValidationState } from '../types';

  interface BarcodeInputProps {
    value?: string;
    label?: string;
    placeholder?: string;
    helperText?: string;
    errorText?: string;
    disabled?: boolean;
    readonly?: boolean;
    required?: boolean;
    size?: Size;
    state?: ValidationState;
    name?: string;
    id?: string;
    barcodeFormat?: string;
    clearable?: boolean;
  }

  export let value: string = '';
  export let label: string = '';
  export let placeholder: string = 'Scan barcode or enter manually';
  export let helperText: string = '';
  export let errorText: string = '';
  export let disabled: boolean = false;
  export let readonly: boolean = false;
  export let required: boolean = false;
  export let size: Size = 'md';
  export let state: ValidationState = 'default';
  export let name: string = '';
  export let id: string = uid('barcode');
  export let barcodeFormat: string = 'CODE128';
  export let clearable: boolean = true;

  let className: string = '';
  export { className as class };

  let isScannerActive = false;
  let scanBuffer = '';
  let scanTimer: NodeJS.Timeout | null = null;
  let inputElement: HTMLInputElement | undefined;

  const dispatch = createEventDispatcher<{
    change: string;
    scan: string;
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

  onMount(() => {
    // Focus input on mount to catch scanner input
    if (inputElement) {
      inputElement.focus();
    }

    // Setup keyboard listener for barcode scanner (typically emits rapidly)
    const handleScannerInput = (e: KeyboardEvent) => {
      if (!isScannerActive) return;

      // Barcode scanners typically end with Enter key
      if (e.key === 'Enter' && scanBuffer) {
        value = scanBuffer;
        dispatch('change', value);
        dispatch('scan', value);
        scanBuffer = '';

        if (scanTimer) clearTimeout(scanTimer);
        return;
      }

      if (e.key.length === 1) {
        scanBuffer += e.key;

        // Reset buffer if no input for 100ms (indicates manual typing)
        if (scanTimer) clearTimeout(scanTimer);
        scanTimer = setTimeout(() => {
          scanBuffer = '';
        }, 100);
      }
    };

    window.addEventListener('keydown', handleScannerInput);

    return () => {
      window.removeEventListener('keydown', handleScannerInput);
      if (scanTimer) clearTimeout(scanTimer);
    };
  });

  function handleInput(e: Event) {
    const target = e.target as HTMLInputElement;
    value = target.value;
    dispatch('change', value);
  }

  function toggleScanner() {
    isScannerActive = !isScannerActive;
    if (isScannerActive && inputElement) {
      inputElement.focus();
    }
  }

  function handleClear() {
    value = '';
    scanBuffer = '';
    dispatch('change', value);
    if (inputElement) {
      inputElement.focus();
    }
  }

  function handleBlur() {
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
      <button
        type="button"
        on:click={toggleScanner}
        disabled={disabled || readonly}
        class={cn(
          'p-1 flex-shrink-0 rounded transition-colors',
          isScannerActive
            ? 'bg-primary-100 text-primary-600'
            : 'text-neutral-400 hover:text-neutral-600',
          disabled && 'opacity-50 cursor-not-allowed'
        )}
        title={isScannerActive ? 'Scanner active' : 'Activate scanner'}
      >
        <Icon name={isScannerActive ? 'radio' : 'barcode'} />
      </button>

      <input
        bind:this={inputElement}
        {id}
        {name}
        {placeholder}
        {disabled}
        {readonly}
        type="text"
        value={value}
        on:input={handleInput}
        on:blur={handleBlur}
        on:focus={handleFocus}
        class={cn(
          'flex-1 bg-transparent border-0 outline-none px-2',
          'placeholder-neutral-400 text-neutral-900',
          disabled && 'cursor-not-allowed'
        )}
        autocomplete="off"
      />

      {#if clearable && value}
        <button
          type="button"
          on:click={handleClear}
          disabled={disabled || readonly}
          class="text-neutral-400 hover:text-neutral-600 flex-shrink-0 p-1"
          aria-label="Clear barcode"
        >
          <Icon name="x" size="sm" />
        </button>
      {/if}
    </div>

    {#if isScannerActive}
      <div class="absolute right-0 mt-1 text-xs bg-primary-100 text-primary-700 px-2 py-1 rounded whitespace-nowrap">
        Scanner ready • Format: {barcodeFormat}
      </div>
    {/if}
  </div>

  {#if errorText}
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
