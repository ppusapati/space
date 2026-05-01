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
  <polyline points="16 18 22 12 16 6"/>
  <polyline points="8 6 2 12 8 18"/>
</IconBase>
