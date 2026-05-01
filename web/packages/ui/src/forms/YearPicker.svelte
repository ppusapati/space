<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import Icon from '../display/Icon.svelte';
  import type { Size, ValidationState } from '../types';

  interface YearPickerProps {
    value?: string | number;
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
    minYear?: number;
    maxYear?: number;
    clearable?: boolean;
  }

  export let value: string | number = '';
  export let label: string = '';
  export let helperText: string = '';
  export let errorText: string = '';
  export let disabled: boolean = false;
  export let readonly: boolean = false;
  export let required: boolean = false;
  export let size: Size = 'md';
  export let state: ValidationState = 'default';
  export let name: string = '';
  export let id: string = uid('year-picker');
  export let minYear: number = 1900;
  export let maxYear: number = new Date().getFullYear() + 10;
  export let clearable: boolean = true;

  let className: string = '';
  export { className as class };

  let isOpen = false;
  let displayDecade = Math.floor((new Date().getFullYear()) / 10) * 10;

  const dispatch = createEventDispatcher<{
    change: string | number;
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

  function selectYear(year: number) {
    value = year;
    isOpen = false;
    dispatch('change', value);
  }

  function previousDecade() {
    displayDecade -= 10;
  }

  function nextDecade() {
    displayDecade += 10;
  }

  function handleClear() {
    value = '';
    dispatch('change', value);
  }

  function getYearRange(): number[] {
    const years: number[] = [];
    for (let i = displayDecade; i < displayDecade + 10; i++) {
      if (i >= minYear && i <= maxYear) {
        years.push(i);
      }
    }
    return years;
  }

  function isSelected(year: number): boolean {
    return Number(value) === year;
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
        'flex items-center border rounded-md transition-colors cursor-pointer',
        'bg-white',
        sizeClasses[size as keyof typeof sizeClasses] ?? sizeClasses.md,
        stateClasses[state as keyof typeof stateClasses] ?? stateClasses.default,
        disabled && 'bg-neutral-50 cursor-not-allowed opacity-50'
      )}
      on:click={() => !disabled && !readonly && (isOpen = !isOpen)}
    >
      <Icon name="calendar" class="text-neutral-400 flex-shrink-0" />

      <input
        {id}
        {name}
        type="text"
        value={value || ''}
        {disabled}
        {readonly}
        on:focus={() => !disabled && !readonly && (isOpen = true)}
        on:blur={() => setTimeout(() => (isOpen = false), 200)}
        class={cn(
          'flex-1 bg-transparent border-0 outline-none px-2',
          'placeholder-neutral-400 text-neutral-900',
          disabled && 'cursor-not-allowed'
        )}
        placeholder="Select year..."
      />

      {#if clearable && value}
        <button
          type="button"
          on:click={(e) => {
            e.stopPropagation();
            handleClear();
          }}
          disabled={disabled || readonly}
          class="text-neutral-400 hover:text-neutral-600 flex-shrink-0 p-1"
        >
          <Icon name="x" size="sm" />
        </button>
      {/if}

      <Icon
        name={isOpen ? 'chevron-up' : 'chevron-down'}
        size="sm"
        class="text-neutral-400 flex-shrink-0"
      />
    </div>

    {#if isOpen}
      <div class="absolute z-10 w-full mt-1 bg-white border border-neutral-200 rounded-md shadow-lg p-4">
        <!-- Decade Navigation -->
        <div class="flex items-center justify-between mb-4">
          <button
            type="button"
            on:click={previousDecade}
            class="p-1 hover:bg-neutral-100 rounded"
          >
            <Icon name="chevron-left" size="sm" />
          </button>
          <span class="font-medium text-neutral-900 text-sm">
            {displayDecade}-{displayDecade + 9}
          </span>
          <button
            type="button"
            on:click={nextDecade}
            class="p-1 hover:bg-neutral-100 rounded"
          >
            <Icon name="chevron-right" size="sm" />
          </button>
        </div>

        <!-- Year Grid -->
        <div class="grid grid-cols-2 gap-2 max-h-48 overflow-y-auto">
          {#each getYearRange() as year}
            <button
              type="button"
              on:click={() => selectYear(year)}
              class={cn(
                'py-2 px-3 rounded text-sm transition-colors',
                isSelected(year)
                  ? 'bg-primary-500 text-white font-medium'
                  : 'hover:bg-neutral-100 text-neutral-700'
              )}
            >
              {year}
            </button>
          {/each}
        </div>
      </div>
    {/if}
  </div>

  {#if errorText}
    <p class="mt-1 text-sm text-red-500">{errorText}</p>
  {:else if helperText}
    <p class="mt-1 text-sm text-neutral-500">{helperText}</p>
  {/if}
</div>
