<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    /** Size preset or custom size in pixels */
    size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl' | number;
    /** Stroke width */
    strokeWidth?: number;
    /** Additional CSS classes */
    class?: string;
    /** Accessible label */
    label?: string;
    /** Icon path content */
    children: Snippet;
  }

  const {
    size = 'md',
    strokeWidth = 2,
    class: className = '',
    label,
    children,
  }: Props = $props();

  const sizeMap: Record<string, string> = {
    xs: '12',
    sm: '16',
    md: '20',
    lg: '24',
    xl: '32',
    '2xl': '40',
  };

  const computedSize = typeof size === 'number' ? size : sizeMap[size] || '20';
</script>

<svg
  xmlns="http://www.w3.org/2000/svg"
  width={computedSize}
  height={computedSize}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width={strokeWidth}
  stroke-linecap="round"
  stroke-linejoin="round"
  class="inline-block flex-shrink-0 {className}"
  aria-hidden={!label}
  aria-label={label}
  role={label ? 'img' : undefined}
>
  {@render children()}
</svg>
