import { w as writable, d as derived, g as get } from "./index.js";
const INITIAL_STATE = {
  stack: [],
  baseZIndex: 1e3,
  zIndexIncrement: 10
};
function generateId() {
  return `modal-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
}
function createModalStack() {
  const store = writable(INITIAL_STATE);
  const { subscribe, update, set } = store;
  function getNextZIndex(state) {
    return state.baseZIndex + state.stack.length * state.zIndexIncrement;
  }
  function openModal2(config = {}) {
    return new Promise((resolve, reject) => {
      update((state) => {
        const id = generateId();
        const item = {
          id,
          type: "modal",
          zIndex: getNextZIndex(state),
          openedAt: Date.now(),
          props: { ...config },
          resolve,
          reject
        };
        return { ...state, stack: [...state.stack, item] };
      });
    });
  }
  function openDialog2(config = {}) {
    return new Promise((resolve, reject) => {
      update((state) => {
        const id = generateId();
        const item = {
          id,
          type: "dialog",
          zIndex: getNextZIndex(state),
          openedAt: Date.now(),
          props: {
            variant: "confirm",
            confirmText: "Confirm",
            cancelText: "Cancel",
            closeOnBackdrop: false,
            ...config
          },
          resolve,
          reject
        };
        return { ...state, stack: [...state.stack, item] };
      });
    });
  }
  function openDrawer2(config = {}) {
    return new Promise((resolve, reject) => {
      update((state) => {
        const id = generateId();
        const item = {
          id,
          type: "drawer",
          zIndex: getNextZIndex(state),
          openedAt: Date.now(),
          props: {
            position: "right",
            closeOnBackdrop: true,
            closeOnEscape: true,
            showClose: true,
            overlay: true,
            ...config
          },
          resolve,
          reject
        };
        return { ...state, stack: [...state.stack, item] };
      });
    });
  }
  function close(id, result) {
    update((state) => {
      const item = state.stack.find((m) => m.id === id);
      if (item?.resolve) {
        item.resolve(result ?? { confirmed: false });
      }
      return {
        ...state,
        stack: state.stack.filter((m) => m.id !== id)
      };
    });
  }
  function closeTop(result) {
    const state = get(store);
    const topModal2 = state.stack[state.stack.length - 1];
    if (topModal2) {
      close(topModal2.id, result);
    }
  }
  function confirm(data) {
    closeTop({ confirmed: true, data });
  }
  function cancel() {
    closeTop({ confirmed: false });
  }
  function closeAll() {
    update((state) => {
      state.stack.forEach((item) => {
        if (item.resolve) {
          item.resolve({ confirmed: false });
        }
      });
      return { ...state, stack: [] };
    });
  }
  function isOpen(id) {
    return get(store).stack.some((m) => m.id === id);
  }
  function getModal(id) {
    return get(store).stack.find((m) => m.id === id);
  }
  const stack = derived(store, ($state) => $state.stack);
  const hasOpenModals = derived(store, ($state) => $state.stack.length > 0);
  const topModal = derived(
    store,
    ($state) => $state.stack[$state.stack.length - 1] ?? null
  );
  const count = derived(store, ($state) => $state.stack.length);
  return {
    subscribe,
    // Opening methods
    openModal: openModal2,
    openDialog: openDialog2,
    openDrawer: openDrawer2,
    // Alias methods
    open: openModal2,
    alert: (config) => openDialog2({ ...config, variant: "info", cancelText: "" }),
    confirm: (config) => openDialog2({ ...config, variant: "confirm" }),
    warning: (config) => openDialog2({ ...config, variant: "warning" }),
    error: (config) => openDialog2({ ...config, variant: "error" }),
    success: (config) => openDialog2({ ...config, variant: "success" }),
    // Closing methods
    close,
    closeTop,
    closeAll,
    // Result methods
    confirmTop: confirm,
    cancelTop: cancel,
    // Query methods
    isOpen,
    getModal,
    // Derived stores
    stack,
    hasOpenModals,
    topModal,
    count
  };
}
createModalStack();
