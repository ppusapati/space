<script lang="ts">
  import type { BadgeProps } from './display.types';
  import { badgeClasses, badgeSizeClasses, badgeVariantClasses, badgeDotVariantClasses } from './display.types';
  import { cn } from '../utils';

  type $$Props = BadgeProps;

  export let variant: $$Props['variant'] = 'primary';
  export let size: $$Props['size'] = 'md';
  export let pill: $$Props['pill'] = false;
  export let dot: $$Props['dot'] = false;
  export let content: $$Props['content'] = undefined;
  export let max: $$Props['max'] = 99;
  export let showZero: $$Props['showZero'] = false;
  let className: $$Props['class'] = undefined;
  export { className as class };

  $: sizeClass = badgeSizeClasses[size || 'md'];
  $: variantClass = dot
    ? badgeDotVariantClasses[variant || 'primary']
    : badgeVariantClasses[variant || 'primary'];

  $: displayContent = (() => {
    if (dot) return '';
    if (content === undefined || content === null) return '';
    if (typeof content === 'number') {
      if (content === 0 && !showZero) return '';
      if (max && content > max) return `${max}+`;
      return String(content);
    }
    return content;
  })();

  $: shouldShow = dot || displayContent !== '' || $$slots.default;
</script>

{#if shouldShow}
  <span
    class={cn(
      badgeClasses.base,
      dot ? badgeClasses.dot : sizeClass,
      variantClass,
      pill || dot ? 'rounded-full' : 'rounded',
      className
    )}
    role={content !== undefined ? 'status' : undefined}
  >
    {#if dot}
      <!-- Dot only, no content -->
    {:else if displayContent}
      {displayContent}
    {:else}
      <slot />
    {/if}
  </span>
{/if}
