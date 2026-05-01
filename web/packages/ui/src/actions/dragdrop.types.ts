/**
 * Drag and Drop Types and Utilities
 * Generic drag-and-drop system for any element
 */

// ============================================================================
// Types
// ============================================================================

/** Data transfer payload for drag operations */
export interface DragData<T = unknown> {
  /** Unique identifier for the dragged item */
  id: string;
  /** Type of the dragged item (for filtering valid drops) */
  type: string;
  /** Custom payload data */
  data: T;
  /** Source container/group ID */
  sourceId?: string;
  /** Index in source list (for sortable) */
  sourceIndex?: number;
}

/** Position relative to target element */
export type DropPosition = 'before' | 'after' | 'inside';

/** Drag effect allowed */
export type DragEffect = 'copy' | 'move' | 'link' | 'none';

/** Draggable action options */
export interface DraggableOptions<T = unknown> {
  /** Unique identifier for this draggable item */
  id: string;
  /** Type identifier (for filtering valid drops) */
  type: string;
  /** Custom data payload */
  data?: T;
  /** Container/group this item belongs to */
  containerId?: string;
  /** Whether dragging is enabled */
  enabled?: boolean;
  /** Allowed drag effects */
  effectAllowed?: DragEffect | 'all';
  /** Handle selector (only allow dragging from this child element) */
  handle?: string;
  /** CSS class applied while dragging */
  draggingClass?: string;
  /** Callback when drag starts */
  onDragStart?: (data: DragData<T>, event: DragEvent) => void;
  /** Callback when drag ends (completed or cancelled) */
  onDragEnd?: (data: DragData<T>, event: DragEvent) => void;
}

/** Droppable action options */
export interface DroppableOptions<T = unknown> {
  /** Unique identifier for this drop zone */
  id: string;
  /** Accepted drag types (empty = accept all) */
  acceptTypes?: string[];
  /** Container/group this drop zone belongs to */
  containerId?: string;
  /** Whether dropping is enabled */
  enabled?: boolean;
  /** Drop effect */
  dropEffect?: DragEffect;
  /** Show position indicators (before/after/inside) */
  showPosition?: boolean;
  /** CSS class applied when valid drag is over */
  dragOverClass?: string;
  /** Callback to validate if drop is allowed */
  canDrop?: (data: DragData<T>) => boolean;
  /** Callback when valid drag enters */
  onDragEnter?: (data: DragData<T>, event: DragEvent) => void;
  /** Callback when valid drag is over */
  onDragOver?: (data: DragData<T>, position: DropPosition, event: DragEvent) => void;
  /** Callback when valid drag leaves */
  onDragLeave?: (data: DragData<T>, event: DragEvent) => void;
  /** Callback when item is dropped */
  onDrop?: (data: DragData<T>, position: DropPosition, event: DragEvent) => void;
}

/** Sortable list options */
export interface SortableOptions<T = unknown> {
  /** Unique identifier for this sortable list */
  id: string;
  /** Type identifier for items in this list */
  type: string;
  /** Whether sorting is enabled */
  enabled?: boolean;
  /** Handle selector (only allow dragging from this child element) */
  handle?: string;
  /** Item selector (which children are sortable) */
  itemSelector?: string;
  /** Data attribute name for item ID */
  itemIdAttribute?: string;
  /** Whether to accept items from other lists */
  acceptFromOther?: boolean;
  /** CSS class applied to dragging item */
  draggingClass?: string;
  /** CSS class applied to ghost/placeholder */
  ghostClass?: string;
  /** Animation duration in ms */
  animationDuration?: number;
  /** Callback when sort starts */
  onSortStart?: (itemId: string, index: number) => void;
  /** Callback when item is over a new position */
  onSortOver?: (itemId: string, overItemId: string, position: DropPosition) => void;
  /** Callback when sort ends */
  onSortEnd?: (result: SortResult<T>) => void;
}

/** Result of a sort operation */
export interface SortResult<T = unknown> {
  /** Item ID that was moved */
  itemId: string;
  /** Source list ID */
  sourceListId: string;
  /** Destination list ID */
  destinationListId: string;
  /** Original index */
  sourceIndex: number;
  /** New index */
  destinationIndex: number;
  /** Whether moved to a different list */
  movedToNewList: boolean;
  /** Custom data from the item */
  data?: T;
}

// ============================================================================
// CSS Classes
// ============================================================================

export const draggableClasses = {
  /** Base draggable styles */
  base: 'cursor-grab select-none',
  /** Applied while dragging */
  dragging: 'opacity-50 cursor-grabbing',
  /** Applied to disabled draggable */
  disabled: 'cursor-default',
};

export const droppableClasses = {
  /** Base drop zone styles */
  base: 'transition-colors duration-150',
  /** Applied when valid drag is over */
  dragOver: 'bg-brand-primary-50 ring-2 ring-brand-primary-500 ring-inset',
  /** Applied when invalid drag is over */
  dragOverInvalid: 'bg-semantic-error-50 ring-2 ring-semantic-error-500 ring-inset',
  /** Drop indicator line */
  indicator: 'absolute bg-brand-primary-500 z-50 pointer-events-none',
  indicatorHorizontal: 'left-0 right-0 h-0.5',
  indicatorVertical: 'top-0 bottom-0 w-0.5',
};

export const sortableClasses = {
  /** Base list styles */
  list: 'relative',
  /** Applied to item while dragging */
  itemDragging: 'opacity-50',
  /** Ghost/placeholder element */
  ghost: 'border-2 border-dashed border-brand-primary-300 bg-brand-primary-50 rounded',
  /** Applied during sort animation */
  animating: 'transition-transform duration-200',
};

// ============================================================================
// Utility Functions
// ============================================================================

/** Global drag state (shared across instances) */
interface DragState {
  isDragging: boolean;
  data: DragData | null;
  sourceElement: HTMLElement | null;
}

const dragState: DragState = {
  isDragging: false,
  data: null,
  sourceElement: null,
};

/** Get current drag state */
export function getDragState(): Readonly<DragState> {
  return dragState;
}

/** Set drag state (internal use) */
export function setDragState(state: Partial<DragState>): void {
  Object.assign(dragState, state);
}

/** Clear drag state (internal use) */
export function clearDragState(): void {
  dragState.isDragging = false;
  dragState.data = null;
  dragState.sourceElement = null;
}

/** Serialize drag data for data transfer */
export function serializeDragData(data: DragData): string {
  return JSON.stringify(data);
}

/** Deserialize drag data from data transfer */
export function deserializeDragData<T = unknown>(dataStr: string): DragData<T> | null {
  try {
    return JSON.parse(dataStr) as DragData<T>;
  } catch {
    return null;
  }
}

/** Calculate drop position based on cursor position */
export function calculateDropPosition(
  event: DragEvent,
  element: HTMLElement,
  threshold = 0.25
): DropPosition {
  const rect = element.getBoundingClientRect();
  const y = event.clientY - rect.top;
  const relativeY = y / rect.height;

  if (relativeY < threshold) {
    return 'before';
  } else if (relativeY > 1 - threshold) {
    return 'after';
  }
  return 'inside';
}

/** Calculate drop position for horizontal lists */
export function calculateDropPositionHorizontal(
  event: DragEvent,
  element: HTMLElement,
  threshold = 0.5
): DropPosition {
  const rect = element.getBoundingClientRect();
  const x = event.clientX - rect.left;
  const relativeX = x / rect.width;

  return relativeX < threshold ? 'before' : 'after';
}

/** Reorder array items (for sortable) */
export function reorderItems<T>(
  items: T[],
  sourceIndex: number,
  destinationIndex: number
): T[] {
  const result = [...items];
  const [removed] = result.splice(sourceIndex, 1);
  result.splice(destinationIndex, 0, removed!);
  return result;
}

/** Move item between arrays (for sortable across lists) */
export function moveItem<T>(
  sourceItems: T[],
  destinationItems: T[],
  sourceIndex: number,
  destinationIndex: number
): { source: T[]; destination: T[] } {
  const source = [...sourceItems];
  const destination = [...destinationItems];
  const [removed] = source.splice(sourceIndex, 1);
  destination.splice(destinationIndex, 0, removed!);
  return { source, destination };
}

/** Create drop indicator element */
export function createDropIndicator(
  orientation: 'horizontal' | 'vertical' = 'horizontal'
): HTMLElement {
  const indicator = document.createElement('div');
  indicator.className = `${droppableClasses.indicator} ${
    orientation === 'horizontal'
      ? droppableClasses.indicatorHorizontal
      : droppableClasses.indicatorVertical
  }`;
  return indicator;
}

/** Position drop indicator */
export function positionDropIndicator(
  indicator: HTMLElement,
  target: HTMLElement,
  position: DropPosition,
  orientation: 'horizontal' | 'vertical' = 'horizontal'
): void {
  const rect = target.getBoundingClientRect();
  const parentRect = target.parentElement?.getBoundingClientRect();

  if (!parentRect) return;

  if (orientation === 'horizontal') {
    indicator.style.width = `${rect.width}px`;
    indicator.style.left = `${rect.left - parentRect.left}px`;
    indicator.style.height = '2px';

    if (position === 'before') {
      indicator.style.top = `${rect.top - parentRect.top}px`;
    } else {
      indicator.style.top = `${rect.bottom - parentRect.top}px`;
    }
  } else {
    indicator.style.height = `${rect.height}px`;
    indicator.style.top = `${rect.top - parentRect.top}px`;
    indicator.style.width = '2px';

    if (position === 'before') {
      indicator.style.left = `${rect.left - parentRect.left}px`;
    } else {
      indicator.style.left = `${rect.right - parentRect.left}px`;
    }
  }
}

// MIME type for drag data
export const DRAG_DATA_MIME = 'application/x-ui-drag-data';
