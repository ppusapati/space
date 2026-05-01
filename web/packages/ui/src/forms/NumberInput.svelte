<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import type { Size } from '../types';

  // Props
  export let value: number | null = null;
  export let disabled: boolean = false;
  export let size: Size = 'md';
  export let label: string = '';
  export let placeholder: string = '';
  export let min: number | undefined = undefined;
  export let max: number | undefined = undefined;
  export let step: number = 1;
  export let precision: number = 2;
  export let showStepper: boolean = true;
  export let thousandSeparator: string = ',';
  export let decimalSeparator: string = '.';
  export let prefix: string = '';
  export let suffix: string = '';
  export let name: string = '';
  export let id: string = uid('number');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    input: { value: number | null };
    change: { value: number | null };
  }>();

  let displayValue = '';
  let isFocused = false;

  // Size configurations
  const sizeConfig = {
    sm: { input: 'h-8 text-sm', button: 'w-6 text-sm' },
    md: { input: 'h-10 text-base', button: 'w-8 text-base' },
    lg: { input: 'h-12 text-lg', button: 'w-10 text-lg' },
  };

  $: config = sizeConfig[size as keyof typeof sizeConfig] ?? sizeConfig.md;

  function formatNumber(num: number | null): string {
    if (num === null) return '';

    const fixed = num.toFixed(precision);
    const [intPart, decPart] = fixed.split('.');

    const formattedInt = intPart!.replace(/\B(?=(\d{3})+(?!\d))/g, thousandSeparator);

    return decPart ? `${formattedInt}${decimalSeparator}${decPart}` : formattedInt;
  }

  function parseNumber(str: string): number | null {
    if (!str) return null;

    // Remove formatting
    const cleaned = str
      .replace(new RegExp(`\\${thousandSeparator}`, 'g'), '')
      .replace(decimalSeparator, '.');

    const num = parseFloat(cleaned);
    return isNaN(num) ? null : num;
  }

  function clampValue(num: number | null): number | null {
    if (num === null) return null;
    if (min !== undefined && num < min) return min;
    if (max !== undefined && num > max) return max;
    return num;
  }

  // Update display when value changes externally
  $: if (!isFocused) {
    displayValue = formatNumber(value);
  }

  function handleInput(event: Event) {
    const target = event.target as HTMLInputElement;
    displayValue = target.value;

    const parsed = parseNumber(target.value);
    if (parsed !== null || target.value === '') {
      value = parsed;
      dispatch('input', { value });
    }
  }

  function handleChange(event: Event) {
    const parsed = parseNumber(displayValue);
    value = clampValue(parsed);
    displayValue = formatNumber(value);
    dispatch('change', { value });
  }

  function handleFocus() {
    isFocused = true;
    // Show raw number on focus for easier editing
    displayValue = value !== null ? String(value) : '';
  }

  function handleBlur() {
    isFocused = false;
    const parsed = parseNumber(displayValue);
    value = clampValue(parsed);
    displayValue = formatNumber(value);
    dispatch('change', { value });
  }

  function increment() {
    if (disabled) return;
    const newValue = (value ?? 0) + step;
    value = clampValue(newValue);
    displayValue = formatNumber(value);
    dispatch('change', { value });
  }

  function decrement() {
    if (disabled) return;
    const newValue = (value ?? 0) - step;
    value = clampValue(newValue);
    displayValue = formatNumber(value);
    dispatch('change', { value });
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'ArrowUp') {
      event.preventDefault();
      increment();
    } else if (event.key === 'ArrowDown') {
      event.preventDefault();
      decrement();
    }
  }
</script>

<div class={cn('w-full', className)}>
  {#if label}
    <label
      for={id}
      class="block mb-2 font-medium text-[var(--color-text-primary)]"
    >
      {label}
    </label>
  {/if}

  <div class="flex">
    {#if showStepper}
      <button
        type="button"
        class={cn(
          'flex items-center justify-center rounded-l-lg',
          'border border-r-0 border-[var(--color-border-primary)]',
          'bg-[var(--color-surface-secondary)] text-[var(--color-text-secondary)]',
          'hover:bg-[var(--color-surface-tertiary)]',
          'focus:outline-none focus:ring-2 focus:ring-inset focus:ring-[var(--color-interactive-primary)]',
          config.button,
          config.input,
          disabled && 'opacity-50 cursor-not-allowed'
        )}
        on:click={decrement}
        {disabled}
        aria-label="Decrease"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" />
        </svg>
      </button>
    {/if}

    <div class="relative flex-1">
      {#if prefix}
        <span class="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-tertiary)]">
          {prefix}
        </span>
      {/if}

      <input
        type="text"
        inputmode="decimal"
        {id}
        {name}
        value={displayValue}
        {disabled}
        {placeholder}
        class={cn(
          'w-full border border-[var(--color-border-primary)]',
          'bg-[var(--color-surface-primary)] text-[var(--color-text-primary)]',
          'text-right font-mono tabular-nums',
          'focus:outline-none focus:ring-2 focus:ring-[var(--color-interactive-primary)]',
          config.input,
          showStepper ? 'rounded-none' : 'rounded-lg',
          prefix && 'pl-8',
          suffix && 'pr-8',
          disabled && 'opacity-50 cursor-not-allowed'
        )}
        data-testid={testId || undefined}
        on:input={handleInput}
        on:change={handleChange}
        on:focus={handleFocus}
        on:blur={handleBlur}
        on:keydown={handleKeydown}
      />

      {#if suffix}
        <span class="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--color-text-tertiary)]">
          {suffix}
        </span>
      {/if}
    </div>

    {#if showStepper}
      <button
        type="button"
        class={cn(
          'flex items-center justify-center rounded-r-lg',
          'border border-l-0 border-[var(--color-border-primary)]',
          'bg-[var(--color-surface-secondary)] text-[var(--color-text-secondary)]',
          'hover:bg-[var(--color-surface-tertiary)]',
          'focus:outline-none focus:ring-2 focus:ring-inset focus:ring-[var(--color-interactive-primary)]',
          config.button,
          config.input,
          disabled && 'opacity-50 cursor-not-allowed'
        )}
        on:click={increment}
        {disabled}
        aria-label="Increase"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
      </button>
    {/if}
  </div>
</div>
