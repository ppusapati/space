<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import Icon from '../display/Icon.svelte';
  import TreeNode from './TreeNode.svelte';
  import type { Size, ValidationState } from '../types';

  interface TreeNodeItem {
    value: string | number;
    label: string;
    children?: TreeNodeItem[];
    disabled?: boolean;
  }

  interface TreeSelectorProps {
    value?: string | number | (string | number)[];
    label?: string;
    helperText?: string;
    errorText?: string;
    disabled?: boolean;
    readonly?: boolean;
    required?: boolean;
    size?: Size;
    state?: ValidationState;
    name?: string;
    id?: string;
    data?: TreeNodeItem[];
    multiple?: boolean;
    expandAll?: boolean;
    clearable?: boolean;
  }

  export let value: string | number | (string | number)[] = '';
  export let label: string = '';
  export let helperText: string = '';
  export let errorText: string = '';
  export let disabled: boolean = false;
  export let readonly: boolean = false;
  export let required: boolean = false;
  export let size: Size = 'md';
  export let state: ValidationState = 'default';
  export let name: string = '';
  export let id: string = uid('tree-selector');
  export let data: TreeNodeItem[] = [];
  export let multiple: boolean = false;
  export let expandAll: boolean = false;
  export let clearable: boolean = true;

  let className: string = '';
  export { className as class };

  let expandedNodes: Set<string | number> = new Set();

  const dispatch = createEventDispatcher<{
    change: string | number | (string | number)[];
  }>();

  const stateClasses = {
    default: 'border-neutral-300 focus:border-primary-500 focus:ring-primary-500',
    success: 'border-green-500 focus:border-green-600 focus:ring-green-500',
    error: 'border-red-500 focus:border-red-600 focus:ring-red-500',
    warning: 'border-yellow-500 focus:border-yellow-600 focus:ring-yellow-500',
  };

  function toggleNode(nodeValue: string | number) {
    if (expandedNodes.has(nodeValue)) {
      expandedNodes.delete(nodeValue);
    } else {
      expandedNodes.add(nodeValue);
    }
    expandedNodes = expandedNodes;
  }

  function selectNode(nodeValue: string | number) {
    if (multiple && Array.isArray(value)) {
      if (value.includes(nodeValue)) {
        value = value.filter((v) => v !== nodeValue);
      } else {
        value = [...value, nodeValue];
      }
    } else {
      value = nodeValue;
    }
    dispatch('change', value);
  }

  function handleClear() {
    value = multiple ? [] : '';
    dispatch('change', value);
  }

  function expandAllNodes() {
    const collectAllValues = (nodes: TreeNodeItem[]) => {
      nodes.forEach((node) => {
        expandedNodes.add(node.value);
        if (node.children) {
          collectAllValues(node.children);
        }
      });
    };
    collectAllValues(data);
    expandedNodes = expandedNodes;
  }

  function collapseAllNodes() {
    expandedNodes.clear();
    expandedNodes = expandedNodes;
  }

  function getSelectedLabels(): string[] {
    const findLabels = (nodes: TreeNodeItem[], selectedValues: (string | number)[]): string[] => {
      const labels: string[] = [];
      nodes.forEach((node) => {
        if (selectedValues.includes(node.value)) {
          labels.push(node.label);
        }
        if (node.children) {
          labels.push(...findLabels(node.children, selectedValues));
        }
      });
      return labels;
    };

    const selectedValues = Array.isArray(value) ? value : [value];
    return findLabels(data, selectedValues);
  }

  $: if (expandAll && expandedNodes.size === 0) {
    expandAllNodes();
  }

  $: stateClass = stateClasses[state as keyof typeof stateClasses] ?? stateClasses.default;
</script>

<div class={cn('w-full', className)}>
  {#if label}
    <label class="block text-sm font-medium text-neutral-700 mb-2">
      {label}
      {#if required}
        <span class="text-red-500 ml-1">*</span>
      {/if}
    </label>
  {/if}

  <div
    class={cn(
      'border rounded-md transition-colors overflow-hidden',
      stateClass,
      disabled && 'bg-neutral-50 cursor-not-allowed opacity-50'
    )}
  >
    <div class="flex gap-1 p-2 border-b border-neutral-200 bg-neutral-50">
      <button
        type="button"
        on:click={expandAllNodes}
        disabled={disabled || readonly}
        class="px-2 py-1 text-xs bg-white hover:bg-neutral-100 border border-neutral-200 rounded disabled:opacity-50"
      >
        Expand All
      </button>
      <button
        type="button"
        on:click={collapseAllNodes}
        disabled={disabled || readonly}
        class="px-2 py-1 text-xs bg-white hover:bg-neutral-100 border border-neutral-200 rounded disabled:opacity-50"
      >
        Collapse All
      </button>
      {#if clearable && (value || (Array.isArray(value) && value.length > 0))}
        <button
          type="button"
          on:click={handleClear}
          disabled={disabled || readonly}
          class="ml-auto px-2 py-1 text-xs bg-red-100 text-red-700 hover:bg-red-200 rounded disabled:opacity-50"
        >
          Clear Selection
        </button>
      {/if}
    </div>

    <div class="max-h-96 overflow-y-auto p-2 bg-white">
      {#if data.length === 0}
        <p class="text-sm text-neutral-500 p-2">No data available</p>
      {:else}
        <TreeNode
          nodes={data}
          {expandedNodes}
          {value}
          {multiple}
          {disabled}
          {readonly}
          on:toggle={(e: CustomEvent<string | number>) => toggleNode(e.detail)}
          on:select={(e: CustomEvent<string | number>) => selectNode(e.detail)}
        />
      {/if}
    </div>
  </div>

  {#if multiple && Array.isArray(value) && value.length > 0}
    <div class="mt-2 flex flex-wrap gap-1">
      {#each getSelectedLabels() as label}
        <span class="bg-primary-100 text-primary-700 px-2 py-0.5 rounded text-xs">
          {label}
        </span>
      {/each}
    </div>
  {:else if !Array.isArray(value) && value}
    <div class="mt-2">
      <span class="bg-primary-100 text-primary-700 px-2 py-0.5 rounded text-xs inline-block">
        {getSelectedLabels()[0]}
      </span>
    </div>
  {/if}

  {#if errorText}
    <p class="mt-1 text-sm text-red-500">{errorText}</p>
  {:else if helperText}
    <p class="mt-1 text-sm text-neutral-500">{helperText}</p>
  {/if}
</div>

