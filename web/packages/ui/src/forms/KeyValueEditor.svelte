<script lang="ts">
  /**
   * KeyValueEditor — UI for `KEYVALUE` (FieldType.KEYVALUE).
   *
   * Renders a list of key/value text pairs with add/remove. Replaces
   * the prior incorrect mapping that routed KEYVALUE to a plain
   * TextArea (which made the user type/parse JSON-ish text by hand
   * and silently ate every malformed entry).
   *
   * Used for: tag dictionaries, custom attribute bags, environment
   * variable lists, HTTP header maps.
   *
   * Value contract: a flat object of string→string. Stored shape
   * matches `Record<string, string>`. On change, emits the full
   * object as the field value.
   */
  import { createEventDispatcher } from 'svelte';

  interface Props {
    name: string;
    label?: string;
    value?: Record<string, string>;
    error?: string;
    onChange?: (next: Record<string, string>) => void;
  }

  let { name, label, value = {}, error, onChange }: Props = $props();

  const dispatch = createEventDispatcher<{ change: Record<string, string> }>();

  // Working copy as an ordered array so the user can sort meaningfully.
  // Map back to a plain object on emit (preserving last-write-wins on
  // duplicate keys — same as JSON.stringify of the object).
  let pairs = $state<{ k: string; v: string }[]>(
    Object.entries(value ?? {}).map(([k, v]) => ({ k, v: String(v ?? '') }))
  );

  function emit(): void {
    const out: Record<string, string> = {};
    for (const p of pairs) {
      if (p.k.length > 0) out[p.k] = p.v;
    }
    onChange?.(out);
    dispatch('change', out);
  }

  function addPair(): void {
    pairs = [...pairs, { k: '', v: '' }];
  }

  function removePair(idx: number): void {
    pairs = pairs.filter((_, i) => i !== idx);
    emit();
  }

  function updateKey(idx: number, key: string): void {
    pairs = pairs.map((p, i) => (i === idx ? { ...p, k: key } : p));
    emit();
  }

  function updateValue(idx: number, val: string): void {
    pairs = pairs.map((p, i) => (i === idx ? { ...p, v: val } : p));
    emit();
  }
</script>

<div class="keyvalue-editor" data-field={name} class:has-error={Boolean(error)}>
  {#if label}
    <div class="keyvalue-label">{label}</div>
  {/if}
  <div class="keyvalue-rows">
    {#each pairs as pair, idx (idx)}
      <div class="keyvalue-row">
        <input
          class="keyvalue-key"
          type="text"
          placeholder="key"
          value={pair.k}
          onchange={(ev) => updateKey(idx, (ev.currentTarget as HTMLInputElement).value)}
        />
        <span class="keyvalue-sep" aria-hidden="true">=</span>
        <input
          class="keyvalue-value"
          type="text"
          placeholder="value"
          value={pair.v}
          onchange={(ev) => updateValue(idx, (ev.currentTarget as HTMLInputElement).value)}
        />
        <button
          type="button"
          class="keyvalue-remove"
          aria-label="Remove pair"
          onclick={() => removePair(idx)}
        >
          ×
        </button>
      </div>
    {/each}
  </div>
  <button type="button" class="keyvalue-add" onclick={addPair}>+ Add pair</button>
  {#if error}
    <div class="keyvalue-error" role="alert">{error}</div>
  {/if}
</div>

<style>
  .keyvalue-editor {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .keyvalue-label {
    font-size: 0.875rem;
    font-weight: 500;
  }

  .keyvalue-rows {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .keyvalue-row {
    display: grid;
    grid-template-columns: 1fr auto 1fr auto;
    align-items: center;
    gap: 0.375rem;
  }

  .keyvalue-key,
  .keyvalue-value {
    padding: 0.375rem 0.5rem;
    border: 1px solid var(--color-border, #e5e7eb);
    border-radius: var(--radius-sm, 0.25rem);
    font-size: 0.875rem;
    min-width: 0;
  }

  .keyvalue-sep {
    color: var(--color-text-secondary, #6b7280);
    font-weight: 600;
  }

  .keyvalue-remove {
    width: 1.75rem;
    height: 1.75rem;
    border: 1px solid var(--color-border, #e5e7eb);
    border-radius: var(--radius-sm, 0.25rem);
    background: var(--color-surface, #fff);
    cursor: pointer;
  }

  .keyvalue-add {
    align-self: flex-start;
    padding: 0.375rem 0.75rem;
    border: 1px dashed var(--color-border, #e5e7eb);
    border-radius: var(--radius-sm, 0.25rem);
    background: transparent;
    color: var(--color-primary, #4f46e5);
    cursor: pointer;
    font-size: 0.875rem;
  }

  .keyvalue-error {
    color: var(--color-destructive, #ef4444);
    font-size: 0.8125rem;
  }

  .has-error .keyvalue-key,
  .has-error .keyvalue-value {
    border-color: var(--color-destructive, #ef4444);
  }
</style>
