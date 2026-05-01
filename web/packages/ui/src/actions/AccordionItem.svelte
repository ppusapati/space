<script lang="ts">
  import { getContext, createEventDispatcher } from 'svelte';
  import type { Writable } from 'svelte/store';
  import type { Size } from '../types';
  import { accordionItemClasses, accordionSizeClasses } from './actions.types';
  import { cn, uid } from '../utils';
  import Collapse from './Collapse.svelte';

  /** Item ID */
  export let id: string = uid('accordion-item');
  /** Item title */
  export let title: string;
  /** Is disabled */
  export let disabled: boolean = false;
  /** Custom class */
  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    toggle: { isOpen: boolean };
  }>();

  // Get context from parent Accordion (if exists)
  const openItems = getContext<Writable<Set<string>>>('accordion-open-items');
  const multiple = getContext<boolean>('accordion-multiple') ?? false;
  const size = getContext<Size>('accordion-size') ?? 'md';
  const baseId = getContext<string>('accordion-base-id') ?? uid('accordion');

  // Local state if no parent context
  let localOpen = false;

  $: isOpen = openItems ? $openItems.has(id) : localOpen;
  $: sizeConfig = accordionSizeClasses[size];

  function toggle() {
    if (disabled) return;

    if (openItems) {
      openItems.update((items) => {
        const newItems = new Set(items);
        if (newItems.has(id)) {
          newItems.delete(id);
        } else {
          if (!multiple) {
            newItems.clear();
          }
          newItems.add(id);
        }
        return newItems;
      });
    } else {
      localOpen = !localOpen;
    }

    dispatch('toggle', { isOpen: !isOpen });
  }

  function handleKeydown(event: KeyboardEvent) {
    if (disabled) return;

    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      toggle();
    }
  }
</script>

<div class={cn('accordion-item', className)}>
  <h3>
    <button
      type="button"
      id="{baseId}-header-{id}"
      class={cn(
        accordionItemClasses.header,
        sizeConfig.header,
        disabled
          ? accordionItemClasses.headerDisabled
          : isOpen
            ? accordionItemClasses.headerOpen
            : accordionItemClasses.headerClosed
      )}
      aria-expanded={isOpen}
      aria-controls="{baseId}-panel-{id}"
      aria-disabled={disabled}
      {disabled}
      on:click={toggle}
      on:keydown={handleKeydown}
    >
      <span class="flex-1 text-left">
        <slot name="title">
          {title}
        </slot>
      </span>
      <slot name="icon" {isOpen}>
        <svg
          class={cn(
            accordionItemClasses.chevron,
            isOpen && accordionItemClasses.chevronOpen
          )}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
          aria-hidden="true"
        >
          <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
        </svg>
      </slot>
    </button>
  </h3>
  <Collapse open={isOpen}>
    <div
      id="{baseId}-panel-{id}"
      role="region"
      aria-labelledby="{baseId}-header-{id}"
      class={cn(accordionItemClasses.content, sizeConfig.content)}
    >
      <slot />
    </div>
  </Collapse>
</div>
