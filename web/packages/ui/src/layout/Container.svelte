<script lang="ts">
  import type { ContainerProps } from './layout.types';
  import { containerClasses, containerMaxWidthClasses, containerPaddingClasses } from './layout.types';
  import { cn } from '../utils';

  type $$Props = ContainerProps;

  export let maxWidth: $$Props['maxWidth'] = 'lg';
  export let centered: $$Props['centered'] = true;
  export let padding: $$Props['padding'] = 'md';
  let className: $$Props['class'] = undefined;
  export { className as class };

  $: maxWidthClass = containerMaxWidthClasses[maxWidth || 'lg'];
  $: paddingClass = containerPaddingClasses[padding || 'md'];
</script>

<div
  class={cn(
    containerClasses.base,
    maxWidthClass,
    centered && containerClasses.centered,
    paddingClass,
    className
  )}
>
  <slot />
</div>
