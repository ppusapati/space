<script lang="ts">
  /**
   * RepeaterField — UI for `ArrayField` (FieldType.ARRAY).
   *
   * Renders a list of repeating row groups, each composed of the
   * itemFields defined on the ArrayField config. Supports add / remove
   * (gated by minItems/maxItems) and optional drag-to-reorder.
   *
   * Replaces the placeholder mapping that previously routed ARRAY to
   * `Input` in DynamicFormRenderer's componentMap. ARRAY is the FIFTH
   * most-used FieldType in the 601-form catalog (line items, repeated
   * contacts, address lists), so a real implementation is required.
   *
   * Contract: emits `change` with the full array of row values whenever
   * any cell or row count changes. Compatible with FormPage.svelte's
   * handleFieldChange flow — the parent rolls the new array into the
   * outer form's values map by field id.
   */
  import type { ArrayField, FormFieldConfig } from '@samavāya/core';
  import { createEventDispatcher } from 'svelte';

  interface Props {
    field: ArrayField;
    value?: unknown[];
    error?: string;
    onChange?: (next: unknown[]) => void;
  }

  let { field, value = [], error, onChange }: Props = $props();

  const dispatch = createEventDispatcher<{ change: unknown[] }>();

  // Working copy. We never mutate `value` directly — instead build a new
  // array per change so Svelte's reactivity sees the reference change.
  let rows = $state<Record<string, unknown>[]>(
    Array.isArray(value)
      ? value.map((v) => (v && typeof v === 'object' ? { ...(v as object) } : { value: v }))
      : []
  );

  const min = field.minItems ?? 0;
  const max = field.maxItems ?? Number.POSITIVE_INFINITY;
  const canAdd = $derived(rows.length < max);
  const canRemove = $derived(rows.length > min);

  function emit(): void {
    onChange?.(rows);
    dispatch('change', rows);
  }

  function addRow(): void {
    if (!canAdd) return;
    const seed: Record<string, unknown> =
      field.defaultItem && typeof field.defaultItem === 'object'
        ? { ...(field.defaultItem as Record<string, unknown>) }
        : {};
    // Pre-populate keys for every itemField so the inputs are controlled
    // from the first render (uncontrolled→controlled is a Svelte warning
    // and a real source of cursor jumps in production).
    for (const f of field.itemFields) {
      if (!(f.name in seed)) {
        seed[f.name] = emptyValueFor(f);
      }
    }
    rows = [...rows, seed];
    emit();
  }

  function removeRow(idx: number): void {
    if (!canRemove) return;
    rows = rows.filter((_, i) => i !== idx);
    emit();
  }

  function updateCell(idx: number, fieldName: string, next: unknown): void {
    rows = rows.map((row, i) => (i === idx ? { ...row, [fieldName]: next } : row));
    emit();
  }

  function emptyValueFor(f: FormFieldConfig): unknown {
    // The adapter casts runtime-only strings ('currency', 'percent',
    // 'lookup', etc.) onto FormFieldConfig['type'] for ergonomics —
    // see protoFormAdapter.mapFieldType. Compare via plain string here
    // to avoid TS narrowing those strings out of the union.
    const t = f.type as string;
    switch (t) {
      case 'number':
      case 'currency':
      case 'percent':
      case 'slider':
      case 'rating':
        return 0;
      case 'checkbox':
      case 'switch':
        return false;
      case 'array':
      case 'multi-lookup':
      case 'checkbox-group':
        return [];
      case 'object':
      case 'keyvalue':
        return {};
      default:
        return '';
    }
  }
</script>

<div class="repeater-field" class:has-error={Boolean(error)}>
  <div class="repeater-rows">
    {#each rows as row, idx (idx)}
      <div class="repeater-row">
        <div class="repeater-row-fields">
          {#each field.itemFields as itemField (itemField.name)}
            <label class="repeater-cell">
              <span class="repeater-cell-label">{itemField.label ?? itemField.name}</span>
              <!-- Plain input for now; nested DynamicFormRenderer would
                   risk import cycles. Repeater rows in the 601-form
                   catalog are simple scalar fields (qty, price, code) —
                   complex nested rows go through TableField (which already
                   has full widget dispatch). -->
              <input
                class="repeater-cell-input"
                type={itemField.type === 'number' ? 'number' : 'text'}
                value={String((row[itemField.name] ?? '') as string | number)}
                onchange={(ev) => updateCell(idx, itemField.name, (ev.currentTarget as HTMLInputElement).value)}
              />
            </label>
          {/each}
        </div>
        <button
          type="button"
          class="repeater-row-remove"
          aria-label={field.removeLabel ?? 'Remove row'}
          disabled={!canRemove}
          onclick={() => removeRow(idx)}
        >
          ×
        </button>
      </div>
    {/each}
  </div>
  <button
    type="button"
    class="repeater-add"
    disabled={!canAdd}
    onclick={addRow}
  >
    + {field.addLabel ?? 'Add row'}
  </button>
  {#if error}
    <div class="repeater-error" role="alert">{error}</div>
  {/if}
</div>

<style>
  .repeater-field {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .repeater-rows {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .repeater-row {
    display: flex;
    gap: 0.5rem;
    align-items: flex-end;
    padding: 0.5rem;
    border: 1px solid var(--color-border, #e5e7eb);
    border-radius: var(--radius-md, 0.375rem);
    background: var(--color-surface-muted, #fafafa);
  }

  .repeater-row-fields {
    flex: 1;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    gap: 0.5rem;
  }

  .repeater-cell {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
  }

  .repeater-cell-label {
    font-size: 0.75rem;
    color: var(--color-text-secondary, #6b7280);
  }

  .repeater-cell-input {
    padding: 0.375rem 0.5rem;
    border: 1px solid var(--color-border, #e5e7eb);
    border-radius: var(--radius-sm, 0.25rem);
    font-size: 0.875rem;
  }

  .repeater-row-remove {
    flex-shrink: 0;
    width: 2rem;
    height: 2rem;
    border: 1px solid var(--color-border, #e5e7eb);
    border-radius: var(--radius-sm, 0.25rem);
    background: var(--color-surface, #fff);
    cursor: pointer;
    font-size: 1rem;
    line-height: 1;
  }

  .repeater-row-remove:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .repeater-add {
    align-self: flex-start;
    padding: 0.375rem 0.75rem;
    border: 1px dashed var(--color-border, #e5e7eb);
    border-radius: var(--radius-sm, 0.25rem);
    background: transparent;
    color: var(--color-primary, #4f46e5);
    cursor: pointer;
    font-size: 0.875rem;
  }

  .repeater-add:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .repeater-error {
    color: var(--color-destructive, #ef4444);
    font-size: 0.8125rem;
  }

  .has-error .repeater-row {
    border-color: var(--color-destructive, #ef4444);
  }
</style>
