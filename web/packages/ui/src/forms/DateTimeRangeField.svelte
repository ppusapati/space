<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import Icon from '../display/Icon.svelte';
  import type { Size, ValidationState } from '../types';

  interface DateTimeRange {
    start: string;
    end: string;
  }

  interface DateTimeRangeFieldProps {
    value?: DateTimeRange;
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
  }

  export let value: DateTimeRange = { start: '', end: '' };
  export let label: string = '';
  export let helperText: string = '';
  export let errorText: string = '';
  export let disabled: boolean = false;
  export let readonly: boolean = false;
  export let required: boolean = false;
  export let size: Size = 'md';
  export let state: ValidationState = 'default';
  export let name: string = '';
  export let id: string = uid('datetime-range');
  export let clearable: boolean = true;

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: DateTimeRange;
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

  function handleStartChange(e: Event) {
    const target = e.target as HTMLInputElement;
    value = { ...value, start: target.value };
    dispatch('change', value);
  }

  function handleEndChange(e: Event) {
    const target = e.target as HTMLInputElement;
    value = { ...value, end: target.value };
    dispatch('change', value);
  }

  function handleClear() {
    value = { start: '', end: '' };
    dispatch('change', value);
  }

  function getDisplayValue(): string {
    if (!value.start && !value.end) return '';
    if (value.start && value.end) {
      return `${value.start} → ${value.end}`;
    }
    return value.start || value.end || '';
  }

  function isValidRange(): boolean {
    if (!value.start || !value.end) return true;
    return new Date(value.start) <= new Date(value.end);
  }

  $: isValid = isValidRange();
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
    <!-- Start DateTime -->
    <div>
      <label class="block text-xs font-medium text-neutral-600 mb-1">Start Date & Time</label>
      <div
        class={cn(
          'flex items-center border rounded-md transition-colors',
          'bg-white',
          sizeClasses[size as keyof typeof sizeClasses] ?? sizeClasses.md,
          stateClasses[state as keyof typeof stateClasses] ?? stateClasses.default,
          disabled && 'bg-neutral-50 cursor-not-allowed opacity-50'
        )}
      >
        <Icon name="calendar" class="text-neutral-400 flex-shrink-0 ml-2" />
        <input
          type="datetime-local"
          value={value.start}
          on:change={handleStartChange}
          disabled={disabled || readonly}
          class={cn(
            'flex-1 bg-transparent border-0 outline-none px-2',
            'placeholder-neutral-400 text-neutral-900',
            disabled && 'cursor-not-allowed'
          )}
        />
      </div>
    </div>

    <!-- End DateTime -->
    <div>
      <label class="block text-xs font-medium text-neutral-600 mb-1">End Date & Time</label>
      <div
        class={cn(
          'flex items-center border rounded-md transition-colors',
          'bg-white',
          sizeClasses[size as keyof typeof sizeClasses] ?? sizeClasses.md,
          stateClasses[state as keyof typeof stateClasses] ?? stateClasses.default,
          disabled && 'bg-neutral-50 cursor-not-allowed opacity-50'
        )}
      >
        <Icon name="calendar" class="text-neutral-400 flex-shrink-0 ml-2" />
        <input
          type="datetime-local"
          value={value.end}
          on:change={handleEndChange}
          disabled={disabled || readonly}
          class={cn(
            'flex-1 bg-transparent border-0 outline-none px-2',
            'placeholder-neutral-400 text-neutral-900',
            disabled && 'cursor-not-allowed'
          )}
        />
        {#if clearable && (value.start || value.end)}
          <button
            type="button"
            on:click={handleClear}
            disabled={disabled || readonly}
            class="text-neutral-400 hover:text-neutral-600 flex-shrink-0 p-1 mr-2"
          >
            <Icon name="x" size="sm" />
          </button>
        {/if}
      </div>
    </div>

    <!-- Validation Message -->
    {#if value.start && value.end && !isValid}
      <p class="text-sm text-red-500">End date must be after start date</p>
    {/if}

    <!-- Display Value -->
    {#if getDisplayValue()}
      <div class="p-2 bg-neutral-50 rounded border border-neutral-200 text-sm">
        {getDisplayValue()}
      </div>
    {/if}
  </div>

  {#if errorText}
    <p class="mt-2 text-sm text-red-500">{errorText}</p>
  {:else if helperText}
    <p class="mt-2 text-sm text-neutral-500">{helperText}</p>
  {/if}
</div>
