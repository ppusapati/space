<script lang="ts">
  import { sidebarStore, themeStore, authStore, moduleStore } from '@samavāya/stores';
  import { MODULE_REGISTRY, getEnabledModules, type ModuleDef, type ModuleSection } from './modules.js';
  import type { ApiFormSummary } from '@samavāya/stores/modules';

  interface Props {
    /** Current active module ID (e.g., 'finance', 'hr') */
    activeModule: string;
    /** Current URL path for highlighting active nav items */
    currentPath?: string;
    /** List of enabled module IDs. If undefined, all modules are shown. */
    enabledModules?: string[];
    /** Page title shown in the header. Auto-derived from path if not provided. */
    pageTitle?: string;
    /** App brand name */
    brandName?: string;
    /** Content slot */
    children: import('svelte').Snippet;
  }

  let {
    activeModule,
    currentPath = '',
    enabledModules,
    pageTitle,
    brandName = 'samavāya',
    children,
  }: Props = $props();

  const isCollapsed = $derived($sidebarStore.isCollapsed);
  const userName = $derived(($authStore.user as any)?.name || ($authStore.user as any)?.displayName || 'User');
  const userInitial = $derived(userName.charAt(0).toUpperCase());

  // Logout flow: revoke the server-side session via authStore.logout()
  // (calls AuthService/RevokeSession + clears local + sessionStorage +
  // resets in-memory state), then full-page-reload to /login. The hard
  // navigation is intentional: it drops every in-memory store + cached
  // query state on every other open tab is unaffected (sessionStorage
  // is per-tab; localStorage clears via authStore.logout). A SPA-style
  // goto() would leave RPC clients with the now-revoked token cached.
  let isLoggingOut = $state(false);
  async function handleLogout(): Promise<void> {
    if (isLoggingOut) return;
    isLoggingOut = true;
    try {
      await authStore.logout();
    } finally {
      // Logout always navigates, even on transport failure — local
      // state is already cleared by authStore.logout's finally block.
      if (typeof window !== 'undefined') {
        window.location.assign('/login');
      }
    }
  }

  // Filtered modules based on deployment (static registry)
  const staticModules = $derived(getEnabledModules(enabledModules));

  // Merge API-discovered modules into static registry:
  // Static registry provides icons/paths/order, API adds form counts
  const modules = $derived(() => {
    if (!moduleStore.isApiDriven) return staticModules;

    const apiMap = new Map(moduleStore.modules.map((m) => [m.moduleId, m]));
    return staticModules.map((mod) => {
      const apiMod = apiMap.get(mod.id);
      if (apiMod) {
        return { ...mod, label: apiMod.label || mod.label };
      }
      return mod;
    });
  });

  // Get current module's sub-navigation sections from static registry
  const currentModuleDef = $derived(MODULE_REGISTRY.find((m) => m.id === activeModule));
  const staticSections = $derived(currentModuleDef?.sections ?? []);

  // Build dynamic form items from the API
  const apiFormItems = $derived<ModuleSection | null>(() => {
    if (!moduleStore.isApiDriven || moduleStore.forms.length === 0) return null;
    if (moduleStore.selectedModuleId !== activeModule) return null;

    return {
      title: 'Forms',
      items: moduleStore.forms.map((f: ApiFormSummary) => ({
        label: f.title,
        path: `/forms/${f.moduleId || activeModule}/${f.formId}`,
      })),
    };
  });

  // Merge static sections with API form sections
  const sections = $derived(() => {
    const base = [...staticSections];
    const apiSection = apiFormItems();
    if (apiSection && apiSection.items.length > 0) {
      base.push(apiSection);
    }
    return base;
  });

  // Load API modules on mount
  $effect(() => {
    moduleStore.loadModules();
  });

  // When active module changes, load its forms
  $effect(() => {
    if (activeModule && activeModule !== 'dashboard') {
      moduleStore.selectModule(activeModule);
    }
  });

  // Auto-derive page title from path if not provided
  const derivedTitle = $derived(() => {
    if (pageTitle) return pageTitle;
    if (!currentPath) return currentModuleDef?.label ?? 'Dashboard';
    // Check static sections
    for (const section of sections()) {
      for (const item of section.items) {
        if (currentPath.startsWith(item.path)) return item.label;
      }
    }
    // Check form routes: /forms/[module]/[formId] (new) and /forms/[formId] (legacy redirect)
    if (currentPath.startsWith('/forms/')) {
      const segments = currentPath.split('/forms/')[1]?.split('/') ?? [];
      // Try the last segment first — it's the formId in both URL shapes
      const candidate = segments[segments.length - 1] ?? '';
      const formId = candidate.split('?')[0];
      const form = moduleStore.forms.find((f: ApiFormSummary) => f.formId === formId);
      if (form) return form.title;
    }
    return currentModuleDef?.label ?? 'Dashboard';
  });

  function toggleSidebar() {
    sidebarStore.toggleCollapsed();
  }

  function toggleTheme() {
    const mode = $themeStore.mode;
    const next = mode === 'light' ? 'dark' : mode === 'dark' ? 'system' : 'light';
    themeStore.setMode(next);
  }

  function isModuleActive(mod: ModuleDef): boolean {
    if (mod.id === 'dashboard') return currentPath === '/' || currentPath === '/dashboard';
    return mod.id === activeModule;
  }

  function isNavItemActive(path: string): boolean {
    return currentPath.startsWith(path);
  }
</script>

<div class="erp-layout" class:sidebar-collapsed={isCollapsed}>
  <!-- Module Sidebar (left) -->
  <aside class="module-sidebar">
    <div class="module-sidebar-header">
      <span class="brand-icon">S</span>
      {#if !isCollapsed}
        <span class="brand-name">{brandName}</span>
      {/if}
    </div>

    <nav class="module-nav">
      {#each modules() as mod}
        <a
          href={mod.path}
          class="module-item"
          class:active={isModuleActive(mod)}
          title={isCollapsed ? mod.label : undefined}
        >
          <svg class="module-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d={mod.icon} />
          </svg>
          {#if !isCollapsed}
            <span class="module-label">{mod.label}</span>
          {/if}
        </a>
      {/each}
    </nav>

    <div class="module-sidebar-footer">
      <button class="sidebar-action" onclick={toggleSidebar} title="Toggle sidebar">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          {#if isCollapsed}
            <path d="M9 18l6-6-6-6" />
          {:else}
            <path d="M15 18l-6-6 6-6" />
          {/if}
        </svg>
      </button>
      <button class="sidebar-action" onclick={toggleTheme} title="Toggle theme">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          {#if $themeStore.mode === 'dark'}
            <circle cx="12" cy="12" r="5" /><line x1="12" y1="1" x2="12" y2="3" /><line x1="12" y1="21" x2="12" y2="23" /><line x1="4.22" y1="4.22" x2="5.64" y2="5.64" /><line x1="18.36" y1="18.36" x2="19.78" y2="19.78" /><line x1="1" y1="12" x2="3" y2="12" /><line x1="21" y1="12" x2="23" y2="12" /><line x1="4.22" y1="19.78" x2="5.64" y2="18.36" /><line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
          {:else}
            <path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z" />
          {/if}
        </svg>
      </button>
    </div>
  </aside>

  <!-- Sub-navigation (for active module sections) -->
  {#if sections().length > 0 && !isCollapsed}
    <aside class="sub-nav">
      <div class="sub-nav-header">
        <h2 class="sub-nav-title">{currentModuleDef?.label}</h2>
      </div>
      <nav class="sub-nav-content">
        {#each sections() as section}
          <div class="sub-nav-section">
            <h3 class="sub-nav-section-title">{section.title}</h3>
            <ul class="sub-nav-list">
              {#each section.items as item}
                <li>
                  <a
                    href={item.path}
                    class="sub-nav-link"
                    class:active={isNavItemActive(item.path)}
                  >
                    {item.label}
                  </a>
                </li>
              {/each}
            </ul>
          </div>
        {/each}
      </nav>
    </aside>
  {/if}

  <!-- Main content area -->
  <div class="main-wrapper">
    <header class="app-header">
      <div class="header-left">
        <h1 class="page-title">{derivedTitle()}</h1>
      </div>
      <div class="header-right">
        <div class="user-menu">
          <span class="user-name">{userName}</span>
          <div class="user-avatar">{userInitial}</div>
          <button
            type="button"
            class="logout-btn"
            onclick={handleLogout}
            disabled={isLoggingOut}
            aria-label="Log out"
            title="Log out"
          >
            {#if isLoggingOut}
              <span class="logout-label">Signing out…</span>
            {:else}
              <svg
                class="logout-icon"
                xmlns="http://www.w3.org/2000/svg"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                aria-hidden="true"
              >
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                <polyline points="16 17 21 12 16 7" />
                <line x1="21" y1="12" x2="9" y2="12" />
              </svg>
              <span class="logout-label">Logout</span>
            {/if}
          </button>
        </div>
      </div>
    </header>

    <main class="main-content">
      {@render children()}
    </main>
  </div>
</div>

<style>
  /* ============================================================
   * Layout variables
   * ============================================================ */
  .erp-layout {
    --module-sidebar-width: 220px;
    --module-sidebar-collapsed: 56px;
    --sub-nav-width: 220px;
    --header-height: 56px;
    display: flex;
    min-height: 100vh;
  }

  /* ============================================================
   * Module Sidebar (left icon rail / expanded nav)
   * ============================================================ */
  .module-sidebar {
    width: var(--module-sidebar-width);
    background-color: var(--color-surface, #fff);
    border-right: 1px solid var(--color-border, #e5e7eb);
    display: flex;
    flex-direction: column;
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    z-index: 50;
    transition: width 0.2s ease;
  }

  .sidebar-collapsed .module-sidebar {
    width: var(--module-sidebar-collapsed);
  }

  .module-sidebar-header {
    height: var(--header-height);
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0 0.75rem;
    border-bottom: 1px solid var(--color-border, #e5e7eb);
    overflow: hidden;
  }

  .brand-icon {
    width: 32px;
    height: 32px;
    min-width: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--color-primary, #4f46e5);
    color: white;
    font-weight: 700;
    border-radius: 0.5rem;
    font-size: 0.875rem;
  }

  .brand-name {
    font-weight: 600;
    font-size: 1rem;
    color: var(--color-text, #111827);
    white-space: nowrap;
  }

  .module-nav {
    flex: 1;
    padding: 0.5rem;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .module-item {
    display: flex;
    align-items: center;
    gap: 0.625rem;
    padding: 0.5rem 0.625rem;
    border-radius: 0.375rem;
    color: var(--color-text-secondary, #6b7280);
    text-decoration: none;
    font-size: 0.8125rem;
    font-weight: 500;
    transition: background-color 0.15s, color 0.15s;
    white-space: nowrap;
    overflow: hidden;
  }

  .module-item:hover {
    background-color: var(--color-border, #f3f4f6);
    color: var(--color-text, #111827);
    text-decoration: none;
  }

  .module-item.active {
    background-color: var(--color-primary-light, #eef2ff);
    color: var(--color-primary, #4f46e5);
  }

  .module-icon {
    width: 18px;
    height: 18px;
    min-width: 18px;
  }

  .module-label {
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .module-sidebar-footer {
    padding: 0.5rem;
    border-top: 1px solid var(--color-border, #e5e7eb);
    display: flex;
    gap: 0.25rem;
  }

  .sidebar-action {
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 0.375rem;
    color: var(--color-text-secondary, #6b7280);
    transition: background-color 0.15s;
    cursor: pointer;
    background: none;
    border: none;
    font: inherit;
  }

  .sidebar-action:hover {
    background-color: var(--color-border, #f3f4f6);
  }

  .sidebar-action svg {
    width: 16px;
    height: 16px;
  }

  /* ============================================================
   * Sub-navigation (right of module sidebar, for section nav)
   * ============================================================ */
  .sub-nav {
    width: var(--sub-nav-width);
    background-color: var(--color-background, #fafafa);
    border-right: 1px solid var(--color-border, #e5e7eb);
    position: fixed;
    top: 0;
    left: var(--module-sidebar-width);
    bottom: 0;
    z-index: 40;
    display: flex;
    flex-direction: column;
    transition: left 0.2s ease;
  }

  .sidebar-collapsed .sub-nav {
    left: var(--module-sidebar-collapsed);
  }

  .sub-nav-header {
    height: var(--header-height);
    display: flex;
    align-items: center;
    padding: 0 1rem;
    border-bottom: 1px solid var(--color-border, #e5e7eb);
  }

  .sub-nav-title {
    font-size: 0.9375rem;
    font-weight: 600;
    color: var(--color-text, #111827);
  }

  .sub-nav-content {
    flex: 1;
    padding: 0.75rem 0;
    overflow-y: auto;
  }

  .sub-nav-section {
    padding: 0 0.5rem;
    margin-bottom: 1.25rem;
  }

  .sub-nav-section-title {
    font-size: 0.6875rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--color-text-secondary, #9ca3af);
    padding: 0 0.625rem;
    margin-bottom: 0.375rem;
  }

  .sub-nav-list {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .sub-nav-link {
    display: block;
    padding: 0.375rem 0.625rem;
    border-radius: 0.375rem;
    color: var(--color-text-secondary, #6b7280);
    text-decoration: none;
    font-size: 0.8125rem;
    transition: background-color 0.15s, color 0.15s;
  }

  .sub-nav-link:hover {
    background-color: var(--color-border, #e5e7eb);
    color: var(--color-text, #111827);
    text-decoration: none;
  }

  .sub-nav-link.active {
    background-color: var(--color-primary, #4f46e5);
    color: white;
  }

  /* ============================================================
   * Main content wrapper
   * ============================================================ */
  .main-wrapper {
    flex: 1;
    display: flex;
    flex-direction: column;
    /* Offset for both sidebars */
    margin-left: calc(var(--module-sidebar-width) + var(--sub-nav-width));
    transition: margin-left 0.2s ease;
  }

  .sidebar-collapsed .main-wrapper {
    margin-left: calc(var(--module-sidebar-collapsed) + var(--sub-nav-width));
  }

  /* When no sub-nav (dashboard or collapsed) */
  .erp-layout:not(:has(.sub-nav)) .main-wrapper {
    margin-left: var(--module-sidebar-width);
  }

  .sidebar-collapsed:not(:has(.sub-nav)) .main-wrapper {
    margin-left: var(--module-sidebar-collapsed);
  }

  .app-header {
    height: var(--header-height);
    background-color: var(--color-background, #fff);
    border-bottom: 1px solid var(--color-border, #e5e7eb);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 1.5rem;
    position: sticky;
    top: 0;
    z-index: 30;
  }

  .page-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-text, #111827);
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .user-menu {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .user-name {
    font-size: 0.8125rem;
    color: var(--color-text-secondary, #6b7280);
  }

  .user-avatar {
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--color-primary, #4f46e5);
    color: white;
    font-weight: 600;
    font-size: 0.8125rem;
    border-radius: 9999px;
  }

  .logout-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0.375rem 0.75rem;
    margin-left: 0.5rem;
    border: 1px solid var(--color-border, #e5e7eb);
    border-radius: 0.375rem;
    background: var(--color-surface, #fff);
    color: var(--color-text-secondary, #6b7280);
    font-size: 0.8125rem;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.15s ease, color 0.15s ease, border-color 0.15s ease;
  }

  .logout-btn:hover:not(:disabled) {
    background: var(--color-danger-bg, #fee2e2);
    color: var(--color-danger, #b91c1c);
    border-color: var(--color-danger, #b91c1c);
  }

  .logout-btn:focus-visible {
    outline: 2px solid var(--color-primary, #4f46e5);
    outline-offset: 2px;
  }

  .logout-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .logout-icon {
    flex-shrink: 0;
  }

  .main-content {
    flex: 1;
    padding: 1.5rem;
    background-color: var(--color-surface, #f9fafb);
  }
</style>
