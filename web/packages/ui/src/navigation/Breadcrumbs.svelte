<script lang="ts">
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import { breadcrumbsClasses, breadcrumbsSizeClasses } from './navigation.types';
  import type { BreadcrumbItem, Size } from '../types';

  // Props
  export let items: BreadcrumbItem[] = [];
  export let separator: string = '/';
  export let maxItems: number = 0;
  export let size: Size = 'md';
  export let id: string = uid('breadcrumbs');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  // Collapse items if maxItems is set
  $: displayItems = maxItems > 0 && items.length > maxItems
    ? [
        items[0]!,
        { label: '...', href: undefined, icon: undefined },
        ...items.slice(-(maxItems - 1)),
      ]
    : items;

  $: containerClasses = cn(
    breadcrumbsClasses.container,
    breadcrumbsSizeClasses[size],
    className
  );
</script>

<nav
  {id}
  class={containerClasses}
  data-testid={testId || undefined}
  aria-label="Breadcrumb"
>
  <ol class="flex items-center flex-wrap gap-1">
    {#each displayItems as item, index (index)}
      <li class={breadcrumbsClasses.item}>
        {#if index > 0}
          <span class={breadcrumbsClasses.separator} aria-hidden="true">
            <slot name="separator">
              {separator}
            </slot>
          </span>
        {/if}

        {#if index === displayItems.length - 1}
          <!-- Current page -->
          <span class={breadcrumbsClasses.current} aria-current="page">
            {#if item.icon}
              <slot name="icon" icon={item.icon}>
                <span class={breadcrumbsClasses.icon}>{item.icon}</span>
              </slot>
            {/if}
            {item.label}
          </span>
        {:else if item.href}
          <!-- Link -->
          <a href={item.href} class={breadcrumbsClasses.link}>
            {#if item.icon}
              <slot name="icon" icon={item.icon}>
                <span class={breadcrumbsClasses.icon}>{item.icon}</span>
              </slot>
            {/if}
            {item.label}
          </a>
        {:else}
          <!-- Non-clickable -->
          <span class={breadcrumbsClasses.link}>
            {item.label}
          </span>
        {/if}
      </li>
    {/each}
  </ol>
</nav>
