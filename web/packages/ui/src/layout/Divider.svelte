<script lang="ts">
  import type { DividerProps } from './layout.types';
  import { dividerClasses } from './layout.types';
  import { cn } from '../utils';

  type $$Props = DividerProps;

  export let orientation: $$Props['orientation'] = 'horizontal';
  export let variant: $$Props['variant'] = 'solid';
  export let color: $$Props['color'] = 'medium';
  export let label: $$Props['label'] = undefined;
  export let labelPosition: $$Props['labelPosition'] = 'center';
  let className: $$Props['class'] = undefined;
  export { className as class };

  $: orientationClass = dividerClasses[orientation || 'horizontal'];
  $: colorClass = dividerClasses[color || 'medium'];
</script>

{#if label && orientation === 'horizontal'}
  <div
    class={cn(dividerClasses.withLabel, className)}
    role="separator"
  >
    {#if labelPosition === 'start'}
      <span class={dividerClasses.label}>{label}</span>
      <div class={cn(dividerClasses.line, colorClass)} />
    {:else if labelPosition === 'end'}
      <div class={cn(dividerClasses.line, colorClass)} />
      <span class={dividerClasses.label}>{label}</span>
    {:else}
      <div class={cn(dividerClasses.line, colorClass)} />
      <span class={dividerClasses.label}>{label}</span>
      <div class={cn(dividerClasses.line, colorClass)} />
    {/if}
  </div>
{:else}
  <hr
    class={cn(
      dividerClasses.base,
      orientationClass,
      colorClass,
      className
    )}
    role="separator"
    aria-orientation={orientation}
  />
{/if}
