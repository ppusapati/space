<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import type { Size, ValidationState } from '../types';

  interface JsonEditorProps {
    value?: string | object;
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
    rows?: number;
    minRows?: number;
    maxRows?: number;
  }

  // Props
  export let value: string | object = '{}';
  export let label: string = '';
  export let helperText: string = 'Enter valid JSON';
  export let errorText: string = '';
  export let disabled: boolean = false;
  export let readonly: boolean = false;
  export let required: boolean = false;
  export let size: Size = 'md';
  export let state: ValidationState = 'default';
  export let name: string = '';
  export let id: string = uid('json-editor');
  export let rows: number = 6;
  export let minRows: number = 4;
  export let maxRows: number = 12;

  let className: string = '';
  export { className as class };

  let textValue: string = typeof value === 'string' ? value : JSON.stringify(value, null, 2);
  let isValidJson = true;
  let parseError = '';

  const dispatch = createEventDispatcher<{
    change: object | string;
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
    sm: 'px-3 py-2 text-sm',
    md: 'px-4 py-2 text-base',
    lg: 'px-4 py-3 text-lg',
  };

  function validateJson(str: string): boolean {
    if (!str || str.trim() === '') {
      parseError = '';
      return true;
    }

    try {
      JSON.parse(str);
      parseError = '';
      return true;
    } catch (e) {
      parseError = `Invalid JSON: ${e instanceof Error ? e.message : 'Unknown error'}`;
      return false;
    }
  }

  function handleInput(e: Event) {
    const target = e.target as HTMLTextAreaElement;
    textValue = target.value;

    // Validate while typing
    const isValid = validateJson(textValue);
    isValidJson = isValid;

    if (isValid && textValue.trim()) {
      try {
        const parsed = JSON.parse(textValue);
        value = parsed;
        dispatch('change', parsed);
      } catch {
        // Invalid JSON, don't dispatch
      }
    } else if (!textValue.trim()) {
      // Empty is allowed
      value = {};
      dispatch('change', value);
    }
  }

  function handleBlur() {
    // Format on blur
    if (validateJson(textValue) && textValue.trim()) {
      try {
        const parsed = JSON.parse(textValue);
        textValue = JSON.stringify(parsed, null, 2);
        value = parsed;
      } catch {
        // Keep as is if parsing fails
      }
    }
    dispatch('blur');
  }

  function handleFocus() {
    dispatch('focus');
  }

  function formatJson() {
    if (validateJson(textValue) && textValue.trim()) {
      try {
        const parsed = JSON.parse(textValue);
        textValue = JSON.stringify(parsed, null, 2);
        value = parsed;
      } catch {
        // Keep as is
      }
    }
  }

  function minifyJson() {
    if (validateJson(textValue) && textValue.trim()) {
      try {
        const parsed = JSON.parse(textValue);
        textValue = JSON.stringify(parsed);
        value = parsed;
      } catch {
        // Keep as is
      }
    }
  }

  // Watch for external value changes
  $: if (typeof value === 'object' && value !== null) {
    textValue = JSON.stringify(value, null, 2);
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
    <textarea
      {id}
      {name}
      {rows}
      {disabled}
      {readonly}
      value={textValue}
      on:input={handleInput}
      on:blur={handleBlur}
      on:focus={handleFocus}
      class={cn(
        'w-full border rounded-md font-mono text-sm transition-colors',
        'bg-white placeholder-neutral-400 text-neutral-900',
        sizeClasses[size as keyof typeof sizeClasses] ?? sizeClasses.md,
        isValidJson ? stateClasses[state as keyof typeof stateClasses] ?? stateClasses.default : stateClasses.error,
        disabled && 'bg-neutral-50 cursor-not-allowed opacity-50'
      )}
      style={`min-height: ${minRows}rem; max-height: ${maxRows}rem; resize: vertical;`}
    />

    <div class="flex gap-2 mt-2">
      <button
        type="button"
        on:click={formatJson}
        disabled={disabled || readonly || !textValue.trim()}
        class="px-2 py-1 text-xs bg-neutral-100 hover:bg-neutral-200 rounded disabled:opacity-50 transition-colors"
      >
        Format
      </button>
      <button
        type="button"
        on:click={minifyJson}
        disabled={disabled || readonly || !textValue.trim()}
        class="px-2 py-1 text-xs bg-neutral-100 hover:bg-neutral-200 rounded disabled:opacity-50 transition-colors"
      >
        Minify
      </button>
    </div>
  </div>

  {#if parseError}
    <p class="mt-1 text-sm text-red-500">{parseError}</p>
  {:else if errorText}
    <p class="mt-1 text-sm text-red-500">{errorText}</p>
  {:else if helperText}
    <p class="mt-1 text-sm text-neutral-500">{helperText}</p>
  {/if}
</div>

<style>
  textarea {
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', 'source-code-pro', monospace;
  }
</style>
