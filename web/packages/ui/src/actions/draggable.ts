/**
 * Draggable Svelte Action
 * Makes any element draggable with HTML5 Drag and Drop API
 */

import type { DraggableOptions, DragData } from './dragdrop.types';
import {
  draggableClasses,
  setDragState,
  clearDragState,
  serializeDragData,
  DRAG_DATA_MIME,
} from './dragdrop.types';

export interface DraggableReturn {
  update: (options: DraggableOptions) => void;
  destroy: () => void;
}

/**
 * Svelte action for making elements draggable
 *
 * @example
 * ```svelte
 * <div
 *   use:draggable={{
 *     id: 'item-1',
 *     type: 'card',
 *     data: { title: 'My Card' },
 *     onDragStart: (data) => console.log('Started dragging', data),
 *     onDragEnd: (data) => console.log('Finished dragging', data)
 *   }}
 * >
 *   Drag me
 * </div>
 * ```
 */
export function draggable<T = unknown>(
  node: HTMLElement,
  options: DraggableOptions<T>
): DraggableReturn {
  let opts = { ...options };
  let handleElement: HTMLElement | null = null;
  let isMouseDownOnHandle = false;

  function getDragData(): DragData<T> {
    return {
      id: opts.id,
      type: opts.type,
      data: opts.data as T,
      sourceId: opts.containerId,
    };
  }

  function handleMouseDown(e: MouseEvent) {
    // Track if mousedown is on handle
    if (handleElement && handleElement.contains(e.target as Node)) {
      isMouseDownOnHandle = true;
    } else if (opts.handle) {
      isMouseDownOnHandle = false;
    } else {
      isMouseDownOnHandle = true;
    }
  }

  function handleMouseUp() {
    isMouseDownOnHandle = false;
  }

  function handleDragStart(e: DragEvent) {
    if (opts.enabled === false) {
      e.preventDefault();
      return;
    }

    // If handle is specified, only allow drag from handle
    if (opts.handle && !isMouseDownOnHandle) {
      e.preventDefault();
      return;
    }

    const dragData = getDragData();

    // Set drag data
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = opts.effectAllowed || 'move';
      e.dataTransfer.setData('text/plain', opts.id);
      e.dataTransfer.setData(DRAG_DATA_MIME, serializeDragData(dragData));
    }

    // Update global state
    setDragState({
      isDragging: true,
      data: dragData,
      sourceElement: node,
    });

    // Apply dragging class
    const draggingClass = opts.draggingClass || draggableClasses.dragging;
    node.classList.add(...draggingClass.split(' '));

    // Callback
    opts.onDragStart?.(dragData, e);
  }

  function handleDragEnd(e: DragEvent) {
    const dragData = getDragData();

    // Remove dragging class
    const draggingClass = opts.draggingClass || draggableClasses.dragging;
    node.classList.remove(...draggingClass.split(' '));

    // Clear global state
    clearDragState();

    // Callback
    opts.onDragEnd?.(dragData, e);

    isMouseDownOnHandle = false;
  }

  function setupDraggable() {
    if (opts.enabled === false) {
      node.removeAttribute('draggable');
      node.classList.remove(...draggableClasses.base.split(' '));
      node.classList.add(...draggableClasses.disabled.split(' '));
      return;
    }

    node.setAttribute('draggable', 'true');
    node.classList.add(...draggableClasses.base.split(' '));
    node.classList.remove(...draggableClasses.disabled.split(' '));

    // Setup handle if specified
    if (opts.handle) {
      handleElement = node.querySelector(opts.handle);
      if (handleElement) {
        handleElement.style.cursor = 'grab';
        node.style.cursor = 'default';
      }
    } else {
      handleElement = null;
    }
  }

  // Initial setup
  setupDraggable();

  // Add event listeners
  node.addEventListener('mousedown', handleMouseDown);
  node.addEventListener('mouseup', handleMouseUp);
  node.addEventListener('dragstart', handleDragStart);
  node.addEventListener('dragend', handleDragEnd);

  return {
    update(newOptions: DraggableOptions<unknown>) {
      opts = { ...(newOptions as DraggableOptions<T>) };
      setupDraggable();
    },
    destroy() {
      node.removeEventListener('mousedown', handleMouseDown);
      node.removeEventListener('mouseup', handleMouseUp);
      node.removeEventListener('dragstart', handleDragStart);
      node.removeEventListener('dragend', handleDragEnd);
      node.removeAttribute('draggable');
      node.classList.remove(
        ...draggableClasses.base.split(' '),
        ...draggableClasses.dragging.split(' '),
        ...draggableClasses.disabled.split(' ')
      );
    },
  };
}
