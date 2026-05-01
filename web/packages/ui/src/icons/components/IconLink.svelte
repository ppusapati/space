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
  <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
  <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
</IconBase>
