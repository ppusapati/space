/**
 * Droppable Svelte Action
 * Makes any element a drop zone with HTML5 Drag and Drop API
 */

import type { DroppableOptions, DragData, DropPosition } from './dragdrop.types';
import {
  droppableClasses,
  getDragState,
  deserializeDragData,
  calculateDropPosition,
  createDropIndicator,
  DRAG_DATA_MIME,
} from './dragdrop.types';

export interface DroppableReturn {
  update: (options: DroppableOptions) => void;
  destroy: () => void;
}

/**
 * Svelte action for making elements droppable
 *
 * @example
 * ```svelte
 * <div
 *   use:droppable={{
 *     id: 'drop-zone-1',
 *     acceptTypes: ['card', 'item'],
 *     onDragEnter: (data) => console.log('Entered with', data),
 *     onDrop: (data, position) => console.log('Dropped', data, 'at', position)
 *   }}
 * >
 *   Drop here
 * </div>
 * ```
 */
export function droppable<T = unknown>(
  node: HTMLElement,
  options: DroppableOptions<T>
): DroppableReturn {
  let opts = { ...options };
  let dropIndicator: HTMLElement | null = null;
  let currentPosition: DropPosition = 'inside';
  let isValidDrag = false;

  function getDragData(e: DragEvent): DragData<T> | null {
    // First try global state (same window)
    const state = getDragState();
    if (state.data) {
      return state.data as DragData<T>;
    }

    // Fall back to data transfer (cross-window or external)
    const dataStr = e.dataTransfer?.getData(DRAG_DATA_MIME);
    if (dataStr) {
      return deserializeDragData<T>(dataStr);
    }

    return null;
  }

  function canAcceptDrop(data: DragData<T> | null): boolean {
    if (!data) return false;
    if (opts.enabled === false) return false;

    // Check accepted types
    if (opts.acceptTypes && opts.acceptTypes.length > 0) {
      if (!opts.acceptTypes.includes(data.type)) {
        return false;
      }
    }

    // Check custom validation
    if (opts.canDrop && !opts.canDrop(data)) {
      return false;
    }

    return true;
  }

  function showDropIndicator(position: DropPosition) {
    if (!opts.showPosition) return;

    if (!dropIndicator) {
      dropIndicator = createDropIndicator('horizontal');
      node.style.position = 'relative';
      node.appendChild(dropIndicator);
    }

    const rect = node.getBoundingClientRect();

    if (position === 'before') {
      dropIndicator.style.top = '0';
      dropIndicator.style.bottom = '';
      dropIndicator.style.left = '0';
      dropIndicator.style.right = '0';
      dropIndicator.style.width = '100%';
      dropIndicator.style.height = '2px';
    } else if (position === 'after') {
      dropIndicator.style.top = '';
      dropIndicator.style.bottom = '0';
      dropIndicator.style.left = '0';
      dropIndicator.style.right = '0';
      dropIndicator.style.width = '100%';
      dropIndicator.style.height = '2px';
    } else {
      // inside - hide indicator
      dropIndicator.style.display = 'none';
    }

    if (position !== 'inside') {
      dropIndicator.style.display = 'block';
    }
  }

  function hideDropIndicator() {
    if (dropIndicator) {
      dropIndicator.remove();
      dropIndicator = null;
    }
  }

  function handleDragEnter(e: DragEvent) {
    if (opts.enabled === false) return;

    const data = getDragData(e);
    isValidDrag = canAcceptDrop(data);

    if (!isValidDrag) {
      // Apply invalid class
      node.classList.add(...droppableClasses.dragOverInvalid.split(' '));
      return;
    }

    e.preventDefault();

    // Apply valid class
    const dragOverClass = opts.dragOverClass || droppableClasses.dragOver;
    node.classList.add(...dragOverClass.split(' '));

    opts.onDragEnter?.(data!, e);
  }

  function handleDragOver(e: DragEvent) {
    if (opts.enabled === false) return;

    const data = getDragData(e);

    if (!canAcceptDrop(data)) {
      return;
    }

    e.preventDefault();

    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = opts.dropEffect || 'move';
    }

    // Calculate position
    currentPosition = opts.showPosition
      ? calculateDropPosition(e, node)
      : 'inside';

    // Show indicator
    showDropIndicator(currentPosition);

    opts.onDragOver?.(data!, currentPosition, e);
  }

  function handleDragLeave(e: DragEvent) {
    // Check if leaving to a child element
    const relatedTarget = e.relatedTarget as Node;
    if (relatedTarget && node.contains(relatedTarget)) {
      return;
    }

    // Remove classes
    const dragOverClass = opts.dragOverClass || droppableClasses.dragOver;
    node.classList.remove(
      ...dragOverClass.split(' '),
      ...droppableClasses.dragOverInvalid.split(' ')
    );

    hideDropIndicator();

    if (isValidDrag) {
      const data = getDragData(e);
      if (data) {
        opts.onDragLeave?.(data, e);
      }
    }

    isValidDrag = false;
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();

    const data = getDragData(e);

    // Remove classes
    const dragOverClass = opts.dragOverClass || droppableClasses.dragOver;
    node.classList.remove(
      ...dragOverClass.split(' '),
      ...droppableClasses.dragOverInvalid.split(' ')
    );

    hideDropIndicator();

    if (!canAcceptDrop(data)) {
      return;
    }

    opts.onDrop?.(data!, currentPosition, e);
  }

  function setupDroppable() {
    node.classList.add(...droppableClasses.base.split(' '));
  }

  // Initial setup
  setupDroppable();

  // Add event listeners
  node.addEventListener('dragenter', handleDragEnter);
  node.addEventListener('dragover', handleDragOver);
  node.addEventListener('dragleave', handleDragLeave);
  node.addEventListener('drop', handleDrop);

  return {
    update(newOptions: DroppableOptions<T>) {
      opts = { ...newOptions };
    },
    destroy() {
      node.removeEventListener('dragenter', handleDragEnter);
      node.removeEventListener('dragover', handleDragOver);
      node.removeEventListener('dragleave', handleDragLeave);
      node.removeEventListener('drop', handleDrop);
      node.classList.remove(
        ...droppableClasses.base.split(' '),
        ...droppableClasses.dragOver.split(' '),
        ...droppableClasses.dragOverInvalid.split(' ')
      );
      hideDropIndicator();
    },
  };
}
