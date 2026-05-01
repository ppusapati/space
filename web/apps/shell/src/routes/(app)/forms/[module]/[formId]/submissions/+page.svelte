<script lang="ts">
  import { onMount, untrack } from 'svelte';
  import { page } from '$app/stores';
  import { getAuditReadService, type AuditLog } from '@samavāya/api';

  const moduleId = $derived($page.params.module ?? '');
  const formId = $derived($page.params.formId ?? '');

  let logs = $state<AuditLog[]>([]);
  let isLoading = $state(true);
  let loadError = $state<string | null>(null);

  async function loadHistory(form: string) {
    if (!form) return;
    isLoading = true;
    loadError = null;
    try {
      const audit = getAuditReadService();
      // The backend audit service indexes by (entity_type, entity_id). Form
      // submissions are logged with entity_type=form_submission and entity_id
      // equal to the form_id; server returns chronological logs.
      const response = await audit.getEntityAuditLogs({
        entityType: 'form_submission',
        entityId: form,
      } as never);
      logs = ((response as unknown as { logs?: AuditLog[] }).logs ?? []) as AuditLog[];
    } catch (err) {
      loadError = err instanceof Error ? err.message : 'Failed to load submission history';
    } finally {
      isLoading = false;
    }
  }

  $effect(() => {
    const id = formId;
    untrack(() => void loadHistory(id));
  });

  function formatDate(ts: unknown): string {
    if (!ts) return '—';
    // AuditLog timestamps are google.protobuf.Timestamp — generated TS gives
    // `{ seconds: bigint; nanos: number }` or a Date depending on the binding.
    if (ts instanceof Date) return ts.toLocaleString();
    if (typeof ts === 'string') return new Date(ts).toLocaleString();
    if (typeof ts === 'object' && ts !== null && 'seconds' in ts) {
      const seconds = Number((ts as { seconds: bigint | number }).seconds);
      return new Date(seconds * 1000).toLocaleString();
    }
    return String(ts);
  }

  function formatActor(log: AuditLog): string {
    // AuditLog may encode the actor via several fields depending on the writer.
    const anyLog = log as unknown as Record<string, unknown>;
    return (
      (anyLog.actor_name as string | undefined) ??
      (anyLog.actorName as string | undefined) ??
      (anyLog.user_name as string | undefined) ??
      (anyLog.userName as string | undefined) ??
      (anyLog.actor_id as string | undefined) ??
      (anyLog.actorId as string | undefined) ??
      'system'
    );
  }

  function formatOperation(log: AuditLog): string {
    const anyLog = log as unknown as Record<string, unknown>;
    return (
      (anyLog.operation as string | undefined) ??
      (anyLog.action as string | undefined) ??
      (anyLog.event_type as string | undefined) ??
      (anyLog.eventType as string | undefined) ??
      '—'
    );
  }
</script>

<svelte:head>
  <title>{formId} submissions · Samavāya</title>
</svelte:head>

<div class="submissions">
  <nav aria-label="Breadcrumb" class="crumbs">
    <a href="/forms">Forms</a>
    <span aria-hidden="true">›</span>
    <a href={`/forms/${moduleId}`}>{moduleId}</a>
    <span aria-hidden="true">›</span>
    <a href={`/forms/${moduleId}/${formId}`}>{formId}</a>
    <span aria-hidden="true">›</span>
    <span class="current">Submissions</span>
  </nav>

  <header>
    <h1>Submission history</h1>
    <p class="lead">{formId}</p>
  </header>

  {#if isLoading}
    <div class="status">Loading history…</div>
  {:else if loadError}
    <div class="status error">
      <strong>Could not load history.</strong>
      <p>{loadError}</p>
      <p class="muted">
        The audit service indexes submissions once the backend records them.
        If this is a new form that has never been submitted, this list will be empty.
      </p>
    </div>
  {:else if logs.length === 0}
    <div class="status muted">
      No submissions recorded yet. Once the form is submitted, each entry will appear here.
    </div>
  {:else}
    <table class="history">
      <thead>
        <tr>
          <th>When</th>
          <th>Who</th>
          <th>Operation</th>
          <th>Result</th>
        </tr>
      </thead>
      <tbody>
        {#each logs as log}
          {@const anyLog = log as unknown as Record<string, unknown>}
          <tr>
            <td>{formatDate(anyLog.timestamp ?? anyLog.created_at ?? anyLog.createdAt)}</td>
            <td>{formatActor(log)}</td>
            <td><code>{formatOperation(log)}</code></td>
            <td>{(anyLog.status ?? anyLog.result ?? '—') as string}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<style>
  .submissions {
    padding: 1.5rem;
    max-width: 1100px;
    margin: 0 auto;
  }

  .crumbs {
    display: flex;
    gap: 0.5rem;
    align-items: center;
    font-size: 0.9rem;
    color: var(--color-text-muted, #666);
    margin-bottom: 0.75rem;
  }

  .crumbs a {
    color: inherit;
    text-decoration: none;
  }

  .crumbs a:hover {
    text-decoration: underline;
  }

  .current {
    color: var(--color-text, #222);
  }

  header h1 {
    margin: 0 0 0.25rem;
    font-size: 1.75rem;
  }

  .lead {
    margin: 0 0 1.25rem;
    color: var(--color-text-muted, #666);
    font-family: var(--font-mono, monospace);
  }

  .history {
    width: 100%;
    border-collapse: collapse;
    background: var(--color-bg-surface, #fff);
    border: 1px solid var(--color-border, #e5e5e5);
    border-radius: 6px;
    overflow: hidden;
  }

  .history th,
  .history td {
    padding: 0.6rem 0.75rem;
    text-align: left;
    border-bottom: 1px solid var(--color-border, #eee);
    font-size: 0.9rem;
  }

  .history thead th {
    background: var(--color-bg-subtle, #f7f7f7);
    font-weight: 600;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--color-text-muted, #555);
  }

  code {
    font-family: var(--font-mono, monospace);
    background: var(--color-bg-subtle, #f3f3f3);
    padding: 0 0.25rem;
    border-radius: 3px;
    font-size: 0.85em;
  }

  .status {
    padding: 2rem;
    text-align: center;
  }

  .status.error {
    color: var(--color-danger, #c00);
  }

  .muted {
    color: var(--color-text-muted, #888);
  }
</style>
