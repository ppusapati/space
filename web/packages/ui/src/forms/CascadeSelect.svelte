<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import Select from './Select.svelte';
  import type { Size, ValidationState } from '../types';

  interface CascadeOption {
    value: string | number;
    label: string;
    children?: CascadeOption[];
  }

  interface CascadeSelectProps {
    value?: (string | number)[];
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
    options?: CascadeOption[];
    placeholder?: string;
    clearable?: boolean;
  }

  export let value: (string | number)[] = [];
  export let label: string = '';
  export let helperText: string = '';
  export let errorText: string = '';
  export let disabled: boolean = false;
  export let readonly: boolean = false;
  export let required: boolean = false;
  export let size: Size = 'md';
  export let state: ValidationState = 'default';
  export let name: string = '';
  export let id: string = uid('cascade-select');
  export let options: CascadeOption[] = [];
  export let placeholder: string = 'Select...';
  export let clearable: boolean = true;

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: (string | number)[];
  }>();

  let selectedValues: (string | number)[] = [...value];
  let cascadeOptions: (CascadeOption[] | undefined)[] = [];

  function initializeCascade() {
    cascadeOptions = [options];

    // Find selected path
    for (let i = 0; i < selectedValues.length; i++) {
      const currentLevel = cascadeOptions[i];
      if (currentLevel) {
        const selected = currentLevel.find((opt) => opt.value === selectedValues[i]);
        if (selected && selected.children) {
          cascadeOptions[i + 1] = selected.children;
        } else {
          cascadeOptions = cascadeOptions.slice(0, i + 1);
          break;
        }
      }
    }
  }

  function handleSelect(level: number, selectedValue: string | number) {
    selectedValues = selectedValues.slice(0, level);
    selectedValues[level] = selectedValue;
    selectedValues = selectedValues;

    // Update cascade options
    let current = options;
    for (let i = 0; i <= level; i++) {
      cascadeOptions[i] = current;
      const selected = current.find((opt) => opt.value === selectedValues[i]);
      if (selected && selected.children) {
        current = selected.children;
      } else {
        cascadeOptions = cascadeOptions.slice(0, i + 1);
        break;
      }
    }

    cascadeOptions = cascadeOptions;
    value = selectedValues;
    dispatch('change', value);
  }

  function handleClear() {
    selectedValues = [];
    value = [];
    initializeCascade();
    dispatch('change', value);
  }

  function getSelectedLabels(): string[] {
    const labels: string[] = [];
    let currentOptions = options;

    for (const val of selectedValues) {
      const option = currentOptions.find((opt) => opt.value === val);
      if (option) {
        labels.push(option.label);
        if (option.children) {
          currentOptions = option.children;
        } else {
          break;
        }
      }
    }

    return labels;
  }

  $: {
    value;
    selectedValues = [...value];
    initializeCascade();
  }
</script>

<div class={cn('w-full', className)}>
  {#if label}
    <label class="block text-sm font-medium text-neutral-700 mb-1">
      {label}
      {#if required}
        <span class="text-red-500 ml-1">*</span>
      {/if}
    </label>
  {/if}

  <div class="space-y-2">
    {#each cascadeOptions as levelOptions, level}
      {#if levelOptions}
        <div class="flex gap-2 items-center">
          <Select
            value={selectedValues[level] || ''}
            on:change={(e) => handleSelect(level, e.detail.value as string | number)}
            disabled={disabled || readonly}
            {size}
            {state}
            {placeholder}
            options={levelOptions.map((opt) => ({
              value: opt.value,
              label: opt.label,
            }))}
          />
          {#if selectedValues[level]}
            <span class="text-xs bg-primary-100 text-primary-700 px-2 py-1 rounded">
              Level {level + 1}
            </span>
          {/if}
        </div>
      {/if}
    {/each}
  </div>

  {#if selectedValues.length > 0 && clearable}
    <button
      type="button"
      on:click={handleClear}
      disabled={disabled || readonly}
      class="mt-2 px-3 py-1.5 text-xs bg-red-100 text-red-700 hover:bg-red-200 rounded disabled:opacity-50"
    >
      Clear Selection
    </button>
  {/if}

  {#if selectedValues.length > 0}
    <div class="mt-2 p-2 bg-neutral-50 rounded border border-neutral-200">
      <p class="text-xs font-medium text-neutral-600 mb-1">Selected Path:</p>
      <div class="text-sm text-neutral-700">
        {getSelectedLabels().join(' > ')}
      </div>
    </div>
  {/if}

  {#if errorText}
    <p class="mt-1 text-sm text-red-500">{errorText}</p>
  {:else if helperText}
    <p class="mt-1 text-sm text-neutral-500">{helperText}</p>
  {/if}
</div>
