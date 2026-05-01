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
  <polyline points="21 8 21 21 3 21 3 8"/>
  <rect x="1" y="3" width="22" height="5"/>
  <line x1="10" y1="12" x2="14" y2="12"/>
</IconBase>
