<!--
  Settings → Active sessions. REQ-FUNC-PLT-IAM-009.

  Lists every active session for the authenticated user. The
  current session is flagged so the user can't accidentally
  revoke their own connection.
-->
<script lang="ts">
  import { onMount } from "svelte";
  import * as iam from "@chetana/api-client/iam";

  let sessions = $state<iam.Session[]>([]);
  let isLoading = $state(true);
  let error = $state<string | null>(null);
  let revokingId = $state<string | null>(null);

  function bearer(): string {
    return sessionStorage.getItem("chetana.access_token") ?? "";
  }

  async function load() {
    error = null;
    isLoading = true;
    try {
      sessions = await iam.listSessions(bearer());
    } catch (err) {
      error = (err as Error).message ?? "Failed to load sessions.";
    } finally {
      isLoading = false;
    }
  }

  async function revoke(id: string) {
    if (!confirm("Revoke this session? Any device using it will be signed out.")) return;
    revokingId = id;
    try {
      await iam.revokeSession(bearer(), id);
      sessions = sessions.filter((s) => s.session_id !== id);
    } catch (err) {
      error = (err as Error).message ?? "Failed to revoke session.";
    } finally {
      revokingId = null;
    }
  }

  function formatTime(iso: string): string {
    return new Date(iso).toLocaleString();
  }

  onMount(load);
</script>

<svelte:head><title>Active sessions — Chetana</title></svelte:head>

<div class="flex flex-col gap-md max-w-3xl">
  <div>
    <h1 class="text-xl font-semibold text-text-primary">Active sessions</h1>
    <p class="text-sm text-text-muted mt-2xs">
      You can sign out of devices you no longer use. The session you're using
      right now is marked "Current".
    </p>
  </div>

  {#if error}
    <div role="alert" class="px-md py-sm rounded border border-error/40 bg-error/10 text-sm text-error">
      {error}
    </div>
  {/if}

  {#if isLoading}
    <div class="text-text-muted text-sm" data-testid="sessions-loading">Loading…</div>
  {:else if sessions.length === 0}
    <div class="text-text-muted text-sm">No active sessions.</div>
  {:else}
    <ul class="flex flex-col gap-sm" data-testid="sessions-list">
      {#each sessions as s (s.session_id)}
        <li class="border border-border rounded p-md bg-surface flex justify-between items-start gap-md">
          <div class="flex flex-col gap-2xs text-sm">
            <div class="flex items-center gap-sm">
              <span class="font-medium text-text-primary">{s.client_ip || "(unknown ip)"}</span>
              {#if s.current}
                <span class="text-xs px-sm py-2xs rounded bg-primary/10 text-primary">Current</span>
              {/if}
            </div>
            <div class="text-xs text-text-secondary">{s.user_agent || "(unknown agent)"}</div>
            <div class="text-xs text-text-muted">
              Signed in {formatTime(s.issued_at)} · Last seen {formatTime(s.last_seen_at)}
            </div>
          </div>
          {#if !s.current}
            <button
              type="button"
              disabled={revokingId === s.session_id}
              onclick={() => revoke(s.session_id)}
              class="px-md py-xs text-xs rounded border border-error/40 text-error hover:bg-error/5 disabled:opacity-60"
              data-testid="revoke-{s.session_id}"
            >
              {revokingId === s.session_id ? "Revoking…" : "Revoke"}
            </button>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</div>
