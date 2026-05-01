<script lang="ts">
  import { authStore, tenantStore } from '@samavāya/stores';
  import { onMount } from 'svelte';

  // Mock data for dashboard widgets
  const metrics = [
    { label: 'Total Revenue', value: '$125,430', change: '+12.5%', changeType: 'positive' as const },
    { label: 'Orders', value: '1,234', change: '+8.2%', changeType: 'positive' as const },
    { label: 'Customers', value: '856', change: '+5.1%', changeType: 'positive' as const },
    { label: 'Pending Invoices', value: '23', change: '-3.4%', changeType: 'negative' as const },
  ];

  const recentActivities = [
    { id: 1, action: 'New order received', time: '5 minutes ago', icon: 'order' },
    { id: 2, action: 'Payment confirmed for INV-2024-001', time: '1 hour ago', icon: 'payment' },
    { id: 3, action: 'New customer registered', time: '2 hours ago', icon: 'customer' },
    { id: 4, action: 'Stock alert: Widget A running low', time: '3 hours ago', icon: 'alert' },
    { id: 5, action: 'Invoice INV-2024-002 sent', time: '4 hours ago', icon: 'invoice' },
  ];

  onMount(() => {
    // Set navigation context
    // navigationStore.setCurrentModule('Dashboard');
  });
</script>

<svelte:head>
  <title>Dashboard - samavāya ERP</title>
</svelte:head>

<div class="dashboard">
  <!-- Welcome Section -->
  <section class="welcome-section">
    <h2 class="welcome-title">
      Welcome back, {$authStore.user?.displayName || 'User'}!
    </h2>
    <p class="welcome-subtitle">
      Here's what's happening with your business today.
    </p>
  </section>

  <!-- Metrics Grid -->
  <section class="metrics-section">
    <div class="metrics-grid">
      {#each metrics as metric}
        <div class="metric-card">
          <div class="metric-header">
            <span class="metric-label">{metric.label}</span>
            <span
              class="metric-change"
              class:positive={metric.changeType === 'positive'}
              class:negative={metric.changeType === 'negative'}
            >
              {metric.change}
            </span>
          </div>
          <div class="metric-value">{metric.value}</div>
        </div>
      {/each}
    </div>
  </section>

  <!-- Content Grid -->
  <div class="content-grid">
    <!-- Recent Activity -->
    <section class="activity-section card">
      <h3 class="section-title">Recent Activity</h3>
      <ul class="activity-list">
        {#each recentActivities as activity}
          <li class="activity-item">
            <div class="activity-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                {#if activity.icon === 'order'}
                  <path d="M6 2L3 6v14a2 2 0 002 2h14a2 2 0 002-2V6l-3-4z" />
                  <line x1="3" y1="6" x2="21" y2="6" />
                  <path d="M16 10a4 4 0 01-8 0" />
                {:else if activity.icon === 'payment'}
                  <rect x="1" y="4" width="22" height="16" rx="2" ry="2" />
                  <line x1="1" y1="10" x2="23" y2="10" />
                {:else if activity.icon === 'customer'}
                  <path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2" />
                  <circle cx="12" cy="7" r="4" />
                {:else if activity.icon === 'alert'}
                  <path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
                  <line x1="12" y1="9" x2="12" y2="13" />
                  <line x1="12" y1="17" x2="12.01" y2="17" />
                {:else}
                  <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" />
                  <polyline points="14 2 14 8 20 8" />
                  <line x1="16" y1="13" x2="8" y2="13" />
                  <line x1="16" y1="17" x2="8" y2="17" />
                  <polyline points="10 9 9 9 8 9" />
                {/if}
              </svg>
            </div>
            <div class="activity-content">
              <p class="activity-action">{activity.action}</p>
              <span class="activity-time">{activity.time}</span>
            </div>
          </li>
        {/each}
      </ul>
      <a href="/activity" class="view-all-link">View all activity</a>
    </section>

    <!-- Quick Actions -->
    <section class="quick-actions-section card">
      <h3 class="section-title">Quick Actions</h3>
      <div class="quick-actions-grid">
        <a href="/sales/orders/new" class="quick-action">
          <svg class="quick-action-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          <span>New Order</span>
        </a>
        <a href="/finance/invoices/new" class="quick-action">
          <svg class="quick-action-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" />
            <polyline points="14 2 14 8 20 8" />
            <line x1="12" y1="18" x2="12" y2="12" />
            <line x1="9" y1="15" x2="15" y2="15" />
          </svg>
          <span>New Invoice</span>
        </a>
        <a href="/masters/customers/new" class="quick-action">
          <svg class="quick-action-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M16 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2" />
            <circle cx="8.5" cy="7" r="4" />
            <line x1="20" y1="8" x2="20" y2="14" />
            <line x1="23" y1="11" x2="17" y2="11" />
          </svg>
          <span>Add Customer</span>
        </a>
        <a href="/inventory/items/new" class="quick-action">
          <svg class="quick-action-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 003 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16z" />
            <line x1="12" y1="8" x2="12" y2="16" />
            <line x1="8" y1="12" x2="16" y2="12" />
          </svg>
          <span>Add Item</span>
        </a>
      </div>
    </section>
  </div>
</div>

<style>
  .dashboard {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-xl);
  }

  /* Welcome Section */
  .welcome-section {
    margin-bottom: var(--spacing-md);
  }

  .welcome-title {
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--color-text);
    margin-bottom: var(--spacing-xs);
  }

  .welcome-subtitle {
    color: var(--color-text-secondary);
  }

  /* Metrics */
  .metrics-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: var(--spacing-lg);
  }

  .metric-card {
    background-color: var(--color-background);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    padding: var(--spacing-lg);
  }

  .metric-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--spacing-sm);
  }

  .metric-label {
    font-size: 0.875rem;
    color: var(--color-text-secondary);
  }

  .metric-change {
    font-size: 0.75rem;
    font-weight: 500;
    padding: var(--spacing-xs) var(--spacing-sm);
    border-radius: var(--radius-full);
  }

  .metric-change.positive {
    background-color: #dcfce7;
    color: #166534;
  }

  .metric-change.negative {
    background-color: #fef2f2;
    color: #991b1b;
  }

  :global([data-theme='dark']) .metric-change.positive {
    background-color: rgba(34, 197, 94, 0.15);
    color: #86efac;
  }

  :global([data-theme='dark']) .metric-change.negative {
    background-color: rgba(239, 68, 68, 0.15);
    color: #fca5a5;
  }

  .metric-value {
    font-size: 1.75rem;
    font-weight: 700;
    color: var(--color-text);
  }

  /* Content Grid */
  .content-grid {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: var(--spacing-lg);
  }

  @media (max-width: 1024px) {
    .content-grid {
      grid-template-columns: 1fr;
    }
  }

  .card {
    background-color: var(--color-background);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    padding: var(--spacing-lg);
  }

  .section-title {
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-text);
    margin-bottom: var(--spacing-lg);
  }

  /* Activity */
  .activity-list {
    list-style: none;
    display: flex;
    flex-direction: column;
    gap: var(--spacing-md);
  }

  .activity-item {
    display: flex;
    gap: var(--spacing-md);
  }

  .activity-icon {
    width: 36px;
    height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--color-surface);
    border-radius: var(--radius-md);
    flex-shrink: 0;
  }

  .activity-icon svg {
    width: 18px;
    height: 18px;
    color: var(--color-text-secondary);
  }

  .activity-content {
    flex: 1;
  }

  .activity-action {
    font-size: 0.875rem;
    color: var(--color-text);
  }

  .activity-time {
    font-size: 0.75rem;
    color: var(--color-text-muted);
  }

  .view-all-link {
    display: block;
    text-align: center;
    margin-top: var(--spacing-lg);
    font-size: 0.875rem;
    color: var(--color-primary);
    font-weight: 500;
  }

  /* Quick Actions */
  .quick-actions-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: var(--spacing-md);
  }

  .quick-action {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--spacing-sm);
    padding: var(--spacing-lg);
    background-color: var(--color-surface);
    border-radius: var(--radius-md);
    text-decoration: none;
    transition: background-color var(--transition-fast);
  }

  .quick-action:hover {
    background-color: var(--color-border);
    text-decoration: none;
  }

  .quick-action-icon {
    width: 24px;
    height: 24px;
    color: var(--color-primary);
  }

  .quick-action span {
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--color-text);
  }
</style>
