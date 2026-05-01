<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import { Keys, isKey } from '../utils/keyboard';
  import { tabsClasses, tabsSizeClasses } from './navigation.types';
  import type { TabItem, Size } from '../types';

  // Props
  export let items: TabItem[] = [];
  export let activeId: string = '';
  export let variant: 'line' | 'enclosed' | 'pills' | 'soft' = 'line';
  export let size: Size = 'md';
  export let orientation: 'horizontal' | 'vertical' = 'horizontal';
  export let fullWidth: boolean = false;
  export let id: string = uid('tabs');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: { id: string; item: TabItem };
  }>();

  // Set initial active tab
  $: if (!activeId && items.length > 0) {
    activeId = items[0]!.id;
  }

  $: variantStyles = tabsClasses[variant];

  $: listClasses = cn(
    tabsClasses.list,
    orientation === 'vertical' && tabsClasses.listVertical,
    variantStyles.list,
    fullWidth && 'w-full'
  );

  function getTabClasses(item: TabItem): string {
    const isActive = item.id === activeId;
    return cn(
      tabsClasses.tab,
      tabsSizeClasses[size],
      variantStyles.tab,
      isActive
        ? cn(tabsClasses.tabActive, variantStyles.tabActive)
        : tabsClasses.tabInactive,
      item.disabled && 'opacity-50 cursor-not-allowed',
      fullWidth && 'flex-1 text-center'
    );
  }

  function selectTab(item: TabItem) {
    if (item.disabled) return;
    activeId = item.id;
    dispatch('change', { id: item.id, item });
  }

  function handleKeydown(event: KeyboardEvent, currentIndex: number) {
    const enabledItems = items.filter(i => !i.disabled);
    const currentEnabledIndex = enabledItems.findIndex(i => i.id === items[currentIndex]!.id);

    let newIndex = currentEnabledIndex;

    if (orientation === 'horizontal') {
      if (isKey(event, 'ArrowRight')) {
        event.preventDefault();
        newIndex = (currentEnabledIndex + 1) % enabledItems.length;
      } else if (isKey(event, 'ArrowLeft')) {
        event.preventDefault();
        newIndex = (currentEnabledIndex - 1 + enabledItems.length) % enabledItems.length;
      }
    } else {
      if (isKey(event, 'ArrowDown')) {
        event.preventDefault();
        newIndex = (currentEnabledIndex + 1) % enabledItems.length;
      } else if (isKey(event, 'ArrowUp')) {
        event.preventDefault();
        newIndex = (currentEnabledIndex - 1 + enabledItems.length) % enabledItems.length;
      }
    }

    if (isKey(event, 'Home')) {
      event.preventDefault();
      newIndex = 0;
    } else if (isKey(event, 'End')) {
      event.preventDefault();
      newIndex = enabledItems.length - 1;
    }

    if (newIndex !== currentEnabledIndex) {
      selectTab(enabledItems[newIndex]!);
    }
  }
</script>

<div
  {id}
  class={cn(tabsClasses.container, className)}
  data-testid={testId || undefined}
>
  <div
    class={listClasses}
    role="tablist"
    aria-orientation={orientation}
  >
    {#each items as item, index (item.id)}
      <button
        type="button"
        role="tab"
        id="{id}-tab-{item.id}"
        class={getTabClasses(item)}
        aria-selected={item.id === activeId}
        aria-controls="{id}-panel-{item.id}"
        tabindex={item.id === activeId ? 0 : -1}
        disabled={item.disabled}
        on:click={() => selectTab(item)}
        on:keydown={(e) => handleKeydown(e, index)}
      >
        {#if item.icon}
          <slot name="icon" icon={item.icon}>
            <span class="mr-2">{item.icon}</span>
          </slot>
        {/if}

        {item.label}

        {#if item.badge !== undefined}
          <span class="ml-2 px-1.5 py-0.5 text-xs bg-brand-primary-100 text-brand-primary-700 rounded-full">
            {item.badge}
          </span>
        {/if}
      </button>
    {/each}
  </div>

  <div class={tabsClasses.panel}>
    {#each items as item (item.id)}
      <div
        role="tabpanel"
        id="{id}-panel-{item.id}"
        aria-labelledby="{id}-tab-{item.id}"
        hidden={item.id !== activeId}
        tabindex="0"
      >
        {#if item.id === activeId}
          <slot name="panel" {item} id={item.id}>
            <slot {item} id={item.id} />
          </slot>
        {/if}
      </div>
    {/each}
  </div>
</div>
