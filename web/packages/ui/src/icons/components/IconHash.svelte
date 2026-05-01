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
  <line x1="4" y1="9" x2="20" y2="9"/>
  <line x1="4" y1="15" x2="20" y2="15"/>
  <line x1="10" y1="3" x2="8" y2="21"/>
  <line x1="16" y1="3" x2="14" y2="21"/>
</IconBase>
