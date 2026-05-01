<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { TreeProps, TreeNode } from './display.types';
  import { treeClasses } from './display.types';
  import { cn } from '../utils';

  type $$Props = TreeProps;

  export let nodes: $$Props['nodes'] = [];
  export let multiSelect: $$Props['multiSelect'] = false;
  export let checkable: $$Props['checkable'] = false;
  export let expandedKeys: $$Props['expandedKeys'] = [];
  export let selectedKeys: $$Props['selectedKeys'] = [];
  export let showLines: $$Props['showLines'] = false;
  export let size: $$Props['size'] = 'md';
  let className: $$Props['class'] = undefined;
  export { className as class };

  const dispatch = createEventDispatcher<{
    expand: { node: TreeNode; expanded: boolean };
    select: { node: TreeNode; selected: boolean; selectedKeys: string[] };
    check: { node: TreeNode; checked: boolean; checkedKeys: string[] };
  }>();

  let internalExpandedKeys = new Set(expandedKeys);
  let internalSelectedKeys = new Set(selectedKeys);
  let checkedKeys = new Set<string>();

  $: internalExpandedKeys = new Set(expandedKeys);
  $: internalSelectedKeys = new Set(selectedKeys);

  function toggleExpand(node: TreeNode) {
    if (node.disabled || !node.children?.length) return;

    const expanded = !internalExpandedKeys.has(node.id);
    if (expanded) {
      internalExpandedKeys.add(node.id);
    } else {
      internalExpandedKeys.delete(node.id);
    }
    internalExpandedKeys = internalExpandedKeys;
    dispatch('expand', { node, expanded });
  }

  function toggleSelect(node: TreeNode) {
    if (node.disabled) return;

    const selected = !internalSelectedKeys.has(node.id);

    if (!multiSelect) {
      internalSelectedKeys.clear();
    }

    if (selected) {
      internalSelectedKeys.add(node.id);
    } else {
      internalSelectedKeys.delete(node.id);
    }
    internalSelectedKeys = internalSelectedKeys;

    dispatch('select', {
      node,
      selected,
      selectedKeys: Array.from(internalSelectedKeys),
    });
  }

  function toggleCheck(node: TreeNode) {
    if (node.disabled) return;

    const checked = !checkedKeys.has(node.id);

    if (checked) {
      checkedKeys.add(node.id);
    } else {
      checkedKeys.delete(node.id);
    }
    checkedKeys = checkedKeys;

    dispatch('check', {
      node,
      checked,
      checkedKeys: Array.from(checkedKeys),
    });
  }

  function handleKeydown(e: KeyboardEvent, node: TreeNode) {
    switch (e.key) {
      case 'Enter':
      case ' ':
        e.preventDefault();
        if (checkable) {
          toggleCheck(node);
        } else {
          toggleSelect(node);
        }
        break;
      case 'ArrowRight':
        if (node.children?.length && !internalExpandedKeys.has(node.id)) {
          toggleExpand(node);
        }
        break;
      case 'ArrowLeft':
        if (node.children?.length && internalExpandedKeys.has(node.id)) {
          toggleExpand(node);
        }
        break;
    }
  }
</script>

<div class={cn(treeClasses.container, className)} role="tree">
  {#each nodes as node (node.id)}
    <div class={treeClasses.node} role="treeitem" aria-expanded={node.children?.length ? internalExpandedKeys.has(node.id) : undefined}>
      <div
        class={cn(
          treeClasses.nodeContent,
          internalSelectedKeys.has(node.id) && treeClasses.nodeSelected,
          node.disabled && treeClasses.nodeDisabled
        )}
        tabindex={node.disabled ? -1 : 0}
        on:click={() => checkable ? toggleCheck(node) : toggleSelect(node)}
        on:keydown={(e) => handleKeydown(e, node)}
      >
        <!-- Expand/Collapse Icon -->
        {#if node.children?.length}
          <button
            type="button"
            class={cn(
              treeClasses.expandIcon,
              internalExpandedKeys.has(node.id) && treeClasses.expandIconExpanded
            )}
            on:click|stopPropagation={() => toggleExpand(node)}
            aria-label={internalExpandedKeys.has(node.id) ? 'Collapse' : 'Expand'}
          >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
            </svg>
          </button>
        {:else}
          <span class="w-4 h-4 shrink-0" />
        {/if}

        <!-- Checkbox -->
        {#if checkable}
          <input
            type="checkbox"
            class={treeClasses.checkbox}
            checked={checkedKeys.has(node.id)}
            disabled={node.disabled}
            on:click|stopPropagation
            on:change={() => toggleCheck(node)}
          />
        {/if}

        <!-- Node Icon -->
        {#if node.icon}
          <svg
            class={treeClasses.icon}
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
            aria-hidden="true"
          >
            <path stroke-linecap="round" stroke-linejoin="round" d={node.icon} />
          </svg>
        {/if}

        <!-- Label -->
        <span class={treeClasses.label}>{node.label}</span>
      </div>

      <!-- Children -->
      {#if node.children?.length && internalExpandedKeys.has(node.id)}
        <div class={cn(treeClasses.children, showLines && treeClasses.lines)} role="group">
          <svelte:self
            nodes={node.children}
            {multiSelect}
            {checkable}
            expandedKeys={Array.from(internalExpandedKeys)}
            selectedKeys={Array.from(internalSelectedKeys)}
            {showLines}
            {size}
            on:expand
            on:select
            on:check
          />
        </div>
      {/if}
    </div>
  {/each}
</div>
