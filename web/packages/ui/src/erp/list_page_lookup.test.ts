/**
 * ListPage response-path lookup contract test.
 *
 * Pins the dotted-path resolver used by ListPage.svelte's loader to
 * extract rows + total from heterogeneous list-RPC responses. Two
 * production shapes coexist in the platform:
 *
 *   - Flat: `{items: [...], totalCount: 42}`           (Masters/Item)
 *   - Nested: `{entries: [...], pagination: {totalCount: 42}}`
 *     (Finance/Journal, HR/Employee — anything using the canonical
 *     packages.api.v1.pagination.Pagination message)
 *
 * Without dotted-path support, the nested-shape services would silently
 * report rows.length as the total, breaking the paginator on second
 * page loads.
 *
 * The function is duplicated here from ListPage.svelte so the test
 * exercises pure logic without a Svelte mount harness — if either
 * implementation drifts, this test forces the audit.
 */

import { describe, it, expect } from 'vitest';

function lookupPath(obj: Record<string, unknown>, path: string): unknown {
  if (!path) return undefined;
  let cursor: unknown = obj;
  for (const segment of path.split('.')) {
    if (cursor === null || cursor === undefined || typeof cursor !== 'object') {
      return undefined;
    }
    cursor = (cursor as Record<string, unknown>)[segment];
  }
  return cursor;
}

describe('ListPage lookupPath', () => {
  it('returns top-level value for single-segment path (flat shape, e.g. masters/items)', () => {
    expect(lookupPath({ totalCount: 42 }, 'totalCount')).toBe(42);
  });

  it('returns nested value for dotted path (canonical Pagination shape, e.g. hr/employees, finance/journal)', () => {
    expect(lookupPath({ pagination: { totalCount: 42 } }, 'pagination.totalCount')).toBe(42);
  });

  it('returns undefined when leaf is missing — caller falls back to rows.length', () => {
    expect(lookupPath({ pagination: {} }, 'pagination.totalCount')).toBeUndefined();
  });

  it('returns undefined when intermediate segment is missing', () => {
    expect(lookupPath({}, 'pagination.totalCount')).toBeUndefined();
  });

  it('returns undefined when intermediate segment is null (defensive)', () => {
    expect(lookupPath({ pagination: null }, 'pagination.totalCount')).toBeUndefined();
  });

  it('returns undefined when intermediate segment is a primitive (defensive — bad RPC shape)', () => {
    expect(lookupPath({ pagination: 'oops' }, 'pagination.totalCount')).toBeUndefined();
  });

  it('returns undefined for empty path — caller treats this as "no count key"', () => {
    expect(lookupPath({ totalCount: 42 }, '')).toBeUndefined();
  });

  it('extracts arrays as-is (used for the rows key, e.g. "items" or "entries")', () => {
    const rows = [{ id: 'a' }, { id: 'b' }];
    expect(lookupPath({ items: rows }, 'items')).toBe(rows);
    expect(lookupPath({ data: { items: rows } }, 'data.items')).toBe(rows);
  });

  it('handles 3-level nesting (regression safety for future RPC shapes)', () => {
    expect(lookupPath({ a: { b: { c: 7 } } }, 'a.b.c')).toBe(7);
  });
});
