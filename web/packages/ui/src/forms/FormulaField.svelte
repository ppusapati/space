<script lang="ts">
  /**
   * FormulaField — UI for `FORMULA` (FieldType.FORMULA).
   *
   * Renders a read-only computed value with a helper hint that this
   * field is derived. Replaces the prior incorrect mapping that routed
   * FORMULA to a free-edit NumberInput — letting the user type values
   * that the form generator's expression engine would silently
   * overwrite or ignore.
   *
   * The actual formula evaluation happens server-side (per the
   * formbuilder spec — derived attributes use the bounded expression
   * language documented in classregistry). The frontend's job is to
   * (a) display the latest computed value, (b) make clear it can't be
   * edited, (c) re-render when upstream fields change so the user
   * sees a fresh result.
   *
   * Value contract: read-only, no onChange callback. The value comes
   * from the form values map; whenever upstream fields change, the
   * server may re-compute and the new value lands here on next render.
   */

  interface Props {
    name: string;
    label?: string;
    value?: number | string | null;
    hint?: string;
    error?: string;
    formula?: string; // optional: shown as a tooltip / hint
  }

  let { name, label, value, hint, error, formula }: Props = $props();

  const display = $derived(
    value === null || value === undefined || value === ''
      ? '—'
      : typeof value === 'number'
        ? value.toLocaleString()
        : String(value)
  );
</script>

<div class="formula-field" data-field={name} class:has-error={Boolean(error)}>
  {#if label}
    <div class="formula-field-label">
      {label}
      <span class="formula-field-badge" title={formula ?? 'Computed value'}>
        ƒ
      </span>
    </div>
  {/if}
  <output class="formula-field-value" for={name}>{display}</output>
  {#if hint}
    <div class="formula-field-hint">{hint}</div>
  {/if}
  {#if error}
    <div class="formula-field-error" role="alert">{error}</div>
  {/if}
</div>

<style>
  .formula-field {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
  }

  .formula-field-label {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    font-size: 0.875rem;
    font-weight: 500;
  }

  .formula-field-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 1.25rem;
    height: 1.25rem;
    border-radius: 50%;
    background: var(--color-info-muted, #dbeafe);
    color: var(--color-info, #2563eb);
    font-style: italic;
    font-weight: 700;
    font-size: 0.75rem;
    cursor: help;
  }

  .formula-field-value {
    display: block;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--color-border, #e5e7eb);
    border-radius: var(--radius-md, 0.375rem);
    background: var(--color-surface-muted, #f9fafb);
    color: var(--color-text-primary, #111827);
    font-variant-numeric: tabular-nums;
    cursor: not-allowed;
  }

  .formula-field-hint {
    font-size: 0.75rem;
    color: var(--color-text-secondary, #6b7280);
  }

  .formula-field-error {
    color: var(--color-destructive, #ef4444);
    font-size: 0.8125rem;
  }

  .has-error .formula-field-value {
    border-color: var(--color-destructive, #ef4444);
  }
</style>
