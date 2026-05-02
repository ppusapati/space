<script lang="ts">
  import { ErpShell } from '@chetana/ui';
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { moduleStore } from '@chetana/stores';

  let { children } = $props();

  // Module IDs that the top-level sidebar highlights for the active module.
  // The sub-nav is populated dynamically from moduleStore (API-driven) when
  // available, with ErpShell falling back to the static MODULE_REGISTRY.
  const MODULE_IDS = [
    'identity', 'masters', 'finance', 'sales', 'purchase',
    'inventory', 'hr', 'manufacturing', 'projects', 'asset',
    'fulfillment', 'insights', 'workflow', 'budget', 'banking',
    'notifications', 'audit', 'platform', 'communication', 'data', 'land',
    'approvals',
  ];

  const activeModule = $derived(() => {
    const path = $page.url.pathname;
    const segments = path.split('/').filter(Boolean);
    // /forms/[module]/... — treat the module segment as the active module so
    // ErpShell highlights the right sidebar entry.
    if (segments[0] === 'forms' && segments[1] && MODULE_IDS.includes(segments[1])) {
      return segments[1];
    }
    const segment = segments[0] ?? '';
    return MODULE_IDS.includes(segment) ? segment : 'dashboard';
  });

  onMount(() => {
    // Seed the API-driven module store once per session. ErpShell will merge
    // this into the static registry (API provides labels + form counts,
    // static registry provides icons + order).
    void moduleStore.loadModules();
  });

  // Keep the sub-nav in sync with whichever module the user navigated into.
  $effect(() => {
    const mod = activeModule();
    if (mod && mod !== 'dashboard' && moduleStore.selectedModuleId !== mod) {
      void moduleStore.selectModule(mod);
    }
  });
</script>

<ErpShell activeModule={activeModule()} currentPath={$page.url.pathname}>
  {@render children()}
</ErpShell>
