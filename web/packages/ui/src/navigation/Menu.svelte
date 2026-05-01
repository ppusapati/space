<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import { Keys, isKey } from '../utils/keyboard';
  import { menuClasses } from './navigation.types';
  import type { MenuItem, Size } from '../types';

  // Props
  export let items: MenuItem[] = [];
  export let open: boolean = false;
  export let trigger: HTMLElement | undefined = undefined;
  export let position: 'bottom-start' | 'bottom-end' | 'top-start' | 'top-end' = 'bottom-start';
  export let size: Size = 'md';
  export let id: string = uid('menu');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  let menuRef: HTMLDivElement;
  let focusedIndex = -1;

  const dispatch = createEventDispatcher<{
    select: { item: MenuItem };
    close: void;
  }>();

  // Position styles
  $: positionClasses = {
    'bottom-start': 'top-full left-0 mt-1',
    'bottom-end': 'top-full right-0 mt-1',
    'top-start': 'bottom-full left-0 mb-1',
    'top-end': 'bottom-full right-0 mb-1',
  }[position];

  // Enabled items for keyboard navigation
  $: enabledItems = items.filter(i => !i.disabled && !i.divider);

  function handleSelect(item: MenuItem) {
    if (item.disabled) return;
    dispatch('select', { item });
    open = false;
    dispatch('close');
  }

  function handleKeydown(event: KeyboardEvent) {
    if (!open) return;

    if (isKey(event, 'Escape')) {
      event.preventDefault();
      open = false;
      dispatch('close');
      trigger?.focus();
      return;
    }

    if (isKey(event, 'ArrowDown')) {
      event.preventDefault();
      focusedIndex = (focusedIndex + 1) % enabledItems.length;
      focusItem(focusedIndex);
    } else if (isKey(event, 'ArrowUp')) {
      event.preventDefault();
      focusedIndex = (focusedIndex - 1 + enabledItems.length) % enabledItems.length;
      focusItem(focusedIndex);
    } else if (isKey(event, 'Home')) {
      event.preventDefault();
      focusedIndex = 0;
      focusItem(focusedIndex);
    } else if (isKey(event, 'End')) {
      event.preventDefault();
      focusedIndex = enabledItems.length - 1;
      focusItem(focusedIndex);
    } else if (isKey(event, 'Enter') || isKey(event, 'Space')) {
      event.preventDefault();
      if (focusedIndex >= 0) {
        handleSelect(enabledItems[focusedIndex]!);
      }
    }
  }

  function focusItem(index: number) {
    const item = enabledItems[index];
    if (item && menuRef) {
      const button = menuRef.querySelector(`[data-menu-item="${item.id}"]`) as HTMLElement;
      button?.focus();
    }
  }

  function handleClickOutside(event: MouseEvent) {
    if (menuRef && !menuRef.contains(event.target as Node) && !trigger?.contains(event.target as Node)) {
      open = false;
      dispatch('close');
    }
  }

  onMount(() => {
    document.addEventListener('click', handleClickOutside);
    document.addEventListener('keydown', handleKeydown);
  });

  onDestroy(() => {
    document.removeEventListener('click', handleClickOutside);
    document.removeEventListener('keydown', handleKeydown);
  });

  // Focus first item when opened
  $: if (open && menuRef) {
    focusedIndex = 0;
    setTimeout(() => focusItem(0), 10);
  }
</script>

{#if open}
  <div
    bind:this={menuRef}
    {id}
    class={cn(menuClasses.container, positionClasses, className)}
    data-testid={testId || undefined}
    role="menu"
    aria-orientation="vertical"
    tabindex="-1"
  >
    {#each items as item (item.id)}
      {#if item.divider}
        <div class={menuClasses.divider} role="separator"></div>
      {:else}
        <button
          type="button"
          class={cn(
            menuClasses.item,
            item.disabled && menuClasses.itemDisabled
          )}
          role="menuitem"
          data-menu-item={item.id}
          disabled={item.disabled}
          on:click={() => handleSelect(item)}
        >
          {#if item.icon}
            <slot name="icon" icon={item.icon}>
              <span class={menuClasses.itemIcon}>{item.icon}</span>
            </slot>
          {/if}

          <span>{item.label}</span>

          {#if item.children && item.children.length > 0}
            <svg class={menuClasses.submenuIndicator} width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          {/if}
        </button>
      {/if}
    {/each}
  </div>
{/if}
