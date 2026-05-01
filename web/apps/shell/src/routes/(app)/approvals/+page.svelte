<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '@samavāya/stores';
  import {
    getApprovalService,
    ApprovalStatus,
    type ApprovalRequest,
    type ApprovalStage,
    type ListPendingApprovalsResponse,
  } from '@samavāya/api';

  let pending = $state<ApprovalRequest[]>([]);
  let isLoading = $state(true);
  let loadError = $state<string | null>(null);
  let actionError = $state<string | null>(null);
  let busyStageId = $state<string | null>(null);
  let filter = $state<'all' | 'mine' | 'escalated'>('mine');

  const userId = $derived($authStore.user?.id ?? '');
  const userName = $derived(
    $authStore.user?.displayName || $authStore.user?.firstName || 'You',
  );

  const visible: ApprovalRequest[] = $derived.by(() => {
    if (filter === 'all') return pending;
    if (filter === 'escalated') {
      return pending.filter((r: ApprovalRequest) => r.overallStatus === ApprovalStatus.ESCALATED);
    }
    // "mine" — requests where I own at least one PENDING stage
    return pending.filter((r: ApprovalRequest) =>
      r.stages.some(
        (s: ApprovalStage) => s.approverId === userId && s.status === ApprovalStatus.PENDING_APPROVAL,
      ),
    );
  });

  async function load() {
    if (!userId) {
      loadError = 'Sign in to see pending approvals.';
      isLoading = false;
      return;
    }
    isLoading = true;
    loadError = null;
    try {
      const svc = getApprovalService();
      const response = (await svc.listPendingApprovals({
        userId,
        role: '',
        entityType: '',
        status: ApprovalStatus.PENDING_APPROVAL,
      } as never)) as unknown as ListPendingApprovalsResponse;
      pending = response.pendingRequests ?? [];
    } catch (err) {
      loadError = err instanceof Error ? err.message : 'Failed to load approvals';
    } finally {
      isLoading = false;
    }
  }

  onMount(() => {
    void load();
  });

  async function approve(req: ApprovalRequest, stage: ApprovalStage) {
    const comments = prompt(`Approve ${req.entityType} ${req.entityId}?\nOptional comment:`);
    if (comments === null) return;
    busyStageId = stage.stageId;
    actionError = null;
    try {
      const svc = getApprovalService();
      await svc.approveStage({
        requestId: req.requestId,
        stageId: stage.stageId,
        approverId: userId,
        comments,
      } as never);
      await load();
    } catch (err) {
      actionError = err instanceof Error ? err.message : 'Approval failed';
    } finally {
      busyStageId = null;
    }
  }

  async function reject(req: ApprovalRequest, stage: ApprovalStage) {
    const comments = prompt(
      `Reject ${req.entityType} ${req.entityId}?\nPlease provide a reason:`,
    );
    if (!comments) return;
    busyStageId = stage.stageId;
    actionError = null;
    try {
      const svc = getApprovalService();
      await svc.rejectStage({
        requestId: req.requestId,
        stageId: stage.stageId,
        approverId: userId,
        comments,
      } as never);
      await load();
    } catch (err) {
      actionError = err instanceof Error ? err.message : 'Rejection failed';
    } finally {
      busyStageId = null;
    }
  }

  function myStage(req: ApprovalRequest): ApprovalStage | null {
    return (
      req.stages.find(
        (s: ApprovalStage) =>
          s.approverId === userId && s.status === ApprovalStatus.PENDING_APPROVAL,
      ) ?? null
    );
  }

  function formatDate(ts: unknown): string {
    if (!ts) return '—';
    if (typeof ts === 'object' && ts !== null && 'seconds' in ts) {
      const seconds = Number((ts as { seconds: bigint | number }).seconds);
      return new Date(seconds * 1000).toLocaleString();
    }
    return String(ts);
  }

  function statusLabel(s: ApprovalStatus): string {
    switch (s) {
      case ApprovalStatus.PENDING_APPROVAL:
        return 'Pending';
      case ApprovalStatus.APPROVED:
        return 'Approved';
      case ApprovalStatus.REJECTED:
        return 'Rejected';
      case ApprovalStatus.ESCALATED:
        return 'Escalated';
      case ApprovalStatus.REMINDER:
        return 'Reminder';
      default:
        return 'Unknown';
    }
  }
</script>

<svelte:head>
  <title>Approvals · Samavāya</title>
</svelte:head>

<div class="approvals">
  <header>
    <div class="header-row">
      <div>
        <h1>Approvals</h1>
        <p class="lead">Pending items awaiting {userName}'s review.</p>
      </div>
      <!--
        Phase 6a link-in (BI roadmap task A.6, 2026-04-19).
        Cross-app deep link to the seeded `form_operations` dashboard in the
        BI app, which publishes submission throughput, approval-cycle time,
        and SLA compliance for forms flowing through this approvals queue.
        The BI app is mounted at `/bi` (apps/bi/svelte.config.js
        paths.base) so this is an absolute URL from the shell's perspective.
      -->
      <a class="ops-link" href="/bi/dashboards/forms">
        <span class="ops-link-label">Forms Operations</span>
        <span class="ops-link-sub">throughput, cycle time, SLA</span>
      </a>
    </div>
  </header>

  <div class="toolbar">
    <div class="filters" role="tablist">
      <button
        class="filter"
        class:active={filter === 'mine'}
        onclick={() => (filter = 'mine')}
      >Mine ({pending.filter((r) => myStage(r)).length})</button>
      <button
        class="filter"
        class:active={filter === 'all'}
        onclick={() => (filter = 'all')}
      >All ({pending.length})</button>
      <button
        class="filter"
        class:active={filter === 'escalated'}
        onclick={() => (filter = 'escalated')}
      >Escalated</button>
    </div>
    <button class="refresh" onclick={load} disabled={isLoading}>
      {isLoading ? 'Refreshing…' : 'Refresh'}
    </button>
  </div>

  {#if actionError}
    <div class="banner error">{actionError}</div>
  {/if}

  {#if isLoading}
    <div class="status">Loading approvals…</div>
  {:else if loadError}
    <div class="status error">
      <strong>Could not load approvals.</strong>
      <p>{loadError}</p>
      <p class="muted">
        Check that the ApprovalService is running and that you're signed in.
      </p>
    </div>
  {:else if visible.length === 0}
    <div class="status muted">
      {#if filter === 'mine'}
        Nothing waiting on you. 🎉
      {:else if filter === 'escalated'}
        No escalated approvals.
      {:else}
        Nothing pending across the system.
      {/if}
    </div>
  {:else}
    <ul class="list">
      {#each visible as req (req.requestId)}
        {@const mine = myStage(req)}
        <li class="card">
          <div class="card-head">
            <div>
              <h3>
                <code>{req.entityType}</code>
                / <span class="entity-id">{req.entityId}</span>
              </h3>
              <p class="meta">
                Submitted by <strong>{req.submittedBy}</strong> on
                {formatDate(req.submittedAt)}
              </p>
            </div>
            <span class="overall-status status-{statusLabel(req.overallStatus).toLowerCase()}">
              {statusLabel(req.overallStatus)}
            </span>
          </div>

          <ol class="stages">
            {#each req.stages as stage (stage.stageId)}
              <li class="stage" class:active={stage.stageId === mine?.stageId}>
                <div class="stage-main">
                  <strong>{stage.levelName || stage.department || `Stage ${stage.stageId}`}</strong>
                  <span class="stage-approver">{stage.approverName || stage.approverId}</span>
                </div>
                <div class="stage-meta">
                  <span class="pill status-{statusLabel(stage.status).toLowerCase()}">
                    {statusLabel(stage.status)}
                  </span>
                  {#if stage.actionDate}
                    <span class="stage-date">{formatDate(stage.actionDate)}</span>
                  {/if}
                </div>
                {#if stage.comments}
                  <p class="comments">"{stage.comments}"</p>
                {/if}
              </li>
            {/each}
          </ol>

          {#if mine}
            <div class="actions">
              <button
                class="btn reject"
                onclick={() => reject(req, mine)}
                disabled={busyStageId === mine.stageId}
              >
                Reject
              </button>
              <button
                class="btn approve"
                onclick={() => approve(req, mine)}
                disabled={busyStageId === mine.stageId}
              >
                {busyStageId === mine.stageId ? 'Working…' : 'Approve'}
              </button>
            </div>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .approvals {
    padding: 1.5rem;
    max-width: 1000px;
    margin: 0 auto;
  }

  header h1 {
    margin: 0 0 0.25rem;
    font-size: 1.75rem;
  }

  .lead {
    margin: 0 0 1.25rem;
    color: var(--color-text-muted, #666);
  }

  .header-row {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .ops-link {
    display: inline-flex;
    flex-direction: column;
    gap: 0.125rem;
    padding: 0.625rem 0.875rem;
    background: var(--color-bg-subtle, #f1f1f1);
    border: 1px solid var(--color-border, #ddd);
    border-radius: 8px;
    text-decoration: none;
    color: inherit;
    font-size: 0.875rem;
    transition: background 0.15s, border-color 0.15s;
  }

  .ops-link:hover {
    background: var(--color-bg-hover, #e9ecef);
    border-color: var(--color-accent, #2563eb);
  }

  .ops-link-label {
    font-weight: 600;
    color: var(--color-text, #222);
  }

  .ops-link-sub {
    font-size: 0.7rem;
    color: var(--color-text-muted, #666);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
    gap: 1rem;
  }

  .filters {
    display: flex;
    gap: 0.25rem;
    background: var(--color-bg-subtle, #f1f1f1);
    padding: 0.25rem;
    border-radius: 8px;
  }

  .filter {
    padding: 0.375rem 0.875rem;
    border: 0;
    background: transparent;
    border-radius: 6px;
    font-size: 0.875rem;
    cursor: pointer;
    color: var(--color-text-muted, #555);
  }

  .filter.active {
    background: var(--color-bg-surface, #fff);
    color: var(--color-text, #222);
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  }

  .refresh {
    padding: 0.375rem 0.75rem;
    border: 1px solid var(--color-border, #ddd);
    background: var(--color-bg-surface, #fff);
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.875rem;
  }

  .refresh:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .banner {
    padding: 0.75rem 1rem;
    border-radius: 6px;
    margin-bottom: 1rem;
  }

  .banner.error {
    background: var(--color-danger-bg, #fee);
    color: var(--color-danger, #900);
  }

  .list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .card {
    background: var(--color-bg-surface, #fff);
    border: 1px solid var(--color-border, #e5e5e5);
    border-radius: 8px;
    padding: 1rem 1.25rem;
  }

  .card-head {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1rem;
    margin-bottom: 0.75rem;
  }

  .card-head h3 {
    margin: 0 0 0.25rem;
    font-size: 1rem;
  }

  .entity-id {
    font-family: var(--font-mono, monospace);
    color: var(--color-text-muted, #444);
  }

  .meta {
    margin: 0;
    font-size: 0.8rem;
    color: var(--color-text-muted, #888);
  }

  code {
    font-family: var(--font-mono, monospace);
    background: var(--color-bg-subtle, #f0f0f0);
    padding: 0.05rem 0.3rem;
    border-radius: 3px;
    font-size: 0.85em;
  }

  .overall-status,
  .pill {
    display: inline-block;
    padding: 0.125rem 0.5rem;
    border-radius: 999px;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    background: var(--color-bg-subtle, #eee);
    color: var(--color-text-muted, #555);
    white-space: nowrap;
  }

  .status-pending {
    background: var(--color-warning-bg, #fff4e5);
    color: var(--color-warning, #a86100);
  }
  .status-approved {
    background: var(--color-success-bg, #e7f7ec);
    color: var(--color-success, #1a7a3a);
  }
  .status-rejected {
    background: var(--color-danger-bg, #fdecec);
    color: var(--color-danger, #a60000);
  }
  .status-escalated {
    background: #ede7f6;
    color: #5b3ea5;
  }

  .stages {
    list-style: none;
    padding: 0;
    margin: 0 0 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .stage {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 0.5rem;
    padding: 0.5rem 0.75rem;
    background: var(--color-bg-subtle, #fafafa);
    border: 1px solid transparent;
    border-radius: 6px;
    font-size: 0.875rem;
  }

  .stage.active {
    border-color: var(--color-accent, #2563eb);
    background: var(--color-bg-hover, #f3f7ff);
  }

  .stage-main {
    display: flex;
    flex-direction: column;
  }

  .stage-approver {
    font-size: 0.75rem;
    color: var(--color-text-muted, #777);
  }

  .stage-meta {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.2rem;
  }

  .stage-date {
    font-size: 0.7rem;
    color: var(--color-text-muted, #888);
  }

  .comments {
    grid-column: 1 / -1;
    margin: 0;
    padding: 0.4rem 0;
    font-style: italic;
    color: var(--color-text-muted, #666);
    font-size: 0.85rem;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
  }

  .btn {
    padding: 0.4rem 1rem;
    border-radius: 6px;
    border: 1px solid transparent;
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
  }

  .btn.approve {
    background: var(--color-success, #1a7a3a);
    color: white;
    border-color: var(--color-success, #1a7a3a);
  }

  .btn.approve:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .btn.reject {
    background: transparent;
    color: var(--color-danger, #a60000);
    border-color: var(--color-danger, #a60000);
  }

  .btn.reject:disabled {
    opacity: 0.6;
    cursor: not-allowed;
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
