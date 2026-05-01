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
  <line x1="19" y1="5" x2="5" y2="19"/>
  <circle cx="6.5" cy="6.5" r="2.5"/>
  <circle cx="17.5" cy="17.5" r="2.5"/>
</IconBase>
