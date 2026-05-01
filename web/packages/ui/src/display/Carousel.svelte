<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import type { CarouselProps, CarouselItem } from './carousel.types';
  import { carouselClasses } from './carousel.types';
  import { cn } from '../utils';

  type $$Props = CarouselProps;

  export let items: $$Props['items'] = [];
  export let currentIndex: $$Props['currentIndex'] = 0;
  export let autoPlay: $$Props['autoPlay'] = false;
  export let interval: $$Props['interval'] = 5000;
  export let pauseOnHover: $$Props['pauseOnHover'] = true;
  export let showControls: $$Props['showControls'] = true;
  export let showIndicators: $$Props['showIndicators'] = true;
  export let indicatorPosition: $$Props['indicatorPosition'] = 'bottom';
  export let loop: $$Props['loop'] = true;
  export let transition: $$Props['transition'] = 'slide';
  export let transitionDuration: $$Props['transitionDuration'] = 300;
  export let showCaptions: $$Props['showCaptions'] = false;
  export let captionPosition: $$Props['captionPosition'] = 'bottom';
  export let aspectRatio: $$Props['aspectRatio'] = '16/9';
  export let fit: $$Props['fit'] = 'cover';
  export let keyboard: $$Props['keyboard'] = true;
  export let touch: $$Props['touch'] = true;
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: { index: number; item: CarouselItem };
    slideClick: { index: number; item: CarouselItem };
  }>();

  let autoPlayTimer: ReturnType<typeof setInterval> | null = null;
  let isPaused = false;
  let touchStartX = 0;
  let touchEndX = 0;
  let containerElement: HTMLDivElement;

  $: activeIndex = Math.max(0, Math.min(currentIndex ?? 0, items.length - 1));
  $: canGoPrev = loop || activeIndex > 0;
  $: canGoNext = loop || activeIndex < items.length - 1;

  function goTo(index: number) {
    if (index < 0) {
      activeIndex = loop ? items.length - 1 : 0;
    } else if (index >= items.length) {
      activeIndex = loop ? 0 : items.length - 1;
    } else {
      activeIndex = index;
    }
    dispatch('change', { index: activeIndex, item: items[activeIndex]! });
  }

  function goToPrev() {
    if (canGoPrev) {
      goTo(activeIndex - 1);
    }
  }

  function goToNext() {
    if (canGoNext) {
      goTo(activeIndex + 1);
    }
  }

  function handleSlideClick(index: number) {
    dispatch('slideClick', { index, item: items[index]! });
  }

  function startAutoPlay() {
    if (autoPlay && !autoPlayTimer && items.length > 1) {
      autoPlayTimer = setInterval(() => {
        if (!isPaused) {
          goToNext();
        }
      }, interval);
    }
  }

  function stopAutoPlay() {
    if (autoPlayTimer) {
      clearInterval(autoPlayTimer);
      autoPlayTimer = null;
    }
  }

  function handleMouseEnter() {
    if (pauseOnHover) {
      isPaused = true;
    }
  }

  function handleMouseLeave() {
    if (pauseOnHover) {
      isPaused = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (!keyboard) return;
    if (e.key === 'ArrowLeft') {
      e.preventDefault();
      goToPrev();
    } else if (e.key === 'ArrowRight') {
      e.preventDefault();
      goToNext();
    }
  }

  function handleTouchStart(e: TouchEvent) {
    if (!touch) return;
    touchStartX = e.touches[0]!.clientX;
  }

  function handleTouchMove(e: TouchEvent) {
    if (!touch) return;
    touchEndX = e.touches[0]!.clientX;
  }

  function handleTouchEnd() {
    if (!touch) return;
    const diff = touchStartX - touchEndX;
    const threshold = 50;

    if (Math.abs(diff) > threshold) {
      if (diff > 0) {
        goToNext();
      } else {
        goToPrev();
      }
    }
    touchStartX = 0;
    touchEndX = 0;
  }

  $: if (autoPlay) {
    stopAutoPlay();
    startAutoPlay();
  }

  onMount(() => {
    if (autoPlay) {
      startAutoPlay();
    }
  });

  onDestroy(() => {
    stopAutoPlay();
  });

  $: trackStyle = transition === 'slide'
    ? `transform: translateX(-${activeIndex * 100}%); transition-duration: ${transitionDuration}ms;`
    : '';

  $: fitClass = fit === 'contain' ? 'object-contain' : fit === 'cover' ? 'object-cover' : 'object-fill';
</script>

<div
  bind:this={containerElement}
  class={cn(carouselClasses.container, className)}
  style="aspect-ratio: {aspectRatio};"
  on:mouseenter={handleMouseEnter}
  on:mouseleave={handleMouseLeave}
  on:keydown={handleKeydown}
  on:touchstart={handleTouchStart}
  on:touchmove={handleTouchMove}
  on:touchend={handleTouchEnd}
  tabindex="0"
  role="region"
  aria-roledescription="carousel"
  aria-label="Image carousel"
>
  <div class={carouselClasses.viewport}>
    {#if transition === 'slide'}
      <div class={carouselClasses.track} style={trackStyle}>
        {#each items as item, index (item.id)}
          <div
            class={carouselClasses.slide}
            role="group"
            aria-roledescription="slide"
            aria-label="Slide {index + 1} of {items.length}"
            aria-hidden={index !== activeIndex}
          >
            <button
              type="button"
              class="w-full h-full"
              on:click={() => handleSlideClick(index)}
            >
              <img
                src={item.src}
                alt={item.alt}
                class={cn(carouselClasses.image, fitClass)}
                loading={index === 0 ? 'eager' : 'lazy'}
                draggable="false"
              />
            </button>
          </div>
        {/each}
      </div>
    {:else if transition === 'fade'}
      {#each items as item, index (item.id)}
        <div
          class={cn(carouselClasses.slide, 'absolute inset-0')}
          style="opacity: {index === activeIndex ? 1 : 0}; transition: opacity {transitionDuration}ms ease-in-out;"
          role="group"
          aria-roledescription="slide"
          aria-label="Slide {index + 1} of {items.length}"
          aria-hidden={index !== activeIndex}
        >
          <button
            type="button"
            class="w-full h-full"
            on:click={() => handleSlideClick(index)}
          >
            <img
              src={item.src}
              alt={item.alt}
              class={cn(carouselClasses.image, fitClass)}
              loading={index === 0 ? 'eager' : 'lazy'}
              draggable="false"
            />
          </button>
        </div>
      {/each}
    {:else}
      <div
        class={carouselClasses.slide}
        role="group"
        aria-roledescription="slide"
      >
        <button
          type="button"
          class="w-full h-full"
          on:click={() => handleSlideClick(activeIndex)}
        >
          <img
            src={items[activeIndex]?.src}
            alt={items[activeIndex]?.alt}
            class={cn(carouselClasses.image, fitClass)}
            draggable="false"
          />
        </button>
      </div>
    {/if}

    {#if showCaptions && items[activeIndex]?.caption}
      <div class={carouselClasses.caption[captionPosition!]}>
        <p class={carouselClasses.caption.text}>{items[activeIndex]!.caption}</p>
      </div>
    {/if}
  </div>

  {#if showControls && items.length > 1}
    <button
      type="button"
      class={cn(carouselClasses.controls.base, carouselClasses.controls.button, carouselClasses.controls.prev)}
      on:click={goToPrev}
      disabled={!canGoPrev}
      aria-label="Previous slide"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
      </svg>
    </button>
    <button
      type="button"
      class={cn(carouselClasses.controls.base, carouselClasses.controls.button, carouselClasses.controls.next)}
      on:click={goToNext}
      disabled={!canGoNext}
      aria-label="Next slide"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
      </svg>
    </button>
  {/if}

  {#if showIndicators && items.length > 1}
    <div class={cn(carouselClasses.indicators.container, carouselClasses.indicators[indicatorPosition!])}>
      {#each items as item, index (item.id)}
        <button
          type="button"
          class={cn(
            carouselClasses.indicators.dot,
            index === activeIndex && carouselClasses.indicators.dotActive
          )}
          on:click={() => goTo(index)}
          aria-label="Go to slide {index + 1}"
          aria-current={index === activeIndex ? 'true' : undefined}
        />
      {/each}
    </div>
  {/if}
</div>
