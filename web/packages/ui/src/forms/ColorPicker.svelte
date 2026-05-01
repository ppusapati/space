<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import type { Size } from '../types';

  // Props
  export let value: string = '#000000';
  export let disabled: boolean = false;
  export let size: Size = 'md';
  export let label: string = '';
  export let showInput: boolean = true;
  export let showPresets: boolean = true;
  export let presets: string[] = [
    '#ef4444', '#f97316', '#f59e0b', '#eab308', '#84cc16',
    '#22c55e', '#14b8a6', '#06b6d4', '#0ea5e9', '#3b82f6',
    '#6366f1', '#8b5cf6', '#a855f7', '#d946ef', '#ec4899',
    '#f43f5e', '#78716c', '#71717a', '#737373', '#000000',
  ];
  export let name: string = '';
  export let id: string = uid('color');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    input: { value: string };
    change: { value: string };
  }>();

  let inputRef: HTMLInputElement;
  let isOpen = false;

  // Size configurations
  const sizeConfig = {
    sm: { swatch: 'w-6 h-6', input: 'text-sm h-8', preset: 'w-5 h-5' },
    md: { swatch: 'w-8 h-8', input: 'text-base h-10', preset: 'w-6 h-6' },
    lg: { swatch: 'w-10 h-10', input: 'text-lg h-12', preset: 'w-7 h-7' },
  };

  $: config = sizeConfig[size as keyof typeof sizeConfig] ?? sizeConfig.md;

  function handleColorInput(event: Event) {
    const target = event.target as HTMLInputElement;
    value = target.value;
    dispatch('input', { value });
  }

  function handleColorChange(event: Event) {
    const target = event.target as HTMLInputElement;
    value = target.value;
    dispatch('change', { value });
  }

  function handleTextInput(event: Event) {
    const target = event.target as HTMLInputElement;
    let newValue = target.value;

    // Ensure it starts with #
    if (!newValue.startsWith('#')) {
      newValue = '#' + newValue;
    }

    // Validate hex format
    if (/^#[0-9A-Fa-f]{0,6}$/.test(newValue)) {
      value = newValue;
      if (newValue.length === 7) {
        dispatch('change', { value });
      }
    }
  }

  function selectPreset(color: string) {
    value = color;
    dispatch('change', { value });
  }

  function openColorPicker() {
    inputRef?.click();
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

  <div class="flex items-center gap-3">
    <!-- Color swatch with hidden native picker -->
    <button
      type="button"
      class={cn(
        'relative rounded-lg border-2 border-[var(--color-border-primary)]',
        'cursor-pointer overflow-hidden shadow-sm',
        'focus:outline-none focus:ring-2 focus:ring-[var(--color-interactive-primary)]',
        config.swatch,
        disabled && 'opacity-50 cursor-not-allowed'
      )}
      style="background-color: {value}"
      on:click={openColorPicker}
      {disabled}
      aria-label="Pick color"
    >
      <input
        bind:this={inputRef}
        type="color"
        {id}
        {name}
        {value}
        {disabled}
        class="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
        data-testid={testId || undefined}
        on:input={handleColorInput}
        on:change={handleColorChange}
      />
    </button>

    <!-- Hex input -->
    {#if showInput}
      <input
        type="text"
        value={value}
        {disabled}
        maxlength="7"
        class={cn(
          'flex-1 px-3 rounded-lg border border-[var(--color-border-primary)]',
          'bg-[var(--color-surface-primary)] text-[var(--color-text-primary)]',
          'font-mono uppercase',
          'focus:outline-none focus:ring-2 focus:ring-[var(--color-interactive-primary)]',
          config.input,
          disabled && 'opacity-50 cursor-not-allowed'
        )}
        on:input={handleTextInput}
      />
    {/if}
  </div>

  <!-- Preset colors -->
  {#if showPresets && presets.length > 0}
    <div class="mt-3">
      <div class="text-xs text-[var(--color-text-tertiary)] mb-2">Presets</div>
      <div class="flex flex-wrap gap-2">
        {#each presets as color}
          <button
            type="button"
            class={cn(
              'rounded-md border-2 transition-transform hover:scale-110',
              'focus:outline-none focus:ring-2 focus:ring-[var(--color-interactive-primary)]',
              config.preset,
              value === color
                ? 'border-[var(--color-interactive-primary)]'
                : 'border-transparent'
            )}
            style="background-color: {color}"
            on:click={() => selectPreset(color)}
            {disabled}
            aria-label="Select {color}"
          />
        {/each}
      </div>
    </div>
  {/if}
</div>
