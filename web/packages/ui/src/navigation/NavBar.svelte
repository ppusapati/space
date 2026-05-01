<script lang="ts">
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import { navbarClasses } from './navigation.types';

  // Props
  export let fixed: boolean = false;
  export let height: string = '64px';
  export let shadow: boolean = true;
  export let transparent: boolean = false;
  export let id: string = uid('navbar');
  export let testId: string = '';

  let className: string = '';
  export { className as class };

  $: containerClasses = cn(
    navbarClasses.container,
    fixed && navbarClasses.fixed,
    shadow && !transparent && navbarClasses.shadow,
    transparent && navbarClasses.transparent,
    className
  );
</script>

<header
  {id}
  class={containerClasses}
  data-testid={testId || undefined}
  role="banner"
>
  <div class={navbarClasses.inner} style="height: {height};">
    {#if $$slots.brand}
      <div class={navbarClasses.brand}>
        <slot name="brand" />
      </div>
    {/if}

    {#if $$slots.nav}
      <nav class={cn(navbarClasses.nav, navbarClasses.desktop)}>
        <slot name="nav" />
      </nav>
    {/if}

    <div class={navbarClasses.actions}>
      <slot name="actions" />

      {#if $$slots.mobile}
        <div class={navbarClasses.mobile}>
          <slot name="mobile" />
        </div>
      {/if}
    </div>
  </div>

  <slot />
</header>
