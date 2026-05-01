/**
 * Layout Generic Types
 * Layouts define the structural composition of pages
 */

import type { Component, Snippet } from 'svelte';
import type { Size } from './common.types.js';

// ============================================================================
// LAYOUT GENERIC TYPE
// ============================================================================

/**
 * Generic Layout Type
 * @template TSlots - Available slot names
 * @template TConfig - Layout configuration type
 */
export interface Layout<
  TSlots extends string = 'default',
  TConfig extends LayoutConfig = LayoutConfig
> {
  // Configuration
  config: TConfig;

  // Slots
  slots: Partial<Record<TSlots, Snippet | Component | null>>;

  // State
  isCollapsed?: boolean;
  isResponsive?: boolean;
  breakpoint?: Breakpoint;

  // Methods
  toggleCollapse?: () => void;
  setBreakpoint?: (breakpoint: Breakpoint) => void;
}

/** Breakpoint type */
export type Breakpoint = 'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl';

/** Base layout configuration */
export interface LayoutConfig {
  id: string;
  variant: string;
  gap?: Size | 'none';
  padding?: Size | 'none';
  maxWidth?: Size | 'full' | 'prose';
}

// ============================================================================
// LAYOUT VARIANTS
// ============================================================================

/**
 * Master-Detail Layout
 * Two-pane layout with master list and detail view
 */
export interface MasterDetailLayout<TMaster = unknown, TDetail = unknown>
  extends Layout<'master' | 'detail', MasterDetailConfig> {
  // Data
  masterData: TMaster[];
  detailData: TDetail | null;
  selectedId: string | null;

  // State
  detailVisible: boolean;

  // Methods
  onMasterSelect: (id: string) => void;
  onDetailClose: () => void;
}

/** Master-detail configuration */
export interface MasterDetailConfig extends LayoutConfig {
  variant: 'master-detail';
  masterWidth: string;
  detailWidth: string;
  showDetailOnSelect: boolean;
  responsiveBreakpoint: Breakpoint;
  resizable?: boolean;
  collapsibleDetail?: boolean;
}

/**
 * Dashboard Layout
 * Grid-based layout for widgets
 */
export interface DashboardLayout<TWidget extends string = string>
  extends Layout<'widgets' | 'sidebar' | 'header', DashboardLayoutConfig> {
  // Widget layout
  widgets: WidgetLayout<TWidget>[];
  gridCols: number;
  gridRows: number;

  // Methods
  onWidgetResize: (widgetId: TWidget, size: { cols: number; rows: number }) => void;
  onWidgetMove: (widgetId: TWidget, position: { col: number; row: number }) => void;
  onWidgetRemove: (widgetId: TWidget) => void;
  onWidgetAdd: (widget: WidgetLayout<TWidget>) => void;
  saveLayout: () => Promise<void>;
}

/** Dashboard layout configuration */
export interface DashboardLayoutConfig extends LayoutConfig {
  variant: 'dashboard';
  gridCols: number;
  gridRows: number;
  cellHeight: string;
  gap: Size;
  editable: boolean;
  compactMode?: 'vertical' | 'horizontal' | 'none';
}

/** Widget layout position */
export interface WidgetLayout<TId extends string = string> {
  id: TId;
  col: number;
  row: number;
  cols: number;
  rows: number;
  minCols?: number;
  minRows?: number;
  maxCols?: number;
  maxRows?: number;
  locked?: boolean;
  static?: boolean;
}

/**
 * Split Layout
 * Two-pane resizable layout
 */
export interface SplitLayout extends Layout<'primary' | 'secondary', SplitLayoutConfig> {
  // State
  splitRatio: number;
  primaryCollapsed: boolean;
  secondaryCollapsed: boolean;

  // Methods
  onSplitChange: (ratio: number) => void;
  onCollapse: (pane: 'primary' | 'secondary') => void;
  onExpand: (pane: 'primary' | 'secondary') => void;
  resetSplit: () => void;
}

/** Split layout configuration */
export interface SplitLayoutConfig extends LayoutConfig {
  variant: 'split';
  direction: 'horizontal' | 'vertical';
  initialRatio: number;
  minRatio: number;
  maxRatio: number;
  resizable: boolean;
  collapsible: 'none' | 'primary' | 'secondary' | 'both';
  gutterSize: number;
}

/**
 * Tab Layout
 * Tabbed content layout
 */
export interface TabLayout<TTab extends string = string>
  extends Layout<TTab, TabLayoutConfig> {
  // State
  activeTab: TTab;
  tabs: TabConfig<TTab>[];

  // Methods
  onTabChange: (tabId: TTab) => void;
  onTabClose?: (tabId: TTab) => void;
  onTabReorder?: (tabs: TTab[]) => void;
  addTab?: (tab: TabConfig<TTab>) => void;
}

/** Tab layout configuration */
export interface TabLayoutConfig extends LayoutConfig {
  variant: 'tabs';
  position: 'top' | 'bottom' | 'left' | 'right';
  closable: boolean;
  draggable: boolean;
  overflow: 'scroll' | 'dropdown' | 'wrap';
  lazy?: boolean;
}

/** Tab configuration */
export interface TabConfig<TId extends string = string> {
  id: TId;
  label: string;
  icon?: string;
  badge?: string | number;
  closable?: boolean;
  disabled?: boolean;
  keepAlive?: boolean;
}

/**
 * Wizard Layout
 * Step-by-step wizard layout
 */
export interface WizardLayout<TStep extends string = string>
  extends Layout<TStep, WizardLayoutConfig> {
  // State
  currentStep: TStep;
  steps: WizardStep<TStep>[];
  completedSteps: Set<TStep>;
  visitedSteps: Set<TStep>;

  // Computed
  isFirstStep: boolean;
  isLastStep: boolean;
  canProceed: boolean;
  canGoBack: boolean;
  progress: number;

  // Methods
  goToStep: (step: TStep) => void;
  nextStep: () => void;
  prevStep: () => void;
  completeStep: (step: TStep) => void;
  resetWizard: () => void;
}

/** Wizard layout configuration */
export interface WizardLayoutConfig extends LayoutConfig {
  variant: 'wizard';
  orientation: 'horizontal' | 'vertical';
  linear: boolean; // Must complete steps in order
  showStepNumbers: boolean;
  showProgress: boolean;
  allowSkip: boolean;
}

/** Wizard step */
export interface WizardStep<TId extends string = string> {
  id: TId;
  title: string;
  description?: string;
  icon?: string;
  optional?: boolean;
  disabled?: boolean;
  validate?: () => boolean | Promise<boolean>;
}

/**
 * App Shell Layout
 * Main application shell with header, sidebar, content
 */
export interface AppShellLayout
  extends Layout<'header' | 'sidebar' | 'main' | 'footer', AppShellConfig> {
  // State
  sidebarCollapsed: boolean;
  sidebarVisible: boolean;
  sidebarHovered: boolean;

  // Methods
  toggleSidebar: () => void;
  collapseSidebar: () => void;
  expandSidebar: () => void;
  showSidebar: () => void;
  hideSidebar: () => void;
}

/** App shell configuration */
export interface AppShellConfig extends LayoutConfig {
  variant: 'app-shell';
  headerHeight: string;
  footerHeight?: string;
  sidebarWidth: string;
  sidebarCollapsedWidth: string;
  sidebarPosition: 'left' | 'right';
  fixedHeader: boolean;
  fixedSidebar: boolean;
  fixedFooter?: boolean;
  sidebarCollapsible: boolean;
  sidebarBreakpoint?: Breakpoint;
  overlay?: boolean;
}

/**
 * Page Layout
 * Generic page content layout
 */
export interface PageLayout extends Layout<'default', PageLayoutConfig> {
  // No additional state needed
}

/** Page layout configuration */
export interface PageLayoutConfig extends LayoutConfig {
  variant: 'default' | 'centered' | 'sidebar' | 'split';
  maxWidth: Size | 'full' | 'prose';
  padding: Size | 'none';
  header?: boolean;
  footer?: boolean;
}

/**
 * Card Layout
 * Layout for card-based content
 */
export interface CardLayout extends Layout<'cards', CardLayoutConfig> {
  // State
  viewMode: 'grid' | 'list' | 'masonry';

  // Methods
  setViewMode: (mode: 'grid' | 'list' | 'masonry') => void;
}

/** Card layout configuration */
export interface CardLayoutConfig extends LayoutConfig {
  variant: 'cards';
  columns: number | 'auto';
  gap: Size;
  minCardWidth?: string;
  maxCardWidth?: string;
}

/**
 * Form Layout
 * Layout for forms
 */
export interface FormLayout extends Layout<'form' | 'actions', FormLayoutConfig> {
  // No additional state needed
}

/** Form layout configuration */
export interface FormLayoutConfig extends LayoutConfig {
  variant: 'vertical' | 'horizontal' | 'inline' | 'grid';
  columns?: number;
  labelPosition: 'top' | 'left' | 'right';
  labelWidth?: string;
  gap: Size;
  dense?: boolean;
}

// ============================================================================
// RESPONSIVE UTILITIES
// ============================================================================

/** Responsive value */
export type ResponsiveValue<T> = T | Partial<Record<Breakpoint, T>>;

/** Breakpoint values */
export const BREAKPOINTS: Record<Breakpoint, number> = {
  xs: 0,
  sm: 640,
  md: 768,
  lg: 1024,
  xl: 1280,
  '2xl': 1536,
};
