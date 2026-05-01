<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { AppShellProps } from './layout.types';
  import { appShellClasses } from './layout.types';
  import { cn } from '../utils';

  type $$Props = AppShellProps;

  export let fixedHeader: $$Props['fixedHeader'] = true;
  export let fixedSidebar: $$Props['fixedSidebar'] = true;
  export let sidebarCollapsed: $$Props['sidebarCollapsed'] = false;
  export let sidebarPosition: $$Props['sidebarPosition'] = 'left';
  export let headerHeight: $$Props['headerHeight'] = '64px';
  export let sidebarWidth: $$Props['sidebarWidth'] = '256px';
  export let collapsedWidth: $$Props['collapsedWidth'] = '64px';
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{ toggleSidebar: boolean }>();

  $: currentSidebarWidth = sidebarCollapsed ? collapsedWidth : sidebarWidth;
  $: positionClass = sidebarPosition === 'right' ? appShellClasses.sidebarRight : appShellClasses.sidebarLeft;

  $: mainMargin = (() => {
    if (!$$slots.sidebar) return '';
    const marginProp = sidebarPosition === 'right' ? 'margin-right' : 'margin-left';
    return `${marginProp}: ${currentSidebarWidth}`;
  })();

  $: mainPaddingTop = fixedHeader && $$slots.header ? `padding-top: ${headerHeight}` : '';

  function toggleSidebar() {
    sidebarCollapsed = !sidebarCollapsed;
    dispatch('toggleSidebar', sidebarCollapsed);
  }
</script>

<div class={cn(appShellClasses.container, className)}>
  <!-- Header -->
  {#if $$slots.header}
    <header
      class={cn(
        appShellClasses.header,
        fixedHeader && appShellClasses.headerFixed
      )}
      style="height: {headerHeight}; {$$slots.sidebar && fixedSidebar ? mainMargin : ''}"
    >
      <slot name="header" {toggleSidebar} {sidebarCollapsed} />
    </header>
  {/if}

  <!-- Sidebar -->
  {#if $$slots.sidebar}
    <aside
      class={cn(
        appShellClasses.sidebar,
        fixedSidebar && appShellClasses.sidebarFixed,
        positionClass
      )}
      style="width: {currentSidebarWidth}; {fixedHeader && fixedSidebar ? `top: ${headerHeight}` : ''}"
    >
      <slot name="sidebar" {sidebarCollapsed} {toggleSidebar} />
    </aside>
  {/if}

  <!-- Main Content -->
  <main
    class={cn(appShellClasses.main, appShellClasses.mainWithSidebar)}
    style="{mainPaddingTop}; {$$slots.sidebar && fixedSidebar ? mainMargin : ''}"
  >
    <slot />
  </main>

  <!-- Footer -->
  {#if $$slots.footer}
    <footer
      class={appShellClasses.footer}
      style={$$slots.sidebar && fixedSidebar ? mainMargin : ''}
    >
      <slot name="footer" />
    </footer>
  {/if}
</div>
