<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import { sidebarClasses } from './navigation.types';

  // Props
  export let collapsed: boolean = false;
  export let width: string = '256px';
  export let collapsedWidth: string = '64px';
  export let position: 'left' | 'right' = 'left';
  export let fixed: boolean = false;
  export let collapsible: boolean = true;
  export let overlay: boolean = false;
  export let id: string = uid('sidebar');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    collapse: { collapsed: boolean };
  }>();

  $: currentWidth = collapsed ? collapsedWidth : width;

  $: containerClasses = cn(
    sidebarClasses.container,
    position === 'right' && sidebarClasses.containerRight,
    fixed && sidebarClasses.fixed,
    position === 'left' ? 'left-0' : 'right-0',
    className
  );

  function toggleCollapse() {
    collapsed = !collapsed;
    dispatch('collapse', { collapsed });
  }

  function handleOverlayClick() {
    if (overlay) {
      collapsed = true;
      dispatch('collapse', { collapsed: true });
    }
  }
</script>

{#if overlay && !collapsed}
  <div
    class={sidebarClasses.overlay}
    on:click={handleOverlayClick}
    on:keydown={(e) => e.key === 'Escape' && handleOverlayClick()}
    role="button"
    tabindex="-1"
    aria-label="Close sidebar"
  >
    <div class={sidebarClasses.overlayBackdrop}></div>
  </div>
{/if}

<aside
  {id}
  class={containerClasses}
  style="width: {currentWidth};"
  data-testid={testId || undefined}
  data-collapsed={collapsed}
>
  {#if $$slots.header}
    <div class={sidebarClasses.header}>
      <slot name="header" {collapsed} />

      {#if collapsible && !overlay}
        <button
          type="button"
          class={sidebarClasses.collapseBtn}
          on:click={toggleCollapse}
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          <svg
            class="w-5 h-5 transition-transform"
            class:rotate-180={position === 'left' ? collapsed : !collapsed}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
      {/if}
    </div>
  {/if}

  <nav class={sidebarClasses.content}>
    <slot {collapsed} />
  </nav>

  {#if $$slots.footer}
    <div class={sidebarClasses.footer}>
      <slot name="footer" {collapsed} />
    </div>
  {/if}
</aside>
