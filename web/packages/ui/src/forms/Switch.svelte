<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import type { Size } from '../types';

  // Props
  export let checked: boolean = false;
  export let disabled: boolean = false;
  export let size: Size = 'md';
  export let label: string = '';
  export let labelPosition: 'left' | 'right' = 'right';
  export let name: string = '';
  export let id: string = uid('switch');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: { checked: boolean };
  }>();

  // Size configurations
  const sizeConfig = {
    sm: {
      track: 'w-8 h-4',
      thumb: 'w-3 h-3',
      translate: 'translate-x-4',
      label: 'text-sm',
    },
    md: {
      track: 'w-11 h-6',
      thumb: 'w-5 h-5',
      translate: 'translate-x-5',
      label: 'text-base',
    },
    lg: {
      track: 'w-14 h-7',
      thumb: 'w-6 h-6',
      translate: 'translate-x-7',
      label: 'text-lg',
    },
  };

  $: config = sizeConfig[size as keyof typeof sizeConfig] ?? sizeConfig.md;

  $: trackClasses = cn(
    'relative inline-flex shrink-0 cursor-pointer rounded-full border-2 border-transparent',
    'transition-colors duration-200 ease-in-out',
    'focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2',
    'focus-visible:ring-[var(--color-interactive-primary)]',
    config.track,
    checked
      ? 'bg-[var(--color-interactive-primary)]'
      : 'bg-[var(--color-neutral-300)]',
    disabled && 'opacity-50 cursor-not-allowed',
    className
  );

  $: thumbClasses = cn(
    'pointer-events-none inline-block rounded-full bg-white shadow-lg ring-0',
    'transform transition duration-200 ease-in-out',
    config.thumb,
    checked ? config.translate : 'translate-x-0'
  );

  $: labelClasses = cn(
    'font-medium text-[var(--color-text-primary)]',
    config.label,
    disabled && 'opacity-50'
  );

  function handleChange() {
    if (disabled) return;
    checked = !checked;
    dispatch('change', { checked });
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      handleChange();
    }
  }
</script>

<label
  class="inline-flex items-center gap-3 cursor-pointer"
  class:cursor-not-allowed={disabled}
  class:flex-row-reverse={labelPosition === 'left'}
>
  {#if label}
    <span class={labelClasses}>{label}</span>
  {/if}

  <button
    type="button"
    role="switch"
    {id}
    aria-checked={checked}
    aria-label={label || 'Toggle switch'}
    {disabled}
    class={trackClasses}
    data-testid={testId || undefined}
    on:click={handleChange}
    on:keydown={handleKeydown}
  >
    <span aria-hidden="true" class={thumbClasses} />
  </button>

  {#if name}
    <input type="hidden" {name} value={checked ? 'on' : 'off'} />
  {/if}
</label>
