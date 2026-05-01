// Svelte Actions
export { portal, type PortalOptions } from './portal';
export { clickOutside, type ClickOutsideOptions } from './clickOutside';
export { focusTrap, type FocusTrapOptions } from './focusTrap';
export { columnResize, columnResizeClasses, type ColumnResizeOptions } from './columnResize';
export { columnReorder, reorderColumns, columnReorderClasses, type ColumnReorderOptions } from './columnReorder';

// Drag and Drop Actions
export { draggable, type DraggableReturn } from './draggable';
export { droppable, type DroppableReturn } from './droppable';
export { sortable, reorderItems, type SortableReturn } from './sortable';
export {
  type DragData,
  type DropPosition,
  type DragEffect,
  type DraggableOptions,
  type DroppableOptions,
  type SortableOptions,
  type SortResult,
  draggableClasses,
  droppableClasses,
  sortableClasses,
  getDragState,
  calculateDropPosition,
  calculateDropPositionHorizontal,
  moveItem,
  DRAG_DATA_MIME,
} from './dragdrop.types';

// Custom Transitions
export {
  fadeScale,
  slideUp,
  slideDown,
  slideLeft,
  slideRight,
  pop,
  collapse,
  blur,
  typewriter,
  type TransitionOptions,
} from './transition';

// Action Components
export { default as Button } from './Button.svelte';
export { default as ButtonGroup } from './ButtonGroup.svelte';
export { default as CloseButton } from './CloseButton.svelte';
export { default as Collapse } from './Collapse.svelte';
export { default as Accordion } from './Accordion.svelte';
export { default as AccordionItem } from './AccordionItem.svelte';
export { default as Scrollspy } from './Scrollspy.svelte';

// Action Component Types
export type {
  ButtonProps,
  ButtonVariant,
  ButtonGroupProps,
  CloseButtonProps,
  CollapseProps,
  AccordionProps,
  AccordionItemData,
  AccordionItemProps,
  ScrollspyProps,
  ScrollspyItem,
} from './actions.types';

// Action Component Classes (for custom implementations)
export {
  buttonBaseClasses,
  buttonSizeClasses,
  buttonIconOnlySizeClasses,
  buttonVariantClasses,
  buttonLoadingSpinnerClasses,
  buttonGroupClasses,
  closeButtonClasses,
  closeButtonSizeClasses,
  collapseClasses,
  accordionClasses,
  accordionItemClasses,
  accordionSizeClasses,
  scrollspyClasses,
} from './actions.types';
