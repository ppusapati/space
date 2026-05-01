<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import type { Size } from '../types';

  // Props
  export let value: number = 0;
  export let min: number = 0;
  export let max: number = 100;
  export let step: number = 1;
  export let disabled: boolean = false;
  export let size: Size = 'md';
  export let label: string = '';
  export let showValue: boolean = true;
  export let showMinMax: boolean = false;
  export let showTicks: boolean = false;
  export let tickCount: number = 5;
  export let formatValue: (val: number) => string = (val) => String(val);
  export let name: string = '';
  export let id: string = uid('slider');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    input: { value: number };
    change: { value: number };
  }>();

  // Size configurations
  const sizeConfig = {
    sm: { track: 'h-1', thumb: 'w-4 h-4', label: 'text-sm' },
    md: { track: 'h-2', thumb: 'w-5 h-5', label: 'text-base' },
    lg: { track: 'h-3', thumb: 'w-6 h-6', label: 'text-lg' },
  };

  $: config = sizeConfig[size as keyof typeof sizeConfig] ?? sizeConfig.md;
  $: percentage = ((value - min) / (max - min)) * 100;
  $: ticks = showTicks
    ? Array.from({ length: tickCount }, (_, i) => min + (i * (max - min)) / (tickCount - 1))
    : [];

  function handleInput(event: Event) {
    const target = event.target as HTMLInputElement;
    value = parseFloat(target.value);
    dispatch('input', { value });
  }

  function handleChange(event: Event) {
    const target = event.target as HTMLInputElement;
    value = parseFloat(target.value);
    dispatch('change', { value });
  }
</script>

<div class={cn('w-full', className)}>
  {#if label || showValue}
    <div class="flex justify-between items-center mb-2">
      {#if label}
        <label for={id} class={cn('font-medium text-[var(--color-text-primary)]', config.label)}>
          {label}
        </label>
      {/if}
      {#if showValue}
        <span class="text-[var(--color-text-secondary)] font-medium tabular-nums">
          {formatValue(value)}
        </span>
      {/if}
    </div>
  {/if}

  <div class="relative">
    <!-- Track background -->
    <div
      class={cn(
        'absolute w-full rounded-full bg-[var(--color-neutral-200)]',
        config.track,
        'top-1/2 -translate-y-1/2'
      )}
    />

    <!-- Filled track -->
    <div
      class={cn(
        'absolute rounded-full bg-[var(--color-interactive-primary)]',
        config.track,
        'top-1/2 -translate-y-1/2'
      )}
      style="width: {percentage}%"
    />

    <!-- Native input -->
    <input
      type="range"
      {id}
      {name}
      {min}
      {max}
      {step}
      {value}
      {disabled}
      class={cn(
        'relative w-full appearance-none bg-transparent cursor-pointer',
        'focus:outline-none',
        '[&::-webkit-slider-thumb]:appearance-none',
        '[&::-webkit-slider-thumb]:rounded-full',
        '[&::-webkit-slider-thumb]:bg-white',
        '[&::-webkit-slider-thumb]:shadow-md',
        '[&::-webkit-slider-thumb]:border-2',
        '[&::-webkit-slider-thumb]:border-[var(--color-interactive-primary)]',
        '[&::-webkit-slider-thumb]:cursor-pointer',
        '[&::-webkit-slider-thumb]:transition-transform',
        '[&::-webkit-slider-thumb]:hover:scale-110',
        size === 'sm' && '[&::-webkit-slider-thumb]:w-4 [&::-webkit-slider-thumb]:h-4',
        size === 'md' && '[&::-webkit-slider-thumb]:w-5 [&::-webkit-slider-thumb]:h-5',
        size === 'lg' && '[&::-webkit-slider-thumb]:w-6 [&::-webkit-slider-thumb]:h-6',
        '[&::-moz-range-thumb]:appearance-none',
        '[&::-moz-range-thumb]:rounded-full',
        '[&::-moz-range-thumb]:bg-white',
        '[&::-moz-range-thumb]:shadow-md',
        '[&::-moz-range-thumb]:border-2',
        '[&::-moz-range-thumb]:border-[var(--color-interactive-primary)]',
        '[&::-moz-range-thumb]:cursor-pointer',
        disabled && 'opacity-50 cursor-not-allowed'
      )}
      data-testid={testId || undefined}
      on:input={handleInput}
      on:change={handleChange}
    />

    <!-- Ticks -->
    {#if showTicks && ticks.length > 0}
      <div class="relative mt-1">
        {#each ticks as tick, i}
          <div
            class="absolute w-1 h-1 rounded-full bg-[var(--color-neutral-400)] -translate-x-1/2"
            style="left: {((tick - min) / (max - min)) * 100}%"
          />
        {/each}
      </div>
    {/if}
  </div>

  {#if showMinMax}
    <div class="flex justify-between mt-1 text-xs text-[var(--color-text-tertiary)]">
      <span>{formatValue(min)}</span>
      <span>{formatValue(max)}</span>
    </div>
  {/if}
</div>
