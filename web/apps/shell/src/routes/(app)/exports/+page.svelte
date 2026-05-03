<!--
  Exports list. REQ-FUNC-CMN-005 + REQ-FUNC-PLT-AUDIT-004.

  Acceptance #3 of WEB-001: surfaces job progress via WS push,
  not polling. The realtime gateway broadcasts
  `notify.inapp.v1` events when an export job transitions; this
  page subscribes to that topic via the chetana realtime client
  and updates rows in place.
-->
<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import * as iam from "@chetana/api-client/iam";
  import { createRealtimeClient, type RealtimeClient } from "@chetana/api-client/realtime";

  interface ExportJob {
    id: string;
    kind: string;
    status: "queued" | "running" | "succeeded" | "failed" | "expired";
    enqueued_at: string;
    completed_at?: string;
    presigned_url?: string;
    presigned_until?: string;
    bytes_total?: number;
    last_error?: string;
  }

  let jobs = $state<ExportJob[]>([]);
  let isLoading = $state(true);
  let error = $state<string | null>(null);
  let rt: RealtimeClient | null = null;
  let rtState = $state<string>("idle");

  function bearer(): string {
    return sessionStorage.getItem("chetana.access_token") ?? "";
  }

  async function load() {
    error = null;
    isLoading = true;
    try {
      // Plain GET against the chetana cmd-layer; api-client doesn't
      // model exports as a typed surface yet because the export
      // RPC shape is small + stable enough that a fetch here is
      // appropriate.
      const res = await fetch("/v1/export/jobs", {
        headers: { Authorization: `Bearer ${bearer()}` },
        credentials: "include",
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const body = (await res.json()) as { jobs: ExportJob[] };
      jobs = body.jobs;
    } catch (err) {
      error = (err as Error).message ?? "Failed to load exports.";
    } finally {
      isLoading = false;
    }
  }

  function pollStateBadge(): string {
    return rtState;
  }

  function formatTime(iso: string | undefined): string {
    return iso ? new Date(iso).toLocaleString() : "—";
  }

  function formatBytes(n: number | undefined): string {
    if (!n) return "—";
    const units = ["B", "KB", "MB", "GB", "TB"];
    let i = 0;
    let v = n;
    while (v >= 1024 && i < units.length - 1) {
      v /= 1024;
      i++;
    }
    return `${v.toFixed(1)} ${units[i]}`;
  }

  onMount(() => {
    void load();

    // Subscribe to in-app notifications for export progress.
    // The notify-svc producer publishes `notify.inapp.v1` events
    // with a typed payload when an export transitions; the shape
    // below mirrors what services/notify/internal/inapp.Message
    // serialises.
    const wsURL =
      (typeof window !== "undefined" && (window as { __CHETANA_WS_URL__?: string }).__CHETANA_WS_URL__) ||
      `${location.protocol === "https:" ? "wss" : "ws"}://${location.host}/v1/rt`;

    rt = createRealtimeClient({ url: wsURL, bearer: bearer() });
    rt.start();

    // Poll the connection state into the local reactive variable so
    // the badge re-renders.
    const stateInterval = setInterval(() => {
      if (rt) rtState = rt.state();
    }, 500);

    const unsub = rt.subscribe("notify.inapp.v1", (payload) => {
      const p = payload as {
        title?: string;
        body?: string;
        metadata?: Record<string, string>;
      };
      const jobID = p.metadata?.export_job_id;
      const status = p.metadata?.export_status as ExportJob["status"] | undefined;
      if (!jobID || !status) return;
      jobs = jobs.map((j) =>
        j.id === jobID
          ? {
              ...j,
              status,
              completed_at: status === "succeeded" ? new Date().toISOString() : j.completed_at,
              presigned_url: p.metadata?.presigned_url ?? j.presigned_url,
              bytes_total: p.metadata?.bytes_total ? Number(p.metadata.bytes_total) : j.bytes_total,
              last_error: status === "failed" ? p.body : j.last_error,
            }
          : j,
      );
    });

    return () => {
      clearInterval(stateInterval);
      unsub();
    };
  });

  onDestroy(() => {
    if (rt) rt.close();
  });
</script>

<svelte:head><title>Exports — Chetana</title></svelte:head>

<div class="flex flex-col gap-md max-w-5xl">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-xl font-semibold text-text-primary">Exports</h1>
      <p class="text-sm text-text-muted mt-2xs">
        Updates push from the realtime gateway — no need to refresh.
      </p>
    </div>
    <span
      class="text-xs px-sm py-2xs rounded border border-border text-text-muted"
      data-testid="rt-state-badge"
    >Realtime: {pollStateBadge()}</span>
  </div>

  {#if error}
    <div role="alert" class="px-md py-sm rounded border border-error/40 bg-error/10 text-sm text-error">
      {error}
    </div>
  {/if}

  {#if isLoading}
    <div class="text-text-muted text-sm">Loading…</div>
  {:else if jobs.length === 0}
    <div class="text-text-muted text-sm">
      No export jobs. Trigger one from the <a href="/audit" class="underline">audit log</a>.
    </div>
  {:else}
    <ul class="flex flex-col gap-sm" data-testid="exports-list">
      {#each jobs as j (j.id)}
        <li class="border border-border rounded p-md bg-surface flex justify-between items-start gap-md">
          <div class="flex flex-col gap-2xs text-sm flex-1">
            <div class="flex items-center gap-sm">
              <span class="font-medium text-text-primary">{j.kind}</span>
              <span
                class="text-xs px-sm py-2xs rounded font-mono
                {j.status === 'succeeded' ? 'bg-success/10 text-success'
                : j.status === 'failed' || j.status === 'expired' ? 'bg-error/10 text-error'
                : j.status === 'running' ? 'bg-primary/10 text-primary'
                : 'bg-surface-hover text-text-secondary'}"
                data-testid="export-status-{j.id}"
              >{j.status}</span>
            </div>
            <span class="text-xs text-text-muted font-mono">{j.id}</span>
            <span class="text-xs text-text-muted">
              Enqueued {formatTime(j.enqueued_at)}
              {#if j.completed_at} · Completed {formatTime(j.completed_at)}{/if}
              {#if j.bytes_total} · {formatBytes(j.bytes_total)}{/if}
            </span>
            {#if j.last_error}
              <span class="text-xs text-error">{j.last_error}</span>
            {/if}
          </div>
          {#if j.status === "succeeded" && j.presigned_url}
            <a
              href={j.presigned_url}
              class="px-md py-xs text-xs rounded bg-primary text-on-primary hover:bg-primary/90"
              data-testid="export-download-{j.id}"
            >
              Download
            </a>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</div>
