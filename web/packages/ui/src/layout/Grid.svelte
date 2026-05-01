<script lang="ts">
  import type { GridProps } from './layout.types';
  import { gridClasses, gridColsClasses, gridResponsiveClasses, gapClasses } from './layout.types';
  import { cn } from '../utils';

  type $$Props = GridProps;

  export let cols: $$Props['cols'] = 3;
  export let gap: $$Props['gap'] = 'md';
  export let responsive: $$Props['responsive'] = true;
  let className: $$Props['class'] = undefined;
  export { className as class };

  $: colsClass = responsive
    ? gridResponsiveClasses[cols || 3]
    : gridColsClasses[cols || 3];
  $: gapClass = gapClasses[gap || 'md'];
</script>

<div class={cn(gridClasses.base, colsClass, gapClass, className)}>
  <slot />
</div>
