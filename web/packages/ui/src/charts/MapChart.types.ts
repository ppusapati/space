export interface MapDataItem {
  name: string;
  value: number;
  lat?: number;
  lng?: number;
  color?: string;
  metadata?: Record<string, unknown>;
}

export type MapType = 'scatter' | 'heatmap' | 'regions';
