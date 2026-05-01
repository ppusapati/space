<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import {
    virtualListClasses,
    calculateVisibleRange,
    getTotalHeight,
    getItemPosition,
  } from './virtuallist.types';

  type T = $$Generic;

  // Props
  export let items: T[] = [];
  export let itemHeight: number;
  export let height: string = '400px';
  export let overscan: number = 3;
  export let getKey: (item: T, index: number) => string | number = (_, index) => index;
  export let id: string = uid('virtuallist');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    scroll: { scrollTop: number; startIndex: number; endIndex: number };
  }>();

  // Internal state
  let containerRef: HTMLDivElement;
  let scrollTop = 0;
  let containerHeight = 0;

  // Calculate visible range
  $: visibleRange = calculateVisibleRange(
    scrollTop,
    containerHeight,
    itemHeight,
    items.length,
    overscan
  );

  // Get visible items
  $: visibleItems = items
    .slice(visibleRange.startIndex, visibleRange.endIndex + 1)
    .map((item, i) => ({
      item,
      index: visibleRange.startIndex + i,
      key: getKey(item, visibleRange.startIndex + i),
    }));

  // Total content height
  $: totalHeight = getTotalHeight(items.length, itemHeight);

  // Handle scroll
  function handleScroll() {
    scrollTop = containerRef.scrollTop;
    dispatch('scroll', {
      scrollTop,
      startIndex: visibleRange.startIndex,
      endIndex: visibleRange.endIndex,
    });
  }

  // Measure container height
  function measureContainer() {
    if (containerRef) {
      containerHeight = containerRef.clientHeight;
    }
  }

  onMount(() => {
    measureContainer();

    // Watch for container resize
    const resizeObserver = new ResizeObserver(() => {
      measureContainer();
    });

    resizeObserver.observe(containerRef);

    return () => {
      resizeObserver.disconnect();
    };
  });

  // Scroll to specific index
  export function scrollToIndex(index: number, align: 'start' | 'center' | 'end' = 'start') {
    if (!containerRef) return;

    let targetTop = getItemPosition(index, itemHeight);

    if (align === 'center') {
      targetTop -= (containerHeight - itemHeight) / 2;
    } else if (align === 'end') {
      targetTop -= containerHeight - itemHeight;
    }

    containerRef.scrollTop = Math.max(0, targetTop);
  }

  // Scroll to top
  export function scrollToTop() {
    if (containerRef) {
      containerRef.scrollTop = 0;
    }
  }

  // Scroll to bottom
  export function scrollToBottom() {
    if (containerRef) {
      containerRef.scrollTop = totalHeight;
    }
  }
</script>

<div
  bind:this={containerRef}
  class={cn(virtualListClasses.container, className)}
  style="height: {height};"
  on:scroll={handleScroll}
  {id}
  data-testid={testId || undefined}
  role="list"
>
  <div
    class={virtualListClasses.content}
    style="height: {totalHeight}px;"
  >
    {#each visibleItems as { item, index, key } (key)}
      <div
        class={virtualListClasses.item}
        style="height: {itemHeight}px; transform: translateY({getItemPosition(index, itemHeight)}px);"
        role="listitem"
      >
        <slot {item} {index}>
          <!-- Default: just render the item as string -->
          {String(item)}
        </slot>
      </div>
    {/each}
  </div>
</div>
