/**
 * Column resize action for tables
 * Enables resizing table columns by dragging their borders
 */

export interface ColumnResizeOptions {
  /** Minimum column width in pixels */
  minWidth?: number;
  /** Maximum column width in pixels */
  maxWidth?: number;
  /** Column key identifier */
  columnKey: string;
  /** Callback when resize starts */
  onResizeStart?: (columnKey: string, initialWidth: number) => void;
  /** Callback during resize */
  onResize?: (columnKey: string, width: number) => void;
  /** Callback when resize ends */
  onResizeEnd?: (columnKey: string, finalWidth: number) => void;
}

export interface ColumnResizeReturn {
  update: (options: ColumnResizeOptions) => void;
  destroy: () => void;
}

/**
 * Svelte action for column resizing
 * Usage: <th use:columnResize={{ columnKey: 'name', onResize: handleResize }}>
 */
export function columnResize(
  node: HTMLElement,
  options: ColumnResizeOptions
): ColumnResizeReturn {
  let { minWidth = 50, maxWidth = 800, columnKey, onResizeStart, onResize, onResizeEnd } = options;

  // Create resize handle element
  const handle = document.createElement('div');
  handle.className = 'column-resize-handle';
  handle.style.cssText = `
    position: absolute;
    right: 0;
    top: 0;
    bottom: 0;
    width: 4px;
    cursor: col-resize;
    user-select: none;
    z-index: 1;
  `;

  // Make parent relative for handle positioning
  const originalPosition = node.style.position;
  node.style.position = 'relative';
  node.appendChild(handle);

  let isResizing = false;
  let startX = 0;
  let startWidth = 0;

  function handleMouseDown(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();

    isResizing = true;
    startX = e.pageX;
    startWidth = node.offsetWidth;

    // Add active state styling
    handle.style.backgroundColor = 'var(--color-brand-primary-500, #0ea5e9)';

    onResizeStart?.(columnKey, startWidth);

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);

    // Prevent text selection during resize
    document.body.style.userSelect = 'none';
    document.body.style.cursor = 'col-resize';
  }

  function handleMouseMove(e: MouseEvent) {
    if (!isResizing) return;

    const diff = e.pageX - startX;
    let newWidth = startWidth + diff;

    // Apply constraints
    newWidth = Math.max(minWidth, Math.min(maxWidth, newWidth));

    // Apply width to the column
    node.style.width = `${newWidth}px`;
    node.style.minWidth = `${newWidth}px`;
    node.style.maxWidth = `${newWidth}px`;

    onResize?.(columnKey, newWidth);
  }

  function handleMouseUp() {
    if (!isResizing) return;

    isResizing = false;
    handle.style.backgroundColor = '';

    const finalWidth = node.offsetWidth;
    onResizeEnd?.(columnKey, finalWidth);

    document.removeEventListener('mousemove', handleMouseMove);
    document.removeEventListener('mouseup', handleMouseUp);

    // Restore normal cursor and selection
    document.body.style.userSelect = '';
    document.body.style.cursor = '';
  }

  // Hover effect for handle
  function handleMouseEnter() {
    if (!isResizing) {
      handle.style.backgroundColor = 'var(--color-neutral-300, #d4d4d4)';
    }
  }

  function handleMouseLeave() {
    if (!isResizing) {
      handle.style.backgroundColor = '';
    }
  }

  handle.addEventListener('mousedown', handleMouseDown);
  handle.addEventListener('mouseenter', handleMouseEnter);
  handle.addEventListener('mouseleave', handleMouseLeave);

  return {
    update(newOptions: ColumnResizeOptions) {
      minWidth = newOptions.minWidth ?? 50;
      maxWidth = newOptions.maxWidth ?? 800;
      columnKey = newOptions.columnKey;
      onResizeStart = newOptions.onResizeStart;
      onResize = newOptions.onResize;
      onResizeEnd = newOptions.onResizeEnd;
    },
    destroy() {
      handle.removeEventListener('mousedown', handleMouseDown);
      handle.removeEventListener('mouseenter', handleMouseEnter);
      handle.removeEventListener('mouseleave', handleMouseLeave);
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
      handle.remove();
      node.style.position = originalPosition;
    }
  };
}

/**
 * CSS classes for column resize styling
 */
export const columnResizeClasses = {
  handle: 'column-resize-handle absolute right-0 top-0 bottom-0 w-1 cursor-col-resize z-10 hover:bg-color-neutral-300 active:bg-color-brand-primary-500',
  resizing: 'select-none',
  column: 'relative',
};
