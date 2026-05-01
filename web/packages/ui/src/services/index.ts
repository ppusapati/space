/**
 * UI Services
 * Programmatic services for managing UI state
 */

// Modal Stack Management
export {
  modalStack,
  openModal,
  openDialog,
  openDrawer,
  closeModal,
  closeTopModal,
  closeAllModals,
  modalStackClasses,
  type ModalStackItem,
  type ModalConfig,
  type DialogConfig,
  type DrawerConfig,
  type ModalResult,
} from './modal-stack';

// Modal Stack Renderer Component
export { default as ModalStackRenderer } from './ModalStackRenderer.svelte';
