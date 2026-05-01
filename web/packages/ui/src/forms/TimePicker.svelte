<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import type { Size } from '../types';

  // Props
  export let value: string = '';
  export let disabled: boolean = false;
  export let size: Size = 'md';
  export let label: string = '';
  export let placeholder: string = 'Select time';
  export let use24Hour: boolean = false;
  export let step: number = 15; // minutes
  export let minTime: string = '';
  export let maxTime: string = '';
  export let name: string = '';
  export let id: string = uid('time');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    input: { value: string };
    change: { value: string };
  }>();

  let isOpen = false;
  let inputRef: HTMLInputElement;

  // Size configurations
  const sizeConfig = {
    sm: { input: 'h-8 text-sm px-2', dropdown: 'text-sm' },
    md: { input: 'h-10 text-base px-3', dropdown: 'text-base' },
    lg: { input: 'h-12 text-lg px-4', dropdown: 'text-lg' },
  };

  $: config = sizeConfig[size as keyof typeof sizeConfig] ?? sizeConfig.md;

  // Generate time options
  function generateTimeOptions(): string[] {
    const options: string[] = [];
    for (let h = 0; h < 24; h++) {
      for (let m = 0; m < 60; m += step) {
        const hour = String(h).padStart(2, '0');
        const minute = String(m).padStart(2, '0');
        const time = `${hour}:${minute}`;

        if (minTime && time < minTime) continue;
        if (maxTime && time > maxTime) continue;

        options.push(time);
      }
    }
    return options;
  }

  $: timeOptions = generateTimeOptions();

  function formatTime(time: string): string {
    if (!time) return '';

    const parts = time.split(':').map(Number);
    const hours = parts[0] ?? 0;
    const minutes = parts[1] ?? 0;

    if (use24Hour) {
      return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`;
    }

    const period = hours >= 12 ? 'PM' : 'AM';
    const displayHour = hours % 12 || 12;
    return `${displayHour}:${String(minutes).padStart(2, '0')} ${period}`;
  }

  function handleInputChange(event: Event) {
    const target = event.target as HTMLInputElement;
    value = target.value;
    dispatch('change', { value });
  }

  function selectTime(time: string) {
    value = time;
    isOpen = false;
    dispatch('change', { value });
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      isOpen = false;
    }
  }

  function handleClickOutside(event: MouseEvent) {
    const target = event.target as HTMLElement;
    if (!target.closest('.time-picker-container')) {
      isOpen = false;
    }
  }
</script>

<svelte:window on:click={handleClickOutside} on:keydown={handleKeydown} />

<div class={cn('w-full', className)}>
  {#if label}
    <label
      for={id}
      class="block mb-2 font-medium text-[var(--color-text-primary)]"
    >
      {label}
    </label>
  {/if}

  <div class="time-picker-container relative">
    <div class="relative">
      <input
        type="time"
        bind:this={inputRef}
        {id}
        {name}
        {value}
        {disabled}
        {placeholder}
        class={cn(
          'w-full rounded-lg border border-[var(--color-border-primary)]',
          'bg-[var(--color-surface-primary)] text-[var(--color-text-primary)]',
          'focus:outline-none focus:ring-2 focus:ring-[var(--color-interactive-primary)]',
          config.input,
          disabled && 'opacity-50 cursor-not-allowed'
        )}
        data-testid={testId || undefined}
        on:change={handleInputChange}
        on:click={() => isOpen = true}
      />
      <button
        type="button"
        class="absolute right-2 top-1/2 -translate-y-1/2 text-[var(--color-text-tertiary)]"
        on:click={() => isOpen = !isOpen}
        {disabled}
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </button>
    </div>

    {#if isOpen && !disabled}
      <div
        class={cn(
          'absolute z-50 mt-1 w-full max-h-60 overflow-auto',
          'bg-[var(--color-surface-primary)] border border-[var(--color-border-primary)]',
          'rounded-lg shadow-lg',
          config.dropdown
        )}
      >
        {#each timeOptions as time}
          <button
            type="button"
            class={cn(
              'w-full px-3 py-2 text-left',
              'hover:bg-[var(--color-surface-secondary)]',
              'focus:outline-none focus:bg-[var(--color-surface-secondary)]',
              value === time && 'bg-[var(--color-interactive-primary)] text-white'
            )}
            on:click={() => selectTime(time)}
          >
            {formatTime(time)}
          </button>
        {/each}
      </div>
    {/if}
  </div>
</div>
