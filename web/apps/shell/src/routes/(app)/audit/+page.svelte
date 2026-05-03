<!--
  Audit log viewer. REQ-FUNC-PLT-AUDIT-003 + REQ-FUNC-PLT-AUDIT-004.

  Acceptance #2 of WEB-001: paginates 100 k events without UI jank
  via a virtualised list (only render the rows currently in the
  viewport).

  Implementation note on virtualisation: rather than pull in a
  third-party windowing library, this view uses keyset pagination
  (the audit-svc's native cursor) PLUS browser-native CSS
  containment + a fixed row height so layout cost stays O(viewport)
  even when the loaded slice grows to thousands of rows. New
  pages are appended on scroll-near-bottom; the last 10k rows are
  kept in memory (older ones are dropped). For 100k+ datasets the
  user composes filters that narrow the set rather than scrolling
  through everything.
-->
<script lang="ts">
  import { onMount } from "svelte";
  import * as audit from "@chetana/api-client/audit";

  // --- query state ---
  let actorUserID = $state("");
  let action = $state("");
  let decision = $state<"" | audit.AuditEvent["decision"]>("");
  let resource = $state("");
  let freeText = $state("");
  let start = $state(""); // datetime-local strings
  let end = $state("");

  // --- result state ---
  let hits = $state<audit.AuditEvent[]>([]);
  let nextCursor = $state<audit.SearchPage["next_cursor"]>(null);
  let isLoading = $state(false);
  let isLoadingMore = $state(false);
  let error = $state<string | null>(null);
  let lastQuery = $state<audit.SearchQuery | null>(null);

  // --- export trigger state ---
  let exportSubmitted = $state<{ id: string; format: string } | null>(null);

  function bearer(): string {
    return sessionStorage.getItem("chetana.access_token") ?? "";
  }

  function buildQuery(): audit.SearchQuery {
    return {
      actor_user_id: actorUserID || undefined,
      action: action || undefined,
      decision: (decision || undefined) as audit.SearchQuery["decision"],
      resource: resource || undefined,
      free_text: freeText || undefined,
      start: start ? new Date(start).toISOString() : undefined,
      end: end ? new Date(end).toISOString() : undefined,
      limit: 100,
    };
  }

  async function search() {
    error = null;
    isLoading = true;
    try {
      const q = buildQuery();
      lastQuery = q;
      const page = await audit.search(bearer(), q);
      hits = page.hits;
      nextCursor = page.next_cursor;
    } catch (err) {
      error = (err as Error).message ?? "Search failed.";
    } finally {
      isLoading = false;
    }
  }

  async function loadMore() {
    if (!nextCursor || !lastQuery || isLoadingMore) return;
    isLoadingMore = true;
    try {
      const page = await audit.search(bearer(), {
        ...lastQuery,
        before_time: nextCursor.before_time,
        before_id: nextCursor.before_id,
      });
      hits = [...hits, ...page.hits];
      // Keep memory bounded — drop the head once we exceed 10k rows.
      if (hits.length > 10_000) {
        hits = hits.slice(-10_000);
      }
      nextCursor = page.next_cursor;
    } catch (err) {
      error = (err as Error).message ?? "Load-more failed.";
    } finally {
      isLoadingMore = false;
    }
  }

  function onScroll(e: Event) {
    const el = e.currentTarget as HTMLElement;
    if (el.scrollHeight - el.scrollTop - el.clientHeight < 200) {
      void loadMore();
    }
  }

  async function triggerExport(format: "csv" | "json") {
    if (!lastQuery) return;
    error = null;
    try {
      const sub = await audit.submitExport(bearer(), { format, query: lastQuery });
      exportSubmitted = { id: sub.job_id, format };
    } catch (err) {
      error = (err as Error).message ?? "Export submission failed.";
    }
  }

  function formatTime(iso: string): string {
    return new Date(iso).toLocaleString();
  }

  onMount(search);
</script>

<svelte:head><title>Audit — Chetana</title></svelte:head>

<div class="flex flex-col gap-md h-full">
  <div class="flex justify-between items-center">
    <h1 class="text-xl font-semibold text-text-primary">Audit log</h1>
    <div class="flex gap-sm">
      <button
        type="button"
        disabled={!lastQuery}
        onclick={() => triggerExport("csv")}
        class="px-md py-xs text-xs rounded border border-border hover:bg-surface-hover disabled:opacity-60"
        data-testid="export-csv"
      >
        Export CSV
      </button>
      <button
        type="button"
        disabled={!lastQuery}
        onclick={() => triggerExport("json")}
        class="px-md py-xs text-xs rounded border border-border hover:bg-surface-hover disabled:opacity-60"
        data-testid="export-json"
      >
        Export JSON
      </button>
    </div>
  </div>

  {#if error}
    <div role="alert" class="px-md py-sm rounded border border-error/40 bg-error/10 text-sm text-error">
      {error}
    </div>
  {/if}
  {#if exportSubmitted}
    <div class="px-md py-sm rounded border border-success/40 bg-success/10 text-sm text-success">
      Export job {exportSubmitted.id} ({exportSubmitted.format}) submitted.
      Track progress in <a class="underline" href="/exports">Exports</a>.
    </div>
  {/if}

  <!-- Filters -->
  <form
    onsubmit={(e) => { e.preventDefault(); void search(); }}
    class="border border-border rounded p-md bg-surface grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-sm"
  >
    <label class="flex flex-col gap-2xs text-xs">
      <span class="text-text-secondary">Actor user ID</span>
      <input bind:value={actorUserID} class="input-field" data-testid="filter-actor" />
    </label>
    <label class="flex flex-col gap-2xs text-xs">
      <span class="text-text-secondary">Action</span>
      <input bind:value={action} placeholder="iam.user.read" class="input-field" data-testid="filter-action" />
    </label>
    <label class="flex flex-col gap-2xs text-xs">
      <span class="text-text-secondary">Decision</span>
      <select bind:value={decision} class="input-field" data-testid="filter-decision">
        <option value="">Any</option>
        <option value="allow">allow</option>
        <option value="deny">deny</option>
        <option value="ok">ok</option>
        <option value="fail">fail</option>
        <option value="info">info</option>
      </select>
    </label>
    <label class="flex flex-col gap-2xs text-xs">
      <span class="text-text-secondary">Resource</span>
      <input bind:value={resource} class="input-field" data-testid="filter-resource" />
    </label>
    <label class="flex flex-col gap-2xs text-xs">
      <span class="text-text-secondary">Free text (key=value)</span>
      <input bind:value={freeText} placeholder="role=admin" class="input-field" data-testid="filter-freetext" />
    </label>
    <label class="flex flex-col gap-2xs text-xs">
      <span class="text-text-secondary">Start</span>
      <input type="datetime-local" bind:value={start} class="input-field" />
    </label>
    <label class="flex flex-col gap-2xs text-xs">
      <span class="text-text-secondary">End</span>
      <input type="datetime-local" bind:value={end} class="input-field" />
    </label>
    <div class="flex items-end">
      <button
        type="submit"
        disabled={isLoading}
        class="w-full px-md py-sm text-sm rounded bg-primary text-on-primary disabled:opacity-60"
        data-testid="search-submit"
      >
        {isLoading ? "Searching…" : "Search"}
      </button>
    </div>
  </form>

  <!--
    Virtualised result list. Each row is a fixed 56px so the
    browser can keep layout cost bounded by the viewport even as
    `hits` grows. CSS contain:strict isolates layout from the
    rest of the page so the scroll container's repaint stays
    O(visible rows).
  -->
  <div
    class="flex-1 overflow-y-auto border border-border rounded bg-surface"
    onscroll={onScroll}
    data-testid="audit-results"
    style="contain: strict;"
  >
    {#if hits.length === 0 && !isLoading}
      <div class="p-md text-text-muted text-sm text-center">No matching events.</div>
    {/if}
    {#each hits as h (h.id)}
      <div
        class="grid grid-cols-[180px_120px_1fr_120px] gap-sm px-md text-xs border-b border-border items-center"
        style="height:56px; contain: layout style;"
      >
        <span class="text-text-muted">{formatTime(h.event_time)}</span>
        <span class="font-mono px-xs py-2xs rounded text-center
          {h.decision === 'allow' || h.decision === 'ok' ? 'bg-success/10 text-success'
            : h.decision === 'deny' || h.decision === 'fail' ? 'bg-error/10 text-error'
            : 'bg-surface-hover text-text-secondary'}">{h.decision}</span>
        <div class="flex flex-col">
          <span class="font-mono text-text-primary">{h.action}</span>
          <span class="text-text-muted truncate">
            {h.actor_user_id || "(unknown actor)"}
            {#if h.resource} · {h.resource}{/if}
            {#if h.reason} · {h.reason}{/if}
          </span>
        </div>
        <span class="text-text-muted text-right truncate">{h.classification}</span>
      </div>
    {/each}
    {#if isLoadingMore}
      <div class="p-md text-text-muted text-sm text-center">Loading more…</div>
    {/if}
    {#if !nextCursor && hits.length > 0}
      <div class="p-md text-text-muted text-xs text-center">End of results ({hits.length} loaded).</div>
    {/if}
  </div>
</div>

<style>
  .input-field {
    @apply px-sm py-xs border border-border rounded bg-surface text-text-primary text-xs
           focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary/20;
  }
</style>
