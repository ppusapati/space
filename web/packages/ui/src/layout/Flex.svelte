<script lang="ts">
  import type { FlexProps } from './layout.types';
  import {
    flexClasses,
    flexDirectionClasses,
    flexJustifyClasses,
    flexAlignClasses,
    flexWrapClasses,
    gapClasses,
  } from './layout.types';
  import { cn } from '../utils';

  type $$Props = FlexProps;

  export let direction: $$Props['direction'] = 'row';
  export let justify: $$Props['justify'] = 'start';
  export let align: $$Props['align'] = 'stretch';
  export let wrap: $$Props['wrap'] = 'nowrap';
  export let gap: $$Props['gap'] = 'none';
  export let inline: $$Props['inline'] = false;
  let className: $$Props['class'] = undefined;
  export { className as class };

  $: directionClass = flexDirectionClasses[direction || 'row'];
  $: justifyClass = flexJustifyClasses[justify || 'start'];
  $: alignClass = flexAlignClasses[align || 'stretch'];
  $: wrapClass = flexWrapClasses[wrap || 'nowrap'];
  $: gapClass = gapClasses[gap || 'none'];
</script>

<div
  class={cn(
    inline ? flexClasses.inline : flexClasses.base,
    directionClass,
    justifyClass,
    alignClass,
    wrapClass,
    gapClass,
    className
  )}
>
  <slot />
</div>
