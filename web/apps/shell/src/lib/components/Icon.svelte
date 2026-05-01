<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    name?: string;
    size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
    class?: string;
    /** For inline SVG icons passed as children */
    children?: Snippet;
  }

  const { name, size = 'md', class: className = '', children }: Props = $props();

  const sizeMap = {
    xs: 'w-3 h-3',
    sm: 'w-4 h-4',
    md: 'w-5 h-5',
    lg: 'w-6 h-6',
    xl: 'w-8 h-8',
  };

  const sizeClass = sizeMap[size];
</script>

{#if children}
  <svg
    class="{sizeClass} {className} flex-shrink-0"
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
    stroke-width="2"
    aria-hidden="true"
  >
    {@render children()}
  </svg>
{:else if name}
  <svg
    class="{sizeClass} {className} flex-shrink-0"
    aria-hidden="true"
  >
    <use href="/icons.svg#{name}" />
  </svg>
{/if}
