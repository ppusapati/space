<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import Icon from '../display/Icon.svelte';
  import type { Size, ValidationState } from '../types';

  interface LookupOption {
    value: string | number;
    label: string;
    description?: string;
  }

  interface LookupFieldProps {
    value?: string | number | (string | number)[];
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
    options?: LookupOption[];
    multiple?: boolean;
    searchable?: boolean;
    clearable?: boolean;
    isLoading?: boolean;
    onSearch?: (query: string) => Promise<LookupOption[]>;
  }

  // Props
  export let value: string | number | (string | number)[] = '';
  export let label: string = '';
  export let placeholder: string = 'Search...';
  export let helperText: string = '';
  export let errorText: string = '';
  export let disabled: boolean = false;
  export let readonly: boolean = false;
  export let required: boolean = false;
  export let size: Size = 'md';
  export let state: ValidationState = 'default';
  export let name: string = '';
  export let id: string = uid('lookup');
  export let options: LookupOption[] = [];
  export let multiple: boolean = false;
  export let searchable: boolean = true;
  export let clearable: boolean = true;
  export let isLoading: boolean = false;
  export let onSearch: ((query: string) => Promise<LookupOption[]>) | undefined = undefined;

  let className: string = '';
  export { className as class };

  let isOpen = false;
  let searchQuery = '';
  let filteredOptions: LookupOption[] = [];
  let highlightedIndex = -1;

  const dispatch = createEventDispatcher<{
    change: string | number | (string | number)[];
    search: string;
    blur: void;
    focus: void;
  }>();

  const stateClasses = {
    default: 'border-neutral-300 focus-within:border-primary-500 focus-within:ring-primary-500',
    success: 'border-green-500 focus-within:border-green-600 focus-within:ring-green-500',
    error: 'border-red-500 focus-within:border-red-600 focus-within:ring-red-500',
    warning: 'border-yellow-500 focus-within:border-yellow-600 focus-within:ring-yellow-500',
  };

  const sizeClasses = {
    sm: 'px-3 py-1.5 text-sm',
    md: 'px-4 py-2 text-base',
    lg: 'px-4 py-3 text-lg',
  };

  async function handleSearch(query: string) {
    searchQuery = query;
    dispatch('search', query);

    if (onSearch) {
      try {
        filteredOptions = await onSearch(query);
      } catch {
        filteredOptions = [];
      }
    } else {
      // Client-side filtering
      const lowerQuery = query.toLowerCase();
      filteredOptions = options.filter(
        (opt) =>
          opt.label.toLowerCase().includes(lowerQuery) ||
          opt.value.toString().toLowerCase().includes(lowerQuery)
      );
    }

    isOpen = true;
    highlightedIndex = -1;
  }

  function selectOption(option: LookupOption) {
    if (multiple && Array.isArray(value)) {
      if (value.includes(option.value)) {
        value = value.filter((v) => v !== option.value);
      } else {
        value = [...value, option.value];
      }
    } else {
      value = option.value;
      isOpen = false;
    }

    searchQuery = '';
    dispatch('change', value);
  }

  function removeOption(optionValue: string | number) {
    if (multiple && Array.isArray(value)) {
      value = value.filter((v) => v !== optionValue);
      dispatch('change', value);
    }
  }

  function handleClear() {
    value = multiple ? [] : '';
    searchQuery = '';
    dispatch('change', value);
  }

  function getSelectedLabels(): string[] {
    if (Array.isArray(value)) {
      return value
        .map((v) => options.find((opt) => opt.value === v)?.label || v.toString())
        .filter(Boolean);
    } else if (value) {
      const selected = options.find((opt) => opt.value === value);
      return selected ? [selected.label] : [value.toString()];
    }
    return [];
  }

  function handleKeydown(e: KeyboardEvent) {
    if (!isOpen) return;

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        highlightedIndex = Math.min(highlightedIndex + 1, filteredOptions.length - 1);
        break;
      case 'ArrowUp':
        e.preventDefault();
        highlightedIndex = Math.max(highlightedIndex - 1, -1);
        break;
      case 'Enter':
        e.preventDefault();
        if (highlightedIndex >= 0) {
          selectOption(filteredOptions[highlightedIndex]!);
        }
        break;
      case 'Escape':
        e.preventDefault();
        isOpen = false;
        break;
    }
  }

  $: if (isOpen && searchQuery === '') {
    filteredOptions = options;
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
        'flex flex-wrap items-center gap-2 border rounded-md transition-colors',
        'bg-white',
        sizeClasses[size as keyof typeof sizeClasses] ?? sizeClasses.md,
        stateClasses[state as keyof typeof stateClasses] ?? stateClasses.default,
        disabled && 'bg-neutral-50 cursor-not-allowed opacity-50'
      )}
    >
      {#if multiple && Array.isArray(value) && value.length > 0}
        {#each getSelectedLabels() as label, idx}
          <div class="flex items-center gap-1 bg-primary-100 text-primary-700 px-2 py-0.5 rounded text-sm">
            <span>{label}</span>
            <button
              type="button"
              on:click={() => removeOption(Array.isArray(value) ? value[idx]! : value as string | number)}
              disabled={disabled || readonly}
              class="text-primary-700 hover:text-primary-900"
            >
              <Icon name="x" size="sm" />
            </button>
          </div>
        {/each}
      {/if}

      <input
        {id}
        {name}
        {placeholder}
        {disabled}
        {readonly}
        type="text"
        value={searchQuery}
        on:input={(e) => handleSearch((e.target as HTMLInputElement).value)}
        on:keydown={handleKeydown}
        on:focus={() => (isOpen = true)}
        on:blur={() => setTimeout(() => (isOpen = false), 200)}
        class={cn(
          'flex-1 bg-transparent border-0 outline-none',
          'placeholder-neutral-400 text-neutral-900',
          disabled && 'cursor-not-allowed'
        )}
        autocomplete="off"
      />

      <div class="flex gap-1 flex-shrink-0">
        {#if isLoading}
          <Icon name="loader" size="sm" class="text-neutral-400 animate-spin" />
        {/if}

        {#if clearable && (value || searchQuery)}
          <button
            type="button"
            on:click={handleClear}
            disabled={disabled || readonly}
            class="text-neutral-400 hover:text-neutral-600 p-1"
          >
            <Icon name="x" size="sm" />
          </button>
        {/if}

        <Icon name={isOpen ? 'chevron-up' : 'chevron-down'} size="sm" class="text-neutral-400" />
      </div>
    </div>

    {#if isOpen && filteredOptions.length > 0}
      <div
        class="absolute z-10 w-full mt-1 bg-white border border-neutral-200 rounded-md shadow-lg max-h-60 overflow-y-auto"
      >
        {#each filteredOptions as option, idx}
          <button
            type="button"
            on:click={() => selectOption(option)}
            class={cn(
              'w-full text-left px-4 py-2 transition-colors',
              idx === highlightedIndex ? 'bg-primary-100' : 'hover:bg-neutral-100',
              Array.isArray(value) && value.includes(option.value) && 'bg-primary-50'
            )}
          >
            <div class="font-medium text-sm">{option.label}</div>
            {#if option.description}
              <div class="text-xs text-neutral-500">{option.description}</div>
            {/if}
          </button>
        {/each}
      </div>
    {/if}

    {#if isOpen && filteredOptions.length === 0 && searchQuery}
      <div class="absolute z-10 w-full mt-1 bg-white border border-neutral-200 rounded-md shadow-lg p-4 text-center text-neutral-500">
        No results found
      </div>
    {/if}
  </div>

  {#if errorText}
    <p class="mt-1 text-sm text-red-500">{errorText}</p>
  {:else if helperText}
    <p class="mt-1 text-sm text-neutral-500">{helperText}</p>
  {/if}
</div>
