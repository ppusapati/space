<script context="module" lang="ts">
  export interface ActivityItem {
    id: string;
    type: 'create' | 'update' | 'delete' | 'comment' | 'assign' | 'status' | 'custom';
    actor: {
      name: string;
      avatar?: string;
    };
    action: string;
    target?: string;
    timestamp: Date | string;
    metadata?: Record<string, unknown>;
    icon?: string;
    color?: string;
  }
</script>

<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';

  // Props
  export let activities: ActivityItem[] = [];
  export let showAvatar: boolean = true;
  export let showTimestamp: boolean = true;
  export let maxItems: number = 0; // 0 = show all
  export let groupByDate: boolean = false;

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    itemClick: { activity: ActivityItem };
    loadMore: void;
  }>();

  const typeIcons: Record<string, string> = {
    create: `<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>`,
    update: `<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>`,
    delete: `<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>`,
    comment: `<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/></svg>`,
    assign: `<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/></svg>`,
    status: `<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>`,
    custom: `<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/></svg>`,
  };

  const typeColors: Record<string, string> = {
    create: 'bg-[var(--color-success)]',
    update: 'bg-[var(--color-info)]',
    delete: 'bg-[var(--color-error)]',
    comment: 'bg-[var(--color-interactive-primary)]',
    assign: 'bg-[var(--color-warning)]',
    status: 'bg-[var(--color-success)]',
    custom: 'bg-[var(--color-neutral-500)]',
  };

  function formatTimestamp(date: Date | string): string {
    const d = typeof date === 'string' ? new Date(date) : date;
    const now = new Date();
    const diff = now.getTime() - d.getTime();

    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (minutes < 1) return 'Just now';
    if (minutes < 60) return `${minutes}m ago`;
    if (hours < 24) return `${hours}h ago`;
    if (days < 7) return `${days}d ago`;

    return d.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: d.getFullYear() !== now.getFullYear() ? 'numeric' : undefined,
    });
  }

  function formatDate(date: Date | string): string {
    const d = typeof date === 'string' ? new Date(date) : date;
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const yesterday = new Date(today.getTime() - 86400000);
    const dateStart = new Date(d.getFullYear(), d.getMonth(), d.getDate());

    if (dateStart.getTime() === today.getTime()) return 'Today';
    if (dateStart.getTime() === yesterday.getTime()) return 'Yesterday';

    return d.toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric' });
  }

  function groupActivitiesByDate(items: ActivityItem[]): Map<string, ActivityItem[]> {
    const grouped = new Map<string, ActivityItem[]>();

    for (const item of items) {
      const date = new Date(item.timestamp);
      const dateKey = new Date(date.getFullYear(), date.getMonth(), date.getDate()).toISOString();

      if (!grouped.has(dateKey)) {
        grouped.set(dateKey, []);
      }
      grouped.get(dateKey)!.push(item);
    }

    return grouped;
  }

  $: displayActivities = maxItems > 0 ? activities.slice(0, maxItems) : activities;
  $: groupedActivities = groupByDate ? groupActivitiesByDate(displayActivities) : null;
  $: hasMore = maxItems > 0 && activities.length > maxItems;

  function handleItemClick(activity: ActivityItem) {
    dispatch('itemClick', { activity });
  }

  function handleLoadMore() {
    dispatch('loadMore');
  }
</script>

<div class={cn('activity-feed', className)}>
  {#if groupByDate && groupedActivities}
    {#each [...groupedActivities.entries()] as [dateKey, items]}
      <div class="mb-6">
        <h4 class="text-sm font-medium text-[var(--color-text-tertiary)] mb-3">
          {formatDate(dateKey)}
        </h4>
        <div class="space-y-4">
          {#each items as activity (activity.id)}
            <button
              type="button"
              class="flex gap-3 w-full text-left hover:bg-[var(--color-surface-secondary)] p-2 -mx-2 rounded-lg transition-colors"
              on:click={() => handleItemClick(activity)}
            >
              {#if showAvatar}
                <div class="relative flex-shrink-0">
                  {#if activity.actor.avatar}
                    <img
                      src={activity.actor.avatar}
                      alt={activity.actor.name}
                      class="w-8 h-8 rounded-full"
                    />
                  {:else}
                    <div class="w-8 h-8 rounded-full bg-[var(--color-interactive-primary)] flex items-center justify-center text-white text-sm font-medium">
                      {activity.actor.name.charAt(0).toUpperCase()}
                    </div>
                  {/if}
                  <div
                    class={cn(
                      'absolute -bottom-1 -right-1 w-5 h-5 rounded-full flex items-center justify-center text-white',
                      activity.color || typeColors[activity.type]
                    )}
                  >
                    {@html activity.icon || typeIcons[activity.type]}
                  </div>
                </div>
              {/if}

              <div class="flex-1 min-w-0">
                <p class="text-sm text-[var(--color-text-primary)]">
                  <span class="font-medium">{activity.actor.name}</span>
                  <span class="text-[var(--color-text-secondary)]"> {activity.action}</span>
                  {#if activity.target}
                    <span class="font-medium"> {activity.target}</span>
                  {/if}
                </p>
                {#if showTimestamp}
                  <p class="text-xs text-[var(--color-text-tertiary)] mt-0.5">
                    {formatTimestamp(activity.timestamp)}
                  </p>
                {/if}
              </div>
            </button>
          {/each}
        </div>
      </div>
    {/each}
  {:else}
    <div class="space-y-4">
      {#each displayActivities as activity (activity.id)}
        <button
          type="button"
          class="flex gap-3 w-full text-left hover:bg-[var(--color-surface-secondary)] p-2 -mx-2 rounded-lg transition-colors"
          on:click={() => handleItemClick(activity)}
        >
          {#if showAvatar}
            <div class="relative flex-shrink-0">
              {#if activity.actor.avatar}
                <img
                  src={activity.actor.avatar}
                  alt={activity.actor.name}
                  class="w-8 h-8 rounded-full"
                />
              {:else}
                <div class="w-8 h-8 rounded-full bg-[var(--color-interactive-primary)] flex items-center justify-center text-white text-sm font-medium">
                  {activity.actor.name.charAt(0).toUpperCase()}
                </div>
              {/if}
              <div
                class={cn(
                  'absolute -bottom-1 -right-1 w-5 h-5 rounded-full flex items-center justify-center text-white',
                  activity.color || typeColors[activity.type]
                )}
              >
                {@html activity.icon || typeIcons[activity.type]}
              </div>
            </div>
          {/if}

          <div class="flex-1 min-w-0">
            <p class="text-sm text-[var(--color-text-primary)]">
              <span class="font-medium">{activity.actor.name}</span>
              <span class="text-[var(--color-text-secondary)]"> {activity.action}</span>
              {#if activity.target}
                <span class="font-medium"> {activity.target}</span>
              {/if}
            </p>
            {#if showTimestamp}
              <p class="text-xs text-[var(--color-text-tertiary)] mt-0.5">
                {formatTimestamp(activity.timestamp)}
              </p>
            {/if}
          </div>
        </button>
      {/each}
    </div>
  {/if}

  {#if hasMore}
    <button
      type="button"
      class="w-full mt-4 py-2 text-sm text-[var(--color-interactive-primary)] hover:underline"
      on:click={handleLoadMore}
    >
      Load more activity
    </button>
  {/if}
</div>
