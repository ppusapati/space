<script context="module" lang="ts">
  export type StatusVariant = 'success' | 'warning' | 'error' | 'info' | 'neutral' | 'pending' | 'active';
</script>

<script lang="ts">
  import { cn } from '../utils/classnames';
  import type { Size } from '../types';

  // Props
  export let status: StatusVariant = 'neutral';
  export let label: string = '';
  export let size: Size = 'md';
  export let showDot: boolean = true;
  export let pulse: boolean = false;
  export let uppercase: boolean = false;

  let className: string = '';
  export { className as class };

  const sizeConfig = {
    sm: { badge: 'text-xs px-1.5 py-0.5', dot: 'w-1.5 h-1.5' },
    md: { badge: 'text-sm px-2 py-1', dot: 'w-2 h-2' },
    lg: { badge: 'text-base px-3 py-1.5', dot: 'w-2.5 h-2.5' },
  };

  const variantConfig = {
    success: {
      bg: 'bg-[var(--color-success)]/15',
      text: 'text-[var(--color-success)]',
      dot: 'bg-[var(--color-success)]',
    },
    warning: {
      bg: 'bg-[var(--color-warning)]/15',
      text: 'text-[var(--color-warning)]',
      dot: 'bg-[var(--color-warning)]',
    },
    error: {
      bg: 'bg-[var(--color-error)]/15',
      text: 'text-[var(--color-error)]',
      dot: 'bg-[var(--color-error)]',
    },
    info: {
      bg: 'bg-[var(--color-info)]/15',
      text: 'text-[var(--color-info)]',
      dot: 'bg-[var(--color-info)]',
    },
    neutral: {
      bg: 'bg-[var(--color-neutral-200)]',
      text: 'text-[var(--color-neutral-700)]',
      dot: 'bg-[var(--color-neutral-500)]',
    },
    pending: {
      bg: 'bg-[var(--color-warning)]/15',
      text: 'text-[var(--color-warning)]',
      dot: 'bg-[var(--color-warning)]',
    },
    active: {
      bg: 'bg-[var(--color-interactive-primary)]/15',
      text: 'text-[var(--color-interactive-primary)]',
      dot: 'bg-[var(--color-interactive-primary)]',
    },
  };

  $: config = sizeConfig[size as keyof typeof sizeConfig] ?? sizeConfig.md;
  $: variant = variantConfig[status];

  const defaultLabels: Record<StatusVariant, string> = {
    success: 'Success',
    warning: 'Warning',
    error: 'Error',
    info: 'Info',
    neutral: 'Neutral',
    pending: 'Pending',
    active: 'Active',
  };

  $: displayLabel = label || defaultLabels[status];
</script>

<span
  class={cn(
    'inline-flex items-center gap-1.5 rounded-full font-medium',
    config.badge,
    variant.bg,
    variant.text,
    uppercase && 'uppercase tracking-wide',
    className
  )}
>
  {#if showDot}
    <span
      class={cn(
        'rounded-full',
        config.dot,
        variant.dot,
        pulse && 'animate-pulse'
      )}
    />
  {/if}
  {displayLabel}
</span>
