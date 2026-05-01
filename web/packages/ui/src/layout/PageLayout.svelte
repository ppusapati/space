<script lang="ts">
  import type { PageLayoutProps } from './layout.types';
  import { pageLayoutClasses, pageLayoutPaddingClasses, containerMaxWidthClasses } from './layout.types';
  import { cn } from '../utils';

  type $$Props = PageLayoutProps;

  export let layout: $$Props['layout'] = 'default';
  export let maxWidth: $$Props['maxWidth'] = 'xl';
  export let padding: $$Props['padding'] = 'lg';
  let className: $$Props['class'] = undefined;
  export { className as class };

  $: layoutClass = pageLayoutClasses[layout || 'default'];
  $: paddingClass = pageLayoutPaddingClasses[padding || 'lg'];
  $: maxWidthClass = maxWidth !== 'full' ? containerMaxWidthClasses[maxWidth || 'xl'] : '';
</script>

<div
  class={cn(
    pageLayoutClasses.base,
    layoutClass,
    paddingClass,
    maxWidthClass,
    maxWidth !== 'full' && 'mx-auto',
    className
  )}
>
  {#if $$slots.header}
    <div class="mb-6">
      <slot name="header" />
    </div>
  {/if}

  {#if layout === 'sidebar'}
    <div class="flex gap-6">
      {#if $$slots.sidebar}
        <aside class="w-64 shrink-0">
          <slot name="sidebar" />
        </aside>
      {/if}
      <div class="flex-1 min-w-0">
        <slot />
      </div>
    </div>
  {:else if layout === 'split'}
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <slot />
    </div>
  {:else if layout === 'centered'}
    <div class="max-w-2xl mx-auto">
      <slot />
    </div>
  {:else}
    <slot />
  {/if}

  {#if $$slots.footer}
    <div class="mt-6">
      <slot name="footer" />
    </div>
  {/if}
</div>
