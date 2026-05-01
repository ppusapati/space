export interface FieldWellItem {
  id: string;
  field_id: string;
  label: string;
  role: 'dimension' | 'measure' | 'attribute';
  data_type: string;
  aggregate?: string; // for measures
  granularity?: string; // for date dimensions
  alias?: string;
  sort_direction?: 'asc' | 'desc' | null;
}
