<!--
  Settings → API keys. Long-lived bearer tokens for service-to-
  service automation.

  The chetana convention: the bearer is shown EXACTLY ONCE on
  creation. After that the user only sees the label + scopes +
  last-used metadata.
-->
<script lang="ts">
  import { onMount } from "svelte";
  import * as iam from "@chetana/api-client/iam";

  let keys = $state<iam.ApiKey[]>([]);
  let isLoading = $state(true);
  let error = $state<string | null>(null);
  let label = $state("");
  let scopesText = $state("");
  let ttlDays = $state<number | undefined>(undefined);
  let isCreating = $state(false);
  let revealedBearer = $state<{ id: string; bearer: string; label: string } | null>(null);

  function bearerToken(): string {
    return sessionStorage.getItem("chetana.access_token") ?? "";
  }

  async function load() {
    error = null;
    isLoading = true;
    try {
      keys = await iam.listApiKeys(bearerToken());
    } catch (err) {
      error = (err as Error).message ?? "Failed to load API keys.";
    } finally {
      isLoading = false;
    }
  }

  async function create(e: Event) {
    e.preventDefault();
    error = null;
    isCreating = true;
    try {
      const scopes = scopesText
        .split(/[\s,]+/)
        .map((s) => s.trim())
        .filter((s) => s.length > 0);
      const created = await iam.createApiKey(bearerToken(), label, scopes, ttlDays);
      revealedBearer = { id: created.id, bearer: created.bearer, label: created.label };
      label = "";
      scopesText = "";
      ttlDays = undefined;
      await load();
    } catch (err) {
      error = (err as Error).message ?? "Failed to create API key.";
    } finally {
      isCreating = false;
    }
  }

  async function revoke(id: string) {
    if (!confirm("Revoke this API key? Anything using it will stop working.")) return;
    try {
      await iam.revokeApiKey(bearerToken(), id);
      keys = keys.filter((k) => k.id !== id);
    } catch (err) {
      error = (err as Error).message ?? "Failed to revoke key.";
    }
  }

  function formatTime(iso: string | null): string {
    return iso ? new Date(iso).toLocaleString() : "—";
  }

  function copyBearer() {
    if (revealedBearer) {
      void navigator.clipboard.writeText(revealedBearer.bearer);
    }
  }

  onMount(load);
</script>

<svelte:head><title>API keys — Chetana</title></svelte:head>

<div class="flex flex-col gap-lg max-w-3xl">
  <div>
    <h1 class="text-xl font-semibold text-text-primary">API keys</h1>
    <p class="text-sm text-text-muted mt-2xs">
      Long-lived tokens for service-to-service automation. The bearer is shown
      exactly once on creation — copy it before closing the dialog.
    </p>
  </div>

  {#if error}
    <div role="alert" class="px-md py-sm rounded border border-error/40 bg-error/10 text-sm text-error">
      {error}
    </div>
  {/if}

  {#if revealedBearer}
    <div class="border border-warning/40 bg-warning/5 rounded p-md flex flex-col gap-sm">
      <div class="text-sm font-medium text-warning">
        Save this bearer for {revealedBearer.label} — it will not be shown again.
      </div>
      <code
        class="block px-md py-sm bg-surface border border-border rounded text-xs break-all"
        data-testid="api-key-bearer"
      >{revealedBearer.bearer}</code>
      <div class="flex gap-sm">
        <button type="button" onclick={copyBearer} class="px-md py-xs text-xs rounded border border-border hover:bg-surface-hover">
          Copy
        </button>
        <button
          type="button"
          onclick={() => (revealedBearer = null)}
          class="px-md py-xs text-xs rounded border border-border hover:bg-surface-hover"
        >
          Done
        </button>
      </div>
    </div>
  {/if}

  <form onsubmit={create} class="border border-border rounded p-md bg-surface flex flex-col gap-md">
    <h2 class="text-sm font-semibold text-text-secondary">Create new key</h2>
    <label class="flex flex-col gap-2xs text-sm">
      <span class="text-text-secondary">Label</span>
      <input type="text" required bind:value={label} class="input-field" data-testid="api-key-label" />
    </label>
    <label class="flex flex-col gap-2xs text-sm">
      <span class="text-text-secondary">Scopes (space- or comma-separated)</span>
      <input type="text" bind:value={scopesText} placeholder="audit.read export.read" class="input-field" data-testid="api-key-scopes" />
    </label>
    <label class="flex flex-col gap-2xs text-sm">
      <span class="text-text-secondary">TTL (days, optional)</span>
      <input type="number" min="1" bind:value={ttlDays} class="input-field w-32" data-testid="api-key-ttl" />
    </label>
    <button
      type="submit"
      disabled={isCreating || !label}
      class="self-start px-md py-xs text-sm rounded bg-primary text-on-primary hover:bg-primary/90 disabled:opacity-60"
      data-testid="api-key-create"
    >
      {isCreating ? "Creating…" : "Create key"}
    </button>
  </form>

  <div class="flex flex-col gap-sm">
    <h2 class="text-sm font-semibold text-text-secondary">Existing keys</h2>
    {#if isLoading}
      <div class="text-text-muted text-sm">Loading…</div>
    {:else if keys.length === 0}
      <div class="text-text-muted text-sm">No API keys yet.</div>
    {:else}
      <ul class="flex flex-col gap-sm">
        {#each keys as k (k.id)}
          <li class="border border-border rounded p-md bg-surface flex justify-between items-start gap-md">
            <div class="flex flex-col gap-2xs text-sm">
              <span class="font-medium text-text-primary">{k.label}</span>
              <span class="text-xs text-text-muted">
                {k.scopes.join(", ") || "no scopes"}
              </span>
              <span class="text-xs text-text-muted">
                Created {formatTime(k.created_at)} · Last used {formatTime(k.last_used_at)}
                {#if k.expires_at}· Expires {formatTime(k.expires_at)}{/if}
              </span>
            </div>
            <button
              type="button"
              onclick={() => revoke(k.id)}
              class="px-md py-xs text-xs rounded border border-error/40 text-error hover:bg-error/5"
            >Revoke</button>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</div>

<style>
  .input-field {
    @apply px-md py-sm border border-border rounded bg-surface text-text-primary
           focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20;
  }
</style>
