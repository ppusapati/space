<script lang="ts">
  /**
   * CustomFieldRenderer — UI for `CustomField` (a field whose component
   * the form schema specifies directly). Replaces the placeholder
   * mapping that used to route `custom` to `Input` (which silently
   * dropped the schema's `component` reference).
   *
   * The CustomField shape carries:
   *   - component: Component  — Svelte 5 component constructor to mount
   *   - props?: Record<string, unknown>  — extra props to forward
   *
   * This renderer mounts the specified component with the form value,
   * standard props (label, error), and any extra props the schema
   * provides. If `component` is null/undefined, falls back to nothing
   * rendered (with a console warning visible only in dev).
   *
   * Used for: domain-specific widgets that don't fit the 39 built-in
   * FieldType enum values — e.g. a BOM tree editor, a gantt picker,
   * a pivot configurator. Forms wire these by declaring `type: 'custom'`
   * with their own component import.
   */
  import type { CustomField } from '@samavāya/core';
  import type { Component } from 'svelte';

  interface Props {
    field: CustomField;
    value?: unknown;
    error?: string;
    onChange?: (next: unknown) => void;
  }

  let { field, value, error, onChange }: Props = $props();

  // The component reference is required by the CustomField contract;
  // emit a one-line dev-only warning when it's missing so the form
  // doesn't silently render nothing.
  $effect(() => {
    if (!field.component && typeof console !== 'undefined') {
      console.warn(
        `[CustomFieldRenderer] field "${field.name}" is type:'custom' but has no component reference; rendering nothing.`
      );
    }
  });

  // Aliased to FieldComponent to avoid colliding with the `Component`
  // type imported from svelte at the top of the file.
  const FieldComponent = $derived(field.component as Component | undefined);
</script>

{#if FieldComponent}
  <svelte:component
    this={FieldComponent}
    {value}
    {error}
    label={field.label}
    name={field.name}
    onChange={onChange}
    {...(field.props ?? {})}
  />
{/if}
