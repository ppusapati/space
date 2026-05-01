<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import type { Size } from '../types';

  // Props
  export let value: number = 0;
  export let max: number = 5;
  export let disabled: boolean = false;
  export let readonly: boolean = false;
  export let size: Size = 'md';
  export let label: string = '';
  export let showValue: boolean = false;
  export let allowHalf: boolean = false;
  export let allowClear: boolean = true;
  export let icon: 'star' | 'heart' | 'circle' = 'star';
  export let activeColor: string = 'var(--color-warning)';
  export let inactiveColor: string = 'var(--color-neutral-300)';
  export let name: string = '';
  export let id: string = uid('rating');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: { value: number };
  }>();

  let hoverValue: number | null = null;

  // Size configurations
  const sizeConfig = {
    sm: { icon: 'w-4 h-4', gap: 'gap-0.5' },
    md: { icon: 'w-6 h-6', gap: 'gap-1' },
    lg: { icon: 'w-8 h-8', gap: 'gap-1.5' },
  };

  $: config = sizeConfig[size as keyof typeof sizeConfig] ?? sizeConfig.md;
  $: displayValue = hoverValue ?? value;

  function handleClick(index: number, isHalf: boolean = false) {
    if (disabled || readonly) return;

    const newValue = isHalf ? index + 0.5 : index + 1;

    // Allow clearing if clicking the same value
    if (allowClear && newValue === value) {
      value = 0;
    } else {
      value = newValue;
    }

    dispatch('change', { value });
  }

  function handleMouseEnter(index: number, isHalf: boolean = false) {
    if (disabled || readonly) return;
    hoverValue = isHalf ? index + 0.5 : index + 1;
  }

  function handleMouseLeave() {
    hoverValue = null;
  }

  function handleKeydown(event: KeyboardEvent) {
    if (disabled || readonly) return;

    const step = allowHalf ? 0.5 : 1;

    if (event.key === 'ArrowRight' || event.key === 'ArrowUp') {
      event.preventDefault();
      value = Math.min(value + step, max);
      dispatch('change', { value });
    } else if (event.key === 'ArrowLeft' || event.key === 'ArrowDown') {
      event.preventDefault();
      value = Math.max(value - step, 0);
      dispatch('change', { value });
    }
  }

  function getIconFill(index: number): 'full' | 'half' | 'empty' {
    if (displayValue >= index + 1) return 'full';
    if (allowHalf && displayValue >= index + 0.5) return 'half';
    return 'empty';
  }
</script>

<div class={cn('inline-flex flex-col', className)}>
  {#if label}
    <label
      for={id}
      class="mb-2 font-medium text-[var(--color-text-primary)]"
    >
      {label}
    </label>
  {/if}

  <div
    class={cn(
      'inline-flex items-center',
      config.gap,
      (disabled || readonly) && 'cursor-default',
      !disabled && !readonly && 'cursor-pointer'
    )}
    role="slider"
    aria-valuemin="0"
    aria-valuemax={max}
    aria-valuenow={value}
    aria-label={label || 'Rating'}
    tabindex={disabled ? -1 : 0}
    on:keydown={handleKeydown}
    on:mouseleave={handleMouseLeave}
    data-testid={testId || undefined}
  >
    {#each Array(max) as _, i}
      {@const fill = getIconFill(i)}
      <div
        class="relative"
        on:click={() => handleClick(i)}
        on:mouseenter={() => handleMouseEnter(i)}
      >
        {#if allowHalf}
          <!-- Left half (for half rating) -->
          <div
            class="absolute inset-y-0 left-0 w-1/2 z-10"
            on:click|stopPropagation={() => handleClick(i, true)}
            on:mouseenter|stopPropagation={() => handleMouseEnter(i, true)}
          />
        {/if}

        {#if icon === 'star'}
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class={cn(config.icon, 'transition-colors')}
            viewBox="0 0 24 24"
            fill={fill === 'empty' ? 'none' : activeColor}
            stroke={fill === 'empty' ? inactiveColor : activeColor}
            stroke-width="2"
          >
            {#if fill === 'half'}
              <defs>
                <linearGradient id="half-{i}">
                  <stop offset="50%" stop-color={activeColor} />
                  <stop offset="50%" stop-color="transparent" />
                </linearGradient>
              </defs>
              <path
                fill="url(#half-{i})"
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
              />
            {:else}
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
              />
            {/if}
          </svg>
        {:else if icon === 'heart'}
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class={cn(config.icon, 'transition-colors')}
            viewBox="0 0 24 24"
            fill={fill === 'empty' ? 'none' : activeColor}
            stroke={fill === 'empty' ? inactiveColor : activeColor}
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"
            />
          </svg>
        {:else}
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class={cn(config.icon, 'transition-colors')}
            viewBox="0 0 24 24"
            fill={fill === 'empty' ? 'none' : activeColor}
            stroke={fill === 'empty' ? inactiveColor : activeColor}
            stroke-width="2"
          >
            <circle cx="12" cy="12" r="10" />
          </svg>
        {/if}
      </div>
    {/each}

    {#if showValue}
      <span class="ml-2 text-[var(--color-text-secondary)] tabular-nums">
        {value.toFixed(allowHalf ? 1 : 0)} / {max}
      </span>
    {/if}
  </div>

  {#if name}
    <input type="hidden" {name} value={value} />
  {/if}
</div>
