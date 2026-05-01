// =============================================================================
// BI Designer Components — Drag-and-drop report & dashboard builder
// =============================================================================

// ─── Field Palette (draggable field list) ────────────────────────────────────
export { default as FieldPalette } from './FieldPalette.svelte';
export type { DatasetFieldItem } from './FieldPalette.types';

// ─── Field Well (drop zones for dimensions, measures, etc.) ──────────────────
export { default as FieldWell } from './FieldWell.svelte';
export type { FieldWellItem } from './FieldWell.types';

// ─── Visual Canvas (live chart/table preview) ───────────────────────────────
export { default as VisualCanvas } from './VisualCanvas.svelte';

// ─── Dashboard Grid (drag-and-resize widget layout) ─────────────────────────
export { default as DashboardGrid } from './DashboardGrid.svelte';
export type { DashboardWidget, DashboardViewport } from './DashboardGrid.types';

// ─── Data Model Diagram (ER relationship diagram) ───────────────────────────
export { default as DataModelDiagram } from './DataModelDiagram.svelte';
export type { DatasetNode, RelationshipEdge } from './DataModelDiagram.types';

// ─── Expression Editor (formula/expression editor with autocomplete) ─────────
export { default as ExpressionEditor } from './ExpressionEditor.svelte';

// ─── Filter Slicer (cross-visual filter component) ──────────────────────────
export { default as FilterSlicer } from './FilterSlicer.svelte';
export type { SlicerFilter } from './FilterSlicer.types';

// ─── Drilldown Manager (drill-through navigation handler) ───────────────────
export { default as DrilldownManager } from './DrilldownManager.svelte';

// ─── Chart Type Picker (visualization type selector) ────────────────────────
export { default as ChartTypePicker } from './ChartTypePicker.svelte';
