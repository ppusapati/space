<script lang="ts">
  import IconBase from '../IconBase.svelte';

  interface Props {
    size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl' | number;
    strokeWidth?: number;
    class?: string;
    label?: string;
  }

  const { size, strokeWidth, class: className, label }: Props = $props();
</script>

<IconBase {size} {strokeWidth} class={className} {label}>
  <rect x="4" y="2" width="16" height="20" rx="2" ry="2"/>
  <path d="M9 22v-4h6v4"/>
  <line x1="8" y1="6" x2="8" y2="6"/>
  <line x1="16" y1="6" x2="16" y2="6"/>
  <line x1="8" y1="10" x2="8" y2="10"/>
  <line x1="16" y1="10" x2="16" y2="10"/>
  <line x1="8" y1="14" x2="8" y2="14"/>
  <line x1="16" y1="14" x2="16" y2="14"/>
</IconBase>
