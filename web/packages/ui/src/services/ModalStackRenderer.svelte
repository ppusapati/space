<script lang="ts">
  /**
   * ModalStackRenderer
   * Renders all modals in the modal stack
   * Place this component once at the root of your application
   */
  import { modalStack, type ModalStackItem, type ModalResult } from './modal-stack';
  import type { ModalConfig, DialogConfig, DrawerConfig } from './modal-stack';
  import Modal from '../feedback/Modal.svelte';
  import Dialog from '../feedback/Dialog.svelte';
  import Drawer from '../feedback/Drawer.svelte';
  import type { Component } from 'svelte';

  // Subscribe to the stack
  $: items = $modalStack.stack;

  function handleClose(item: ModalStackItem) {
    modalStack.close(item.id, { confirmed: false });
  }

  function handleConfirm(item: ModalStackItem, data?: unknown) {
    modalStack.close(item.id, { confirmed: true, data });
  }

  function handleCancel(item: ModalStackItem) {
    modalStack.close(item.id, { confirmed: false });
  }

  function getModalProps(item: ModalStackItem): ModalConfig {
    return item.props as ModalConfig;
  }

  function getDialogProps(item: ModalStackItem): DialogConfig {
    return item.props as DialogConfig;
  }

  function getDrawerProps(item: ModalStackItem): DrawerConfig {
    return item.props as DrawerConfig;
  }

  function getComponent(item: ModalStackItem): Component<any> | null {
    return (item.props.component as Component<any>) ?? null;
  }
</script>

{#each items as item (item.id)}
  {#if item.type === 'modal'}
    {@const props = getModalProps(item)}
    <div style="z-index: {item.zIndex};" class="modal-stack-item">
      <Modal
        open={true}
        title={props.title}
        size={props.size}
        closeOnBackdrop={props.closeOnBackdrop ?? true}
        closeOnEscape={props.closeOnEscape ?? true}
        showClose={props.showClose ?? true}
        centered={props.centered ?? true}
        preventScroll={props.preventScroll ?? true}
        on:close={() => handleClose(item)}
      >
        {#if item.props.content}
          {item.props.content}
        {:else if item.props.component}
          <svelte:component
            this={getComponent(item)}
            {...(item.props.componentProps as Record<string, unknown>) ?? {}}
            on:confirm={(e: CustomEvent) => handleConfirm(item, e.detail)}
            on:cancel={() => handleCancel(item)}
            on:close={() => handleClose(item)}
          />
        {/if}
        <svelte:fragment slot="footer">
          {#if item.props.showFooter !== false && !item.props.component}
            <button
              type="button"
              class="px-4 py-2 text-sm font-medium text-neutral-700 bg-neutral-white border border-neutral-300 rounded-md hover:bg-neutral-50"
              on:click={() => handleCancel(item)}
            >
              {item.props.cancelText ?? 'Cancel'}
            </button>
            <button
              type="button"
              class="px-4 py-2 text-sm font-medium text-neutral-white bg-brand-primary-500 rounded-md hover:bg-brand-primary-600"
              on:click={() => handleConfirm(item)}
            >
              {item.props.confirmText ?? 'OK'}
            </button>
          {/if}
        </svelte:fragment>
      </Modal>
    </div>
  {:else if item.type === 'dialog'}
    {@const props = getDialogProps(item)}
    <div style="z-index: {item.zIndex};" class="modal-stack-item">
      <Dialog
        open={true}
        title={props.title}
        size={props.size ?? 'sm'}
        variant={props.variant ?? 'confirm'}
        confirmText={props.confirmText ?? 'Confirm'}
        cancelText={props.cancelText ?? 'Cancel'}
        destructive={props.destructive ?? false}
        closeOnBackdrop={props.closeOnBackdrop ?? false}
        closeOnEscape={props.closeOnEscape ?? true}
        on:confirm={() => handleConfirm(item)}
        on:cancel={() => handleCancel(item)}
        on:close={() => handleClose(item)}
      >
        {#if item.props.message}
          <p class="text-sm text-neutral-600">{item.props.message}</p>
        {:else if item.props.content}
          {item.props.content}
        {:else if item.props.component}
          <svelte:component
            this={getComponent(item)}
            {...(item.props.componentProps as Record<string, unknown>) ?? {}}
            on:confirm={(e: CustomEvent) => handleConfirm(item, e.detail)}
            on:cancel={() => handleCancel(item)}
          />
        {/if}
      </Dialog>
    </div>
  {:else if item.type === 'drawer'}
    {@const props = getDrawerProps(item)}
    <div style="z-index: {item.zIndex};" class="modal-stack-item">
      <Drawer
        open={true}
        title={props.title}
        position={props.position ?? 'right'}
        size={props.size ?? 'md'}
        closeOnBackdrop={props.closeOnBackdrop ?? true}
        closeOnEscape={props.closeOnEscape ?? true}
        showClose={props.showClose ?? true}
        overlay={props.overlay ?? true}
        on:close={() => handleClose(item)}
      >
        {#if item.props.content}
          {item.props.content}
        {:else if item.props.component}
          <svelte:component
            this={getComponent(item)}
            {...(item.props.componentProps as Record<string, unknown>) ?? {}}
            on:confirm={(e: CustomEvent) => handleConfirm(item, e.detail)}
            on:cancel={() => handleCancel(item)}
            on:close={() => handleClose(item)}
          />
        {/if}
        <svelte:fragment slot="footer">
          {#if item.props.showFooter && !item.props.component}
            <button
              type="button"
              class="px-4 py-2 text-sm font-medium text-neutral-700 bg-neutral-white border border-neutral-300 rounded-md hover:bg-neutral-50"
              on:click={() => handleCancel(item)}
            >
              {item.props.cancelText ?? 'Cancel'}
            </button>
            <button
              type="button"
              class="px-4 py-2 text-sm font-medium text-neutral-white bg-brand-primary-500 rounded-md hover:bg-brand-primary-600"
              on:click={() => handleConfirm(item)}
            >
              {item.props.confirmText ?? 'OK'}
            </button>
          {/if}
        </svelte:fragment>
      </Drawer>
    </div>
  {/if}
{/each}
