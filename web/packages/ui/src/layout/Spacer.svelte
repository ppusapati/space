<script lang="ts">
  import type { SpacerProps } from './layout.types';
  import { spacerClasses } from './layout.types';
  import { cn } from '../utils';

  type $$Props = SpacerProps;

  export let size: $$Props['size'] = 'md';
  export let axis: $$Props['axis'] = 'both';
  let className: $$Props['class'] = undefined;
  export { className as class };

  $: sizeStyles = spacerClasses[size || 'md'];

  $: spacerClass = (() => {
    switch (axis) {
      case 'horizontal':
        return sizeStyles.x;
      case 'vertical':
        return sizeStyles.y;
      case 'both':
      default:
        return `${sizeStyles.x} ${sizeStyles.y}`;
    }
  })();
</script>

<div class={cn(spacerClass, className)} aria-hidden="true" />
