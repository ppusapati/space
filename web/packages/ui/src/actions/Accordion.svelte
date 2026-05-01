<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { AccordionProps, AccordionItemData } from './actions.types';
  import {
    accordionClasses,
    accordionItemClasses,
    accordionSizeClasses,
  } from './actions.types';
  import { cn, uid } from '../utils';
  import Collapse from './Collapse.svelte';

  type $$Props = AccordionProps;

  export let items: $$Props['items'] = [];
  export let size: $$Props['size'] = 'md';
  export let multiple: $$Props['multiple'] = false;
  export let defaultOpen: $$Props['defaultOpen'] = [];
  export let flush: $$Props['flush'] = false;
  export let id: $$Props['id'] = undefined;
  export let testId: $$Props['testId'] = undefined;
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: { openItems: string[] };
    toggle: { itemId: string; isOpen: boolean };
  }>();

  const baseId = id || uid('accordion');

  // Track open items
  let openItems: Set<string> = new Set(defaultOpen || []);

  $: sizeConfig = accordionSizeClasses[size || 'md'];

  $: containerClasses = cn(
    flush ? accordionClasses.containerFlush : accordionClasses.container,
    className
  );

  function toggleItem(item: AccordionItemData) {
    if (item.disabled) return;

    const isOpen = openItems.has(item.id);

    if (isOpen) {
      openItems.delete(item.id);
    } else {
      if (!multiple) {
        openItems.clear();
      }
      openItems.add(item.id);
    }

    openItems = openItems; // Trigger reactivity

    dispatch('toggle', { itemId: item.id, isOpen: !isOpen });
    dispatch('change', { openItems: Array.from(openItems) });
  }

  function handleKeydown(event: KeyboardEvent, item: AccordionItemData, index: number) {
    if (item.disabled) return;

    const enabledItems = items?.filter((i) => !i.disabled) || [];
    const currentEnabledIndex = enabledItems.findIndex((i) => i.id === item.id);

    switch (event.key) {
      case 'Enter':
      case ' ':
        event.preventDefault();
        toggleItem(item);
        break;
      case 'ArrowDown':
        event.preventDefault();
        if (currentEnabledIndex < enabledItems.length - 1) {
          const nextItem = enabledItems[currentEnabledIndex + 1]!;
          document.getElementById(`${baseId}-header-${nextItem.id}`)?.focus();
        }
        break;
      case 'ArrowUp':
        event.preventDefault();
        if (currentEnabledIndex > 0) {
          const prevItem = enabledItems[currentEnabledIndex - 1]!;
          document.getElementById(`${baseId}-header-${prevItem.id}`)?.focus();
        }
        break;
      case 'Home':
        event.preventDefault();
        if (enabledItems.length > 0) {
          document.getElementById(`${baseId}-header-${enabledItems[0]!.id}`)?.focus();
        }
        break;
      case 'End':
        event.preventDefault();
        if (enabledItems.length > 0) {
          document
            .getElementById(`${baseId}-header-${enabledItems[enabledItems.length - 1]!.id}`)
            ?.focus();
        }
        break;
    }
  }

  function isOpen(itemId: string): boolean {
    return openItems.has(itemId);
  }
</script>

<div
  id={baseId}
  class={containerClasses}
  data-testid={testId}
>
  {#if items && items.length > 0}
    {#each items as item, index (item.id)}
      {@const itemOpen = isOpen(item.id)}
      <div class="accordion-item">
        <h3>
          <button
            type="button"
            id="{baseId}-header-{item.id}"
            class={cn(
              accordionItemClasses.header,
              sizeConfig.header,
              item.disabled
                ? accordionItemClasses.headerDisabled
                : itemOpen
                  ? accordionItemClasses.headerOpen
                  : accordionItemClasses.headerClosed
            )}
            aria-expanded={itemOpen}
            aria-controls="{baseId}-panel-{item.id}"
            aria-disabled={item.disabled}
            disabled={item.disabled}
            on:click={() => toggleItem(item)}
            on:keydown={(e) => handleKeydown(e, item, index)}
          >
            <span>{item.title}</span>
            <svg
              class={cn(
                accordionItemClasses.chevron,
                itemOpen && accordionItemClasses.chevronOpen
              )}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
              aria-hidden="true"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
            </svg>
          </button>
        </h3>
        <Collapse open={itemOpen}>
          <div
            id="{baseId}-panel-{item.id}"
            role="region"
            aria-labelledby="{baseId}-header-{item.id}"
            class={cn(accordionItemClasses.content, sizeConfig.content)}
          >
            {item.content || ''}
          </div>
        </Collapse>
      </div>
    {/each}
  {:else}
    <!-- Slot-based accordion for custom content -->
    <slot />
  {/if}
</div>
