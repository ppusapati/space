<script context="module" lang="ts">
  export interface Currency {
    code: string;
    symbol: string;
    name: string;
    decimalPlaces: number;
  }
</script>

<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import type { Size } from '../types';

  // Common currencies
  const CURRENCIES: Currency[] = [
    { code: 'USD', symbol: '$', name: 'US Dollar', decimalPlaces: 2 },
    { code: 'EUR', symbol: '€', name: 'Euro', decimalPlaces: 2 },
    { code: 'GBP', symbol: '£', name: 'British Pound', decimalPlaces: 2 },
    { code: 'INR', symbol: '₹', name: 'Indian Rupee', decimalPlaces: 2 },
    { code: 'JPY', symbol: '¥', name: 'Japanese Yen', decimalPlaces: 0 },
    { code: 'CNY', symbol: '¥', name: 'Chinese Yuan', decimalPlaces: 2 },
    { code: 'CAD', symbol: 'C$', name: 'Canadian Dollar', decimalPlaces: 2 },
    { code: 'AUD', symbol: 'A$', name: 'Australian Dollar', decimalPlaces: 2 },
    { code: 'CHF', symbol: 'Fr', name: 'Swiss Franc', decimalPlaces: 2 },
  ];

  // Props
  export let value: number | null = null;
  export let currencyCode: string = 'USD';
  export let disabled: boolean = false;
  export let size: Size = 'md';
  export let label: string = '';
  export let placeholder: string = '0.00';
  export let showCurrencySelector: boolean = false;
  export let availableCurrencies: Currency[] = CURRENCIES;
  export let name: string = '';
  export let id: string = uid('currency');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    input: { value: number | null; currencyCode: string };
    change: { value: number | null; currencyCode: string };
  }>();

  let displayValue = '';
  let isFocused = false;
  let showCurrencyDropdown = false;

  // Size configurations
  const sizeConfig = {
    sm: { input: 'h-8 text-sm', selector: 'text-sm px-2' },
    md: { input: 'h-10 text-base', selector: 'text-base px-3' },
    lg: { input: 'h-12 text-lg', selector: 'text-lg px-4' },
  };

  $: config = sizeConfig[size as keyof typeof sizeConfig] ?? sizeConfig.md;
  $: currency = availableCurrencies.find(c => c.code === currencyCode) || CURRENCIES[0]!;

  function formatCurrency(num: number | null): string {
    if (num === null) return '';

    return new Intl.NumberFormat('en-US', {
      minimumFractionDigits: currency.decimalPlaces,
      maximumFractionDigits: currency.decimalPlaces,
    }).format(num);
  }

  function parseNumber(str: string): number | null {
    if (!str) return null;

    // Remove formatting
    const cleaned = str.replace(/[^\d.-]/g, '');
    const num = parseFloat(cleaned);
    return isNaN(num) ? null : num;
  }

  // Update display when value changes externally
  $: if (!isFocused) {
    displayValue = formatCurrency(value);
  }

  function handleInput(event: Event) {
    const target = event.target as HTMLInputElement;
    displayValue = target.value;

    const parsed = parseNumber(target.value);
    if (parsed !== null || target.value === '') {
      value = parsed;
      dispatch('input', { value, currencyCode });
    }
  }

  function handleFocus() {
    isFocused = true;
    displayValue = value !== null ? String(value) : '';
  }

  function handleBlur() {
    isFocused = false;
    value = parseNumber(displayValue);
    displayValue = formatCurrency(value);
    dispatch('change', { value, currencyCode });
  }

  function selectCurrency(code: string) {
    currencyCode = code;
    showCurrencyDropdown = false;
    displayValue = formatCurrency(value);
    dispatch('change', { value, currencyCode });
  }

  function handleKeydown(event: KeyboardEvent) {
    // Allow numbers, decimal, minus, backspace, delete, arrows, tab
    const allowed = ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', '-',
                     'Backspace', 'Delete', 'ArrowLeft', 'ArrowRight', 'Tab', 'Enter'];
    if (!allowed.includes(event.key) && !event.ctrlKey && !event.metaKey) {
      event.preventDefault();
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
    {#if showCurrencySelector}
      <div class="relative">
        <button
          type="button"
          class={cn(
            'flex items-center justify-between gap-1 rounded-l-lg',
            'border border-r-0 border-[var(--color-border-primary)]',
            'bg-[var(--color-surface-secondary)] text-[var(--color-text-primary)]',
            'hover:bg-[var(--color-surface-tertiary)]',
            'focus:outline-none focus:ring-2 focus:ring-inset focus:ring-[var(--color-interactive-primary)]',
            config.input,
            config.selector,
            disabled && 'opacity-50 cursor-not-allowed'
          )}
          on:click={() => showCurrencyDropdown = !showCurrencyDropdown}
          {disabled}
        >
          <span>{currency.code}</span>
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>

        {#if showCurrencyDropdown && !disabled}
          <div class="absolute z-50 mt-1 w-48 max-h-60 overflow-auto bg-[var(--color-surface-primary)] border border-[var(--color-border-primary)] rounded-lg shadow-lg">
            {#each availableCurrencies as curr}
              <button
                type="button"
                class={cn(
                  'w-full px-3 py-2 text-left text-sm',
                  'hover:bg-[var(--color-surface-secondary)]',
                  'focus:outline-none focus:bg-[var(--color-surface-secondary)]',
                  curr.code === currencyCode && 'bg-[var(--color-interactive-primary)] text-white'
                )}
                on:click={() => selectCurrency(curr.code)}
              >
                <span class="font-medium">{curr.code}</span>
                <span class="text-[var(--color-text-tertiary)] ml-2">{curr.symbol}</span>
                <span class="block text-xs text-[var(--color-text-tertiary)]">{curr.name}</span>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    {:else}
      <span
        class={cn(
          'flex items-center justify-center rounded-l-lg',
          'border border-r-0 border-[var(--color-border-primary)]',
          'bg-[var(--color-surface-secondary)] text-[var(--color-text-secondary)]',
          'px-3',
          config.input
        )}
      >
        {currency.symbol}
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
        'flex-1 rounded-r-lg border border-[var(--color-border-primary)]',
        'bg-[var(--color-surface-primary)] text-[var(--color-text-primary)]',
        'text-right font-mono tabular-nums',
        'focus:outline-none focus:ring-2 focus:ring-[var(--color-interactive-primary)]',
        config.input,
        'px-3',
        disabled && 'opacity-50 cursor-not-allowed'
      )}
      data-testid={testId || undefined}
      on:input={handleInput}
      on:focus={handleFocus}
      on:blur={handleBlur}
      on:keydown={handleKeydown}
    />
  </div>
</div>
