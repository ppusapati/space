<script lang="ts">
  import type { ListProps } from './display.types';
  import { listClasses, listSizeClasses } from './display.types';
  import { cn } from '../utils';

  type $$Props = ListProps;

  export let variant: $$Props['variant'] = 'simple';
  export let size: $$Props['size'] = 'md';
  export let ordered: $$Props['ordered'] = false;
  let className: $$Props['class'] = undefined;
  export { className as class };

  $: variantClass = listClasses[variant || 'simple'];
  $: sizeClass = listSizeClasses[size || 'md'];
</script>

{#if ordered}
  <ol
    class={cn(listClasses.base, variantClass, sizeClass, className)}
    role="list"
  >
    <slot />
  </ol>
{:else}
  <ul
    class={cn(listClasses.base, variantClass, sizeClass, className)}
    role="list"
  >
    <slot />
  </ul>
{/if}
