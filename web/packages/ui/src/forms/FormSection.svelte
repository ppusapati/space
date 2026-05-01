<script lang="ts">
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import { formSectionClasses } from './formfield.types';

  // Props
  export let title: string = '';
  export let description: string = '';
  export let collapsible: boolean = false;
  export let collapsed: boolean = false;
  export let divider: 'none' | 'top' | 'bottom' | 'both' = 'none';
  export let columns: 1 | 2 | 3 | 4 = 1;
  export let gap: 'none' | 'sm' | 'md' | 'lg' = 'md';
  export let id: string = uid('section');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  // Computed
  $: containerClasses = cn(
    formSectionClasses.container,
    (divider === 'top' || divider === 'both') && formSectionClasses.dividerTop,
    (divider === 'bottom' || divider === 'both') && formSectionClasses.dividerBottom,
    className
  );

  $: contentClasses = cn(
    formSectionClasses.columns[columns],
    formSectionClasses.gap[gap]
  );

  function toggleCollapse() {
    if (collapsible) {
      collapsed = !collapsed;
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (collapsible && (event.key === 'Enter' || event.key === ' ')) {
      event.preventDefault();
      toggleCollapse();
    }
  }
</script>

<section
  {id}
  class={containerClasses}
  data-testid={testId || undefined}
  aria-labelledby={title ? `${id}-title` : undefined}
>
  {#if title || description || $$slots.header}
    <div
      class={cn(
        formSectionClasses.header,
        collapsible && formSectionClasses.headerCollapsible
      )}
      role={collapsible ? 'button' : undefined}
      tabindex={collapsible ? 0 : undefined}
      aria-expanded={collapsible ? !collapsed : undefined}
      aria-controls={collapsible ? `${id}-content` : undefined}
      on:click={toggleCollapse}
      on:keydown={handleKeydown}
    >
      <div>
        {#if $$slots.header}
          <slot name="header" />
        {:else}
          {#if title}
            <h3 id="{id}-title" class={formSectionClasses.title}>{title}</h3>
          {/if}
          {#if description}
            <p class={formSectionClasses.description}>{description}</p>
          {/if}
        {/if}
      </div>

      {#if collapsible}
        <svg
          class={cn(
            formSectionClasses.chevron,
            collapsed && formSectionClasses.chevronCollapsed
          )}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      {/if}
    </div>
  {/if}

  {#if !collapsed}
    <div id="{id}-content" class={contentClasses}>
      <slot />
    </div>
  {/if}

  {#if $$slots.footer}
    <div class="mt-4 pt-4 border-t border-neutral-200">
      <slot name="footer" />
    </div>
  {/if}
</section>
