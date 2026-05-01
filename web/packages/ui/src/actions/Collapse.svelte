<script lang="ts">
  import { onMount } from 'svelte';
  import type { CollapseProps } from './actions.types';
  import { collapseClasses } from './actions.types';
  import { cn } from '../utils';

  type $$Props = CollapseProps;

  export let open: $$Props['open'] = false;
  export let duration: $$Props['duration'] = 200;
  export let id: $$Props['id'] = undefined;
  export let testId: $$Props['testId'] = undefined;
  let className: $$Props['class'] = undefined;
  export { className as class };

  let contentRef: HTMLDivElement;
  let contentHeight = 0;
  let mounted = false;

  onMount(() => {
    mounted = true;
    updateHeight();
  });

  function updateHeight() {
    if (contentRef) {
      contentHeight = contentRef.scrollHeight;
    }
  }

  // Update height when content changes
  $: if (mounted && open) {
    // Use requestAnimationFrame to ensure DOM is updated
    requestAnimationFrame(updateHeight);
  }

  $: containerStyle = open
    ? `max-height: ${contentHeight}px; transition-duration: ${duration}ms;`
    : `max-height: 0; transition-duration: ${duration}ms;`;
</script>

<div
  {id}
  class={cn(collapseClasses.container, className)}
  style={containerStyle}
  aria-hidden={!open}
  data-testid={testId}
>
  <div bind:this={contentRef} class={collapseClasses.content}>
    <slot />
  </div>
</div>
