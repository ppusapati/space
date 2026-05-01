/**
 * CrudListPage column-derivation contract test.
 *
 * Pins the rules:
 *
 *   1. If `columns` prop set, those exactly are the columns.
 *   2. If `columns` unset, fall back to first 5 schema.fields.
 *   3. Column labels prefer the schema field's `label`; fall back to
 *      humanized field name.
 *
 * The actual table render isn't tested here (would need jsdom + Svelte
 * mount harness). The column-derivation logic is a pure function of
 * the inputs; verifying it without a render keeps the test fast and
 * deterministic.
 */

import { describe, it, expect } from 'vitest';
import type { FormSchema, FormFieldConfig } from '@samavāya/core';

/**
 * The same logic CrudListPage's `effectiveColumns` $derived uses.
 * Extracted here so the test can exercise it independently — if the
 * component logic changes, this test must be updated and the fact
 * forces an audit.
 */
function deriveColumns(
  schema: FormSchema<Record<string, unknown>>,
  columns?: string[]
): string[] {
  if (columns && columns.length > 0) return columns;
  return schema.fields.slice(0, 5).map((f) => f.name);
}

function deriveLabels(
  schema: FormSchema<Record<string, unknown>>,
  columns: string[]
): string[] {
  return columns.map((name) => {
    const f = schema.fields.find((field) => field.name === name);
    return f?.label ?? humanize(name);
  });
}

function humanize(s: string): string {
  return s.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

const schema = (fields: Partial<FormFieldConfig>[]): FormSchema<Record<string, unknown>> => ({
  fields: fields.map((f) => ({ type: 'text', name: 'unset', ...f }) as FormFieldConfig),
});

describe('CrudListPage column derivation', () => {
  it('uses explicit columns prop when provided', () => {
    const s = schema([
      { name: 'item_code', label: 'Code' },
      { name: 'item_name', label: 'Name' },
      { name: 'category_id', label: 'Category' },
      { name: 'created_at', label: 'Created' },
    ]);
    expect(deriveColumns(s, ['item_code', 'item_name'])).toEqual(['item_code', 'item_name']);
  });

  it('falls back to first 5 schema fields when columns unset', () => {
    const s = schema([
      { name: 'a' },
      { name: 'b' },
      { name: 'c' },
      { name: 'd' },
      { name: 'e' },
      { name: 'f' }, // 6th — must be excluded
      { name: 'g' },
    ]);
    expect(deriveColumns(s)).toEqual(['a', 'b', 'c', 'd', 'e']);
  });

  it('handles schemas with fewer than 5 fields gracefully', () => {
    const s = schema([{ name: 'item_code' }, { name: 'item_name' }]);
    expect(deriveColumns(s)).toEqual(['item_code', 'item_name']);
  });

  it('returns empty array when schema has no fields and no columns', () => {
    const s = schema([]);
    expect(deriveColumns(s)).toEqual([]);
  });

  it('treats empty columns array as fallback trigger', () => {
    // Regression: an empty array shouldn't be treated as "explicit" —
    // it should fall back so the page still renders something.
    const s = schema([{ name: 'a' }, { name: 'b' }]);
    expect(deriveColumns(s, [])).toEqual(['a', 'b']);
  });
});

describe('CrudListPage label derivation', () => {
  it('prefers schema field label over humanized name', () => {
    const s = schema([
      { name: 'item_code', label: 'SKU' },
      { name: 'item_name', label: 'Product' },
    ]);
    expect(deriveLabels(s, ['item_code', 'item_name'])).toEqual(['SKU', 'Product']);
  });

  it('humanizes field name when no label is set', () => {
    const s = schema([{ name: 'created_at' }, { name: 'updated_by' }]);
    expect(deriveLabels(s, ['created_at', 'updated_by'])).toEqual([
      'Created At',
      'Updated By',
    ]);
  });

  it('humanizes column name when column is not in schema fields', () => {
    // Defensive: if the caller asks for a column the schema doesn't
    // declare (e.g. a derived column), fall back to humanized name
    // rather than crashing.
    const s = schema([{ name: 'item_code', label: 'SKU' }]);
    expect(deriveLabels(s, ['item_code', 'derived_total'])).toEqual([
      'SKU',
      'Derived Total',
    ]);
  });
});
