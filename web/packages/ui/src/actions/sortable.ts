/**
 * Sortable Svelte Action
 * Makes a list sortable via drag-and-drop
 */

import type { SortableOptions, SortResult, DropPosition, DragData } from './dragdrop.types';
import {
  sortableClasses,
  setDragState,
  clearDragState,
  getDragState,
  serializeDragData,
  deserializeDragData,
  calculateDropPosition,
  reorderItems,
  DRAG_DATA_MIME,
} from './dragdrop.types';

export interface SortableReturn {
  update: (options: SortableOptions) => void;
  destroy: () => void;
}

/**
 * Svelte action for making lists sortable
 *
 * @example
 * ```svelte
 * <ul
 *   use:sortable={{
 *     id: 'my-list',
 *     type: 'list-item',
 *     itemSelector: 'li',
 *     itemIdAttribute: 'data-id',
 *     onSortEnd: (result) => {
 *       items = reorderItems(items, result.sourceIndex, result.destinationIndex);
 *     }
 *   }}
 * >
 *   {#each items as item}
 *     <li data-id={item.id}>{item.name}</li>
 *   {/each}
 * </ul>
 * ```
 */
export function sortable<T = unknown>(
  node: HTMLElement,
  options: SortableOptions<T>
): SortableReturn {
  let opts: SortableOptions<T> = {
    itemSelector: 'li',
    itemIdAttribute: 'data-id',
    acceptFromOther: true,
    animationDuration: 200,
    ...options,
  };

  let draggedItem: HTMLElement | null = null;
  let draggedItemId: string | null = null;
  let draggedItemIndex: number = -1;
  let dropIndicator: HTMLElement | null = null;
  let currentDropTarget: HTMLElement | null = null;
  let currentDropPosition: DropPosition = 'after';

  function getItems(): HTMLElement[] {
    return Array.from(node.querySelectorAll(opts.itemSelector!));
  }

  function getItemId(item: HTMLElement): string {
    return item.getAttribute(opts.itemIdAttribute!) || '';
  }

  function getItemIndex(itemId: string): number {
    const items = getItems();
    return items.findIndex((item) => getItemId(item) === itemId);
  }

  function createDropIndicator(): HTMLElement {
    const indicator = document.createElement('div');
    indicator.className = 'sortable-drop-indicator';
    indicator.style.cssText = `
      position: absolute;
      left: 0;
      right: 0;
      height: 2px;
      background-color: var(--color-brand-primary-500, #0ea5e9);
      z-index: 100;
      pointer-events: none;
      transition: top 0.1s ease-out;
    `;
    return indicator;
  }

  function showDropIndicator(target: HTMLElement, position: DropPosition) {
    if (!dropIndicator) {
      dropIndicator = createDropIndicator();
      node.style.position = 'relative';
      node.appendChild(dropIndicator);
    }

    const targetRect = target.getBoundingClientRect();
    const parentRect = node.getBoundingClientRect();

    if (position === 'before') {
      dropIndicator.style.top = `${targetRect.top - parentRect.top}px`;
    } else {
      dropIndicator.style.top = `${targetRect.bottom - parentRect.top}px`;
    }
  }

  function hideDropIndicator() {
    if (dropIndicator) {
      dropIndicator.remove();
      dropIndicator = null;
    }
  }

  function handleDragStart(e: DragEvent) {
    if (opts.enabled === false) return;

    const target = e.target as HTMLElement;
    const item = target.closest(opts.itemSelector!) as HTMLElement;

    if (!item || !node.contains(item)) return;

    // Check handle
    if (opts.handle) {
      const handle = item.querySelector(opts.handle);
      if (!handle || !handle.contains(e.target as Node)) {
        e.preventDefault();
        return;
      }
    }

    draggedItem = item;
    draggedItemId = getItemId(item);
    draggedItemIndex = getItemIndex(draggedItemId);

    const dragData: DragData<T> = {
      id: draggedItemId,
      type: opts.type,
      sourceId: opts.id,
      sourceIndex: draggedItemIndex,
      data: item as unknown as T,
    };

    // Set drag data
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move';
      e.dataTransfer.setData('text/plain', draggedItemId);
      e.dataTransfer.setData(DRAG_DATA_MIME, serializeDragData(dragData));
    }

    // Update global state
    setDragState({
      isDragging: true,
      data: dragData,
      sourceElement: item,
    });

    // Apply dragging class
    const draggingClass = opts.draggingClass || sortableClasses.itemDragging;
    item.classList.add(...draggingClass.split(' '));

    opts.onSortStart?.(draggedItemId, draggedItemIndex);
  }

  function handleDragEnd(e: DragEvent) {
    if (draggedItem) {
      const draggingClass = opts.draggingClass || sortableClasses.itemDragging;
      draggedItem.classList.remove(...draggingClass.split(' '));
    }

    hideDropIndicator();
    clearDragState();

    draggedItem = null;
    draggedItemId = null;
    draggedItemIndex = -1;
    currentDropTarget = null;
  }

  function handleDragOver(e: DragEvent) {
    if (opts.enabled === false) return;

    const state = getDragState();

    // Check if we accept this drag
    if (!state.data) {
      const dataStr = e.dataTransfer?.getData(DRAG_DATA_MIME);
      if (!dataStr) return;

      const data = deserializeDragData(dataStr);
      if (!data || data.type !== opts.type) return;

      // External drag - check if we accept from other lists
      if (data.sourceId !== opts.id && !opts.acceptFromOther) {
        return;
      }
    } else {
      if (state.data.type !== opts.type) return;

      // Same window - check if we accept from other lists
      if (state.data.sourceId !== opts.id && !opts.acceptFromOther) {
        return;
      }
    }

    e.preventDefault();

    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'move';
    }

    // Find the item under the cursor
    const target = e.target as HTMLElement;
    const item = target.closest(opts.itemSelector!) as HTMLElement;

    if (!item || !node.contains(item)) {
      hideDropIndicator();
      currentDropTarget = null;
      return;
    }

    // Don't show indicator on self
    if (state.data && getItemId(item) === state.data.id) {
      hideDropIndicator();
      currentDropTarget = null;
      return;
    }

    currentDropTarget = item;
    currentDropPosition = calculateDropPosition(e, item, 0.5) === 'before' ? 'before' : 'after';

    showDropIndicator(item, currentDropPosition);

    const overId = getItemId(item);
    if (state.data) {
      opts.onSortOver?.(state.data.id, overId, currentDropPosition);
    }
  }

  function handleDragLeave(e: DragEvent) {
    const relatedTarget = e.relatedTarget as Node;

    // Check if leaving the entire sortable container
    if (!relatedTarget || !node.contains(relatedTarget)) {
      hideDropIndicator();
      currentDropTarget = null;
    }
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();

    hideDropIndicator();

    if (!currentDropTarget) return;

    // Get drag data
    const state = getDragState();
    let dragData: DragData<T> | null = state.data as DragData<T> | null;

    if (!dragData) {
      const dataStr = e.dataTransfer?.getData(DRAG_DATA_MIME);
      if (dataStr) {
        dragData = deserializeDragData<T>(dataStr);
      }
    }

    if (!dragData || dragData.type !== opts.type) return;

    // Check if we accept from other lists
    if (dragData.sourceId !== opts.id && !opts.acceptFromOther) {
      return;
    }

    const sourceListId = dragData.sourceId || opts.id;
    const sourceIndex = dragData.sourceIndex ?? -1;
    const targetItemId = getItemId(currentDropTarget);
    const targetIndex = getItemIndex(targetItemId);

    // Calculate destination index
    let destinationIndex = targetIndex;
    if (currentDropPosition === 'after') {
      destinationIndex = targetIndex + 1;
    }

    // Adjust for same-list moves
    if (sourceListId === opts.id && sourceIndex < destinationIndex) {
      destinationIndex -= 1;
    }

    const result: SortResult<T> = {
      itemId: dragData.id,
      sourceListId,
      destinationListId: opts.id,
      sourceIndex,
      destinationIndex,
      movedToNewList: sourceListId !== opts.id,
      data: dragData.data,
    };

    opts.onSortEnd?.(result);

    currentDropTarget = null;
  }

  function setupSortable() {
    node.classList.add(...sortableClasses.list.split(' '));
  }

  // Initial setup
  setupSortable();

  // Add event listeners to container
  node.addEventListener('dragstart', handleDragStart);
  node.addEventListener('dragend', handleDragEnd);
  node.addEventListener('dragover', handleDragOver);
  node.addEventListener('dragleave', handleDragLeave);
  node.addEventListener('drop', handleDrop);

  return {
    update(newOptions: SortableOptions<T>) {
      opts = {
        itemSelector: 'li',
        itemIdAttribute: 'data-id',
        acceptFromOther: true,
        animationDuration: 200,
        ...newOptions,
      };
    },
    destroy() {
      node.removeEventListener('dragstart', handleDragStart);
      node.removeEventListener('dragend', handleDragEnd);
      node.removeEventListener('dragover', handleDragOver);
      node.removeEventListener('dragleave', handleDragLeave);
      node.removeEventListener('drop', handleDrop);
      node.classList.remove(...sortableClasses.list.split(' '));
      hideDropIndicator();
    },
  };
}

// Re-export utility function for convenience
export { reorderItems } from './dragdrop.types';
