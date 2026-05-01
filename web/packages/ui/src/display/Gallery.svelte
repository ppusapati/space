<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { GalleryProps, CarouselItem } from './carousel.types';
  import { galleryClasses, galleryGapClasses, galleryRoundedClasses } from './carousel.types';
  import { cn } from '../utils';
  import Lightbox from './Lightbox.svelte';

  type $$Props = GalleryProps;

  export let items: $$Props['items'] = [];
  export let columns: $$Props['columns'] = 3;
  export let gap: $$Props['gap'] = 'md';
  export let layout: $$Props['layout'] = 'grid';
  export let aspectRatio: $$Props['aspectRatio'] = '1/1';
  export let lightbox: $$Props['lightbox'] = true;
  export let fit: $$Props['fit'] = 'cover';
  export let rounded: $$Props['rounded'] = 'md';
  export let hoverable: $$Props['hoverable'] = true;
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{
    itemClick: { index: number; item: CarouselItem };
  }>();

  let lightboxOpen = false;
  let lightboxIndex = 0;

  $: gridColumns = typeof columns === 'number'
    ? `repeat(${columns}, minmax(0, 1fr))`
    : undefined;

  $: responsiveClasses = typeof columns === 'object'
    ? [
        columns.xs && `grid-cols-${columns.xs}`,
        columns.sm && `sm:grid-cols-${columns.sm}`,
        columns.md && `md:grid-cols-${columns.md}`,
        columns.lg && `lg:grid-cols-${columns.lg}`,
        columns.xl && `xl:grid-cols-${columns.xl}`,
      ].filter(Boolean).join(' ')
    : '';

  $: fitClass = fit === 'contain' ? 'object-contain' : fit === 'cover' ? 'object-cover' : 'object-fill';

  function handleItemClick(index: number) {
    dispatch('itemClick', { index, item: items[index]! });
    if (lightbox) {
      lightboxIndex = index;
      lightboxOpen = true;
    }
  }

  function handleLightboxClose() {
    lightboxOpen = false;
  }
</script>

<div
  class={cn(
    galleryClasses.container,
    galleryClasses.grid,
    galleryGapClasses[gap ?? 'md'],
    responsiveClasses,
    className
  )}
  style={gridColumns ? `grid-template-columns: ${gridColumns};` : undefined}
>
  {#each items as item, index (item.id)}
    <button
      type="button"
      class={cn(
        galleryClasses.item,
        galleryRoundedClasses[rounded ?? 'md'],
        hoverable && galleryClasses.itemHoverable,
        'group'
      )}
      style="aspect-ratio: {aspectRatio};"
      on:click={() => handleItemClick(index)}
      aria-label={item.alt || `Gallery image ${index + 1}`}
    >
      <img
        src={item.thumbnail || item.src}
        alt={item.alt}
        class={cn(
          galleryClasses.image,
          fitClass,
          hoverable && galleryClasses.imageHover
        )}
        loading="lazy"
        decoding="async"
      />
    </button>
  {/each}
</div>

{#if lightbox}
  <Lightbox
    open={lightboxOpen}
    {items}
    startIndex={lightboxIndex}
    on:close={handleLightboxClose}
  />
{/if}
