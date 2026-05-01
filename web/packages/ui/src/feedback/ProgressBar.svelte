<script lang="ts">
  import { cn } from '../utils/classnames';
  import type { Size } from '../types';

  // Props
  export let value: number = 0;
  export let max: number = 100;
  export let size: Size = 'md';
  export let variant: 'default' | 'success' | 'warning' | 'error' | 'info' = 'default';
  export let showLabel: boolean = false;
  export let showValue: boolean = false;
  export let label: string = '';
  export let indeterminate: boolean = false;
  export let striped: boolean = false;
  export let animated: boolean = false;

  let className: string = '';
  export { className as class };

  // Size configurations
  const sizeConfig = {
    sm: 'h-1',
    md: 'h-2',
    lg: 'h-3',
  };

  const variantConfig = {
    default: 'bg-[var(--color-interactive-primary)]',
    success: 'bg-[var(--color-success)]',
    warning: 'bg-[var(--color-warning)]',
    error: 'bg-[var(--color-error)]',
    info: 'bg-[var(--color-info)]',
  };

  $: percentage = Math.min(Math.max((value / max) * 100, 0), 100);
</script>

<div class={cn('w-full', className)}>
  {#if showLabel || showValue}
    <div class="flex justify-between items-center mb-1">
      {#if showLabel && label}
        <span class="text-sm font-medium text-[var(--color-text-primary)]">
          {label}
        </span>
      {/if}
      {#if showValue}
        <span class="text-sm text-[var(--color-text-secondary)]">
          {value} / {max}
        </span>
      {/if}
    </div>
  {/if}

  <div
    class={cn(
      'w-full rounded-full overflow-hidden',
      'bg-[var(--color-surface-tertiary)]',
      sizeConfig[size as keyof typeof sizeConfig] ?? sizeConfig.md
    )}
    role="progressbar"
    aria-valuenow={indeterminate ? undefined : value}
    aria-valuemin={0}
    aria-valuemax={max}
    aria-label={label || 'Progress'}
  >
    <div
      class={cn(
        'h-full rounded-full transition-all duration-300',
        variantConfig[variant],
        striped && 'progress-striped',
        animated && 'progress-animated',
        indeterminate && 'progress-indeterminate'
      )}
      style={indeterminate ? '' : `width: ${percentage}%`}
    />
  </div>
</div>

<style>
  .progress-striped {
    background-image: linear-gradient(
      45deg,
      rgba(255, 255, 255, 0.15) 25%,
      transparent 25%,
      transparent 50%,
      rgba(255, 255, 255, 0.15) 50%,
      rgba(255, 255, 255, 0.15) 75%,
      transparent 75%,
      transparent
    );
    background-size: 1rem 1rem;
  }

  .progress-animated {
    animation: progress-bar-stripes 1s linear infinite;
  }

  .progress-indeterminate {
    width: 30%;
    animation: progress-indeterminate 1.5s ease-in-out infinite;
  }

  @keyframes progress-bar-stripes {
    0% {
      background-position: 1rem 0;
    }
    100% {
      background-position: 0 0;
    }
  }

  @keyframes progress-indeterminate {
    0% {
      transform: translateX(-100%);
    }
    50% {
      transform: translateX(233%);
    }
    100% {
      transform: translateX(-100%);
    }
  }
</style>
