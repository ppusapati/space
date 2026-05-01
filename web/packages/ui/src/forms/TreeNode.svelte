<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Icon from '../display/Icon.svelte';

  interface TreeNode {
    value: string | number;
    label: string;
    children?: TreeNode[];
    disabled?: boolean;
  }

  export let nodes: TreeNode[] = [];
  export let expandedNodes: Set<string | number> = new Set();
  export let value: string | number | (string | number)[] = '';
  export let multiple: boolean = false;
  export let disabled: boolean = false;
  export let readonly: boolean = false;
  export let level: number = 0;

  const dispatch = createEventDispatcher<{
    toggle: string | number;
    select: string | number;
  }>();

  function isSelected(nodeValue: string | number): boolean {
    if (Array.isArray(value)) {
      return value.includes(nodeValue);
    }
    return value === nodeValue;
  }

  function isExpanded(nodeValue: string | number): boolean {
    return expandedNodes.has(nodeValue);
  }
</script>

{#each nodes as node (node.value)}
  <div style={`padding-left: ${level * 1.5}rem;`}>
    <div class="flex items-center gap-2 py-1 group">
      {#if node.children && node.children.length > 0}
        <button
          type="button"
          on:click={() => dispatch('toggle', node.value)}
          disabled={disabled || readonly || node.disabled}
          class="p-0 hover:bg-neutral-100 rounded flex-shrink-0 disabled:opacity-50"
        >
          <Icon
            name={isExpanded(node.value) ? 'chevron-down' : 'chevron-right'}
            size="sm"
          />
        </button>
      {:else}
        <div class="w-5" />
      {/if}

      {#if multiple}
        <input
          type="checkbox"
          checked={isSelected(node.value)}
          disabled={disabled || readonly || node.disabled}
          on:change={() => dispatch('select', node.value)}
          class="rounded cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
        />
      {/if}

      <button
        type="button"
        on:click={() => dispatch('select', node.value)}
        disabled={disabled || readonly || node.disabled}
        class={`flex-1 text-left text-sm px-2 py-1 rounded transition-colors ${
          isSelected(node.value)
            ? 'bg-primary-100 text-primary-700 font-medium'
            : 'hover:bg-neutral-100 text-neutral-700 disabled:opacity-50 disabled:cursor-not-allowed'
        }`}
      >
        {node.label}
      </button>
    </div>

    {#if node.children && isExpanded(node.value)}
      <svelte:self
        nodes={node.children}
        {expandedNodes}
        value={value}
        {multiple}
        {disabled}
        {readonly}
        level={level + 1}
        on:toggle
        on:select
      />
    {/if}
  </div>
{/each}
