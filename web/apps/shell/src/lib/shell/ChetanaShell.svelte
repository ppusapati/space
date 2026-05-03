<!--
  ChetanaShell.svelte — top nav, side nav, content area for the
  chetana-platform routes (audit / exports / settings).

  Why a chetana-flavoured shell rather than reusing the existing
  ErpShell directly:

    • The chetana platform surface is much smaller than the ERP
      surface (4 modules: audit / exports / settings / dashboard
      vs. 22 ERP modules). A fixed left-nav with the chetana
      hierarchy renders cleaner than the dynamic module store.

    • The chetana spec calls out the route registry as the single
      source of truth (acceptance #5). This component reads from
      `chetanaRouteRegistry` so adding a new route = adding one
      entry, no markup edits.

    • The shell expects a Principal in context (set by the
      (app) +layout.server.ts loader after the auth.cookie check).
      It surfaces the user identity, ITAR posture, and a logout
      action in the top right.
-->
<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { chetanaRouteRegistry, type ChetanaRoute } from "./registry";

  interface Props {
    user: {
      email: string;
      display_name: string;
      tenant_id: string;
      is_us_person: boolean;
      clearance_level: string;
    } | null;
    children: import("svelte").Snippet;
  }

  let { user, children }: Props = $props();

  const activeId = $derived(() => {
    const path = $page.url.pathname;
    for (const r of chetanaRouteRegistry) {
      if (path.startsWith(r.path)) return r.id;
    }
    return "dashboard";
  });

  async function logout() {
    // Best-effort — even if the call fails, drop the local state.
    try {
      await fetch("/api/logout", { method: "POST" });
    } catch {
      /* noop */
    }
    await goto("/login");
  }
</script>

<div class="grid grid-cols-[260px_1fr] grid-rows-[56px_1fr] h-screen bg-background">
  <!-- Top nav -->
  <header
    class="col-span-2 flex items-center justify-between px-lg border-b border-border bg-surface"
  >
    <div class="flex items-center gap-md">
      <a href="/dashboard" class="text-lg font-semibold text-primary">Chetana</a>
      <span class="text-xs text-text-muted">Platform</span>
    </div>
    <div class="flex items-center gap-md">
      {#if user}
        <div class="flex flex-col items-end text-xs">
          <span class="font-medium text-text-primary">{user.display_name || user.email}</span>
          <span class="text-text-muted">
            {user.clearance_level}
            {#if user.is_us_person}<span class="ml-xs text-primary">US</span>{/if}
          </span>
        </div>
      {/if}
      <button
        type="button"
        class="px-md py-xs text-xs rounded border border-border hover:bg-surface-hover"
        onclick={logout}
      >
        Sign out
      </button>
    </div>
  </header>

  <!-- Side nav -->
  <nav class="border-r border-border bg-surface overflow-y-auto py-md">
    <ul class="flex flex-col gap-2xs">
      {#each chetanaRouteRegistry as route (route.id)}
        <li>
          <a
            href={route.path}
            class="flex items-center gap-sm px-md py-sm text-sm rounded-l-none rounded-r-md
                   hover:bg-surface-hover transition-colors
                   {activeId() === route.id
              ? 'bg-primary/10 text-primary border-l-2 border-primary font-medium'
              : 'text-text-secondary border-l-2 border-transparent'}"
          >
            <span class="w-4 text-center">{route.icon}</span>
            <span>{route.label}</span>
          </a>
        </li>
      {/each}
    </ul>
  </nav>

  <!-- Content -->
  <main class="overflow-y-auto p-lg">
    {@render children()}
  </main>
</div>

<style>
  /* Reset any wrapper defaults so the shell takes the viewport. */
  :global(body) {
    overflow: hidden;
  }
</style>
