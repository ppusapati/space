<script lang="ts">
  /**
   * CheckboxGroup — UI for `CheckboxGroupField` (FieldType.CHECKBOXGROUP).
   *
   * Renders a list of checkboxes corresponding to the field's `options[]`.
   * Selected values are returned as an array. Replaces the prior incorrect
   * mapping that routed CheckboxGroupField to the single-value Checkbox
   * widget — that mapping silently dropped all-but-the-first option's
   * selection state.
   *
   * Honors minSelections / maxSelections by:
   *   - reporting an error when the count is out of range
   *   - blocking checks beyond maxSelections (the box stays unchecked
   *     and a small message appears below)
   *
   * Used for: multi-select roles, vertical-flag tags, "applicable
   * channels" lists. Distinct from MULTI_SELECT (dropdown) — checkbox
   * groups are preferred when the option set is small (<8) and the
   * user benefits from seeing all options at once.
   */
  import type { CheckboxGroupField } from '@chetana/core';
  import { createEventDispatcher } from 'svelte';

  interface Props {
    field: CheckboxGroupField;
    value?: unknown[];
    error?: string;
    onChange?: (next: unknown[]) => void;
  }

  let { field, value = [], error, onChange }: Props = $props();

  const dispatch = createEventDispatcher<{ change: unknown[] }>();

  let selected = $state<unknown[]>(Array.isArray(value) ? [...value] : []);

  const min = field.minSelections ?? 0;
  const max = field.maxSelections ?? Number.POSITIVE_INFINITY;

  const countError = $derived.by(() => {
    if (selected.length < min) return `Select at least ${min}`;
    if (selected.length > max) return `Select at most ${max}`;
    return undefined;
  });

  function toggle(optionValue: unknown, checked: boolean): void {
    if (checked) {
      if (selected.length >= max) return; // block check beyond max
      selected = [...selected, optionValue];
    } else {
      selected = selected.filter((v) => v !== optionValue);
    }
    onChange?.(selected);
    dispatch('change', selected);
  }

  function isChecked(optionValue: unknown): boolean {
    return selected.includes(optionValue);
  }
</script>

<fieldset class="checkbox-group" class:has-error={Boolean(error || countError)}>
  {#if field.label}
    <legend class="checkbox-group-legend">
      {field.label}
      {#if min > 0}<span class="required">*</span>{/if}
    </legend>
  {/if}
  <div class="checkbox-group-options">
    {#each field.options as opt (opt.value)}
      <label class="checkbox-group-option">
        <input
          type="checkbox"
          checked={isChecked(opt.value)}
          disabled={opt.disabled || (!isChecked(opt.value) && selected.length >= max)}
          onchange={(ev) => toggle(opt.value, (ev.currentTarget as HTMLInputElement).checked)}
        />
        <span>{opt.label}</span>
      </label>
    {/each}
  </div>
  {#if error || countError}
    <div class="checkbox-group-error" role="alert">{error ?? countError}</div>
  {/if}
</fieldset>

<style>
  .checkbox-group {
    border: none;
    margin: 0;
    padding: 0;
  }

  .checkbox-group-legend {
    font-size: 0.875rem;
    font-weight: 500;
    margin-bottom: 0.375rem;
    padding: 0;
  }

  .required {
    color: var(--color-destructive, #ef4444);
    margin-left: 0.125rem;
  }

  .checkbox-group-options {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .checkbox-group-option {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    cursor: pointer;
  }

  .checkbox-group-option input[disabled] + span {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .checkbox-group-error {
    color: var(--color-destructive, #ef4444);
    font-size: 0.8125rem;
    margin-top: 0.375rem;
  }

  .has-error .checkbox-group-options {
    /* Subtle indicator without redrawing each checkbox */
    padding-left: 0.25rem;
    border-left: 2px solid var(--color-destructive, #ef4444);
  }
</style>
