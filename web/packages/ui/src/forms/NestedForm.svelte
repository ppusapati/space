<script lang="ts">
  /**
   * NestedForm — UI for `ObjectField` (FieldType.OBJECT and NESTED_FORM).
   *
   * Renders a nested object as a labeled group of child fields. Supports
   * collapsible groups. Used for grouped attributes like `address`,
   * `shipping_details`, `customer_contact`. ObjectField is THIRD most-used
   * non-trivial FieldType in the 601-form catalog.
   *
   * Replaces the placeholder mapping in DynamicFormRenderer's componentMap
   * that previously routed OBJECT to `Input` (silently dropped child
   * fields). The nested fields render as scalar inputs here for the same
   * reason RepeaterField does it inline — avoids an import cycle with
   * DynamicFormRenderer. Forms needing complex nested widgets should use
   * NESTED_FORM with a custom CustomField, or split into separate steps.
   *
   * Contract: emits `change` with the full updated object whenever any
   * child field changes.
   */
  import type { ObjectField, FormFieldConfig } from '@chetana/core';
  import { createEventDispatcher } from 'svelte';

  interface Props {
    field: ObjectField;
    value?: Record<string, unknown>;
    error?: string;
    onChange?: (next: Record<string, unknown>) => void;
  }

  let { field, value = {}, error, onChange }: Props = $props();

  const dispatch = createEventDispatcher<{ change: Record<string, unknown> }>();

  let working = $state<Record<string, unknown>>({ ...(value ?? {}) });
  let collapsed = $state(field.defaultCollapsed ?? false);

  const columns = field.columns ?? 1;
  const gridStyle = $derived(`grid-template-columns: repeat(${columns}, minmax(0, 1fr));`);

  function emit(): void {
    onChange?.(working);
    dispatch('change', working);
  }

  function updateCell(name: string, next: unknown): void {
    working = { ...working, [name]: next };
    emit();
  }

  function inputType(f: FormFieldConfig): string {
    // The adapter produces runtime-only type strings (e.g. 'currency')
    // that aren't in the canonical FormFieldConfig['type'] union — see
    // protoFormAdapter.mapFieldType. Cast to string for switch.
    const t = f.type as string;
    switch (t) {
      case 'number':
      case 'currency':
      case 'percent':
      case 'slider':
        return 'number';
      case 'date':
        return 'date';
      case 'datetime':
        return 'datetime-local';
      case 'time':
        return 'time';
      case 'email':
        return 'email';
      case 'url':
        return 'url';
      case 'tel':
        return 'tel';
      case 'password':
        return 'password';
      default:
        return 'text';
    }
  }
</script>

<fieldset class="nested-form" class:has-error={Boolean(error)}>
  {#if field.label}
    <legend class="nested-form-legend">
      {#if field.collapsible}
        <button
          type="button"
          class="nested-form-toggle"
          aria-expanded={!collapsed}
          onclick={() => (collapsed = !collapsed)}
        >
          {collapsed ? '▶' : '▼'} {field.label}
        </button>
      {:else}
        {field.label}
      {/if}
    </legend>
  {/if}
  {#if !collapsed}
    <div class="nested-form-grid" style={gridStyle}>
      {#each field.fields as childField (childField.name)}
        <label class="nested-form-cell">
          <span class="nested-form-cell-label">{childField.label ?? childField.name}</span>
          {#if childField.type === 'checkbox' || childField.type === 'switch'}
            <input
              type="checkbox"
              checked={Boolean(working[childField.name])}
              onchange={(ev) => updateCell(childField.name, (ev.currentTarget as HTMLInputElement).checked)}
            />
          {:else}
            <input
              class="nested-form-cell-input"
              type={inputType(childField)}
              value={String((working[childField.name] ?? '') as string | number)}
              onchange={(ev) => updateCell(childField.name, (ev.currentTarget as HTMLInputElement).value)}
            />
          {/if}
        </label>
      {/each}
    </div>
  {/if}
  {#if error}
    <div class="nested-form-error" role="alert">{error}</div>
  {/if}
</fieldset>

<style>
  .nested-form {
    border: 1px solid var(--color-border, #e5e7eb);
    border-radius: var(--radius-md, 0.375rem);
    padding: 0.5rem 0.75rem 0.75rem;
    margin: 0;
  }

  .nested-form-legend {
    padding: 0 0.25rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-text-primary, #111827);
  }

  .nested-form-toggle {
    background: transparent;
    border: none;
    padding: 0;
    font: inherit;
    color: inherit;
    cursor: pointer;
  }

  .nested-form-grid {
    display: grid;
    gap: 0.5rem;
  }

  .nested-form-cell {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
  }

  .nested-form-cell-label {
    font-size: 0.75rem;
    color: var(--color-text-secondary, #6b7280);
  }

  .nested-form-cell-input {
    padding: 0.375rem 0.5rem;
    border: 1px solid var(--color-border, #e5e7eb);
    border-radius: var(--radius-sm, 0.25rem);
    font-size: 0.875rem;
  }

  .nested-form-error {
    color: var(--color-destructive, #ef4444);
    font-size: 0.8125rem;
    margin-top: 0.5rem;
  }

  .has-error {
    border-color: var(--color-destructive, #ef4444);
  }
</style>
