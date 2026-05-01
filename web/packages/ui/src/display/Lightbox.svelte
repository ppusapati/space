<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { fade, scale } from 'svelte/transition';
  import type { LightboxProps, CarouselItem } from './carousel.types';
  import { lightboxClasses } from './carousel.types';
  import { cn } from '../utils';
  import { portal } from '../actions/portal';

  type $$Props = LightboxProps;

  export let open: $$Props['open'] = false;
  export let items: $$Props['items'] = [];
  export let startIndex: $$Props['startIndex'] = 0;
  export let showCaptions: $$Props['showCaptions'] = true;
  export let showThumbnails: $$Props['showThumbnails'] = true;
  export let showCounter: $$Props['showCounter'] = true;
  export let closeOnBackdrop: $$Props['closeOnBackdrop'] = true;
  export let closeOnEscape: $$Props['closeOnEscape'] = true;
  export let zoomable: $$Props['zoomable'] = false;
  export let downloadable: $$Props['downloadable'] = false;
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{
    close: void;
    change: { index: number; item: CarouselItem };
  }>();

  let currentIndex: number = startIndex ?? 0;
  let isZoomed = false;

  $: currentIndex = startIndex ?? 0;
  $: currentItem = items[currentIndex];
  $: canGoPrev = currentIndex > 0;
  $: canGoNext = currentIndex < items.length - 1;

  function close() {
    dispatch('close');
  }

  function goTo(index: number) {
    if (index >= 0 && index < items.length) {
      currentIndex = index;
      isZoomed = false;
      dispatch('change', { index: currentIndex, item: items[currentIndex]! });
    }
  }

  function goToPrev() {
    if (canGoPrev) {
      goTo(currentIndex - 1);
    }
  }

  function goToNext() {
    if (canGoNext) {
      goTo(currentIndex + 1);
    }
  }

  function toggleZoom() {
    if (zoomable) {
      isZoomed = !isZoomed;
    }
  }

  function downloadImage() {
    if (downloadable && currentItem) {
      const link = document.createElement('a');
      link.href = currentItem.src;
      link.download = currentItem.alt || `image-${currentIndex + 1}`;
      link.target = '_blank';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (!open) return;

    switch (e.key) {
      case 'Escape':
        if (closeOnEscape) {
          e.preventDefault();
          close();
        }
        break;
      case 'ArrowLeft':
        e.preventDefault();
        goToPrev();
        break;
      case 'ArrowRight':
        e.preventDefault();
        goToNext();
        break;
    }
  }

  function handleBackdropClick(e: MouseEvent) {
    if (closeOnBackdrop && e.target === e.currentTarget) {
      close();
    }
  }

  onMount(() => {
    if (open) {
      document.body.style.overflow = 'hidden';
    }
  });

  onDestroy(() => {
    document.body.style.overflow = '';
  });

  $: if (typeof document !== 'undefined') {
    if (open) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if open}
  <div use:portal>
    <div
      class={cn(lightboxClasses.overlay, className)}
      transition:fade={{ duration: 200 }}
      aria-hidden="true"
    />

    <div
      class={lightboxClasses.container}
      role="dialog"
      aria-modal="true"
      aria-label="Image lightbox"
      on:click={handleBackdropClick}
    >
      <!-- Header -->
      <header class={lightboxClasses.header}>
        {#if showCounter}
          <span class={lightboxClasses.counter}>
            {currentIndex + 1} / {items.length}
          </span>
        {:else}
          <span />
        {/if}

        <div class={lightboxClasses.actions}>
          {#if zoomable}
            <button
              type="button"
              class={lightboxClasses.actionButton}
              on:click={toggleZoom}
              aria-label={isZoomed ? 'Zoom out' : 'Zoom in'}
            >
              {#if isZoomed}
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM13 10H7" />
                </svg>
              {:else}
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM10 7v6m3-3H7" />
                </svg>
              {/if}
            </button>
          {/if}

          {#if downloadable}
            <button
              type="button"
              class={lightboxClasses.actionButton}
              on:click={downloadImage}
              aria-label="Download image"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
            </button>
          {/if}

          <button
            type="button"
            class={lightboxClasses.closeButton}
            on:click={close}
            aria-label="Close lightbox"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </header>

      <!-- Main content -->
      <main class={lightboxClasses.main} on:click={handleBackdropClick}>
        {#if items.length > 1}
          <button
            type="button"
            class={cn(lightboxClasses.nav.button, lightboxClasses.nav.prev)}
            on:click|stopPropagation={goToPrev}
            disabled={!canGoPrev}
            aria-label="Previous image"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <button
            type="button"
            class={cn(lightboxClasses.nav.button, lightboxClasses.nav.next)}
            on:click|stopPropagation={goToNext}
            disabled={!canGoNext}
            aria-label="Next image"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </button>
        {/if}

        {#if currentItem}
          {#key currentItem.id}
            <img
              src={currentItem.src}
              alt={currentItem.alt}
              class={cn(
                lightboxClasses.image,
                isZoomed && 'cursor-zoom-out scale-150'
              )}
              style={isZoomed ? 'transform: scale(1.5);' : ''}
              on:click|stopPropagation={toggleZoom}
              transition:fade={{ duration: 150 }}
              draggable="false"
            />
          {/key}
        {/if}
      </main>

      <!-- Caption -->
      {#if showCaptions && currentItem?.caption}
        <div class={lightboxClasses.caption}>
          {currentItem.caption}
        </div>
      {/if}

      <!-- Thumbnails -->
      {#if showThumbnails && items.length > 1}
        <div class={lightboxClasses.thumbnails.container}>
          {#each items as item, index (item.id)}
            <button
              type="button"
              class={cn(
                lightboxClasses.thumbnails.item,
                index === currentIndex && lightboxClasses.thumbnails.itemActive
              )}
              on:click={() => goTo(index)}
              aria-label="Go to image {index + 1}"
              aria-current={index === currentIndex ? 'true' : undefined}
            >
              <img
                src={item.thumbnail || item.src}
                alt=""
                class={lightboxClasses.thumbnails.image}
                loading="lazy"
              />
            </button>
          {/each}
        </div>
      {/if}
    </div>
  </div>
{/if}
