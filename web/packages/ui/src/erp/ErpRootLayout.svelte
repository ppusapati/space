<script lang="ts">
  /**
   * ErpRootLayout — place this in every module app's root +layout.svelte.
   * Handles: design token CSS import, theme initialization, color-scheme meta tag.
   *
   * Usage:
   *   <ErpRootLayout>
   *     {@render children()}
   *   </ErpRootLayout>
   */
  import { themeStore } from '@samavāya/stores';
  import { onMount } from 'svelte';

  interface Props {
    children: import('svelte').Snippet;
  }

  let { children }: Props = $props();

  const theme = $derived($themeStore.mode);

  onMount(() => {
    document.documentElement.dataset.theme = resolveTheme(theme);

    const mq = window.matchMedia('(prefers-color-scheme: dark)');
    const handleChange = () => {
      if ($themeStore.mode === 'system') {
        document.documentElement.dataset.theme = mq.matches ? 'dark' : 'light';
      }
    };
    mq.addEventListener('change', handleChange);
    return () => mq.removeEventListener('change', handleChange);
  });

  $effect(() => {
    if (typeof document !== 'undefined') {
      document.documentElement.dataset.theme = resolveTheme(theme);
    }
  });

  function resolveTheme(mode: string): string {
    if (mode === 'system') {
      return typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches
        ? 'dark'
        : 'light';
    }
    return mode;
  }
</script>

<svelte:head>
  <meta name="color-scheme" content={theme === 'dark' ? 'dark' : 'light'} />
</svelte:head>

{@render children()}
