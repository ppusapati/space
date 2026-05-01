<script context="module" lang="ts">
  export interface KanbanCard {
    id: string;
    title: string;
    description?: string;
    labels?: { text: string; color: string }[];
    assignee?: { name: string; avatar?: string };
    dueDate?: string | Date;
    priority?: 'low' | 'medium' | 'high' | 'urgent';
    metadata?: Record<string, string | number>;
  }

  export interface KanbanColumn {
    id: string;
    title: string;
    color?: string;
    cards: KanbanCard[];
    limit?: number;
  }
</script>

<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';

  // Props
  export let columns: KanbanColumn[] = [];
  export let cardDraggable: boolean = true;
  export let showCardCount: boolean = true;
  export let showAddCard: boolean = true;
  export let columnMinWidth: string = '280px';
  export let columnMaxWidth: string = '350px';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    cardClick: { card: KanbanCard; column: KanbanColumn };
    cardMove: { card: KanbanCard; fromColumn: KanbanColumn; toColumn: KanbanColumn; newIndex: number };
    addCard: { column: KanbanColumn };
    columnClick: { column: KanbanColumn };
  }>();

  let draggedCard: KanbanCard | null = null;
  let draggedFromColumn: KanbanColumn | null = null;
  let dropTargetColumn: string | null = null;
  let dropTargetIndex: number | null = null;

  const priorityColors = {
    low: 'bg-[var(--color-success)]/20 text-[var(--color-success)]',
    medium: 'bg-[var(--color-warning)]/20 text-[var(--color-warning)]',
    high: 'bg-[var(--color-error)]/20 text-[var(--color-error)]',
    urgent: 'bg-[var(--color-error)] text-white',
  };

  function handleDragStart(event: DragEvent, card: KanbanCard, column: KanbanColumn) {
    if (!cardDraggable) return;

    draggedCard = card;
    draggedFromColumn = column;

    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = 'move';
      event.dataTransfer.setData('text/plain', card.id);
    }

    const target = event.target as HTMLElement;
    target.classList.add('opacity-50');
  }

  function handleDragEnd(event: DragEvent) {
    const target = event.target as HTMLElement;
    target.classList.remove('opacity-50');

    draggedCard = null;
    draggedFromColumn = null;
    dropTargetColumn = null;
    dropTargetIndex = null;
  }

  function handleDragOver(event: DragEvent, columnId: string, cardIndex: number) {
    event.preventDefault();
    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = 'move';
    }
    dropTargetColumn = columnId;
    dropTargetIndex = cardIndex;
  }

  function handleDragLeave() {
    dropTargetColumn = null;
    dropTargetIndex = null;
  }

  function handleDrop(event: DragEvent, targetColumn: KanbanColumn, targetIndex: number) {
    event.preventDefault();

    if (!draggedCard || !draggedFromColumn) return;

    // Remove from source column
    const sourceColumn = columns.find(c => c.id === draggedFromColumn!.id);
    if (sourceColumn) {
      sourceColumn.cards = sourceColumn.cards.filter(c => c.id !== draggedCard!.id);
    }

    // Add to target column
    const targetCol = columns.find(c => c.id === targetColumn.id);
    if (targetCol) {
      targetCol.cards.splice(targetIndex, 0, draggedCard);
    }

    // Trigger reactivity
    columns = [...columns];

    dispatch('cardMove', {
      card: draggedCard,
      fromColumn: draggedFromColumn,
      toColumn: targetColumn,
      newIndex: targetIndex,
    });

    dropTargetColumn = null;
    dropTargetIndex = null;
  }

  function handleCardClick(card: KanbanCard, column: KanbanColumn) {
    dispatch('cardClick', { card, column });
  }

  function handleAddCard(column: KanbanColumn) {
    dispatch('addCard', { column });
  }

  function handleColumnClick(column: KanbanColumn) {
    dispatch('columnClick', { column });
  }

  function formatDate(date: string | Date): string {
    const d = typeof date === 'string' ? new Date(date) : date;
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }

  function isOverLimit(column: KanbanColumn): boolean {
    return column.limit !== undefined && column.cards.length > column.limit;
  }
</script>

<div
  class={cn(
    'kanban-board flex gap-4 overflow-x-auto pb-4',
    className
  )}
>
  {#each columns as column (column.id)}
    <div
      class={cn(
        'kanban-column flex flex-col rounded-lg',
        'bg-[var(--color-surface-secondary)]',
        isOverLimit(column) && 'ring-2 ring-[var(--color-error)]'
      )}
      style="min-width: {columnMinWidth}; max-width: {columnMaxWidth}; width: {columnMinWidth};"
    >
      <!-- Column Header -->
      <button
        type="button"
        class={cn(
          'flex items-center justify-between p-3 rounded-t-lg',
          'hover:bg-[var(--color-surface-tertiary)] transition-colors',
          'focus:outline-none focus:ring-2 focus:ring-inset focus:ring-[var(--color-interactive-primary)]'
        )}
        style={column.color ? `border-top: 3px solid ${column.color}` : ''}
        on:click={() => handleColumnClick(column)}
      >
        <div class="flex items-center gap-2">
          <h3 class="font-semibold text-[var(--color-text-primary)]">
            {column.title}
          </h3>
          {#if showCardCount}
            <span
              class={cn(
                'text-xs px-2 py-0.5 rounded-full',
                isOverLimit(column)
                  ? 'bg-[var(--color-error)] text-white'
                  : 'bg-[var(--color-surface-tertiary)] text-[var(--color-text-secondary)]'
              )}
            >
              {column.cards.length}{column.limit ? `/${column.limit}` : ''}
            </span>
          {/if}
        </div>
      </button>

      <!-- Cards Container -->
      <div
        class="flex-1 p-2 space-y-2 overflow-y-auto max-h-[calc(100vh-200px)]"
        on:dragover={(e) => handleDragOver(e, column.id, column.cards.length)}
        on:dragleave={handleDragLeave}
        on:drop={(e) => handleDrop(e, column, column.cards.length)}
      >
        {#each column.cards as card, cardIndex (card.id)}
          <div
            class={cn(
              'kanban-card p-3 rounded-lg cursor-pointer',
              'bg-[var(--color-surface-primary)]',
              'border border-[var(--color-border-primary)]',
              'hover:border-[var(--color-interactive-primary)]',
              'hover:shadow-md transition-all',
              dropTargetColumn === column.id && dropTargetIndex === cardIndex && 'border-t-2 border-t-[var(--color-interactive-primary)]'
            )}
            draggable={cardDraggable}
            on:dragstart={(e) => handleDragStart(e, card, column)}
            on:dragend={handleDragEnd}
            on:dragover={(e) => handleDragOver(e, column.id, cardIndex)}
            on:click={() => handleCardClick(card, column)}
            role="button"
            tabindex="0"
            on:keydown={(e) => e.key === 'Enter' && handleCardClick(card, column)}
          >
            <!-- Labels -->
            {#if card.labels && card.labels.length > 0}
              <div class="flex flex-wrap gap-1 mb-2">
                {#each card.labels as label}
                  <span
                    class="text-xs px-2 py-0.5 rounded text-white"
                    style="background-color: {label.color}"
                  >
                    {label.text}
                  </span>
                {/each}
              </div>
            {/if}

            <!-- Title -->
            <h4 class="font-medium text-[var(--color-text-primary)] text-sm">
              {card.title}
            </h4>

            <!-- Description -->
            {#if card.description}
              <p class="text-xs text-[var(--color-text-secondary)] mt-1 line-clamp-2">
                {card.description}
              </p>
            {/if}

            <!-- Footer -->
            <div class="flex items-center justify-between mt-3 pt-2 border-t border-[var(--color-border-secondary)]">
              <div class="flex items-center gap-2">
                <!-- Priority -->
                {#if card.priority}
                  <span class={cn('text-xs px-1.5 py-0.5 rounded font-medium', priorityColors[card.priority])}>
                    {card.priority}
                  </span>
                {/if}

                <!-- Due Date -->
                {#if card.dueDate}
                  {@const isOverdue = new Date(card.dueDate) < new Date()}
                  <span
                    class={cn(
                      'text-xs flex items-center gap-1',
                      isOverdue ? 'text-[var(--color-error)]' : 'text-[var(--color-text-tertiary)]'
                    )}
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                    {formatDate(card.dueDate)}
                  </span>
                {/if}
              </div>

              <!-- Assignee -->
              {#if card.assignee}
                <div class="flex items-center">
                  {#if card.assignee.avatar}
                    <img
                      src={card.assignee.avatar}
                      alt={card.assignee.name}
                      class="w-6 h-6 rounded-full"
                    />
                  {:else}
                    <div class="w-6 h-6 rounded-full bg-[var(--color-interactive-primary)] flex items-center justify-center text-white text-xs font-medium">
                      {card.assignee.name.charAt(0).toUpperCase()}
                    </div>
                  {/if}
                </div>
              {/if}
            </div>
          </div>
        {/each}

        <!-- Drop zone indicator -->
        {#if dropTargetColumn === column.id && dropTargetIndex === column.cards.length}
          <div class="h-16 border-2 border-dashed border-[var(--color-interactive-primary)] rounded-lg bg-[var(--color-interactive-primary)]/10" />
        {/if}
      </div>

      <!-- Add Card Button -->
      {#if showAddCard}
        <button
          type="button"
          class={cn(
            'flex items-center justify-center gap-2 p-3',
            'text-[var(--color-text-secondary)]',
            'hover:text-[var(--color-text-primary)]',
            'hover:bg-[var(--color-surface-tertiary)]',
            'transition-colors rounded-b-lg',
            'focus:outline-none focus:ring-2 focus:ring-inset focus:ring-[var(--color-interactive-primary)]'
          )}
          on:click={() => handleAddCard(column)}
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          <span class="text-sm">Add card</span>
        </button>
      {/if}
    </div>
  {/each}
</div>

<style>
  .line-clamp-2 {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
</style>
