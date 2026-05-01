<script lang="ts">
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  import type { ScrollspyProps, ScrollspyItem } from './actions.types';
  import { scrollspyClasses } from './actions.types';
  import { cn } from '../utils';

  type $$Props = ScrollspyProps;

  export let items: $$Props['items'];
  export let offset: $$Props['offset'] = 100;
  export let container: $$Props['container'] = undefined;
  export let id: $$Props['id'] = undefined;
  export let testId: $$Props['testId'] = undefined;
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: { activeId: string | null; item: ScrollspyItem | null };
  }>();

  let activeId: string | null = null;
  let scrollContainer: Element | Window | null = null;
  let rafId: number | null = null;

  function getScrollContainer(): Element | Window {
    if (container) {
      return document.querySelector(container) || window;
    }
    return window;
  }

  function getScrollTop(): number {
    if (scrollContainer instanceof Window) {
      return window.scrollY || document.documentElement.scrollTop;
    }
    return (scrollContainer as Element).scrollTop;
  }

  function updateActiveItem() {
    if (rafId) {
      cancelAnimationFrame(rafId);
    }

    rafId = requestAnimationFrame(() => {
      const scrollTop = getScrollTop();
      const effectiveOffset = offset || 100;

      let newActiveId: string | null = null;

      // Find the section that's currently in view
      for (let i = items.length - 1; i >= 0; i--) {
        const item = items[i]!;
        const targetId = item.href.replace('#', '');
        const element = document.getElementById(targetId);

        if (element) {
          const elementTop = element.offsetTop;

          if (scrollTop >= elementTop - effectiveOffset) {
            newActiveId = item.id;
            break;
          }
        }
      }

      // Default to first item if we're at the top
      if (!newActiveId && items.length > 0) {
        newActiveId = items[0]!.id;
      }

      if (newActiveId !== activeId) {
        activeId = newActiveId;
        const activeItem = items.find((item) => item.id === activeId) || null;
        dispatch('change', { activeId, item: activeItem });
      }
    });
  }

  function handleClick(event: MouseEvent, item: ScrollspyItem) {
    event.preventDefault();

    const targetId = item.href.replace('#', '');
    const element = document.getElementById(targetId);

    if (element) {
      const elementTop = element.offsetTop;
      const effectiveOffset = offset || 100;

      if (scrollContainer instanceof Window) {
        window.scrollTo({
          top: elementTop - effectiveOffset + 1,
          behavior: 'smooth',
        });
      } else {
        (scrollContainer as Element).scrollTo({
          top: elementTop - effectiveOffset + 1,
          behavior: 'smooth',
        });
      }
    }
  }

  function handleKeydown(event: KeyboardEvent, item: ScrollspyItem, index: number) {
    switch (event.key) {
      case 'Enter':
      case ' ':
        event.preventDefault();
        handleClick(event as unknown as MouseEvent, item);
        break;
      case 'ArrowDown':
        event.preventDefault();
        if (index < items.length - 1) {
          const nextElement = document.querySelector(
            `[data-scrollspy-index="${index + 1}"]`
          ) as HTMLElement;
          nextElement?.focus();
        }
        break;
      case 'ArrowUp':
        event.preventDefault();
        if (index > 0) {
          const prevElement = document.querySelector(
            `[data-scrollspy-index="${index - 1}"]`
          ) as HTMLElement;
          prevElement?.focus();
        }
        break;
      case 'Home':
        event.preventDefault();
        const firstElement = document.querySelector(
          `[data-scrollspy-index="0"]`
        ) as HTMLElement;
        firstElement?.focus();
        break;
      case 'End':
        event.preventDefault();
        const lastElement = document.querySelector(
          `[data-scrollspy-index="${items.length - 1}"]`
        ) as HTMLElement;
        lastElement?.focus();
        break;
    }
  }

  onMount(() => {
    scrollContainer = getScrollContainer();
    scrollContainer.addEventListener('scroll', updateActiveItem, { passive: true });

    // Initial check
    updateActiveItem();
  });

  onDestroy(() => {
    if (scrollContainer) {
      scrollContainer.removeEventListener('scroll', updateActiveItem);
    }
    if (rafId) {
      cancelAnimationFrame(rafId);
    }
  });
</script>

<nav
  {id}
  class={cn(scrollspyClasses.nav, className)}
  aria-label="Page navigation"
  data-testid={testId}
>
  {#each items as item, index (item.id)}
    <a
      href={item.href}
      class={cn(
        scrollspyClasses.item,
        activeId === item.id && scrollspyClasses.itemActive
      )}
      aria-current={activeId === item.id ? 'location' : undefined}
      data-scrollspy-index={index}
      on:click={(e) => handleClick(e, item)}
      on:keydown={(e) => handleKeydown(e, item, index)}
    >
      {item.label}
    </a>
  {/each}
</nav>
