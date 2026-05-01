<script lang="ts">
  import type { ButtonGroupProps } from './actions.types';
  import { buttonGroupClasses } from './actions.types';
  import { cn } from '../utils';

  type $$Props = ButtonGroupProps;

  export let size: $$Props['size'] = 'md';
  export let variant: $$Props['variant'] = undefined;
  export let vertical: $$Props['vertical'] = false;
  export let attached: $$Props['attached'] = false;
  export let id: $$Props['id'] = undefined;
  export let testId: $$Props['testId'] = undefined;
  let className: $$Props['class'] = undefined;
  export { className as class };

  $: groupClasses = cn(
    buttonGroupClasses.base,
    vertical ? buttonGroupClasses.vertical : buttonGroupClasses.horizontal,
    attached
      ? [
          buttonGroupClasses.attached,
          vertical ? buttonGroupClasses.attachedVertical : buttonGroupClasses.attachedHorizontal,
        ]
      : buttonGroupClasses.gap,
    className
  );
</script>

<div
  {id}
  class={groupClasses}
  role="group"
  data-testid={testId}
  data-size={size}
  data-variant={variant}
>
  <slot />
</div>

<style>
  /* Pass down size and variant to child buttons via CSS custom properties */
  div[data-size] :global(button),
  div[data-size] :global(a) {
    /* Buttons will inherit from parent context if needed */
  }
</style>
