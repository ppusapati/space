/**
 * @samavāya/ui - Svelte Component Library for Samavāya ERP
 *
 * All UI components in one package, organized by domain:
 *   - actions     (Button, Accordion, drag-and-drop)
 *   - display     (Card, Badge, Avatar, Carousel)
 *   - feedback    (Modal, Toast, Alert, Spinner)
 *   - layout      (Grid, Stack, Container, AppShell)
 *   - navigation  (Sidebar, NavBar, Tabs, Breadcrumbs)
 *   - forms       (39 field types, FormBuilder, DynamicFormRenderer)
 *   - charts      (ECharts: Bar, Line, Pie, Gauge, etc.)
 *   - tables      (DataGrid, VirtualList, TreeDataTable)
 *   - icons       (400+ SVG icons)
 *   - business    (StatusBadge, ActivityFeed, Calendar, Kanban)
 *   - services    (Modal stack)
 *   - bi          (FieldPalette, FieldWell, VisualCanvas, DashboardGrid, DataModelDiagram, ExpressionEditor, FilterSlicer, DrilldownManager, ChartTypePicker)
 */

// ============================================================================
// CORE (primitives, layout, display, feedback, navigation, actions, services)
// ============================================================================

export * from './types';
export * from './utils';
export * from './actions';
export * from './display';
export * from './feedback';
export * from './layout';
export * from './navigation';
export * from './services';

// ============================================================================
// FORMS (39 field types, builder, renderer)
// ============================================================================

// Exclude TreeNode to avoid conflict with display/display.types.ts TreeNode
export { default as TreeNodeSelector } from './forms/TreeSelector.svelte';
export * from './forms/index';

// ============================================================================
// CHARTS
// ============================================================================

export * from './charts';

// ============================================================================
// TABLES
// ============================================================================

// Exclude treeClasses to avoid conflict with display/display.types.ts treeClasses
export { default as TreeDataTable } from './tables/TreeDataTable.svelte';
export { default as GroupDataTable } from './tables/GroupDataTable.svelte';
export { default as Table } from './tables/Table.svelte';
export { default as DataGrid } from './tables/DataGrid.svelte';
export { default as VirtualList } from './tables/VirtualList.svelte';
export * from './tables/table.types';
export * from './tables/virtuallist.types';
export * from './tables/table.logic';
export * from './tables/table.export';

// ============================================================================
// DATA (Timeline, Kanban, Calendar)
// ============================================================================

export * from './data';

// ============================================================================
// BUSINESS
// ============================================================================

export * from './business';

// ============================================================================
// ICONS
// ============================================================================

// Exclude Icon to avoid conflict with display/Icon.svelte
export { icons, getIconNames, hasIcon } from './icons/icons';
export type { IconName } from './icons/icons';

// ============================================================================
// ERP SHELL (shared layout for all module apps)
// ============================================================================

export * from './erp';

// ============================================================================
// BI (drag-and-drop report & dashboard designer)
// ============================================================================

export * from './bi/index.js';
