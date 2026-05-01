<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { cn } from '../utils/classnames';
  import { uid } from '../utils/uid';
  import Icon from '../display/Icon.svelte';
  import type { Size, ValidationState } from '../types';

  interface TableColumn {
    key: string;
    label: string;
    type?: 'text' | 'number' | 'select' | 'date' | 'checkbox';
    options?: { label: string; value: string | number }[];
    required?: boolean;
  }

  interface TableRow {
    [key: string]: unknown;
  }

  interface TableFieldProps {
    value?: TableRow[];
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
    columns?: TableColumn[];
    maxRows?: number;
    addRowText?: string;
    deleteRowText?: string;
  }

  export let value: TableRow[] = [];
  export let label: string = '';
  export let helperText: string = '';
  export let errorText: string = '';
  export let disabled: boolean = false;
  export let readonly: boolean = false;
  export let required: boolean = false;
  export let size: Size = 'md';
  export let state: ValidationState = 'default';
  export let name: string = '';
  export let id: string = uid('table-field');
  export let columns: TableColumn[] = [];
  export let maxRows: number | undefined = undefined;
  export let addRowText: string = 'Add Row';
  export let deleteRowText: string = 'Delete';

  let className: string = '';
  export { className as class };

  const dispatch = createEventDispatcher<{
    change: TableRow[];
  }>();

  const stateClasses = {
    default: 'border-neutral-300',
    success: 'border-green-500',
    error: 'border-red-500',
    warning: 'border-yellow-500',
  };

  function addRow() {
    if (maxRows && value.length >= maxRows) return;

    const newRow: TableRow = {};
    columns.forEach((col) => {
      newRow[col.key] = '';
    });

    value = [...value, newRow];
    dispatch('change', value);
  }

  function updateCell(rowIdx: number, colKey: string, newValue: unknown) {
    value[rowIdx]![colKey] = newValue;
    value = value;
    dispatch('change', value);
  }

  function deleteRow(rowIdx: number) {
    value = value.filter((_, idx) => idx !== rowIdx);
    dispatch('change', value);
  }

  function duplicateRow(rowIdx: number) {
    const newRow = { ...value[rowIdx] };
    value = [...value.slice(0, rowIdx + 1), newRow, ...value.slice(rowIdx + 1)];
    dispatch('change', value);
  }

  const canAddRow = !maxRows || value.length < maxRows;
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

  <div class={cn('border rounded-lg overflow-hidden', stateClasses[state as keyof typeof stateClasses] ?? stateClasses.default)}>
    <div class="overflow-x-auto">
      <table class="w-full text-sm">
        <thead class="bg-neutral-100 border-b border-neutral-300">
          <tr>
            <th class="px-4 py-2 text-left font-medium text-neutral-700 w-12">#</th>
            {#each columns as column}
              <th class="px-4 py-2 text-left font-medium text-neutral-700 min-w-[150px]">
                {column.label}
                {#if column.required}
                  <span class="text-red-500 ml-1">*</span>
                {/if}
              </th>
            {/each}
            <th class="px-4 py-2 text-left font-medium text-neutral-700 w-24">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#if value.length === 0}
            <tr>
              <td colspan={columns.length + 2} class="px-4 py-4 text-center text-neutral-500">
                No rows added yet
              </td>
            </tr>
          {:else}
            {#each value as row, rowIdx}
              <tr class="border-b border-neutral-200 hover:bg-neutral-50">
                <td class="px-4 py-2 text-neutral-500 font-medium">{rowIdx + 1}</td>
                {#each columns as column}
                  <td class="px-4 py-2">
                    {#if column.type === 'checkbox'}
                      <input
                        type="checkbox"
                        checked={row[column.key] as boolean || false}
                        on:change={(e) => updateCell(rowIdx, column.key, (e.target as HTMLInputElement).checked)}
                        disabled={disabled || readonly}
                        class="rounded disabled:opacity-50"
                      />
                    {:else if column.type === 'select' && column.options}
                      <select
                        value={row[column.key] as string || ''}
                        on:change={(e) => updateCell(rowIdx, column.key, (e.target as HTMLSelectElement).value)}
                        disabled={disabled || readonly}
                        class="w-full px-2 py-1 border border-neutral-200 rounded text-sm disabled:opacity-50"
                      >
                        <option value="">Select...</option>
                        {#each column.options as option}
                          <option value={option.value}>
                            {option.label}
                          </option>
                        {/each}
                      </select>
                    {:else}
                      <input
                        type={column.type === 'number' ? 'number' : column.type === 'date' ? 'date' : 'text'}
                        value={row[column.key] as string || ''}
                        on:change={(e) => updateCell(rowIdx, column.key, (e.target as HTMLInputElement).value)}
                        disabled={disabled || readonly}
                        required={column.required}
                        class="w-full px-2 py-1 border border-neutral-200 rounded text-sm disabled:opacity-50"
                      />
                    {/if}
                  </td>
                {/each}
                <td class="px-4 py-2">
                  <div class="flex gap-1">
                    <button
                      type="button"
                      on:click={() => duplicateRow(rowIdx)}
                      disabled={disabled || readonly || !!(maxRows && value.length >= maxRows)}
                      title="Duplicate row"
                      class="p-1 hover:bg-neutral-200 rounded disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      <Icon name="copy" size="sm" />
                    </button>
                    <button
                      type="button"
                      on:click={() => deleteRow(rowIdx)}
                      disabled={disabled || readonly}
                      title={deleteRowText}
                      class="p-1 text-red-600 hover:bg-red-100 rounded disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      <Icon name="trash" size="sm" />
                    </button>
                  </div>
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>

    {#if canAddRow}
      <div class="border-t border-neutral-200 p-2 bg-neutral-50">
        <button
          type="button"
          on:click={addRow}
          disabled={disabled || readonly}
          class="px-3 py-1.5 text-sm bg-primary-600 text-white rounded hover:bg-primary-700 disabled:opacity-50 flex items-center gap-2"
        >
          <Icon name="plus" size="sm" />
          {addRowText}
        </button>
      </div>
    {/if}

    {#if maxRows}
      <div class="px-4 py-2 bg-neutral-50 border-t border-neutral-200 text-xs text-neutral-600">
        {value.length} of {maxRows} rows
      </div>
    {/if}
  </div>

  {#if errorText}
    <p class="mt-2 text-sm text-red-500">{errorText}</p>
  {:else if helperText}
    <p class="mt-2 text-sm text-neutral-500">{helperText}</p>
  {/if}
</div>
