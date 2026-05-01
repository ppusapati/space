export interface DashboardWidget {
  id: string;
  type: 'chart' | 'table' | 'kpi' | 'map' | 'gauge' | 'pivot' | 'summary';
  title: string;
  col: number;    // 1-based grid column
  row: number;    // 1-based grid row
  colSpan: number;
  rowSpan: number;
  config: Record<string, unknown>; // widget-specific config
}

export type DashboardViewport = 'mobile' | 'tablet' | 'desktop' | 'wide';
