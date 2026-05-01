/**
 * Column reorder action for tables
 * Enables drag-and-drop reordering of table columns
 */

export interface ColumnReorderOptions {
  /** Column key identifier */
  columnKey: string;
  /** Group identifier for columns that can be reordered together */
  group?: string;
  /** Whether this column can be dragged */
  draggable?: boolean;
  /** Whether this column accepts drops */
  droppable?: boolean;
  /** Callback when drag starts */
  onDragStart?: (columnKey: string) => void;
  /** Callback when dragging over a valid target */
  onDragOver?: (sourceKey: string, targetKey: string) => void;
  /** Callback when columns are reordered */
  onReorder?: (sourceKey: string, targetKey: string, position: 'before' | 'after') => void;
  /** Callback when drag ends (cancelled or completed) */
  onDragEnd?: (columnKey: string) => void;
}

export interface ColumnReorderReturn {
  update: (options: ColumnReorderOptions) => void;
  destroy: () => void;
}

// Store the currently dragged column key (shared across instances)
let currentDragKey: string | null = null;
let currentDragGroup: string | null = null;

/**
 * Svelte action for column reordering via drag-and-drop
 * Usage: <th use:columnReorder={{ columnKey: 'name', onReorder: handleReorder }}>
 */
export function columnReorder(
  node: HTMLElement,
  options: ColumnReorderOptions
): ColumnReorderReturn {
  let {
    columnKey,
    group = 'default',
    draggable = true,
    droppable = true,
    onDragStart,
    onDragOver,
    onReorder,
    onDragEnd
  } = options;

  // Set draggable attribute
  if (draggable) {
    node.setAttribute('draggable', 'true');
    node.style.cursor = 'grab';
  }

  let dropIndicator: HTMLElement | null = null;
  let dropPosition: 'before' | 'after' = 'after';

  function createDropIndicator() {
    dropIndicator = document.createElement('div');
    dropIndicator.className = 'column-drop-indicator';
    dropIndicator.style.cssText = `
      position: absolute;
      top: 0;
      bottom: 0;
      width: 3px;
      background-color: var(--color-brand-primary-500, #0ea5e9);
      z-index: 100;
      pointer-events: none;
    `;
    return dropIndicator;
  }

  function handleDragStart(e: DragEvent) {
    if (!draggable) return;

    currentDragKey = columnKey;
    currentDragGroup = group;

    // Set drag data
    e.dataTransfer?.setData('text/plain', columnKey);
    e.dataTransfer?.setData('application/column-key', columnKey);
    e.dataTransfer?.setData('application/column-group', group);

    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move';
    }

    // Add dragging style
    node.style.opacity = '0.5';
    node.style.cursor = 'grabbing';

    onDragStart?.(columnKey);
  }

  function handleDragEnd(e: DragEvent) {
    currentDragKey = null;
    currentDragGroup = null;

    // Remove dragging style
    node.style.opacity = '';
    node.style.cursor = draggable ? 'grab' : '';

    // Remove drop indicator
    dropIndicator?.remove();
    dropIndicator = null;

    onDragEnd?.(columnKey);
  }

  function handleDragOver(e: DragEvent) {
    if (!droppable || !currentDragKey || currentDragKey === columnKey) {
      return;
    }

    // Only allow drops from the same group
    if (currentDragGroup !== group) {
      return;
    }

    e.preventDefault();
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'move';
    }

    // Determine drop position based on mouse position
    const rect = node.getBoundingClientRect();
    const midpoint = rect.left + rect.width / 2;
    dropPosition = e.clientX < midpoint ? 'before' : 'after';

    // Show drop indicator
    if (!dropIndicator) {
      dropIndicator = createDropIndicator();
      node.style.position = 'relative';
      node.appendChild(dropIndicator);
    }

    // Position indicator
    if (dropPosition === 'before') {
      dropIndicator.style.left = '-1px';
      dropIndicator.style.right = '';
    } else {
      dropIndicator.style.left = '';
      dropIndicator.style.right = '-1px';
    }

    onDragOver?.(currentDragKey, columnKey);
  }

  function handleDragEnter(e: DragEvent) {
    if (!droppable || !currentDragKey || currentDragKey === columnKey) {
      return;
    }

    if (currentDragGroup !== group) {
      return;
    }

    e.preventDefault();
    node.classList.add('column-drag-over');
  }

  function handleDragLeave(e: DragEvent) {
    // Check if we're leaving to a child element
    const relatedTarget = e.relatedTarget as Node;
    if (relatedTarget && node.contains(relatedTarget)) {
      return;
    }

    node.classList.remove('column-drag-over');
    dropIndicator?.remove();
    dropIndicator = null;
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();

    if (!droppable || !currentDragKey || currentDragKey === columnKey) {
      return;
    }

    if (currentDragGroup !== group) {
      return;
    }

    const sourceKey = e.dataTransfer?.getData('application/column-key') || currentDragKey;

    node.classList.remove('column-drag-over');
    dropIndicator?.remove();
    dropIndicator = null;

    onReorder?.(sourceKey, columnKey, dropPosition);
  }

  // Add event listeners
  node.addEventListener('dragstart', handleDragStart);
  node.addEventListener('dragend', handleDragEnd);
  node.addEventListener('dragover', handleDragOver);
  node.addEventListener('dragenter', handleDragEnter);
  node.addEventListener('dragleave', handleDragLeave);
  node.addEventListener('drop', handleDrop);

  return {
    update(newOptions: ColumnReorderOptions) {
      columnKey = newOptions.columnKey;
      group = newOptions.group ?? 'default';
      draggable = newOptions.draggable ?? true;
      droppable = newOptions.droppable ?? true;
      onDragStart = newOptions.onDragStart;
      onDragOver = newOptions.onDragOver;
      onReorder = newOptions.onReorder;
      onDragEnd = newOptions.onDragEnd;

      // Update draggable state
      if (draggable) {
        node.setAttribute('draggable', 'true');
        node.style.cursor = 'grab';
      } else {
        node.removeAttribute('draggable');
        node.style.cursor = '';
      }
    },
    destroy() {
      node.removeEventListener('dragstart', handleDragStart);
      node.removeEventListener('dragend', handleDragEnd);
      node.removeEventListener('dragover', handleDragOver);
      node.removeEventListener('dragenter', handleDragEnter);
      node.removeEventListener('dragleave', handleDragLeave);
      node.removeEventListener('drop', handleDrop);
      dropIndicator?.remove();
    }
  };
}

/**
 * Helper function to reorder columns array
 */
export function reorderColumns<T extends { key: string }>(
  columns: T[],
  sourceKey: string,
  targetKey: string,
  position: 'before' | 'after'
): T[] {
  const newColumns = [...columns];
  const sourceIndex = newColumns.findIndex((col) => col.key === sourceKey);
  const targetIndex = newColumns.findIndex((col) => col.key === targetKey);

  if (sourceIndex === -1 || targetIndex === -1) {
    return columns;
  }

  // Remove source column
  const [removed] = newColumns.splice(sourceIndex, 1);

  // Calculate new target index (accounting for the removal)
  let insertIndex = newColumns.findIndex((col) => col.key === targetKey);
  if (position === 'after') {
    insertIndex += 1;
  }

  // Insert at new position
  newColumns.splice(insertIndex, 0, removed!);

  return newColumns;
}

/**
 * CSS classes for column reorder styling
 */
export const columnReorderClasses = {
  draggable: 'cursor-grab active:cursor-grabbing',
  dragging: 'opacity-50',
  dragOver: 'bg-color-brand-primary-50',
  dropIndicator: 'absolute top-0 bottom-0 w-0.5 bg-color-brand-primary-500 z-100',
  dropIndicatorBefore: 'left-0',
  dropIndicatorAfter: 'right-0',
};
