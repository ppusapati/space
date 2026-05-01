<script lang="ts">
  import type { StackProps } from './layout.types';
  import { stackClasses, flexAlignClasses, flexJustifyClasses, gapClasses } from './layout.types';
  import { cn } from '../utils';

  type $$Props = StackProps;

  export let direction: $$Props['direction'] = 'vertical';
  export let gap: $$Props['gap'] = 'md';
  export let align: $$Props['align'] = 'stretch';
  export let justify: $$Props['justify'] = 'start';
  export let wrap: $$Props['wrap'] = false;
  let className: $$Props['class'] = undefined;
  export { className as class };

  $: directionClass = stackClasses[direction || 'vertical'];
  $: alignClass = flexAlignClasses[align || 'stretch'];
  $: justifyClass = flexJustifyClasses[justify || 'start'];
  $: gapClass = gapClasses[gap || 'md'];
</script>

<div
  class={cn(
    stackClasses.base,
    directionClass,
    alignClass,
    justifyClass,
    gapClass,
    wrap && 'flex-wrap',
    className
  )}
>
  <slot />
</div>
