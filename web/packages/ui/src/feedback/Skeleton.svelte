<script lang="ts">
  import type { SkeletonProps } from './feedback.types';
  import { skeletonClasses } from './feedback.types';
  import { cn } from '../utils';

  type $$Props = SkeletonProps;

  export let variant: $$Props['variant'] = 'text';
  export let width: $$Props['width'] = undefined;
  export let height: $$Props['height'] = undefined;
  export let animation: $$Props['animation'] = 'pulse';
  export let lines: $$Props['lines'] = 1;
  let className: $$Props['class'] = undefined;
  export { className as class };

  $: variantClass = skeletonClasses[variant || 'text'];
  $: animationClass = animation === 'none' ? '' : animation === 'wave' ? skeletonClasses.wave : skeletonClasses.pulse;

  $: style = [
    width ? `width: ${width}` : '',
    height ? `height: ${height}` : '',
  ].filter(Boolean).join('; ');
</script>

{#if variant === 'text' && (lines || 1) > 1}
  <div class={cn('flex flex-col gap-2', className)} role="status" aria-busy="true">
    {#each Array(lines) as _, i}
      <div
        class={cn(skeletonClasses.base, variantClass, animationClass)}
        style={i === (lines || 1) - 1 ? `width: ${width || '75%'}` : `width: ${width || '100%'}`}
        aria-hidden="true"
      />
    {/each}
    <span class="sr-only">Loading...</span>
  </div>
{:else}
  <div
    class={cn(skeletonClasses.base, variantClass, animationClass, className)}
    style={style || undefined}
    role="status"
    aria-busy="true"
    aria-hidden="true"
  >
    <span class="sr-only">Loading...</span>
  </div>
{/if}
